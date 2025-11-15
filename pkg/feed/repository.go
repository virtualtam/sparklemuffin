// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"context"
)

// ValidationRepository provides methods for Feed and Subscription validation.
type ValidationRepository interface {
	// FeedCategoryNameAndSlugAreRegistered returns whether a user has already registered
	// a Category with the same name or slug.
	FeedCategoryNameAndSlugAreRegistered(ctx context.Context, userUUID string, name string, slug string) (bool, error)

	// FeedCategoryNameAndSlugAreRegisteredToAnotherCategory returns whether a user has already registered
	// another Category with the same name or slug.
	FeedCategoryNameAndSlugAreRegisteredToAnotherCategory(ctx context.Context, userUUID string, categoryUUID string, name string, slug string) (bool, error)

	// FeedSubscriptionIsRegistered returns whether a user has already registered
	// a Subscription to a given Feed.
	FeedSubscriptionIsRegistered(ctx context.Context, userUUID string, feedUUID string) (bool, error)
}

// Repository provides access to user feeds.
type Repository interface {
	ValidationRepository

	// FeedCreate creates a new Feed.
	FeedCreate(ctx context.Context, feed Feed) error

	// FeedGetBySlug returns the Feed for a given slug.
	FeedGetBySlug(ctx context.Context, feedSlug string) (Feed, error)

	// FeedGetByURL returns the Feed for a given URL.
	FeedGetByURL(ctx context.Context, feedURL string) (Feed, error)

	// FeedCategoryCreate creates a new Category.
	FeedCategoryCreate(ctx context.Context, category Category) error

	// FeedCategoryDelete deletes an existing Category and related Subscriptions.
	FeedCategoryDelete(ctx context.Context, userUUID string, categoryUUID string) error

	// FeedCategoryGetByName returns the Category for a given user and name.
	FeedCategoryGetByName(ctx context.Context, userUUID string, name string) (Category, error)

	// FeedCategoryGetBySlug returns the Category for a given user and slug.
	FeedCategoryGetBySlug(ctx context.Context, userUUID string, slug string) (Category, error)

	// FeedCategoryGetByUUID returns the Category for a given user and UUID.
	FeedCategoryGetByUUID(ctx context.Context, userUUID string, categoryUUID string) (Category, error)

	// FeedCategoryGetMany returns all categories for a giver user.
	FeedCategoryGetMany(ctx context.Context, userUUID string) ([]Category, error)

	// FeedCategoryUpdate updates an existing Category.
	FeedCategoryUpdate(ctx context.Context, category Category) error

	// FeedEntryCreateMany creates a collection of new Entries.
	FeedEntryCreateMany(ctx context.Context, entries []Entry) (int64, error)

	// FeedEntryMarkAllAsRead marks all entries as "read" for a given User.
	FeedEntryMarkAllAsRead(ctx context.Context, userUUID string) error

	// FeedEntryMarkAllAsReadByCategory marks all entries as "read" for a given User and Category.
	FeedEntryMarkAllAsReadByCategory(ctx context.Context, userUUID string, categoryUUID string) error

	// FeedEntryMarkAllAsReadBySubscription marks all entries as "read" for a given User and Subscription.
	FeedEntryMarkAllAsReadBySubscription(ctx context.Context, userUUID string, subscriptionUUID string) error

	// FeedEntryMetadataCreate creates a new EntryStatus.
	FeedEntryMetadataCreate(ctx context.Context, entryMetadata EntryMetadata) error

	// FeedEntryMetadataGetByUID returns the EntryStatus for a given user and Entry.
	FeedEntryMetadataGetByUID(ctx context.Context, userUUID string, entryUID string) (EntryMetadata, error)

	// FeedEntryMetadataUpdate updates an existing EntryStatus.
	FeedEntryMetadataUpdate(ctx context.Context, entryMetadata EntryMetadata) error

	// FeedPreferencesGetByUserUUID returns a user's feed Preferences.
	FeedPreferencesGetByUserUUID(ctx context.Context, userUUID string) (Preferences, error)

	// FeedPreferencesUpdate updates a user's feed Preferences.
	FeedPreferencesUpdate(ctx context.Context, preferences Preferences) error

	// FeedSubscriptionCreate creates a new Feed subscription for a given user.
	FeedSubscriptionCreate(ctx context.Context, subscription Subscription) (Subscription, error)

	// FeedSubscriptionDelete deletes a given Feed subscription.
	FeedSubscriptionDelete(ctx context.Context, userUUID string, subscriptionUUID string) error

	// FeedSubscriptionGetByFeed returns the Subscription for a given user and feed.
	FeedSubscriptionGetByFeed(ctx context.Context, userUUID string, feedUUID string) (Subscription, error)

	// FeedSubscriptionGetByUUID returns the Subscription for a given user and UUID.
	FeedSubscriptionGetByUUID(ctx context.Context, userUUID string, subscriptionUUID string) (Subscription, error)

	// FeedSubscriptionUpdate updates an existing Subscription.
	FeedSubscriptionUpdate(ctx context.Context, subscription Subscription) error
}
