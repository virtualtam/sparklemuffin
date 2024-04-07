// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package postgresql

import (
	"context"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
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
