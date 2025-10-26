// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

// ValidationRepository provides methods for Feed and Subscription validation.
type ValidationRepository interface {
	// FeedCategoryNameAndSlugAreRegistered returns whether a user has already registered
	// a Category with the same name or slug.
	FeedCategoryNameAndSlugAreRegistered(userUUID string, name string, slug string) (bool, error)

	// FeedCategoryNameAndSlugAreRegisteredToAnotherCategory returns whether a user has already registered
	// another Category with the same name or slug.
	FeedCategoryNameAndSlugAreRegisteredToAnotherCategory(userUUID string, categoryUUID string, name string, slug string) (bool, error)

	// FeedSubscriptionIsRegistered returns whether a user has already registered
	// a Subscription to a given Feed.
	FeedSubscriptionIsRegistered(userUUID string, feedUUID string) (bool, error)
}

// Repository provides access to user feeds.
type Repository interface {
	ValidationRepository

	// FeedCreate creates a new Feed.
	FeedCreate(feed Feed) error

	// FeedGetBySlug returns the Feed for a given slug.
	FeedGetBySlug(feedSlug string) (Feed, error)

	// FeedGetByURL returns the Feed for a given URL.
	FeedGetByURL(feedURL string) (Feed, error)

	// FeedCategoryCreate creates a new Category.
	FeedCategoryCreate(category Category) error

	// FeedCategoryDelete deletes an existing Category and related Subscriptions.
	FeedCategoryDelete(userUUID string, categoryUUID string) error

	// FeedCategoryGetByName returns the Category for a given user and name.
	FeedCategoryGetByName(userUUID string, name string) (Category, error)

	// FeedCategoryGetBySlug returns the Category for a given user and slug.
	FeedCategoryGetBySlug(userUUID string, slug string) (Category, error)

	// FeedCategoryGetByUUID returns the Category for a given user and UUID.
	FeedCategoryGetByUUID(userUUID string, categoryUUID string) (Category, error)

	// FeedCategoryGetMany returns all categories for a giver user.
	FeedCategoryGetMany(userUUID string) ([]Category, error)

	// FeedCategoryUpdate updates an existing Category.
	FeedCategoryUpdate(category Category) error

	// FeedEntryCreateMany creates a collection of new Entries.
	FeedEntryCreateMany(entries []Entry) (int64, error)

	// FeedEntryMarkAllAsRead marks all entries as "read" for a given User.
	FeedEntryMarkAllAsRead(userUUID string) error

	// FeedEntryMarkAllAsReadByCategory marks all entries as "read" for a given User and Category.
	FeedEntryMarkAllAsReadByCategory(userUUID string, categoryUUID string) error

	// FeedEntryMarkAllAsReadBySubscription marks all entries as "read" for a given User and Subscription.
	FeedEntryMarkAllAsReadBySubscription(userUUID string, subscriptionUUID string) error

	// FeedEntryMetadataCreate creates a new EntryStatus.
	FeedEntryMetadataCreate(entryMetadata EntryMetadata) error

	// FeedEntryMetadataGetByUID returns the EntryStatus for a given user and Entry.
	FeedEntryMetadataGetByUID(userUUID string, entryUID string) (EntryMetadata, error)

	// FeedEntryMetadataUpdate updates an existing EntryStatus.
	FeedEntryMetadataUpdate(entryMetadata EntryMetadata) error

	// FeedPreferencesByUserUUID returns a user's feed Preferences.
	FeedPreferencesByUserUUID(userUUID string) (Preferences, error)

	// FeedPreferencesUpdate updates a user's feed Preferences.
	FeedPreferencesUpdate(preferences Preferences) error

	// FeedSubscriptionCreate creates a new Feed subscription for a given user.
	FeedSubscriptionCreate(subscription Subscription) (Subscription, error)

	// FeedSubscriptionDelete deletes a given Feed subscription.
	FeedSubscriptionDelete(userUUID string, subscriptionUUID string) error

	// FeedSubscriptionGetByFeed returns the Subscription for a given user and feed.
	FeedSubscriptionGetByFeed(userUUID string, feedUUID string) (Subscription, error)

	// FeedSubscriptionGetByUUID returns the Subscription for a given user and UUID.
	FeedSubscriptionGetByUUID(userUUID string, subscriptionUUID string) (Subscription, error)

	// FeedSubscriptionUpdate updates an existing Subscription.
	FeedSubscriptionUpdate(subscription Subscription) error
}
