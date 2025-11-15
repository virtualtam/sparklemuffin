// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"context"
	"errors"
	"slices"

	"github.com/virtualtam/sparklemuffin/pkg/user"
)

var _ Repository = &FakeRepository{}

type FakeRepository struct {
	Categories      []Category
	Entries         []Entry
	EntriesMetadata []EntryMetadata
	Feeds           []Feed
	Preferences     map[string]Preferences
	Subscriptions   []Subscription
}

func (r *FakeRepository) FeedCreate(_ context.Context, feed Feed) error {
	r.Feeds = append(r.Feeds, feed)
	return nil
}

func (r *FakeRepository) FeedDelete(_ context.Context, feedUUID string) error {
	for _, entry := range r.Entries {
		if entry.FeedUUID == feedUUID {
			r.EntriesMetadata = slices.DeleteFunc(r.EntriesMetadata, func(em EntryMetadata) bool {
				return em.EntryUID == entry.UID
			})
		}
	}

	r.Entries = slices.DeleteFunc(r.Entries, func(e Entry) bool {
		return e.FeedUUID == feedUUID
	})

	r.Feeds = slices.DeleteFunc(r.Feeds, func(f Feed) bool {
		return f.UUID == feedUUID
	})

	return nil
}

func (r *FakeRepository) FeedGetByURL(_ context.Context, feedURL string) (Feed, error) {
	for _, f := range r.Feeds {
		if f.FeedURL == feedURL {
			return f, nil
		}
	}

	return Feed{}, ErrFeedNotFound
}

func (r *FakeRepository) FeedGetBySlug(_ context.Context, feedSlug string) (Feed, error) {
	for _, feed := range r.Feeds {
		if feed.Slug == feedSlug {
			return feed, nil
		}
	}

	return Feed{}, ErrFeedNotFound
}

func (r *FakeRepository) FeedCategoryCreate(_ context.Context, category Category) error {
	r.Categories = append(r.Categories, category)
	return nil
}

func (r *FakeRepository) FeedCategoryDelete(ctx context.Context, userUUID string, categoryUUID string) error {
	r.Categories = slices.DeleteFunc(r.Categories, func(c Category) bool {
		return c.UserUUID == userUUID && c.UUID == categoryUUID
	})

	var subscriptionUUIDs []string
	for _, subscription := range r.Subscriptions {
		if subscription.UserUUID == userUUID && subscription.CategoryUUID == categoryUUID {
			subscriptionUUIDs = append(subscriptionUUIDs, subscription.UUID)
		}
	}

	for _, subscriptionUUID := range subscriptionUUIDs {
		if err := r.FeedSubscriptionDelete(ctx, userUUID, subscriptionUUID); err != nil {
			return err
		}
	}

	return nil
}

func (r *FakeRepository) FeedCategoryGetByName(_ context.Context, userUUID string, name string) (Category, error) {
	for _, category := range r.Categories {
		if category.UserUUID == userUUID && category.Name == name {
			return category, nil
		}
	}

	return Category{}, ErrCategoryNotFound
}

func (r *FakeRepository) FeedCategoryGetBySlug(_ context.Context, userUUID string, slug string) (Category, error) {
	for _, category := range r.Categories {
		if category.UserUUID == userUUID && category.Slug == slug {
			return category, nil
		}
	}

	return Category{}, ErrCategoryNotFound
}

func (r *FakeRepository) FeedCategoryGetByUUID(_ context.Context, userUUID string, categoryUUID string) (Category, error) {
	for _, category := range r.Categories {
		if category.UserUUID == userUUID && category.UUID == categoryUUID {
			return category, nil
		}
	}

	return Category{}, ErrCategoryNotFound
}

func (r *FakeRepository) FeedCategoryGetMany(_ context.Context, userUUID string) ([]Category, error) {
	panic("unimplemented")
}

func (r *FakeRepository) FeedCategoryNameAndSlugAreRegistered(_ context.Context, userUUID string, name string, slug string) (bool, error) {
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

func (r *FakeRepository) FeedCategoryNameAndSlugAreRegisteredToAnotherCategory(_ context.Context, userUUID string, categoryUUID string, name string, slug string) (bool, error) {
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

func (r *FakeRepository) FeedCategoryUpdate(_ context.Context, category Category) error {
	for index, c := range r.Categories {
		if c.UserUUID == category.UserUUID && c.UUID == category.UUID {
			r.Categories[index] = category
		}
	}
	return nil
}

func (r *FakeRepository) FeedEntryCreateMany(_ context.Context, entries []Entry) (int64, error) {
	r.Entries = append(r.Entries, entries...)
	return int64(len(entries)), nil
}

func (r *FakeRepository) feedEntryExists(entryUID string) bool {
	for _, entry := range r.Entries {
		if entry.UID == entryUID {
			return true
		}
	}

	return false
}

func (r *FakeRepository) FeedEntryMarkAllAsRead(_ context.Context, userUUID string) error {
	return errors.New("not implemented")
}

func (r *FakeRepository) FeedEntryMarkAllAsReadByCategory(_ context.Context, userUUID string, categoryUUID string) error {
	return errors.New("not implemented")
}

func (r *FakeRepository) FeedEntryMarkAllAsReadBySubscription(_ context.Context, userUUID string, subscriptionUUID string) error {
	return errors.New("not implemented")
}

func (r *FakeRepository) FeedEntryMetadataCreate(_ context.Context, newEntryMetadata EntryMetadata) error {
	if !r.feedEntryExists(newEntryMetadata.EntryUID) {
		return ErrEntryNotFound
	}

	r.EntriesMetadata = append(r.EntriesMetadata, newEntryMetadata)
	return nil
}

func (r *FakeRepository) FeedEntryMetadataGetByUID(_ context.Context, userUUID string, entryUID string) (EntryMetadata, error) {
	for _, entryMetadata := range r.EntriesMetadata {
		if entryMetadata.UserUUID == userUUID && entryMetadata.EntryUID == entryUID {
			return entryMetadata, nil
		}
	}

	return EntryMetadata{}, ErrEntryMetadataNotFound
}

func (r *FakeRepository) FeedEntryMetadataUpdate(_ context.Context, updatedEntryMetadata EntryMetadata) error {
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

func (r *FakeRepository) FeedPreferencesGetByUserUUID(_ context.Context, userUUID string) (Preferences, error) {
	preferences, ok := r.Preferences[userUUID]
	if !ok {
		return Preferences{}, user.ErrNotFound
	}

	return preferences, nil
}

func (r *FakeRepository) FeedPreferencesUpdate(_ context.Context, preferences Preferences) error {
	_, ok := r.Preferences[preferences.UserUUID]
	if !ok {
		return user.ErrNotFound
	}

	r.Preferences[preferences.UserUUID] = preferences

	return nil
}

func (r *FakeRepository) FeedSubscriptionCountByFeed(_ context.Context, feedUUID string) (uint, error) {
	var count uint
	for _, s := range r.Subscriptions {
		if s.FeedUUID == feedUUID {
			count++
		}
	}

	return count, nil
}

func (r *FakeRepository) FeedSubscriptionIsRegistered(_ context.Context, userUUID string, feedUUID string) (bool, error) {
	for _, s := range r.Subscriptions {
		if s.UserUUID == userUUID && s.FeedUUID == feedUUID {
			return true, nil
		}
	}

	return false, nil
}

func (r *FakeRepository) FeedSubscriptionCreate(_ context.Context, subscription Subscription) (Subscription, error) {
	r.Subscriptions = append(r.Subscriptions, subscription)
	return subscription, nil
}

func (r *FakeRepository) FeedSubscriptionDelete(ctx context.Context, userUUID string, subscriptionUUID string) error {
	if _, err := r.FeedSubscriptionGetByUUID(ctx, userUUID, subscriptionUUID); err != nil {
		return err
	}

	// 1. Delete the subscription
	r.Subscriptions = slices.DeleteFunc(r.Subscriptions, func(s Subscription) bool {
		return s.UserUUID == userUUID && s.UUID == subscriptionUUID
	})

	// 2. Propagate the deletion to feeds that are not referenced anymore
	var deletedFeedUUIDs []string
	r.Feeds = slices.DeleteFunc(r.Feeds, func(f Feed) bool {
		for _, s := range r.Subscriptions {
			if s.FeedUUID == f.UUID {
				// This feed is referenced by subscriptions for other users
				return false
			}
		}

		deletedFeedUUIDs = append(deletedFeedUUIDs, f.UUID)

		return true
	})

	var deletedEntryUIDs []string
	r.Entries = slices.DeleteFunc(r.Entries, func(e Entry) bool {
		if slices.Contains(deletedFeedUUIDs, e.FeedUUID) {
			deletedEntryUIDs = append(deletedEntryUIDs, e.UID)
			return true
		}

		return false
	})

	r.EntriesMetadata = slices.DeleteFunc(r.EntriesMetadata, func(em EntryMetadata) bool {
		return slices.Contains(deletedEntryUIDs, em.EntryUID)
	})

	return nil
}

func (r *FakeRepository) FeedSubscriptionGetByFeed(_ context.Context, userUUID string, feedUUID string) (Subscription, error) {
	for _, subscription := range r.Subscriptions {
		if subscription.UserUUID == userUUID && subscription.FeedUUID == feedUUID {
			return subscription, nil
		}
	}

	return Subscription{}, ErrSubscriptionNotFound
}

func (r *FakeRepository) FeedSubscriptionGetByUUID(_ context.Context, userUUID string, subscriptionUUID string) (Subscription, error) {
	for _, subscription := range r.Subscriptions {
		if subscription.UserUUID == userUUID && subscription.UUID == subscriptionUUID {
			return subscription, nil
		}
	}

	return Subscription{}, ErrSubscriptionNotFound
}

func (r *FakeRepository) FeedSubscriptionUpdate(_ context.Context, subscription Subscription) error {
	for index, c := range r.Subscriptions {
		if c.UserUUID == subscription.UserUUID && c.UUID == subscription.UUID {
			r.Subscriptions[index] = subscription
			return nil
		}
	}

	return ErrSubscriptionNotFound
}
