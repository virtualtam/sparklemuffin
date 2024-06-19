// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"slices"
)

var _ Repository = &fakeRepository{}

type fakeRepository struct {
	Categories    []Category
	Entries       []Entry
	Feeds         []Feed
	Subscriptions []Subscription
}

func (r *fakeRepository) FeedAdd(feed Feed) error {
	r.Feeds = append(r.Feeds, feed)
	return nil
}

func (r *fakeRepository) FeedGetByURL(feedURL string) (Feed, error) {
	for _, f := range r.Feeds {
		if f.FeedURL == feedURL {
			return f, nil
		}
	}

	return Feed{}, ErrFeedNotFound
}

func (r *fakeRepository) FeedGetBySlug(feedSlug string) (Feed, error) {
	for _, feed := range r.Feeds {
		if feed.Slug == feedSlug {
			return feed, nil
		}
	}

	return Feed{}, ErrFeedNotFound
}

func (r *fakeRepository) FeedCategoryAdd(category Category) error {
	r.Categories = append(r.Categories, category)
	return nil
}

func (r *fakeRepository) FeedCategoryDelete(userUUID string, categoryUUID string) error {
	r.Categories = slices.DeleteFunc(r.Categories, func(c Category) bool {
		return c.UserUUID == userUUID && c.UUID == categoryUUID
	})

	deletedSubscriptionsUUIDs := []string{}

	r.Subscriptions = slices.DeleteFunc(r.Subscriptions, func(s Subscription) bool {
		deletedSubscriptionsUUIDs = append(deletedSubscriptionsUUIDs, s.UUID)

		return s.UserUUID == userUUID && s.CategoryUUID == categoryUUID
	})

	return nil
}

func (r *fakeRepository) FeedCategoryGetBySlug(userUUID string, slug string) (Category, error) {
	for _, category := range r.Categories {
		if category.UserUUID == userUUID && category.Slug == slug {
			return category, nil
		}
	}

	return Category{}, ErrCategoryNotFound
}

func (r *fakeRepository) FeedCategoryGetByUUID(userUUID string, categoryUUID string) (Category, error) {
	for _, category := range r.Categories {
		if category.UserUUID == userUUID && category.UUID == categoryUUID {
			return category, nil
		}
	}

	return Category{}, ErrCategoryNotFound
}

func (r *fakeRepository) FeedCategoryGetMany(userUUID string) ([]Category, error) {
	panic("unimplemented")
}

func (r *fakeRepository) FeedCategoryNameAndSlugAreRegistered(userUUID string, name string, slug string) (bool, error) {
	for _, category := range r.Categories {
		if category.UserUUID != userUUID {
			continue
		}

		if category.Name == name || category.Slug == slug {
			return true, nil
		}
	}

	return false, nil
}

func (r *fakeRepository) FeedCategoryNameAndSlugAreRegisteredToAnotherCategory(userUUID string, categoryUUID string, name string, slug string) (bool, error) {
	for _, category := range r.Categories {
		if category.UserUUID != userUUID {
			continue
		}

		if category.UUID == categoryUUID {
			continue
		}

		if category.Name == name || category.Slug == slug {
			return true, nil
		}
	}

	return false, nil
}

func (r *fakeRepository) FeedCategoryUpdate(category Category) error {
	for index, c := range r.Categories {
		if c.UserUUID == category.UserUUID && c.UUID == category.UUID {
			r.Categories[index] = category
		}
	}
	return nil
}

func (r *fakeRepository) FeedEntryAddMany(entries []Entry) (int64, error) {
	r.Entries = append(r.Entries, entries...)
	return int64(len(entries)), nil
}

func (r *fakeRepository) FeedEntryGetN(feedUUID string, n uint) ([]Entry, error) {
	var entries []Entry
	var count uint

	for _, entry := range r.Entries {
		if entry.FeedUUID != feedUUID {
			continue
		}

		count++
		entries = append(entries, entry)

		if count == n {
			break
		}
	}

	return entries, nil
}

func (r *fakeRepository) FeedSubscriptionIsRegistered(userUUID string, feedUUID string) (bool, error) {
	for _, s := range r.Subscriptions {
		if s.UserUUID == userUUID && s.FeedUUID == feedUUID {
			return true, nil
		}
	}

	return false, nil
}

func (r *fakeRepository) FeedSubscriptionAdd(subscription Subscription) error {
	r.Subscriptions = append(r.Subscriptions, subscription)
	return nil
}

func (r *fakeRepository) FeedSubscriptionGetByFeed(userUUID string, feedUUID string) (Subscription, error) {
	for _, subscription := range r.Subscriptions {
		if subscription.UserUUID == userUUID && subscription.FeedUUID == feedUUID {
			return subscription, nil
		}
	}

	return Subscription{}, ErrSubscriptionNotFound
}
