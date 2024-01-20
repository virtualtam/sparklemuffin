// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark/exporting"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark/importing"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark/querying"
)

var (
	fullTextSearchReplacer = strings.NewReplacer("/", " ", ".", " ")
)

type DBBookmark struct {
	UID      string `db:"uid"`
	UserUUID string `db:"user_uuid"`

	URL         string `db:"url"`
	Title       string `db:"title"`
	Description string `db:"description"`

	Private bool     `db:"private"`
	Tags    []string `db:"tags"`

	FullTextSearchString string `db:"fulltextsearch_string"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func bookmarkToFullTextSearchString(b bookmark.Bookmark) string {
	return fmt.Sprintf(
		"%s %s %s",
		b.Title,
		fullTextSearchReplacer.Replace(b.Description),
		fullTextSearchReplacer.Replace(strings.Join(b.Tags, " ")),
	)
}

type DBTag struct {
	Name  string `db:"name"`
	Count uint   `db:"count"`
}

var _ bookmark.Repository = &Repository{}
var _ exporting.Repository = &Repository{}
var _ importing.Repository = &Repository{}
var _ querying.Repository = &Repository{}

func (r *Repository) BookmarkAdd(b bookmark.Bookmark) error {
	query := `
	INSERT INTO bookmarks(
		uid,
		user_uuid,
		url,
		title,
		description,
		private,
		tags,
		fulltextsearch_tsv,
		created_at,
		updated_at
	)
	VALUES(
		@uid,
		@user_uuid,
		@url,
		@title,
		@description,
		@private,
		@tags,
		to_tsvector(@fulltextsearch_string),
		@created_at,
		@updated_at
	)`

	fullTextSearchString := bookmarkToFullTextSearchString(b)

	args := pgx.NamedArgs{
		"uid":                   b.UID,
		"user_uuid":             b.UserUUID,
		"url":                   b.URL,
		"title":                 b.Title,
		"description":           b.Description,
		"private":               b.Private,
		"tags":                  b.Tags,
		"fulltextsearch_string": fullTextSearchString,
		"created_at":            b.CreatedAt,
		"updated_at":            b.UpdatedAt,
	}

	ctx := context.Background()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer r.rollback(ctx, tx, "bookmarks", "add")

	_, err = tx.Exec(ctx, query, args)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) BookmarkAddMany(bookmarks []bookmark.Bookmark) (int64, error) {
	return r.bookmarkUpsertMany("ON CONFLICT DO NOTHING", bookmarks)
}

func (r *Repository) BookmarkUpsertMany(bookmarks []bookmark.Bookmark) (int64, error) {
	return r.bookmarkUpsertMany(
		`
ON CONFLICT (user_uuid, url) DO UPDATE
SET
	title              = EXCLUDED.title,
	description        = EXCLUDED.description,
	private            = EXCLUDED.private,
	tags               = EXCLUDED.tags,
	fulltextsearch_tsv = EXCLUDED.fulltextsearch_tsv,
	created_at         = EXCLUDED.created_at,
	updated_at         = EXCLUDED.updated_at
`,
		bookmarks,
	)
}

func (r *Repository) bookmarkUpsertMany(onConflictStmt string, bookmarks []bookmark.Bookmark) (int64, error) {
	insertQuery := `
	INSERT INTO bookmarks(
		uid,
		user_uuid,
		url,
		title,
		description,
		private,
		tags,
		fulltextsearch_tsv,
		created_at,
		updated_at
	)
	VALUES(
		@uid,
		@user_uuid,
		@url,
		@title,
		@description,
		@private,
		@tags,
		to_tsvector(@fulltextsearch_string),
		@created_at,
		@updated_at
	)`

	query := insertQuery + onConflictStmt

	batch := &pgx.Batch{}

	for _, b := range bookmarks {
		fullTextSearchString := bookmarkToFullTextSearchString(b)

		args := pgx.NamedArgs{
			"uid":                   b.UID,
			"user_uuid":             b.UserUUID,
			"url":                   b.URL,
			"title":                 b.Title,
			"description":           b.Description,
			"private":               b.Private,
			"tags":                  b.Tags,
			"fulltextsearch_string": fullTextSearchString,
			"created_at":            b.CreatedAt,
			"updated_at":            b.UpdatedAt,
		}

		batch.Queue(query, args)
	}

	ctx := context.Background()

	batchResults := r.pool.SendBatch(ctx, batch)
	defer func() {
		if err := batchResults.Close(); err != nil {
			log.Error().
				Err(err).
				Str("domain", "bookmarks").
				Str("operation", "upsert_many").
				Msg("failed to close batch results")
		}
	}()

	var rowsAffected int64

	for i := 0; i < len(bookmarks); i++ {
		commandTag, qerr := batchResults.Exec()
		if qerr != nil {
			return 0, qerr
		}

		rowsAffected += commandTag.RowsAffected()
	}

	return rowsAffected, nil
}

func (r *Repository) BookmarkDelete(userUUID, uid string) error {
	ctx := context.Background()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer r.rollback(ctx, tx, "bookmarks", "delete")

	commandTag, err := tx.Exec(
		context.Background(),
		"DELETE FROM bookmarks WHERE user_uuid=$1 AND uid=$2",
		userUUID,
		uid,
	)
	if err != nil {
		return err
	}

	rowsAffected := commandTag.RowsAffected()

	if rowsAffected != 1 {
		return bookmark.ErrNotFound
	}

	return tx.Commit(ctx)
}

func (r *Repository) bookmarkGetManyQuery(query string, queryParams ...any) ([]bookmark.Bookmark, error) {
	rows, err := r.pool.Query(context.Background(), query, queryParams...)
	if err != nil {
		return []bookmark.Bookmark{}, err
	}
	defer rows.Close()

	dbBookmarks := []DBBookmark{}

	if err := pgxscan.ScanAll(&dbBookmarks, rows); err != nil {
		return []bookmark.Bookmark{}, err
	}

	bookmarks := []bookmark.Bookmark{}

	for _, dbBookmark := range dbBookmarks {
		bookmark := bookmark.Bookmark{
			UserUUID:    dbBookmark.UserUUID,
			UID:         dbBookmark.UID,
			URL:         dbBookmark.URL,
			Title:       dbBookmark.Title,
			Description: dbBookmark.Description,
			Private:     dbBookmark.Private,
			Tags:        dbBookmark.Tags,
			CreatedAt:   dbBookmark.CreatedAt,
			UpdatedAt:   dbBookmark.UpdatedAt,
		}

		bookmarks = append(bookmarks, bookmark)
	}

	return bookmarks, nil
}

func (r *Repository) bookmarkGetQuery(query string, queryParams ...any) (bookmark.Bookmark, error) {
	rows, err := r.pool.Query(context.Background(), query, queryParams...)
	if err != nil {
		return bookmark.Bookmark{}, err
	}
	defer rows.Close()

	dbBookmark := &DBBookmark{}
	err = pgxscan.ScanOne(dbBookmark, rows)

	if errors.Is(err, pgx.ErrNoRows) {
		return bookmark.Bookmark{}, bookmark.ErrNotFound
	}
	if err != nil {
		return bookmark.Bookmark{}, err
	}

	return bookmark.Bookmark{
		UserUUID:    dbBookmark.UserUUID,
		UID:         dbBookmark.UID,
		URL:         dbBookmark.URL,
		Title:       dbBookmark.Title,
		Description: dbBookmark.Description,
		Private:     dbBookmark.Private,
		Tags:        dbBookmark.Tags,
		CreatedAt:   dbBookmark.CreatedAt,
		UpdatedAt:   dbBookmark.UpdatedAt,
	}, nil
}

func (r *Repository) BookmarkGetAll(userUUID string) ([]bookmark.Bookmark, error) {
	return r.bookmarkGetManyQuery(
		`
SELECT user_uuid, uid, url, title, description, private, tags, created_at, updated_at
FROM bookmarks
WHERE user_uuid=$1
ORDER BY created_at DESC`,
		userUUID,
	)
}

func (r *Repository) BookmarkGetAllPrivate(userUUID string) ([]bookmark.Bookmark, error) {
	return r.bookmarkGetManyQuery(
		`
SELECT user_uuid, uid, url, title, description, private, tags, created_at, updated_at
FROM bookmarks
WHERE user_uuid=$1
AND   private=TRUE
ORDER BY created_at DESC`,
		userUUID,
	)
}

func (r *Repository) BookmarkGetAllPublic(userUUID string) ([]bookmark.Bookmark, error) {
	return r.bookmarkGetManyQuery(
		`
SELECT user_uuid, uid, url, title, description, private, tags, created_at, updated_at
FROM bookmarks
WHERE user_uuid=$1
AND   private=FALSE
ORDER BY created_at DESC`,
		userUUID,
	)
}

func (r *Repository) BookmarkGetByTag(userUUID string, tag string) ([]bookmark.Bookmark, error) {
	query := `
	SELECT user_uuid, uid, url, title, description, private, tags, created_at, updated_at
	FROM bookmarks
	WHERE user_uuid=$1
	AND   $2=ANY(tags)`

	return r.bookmarkGetManyQuery(
		query,
		userUUID,
		tag,
	)
}

func (r *Repository) BookmarkGetByUID(userUUID, uid string) (bookmark.Bookmark, error) {
	query := `
	SELECT user_uuid, uid, url, title, description, private, tags, created_at, updated_at
	FROM bookmarks
	WHERE user_uuid=$1
	AND uid=$2`

	return r.bookmarkGetQuery(query, userUUID, uid)
}

func (r *Repository) BookmarkGetByURL(userUUID, u string) (bookmark.Bookmark, error) {
	query := `
	SELECT user_uuid, uid, url, title, description, private, tags, created_at, updated_at
	FROM bookmarks
	WHERE user_uuid=$1
	AND url=$2`

	return r.bookmarkGetQuery(query, userUUID, u)
}

func (r *Repository) BookmarkGetCount(userUUID string, visibility querying.Visibility) (uint, error) {
	var query string

	switch visibility {
	case querying.VisibilityPrivate:
		query = `
		SELECT COUNT(*)
		FROM  bookmarks
		WHERE user_uuid=$1
		AND   private=TRUE`

	case querying.VisibilityPublic:
		query = `
		SELECT COUNT(*)
		FROM  bookmarks
		WHERE user_uuid=$1
		AND   private=FALSE`

	default:
		query = `
		SELECT COUNT(*)
		FROM bookmarks
		WHERE user_uuid=$1`
	}

	var count uint

	err := r.pool.QueryRow(
		context.Background(),
		query,
		userUUID,
	).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *Repository) BookmarkGetN(userUUID string, visibility querying.Visibility, n uint, offset uint) ([]bookmark.Bookmark, error) {
	var query string

	switch visibility {
	case querying.VisibilityPrivate:
		query = `
		SELECT user_uuid, uid, url, title, description, private, tags, created_at, updated_at
		FROM  bookmarks
		WHERE user_uuid=$1
		AND   private=TRUE
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	case querying.VisibilityPublic:
		query = `
		SELECT user_uuid, uid, url, title, description, private, tags, created_at, updated_at
		FROM  bookmarks
		WHERE user_uuid=$1
		AND   private=FALSE
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	default:
		query = `
		SELECT user_uuid, uid, url, title, description, private, tags, created_at, updated_at
		FROM  bookmarks
		WHERE user_uuid=$1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`
	}

	return r.bookmarkGetManyQuery(
		query,
		userUUID,
		n,
		offset,
	)
}

func (r *Repository) BookmarkGetPublicByUID(userUUID, uid string) (bookmark.Bookmark, error) {
	query := `
	SELECT user_uuid, uid, url, title, description, private, tags, created_at, updated_at
	FROM bookmarks
	WHERE user_uuid=$1
	AND uid=$2
	AND private=false`

	return r.bookmarkGetQuery(query, userUUID, uid)
}

func (r *Repository) BookmarkSearchCount(userUUID string, visibility querying.Visibility, searchTerms string) (uint, error) {
	var query string

	switch visibility {
	case querying.VisibilityPrivate:
		query = `
		SELECT COUNT(*)
		FROM bookmarks
		WHERE user_uuid=$1
		AND PRIVATE=TRUE
		AND fulltextsearch_tsv @@ websearch_to_tsquery($2)`

	case querying.VisibilityPublic:
		query = `
		SELECT COUNT(*)
		FROM bookmarks
		WHERE user_uuid=$1
		AND PRIVATE=FALSE
		AND fulltextsearch_tsv @@ websearch_to_tsquery($2)`

	default:
		query = `
		SELECT COUNT(*)
		FROM bookmarks
		WHERE user_uuid=$1
		AND fulltextsearch_tsv @@ websearch_to_tsquery($2)`
	}

	var count uint
	fullTextSearchTerms := fullTextSearchReplacer.Replace(searchTerms)

	err := r.pool.QueryRow(
		context.Background(),
		query,
		userUUID,
		fullTextSearchTerms,
	).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *Repository) BookmarkSearchN(userUUID string, visibility querying.Visibility, searchTerms string, n uint, offset uint) ([]bookmark.Bookmark, error) {
	var query string

	switch visibility {
	case querying.VisibilityPrivate:
		query = `
		SELECT user_uuid, uid, url, title, description, private, tags, created_at, updated_at
		FROM bookmarks
		WHERE user_uuid=$1
		AND private=TRUE
		AND fulltextsearch_tsv @@ websearch_to_tsquery($2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	case querying.VisibilityPublic:
		query = `
		SELECT user_uuid, uid, url, title, description, private, tags, created_at, updated_at
		FROM bookmarks
		WHERE user_uuid=$1
		AND private=FALSE
		AND fulltextsearch_tsv @@ websearch_to_tsquery($2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	default:
		query = `
		SELECT user_uuid, uid, url, title, description, private, tags, created_at, updated_at
		FROM bookmarks
		WHERE user_uuid=$1
		AND fulltextsearch_tsv @@ websearch_to_tsquery($2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`
	}

	fullTextSearchTerms := fullTextSearchReplacer.Replace(searchTerms)

	return r.bookmarkGetManyQuery(
		query,
		userUUID,
		fullTextSearchTerms,
		n,
		offset,
	)
}

func (r *Repository) BookmarkIsURLRegistered(userUUID, url string) (bool, error) {
	return r.rowExistsByQuery(
		"SELECT 1 FROM bookmarks WHERE user_uuid=$1 AND url=$2",
		userUUID,
		url,
	)
}

func (r *Repository) BookmarkIsURLRegisteredToAnotherUID(userUUID, url, uid string) (bool, error) {
	return r.rowExistsByQuery(
		"SELECT 1 FROM bookmarks WHERE user_uuid=$1 AND url=$2 AND uid!=$3",
		userUUID,
		url,
		uid,
	)
}

func (r *Repository) BookmarkTagUpdateMany(bookmarks []bookmark.Bookmark) (int64, error) {
	return r.BookmarkUpsertMany(bookmarks)
}

func (r *Repository) BookmarkUpdate(b bookmark.Bookmark) error {
	query := `
	UPDATE bookmarks
	SET
		url=@url,
		title=@title,
		description=@description,
		private=@private,
		tags=@tags,
		fulltextsearch_tsv=to_tsvector(@fulltextsearch_string),
		updated_at=@updated_at
	WHERE user_uuid=@user_uuid
	AND uid=@uid
			`

	fullTextSearchString := bookmarkToFullTextSearchString(b)

	args := pgx.NamedArgs{
		"user_uuid":             b.UserUUID,
		"uid":                   b.UID,
		"url":                   b.URL,
		"title":                 b.Title,
		"description":           b.Description,
		"private":               b.Private,
		"tags":                  b.Tags,
		"fulltextsearch_string": fullTextSearchString,
		"updated_at":            b.UpdatedAt,
	}

	ctx := context.Background()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer r.rollback(ctx, tx, "bookmarks", "update")

	_, err = tx.Exec(ctx, query, args)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) OwnerGetByUUID(userUUID string) (querying.Owner, error) {
	query := `
	SELECT uuid, nick_name, display_name
	FROM users
	WHERE uuid=$1`

	dbUser := &DBUser{}

	rows, err := r.pool.Query(
		context.Background(),
		query,
		userUUID,
	)
	if err != nil {
		return querying.Owner{}, err
	}
	defer rows.Close()

	err = pgxscan.ScanOne(dbUser, rows)

	if errors.Is(err, sql.ErrNoRows) {
		return querying.Owner{}, querying.ErrOwnerNotFound
	}
	if err != nil {
		return querying.Owner{}, err
	}

	return querying.Owner{
		UUID:        dbUser.UUID,
		NickName:    dbUser.NickName,
		DisplayName: dbUser.DisplayName,
	}, nil
}

func (r *Repository) tagGetQuery(query string, queryParams ...any) ([]querying.Tag, error) {
	rows, err := r.pool.Query(context.Background(), query, queryParams...)
	if err != nil {
		return []querying.Tag{}, err
	}
	defer rows.Close()

	var dbTags []DBTag

	if err := pgxscan.ScanAll(&dbTags, rows); err != nil {
		return []querying.Tag{}, err
	}

	var tags []querying.Tag

	for _, dbTag := range dbTags {
		tag := querying.NewTag(dbTag.Name, dbTag.Count)
		tags = append(tags, tag)
	}

	return tags, nil
}

func (r *Repository) TagGetCount(userUUID string, visibility querying.Visibility) (uint, error) {
	var query string

	switch visibility {
	case querying.VisibilityPrivate:
		query = `
		SELECT COUNT(DISTINCT name)
		FROM (
			SELECT unnest(tags) as name
			FROM bookmarks
			WHERE user_uuid=$1
			AND   private=TRUE
		) s`

	case querying.VisibilityPublic:
		query = `
		SELECT COUNT(DISTINCT name)
		FROM (
			SELECT unnest(tags) as name
			FROM bookmarks
			WHERE user_uuid=$1
			AND   private=FALSE
		) s`

	default:
		query = `
		SELECT COUNT(DISTINCT name)
		FROM (
			SELECT unnest(tags) as name
			FROM bookmarks
			WHERE user_uuid=$1
		) s`
	}

	var count uint

	err := r.pool.QueryRow(
		context.Background(),
		query,
		userUUID,
	).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *Repository) TagGetAll(userUUID string, visibility querying.Visibility) ([]querying.Tag, error) {
	var query string

	switch visibility {
	case querying.VisibilityPrivate:
		query = `
		SELECT name, COUNT(name) as count
		FROM (
			SELECT unnest(tags) as name
			FROM  bookmarks
			WHERE user_uuid=$1
			AND   private=TRUE
		) s
		GROUP BY name
		ORDER BY count DESC, name ASC`

	case querying.VisibilityPublic:
		query = `
		SELECT name, COUNT(name) as count
		FROM (
			SELECT unnest(tags) as name
			FROM  bookmarks
			WHERE user_uuid=$1
			AND   private=FALSE
		) s
		GROUP BY name
		ORDER BY count DESC, name ASC`

	default:
		query = `
		SELECT name, COUNT(name) as count
		FROM (
			SELECT unnest(tags) as name
			FROM  bookmarks
			WHERE user_uuid=$1
		) s
		GROUP BY name
		ORDER BY count DESC, name ASC`
	}

	return r.tagGetQuery(query, userUUID)
}

func (r *Repository) TagGetN(userUUID string, visibility querying.Visibility, n uint, offset uint) ([]querying.Tag, error) {
	var query string

	switch visibility {
	case querying.VisibilityPrivate:
		query = `
		SELECT name, COUNT(name) as count
		FROM (
			SELECT unnest(tags) as name
			FROM  bookmarks
			WHERE user_uuid=$1
			AND   private=TRUE
		) s
		GROUP BY name
		ORDER BY count DESC, name ASC
		LIMIT $2 OFFSET $3`

	case querying.VisibilityPublic:
		query = `
		SELECT name, COUNT(name) as count
		FROM (
			SELECT unnest(tags) as name
			FROM  bookmarks
			WHERE user_uuid=$1
			AND   private=FALSE
		) s
		GROUP BY name
		ORDER BY count DESC, name ASC
		LIMIT $2 OFFSET $3`

	default:
		query = `
		SELECT name, COUNT(name) as count
		FROM (
			SELECT unnest(tags) as name
			FROM  bookmarks
			WHERE user_uuid=$1
		) s
		GROUP BY name
		ORDER BY count DESC, name ASC
		LIMIT $2 OFFSET $3`
	}

	return r.tagGetQuery(query, userUUID, n, offset)
}

func (r *Repository) TagFilterCount(userUUID string, visibility querying.Visibility, filterTerm string) (uint, error) {
	var query string

	switch visibility {
	case querying.VisibilityPrivate:
		query = `
		SELECT COUNT(DISTINCT name)
		FROM (
			SELECT unnest(tags) as name
			FROM bookmarks
			WHERE user_uuid=$1
			AND   private=TRUE
		) s
		WHERE name ILIKE $2`

	case querying.VisibilityPublic:
		query = `
		SELECT COUNT(DISTINCT name)
		FROM (
			SELECT unnest(tags) as name
			FROM bookmarks
			WHERE user_uuid=$1
			AND   private=FALSE
		) s
		WHERE name ILIKE $2`

	default:
		query = `
		SELECT COUNT(DISTINCT name)
		FROM (
			SELECT unnest(tags) as name
			FROM bookmarks
			WHERE user_uuid=$1
		) s
		WHERE name ILIKE $2`
	}

	var count uint

	err := r.pool.QueryRow(
		context.Background(),
		query,
		userUUID,
		"%"+filterTerm+"%",
	).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *Repository) TagFilterN(userUUID string, visibility querying.Visibility, filterTerm string, n uint, offset uint) ([]querying.Tag, error) {
	var query string

	switch visibility {
	case querying.VisibilityPrivate:
		query = `
		SELECT name, COUNT(name) as count
		FROM (
			SELECT unnest(tags) as name
			FROM  bookmarks
			WHERE user_uuid=$1
			AND   private=TRUE
		) s
		WHERE name ILIKE $2
		GROUP BY name
		ORDER BY count DESC, name ASC
		LIMIT $3 OFFSET $4`

	case querying.VisibilityPublic:
		query = `
		SELECT name, COUNT(name) as count
		FROM (
			SELECT unnest(tags) as name
			FROM  bookmarks
			WHERE user_uuid=$1
			AND   private=FALSE
		) s
		WHERE name ILIKE $2
		GROUP BY name
		ORDER BY count DESC, name ASC
		LIMIT $3 OFFSET $4`

	default:
		query = `
		SELECT name, COUNT(name) as count
		FROM (
			SELECT unnest(tags) as name
			FROM  bookmarks
			WHERE user_uuid=$1
		) s
		WHERE name ILIKE $2
		GROUP BY name
		ORDER BY count DESC, name ASC
		LIMIT $3 OFFSET $4`
	}

	return r.tagGetQuery(query, userUUID, "%"+filterTerm+"%", n, offset)
}
