// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"errors"
	"slices"
)

var _ Repository = &FakeRepository{}

type FakeRepository struct {
	Categories      []Category
	Entries         []Entry
	EntriesMetadata []EntryMetadata
	Feeds           []Feed
	Subscriptions   []Subscription
}

func (r *FakeRepository) FeedCreate(feed Feed) error {
	r.Feeds = append(r.Feeds, feed)
	return nil
}

func (r *FakeRepository) FeedGetByURL(feedURL string) (Feed, error) {
	for _, f := range r.Feeds {
		if f.FeedURL == feedURL {
			return f, nil
		}
	}

	return Feed{}, ErrFeedNotFound
}

func (r *FakeRepository) FeedGetBySlug(feedSlug string) (Feed, error) {
	for _, feed := range r.Feeds {
		if feed.Slug == feedSlug {
			return feed, nil
		}
	}

	return Feed{}, ErrFeedNotFound
}

func (r *FakeRepository) FeedCategoryCreate(category Category) error {
	r.Categories = append(r.Categories, category)
	return nil
}

func (r *FakeRepository) FeedCategoryDelete(userUUID string, categoryUUID string) error {
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

func (r *FakeRepository) FeedCategoryGetByName(userUUID string, name string) (Category, error) {
	for _, category := range r.Categories {
		if category.UserUUID == userUUID && category.Name == name {
			return category, nil
		}
	}

	return Category{}, ErrCategoryNotFound
}

func (r *FakeRepository) FeedCategoryGetBySlug(userUUID string, slug string) (Category, error) {
	for _, category := range r.Categories {
		if category.UserUUID == userUUID && category.Slug == slug {
			return category, nil
		}
	}

	return Category{}, ErrCategoryNotFound
}

func (r *FakeRepository) FeedCategoryGetByUUID(userUUID string, categoryUUID string) (Category, error) {
	for _, category := range r.Categories {
		if category.UserUUID == userUUID && category.UUID == categoryUUID {
			return category, nil
		}
	}

	return Category{}, ErrCategoryNotFound
}

func (r *FakeRepository) FeedCategoryGetMany(userUUID string) ([]Category, error) {
	panic("unimplemented")
}

func (r *FakeRepository) FeedCategoryNameAndSlugAreRegistered(userUUID string, name string, slug string) (bool, error) {
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

func (r *FakeRepository) FeedCategoryNameAndSlugAreRegisteredToAnotherCategory(userUUID string, categoryUUID string, name string, slug string) (bool, error) {
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

func (r *FakeRepository) FeedCategoryUpdate(category Category) error {
	for index, c := range r.Categories {
		if c.UserUUID == category.UserUUID && c.UUID == category.UUID {
			r.Categories[index] = category
		}
	}
	return nil
}

func (r *FakeRepository) FeedEntryCreateMany(entries []Entry) (int64, error) {
	r.Entries = append(r.Entries, entries...)
	return int64(len(entries)), nil
}

func (r *FakeRepository) FeedEntryGetN(feedUUID string, n uint) ([]Entry, error) {
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

func (r *FakeRepository) feedEntryExists(entryUID string) bool {
	for _, entry := range r.Entries {
		if entry.UID == entryUID {
			return true
		}
	}

	return false
}

func (r *FakeRepository) FeedEntryMarkAllAsRead(userUUID string) error {
	return errors.New("not implemented")
}

func (r *FakeRepository) FeedEntryMarkAllAsReadByCategory(userUUID string, categoryUUID string) error {
	return errors.New("not implemented")
}

func (r *FakeRepository) FeedEntryMarkAllAsReadBySubscription(userUUID string, subscriptionUUID string) error {
	return errors.New("not implemented")
}

func (r *FakeRepository) FeedEntryMetadataCreate(newEntryMetadata EntryMetadata) error {
	if !r.feedEntryExists(newEntryMetadata.EntryUID) {
		return ErrEntryNotFound
	}

	r.EntriesMetadata = append(r.EntriesMetadata, newEntryMetadata)
	return nil
}

func (r *FakeRepository) FeedEntryMetadataGetByUID(userUUID string, entryUID string) (EntryMetadata, error) {
	for _, entryMetadata := range r.EntriesMetadata {
		if entryMetadata.UserUUID == userUUID && entryMetadata.EntryUID == entryUID {
			return entryMetadata, nil
		}
	}

	return EntryMetadata{}, ErrEntryMetadataNotFound
}

func (r *FakeRepository) FeedEntryMetadataUpdate(updatedEntryMetadata EntryMetadata) error {
	if !r.feedEntryExists(updatedEntryMetadata.EntryUID) {
		return ErrEntryNotFound
	}

	for index, entryMetadata := range r.EntriesMetadata {
		if entryMetadata.UserUUID == updatedEntryMetadata.UserUUID && entryMetadata.EntryUID == updatedEntryMetadata.EntryUID {
			r.EntriesMetadata[index] = updatedEntryMetadata
			return nil
		}
	}

	return ErrEntryMetadataNotFound
}

func (r *FakeRepository) FeedSubscriptionIsRegistered(userUUID string, feedUUID string) (bool, error) {
	for _, s := range r.Subscriptions {
		if s.UserUUID == userUUID && s.FeedUUID == feedUUID {
			return true, nil
		}
	}

	return false, nil
}

func (r *FakeRepository) FeedSubscriptionCreate(subscription Subscription) (Subscription, error) {
	r.Subscriptions = append(r.Subscriptions, subscription)
	return subscription, nil
}

func (r *FakeRepository) FeedSubscriptionDelete(userUUID string, subscriptionUUID string) error {
	for index, s := range r.Subscriptions {
		if s.UserUUID == userUUID && s.UUID == subscriptionUUID {
			r.Subscriptions = append(r.Subscriptions[:index], r.Subscriptions[index+1:]...)

			return nil
		}
	}

	return ErrSubscriptionNotFound
}

func (r *FakeRepository) FeedSubscriptionGetByFeed(userUUID string, feedUUID string) (Subscription, error) {
	for _, subscription := range r.Subscriptions {
		if subscription.UserUUID == userUUID && subscription.FeedUUID == feedUUID {
			return subscription, nil
		}
	}

	return Subscription{}, ErrSubscriptionNotFound
}

func (r *FakeRepository) FeedSubscriptionGetByUUID(userUUID string, subscriptionUUID string) (Subscription, error) {
	for _, subscription := range r.Subscriptions {
		if subscription.UserUUID == userUUID && subscription.UUID == subscriptionUUID {
			return subscription, nil
		}
	}

	return Subscription{}, ErrSubscriptionNotFound
}

func (r *FakeRepository) FeedSubscriptionUpdate(subscription Subscription) error {
	for index, c := range r.Subscriptions {
		if c.UserUUID == subscription.UserUUID && c.UUID == subscription.UUID {
			r.Subscriptions[index] = subscription
			return nil
		}
	}

	return ErrSubscriptionNotFound
}
