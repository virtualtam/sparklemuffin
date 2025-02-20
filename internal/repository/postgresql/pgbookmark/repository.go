// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package pgbookmark

import (
	"context"
	"database/sql"
	"errors"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgbase"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pguser"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	bookmarkexporting "github.com/virtualtam/sparklemuffin/pkg/bookmark/exporting"
	bookmarkimporting "github.com/virtualtam/sparklemuffin/pkg/bookmark/importing"
	bookmarkquerying "github.com/virtualtam/sparklemuffin/pkg/bookmark/querying"
)

var _ bookmark.Repository = &Repository{}
var _ bookmarkexporting.Repository = &Repository{}
var _ bookmarkimporting.Repository = &Repository{}
var _ bookmarkquerying.Repository = &Repository{}

type Repository struct {
	*pgbase.Repository
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		Repository: pgbase.NewRepository(pool),
	}
}

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

	return r.QueryTx("bookmarks", "BookmarkAdd", query, args)
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

func (r *Repository) BookmarkDelete(userUUID, uid string) error {
	ctx := context.Background()

	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer r.Rollback(ctx, tx, "bookmarks", "BookmarkDelete")

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

func (r *Repository) BookmarkGetCount(userUUID string, visibility bookmarkquerying.Visibility) (uint, error) {
	var query string

	switch visibility {
	case bookmarkquerying.VisibilityPrivate:
		query = `
		SELECT COUNT(*)
		FROM  bookmarks
		WHERE user_uuid=$1
		AND   private=TRUE`

	case bookmarkquerying.VisibilityPublic:
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

	err := r.Pool.QueryRow(
		context.Background(),
		query,
		userUUID,
	).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *Repository) BookmarkGetN(userUUID string, visibility bookmarkquerying.Visibility, n uint, offset uint) ([]bookmark.Bookmark, error) {
	var query string

	switch visibility {
	case bookmarkquerying.VisibilityPrivate:
		query = `
		SELECT user_uuid, uid, url, title, description, private, tags, created_at, updated_at
		FROM  bookmarks
		WHERE user_uuid=$1
		AND   private=TRUE
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	case bookmarkquerying.VisibilityPublic:
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

func (r *Repository) BookmarkSearchCount(userUUID string, visibility bookmarkquerying.Visibility, searchTerms string) (uint, error) {
	var query string

	switch visibility {
	case bookmarkquerying.VisibilityPrivate:
		query = `
		SELECT COUNT(*)
		FROM bookmarks
		WHERE user_uuid=$1
		AND PRIVATE=TRUE
		AND fulltextsearch_tsv @@ websearch_to_tsquery($2)`

	case bookmarkquerying.VisibilityPublic:
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
	fullTextSearchTerms := pgbase.FullTextSearchReplacer.Replace(searchTerms)

	err := r.Pool.QueryRow(
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

func (r *Repository) BookmarkSearchN(userUUID string, visibility bookmarkquerying.Visibility, searchTerms string, n uint, offset uint) ([]bookmark.Bookmark, error) {
	var query string

	switch visibility {
	case bookmarkquerying.VisibilityPrivate:
		query = `
		SELECT user_uuid, uid, url, title, description, private, tags, created_at, updated_at
		FROM bookmarks
		WHERE user_uuid=$1
		AND private=TRUE
		AND fulltextsearch_tsv @@ websearch_to_tsquery($2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	case bookmarkquerying.VisibilityPublic:
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

	fullTextSearchTerms := pgbase.FullTextSearchReplacer.Replace(searchTerms)

	return r.bookmarkGetManyQuery(
		query,
		userUUID,
		fullTextSearchTerms,
		n,
		offset,
	)
}

func (r *Repository) BookmarkIsURLRegistered(userUUID, url string) (bool, error) {
	return r.RowExistsByQuery(
		"SELECT 1 FROM bookmarks WHERE user_uuid=$1 AND url=$2",
		userUUID,
		url,
	)
}

func (r *Repository) BookmarkIsURLRegisteredToAnotherUID(userUUID, url, uid string) (bool, error) {
	return r.RowExistsByQuery(
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

	return r.QueryTx("bookmarks", "BookmarkUpdate", query, args)
}

func (r *Repository) OwnerGetByUUID(userUUID string) (bookmarkquerying.Owner, error) {
	query := `
	SELECT uuid, nick_name, display_name
	FROM users
	WHERE uuid=$1`

	dbUser := &pguser.DBUser{}

	rows, err := r.Pool.Query(
		context.Background(),
		query,
		userUUID,
	)
	if err != nil {
		return bookmarkquerying.Owner{}, err
	}
	defer rows.Close()

	err = pgxscan.ScanOne(dbUser, rows)

	if errors.Is(err, sql.ErrNoRows) {
		return bookmarkquerying.Owner{}, bookmarkquerying.ErrOwnerNotFound
	}
	if err != nil {
		return bookmarkquerying.Owner{}, err
	}

	return bookmarkquerying.Owner{
		UUID:        dbUser.UUID,
		NickName:    dbUser.NickName,
		DisplayName: dbUser.DisplayName,
	}, nil
}

func (r *Repository) BookmarkTagGetCount(userUUID string, visibility bookmarkquerying.Visibility) (uint, error) {
	var query string

	switch visibility {
	case bookmarkquerying.VisibilityPrivate:
		query = `
		SELECT COUNT(DISTINCT name)
		FROM (
			SELECT unnest(tags) as name
			FROM bookmarks
			WHERE user_uuid=$1
			AND   private=TRUE
		) s`

	case bookmarkquerying.VisibilityPublic:
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

	err := r.Pool.QueryRow(
		context.Background(),
		query,
		userUUID,
	).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *Repository) BookmarkTagGetAll(userUUID string, visibility bookmarkquerying.Visibility) ([]bookmarkquerying.Tag, error) {
	var query string

	switch visibility {
	case bookmarkquerying.VisibilityPrivate:
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

	case bookmarkquerying.VisibilityPublic:
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

func (r *Repository) BookmarkTagGetN(userUUID string, visibility bookmarkquerying.Visibility, n uint, offset uint) ([]bookmarkquerying.Tag, error) {
	var query string

	switch visibility {
	case bookmarkquerying.VisibilityPrivate:
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

	case bookmarkquerying.VisibilityPublic:
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

func (r *Repository) BookmarkTagFilterCount(userUUID string, visibility bookmarkquerying.Visibility, filterTerm string) (uint, error) {
	var query string

	switch visibility {
	case bookmarkquerying.VisibilityPrivate:
		query = `
		SELECT COUNT(DISTINCT name)
		FROM (
			SELECT unnest(tags) as name
			FROM bookmarks
			WHERE user_uuid=$1
			AND   private=TRUE
		) s
		WHERE name ILIKE $2`

	case bookmarkquerying.VisibilityPublic:
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

	err := r.Pool.QueryRow(
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

func (r *Repository) BookmarkTagFilterN(userUUID string, visibility bookmarkquerying.Visibility, filterTerm string, n uint, offset uint) ([]bookmarkquerying.Tag, error) {
	var query string

	switch visibility {
	case bookmarkquerying.VisibilityPrivate:
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

	case bookmarkquerying.VisibilityPublic:
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
