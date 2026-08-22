// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package pgbookmark

import (
	"context"
	"errors"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"

	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgbase"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	"github.com/virtualtam/sparklemuffin/pkg/taxonomy"
)

// getOrCreateTagsTx resolves a set of tag names to Tags for a given user,
// creating any tag that does not already exist, all within the transaction
// carried by q.
func (r *Repository) getOrCreateTagsTx(ctx context.Context, q pgbase.Querier, userUUID string, names []string) ([]taxonomy.Tag, error) {
	tags := make([]taxonomy.Tag, 0, len(names))

	for _, name := range names {
		tag, err := r.taxonomyRepo.TagGetByNameTx(ctx, q, userUUID, name)

		if errors.Is(err, taxonomy.ErrNotFound) {
			tag, err = taxonomy.NewTag(userUUID, name)
			if err != nil {
				return nil, err
			}

			if err := tag.Validate(); err != nil {
				return nil, err
			}

			if err := r.taxonomyRepo.TagAddTx(ctx, q, tag); err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

// insertBookmarkTagsTx inserts one bookmark_tags row per tag, within the
// transaction carried by q.
func insertBookmarkTagsTx(ctx context.Context, q pgbase.Querier, userUUID, bookmarkUID string, tags []taxonomy.Tag) error {
	for _, tag := range tags {
		_, err := q.Exec(
			ctx,
			"INSERT INTO bookmark_tags(user_uuid, bookmark_uid, tag_uuid) VALUES($1, $2, $3)",
			userUUID,
			bookmarkUID,
			tag.UUID,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) bookmarkGetQuery(ctx context.Context, query string, queryParams ...any) (bookmark.Bookmark, error) {
	rows, err := r.Pool.Query(ctx, query, queryParams...)
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

func (r *Repository) bookmarkGetManyQuery(ctx context.Context, query string, queryParams ...any) ([]bookmark.Bookmark, error) {
	rows, err := r.Pool.Query(ctx, query, queryParams...)
	if err != nil {
		return []bookmark.Bookmark{}, err
	}
	defer rows.Close()

	var dbBookmarks []DBBookmark

	if err := pgxscan.ScanAll(&dbBookmarks, rows); err != nil {
		return []bookmark.Bookmark{}, err
	}

	var bookmarks []bookmark.Bookmark

	for _, dbBookmark := range dbBookmarks {
		b := bookmark.Bookmark{
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

		bookmarks = append(bookmarks, b)
	}

	return bookmarks, nil
}

// bookmarkUpsertMany upserts a batch of bookmarks and their tags within a
// single transaction.
func (r *Repository) bookmarkUpsertMany(ctx context.Context, onConflictStmt string, bookmarks []bookmark.Bookmark) (int64, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}

	defer r.Rollback(ctx, tx, domain, "bookmarkUpsertMany")

	insertQuery := `
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

	query := insertQuery + onConflictStmt + "\nRETURNING uid"

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
			"fulltextsearch_string": fullTextSearchString,
			"created_at":            b.CreatedAt,
			"updated_at":            b.UpdatedAt,
		}

		batch.Queue(query, args)
	}

	batchResults := tx.SendBatch(ctx, batch)

	var rowsAffected int64
	upsertedUIDs := make([]string, len(bookmarks))

	for i := range bookmarks {
		var uid string

		err := batchResults.QueryRow().Scan(&uid)
		if errors.Is(err, pgx.ErrNoRows) {
			// ON CONFLICT DO NOTHING skipped this bookmark.
			continue
		}
		if err != nil {
			_ = batchResults.Close()
			return 0, err
		}

		upsertedUIDs[i] = uid
		rowsAffected++
	}

	if err := batchResults.Close(); err != nil {
		return 0, err
	}

	tagsBatch := &pgx.Batch{}

	for i, uid := range upsertedUIDs {
		if uid == "" {
			// ON CONFLICT DO NOTHING skipped this bookmark: resolving its
			// tags now would create taxonomy_tags rows that never get
			// attached to anything.
			continue
		}

		tags, err := r.getOrCreateTagsTx(ctx, tx, bookmarks[i].UserUUID, bookmarks[i].Tags)
		if err != nil {
			return 0, err
		}

		tagsBatch.Queue("DELETE FROM bookmark_tags WHERE user_uuid=$1 AND bookmark_uid=$2", bookmarks[i].UserUUID, uid)

		for _, tag := range tags {
			tagsBatch.Queue(
				"INSERT INTO bookmark_tags(user_uuid, bookmark_uid, tag_uuid) VALUES($1, $2, $3)",
				bookmarks[i].UserUUID,
				uid,
				tag.UUID,
			)
		}
	}

	if tagsBatch.Len() > 0 {
		tagsBatchResults := tx.SendBatch(ctx, tagsBatch)

		for range tagsBatch.Len() {
			if _, err := tagsBatchResults.Exec(); err != nil {
				_ = tagsBatchResults.Close()
				return 0, err
			}
		}

		if err := tagsBatchResults.Close(); err != nil {
			return 0, err
		}
	}

	return rowsAffected, tx.Commit(ctx)
}
