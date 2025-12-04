// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package pgfeed

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

func (r *Repository) feedGetQuery(ctx context.Context, query string, queryParams ...any) (feed.Feed, error) {
	rows, err := r.Pool.Query(ctx, query, queryParams...)
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

func (r *Repository) feedGetManyQuery(ctx context.Context, query string, queryParams ...any) ([]feed.Feed, error) {
	rows, err := r.Pool.Query(ctx, query, queryParams...)
	if err != nil {
		return []feed.Feed{}, err
	}
	defer rows.Close()

	var dbFeeds []DBFeed

	if err := pgxscan.ScanAll(&dbFeeds, rows); err != nil {
		return []feed.Feed{}, err
	}

	feeds := make([]feed.Feed, len(dbFeeds))

	for i, dbFeed := range dbFeeds {
		feeds[i] = dbFeed.asFeed()
	}

	return feeds, nil
}

func (r *Repository) feedGetAllByCategory(ctx context.Context, userUUID string, categoryUUID string) ([]feed.Feed, error) {
	query := `
SELECT f.feed_url, f.title, f.slug
FROM feed_subscriptions fs
JOIN feed_feeds f ON f.uuid=fs.feed_uuid
WHERE fs.user_uuid=$1
AND   fs.category_uuid=$2
ORDER BY f.title`

	return r.feedGetManyQuery(ctx, query, userUUID, categoryUUID)
}

func (r *Repository) feedCategoryGetQuery(ctx context.Context, query string, queryParams ...any) (feed.Category, error) {
	rows, err := r.Pool.Query(ctx, query, queryParams...)
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

func (r *Repository) feedGetCategories(ctx context.Context, userUUID string) ([]DBCategory, error) {
	query := `
	SELECT uuid, name, slug
	FROM feed_categories
	WHERE user_uuid=$1
	ORDER BY name`

	rows, err := r.Pool.Query(ctx, query, userUUID)
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

func (r *Repository) feedEntryUpsertMany(ctx context.Context, operation string, onConflictStmt string, entries []feed.Entry) (int64, error) {
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
		TO_TSVECTOR(@fulltextsearch_string),
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

	batchResults := r.Pool.SendBatch(ctx, batch)
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

	for range entries {
		commandTag, qerr := batchResults.Exec()
		if qerr != nil {
			return 0, qerr
		}

		rowsAffected += commandTag.RowsAffected()
	}

	return rowsAffected, nil
}

func (r *Repository) feedEntryGetCount(ctx context.Context, and string, showEntries feed.EntryVisibility, args pgx.NamedArgs) (uint, error) {
	const baseQuery = `
		SELECT COUNT(*)
		FROM feed_entries fe
		JOIN feed_feeds f ON f.uuid = fe.feed_uuid
		JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
		LEFT JOIN feed_entries_metadata fem ON fem.entry_uid = fe.uid
		WHERE fs.user_uuid=@user_uuid`

	query := fmt.Sprintf("%s\n%s", baseQuery, and)

	switch showEntries {
	case feed.EntryVisibilityRead:
		query += " AND fem.read = TRUE"
	case feed.EntryVisibilityUnread:
		query += " AND COALESCE(fem.read, FALSE) = FALSE"
	}

	var count uint

	err := r.Pool.QueryRow(
		ctx,
		query,
		args,
	).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *Repository) feedSubscriptionEntryGetN(ctx context.Context, where string, showEntries feed.EntryVisibility, args pgx.NamedArgs) ([]feedquerying.SubscribedFeedEntry, error) {
	const baseQuery = `
		SELECT
			fe.uid,
			fe.url,
			fe.title,
			fe.summary,
			fe.published_at,
			fe.updated_at,
			fs.alias AS subscription_alias,
			f.uuid AS feed_uuid,
			f.title AS feed_title,
			f.slug AS feed_slug,
			COALESCE(fem.read, FALSE) AS read
		FROM feed_entries fe
		LEFT JOIN feed_entries_metadata fem ON fem.entry_uid = fe.uid
		JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
		JOIN feed_feeds f ON f.uuid = fe.feed_uuid`

	query := fmt.Sprintf("%s\n%s", baseQuery, where)

	switch showEntries {
	case feed.EntryVisibilityRead:
		query += " AND read = TRUE"
	case feed.EntryVisibilityUnread:
		query += " AND COALESCE(fem.read, FALSE) = FALSE"
	}

	query += `
	ORDER BY fe.published_at DESC
	LIMIT @limit OFFSET @offset`

	rows, err := r.Pool.Query(ctx, query, args)
	if err != nil {
		return []feedquerying.SubscribedFeedEntry{}, err
	}
	defer rows.Close()

	var dbQueryingEntries []DBQueryingSubscribedFeedEntry

	if err := pgxscan.ScanAll(&dbQueryingEntries, rows); err != nil {
		return []feedquerying.SubscribedFeedEntry{}, err
	}

	queryingEntries := make([]feedquerying.SubscribedFeedEntry, len(dbQueryingEntries))

	for i, dbQueryingEntry := range dbQueryingEntries {
		queryingEntries[i] = dbQueryingEntry.asQueryingSubscribedFeedEntry()
	}

	return queryingEntries, nil
}

func (r *Repository) feedGetSubscriptionsByCategory(ctx context.Context, userUUID string, categoryUUID string) ([]DBSubscribedFeed, error) {
	query := `
SELECT
    f.feed_url,
    f.title,
    f.slug,
	f.created_at,
	f.updated_at,
	f.fetched_at,
    fs.alias,
    COUNT(NULLIF(COALESCE(fem.read, FALSE) = TRUE, TRUE)) AS unread
FROM feed_subscriptions fs
JOIN feed_feeds f ON f.uuid = fs.feed_uuid
JOIN feed_entries fe ON fe.feed_uuid = fs.feed_uuid
LEFT JOIN feed_entries_metadata fem ON fem.entry_uid = fe.uid
WHERE
    fs.user_uuid = $1
    AND fs.category_uuid = $2
GROUP BY f.feed_url, f.title, f.slug, f.created_at, f.updated_at, f.fetched_at, fs.alias
ORDER BY
    CASE
        WHEN fs.alias != '' THEN fs.alias
        ELSE f.title
    END`

	rows, err := r.Pool.Query(ctx, query, userUUID, categoryUUID)
	if err != nil {
		return []DBSubscribedFeed{}, err
	}
	defer rows.Close()

	var dbFeeds []DBSubscribedFeed

	if err := pgxscan.ScanAll(&dbFeeds, rows); err != nil {
		return []DBSubscribedFeed{}, err
	}

	return dbFeeds, nil
}

func (r *Repository) feedSubscriptionGetQuery(ctx context.Context, query string, queryParams ...any) (feed.Subscription, error) {
	rows, err := r.Pool.Query(ctx, query, queryParams...)
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

func (r *Repository) feedSubscriptionTitleGetQuery(ctx context.Context, query string, queryParams ...any) (feedquerying.Subscription, error) {
	rows, err := r.Pool.Query(ctx, query, queryParams...)
	if err != nil {
		return feedquerying.Subscription{}, err
	}
	defer rows.Close()

	dbSubscriptionTitle := &DBQueryingSubscription{}
	err = pgxscan.ScanOne(dbSubscriptionTitle, rows)

	if errors.Is(err, pgx.ErrNoRows) {
		return feedquerying.Subscription{}, feed.ErrSubscriptionNotFound
	}
	if err != nil {
		return feedquerying.Subscription{}, err
	}

	return dbSubscriptionTitle.asQueryingSubscription(), nil
}

func (r *Repository) feedGetSubscriptionTitlesByCategory(ctx context.Context, userUUID string, categoryUUID string) ([]DBQueryingSubscription, error) {
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

	rows, err := r.Pool.Query(ctx, query, userUUID, categoryUUID)
	if err != nil {
		return []DBQueryingSubscription{}, err
	}
	defer rows.Close()

	var dbSubscriptionTitles []DBQueryingSubscription

	if err := pgxscan.ScanAll(&dbSubscriptionTitles, rows); err != nil {
		return []DBQueryingSubscription{}, err
	}

	return dbSubscriptionTitles, nil
}
