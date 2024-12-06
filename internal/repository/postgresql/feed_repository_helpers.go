// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package postgresql

import (
	"context"
	"errors"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	feedquerying "github.com/virtualtam/sparklemuffin/pkg/feed/querying"
)

func (r *Repository) feedGetQuery(query string, queryParams ...any) (feed.Feed, error) {
	rows, err := r.pool.Query(context.Background(), query, queryParams...)
	if err != nil {
		return feed.Feed{}, err
	}
	defer rows.Close()

	dbFeed := &DBFeed{}
	err = pgxscan.ScanOne(dbFeed, rows)

	if errors.Is(err, pgx.ErrNoRows) {
		return feed.Feed{}, feed.ErrFeedNotFound
	}
	if err != nil {
		return feed.Feed{}, err
	}

	return dbFeed.asFeed(), nil
}

func (r *Repository) feedGetManyQuery(query string, queryParams ...any) ([]feed.Feed, error) {
	rows, err := r.pool.Query(context.Background(), query, queryParams...)
	if err != nil {
		return []feed.Feed{}, err
	}
	defer rows.Close()

	dbFeeds := []DBFeed{}

	if err := pgxscan.ScanAll(&dbFeeds, rows); err != nil {
		return []feed.Feed{}, err
	}

	feeds := make([]feed.Feed, len(dbFeeds))

	for i, dbFeed := range dbFeeds {
		feeds[i] = dbFeed.asFeed()
	}

	return feeds, nil
}

func (r *Repository) feedGetAllByCategory(userUUID string, categoryUUID string) ([]feed.Feed, error) {
	query := `
SELECT f.feed_url, f.title, f.slug
FROM feed_subscriptions fs
JOIN feed_feeds f ON f.uuid=fs.feed_uuid
WHERE fs.user_uuid=$1
AND   fs.category_uuid=$2
ORDER BY f.title`

	return r.feedGetManyQuery(query, userUUID, categoryUUID)
}

func (r *Repository) feedCategoryGetQuery(query string, queryParams ...any) (feed.Category, error) {
	rows, err := r.pool.Query(context.Background(), query, queryParams...)
	if err != nil {
		return feed.Category{}, err
	}
	defer rows.Close()

	dbCategory := &DBCategory{}
	err = pgxscan.ScanOne(dbCategory, rows)

	if errors.Is(err, pgx.ErrNoRows) {
		return feed.Category{}, feed.ErrCategoryNotFound
	}
	if err != nil {
		return feed.Category{}, err
	}

	return dbCategory.asCategory(), nil
}

func (r *Repository) feedGetCategories(userUUID string) ([]DBCategory, error) {
	query := `
	SELECT uuid, name, slug
	FROM feed_categories
	WHERE user_uuid=$1
	ORDER BY name`

	rows, err := r.pool.Query(context.Background(), query, userUUID)
	if err != nil {
		return []DBCategory{}, fmt.Errorf("failed to retrieve categories: %w", err)
	}
	defer rows.Close()

	var dbCategories []DBCategory
	if err := pgxscan.ScanAll(&dbCategories, rows); err != nil {
		return []DBCategory{}, fmt.Errorf("failed to scan category: %w", err)
	}

	return dbCategories, nil
}

func (r *Repository) feedEntryUpsertMany(operation string, onConflictStmt string, entries []feed.Entry) (int64, error) {
	insertQuery := `
	INSERT INTO feed_entries(
		uid,
		feed_uuid,
		url,
		title,
		summary,
		textrank_terms,
		fulltextsearch_tsv,
		published_at,
		updated_at
	)
	VALUES(
		@uid,
		@feed_uuid,
		@url,
		@title,
		@summary,
		@textrank_terms,
		to_tsvector(@fulltextsearch_string),
		@published_at,
		@updated_at
	)`

	query := insertQuery + onConflictStmt

	batch := &pgx.Batch{}

	for _, entry := range entries {
		fullTextSearchString := feedEntryToFullTextSearchString(entry)

		args := pgx.NamedArgs{
			"uid":                   entry.UID,
			"feed_uuid":             entry.FeedUUID,
			"url":                   entry.URL,
			"title":                 entry.Title,
			"summary":               entry.Summary,
			"textrank_terms":        entry.TextRankTerms,
			"fulltextsearch_string": fullTextSearchString,
			"published_at":          entry.PublishedAt,
			"updated_at":            entry.UpdatedAt,
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
				Str("operation", operation).
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

func (r *Repository) feedSubscriptionEntryGetN(query string, queryParams ...any) ([]feedquerying.SubscribedFeedEntry, error) {
	rows, err := r.pool.Query(context.Background(), query, queryParams...)
	if err != nil {
		return []feedquerying.SubscribedFeedEntry{}, err
	}
	defer rows.Close()

	dbQueryingEntries := []DBQueryingSubscribedFeedEntry{}

	if err := pgxscan.ScanAll(&dbQueryingEntries, rows); err != nil {
		return []feedquerying.SubscribedFeedEntry{}, err
	}

	queryingEntries := make([]feedquerying.SubscribedFeedEntry, len(dbQueryingEntries))

	for i, dbQueryingEntry := range dbQueryingEntries {
		queryingEntries[i] = dbQueryingEntry.asQueryingSubscribedFeedEntry()
	}

	return queryingEntries, nil
}

func (r *Repository) feedGetSubscriptionsByCategory(userUUID string, categoryUUID string) ([]DBSubscribedFeed, error) {
	query := `
SELECT
    f.feed_url,
    f.title,
    f.slug,
    fs.alias,
    COUNT(NULLIF(COALESCE(fem.read, FALSE) = TRUE, TRUE)) AS unread
FROM feed_subscriptions fs
JOIN feed_feeds f ON f.uuid = fs.feed_uuid
JOIN feed_entries fe ON fe.feed_uuid = fs.feed_uuid
LEFT JOIN feed_entries_metadata fem ON fem.entry_uid = fe.uid
WHERE
    fs.user_uuid = $1
    AND fs.category_uuid = $2
GROUP BY f.feed_url, f.title, f.slug, fs.alias
ORDER BY
    CASE
        WHEN fs.alias != '' THEN fs.alias
        ELSE f.title
    END`

	rows, err := r.pool.Query(context.Background(), query, userUUID, categoryUUID)
	if err != nil {
		return []DBSubscribedFeed{}, err
	}
	defer rows.Close()

	dbFeeds := []DBSubscribedFeed{}

	if err := pgxscan.ScanAll(&dbFeeds, rows); err != nil {
		return []DBSubscribedFeed{}, err
	}

	return dbFeeds, nil
}

func (r *Repository) feedSubscriptionGetQuery(query string, queryParams ...any) (feed.Subscription, error) {
	rows, err := r.pool.Query(context.Background(), query, queryParams...)
	if err != nil {
		return feed.Subscription{}, err
	}
	defer rows.Close()

	dbSubscription := &DBSubscription{}
	err = pgxscan.ScanOne(dbSubscription, rows)

	if errors.Is(err, pgx.ErrNoRows) {
		return feed.Subscription{}, feed.ErrSubscriptionNotFound
	}
	if err != nil {
		return feed.Subscription{}, err
	}

	return dbSubscription.asSubscription(), nil
}

func (r *Repository) feedSubscriptionTitleGetQuery(query string, queryParams ...any) (feedquerying.SubscriptionTitle, error) {
	rows, err := r.pool.Query(context.Background(), query, queryParams...)
	if err != nil {
		return feedquerying.SubscriptionTitle{}, err
	}
	defer rows.Close()

	dbSubscriptionTitle := &DBSubscriptionTitle{}
	err = pgxscan.ScanOne(dbSubscriptionTitle, rows)

	if errors.Is(err, pgx.ErrNoRows) {
		return feedquerying.SubscriptionTitle{}, feed.ErrSubscriptionNotFound
	}
	if err != nil {
		return feedquerying.SubscriptionTitle{}, err
	}

	return dbSubscriptionTitle.asSubscriptionTitle(), nil
}

func (r *Repository) feedGetSubscriptionTitlesByCategory(userUUID string, categoryUUID string) ([]DBSubscriptionTitle, error) {
	query := `
SELECT
    fs.uuid,
    fs.alias,
    f.title,
    f.description
FROM feed_subscriptions fs
JOIN feed_feeds f ON f.uuid = fs.feed_uuid
WHERE
    fs.user_uuid = $1
    AND fs.category_uuid = $2
ORDER BY
    CASE
        WHEN fs.alias != '' THEN fs.alias
        ELSE f.title
    END`

	rows, err := r.pool.Query(context.Background(), query, userUUID, categoryUUID)
	if err != nil {
		return []DBSubscriptionTitle{}, err
	}
	defer rows.Close()

	dbSubscriptionTitles := []DBSubscriptionTitle{}

	if err := pgxscan.ScanAll(&dbSubscriptionTitles, rows); err != nil {
		return []DBSubscriptionTitle{}, err
	}

	return dbSubscriptionTitles, nil
}
