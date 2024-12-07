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

func (r *Repository) FeedCreate(f feed.Feed) error {
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
		to_tsvector(@fulltextsearch_string),
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

	return r.QueryTx("feeds", "FeedCreate", query, args)
}

func (r *Repository) FeedGetBySlug(feedSlug string) (feed.Feed, error) {
	query := `
	SELECT uuid, feed_url, title, description, slug, etag, last_modified, hash_xxhash64, created_at, updated_at, fetched_at
	FROM feed_feeds
	WHERE slug=$1`

	return r.feedGetQuery(query, feedSlug)
}

func (r *Repository) FeedGetByURL(feedURL string) (feed.Feed, error) {
	query := `
	SELECT uuid, feed_url, title, description, slug, etag, last_modified, hash_xxhash64, created_at, updated_at, fetched_at
	FROM feed_feeds
	WHERE feed_url=$1`

	return r.feedGetQuery(query, feedURL)
}

func (r *Repository) FeedGetByUUID(feedUUID string) (feed.Feed, error) {
	query := `
	SELECT uuid, feed_url, title, description, slug, etag, last_modified, hash_xxhash64, created_at, updated_at, fetched_at
	FROM feed_feeds
	WHERE uuid=$1`

	return r.feedGetQuery(query, feedUUID)
}

func (r *Repository) FeedGetNByLastSynchronizationTime(n uint, before time.Time) ([]feed.Feed, error) {
	query := `
	SELECT f.uuid, f.feed_url, f.title, f.description, f.slug, f.etag, f.last_modified, f.hash_xxhash64, f.created_at, f.updated_at, f.fetched_at
	FROM feed_feeds f
	INNER JOIN feed_subscriptions fs ON f.uuid = fs.feed_uuid
	WHERE fetched_at < $1
	OR    fetched_at IS NULL
	LIMIT $2`

	return r.feedGetManyQuery(query, before, n)
}

func (r *Repository) FeedUpdateFetchMetadata(feedFetchMetadata feedsynchronizing.FeedFetchMetadata) error {
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

	return r.QueryTx("feeds", "FeedUpdateFetchMetadata", query, args)
}

func (r *Repository) FeedUpdateMetadata(feedMetadata feedsynchronizing.FeedMetadata) error {
	query := `
	UPDATE feed_feeds
	SET
		title=@title,
		description=@description,
		hash_xxhash64=@hash_xxhash64,
		fulltextsearch_tsv=to_tsvector(@fulltextsearch_string),
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

	return r.QueryTx("feeds", "FeedUpdateFetchMetadata", query, args)
}

func (r *Repository) FeedCategoryCreate(c feed.Category) error {
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

	return r.QueryTx("feeds", "FeedCategoryCreate", query, args)
}

func (r *Repository) FeedCategoryDelete(userUUID string, categoryUUID string) error {
	ctx := context.Background()

	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer r.Rollback(ctx, tx, "feeds", "FeedCategoryDelete")

	commandTag, err := tx.Exec(
		context.Background(),
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

	return tx.Commit(ctx)
}

func (r *Repository) FeedCategoryGetByName(userUUID string, name string) (feed.Category, error) {
	query := `
	SELECT uuid, user_uuid, name, slug, created_at, updated_at
	FROM feed_categories
	WHERE user_uuid=$1
	AND name=$2`

	return r.feedCategoryGetQuery(query, userUUID, name)
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

	rows, err := r.Pool.Query(context.Background(), query, userUUID)
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
	return r.RowExistsByQuery(
		"SELECT 1 FROM feed_categories WHERE user_uuid=$1 AND (name=$2 OR slug=$3)",
		userUUID,
		name,
		slug,
	)
}

func (r *Repository) FeedCategoryNameAndSlugAreRegisteredToAnotherCategory(userUUID string, categoryUUID string, name string, slug string) (bool, error) {
	return r.RowExistsByQuery(
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

	return r.QueryTx("feeds", "FeedCategoryUpdate", query, args)
}

func (r *Repository) FeedEntryCreateMany(entries []feed.Entry) (int64, error) {
	return r.feedEntryUpsertMany("FeedEntryCreateMany", "ON CONFLICT DO NOTHING", entries)
}

func (r *Repository) FeedEntryUpsertMany(entries []feed.Entry) (int64, error) {
	return r.feedEntryUpsertMany(
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

func (r *Repository) FeedEntryGetN(feedUUID string, n uint) ([]feed.Entry, error) {
	query := `
	SELECT feed_uuid, uid, url, title, published_at, updated_at
	FROM  feed_entries
	WHERE feed_uuid=$1
	ORDER BY published_at DESC
	LIMIT $2`

	rows, err := r.Pool.Query(context.Background(), query, feedUUID, n)
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

func (r *Repository) FeedEntryGetCount(userUUID string) (uint, error) {
	query := `
	SELECT COUNT(*)
	FROM feed_entries fe
	JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
	WHERE fs.user_uuid=$1`

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

func (r *Repository) FeedEntryGetCountByCategory(userUUID string, categoryUUID string) (uint, error) {
	query := `
	SELECT COUNT(*)
	FROM feed_entries fe
	JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
	WHERE fs.user_uuid=$1
	AND   fs.category_uuid=$2`

	var count uint

	err := r.Pool.QueryRow(
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

	err := r.Pool.QueryRow(
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

func (r *Repository) FeedEntryGetCountByQuery(userUUID string, searchTerms string) (uint, error) {
	query := `
	SELECT COUNT(*)
	FROM feed_entries fe
	JOIN feed_feeds f ON f.uuid = fe.feed_uuid
	JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
	WHERE fs.user_uuid=$1
	AND   (f.fulltextsearch_tsv || fe.fulltextsearch_tsv) @@ websearch_to_tsquery($2)`

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

func (r *Repository) FeedEntryGetCountByCategoryAndQuery(userUUID string, categoryUUID string, searchTerms string) (uint, error) {
	query := `
	SELECT COUNT(*)
	FROM feed_entries fe
	JOIN feed_feeds f ON f.uuid = fe.feed_uuid
	JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
	WHERE fs.user_uuid=$1
	AND   fs.category_uuid=$2
	AND   (f.fulltextsearch_tsv || fe.fulltextsearch_tsv) @@ websearch_to_tsquery($3)`

	var count uint
	fullTextSearchTerms := pgbase.FullTextSearchReplacer.Replace(searchTerms)

	err := r.Pool.QueryRow(
		context.Background(),
		query,
		userUUID,
		categoryUUID,
		fullTextSearchTerms,
	).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *Repository) FeedEntryGetCountBySubscriptionAndQuery(userUUID string, subscriptionUUID string, searchTerms string) (uint, error) {
	query := `
	SELECT COUNT(*)
	FROM feed_entries fe
	JOIN feed_feeds f ON f.uuid = fe.feed_uuid
	JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
	WHERE user_uuid=$1
	AND   fs.uuid=$2
	AND   (f.fulltextsearch_tsv || fe.fulltextsearch_tsv) @@ websearch_to_tsquery($3)`

	var count uint
	fullTextSearchTerms := pgbase.FullTextSearchReplacer.Replace(searchTerms)

	err := r.Pool.QueryRow(
		context.Background(),
		query,
		userUUID,
		subscriptionUUID,
		fullTextSearchTerms,
	).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *Repository) FeedEntryMarkAllAsRead(userUUID string) error {
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

	return r.QueryTx("feeds", "FeedEntryMarkAllAsRead", query, args)
}

func (r *Repository) FeedEntryMarkAllAsReadByCategory(userUUID string, categoryUUID string) error {
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

	return r.QueryTx("feeds", "FeedEntryMarkAllAsReadByCategory", query, args)
}

func (r *Repository) FeedEntryMarkAllAsReadBySubscription(userUUID string, subscriptionUUID string) error {
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

	return r.QueryTx("feeds", "FeedEntryMarkAllAsReadBySubscription", query, args)
}

func (r *Repository) FeedEntryMetadataCreate(entryMetadata feed.EntryMetadata) error {
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

	return r.QueryTx("feeds", "FeedEntryMetadataCreate", query, args)
}

func (r *Repository) FeedEntryMetadataGetByUID(userUUID string, entryUID string) (feed.EntryMetadata, error) {
	query := `
	SELECT user_uuid, entry_uid, read
	FROM feed_entries_metadata
	WHERE user_uuid=$1
	AND   entry_uid=$2
	`

	rows, err := r.Pool.Query(context.Background(), query, userUUID, entryUID)
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

func (r *Repository) FeedEntryMetadataUpdate(entryMetadata feed.EntryMetadata) error {
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

	return r.QueryTx("feeds", "FeedEntryMetadataUpdate", query, args)
}

func (r *Repository) FeedCategorySubscriptionsGetAll(userUUID string) ([]feedexporting.CategorySubscriptions, error) {
	dbCategories, err := r.feedGetCategories(userUUID)
	if err != nil {
		return []feedexporting.CategorySubscriptions{}, err
	}

	categoriesSubscriptions := make([]feedexporting.CategorySubscriptions, len(dbCategories))

	for i, dbCategory := range dbCategories {
		subscribedFeeds, err := r.feedGetAllByCategory(userUUID, dbCategory.UUID)
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

func (r *Repository) FeedSubscriptionCategoryGetAll(userUUID string) ([]feedquerying.SubscribedFeedsByCategory, error) {
	dbCategories, err := r.feedGetCategories(userUUID)
	if err != nil {
		return []feedquerying.SubscribedFeedsByCategory{}, err
	}

	categories := make([]feedquerying.SubscribedFeedsByCategory, len(dbCategories))

	for i, dbCategory := range dbCategories {
		dbFeeds, err := r.feedGetSubscriptionsByCategory(userUUID, dbCategory.UUID)
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

func (r *Repository) FeedSubscriptionEntryGetN(userUUID string, n uint, offset uint) ([]feedquerying.SubscribedFeedEntry, error) {
	query := `
	SELECT fe.uid, fe.url, fe.title, fe.summary, fe.published_at, f.title AS feed_title, COALESCE(fem.read, FALSE) AS read
	FROM feed_entries fe
	LEFT JOIN feed_entries_metadata fem ON fem.entry_uid = fe.uid
	JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
	JOIN feed_feeds f ON f.uuid = fe.feed_uuid
	WHERE fs.user_uuid=$1
	ORDER BY fe.published_at DESC
	LIMIT $2 OFFSET $3`

	return r.feedSubscriptionEntryGetN(query, userUUID, n, offset)
}

func (r *Repository) FeedSubscriptionEntryGetNByCategory(userUUID string, categoryUUID string, n uint, offset uint) ([]feedquerying.SubscribedFeedEntry, error) {
	query := `
	SELECT  fe.uid, fe.url, fe.title, fe.summary, fe.published_at, f.title AS feed_title, COALESCE(fem.read, FALSE) AS read
	FROM feed_entries fe
	LEFT JOIN feed_entries_metadata fem ON fem.entry_uid = fe.uid
	JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
	JOIN feed_feeds f ON f.uuid = fe.feed_uuid
	WHERE fs.user_uuid=$1
	AND   fs.category_uuid=$2
	ORDER BY fe.published_at DESC
	LIMIT $3 OFFSET $4`

	return r.feedSubscriptionEntryGetN(query, userUUID, categoryUUID, n, offset)
}

func (r *Repository) FeedSubscriptionEntryGetNBySubscription(userUUID string, subscriptionUUID string, n uint, offset uint) ([]feedquerying.SubscribedFeedEntry, error) {
	query := `
	SELECT  fe.uid, fe.url, fe.title, fe.summary, fe.published_at, f.title AS feed_title, COALESCE(fem.read, FALSE) AS read
	FROM feed_entries fe
	LEFT JOIN feed_entries_metadata fem ON fem.entry_uid = fe.uid
	JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
	JOIN feed_feeds f ON f.uuid = fe.feed_uuid
	WHERE fs.user_uuid=$1
	AND   fs.uuid=$2
	ORDER BY fe.published_at DESC
	LIMIT $3 OFFSET $4`

	return r.feedSubscriptionEntryGetN(query, userUUID, subscriptionUUID, n, offset)
}

func (r *Repository) FeedSubscriptionEntryGetNByQuery(userUUID string, searchTerms string, n uint, offset uint) ([]feedquerying.SubscribedFeedEntry, error) {
	query := `
	SELECT fe.uid, fe.url, fe.title, fe.summary, fe.published_at, f.title AS feed_title, COALESCE(fem.read, FALSE) AS read
	FROM feed_entries fe
	LEFT JOIN feed_entries_metadata fem ON fem.entry_uid = fe.uid
	JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
	JOIN feed_feeds f ON f.uuid = fe.feed_uuid
	WHERE fs.user_uuid=$1
	AND   (f.fulltextsearch_tsv || fe.fulltextsearch_tsv) @@ websearch_to_tsquery($2)
	ORDER BY fe.published_at DESC
	LIMIT $3 OFFSET $4`

	fullTextSearchTerms := pgbase.FullTextSearchReplacer.Replace(searchTerms)

	return r.feedSubscriptionEntryGetN(query, userUUID, fullTextSearchTerms, n, offset)
}

func (r *Repository) FeedSubscriptionEntryGetNByCategoryAndQuery(userUUID string, categoryUUID string, searchTerms string, n uint, offset uint) ([]feedquerying.SubscribedFeedEntry, error) {
	query := `
	SELECT fe.uid, fe.url, fe.title, fe.summary, fe.published_at, f.title AS feed_title, COALESCE(fem.read, FALSE) AS read
	FROM feed_entries fe
	LEFT JOIN feed_entries_metadata fem ON fem.entry_uid = fe.uid
	JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
	JOIN feed_feeds f ON f.uuid = fe.feed_uuid
	WHERE fs.user_uuid=$1
	AND   fs.category_uuid=$2
	AND   (f.fulltextsearch_tsv || fe.fulltextsearch_tsv) @@ websearch_to_tsquery($3)
	ORDER BY fe.published_at DESC
	LIMIT $4 OFFSET $5`

	fullTextSearchTerms := pgbase.FullTextSearchReplacer.Replace(searchTerms)

	return r.feedSubscriptionEntryGetN(query, userUUID, categoryUUID, fullTextSearchTerms, n, offset)
}

func (r *Repository) FeedSubscriptionEntryGetNBySubscriptionAndQuery(userUUID string, subscriptionUUID string, searchTerms string, n uint, offset uint) ([]feedquerying.SubscribedFeedEntry, error) {
	query := `
	SELECT fe.uid, fe.url, fe.title, fe.summary, fe.published_at, f.title AS feed_title, COALESCE(fem.read, FALSE) AS read
	FROM feed_entries fe
	LEFT JOIN feed_entries_metadata fem ON fem.entry_uid = fe.uid
	JOIN feed_subscriptions fs ON fs.feed_uuid = fe.feed_uuid
	JOIN feed_feeds f ON f.uuid = fe.feed_uuid
	WHERE fs.user_uuid=$1
	AND   fs.uuid=$2
	AND   (f.fulltextsearch_tsv || fe.fulltextsearch_tsv) @@ websearch_to_tsquery($3)
	ORDER BY fe.published_at DESC
	LIMIT $4 OFFSET $5`

	fullTextSearchTerms := pgbase.FullTextSearchReplacer.Replace(searchTerms)

	return r.feedSubscriptionEntryGetN(query, userUUID, subscriptionUUID, fullTextSearchTerms, n, offset)
}

func (r *Repository) FeedSubscriptionIsRegistered(userUUID string, feedUUID string) (bool, error) {
	return r.RowExistsByQuery(
		"SELECT 1 FROM feed_subscriptions WHERE user_uuid=$1 AND feed_uuid=$2",
		userUUID,
		feedUUID,
	)
}

func (r *Repository) FeedSubscriptionCreate(s feed.Subscription) (feed.Subscription, error) {
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

	if err := r.QueryTx("feeds", "FeedSubscriptionCreate", query, args); err != nil {
		return feed.Subscription{}, err
	}

	return s, nil
}

func (r *Repository) FeedSubscriptionDelete(userUUID string, subscriptionUUID string) error {
	ctx := context.Background()

	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer r.Rollback(ctx, tx, "feeds", "FeedSubscriptionDelete")

	commandTag, err := tx.Exec(
		context.Background(),
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

	return tx.Commit(ctx)
}

func (r *Repository) FeedSubscriptionGetByFeed(userUUID string, feedUUID string) (feed.Subscription, error) {
	query := `
	SELECT uuid, category_uuid, feed_uuid, user_uuid, alias, created_at, updated_at
	FROM feed_subscriptions
	WHERE feed_uuid=$1`

	return r.feedSubscriptionGetQuery(query, feedUUID)
}

func (r *Repository) FeedSubscriptionGetByUUID(userUUID string, subscriptionUUID string) (feed.Subscription, error) {
	query := `
	SELECT uuid, category_uuid, feed_uuid, user_uuid, alias, created_at, updated_at
	FROM feed_subscriptions
	WHERE uuid=$1`

	return r.feedSubscriptionGetQuery(query, subscriptionUUID)
}

func (r *Repository) FeedSubscriptionUpdate(s feed.Subscription) error {
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

	return r.QueryTx("feeds", "FeedSubscriptionUpdate", query, args)
}

func (r *Repository) FeedQueryingSubscriptionByUUID(userUUID string, subscriptionUUID string) (feedquerying.Subscription, error) {
	query := `
	SELECT fs.uuid, fs.alias, fs.category_uuid, f.title, f.description
	FROM   feed_subscriptions fs
	JOIN   feed_feeds f ON f.uuid = fs.feed_uuid
	WHERE  fs.user_uuid=$1
	AND    fs.uuid=$2`

	return r.feedSubscriptionTitleGetQuery(query, userUUID, subscriptionUUID)
}

func (r *Repository) FeedQueryingSubscriptionsByCategory(userUUID string) ([]feedquerying.SubscriptionsByCategory, error) {
	dbCategories, err := r.feedGetCategories(userUUID)
	if err != nil {
		return []feedquerying.SubscriptionsByCategory{}, err
	}

	categories := make([]feedquerying.SubscriptionsByCategory, len(dbCategories))

	for i, dbCategory := range dbCategories {
		dbSubscriptionTitles, err := r.feedGetSubscriptionTitlesByCategory(userUUID, dbCategory.UUID)
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
