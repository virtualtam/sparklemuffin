// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package postgresql

import (
	"context"
	"errors"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	fquerying "github.com/virtualtam/sparklemuffin/pkg/feed/querying"
)

var _ feed.Repository = &Repository{}
var _ fquerying.Repository = &Repository{}

type DBCategory struct {
	UUID     string `db:"uuid"`
	UserUUID string `db:"user_uuid"`

	Name string `db:"name"`
	Slug string `db:"slug"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type DBFeed struct {
	UUID string `db:"uuid"`

	URL   string `db:"url"`
	Title string `db:"title"`
	Slug  string `db:"slug"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	FetchedAt time.Time `db:"fetched_at"`
}

func (f *DBFeed) asFeed() feed.Feed {
	return feed.Feed{
		UUID:      f.UUID,
		URL:       f.URL,
		Title:     f.Title,
		Slug:      f.Slug,
		CreatedAt: f.CreatedAt,
		UpdatedAt: f.UpdatedAt,
		FetchedAt: f.FetchedAt,
	}
}

type DBEntry struct {
	UID      string `db:"uid"`
	FeedUUID string `db:"feed_uuid"`

	URL   string `db:"url"`
	Title string `db:"title"`

	PublishedAt time.Time `db:"published_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (e *DBEntry) asEntry() feed.Entry {
	return feed.Entry{
		UID:         e.UID,
		FeedUUID:    e.FeedUUID,
		URL:         e.URL,
		Title:       e.Title,
		PublishedAt: e.PublishedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

type DBSubscribedFeed struct {
	UUID    string `db:"uuid"`
	FeedURL string `db:"feed_url"`
	Title   string `db:"title"`
	Slug    string `db:"slug"`

	Unread uint `db:"unread"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	FetchedAt time.Time `db:"fetched_at"`
}

func (f *DBSubscribedFeed) asSubscribedFeed() fquerying.SubscribedFeed {
	return fquerying.SubscribedFeed{
		Feed: feed.Feed{
			UUID:      f.UUID,
			FeedURL:   f.FeedURL,
			Title:     f.Title,
			Slug:      f.Slug,
			CreatedAt: f.CreatedAt,
			UpdatedAt: f.UpdatedAt,
			FetchedAt: f.FetchedAt,
		},
		Unread: f.Unread,
	}
}

func (r *Repository) FeedCreate(f feed.Feed) error {
	query := `
	INSERT INTO feed_feeds(
		uuid,
		url,
		title,
		slug,
		created_at,
		updated_at
	)
	VALUES(
		@uuid,
		@url,
		@title,
		@slug,
		@created_at,
		@updated_at
	)`

	args := pgx.NamedArgs{
		"uuid":       f.UUID,
		"url":        f.URL,
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

func (r *Repository) FeedGetByURL(feedURL string) (feed.Feed, error) {
	query := `
SELECT uuid, url, title, slug
FROM feed_feeds
WHERE url=$1`

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

func (r *Repository) feedGetCategories(userUUID string) ([]DBCategory, error) {
	query := `SELECT uuid, name, slug FROM feed_categories WHERE user_uuid=$1`

	rows, err := r.pool.Query(context.Background(), query, userUUID)
	if err != nil {
		return []DBCategory{}, err
	}
	defer rows.Close()

	var dbCategories []DBCategory
	if err := pgxscan.ScanAll(&dbCategories, rows); err != nil {
		return []DBCategory{}, err
	}

	return dbCategories, nil
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

	for _, b := range entries {
		args := pgx.NamedArgs{
			"uid":          b.UID,
			"feed_uuid":    b.FeedUUID,
			"url":          b.URL,
			"title":        b.Title,
			"published_at": b.PublishedAt,
			"updated_at":   b.UpdatedAt,
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
				Str("operation", "create_many").
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

func (r *Repository) feedGetSubscriptionsByCategory(userUUID string, categoryUUID string) ([]DBSubscribedFeed, error) {
	query := `
SELECT f.feed_url, f.title, f.slug, COUNT(*) AS unread
FROM feed_subscriptions fs
JOIN feed_feeds f ON f.uuid=fs.feed_uuid
JOIN feed_entries fe ON fe.feed_uuid=fs.feed_uuid
WHERE fs.user_uuid=$1
AND   fs.category_uuid=$2
GROUP BY f.feed_url, f.title, f.slug`

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
