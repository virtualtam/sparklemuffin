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

func (r *Repository) FeedAdd(f feed.Feed) error {
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

	return r.add("feeds", "FeedAdd", query, args)
}

func (r *Repository) FeedGetBySlug(feedSlug string) (feed.Feed, error) {
	query := `
SELECT uuid, url, title, slug
FROM feed_feeds
WHERE slug=$1`

	return r.feedGetQuery(query, feedSlug)
}

func (r *Repository) FeedGetByURL(feedURL string) (feed.Feed, error) {
	query := `
SELECT uuid, feed_url, title, slug
FROM feed_feeds
WHERE feed_url=$1`

	return r.feedGetQuery(query, feedURL)
}

func (r *Repository) FeedGetByUUID(feedUUID string) (feed.Feed, error) {
	query := `
SELECT uuid, url, title, slug
FROM feed_feeds
WHERE uuid=$1`

	return r.feedGetQuery(query, feedUUID)
}

func (r *Repository) FeedCategoryAdd(c feed.Category) error {
	query := `
	INSERT INTO feed_categories(
		uuid,
		user_uuid,
		name,
		slug,
		created_at,
		updated_at
	)
	VALUES(
		@uuid,
		@user_uuid,
		@name,
		@slug,
		@created_at,
		@updated_at
	)`

	args := pgx.NamedArgs{
		"uuid":       c.UUID,
		"user_uuid":  c.UserUUID,
		"name":       c.Name,
		"slug":       c.Slug,
		"created_at": c.CreatedAt,
		"updated_at": c.UpdatedAt,
	}

	return r.add("feeds", "FeedCategoryAdd", query, args)
}

func (r *Repository) FeedCategoryGetBySlug(userUUID string, slug string) (feed.Category, error) {
	query := `
	SELECT uuid, user_uuid, name, slug, created_at, updated_at
	FROM feed_categories
	WHERE user_uuid=$1
	AND slug=$2`

	return r.feedCategoryGetQuery(query, userUUID, slug)
}

func (r *Repository) FeedCategoryGetByUUID(userUUID string, categoryUUID string) (feed.Category, error) {
	query := `
	SELECT uuid, user_uuid, name, slug, created_at, updated_at
	FROM feed_categories
	WHERE user_uuid=$1
	AND uuid=$2`

	return r.feedCategoryGetQuery(query, userUUID, categoryUUID)
}

func (r *Repository) FeedCategoryGetMany(userUUID string) ([]feed.Category, error) {
	query := `
SELECT uuid, user_uuid, name, slug
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
			UUID:     dbCategory.UUID,
			UserUUID: dbCategory.UserUUID,
			Name:     dbCategory.Name,
			Slug:     dbCategory.Slug,
		}
	}

	return categories, nil
}

func (r *Repository) FeedCategoryNameAndSlugAreRegistered(userUUID string, name string, slug string) (bool, error) {
	return r.rowExistsByQuery(
		"SELECT 1 FROM feed_categories WHERE user_uuid=$1 AND (name=$2 OR slug=$3)",
		userUUID,
		name,
		slug,
	)
}

func (r *Repository) FeedCategoryNameAndSlugAreRegisteredToAnotherCategory(userUUID string, categoryUUID string, name string, slug string) (bool, error) {
	return r.rowExistsByQuery(
		"SELECT 1 FROM feed_categories WHERE user_uuid = $1 AND uuid != $2 AND (name = $3 OR slug = $4)",
		userUUID,
		categoryUUID,
		name,
		slug,
	)
}

func (r *Repository) FeedCategoryUpdate(c feed.Category) error {
	query := `
	UPDATE feed_categories
	SET
		name=@name,
		slug=@slug,
		updated_at=@updated_at
	WHERE user_uuid=@user_uuid
	AND uuid=@uuid`

	args := pgx.NamedArgs{
		"user_uuid":  c.UserUUID,
		"uuid":       c.UUID,
		"name":       c.Name,
		"slug":       c.Slug,
		"updated_at": c.UpdatedAt,
	}

	ctx := context.Background()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer r.rollback(ctx, tx, "feeds", "FeedCategoryUpdate")

	_, err = tx.Exec(ctx, query, args)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) FeedEntryAddMany(entries []feed.Entry) (int64, error) {
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
				Str("operation", "FeedEntryAddMany").
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

func (r *Repository) FeedSubscriptionCategoryGetAll(userUUID string) ([]fquerying.SubscriptionCategory, error) {
	dbCategories, err := r.feedGetCategories(userUUID)
	if err != nil {
		return []fquerying.SubscriptionCategory{}, err
	}

	categories := make([]fquerying.SubscriptionCategory, len(dbCategories))

	for i, dbCategory := range dbCategories {
		dbFeeds, err := r.feedGetSubscriptionsByCategory(userUUID, dbCategory.UUID)
		if err != nil {
			return []fquerying.SubscriptionCategory{}, err
		}

		var unread uint
		subscribedFeeds := make([]fquerying.SubscribedFeed, len(dbFeeds))

		for j, dbFeed := range dbFeeds {
			subscribedFeeds[j] = dbFeed.asSubscribedFeed()
			unread += dbFeed.Unread
		}

		category := fquerying.SubscriptionCategory{
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

func (r *Repository) FeedEntryGetCount(userUUID string) (uint, error) {
	query := `
	SELECT COUNT(*)
	FROM feed_entries fe
	JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
	WHERE fs.user_uuid=$1`

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

func (r *Repository) FeedEntryGetCountByCategory(userUUID string, categoryUUID string) (uint, error) {
	query := `
	SELECT COUNT(*)
	FROM feed_entries fe
	JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
	WHERE fs.user_uuid=$1
	AND   fs.category_uuid=$2`

	var count uint

	err := r.pool.QueryRow(
		context.Background(),
		query,
		userUUID,
		categoryUUID,
	).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *Repository) FeedEntryGetCountBySubscription(userUUID string, subscriptionUUID string) (uint, error) {
	query := `
	SELECT COUNT(*)
	FROM feed_entries fe
	JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
	WHERE user_uuid=$1
	AND   fs.uuid=$2`

	var count uint

	err := r.pool.QueryRow(
		context.Background(),
		query,
		userUUID,
		subscriptionUUID,
	).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *Repository) FeedSubscriptionEntryGetN(userUUID string, n uint, offset uint) ([]fquerying.SubscriptionEntry, error) {
	query := `
	SELECT fe.url, fe.title, fe.published_at, FALSE AS "read"
	FROM feed_entries fe
	JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
	WHERE fs.user_uuid=$1
	ORDER BY fe.published_at DESC
	LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(context.Background(), query, userUUID, n, offset)
	if err != nil {
		return []fquerying.SubscriptionEntry{}, err
	}
	defer rows.Close()

	dbQueryingEntries := []DBQueryingEntry{}

	if err := pgxscan.ScanAll(&dbQueryingEntries, rows); err != nil {
		return []fquerying.SubscriptionEntry{}, err
	}

	queryingEntries := make([]fquerying.SubscriptionEntry, len(dbQueryingEntries))

	for i, dbQueryingEntry := range dbQueryingEntries {
		queryingEntries[i] = dbQueryingEntry.asQueryingEntry()
	}

	return queryingEntries, nil
}

func (r *Repository) FeedSubscriptionEntryGetNByCategory(userUUID string, categoryUUID string, n uint, offset uint) ([]fquerying.SubscriptionEntry, error) {
	query := `
	SELECT fe.url, fe.title, fe.published_at, FALSE AS "read"
	FROM feed_entries fe
	JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
	WHERE fs.user_uuid=$1
	AND   fs.category_uuid=$2
	ORDER BY fe.published_at DESC
	LIMIT $3 OFFSET $4`

	rows, err := r.pool.Query(context.Background(), query, userUUID, categoryUUID, n, offset)
	if err != nil {
		return []fquerying.SubscriptionEntry{}, err
	}
	defer rows.Close()

	dbQueryingEntries := []DBQueryingEntry{}

	if err := pgxscan.ScanAll(&dbQueryingEntries, rows); err != nil {
		return []fquerying.SubscriptionEntry{}, err
	}

	queryingEntries := make([]fquerying.SubscriptionEntry, len(dbQueryingEntries))

	for i, dbQueryingEntry := range dbQueryingEntries {
		queryingEntries[i] = dbQueryingEntry.asQueryingEntry()
	}

	return queryingEntries, nil
}

func (r *Repository) FeedSubscriptionEntryGetNBySubscription(userUUID string, subscriptionUUID string, n uint, offset uint) ([]fquerying.SubscriptionEntry, error) {
	query := `
	SELECT fe.url, fe.title, fe.published_at, FALSE as "read"
	FROM feed_entries fe
	JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
	WHERE fs.user_uuid=$1
	AND   fs.uuid=$2
	ORDER BY fe.published_at DESC
	LIMIT $3 OFFSET $4`

	rows, err := r.pool.Query(context.Background(), query, userUUID, subscriptionUUID, n, offset)
	if err != nil {
		return []fquerying.SubscriptionEntry{}, err
	}
	defer rows.Close()

	dbQueryingEntries := []DBQueryingEntry{}

	if err := pgxscan.ScanAll(&dbQueryingEntries, rows); err != nil {
		return []fquerying.SubscriptionEntry{}, err
	}

	queryingEntries := make([]fquerying.SubscriptionEntry, len(dbQueryingEntries))

	for i, dbQueryingEntry := range dbQueryingEntries {
		queryingEntries[i] = dbQueryingEntry.asQueryingEntry()
	}

	return queryingEntries, nil
}

func (r *Repository) FeedSubscriptionIsRegistered(userUUID string, feedUUID string) (bool, error) {
	return r.rowExistsByQuery(
		"SELECT 1 FROM feed_subscriptions WHERE user_uuid=$1 AND feed_uuid=$2",
		userUUID,
		feedUUID,
	)
}

func (r *Repository) FeedSubscriptionAdd(s feed.Subscription) error {
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

	return r.add("feeds", "FeedSubscriptionAdd", query, args)
}

func (r *Repository) FeedSubscriptionGetByFeed(userUUID string, feedUUID string) (feed.Subscription, error) {
	query := `
	SELECT uuid, category_uuid, feed_uuid, user_uuid, created_at, updated_at
	FROM feed_subscriptions
	WHERE feed_uuid=$1`

	return r.feedSubscriptionGetQuery(query, feedUUID)
}
