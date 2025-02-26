// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package pgbookmark

import (
	"context"
	"errors"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	bookmarkquerying "github.com/virtualtam/sparklemuffin/pkg/bookmark/querying"
)

func (r *Repository) bookmarkGetQuery(query string, queryParams ...any) (bookmark.Bookmark, error) {
	rows, err := r.Pool.Query(context.Background(), query, queryParams...)
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

func (r *Repository) bookmarkGetManyQuery(query string, queryParams ...any) ([]bookmark.Bookmark, error) {
	rows, err := r.Pool.Query(context.Background(), query, queryParams...)
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

	batchResults := r.Pool.SendBatch(ctx, batch)
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

	for range bookmarks {
		commandTag, qerr := batchResults.Exec()
		if qerr != nil {
			return 0, qerr
		}

		rowsAffected += commandTag.RowsAffected()
	}

	return rowsAffected, nil
}

func (r *Repository) tagGetQuery(query string, queryParams ...any) ([]bookmarkquerying.Tag, error) {
	rows, err := r.Pool.Query(context.Background(), query, queryParams...)
	if err != nil {
		return []bookmarkquerying.Tag{}, err
	}
	defer rows.Close()

	var dbTags []DBTag

	if err := pgxscan.ScanAll(&dbTags, rows); err != nil {
		return []bookmarkquerying.Tag{}, err
	}

	var tags []bookmarkquerying.Tag

	for _, dbTag := range dbTags {
		tag := bookmarkquerying.NewTag(dbTag.Name, dbTag.Count)
		tags = append(tags, tag)
	}

	return tags, nil
}
