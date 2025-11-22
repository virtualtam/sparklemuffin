// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package pgfeed

import (
	"context"
	"errors"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgbase"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	feedexporting "github.com/virtualtam/sparklemuffin/pkg/feed/exporting"
	feedquerying "github.com/virtualtam/sparklemuffin/pkg/feed/querying"
	feedsynchronizing "github.com/virtualtam/sparklemuffin/pkg/feed/synchronizing"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

var _ feed.Repository = &Repository{}
var _ feedexporting.Repository = &Repository{}
var _ feedquerying.Repository = &Repository{}
var _ feedsynchronizing.Repository = &Repository{}

type Repository struct {
	*pgbase.Repository
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		Repository: pgbase.NewRepository(pool),
	}
}

const (
	domain = "feeds"
)

func (r *Repository) FeedCreate(ctx context.Context, f feed.Feed) error {
	query := `
	INSERT INTO feed_feeds(
		uuid,
		feed_url,
		title,
		description,
		slug,
		fulltextsearch_tsv,
		etag,
		hash_xxhash64,
		last_modified,
		created_at,
		updated_at,
		fetched_at
	)
	VALUES(
		@uuid,
		@feed_url,
		@title,
		@description,
		@slug,
		TO_TSVECTOR(@fulltextsearch_string),
		@etag,
		@hash_xxhash64,
		@last_modified,
		@created_at,
		@updated_at,
		@fetched_at
	)`

	fullTextSearchString := feedToFullTextSearchString(f)

	args := pgx.NamedArgs{
		"uuid":                  f.UUID,
		"feed_url":              f.FeedURL,
		"title":                 f.Title,
		"description":           f.Description,
		"slug":                  f.Slug,
		"fulltextsearch_string": fullTextSearchString,
		"etag":                  f.ETag,
		"hash_xxhash64":         int64(f.Hash), // uint64 -> int64 (BIGINT)
		"last_modified":         f.LastModified,
		"created_at":            f.CreatedAt,
		"updated_at":            f.UpdatedAt,
		"fetched_at":            f.FetchedAt,
	}

	return r.QueryTx(ctx, domain, "FeedCreate", query, args)
}

func (r *Repository) FeedGetBySlug(ctx context.Context, feedSlug string) (feed.Feed, error) {
	query := `
	SELECT uuid, feed_url, title, description, slug, etag, last_modified, hash_xxhash64, created_at, updated_at, fetched_at
	FROM feed_feeds
	WHERE slug=$1`

	return r.feedGetQuery(ctx, query, feedSlug)
}

func (r *Repository) FeedGetByURL(ctx context.Context, feedURL string) (feed.Feed, error) {
	query := `
	SELECT uuid, feed_url, title, description, slug, etag, last_modified, hash_xxhash64, created_at, updated_at, fetched_at
	FROM feed_feeds
	WHERE feed_url=$1`

	return r.feedGetQuery(ctx, query, feedURL)
}

func (r *Repository) FeedGetByUUID(ctx context.Context, feedUUID string) (feed.Feed, error) {
	query := `
	SELECT uuid, feed_url, title, description, slug, etag, last_modified, hash_xxhash64, created_at, updated_at, fetched_at
	FROM feed_feeds
	WHERE uuid=$1`

	return r.feedGetQuery(ctx, query, feedUUID)
}

func (r *Repository) FeedGetNByLastSynchronizationTime(ctx context.Context, n uint, before time.Time) ([]feed.Feed, error) {
	query := `
	SELECT f.uuid, f.feed_url, f.title, f.description, f.slug, f.etag, f.last_modified, f.hash_xxhash64, f.created_at, f.updated_at, f.fetched_at
	FROM feed_feeds f
	INNER JOIN feed_subscriptions fs ON f.uuid = fs.feed_uuid
	WHERE fetched_at < $1
	OR    fetched_at IS NULL
	LIMIT $2`

	return r.feedGetManyQuery(ctx, query, before, n)
}

func (r *Repository) FeedUpdateFetchMetadata(ctx context.Context, feedFetchMetadata feedsynchronizing.FeedFetchMetadata) error {
	query := `
	UPDATE feed_feeds
	SET
		etag=@etag,
		last_modified=@last_modified,
		updated_at=@updated_at,
		fetched_at=@fetched_at
	WHERE uuid=@uuid`

	args := pgx.NamedArgs{
		"uuid":          feedFetchMetadata.UUID,
		"etag":          feedFetchMetadata.ETag,
		"last_modified": feedFetchMetadata.LastModified,
		"updated_at":    feedFetchMetadata.UpdatedAt,
		"fetched_at":    feedFetchMetadata.FetchedAt,
	}

	return r.QueryTx(ctx, domain, "FeedUpdateFetchMetadata", query, args)
}

func (r *Repository) FeedUpdateMetadata(ctx context.Context, feedMetadata feedsynchronizing.FeedMetadata) error {
	query := `
	UPDATE feed_feeds
	SET
		title=@title,
		description=@description,
		hash_xxhash64=@hash_xxhash64,
		fulltextsearch_tsv=TO_TSVECTOR(@fulltextsearch_string),
		updated_at=@updated_at
	WHERE uuid=@uuid`

	fullTextSearchString := feedMetadataToFullTextSearchString(feedMetadata)

	args := pgx.NamedArgs{
		"uuid":                  feedMetadata.UUID,
		"title":                 feedMetadata.Title,
		"description":           feedMetadata.Description,
		"hash_xxhash64":         int64(feedMetadata.Hash), // uint64 -> int64 (BIGINT)
		"fulltextsearch_string": fullTextSearchString,
		"updated_at":            feedMetadata.UpdatedAt,
	}

	return r.QueryTx(ctx, domain, "FeedUpdateFetchMetadata", query, args)
}

func (r *Repository) FeedCategoryCreate(ctx context.Context, c feed.Category) error {
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

	return r.QueryTx(ctx, domain, "FeedCategoryCreate", query, args)
}

func (r *Repository) FeedCategoryDelete(ctx context.Context, userUUID string, categoryUUID string) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer r.Rollback(ctx, tx, domain, "FeedCategoryDelete")

	// 1. Delete the category (cascaded to subscriptions)
	commandTag, err := tx.Exec(
		ctx,
		"DELETE FROM feed_categories WHERE user_uuid=$1 AND uuid=$2",
		userUUID,
		categoryUUID,
	)
	if err != nil {
		return err
	}

	rowsAffected := commandTag.RowsAffected()

	if rowsAffected != 1 {
		return feed.ErrCategoryNotFound
	}

	// 2. Delete feeds with no remaining subscriptions
	_, err = tx.Exec(
		ctx,
		`
		DELETE FROM feed_feeds
		WHERE uuid NOT IN (
			SELECT feed_uuid
			FROM feed_subscriptions
		)`,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) FeedCategoryGetByName(ctx context.Context, userUUID string, name string) (feed.Category, error) {
	query := `
	SELECT uuid, user_uuid, name, slug, created_at, updated_at
	FROM feed_categories
	WHERE user_uuid=$1
	AND name=$2`

	return r.feedCategoryGetQuery(ctx, query, userUUID, name)
}

func (r *Repository) FeedCategoryGetBySlug(ctx context.Context, userUUID string, slug string) (feed.Category, error) {
	query := `
	SELECT uuid, user_uuid, name, slug, created_at, updated_at
	FROM feed_categories
	WHERE user_uuid=$1
	AND slug=$2`

	return r.feedCategoryGetQuery(ctx, query, userUUID, slug)
}

func (r *Repository) FeedCategoryGetByUUID(ctx context.Context, userUUID string, categoryUUID string) (feed.Category, error) {
	query := `
	SELECT uuid, user_uuid, name, slug, created_at, updated_at
	FROM feed_categories
	WHERE user_uuid=$1
	AND uuid=$2`

	return r.feedCategoryGetQuery(ctx, query, userUUID, categoryUUID)
}

func (r *Repository) FeedCategoryGetMany(ctx context.Context, userUUID string) ([]feed.Category, error) {
	query := `
SELECT uuid, user_uuid, name, slug
FROM feed_categories
WHERE user_uuid=$1
ORDER BY name`

	rows, err := r.Pool.Query(ctx, query, userUUID)
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

func (r *Repository) FeedCategoryNameAndSlugAreRegistered(ctx context.Context, userUUID string, name string, slug string) (bool, error) {
	return r.RowExistsByQuery(
		ctx,
		"SELECT 1 FROM feed_categories WHERE user_uuid=$1 AND (name=$2 OR slug=$3)",
		userUUID,
		name,
		slug,
	)
}

func (r *Repository) FeedCategoryNameAndSlugAreRegisteredToAnotherCategory(ctx context.Context, userUUID string, categoryUUID string, name string, slug string) (bool, error) {
	return r.RowExistsByQuery(
		ctx,
		"SELECT 1 FROM feed_categories WHERE user_uuid = $1 AND uuid != $2 AND (name = $3 OR slug = $4)",
		userUUID,
		categoryUUID,
		name,
		slug,
	)
}

func (r *Repository) FeedCategoryUpdate(ctx context.Context, c feed.Category) error {
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

	return r.QueryTx(ctx, domain, "FeedCategoryUpdate", query, args)
}

func (r *Repository) FeedEntryCreateMany(ctx context.Context, entries []feed.Entry) (int64, error) {
	return r.feedEntryUpsertMany(ctx, "FeedEntryCreateMany", "ON CONFLICT DO NOTHING", entries)
}

func (r *Repository) FeedEntryUpsertMany(ctx context.Context, entries []feed.Entry) (int64, error) {
	return r.feedEntryUpsertMany(
		ctx,
		"FeedEntryUpsertMany",
		`
		ON CONFLICT (feed_uuid, url) DO UPDATE
		SET
			title              = EXCLUDED.title,
			summary            = EXCLUDED.summary,
			textrank_terms     = EXCLUDED.textrank_terms,
			fulltextsearch_tsv = EXCLUDED.fulltextsearch_tsv,
			updated_at         = EXCLUDED.updated_at
		`,
		entries,
	)
}

func (r *Repository) FeedEntryGetCount(ctx context.Context, userUUID string, showEntries feed.EntryVisibility) (uint, error) {
	const (
		query = `
		SELECT COUNT(*)
		FROM feed_entries fe
		JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
		LEFT JOIN feed_entries_metadata fem ON fem.entry_uid = fe.uid
		WHERE fs.user_uuid=@user_uuid`
	)

	args := pgx.NamedArgs{
		"user_uuid": userUUID,
	}

	return r.feedEntryGetCount(ctx, query, showEntries, args)
}

func (r *Repository) FeedEntryGetCountByCategory(ctx context.Context, userUUID string, showEntries feed.EntryVisibility, categoryUUID string) (uint, error) {
	const (
		query = `
		SELECT COUNT(*)
		FROM feed_entries fe
		JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
		LEFT JOIN feed_entries_metadata fem ON fem.entry_uid = fe.uid
		WHERE fs.user_uuid=@user_uuid
		AND   fs.category_uuid=@category_uuid`
	)

	args := pgx.NamedArgs{
		"user_uuid":     userUUID,
		"category_uuid": categoryUUID,
	}

	return r.feedEntryGetCount(ctx, query, showEntries, args)
}

func (r *Repository) FeedEntryGetCountBySubscription(ctx context.Context, userUUID string, showEntries feed.EntryVisibility, subscriptionUUID string) (uint, error) {
	const (
		query = `
		SELECT COUNT(*)
		FROM feed_entries fe
		JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
		LEFT JOIN feed_entries_metadata fem ON fem.entry_uid = fe.uid
		WHERE fs.user_uuid=@user_uuid
		AND   fs.uuid=@subscription_uuid`
	)

	args := pgx.NamedArgs{
		"user_uuid":         userUUID,
		"subscription_uuid": subscriptionUUID,
	}

	return r.feedEntryGetCount(ctx, query, showEntries, args)
}

func (r *Repository) FeedEntryGetCountByQuery(ctx context.Context, userUUID string, showEntries feed.EntryVisibility, searchTerms string) (uint, error) {
	const (
		query = `
		SELECT COUNT(*)
		FROM feed_entries fe
		JOIN feed_feeds f ON f.uuid = fe.feed_uuid
		JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
		LEFT JOIN feed_entries_metadata fem ON fem.entry_uid = fe.uid
		WHERE fs.user_uuid=@user_uuid
		AND   (f.fulltextsearch_tsv || fe.fulltextsearch_tsv) @@ websearch_to_tsquery(@search_terms)`
	)

	args := pgx.NamedArgs{
		"user_uuid":    userUUID,
		"search_terms": pgbase.FullTextSearchReplacer.Replace(searchTerms),
	}

	return r.feedEntryGetCount(ctx, query, showEntries, args)
}

func (r *Repository) FeedEntryGetCountByCategoryAndQuery(ctx context.Context, userUUID string, showEntries feed.EntryVisibility, categoryUUID string, searchTerms string) (uint, error) {
	const (
		query = `
		SELECT COUNT(*)
		FROM feed_entries fe
		JOIN feed_feeds f ON f.uuid = fe.feed_uuid
		JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
		LEFT JOIN feed_entries_metadata fem ON fem.entry_uid = fe.uid
		WHERE fs.user_uuid=@user_uuid
		AND   fs.category_uuid=@category_uuid
		AND   (f.fulltextsearch_tsv || fe.fulltextsearch_tsv) @@ websearch_to_tsquery(@search_terms)`
	)

	args := pgx.NamedArgs{
		"user_uuid":     userUUID,
		"category_uuid": categoryUUID,
		"search_terms":  pgbase.FullTextSearchReplacer.Replace(searchTerms),
	}

	return r.feedEntryGetCount(ctx, query, showEntries, args)
}

func (r *Repository) FeedEntryGetCountBySubscriptionAndQuery(ctx context.Context, userUUID string, showEntries feed.EntryVisibility, subscriptionUUID string, searchTerms string) (uint, error) {
	const (
		query = `
		SELECT COUNT(*)
		FROM feed_entries fe
		JOIN feed_feeds f ON f.uuid = fe.feed_uuid
		JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
		LEFT JOIN feed_entries_metadata fem ON fem.entry_uid = fe.uid
		WHERE fs.user_uuid=@user_uuid
		AND   fs.uuid=@subscription_uuid
		AND   (f.fulltextsearch_tsv || fe.fulltextsearch_tsv) @@ websearch_to_tsquery(@search_terms)`
	)

	args := pgx.NamedArgs{
		"user_uuid":         userUUID,
		"subscription_uuid": subscriptionUUID,
		"search_terms":      pgbase.FullTextSearchReplacer.Replace(searchTerms),
	}

	return r.feedEntryGetCount(ctx, query, showEntries, args)
}

func (r *Repository) FeedEntryMarkAllAsRead(ctx context.Context, userUUID string) error {
	query := `
	INSERT INTO feed_entries_metadata(
		user_uuid,
		entry_uid,
		read
	)

	SELECT @user_uuid, fe.uid, TRUE
	FROM feed_entries fe
	JOIN feed_subscriptions fs ON fs.feed_uuid=fe.feed_uuid
	WHERE fs.user_uuid=@user_uuid

	ON CONFLICT(user_uuid, entry_uid) DO UPDATE SET read=TRUE
	`

	args := pgx.NamedArgs{
		"user_uuid": userUUID,
	}

	return r.QueryTx(ctx, domain, "FeedEntryMarkAllAsRead", query, args)
}

func (r *Repository) FeedEntryMarkAllAsReadByCategory(ctx context.Context, userUUID string, categoryUUID string) error {
	query := `
	INSERT INTO feed_entries_metadata(
		user_uuid,
		entry_uid,
		read
	)

	SELECT @user_uuid, fe.uid, TRUE
	FROM feed_entries fe
	JOIN feed_subscriptions fs ON fs.feed_uuid=fe.feed_uuid
	WHERE fs.user_uuid=@user_uuid
	AND   fs.category_uuid=@category_uuid

	ON CONFLICT(user_uuid, entry_uid) DO UPDATE SET read=TRUE
	`

	args := pgx.NamedArgs{
		"user_uuid":     userUUID,
		"category_uuid": categoryUUID,
	}

	return r.QueryTx(ctx, domain, "FeedEntryMarkAllAsReadByCategory", query, args)
}

func (r *Repository) FeedEntryMarkAllAsReadBySubscription(ctx context.Context, userUUID string, subscriptionUUID string) error {
	query := `
	INSERT INTO feed_entries_metadata(
		user_uuid,
		entry_uid,
		read
	)

	SELECT @user_uuid, fe.uid, TRUE
	FROM feed_entries fe
	JOIN feed_subscriptions fs ON fs.feed_uuid=fe.feed_uuid
	WHERE fs.user_uuid=@user_uuid
	AND   fs.uuid=@subscription_uuid

	ON CONFLICT(user_uuid, entry_uid) DO UPDATE SET read=TRUE
	`

	args := pgx.NamedArgs{
		"user_uuid":         userUUID,
		"subscription_uuid": subscriptionUUID,
	}

	return r.QueryTx(ctx, domain, "FeedEntryMarkAllAsReadBySubscription", query, args)
}

func (r *Repository) FeedEntryMetadataCreate(ctx context.Context, entryMetadata feed.EntryMetadata) error {
	query := `
	INSERT INTO feed_entries_metadata(
		user_uuid,
		entry_uid,
		read
	)
	VALUES(
		@user_uuid,
		@entry_uid,
		@read
	)
	`

	args := pgx.NamedArgs{
		"user_uuid": entryMetadata.UserUUID,
		"entry_uid": entryMetadata.EntryUID,
		"read":      entryMetadata.Read,
	}

	return r.QueryTx(ctx, domain, "FeedEntryMetadataCreate", query, args)
}

func (r *Repository) FeedEntryMetadataGetByUID(ctx context.Context, userUUID string, entryUID string) (feed.EntryMetadata, error) {
	query := `
	SELECT user_uuid, entry_uid, read
	FROM feed_entries_metadata
	WHERE user_uuid=$1
	AND   entry_uid=$2
	`

	rows, err := r.Pool.Query(ctx, query, userUUID, entryUID)
	if err != nil {
		return feed.EntryMetadata{}, err
	}
	defer rows.Close()

	dbEntryMetadata := &DBEntryMetadata{}
	err = pgxscan.ScanOne(dbEntryMetadata, rows)

	if errors.Is(err, pgx.ErrNoRows) {
		return feed.EntryMetadata{}, feed.ErrEntryMetadataNotFound
	}
	if err != nil {
		return feed.EntryMetadata{}, err
	}

	return dbEntryMetadata.asEntryMetadata(), nil
}

func (r *Repository) FeedEntryMetadataUpdate(ctx context.Context, entryMetadata feed.EntryMetadata) error {
	query := `
	UPDATE feed_entries_metadata
	SET read=@read
	WHERE user_uuid=@user_uuid
	AND   entry_uid=@entry_uid
	`

	args := pgx.NamedArgs{
		"user_uuid": entryMetadata.UserUUID,
		"entry_uid": entryMetadata.EntryUID,
		"read":      entryMetadata.Read,
	}

	return r.QueryTx(ctx, domain, "FeedEntryMetadataUpdate", query, args)
}

func (r *Repository) FeedPreferencesGetByUserUUID(ctx context.Context, userUUID string) (feed.Preferences, error) {
	const (
		query = `
		SELECT user_uuid, show_entries, show_entry_summaries, updated_at
		FROM feed_preferences
		WHERE user_uuid=$1`
	)

	rows, err := r.Pool.Query(ctx, query, userUUID)
	if err != nil {
		return feed.Preferences{}, err
	}
	defer rows.Close()

	dbPreferences := DBPreferences{}
	err = pgxscan.ScanOne(&dbPreferences, rows)
	if errors.Is(err, pgx.ErrNoRows) {
		return feed.Preferences{}, user.ErrNotFound
	}
	if err != nil {
		return feed.Preferences{}, err
	}

	return feed.Preferences{
		UserUUID:           dbPreferences.UserUUID,
		ShowEntries:        feed.EntryVisibility(dbPreferences.ShowEntries),
		ShowEntrySummaries: dbPreferences.ShowEntrySummaries,
		UpdatedAt:          dbPreferences.UpdatedAt,
	}, nil
}

func (r *Repository) FeedPreferencesUpdate(ctx context.Context, preferences feed.Preferences) error {
	const (
		query = `
		UPDATE feed_preferences
		SET
			show_entries=@show_entries,
			show_entry_summaries=@show_entry_summaries,
			updated_at=@updated_at
		WHERE user_uuid=@user_uuid`
	)

	args := pgx.NamedArgs{
		"user_uuid":            preferences.UserUUID,
		"show_entries":         preferences.ShowEntries,
		"show_entry_summaries": preferences.ShowEntrySummaries,
		"updated_at":           preferences.UpdatedAt,
	}

	return r.QueryTx(ctx, domain, "FeedPreferencesUpdate", query, args)
}

func (r *Repository) FeedCategorySubscriptionsGetAll(ctx context.Context, userUUID string) ([]feedexporting.CategorySubscriptions, error) {
	dbCategories, err := r.feedGetCategories(ctx, userUUID)
	if err != nil {
		return []feedexporting.CategorySubscriptions{}, err
	}

	categoriesSubscriptions := make([]feedexporting.CategorySubscriptions, len(dbCategories))

	for i, dbCategory := range dbCategories {
		subscribedFeeds, err := r.feedGetAllByCategory(ctx, userUUID, dbCategory.UUID)
		if err != nil {
			return []feedexporting.CategorySubscriptions{}, err
		}

		category := dbCategory.asCategory()
		categorySubscriptions := feedexporting.CategorySubscriptions{
			Category:        category,
			SubscribedFeeds: subscribedFeeds,
		}

		categoriesSubscriptions[i] = categorySubscriptions
	}

	return categoriesSubscriptions, nil
}

func (r *Repository) FeedSubscriptionCategoryGetAll(ctx context.Context, userUUID string) ([]feedquerying.SubscribedFeedsByCategory, error) {
	dbCategories, err := r.feedGetCategories(ctx, userUUID)
	if err != nil {
		return []feedquerying.SubscribedFeedsByCategory{}, err
	}

	categories := make([]feedquerying.SubscribedFeedsByCategory, len(dbCategories))

	for i, dbCategory := range dbCategories {
		dbFeeds, err := r.feedGetSubscriptionsByCategory(ctx, userUUID, dbCategory.UUID)
		if err != nil {
			return []feedquerying.SubscribedFeedsByCategory{}, err
		}

		var unread uint
		subscribedFeeds := make([]feedquerying.SubscribedFeed, len(dbFeeds))

		for j, dbFeed := range dbFeeds {
			subscribedFeeds[j] = dbFeed.asSubscribedFeed()
			unread += dbFeed.Unread
		}

		category := feedquerying.SubscribedFeedsByCategory{
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

func (r *Repository) FeedSubscriptionEntryGetN(ctx context.Context, userUUID string, preferences feed.Preferences, n uint, offset uint) ([]feedquerying.SubscribedFeedEntry, error) {
	const (
		query = `
		SELECT fe.uid, fe.url, fe.title, fe.summary, fe.published_at, fe.updated_at, fs.alias AS subscription_alias, f.uuid AS feed_uuid, f.title AS feed_title, COALESCE(fem.read, FALSE) AS read
		FROM feed_entries fe
		LEFT JOIN feed_entries_metadata fem ON fem.entry_uid = fe.uid
		JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
		JOIN feed_feeds f ON f.uuid = fe.feed_uuid
		WHERE fs.user_uuid=@user_uuid`
	)

	args := pgx.NamedArgs{
		"user_uuid": userUUID,
		"limit":     n,
		"offset":    offset,
	}

	return r.feedSubscriptionEntryGetN(ctx, query, preferences.ShowEntries, args)
}

func (r *Repository) FeedSubscriptionEntryGetNByCategory(ctx context.Context, userUUID string, preferences feed.Preferences, categoryUUID string, n uint, offset uint) ([]feedquerying.SubscribedFeedEntry, error) {
	const (
		query = `
		SELECT  fe.uid, fe.url, fe.title, fe.summary, fe.published_at, fe.updated_at, fs.alias AS subscription_alias, f.uuid AS feed_uuid, f.title AS feed_title, COALESCE(fem.read, FALSE) AS read
		FROM feed_entries fe
		LEFT JOIN feed_entries_metadata fem ON fem.entry_uid = fe.uid
		JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
		JOIN feed_feeds f ON f.uuid = fe.feed_uuid
		WHERE fs.user_uuid=@user_uuid
		AND   fs.category_uuid=@category_uuid`
	)

	args := pgx.NamedArgs{
		"user_uuid":     userUUID,
		"category_uuid": categoryUUID,
		"limit":         n,
		"offset":        offset,
	}

	return r.feedSubscriptionEntryGetN(ctx, query, preferences.ShowEntries, args)
}

func (r *Repository) FeedSubscriptionEntryGetNBySubscription(ctx context.Context, userUUID string, preferences feed.Preferences, subscriptionUUID string, n uint, offset uint) ([]feedquerying.SubscribedFeedEntry, error) {
	const (
		query = `
		SELECT  fe.uid, fe.url, fe.title, fe.summary, fe.published_at, fe.updated_at, fs.alias AS subscription_alias, f.uuid AS feed_uuid, f.title AS feed_title, COALESCE(fem.read, FALSE) AS read
		FROM feed_entries fe
		LEFT JOIN feed_entries_metadata fem ON fem.entry_uid = fe.uid
		JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
		JOIN feed_feeds f ON f.uuid = fe.feed_uuid
		WHERE fs.user_uuid=@user_uuid
		AND   fs.uuid=@subscription_uuid`
	)

	args := pgx.NamedArgs{
		"user_uuid":         userUUID,
		"subscription_uuid": subscriptionUUID,
		"limit":             n,
		"offset":            offset,
	}

	return r.feedSubscriptionEntryGetN(ctx, query, preferences.ShowEntries, args)
}

func (r *Repository) FeedSubscriptionEntryGetNByQuery(ctx context.Context, userUUID string, preferences feed.Preferences, searchTerms string, n uint, offset uint) ([]feedquerying.SubscribedFeedEntry, error) {
	const (
		query = `
		SELECT fe.uid, fe.url, fe.title, fe.summary, fe.published_at, fe.updated_at, fs.alias AS subscription_alias, f.uuid AS feed_uuid, f.title AS feed_title, COALESCE(fem.read, FALSE) AS read
		FROM feed_entries fe
		LEFT JOIN feed_entries_metadata fem ON fem.entry_uid = fe.uid
		JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
		JOIN feed_feeds f ON f.uuid = fe.feed_uuid
		WHERE fs.user_uuid=@user_uuid
		AND   (f.fulltextsearch_tsv || fe.fulltextsearch_tsv) @@ websearch_to_tsquery(@search_terms)`
	)

	args := pgx.NamedArgs{
		"user_uuid":    userUUID,
		"search_terms": pgbase.FullTextSearchReplacer.Replace(searchTerms),
		"limit":        n,
		"offset":       offset,
	}

	return r.feedSubscriptionEntryGetN(ctx, query, preferences.ShowEntries, args)
}

func (r *Repository) FeedSubscriptionEntryGetNByCategoryAndQuery(ctx context.Context, userUUID string, preferences feed.Preferences, categoryUUID string, searchTerms string, n uint, offset uint) ([]feedquerying.SubscribedFeedEntry, error) {
	const (
		query = `
		SELECT fe.uid, fe.url, fe.title, fe.summary, fe.published_at, fe.updated_at, fs.alias AS subscription_alias, f.uuid AS feed_uuid, f.title AS feed_title, COALESCE(fem.read, FALSE) AS read
		FROM feed_entries fe
		LEFT JOIN feed_entries_metadata fem ON fem.entry_uid = fe.uid
		JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
		JOIN feed_feeds f ON f.uuid = fe.feed_uuid
		WHERE fs.user_uuid=@user_uuid
		AND   fs.category_uuid=@category_uuid
		AND   (f.fulltextsearch_tsv || fe.fulltextsearch_tsv) @@ websearch_to_tsquery(@search_terms)`
	)

	args := pgx.NamedArgs{
		"user_uuid":     userUUID,
		"category_uuid": categoryUUID,
		"search_terms":  pgbase.FullTextSearchReplacer.Replace(searchTerms),
		"limit":         n,
		"offset":        offset,
	}

	return r.feedSubscriptionEntryGetN(ctx, query, preferences.ShowEntries, args)
}

func (r *Repository) FeedSubscriptionEntryGetNBySubscriptionAndQuery(ctx context.Context, userUUID string, preferences feed.Preferences, subscriptionUUID string, searchTerms string, n uint, offset uint) ([]feedquerying.SubscribedFeedEntry, error) {
	const (
		query = `
		SELECT fe.uid, fe.url, fe.title, fe.summary, fe.published_at, fe.updated_at, fs.alias AS subscription_alias, f.uuid AS feed_uuid, f.title AS feed_title, COALESCE(fem.read, FALSE) AS read
		FROM feed_entries fe
		LEFT JOIN feed_entries_metadata fem ON fem.entry_uid = fe.uid
		JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
		JOIN feed_feeds f ON f.uuid = fe.feed_uuid
		WHERE fs.user_uuid=@user_uuid
		AND   fs.uuid=@subscription_uuid
		AND   (f.fulltextsearch_tsv || fe.fulltextsearch_tsv) @@ websearch_to_tsquery(@search_terms)`
	)

	args := pgx.NamedArgs{
		"user_uuid":         userUUID,
		"subscription_uuid": subscriptionUUID,
		"search_terms":      pgbase.FullTextSearchReplacer.Replace(searchTerms),
		"limit":             n,
		"offset":            offset,
	}
	return r.feedSubscriptionEntryGetN(ctx, query, preferences.ShowEntries, args)
}

func (r *Repository) FeedSubscriptionIsRegistered(ctx context.Context, userUUID string, feedUUID string) (bool, error) {
	return r.RowExistsByQuery(
		ctx,
		"SELECT 1 FROM feed_subscriptions WHERE user_uuid=$1 AND feed_uuid=$2",
		userUUID,
		feedUUID,
	)
}

func (r *Repository) FeedSubscriptionCreate(ctx context.Context, s feed.Subscription) (feed.Subscription, error) {
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

	if err := r.QueryTx(ctx, domain, "FeedSubscriptionCreate", query, args); err != nil {
		return feed.Subscription{}, err
	}

	return s, nil
}

func (r *Repository) FeedSubscriptionDelete(ctx context.Context, userUUID string, subscriptionUUID string) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer r.Rollback(ctx, tx, domain, "FeedSubscriptionDelete")

	// 1. Delete the subscription
	commandTag, err := tx.Exec(
		ctx,
		"DELETE FROM feed_subscriptions WHERE user_uuid=$1 AND uuid=$2",
		userUUID,
		subscriptionUUID,
	)
	if err != nil {
		return err
	}

	rowsAffected := commandTag.RowsAffected()

	if rowsAffected != 1 {
		return feed.ErrSubscriptionNotFound
	}

	// 2. Delete feeds with no remaining subscriptions
	_, err = tx.Exec(
		ctx,
		`
		DELETE FROM feed_feeds
		WHERE uuid NOT IN (
			SELECT feed_uuid
			FROM feed_subscriptions
		)`,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) FeedSubscriptionGetByFeed(ctx context.Context, userUUID string, feedUUID string) (feed.Subscription, error) {
	query := `
	SELECT uuid, category_uuid, feed_uuid, user_uuid, alias, created_at, updated_at
	  FROM feed_subscriptions
	 WHERE user_uuid=$1
	   AND feed_uuid=$2`

	return r.feedSubscriptionGetQuery(ctx, query, userUUID, feedUUID)
}

func (r *Repository) FeedSubscriptionGetByUUID(ctx context.Context, userUUID string, subscriptionUUID string) (feed.Subscription, error) {
	query := `
	SELECT uuid, category_uuid, feed_uuid, user_uuid, alias, created_at, updated_at
	  FROM feed_subscriptions
	 WHERE user_uuid=$1
	   AND uuid=$2`

	return r.feedSubscriptionGetQuery(ctx, query, userUUID, subscriptionUUID)
}

func (r *Repository) FeedSubscriptionUpdate(ctx context.Context, s feed.Subscription) error {
	query := `
	UPDATE feed_subscriptions
	SET
		category_uuid=@category_uuid,
		updated_at=@updated_at,
		alias=@alias
	WHERE user_uuid=@user_uuid
	AND uuid=@uuid`

	args := pgx.NamedArgs{
		"user_uuid":     s.UserUUID,
		"uuid":          s.UUID,
		"category_uuid": s.CategoryUUID,
		"alias":         s.Alias,
		"updated_at":    s.UpdatedAt,
	}

	return r.QueryTx(ctx, domain, "FeedSubscriptionUpdate", query, args)
}

func (r *Repository) FeedQueryingSubscriptionByUUID(ctx context.Context, userUUID string, subscriptionUUID string) (feedquerying.Subscription, error) {
	query := `
	SELECT fs.uuid, fs.alias, fs.category_uuid, f.title, f.description
	FROM   feed_subscriptions fs
	JOIN   feed_feeds f ON f.uuid = fs.feed_uuid
	WHERE  fs.user_uuid=$1
	AND    fs.uuid=$2`

	return r.feedSubscriptionTitleGetQuery(ctx, query, userUUID, subscriptionUUID)
}

func (r *Repository) FeedQueryingSubscriptionsByCategory(ctx context.Context, userUUID string) ([]feedquerying.SubscriptionsByCategory, error) {
	dbCategories, err := r.feedGetCategories(ctx, userUUID)
	if err != nil {
		return []feedquerying.SubscriptionsByCategory{}, err
	}

	categories := make([]feedquerying.SubscriptionsByCategory, len(dbCategories))

	for i, dbCategory := range dbCategories {
		dbSubscriptionTitles, err := r.feedGetSubscriptionTitlesByCategory(ctx, userUUID, dbCategory.UUID)
		if err != nil {
			return []feedquerying.SubscriptionsByCategory{}, err
		}

		subscriptionTitles := make([]feedquerying.Subscription, len(dbSubscriptionTitles))

		for j, dbSubscriptionTitle := range dbSubscriptionTitles {
			subscriptionTitles[j] = dbSubscriptionTitle.asQueryingSubscription()
		}

		category := feedquerying.SubscriptionsByCategory{
			Category:      dbCategory.asCategory(),
			Subscriptions: subscriptionTitles,
		}

		categories[i] = category
	}

	return categories, nil
}
