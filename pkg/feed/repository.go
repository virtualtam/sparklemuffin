// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

// ValidationRepository provides methods for Feed and Subscription validation.
type ValidationRepository interface {
	// FeedCategoryNameAndSlugAreRegistered returns whether a user has already registered
	// a Category with the same name or slug.
	FeedCategoryNameAndSlugAreRegistered(userUUID string, name string, slug string) (bool, error)

	// FeedCategoryNameAndSlugAreRegistered returns whether a user has already registered
	// another Category with the same name or slug.
	FeedCategoryNameAndSlugAreRegisteredToAnotherCategory(userUUID string, categoryUUID string, name string, slug string) (bool, error)

	// FeedSubscriptionIsRegistered returns whether a user has already registered
	// a Subscription to a given Feed.
	FeedSubscriptionIsRegistered(userUUID string, feedUUID string) (bool, error)
}

// Repository provides access to user feeds.
type Repository interface {
	ValidationRepository

	// FeedAdd creates a new Feed.
	FeedAdd(feed Feed) error

	// FeedGetBySlug returns the Feed for a given slug.
	FeedGetBySlug(feedSlug string) (Feed, error)

	// FeedGetByURL returns the Feed for a given URL.
	FeedGetByURL(feedURL string) (Feed, error)

	// FeedCategoryAdd creates a new Category.
	FeedCategoryAdd(category Category) error

	// FeedCategoryDelete deletes an existing Category and related Subscriptions.
	FeedCategoryDelete(userUUID string, categoryUUID string) error

	// FeedCategoryGetBySlug returns the Category for a given user and slug.
	FeedCategoryGetBySlug(userUUID string, slug string) (Category, error)

	// FeedCategoryGetByUUID returns the Category for a given user and UUID.
	FeedCategoryGetByUUID(userUUID string, categoryUUID string) (Category, error)

	// FeedCategoryGetMany returns all categories for a giver user.
	FeedCategoryGetMany(userUUID string) ([]Category, error)

	// FeedCategoryUpdate updates an existing Category.
	FeedCategoryUpdate(category Category) error

	// FeedEntryAddMany creates a collection of new Entries.
	FeedEntryAddMany(entries []Entry) (int64, error)

	// FeedEntryGetN returns at most N entries for a given Feed.
	FeedEntryGetN(feedUUID string, n uint) ([]Entry, error)

	// FeedSubscriptionAdd creates a new Feed subscription for a given user.
	FeedSubscriptionAdd(subscription Subscription) error

	// FeedSubscriptionGetByFeed returns the Subscription for a giver user and feed.
	FeedSubscriptionGetByFeed(userUUID string, feedUUID string) (Subscription, error)
}
