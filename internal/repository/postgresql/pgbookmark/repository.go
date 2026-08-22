// Copyright VirtualTam 2022, 2026
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
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgtaxonomy"
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

	taxonomyRepo *pgtaxonomy.Repository
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		Repository:   pgbase.NewRepository(pool),
		taxonomyRepo: pgtaxonomy.NewRepository(pool),
	}
}

const (
	domain = "bookmarks"

	bookmarkSelectQuery = `
	SELECT
		b.user_uuid,
		b.uid,
		b.url,
		b.title,
		b.description,
		b.private,
		COALESCE(ARRAY_AGG(tt.tag_name ORDER BY tt.tag_name) FILTER (WHERE tt.tag_name IS NOT NULL), '{}') AS tags,
		b.created_at,
		b.updated_at
	FROM bookmarks b
	LEFT JOIN bookmark_tags bt ON bt.user_uuid = b.user_uuid AND bt.bookmark_uid = b.uid
	LEFT JOIN taxonomy_tags tt ON tt.tag_uuid = bt.tag_uuid
	`

	bookmarkGroupByClause = `
	GROUP BY b.user_uuid, b.uid, b.url, b.title, b.description, b.private, b.created_at, b.updated_at
	`

	// bookmarkPageQuery aggregates tags onto an already-paginated "page" CTE
	// (see BookmarkGetN/BookmarkSearchN), instead of onto every bookmark
	// matching the WHERE clause: the join+aggregate work is bounded by the
	// page size, not by how many bookmarks a user has.
	bookmarkPageQuery = `
	SELECT
		p.user_uuid,
		p.uid,
		p.url,
		p.title,
		p.description,
		p.private,
		COALESCE(ARRAY_AGG(tt.tag_name ORDER BY tt.tag_name) FILTER (WHERE tt.tag_name IS NOT NULL), '{}') AS tags,
		p.created_at,
		p.updated_at
	FROM page p
	LEFT JOIN bookmark_tags bt ON bt.user_uuid = p.user_uuid AND bt.bookmark_uid = p.uid
	LEFT JOIN taxonomy_tags tt ON tt.tag_uuid = bt.tag_uuid
	GROUP BY p.user_uuid, p.uid, p.url, p.title, p.description, p.private, p.created_at, p.updated_at
	ORDER BY p.created_at DESC
	`

	bookmarkTagsSearchCondition = `
	EXISTS (
		SELECT 1
		FROM bookmark_tags bt2
		JOIN taxonomy_tags tt2 ON tt2.tag_uuid = bt2.tag_uuid
		WHERE bt2.user_uuid = b.user_uuid
		AND   bt2.bookmark_uid = b.uid
		AND   tt2.tag_name_tsv @@ websearch_to_tsquery($2)
	)
	`
)

func (r *Repository) BookmarkAdd(ctx context.Context, b bookmark.Bookmark) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer r.Rollback(ctx, tx, domain, "BookmarkAdd")

	tags, err := r.getOrCreateTagsTx(ctx, tx, b.UserUUID, b.Tags)
	if err != nil {
		return err
	}

	query := `
	INSERT INTO bookmarks(
		uid,
		user_uuid,
		url,
		title,
		description,
		private,
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
		TO_TSVECTOR(@fulltextsearch_string),
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
		"fulltextsearch_string": fullTextSearchString,
		"created_at":            b.CreatedAt,
		"updated_at":            b.UpdatedAt,
	}

	if _, err := tx.Exec(ctx, query, args); err != nil {
		return err
	}

	if err := insertBookmarkTagsTx(ctx, tx, b.UserUUID, b.UID, tags); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) BookmarkAddMany(ctx context.Context, bookmarks []bookmark.Bookmark) (int64, error) {
	return r.bookmarkUpsertMany(ctx, "ON CONFLICT DO NOTHING", bookmarks)
}

func (r *Repository) BookmarkUpsertMany(ctx context.Context, bookmarks []bookmark.Bookmark) (int64, error) {
	return r.bookmarkUpsertMany(
		ctx,
		`
ON CONFLICT (user_uuid, url) DO UPDATE
SET
	title              = EXCLUDED.title,
	description        = EXCLUDED.description,
	private            = EXCLUDED.private,
	fulltextsearch_tsv = EXCLUDED.fulltextsearch_tsv,
	created_at         = EXCLUDED.created_at,
	updated_at         = EXCLUDED.updated_at
`,
		bookmarks,
	)
}

func (r *Repository) BookmarkDelete(ctx context.Context, userUUID, uid string) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer r.Rollback(ctx, tx, domain, "BookmarkDelete")

	commandTag, err := tx.Exec(
		ctx,
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

func (r *Repository) BookmarkGetAll(ctx context.Context, userUUID string) ([]bookmark.Bookmark, error) {
	query := bookmarkSelectQuery + `
	WHERE b.user_uuid=$1
	` + bookmarkGroupByClause + `
	ORDER BY b.created_at DESC`

	return r.bookmarkGetManyQuery(ctx, query, userUUID)
}

func (r *Repository) BookmarkGetAllPrivate(ctx context.Context, userUUID string) ([]bookmark.Bookmark, error) {
	query := bookmarkSelectQuery + `
	WHERE b.user_uuid=$1
	AND   b.private=TRUE
	` + bookmarkGroupByClause + `
	ORDER BY b.created_at DESC`

	return r.bookmarkGetManyQuery(ctx, query, userUUID)
}

func (r *Repository) BookmarkGetAllPublic(ctx context.Context, userUUID string) ([]bookmark.Bookmark, error) {
	query := bookmarkSelectQuery + `
	WHERE b.user_uuid=$1
	AND   b.private=FALSE
	` + bookmarkGroupByClause + `
	ORDER BY b.created_at DESC`

	return r.bookmarkGetManyQuery(ctx, query, userUUID)
}

func (r *Repository) BookmarkGetByUID(ctx context.Context, userUUID, uid string) (bookmark.Bookmark, error) {
	query := bookmarkSelectQuery + `
	WHERE b.user_uuid=$1
	AND   b.uid=$2
	` + bookmarkGroupByClause

	return r.bookmarkGetQuery(ctx, query, userUUID, uid)
}

func (r *Repository) BookmarkGetByURL(ctx context.Context, userUUID, u string) (bookmark.Bookmark, error) {
	query := bookmarkSelectQuery + `
	WHERE b.user_uuid=$1
	AND   b.url=$2
	` + bookmarkGroupByClause

	return r.bookmarkGetQuery(ctx, query, userUUID, u)
}

func (r *Repository) BookmarkGetCount(ctx context.Context, userUUID string, visibility bookmarkquerying.Visibility) (uint, error) {
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
		ctx,
		query,
		userUUID,
	).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *Repository) BookmarkGetN(ctx context.Context, userUUID string, visibility bookmarkquerying.Visibility, n uint, offset uint) ([]bookmark.Bookmark, error) {
	var query string

	switch visibility {
	case bookmarkquerying.VisibilityPrivate:
		query = `
		WITH page AS (
			SELECT b.user_uuid, b.uid, b.url, b.title, b.description, b.private, b.created_at, b.updated_at
			FROM bookmarks b
			WHERE b.user_uuid=$1
			AND   b.private=TRUE
			ORDER BY b.created_at DESC
			LIMIT $2 OFFSET $3
		)
		` + bookmarkPageQuery

	case bookmarkquerying.VisibilityPublic:
		query = `
		WITH page AS (
			SELECT b.user_uuid, b.uid, b.url, b.title, b.description, b.private, b.created_at, b.updated_at
			FROM bookmarks b
			WHERE b.user_uuid=$1
			AND   b.private=FALSE
			ORDER BY b.created_at DESC
			LIMIT $2 OFFSET $3
		)
		` + bookmarkPageQuery

	default:
		query = `
		WITH page AS (
			SELECT b.user_uuid, b.uid, b.url, b.title, b.description, b.private, b.created_at, b.updated_at
			FROM bookmarks b
			WHERE b.user_uuid=$1
			ORDER BY b.created_at DESC
			LIMIT $2 OFFSET $3
		)
		` + bookmarkPageQuery
	}

	return r.bookmarkGetManyQuery(
		ctx,
		query,
		userUUID,
		n,
		offset,
	)
}

func (r *Repository) BookmarkGetPublicByUID(ctx context.Context, userUUID, uid string) (bookmark.Bookmark, error) {
	query := bookmarkSelectQuery + `
	WHERE b.user_uuid=$1
	AND   b.uid=$2
	AND   b.private=FALSE
	` + bookmarkGroupByClause

	return r.bookmarkGetQuery(ctx, query, userUUID, uid)
}

func (r *Repository) BookmarkSearchCount(ctx context.Context, userUUID string, visibility bookmarkquerying.Visibility, searchTerms string) (uint, error) {
	var query string

	switch visibility {
	case bookmarkquerying.VisibilityPrivate:
		query = `
		SELECT COUNT(*)
		FROM bookmarks b
		WHERE b.user_uuid=$1
		AND   b.private=TRUE
		AND   (b.fulltextsearch_tsv @@ websearch_to_tsquery($2) OR ` + bookmarkTagsSearchCondition + `)`

	case bookmarkquerying.VisibilityPublic:
		query = `
		SELECT COUNT(*)
		FROM bookmarks b
		WHERE b.user_uuid=$1
		AND   b.private=FALSE
		AND   (b.fulltextsearch_tsv @@ websearch_to_tsquery($2) OR ` + bookmarkTagsSearchCondition + `)`

	default:
		query = `
		SELECT COUNT(*)
		FROM bookmarks b
		WHERE b.user_uuid=$1
		AND   (b.fulltextsearch_tsv @@ websearch_to_tsquery($2) OR ` + bookmarkTagsSearchCondition + `)`
	}

	var count uint
	fullTextSearchTerms := pgbase.FullTextSearchReplacer.Replace(searchTerms)

	err := r.Pool.QueryRow(
		ctx,
		query,
		userUUID,
		fullTextSearchTerms,
	).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *Repository) BookmarkSearchN(ctx context.Context, userUUID string, visibility bookmarkquerying.Visibility, searchTerms string, n uint, offset uint) ([]bookmark.Bookmark, error) {
	var query string

	switch visibility {
	case bookmarkquerying.VisibilityPrivate:
		query = `
		WITH page AS (
			SELECT b.user_uuid, b.uid, b.url, b.title, b.description, b.private, b.created_at, b.updated_at
			FROM bookmarks b
			WHERE b.user_uuid=$1
			AND   b.private=TRUE
			AND   (b.fulltextsearch_tsv @@ websearch_to_tsquery($2) OR ` + bookmarkTagsSearchCondition + `)
			ORDER BY b.created_at DESC
			LIMIT $3 OFFSET $4
		)
		` + bookmarkPageQuery

	case bookmarkquerying.VisibilityPublic:
		query = `
		WITH page AS (
			SELECT b.user_uuid, b.uid, b.url, b.title, b.description, b.private, b.created_at, b.updated_at
			FROM bookmarks b
			WHERE b.user_uuid=$1
			AND   b.private=FALSE
			AND   (b.fulltextsearch_tsv @@ websearch_to_tsquery($2) OR ` + bookmarkTagsSearchCondition + `)
			ORDER BY b.created_at DESC
			LIMIT $3 OFFSET $4
		)
		` + bookmarkPageQuery

	default:
		query = `
		WITH page AS (
			SELECT b.user_uuid, b.uid, b.url, b.title, b.description, b.private, b.created_at, b.updated_at
			FROM bookmarks b
			WHERE b.user_uuid=$1
			AND   (b.fulltextsearch_tsv @@ websearch_to_tsquery($2) OR ` + bookmarkTagsSearchCondition + `)
			ORDER BY b.created_at DESC
			LIMIT $3 OFFSET $4
		)
		` + bookmarkPageQuery
	}

	fullTextSearchTerms := pgbase.FullTextSearchReplacer.Replace(searchTerms)

	return r.bookmarkGetManyQuery(
		ctx,
		query,
		userUUID,
		fullTextSearchTerms,
		n,
		offset,
	)
}

func (r *Repository) BookmarkIsURLRegistered(ctx context.Context, userUUID, url string) (bool, error) {
	return r.RowExistsByQuery(
		ctx,
		"SELECT 1 FROM bookmarks WHERE user_uuid=$1 AND url=$2",
		userUUID,
		url,
	)
}

func (r *Repository) BookmarkIsURLRegisteredToAnotherUID(ctx context.Context, userUUID, url, uid string) (bool, error) {
	return r.RowExistsByQuery(
		ctx,
		"SELECT 1 FROM bookmarks WHERE user_uuid=$1 AND url=$2 AND uid!=$3",
		userUUID,
		url,
		uid,
	)
}

func (r *Repository) BookmarkUpdate(ctx context.Context, b bookmark.Bookmark) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer r.Rollback(ctx, tx, domain, "BookmarkUpdate")

	tags, err := r.getOrCreateTagsTx(ctx, tx, b.UserUUID, b.Tags)
	if err != nil {
		return err
	}

	query := `
	UPDATE bookmarks
	SET
		url=@url,
		title=@title,
		description=@description,
		private=@private,
		fulltextsearch_tsv=TO_TSVECTOR(@fulltextsearch_string),
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
		"fulltextsearch_string": fullTextSearchString,
		"updated_at":            b.UpdatedAt,
	}

	if _, err := tx.Exec(ctx, query, args); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, "DELETE FROM bookmark_tags WHERE user_uuid=$1 AND bookmark_uid=$2", b.UserUUID, b.UID); err != nil {
		return err
	}

	if err := insertBookmarkTagsTx(ctx, tx, b.UserUUID, b.UID, tags); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) OwnerGetByUUID(ctx context.Context, userUUID string) (bookmarkquerying.Owner, error) {
	query := `
	SELECT uuid, nick_name, display_name
	FROM users
	WHERE uuid=$1`

	dbUser := &pguser.DBUser{}

	rows, err := r.Pool.Query(
		ctx,
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
