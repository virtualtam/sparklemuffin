// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package postgresql

import (
	"context"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	fquerying "github.com/virtualtam/sparklemuffin/pkg/feed/querying"
)

var _ feed.Repository = &Repository{}
var _ fquerying.Repository = &Repository{}

func (r *Repository) FeedCreate(f feed.Feed) error {
	query := `
	INSERT INTO feed_feeds(
		uuid,
		feed_url,
		title,
		slug,
		created_at,
		updated_at
	)
	VALUES(
		@uuid,
		@feed_url,
		@title,
		@slug,
		@created_at,
		@updated_at
	)`

	args := pgx.NamedArgs{
		"uuid":       f.UUID,
		"feed_url":   f.FeedURL,
		"title":      f.Title,
		"slug":       f.Slug,
		"created_at": f.CreatedAt,
		"updated_at": f.UpdatedAt,
	}

	ctx := context.Background()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer r.rollback(ctx, tx, "feeds", "create feed")

	_, err = tx.Exec(ctx, query, args)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) FeedGetByURL(feedURL string) (feed.Feed, error) {
	query := `
SELECT uuid, feed_url, title, slug
FROM feed_feeds
WHERE feed_url=$1`

	return r.feedGetQuery(query, feedURL)
}

func (r *Repository) FeedGetCategories(userUUID string) ([]feed.Category, error) {
	query := `
SELECT uuid, name, slug
FROM feed_categories
WHERE user_uuid=$1
ORDER BY name`

	rows, err := r.pool.Query(context.Background(), query, userUUID)
	if err != nil {
		return []feed.Category{}, err
	}
	defer rows.Close()

	var dbCategories []DBCategory
	if err := pgxscan.ScanAll(&dbCategories, rows); err != nil {
		return []feed.Category{}, err
	}

	categories := make([]feed.Category, len(dbCategories))
	for i, dbCategory := range dbCategories {
		categories[i] = feed.Category{
			UUID: dbCategory.UUID,
			Name: dbCategory.Name,
			Slug: dbCategory.Slug,
		}
	}

	return categories, nil
}

func (r *Repository) FeedEntryCreateMany(entries []feed.Entry) (int64, error) {
	query := `
	INSERT INTO feed_entries(
		uid,
		feed_uuid,
		url,
		title,
		published_at,
		updated_at
	)
	VALUES(
		@uid,
		@feed_uuid,
		@url,
		@title,
		@published_at,
		@updated_at
	)`

	batch := &pgx.Batch{}

	for _, entry := range entries {
		args := pgx.NamedArgs{
			"uid":          entry.UID,
			"feed_uuid":    entry.FeedUUID,
			"url":          entry.URL,
			"title":        entry.Title,
			"published_at": entry.PublishedAt,
			"updated_at":   entry.UpdatedAt,
		}

		batch.Queue(query, args)
	}

	ctx := context.Background()

	batchResults := r.pool.SendBatch(ctx, batch)
	defer func() {
		if err := batchResults.Close(); err != nil {
			log.Error().
				Err(err).
				Str("domain", "feeds").
				Str("operation", "entries_create_many").
				Msg("failed to close batch results")
		}
	}()

	var rowsAffected int64

	for i := 0; i < len(entries); i++ {
		commandTag, qerr := batchResults.Exec()
		if qerr != nil {
			return 0, qerr
		}

		rowsAffected += commandTag.RowsAffected()
	}

	return rowsAffected, nil
}

func (r *Repository) FeedEntryGetN(feedUUID string, n uint) ([]feed.Entry, error) {
	query := `
	SELECT feed_uuid, uid, url, title, published_at, updated_at
	FROM  feed_entries
	WHERE feed_uuid=$1
	ORDER BY published_at DESC
	LIMIT $2`

	rows, err := r.pool.Query(context.Background(), query, feedUUID, n)
	if err != nil {
		return []feed.Entry{}, err
	}
	defer rows.Close()

	dbEntries := []DBEntry{}

	if err := pgxscan.ScanAll(&dbEntries, rows); err != nil {
		return []feed.Entry{}, err
	}

	entries := make([]feed.Entry, len(dbEntries))

	for i, dbEntry := range dbEntries {
		entries[i] = dbEntry.asEntry()
	}

	return entries, nil
}

func (r *Repository) FeedGetSubscriptionsByCategories(userUUID string) ([]fquerying.Category, error) {
	dbCategories, err := r.feedGetCategories(userUUID)
	if err != nil {
		return []fquerying.Category{}, err
	}

	categories := make([]fquerying.Category, len(dbCategories))

	for i, dbCategory := range dbCategories {
		dbFeeds, err := r.feedGetSubscriptionsByCategory(userUUID, dbCategory.UUID)
		if err != nil {
			return []fquerying.Category{}, err
		}

		var unread uint
		subscribedFeeds := make([]fquerying.SubscribedFeed, len(dbFeeds))

		for j, dbFeed := range dbFeeds {
			subscribedFeeds[j] = dbFeed.asSubscribedFeed()
			unread += dbFeed.Unread
		}

		category := fquerying.Category{
			Category: feed.Category{
				UUID: dbCategory.UUID,
				Name: dbCategory.Name,
				Slug: dbCategory.Slug,
			},
			Unread:          unread,
			SubscribedFeeds: subscribedFeeds,
		}

		categories[i] = category
	}

	return categories, nil
}

func (r *Repository) FeedGetEntriesByPage(userUUID string) ([]feed.Entry, error) {
	query := `
SELECT fe.url, fe.title, fe.published_at
FROM feed_entries fe
JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
WHERE user_uuid=$1
ORDER BY fe.published_at DESC`

	rows, err := r.pool.Query(context.Background(), query, userUUID)
	if err != nil {
		return []feed.Entry{}, err
	}
	defer rows.Close()

	dbEntries := []DBEntry{}

	if err := pgxscan.ScanAll(&dbEntries, rows); err != nil {
		return []feed.Entry{}, err
	}

	entries := make([]feed.Entry, len(dbEntries))

	for i, dbEntry := range dbEntries {
		entries[i] = dbEntry.asEntry()
	}

	return entries, nil
}

func (r *Repository) FeedIsSubscriptionRegistered(userUUID string, feedUUID string) (bool, error) {
	return r.rowExistsByQuery(
		"SELECT 1 FROM feed_subscriptions WHERE user_uuid=$1 AND feed_uuid=$2",
		userUUID,
		feedUUID,
	)
}

func (r *Repository) FeedSubscriptionCreate(s feed.Subscription) error {
	query := `
	INSERT INTO feed_subscriptions(
		uuid,
		feed_uuid,
		category_uuid,
		user_uuid,
		created_at,
		updated_at
	)
	VALUES(
		@uuid,
		@feed_uuid,
		@category_uuid,
		@user_uuid,
		@created_at,
		@updated_at
	)`

	args := pgx.NamedArgs{
		"uuid":          s.UUID,
		"feed_uuid":     s.FeedUUID,
		"category_uuid": s.CategoryUUID,
		"user_uuid":     s.UserUUID,
		"created_at":    s.CreatedAt,
		"updated_at":    s.UpdatedAt,
	}

	ctx := context.Background()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer r.rollback(ctx, tx, "feeds", "create subscription")

	_, err = tx.Exec(ctx, query, args)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
