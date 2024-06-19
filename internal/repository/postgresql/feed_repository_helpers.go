// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package postgresql

import (
	"context"
	"errors"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
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
SELECT fs.uuid, f.title
FROM feed_subscriptions fs
JOIN feed_feeds f ON f.uuid=fs.feed_uuid
WHERE fs.user_uuid=$1
AND   fs.category_uuid=$2
ORDER BY f.title`

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
