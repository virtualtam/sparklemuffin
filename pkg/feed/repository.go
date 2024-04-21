// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

// ValidationRepository provides methods for Feed and Subscription validation.
type ValidationRepository interface {
	// FeedSubscriptionIsRegistered returns whether a user has already registered
	// a Subscription to a given Feed.
	FeedSubscriptionIsRegistered(userUUID string, feedUUID string) (bool, error)
}

// Repository provides access to user feeds.
type Repository interface {
	ValidationRepository

	// FeedAdd creates a new Feed.
	FeedAdd(feed Feed) error

	// FeedGetByURL returns the Feed for a given URL.
	FeedGetByURL(feedURL string) (Feed, error)

	// FeedCategoryGetMany returns all categories for a giver user.
	FeedCategoryGetMany(userUUID string) ([]Category, error)

	// FeedEntryAddMany creates a collection of new Entries.
	FeedEntryAddMany(entries []Entry) (int64, error)

	// FeedEntryGetN returns at most N entries for a given Feed.
	FeedEntryGetN(feedUUID string, n uint) ([]Entry, error)

	// FeedSubscriptionAdd creates a new Feed subscription for a given user.
	FeedSubscriptionAdd(subscription Subscription) error
}
