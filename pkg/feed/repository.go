// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

// ValidationRepository provides methods for Feed and Subscription validation.
type ValidationRepository interface {
	// FeedIsSubscriptionRegistered returns whether a user has already registered
	// a Subscription to a given Feed.
	FeedIsSubscriptionRegistered(userUUID string, feedUUID string) (bool, error)
}

// Repository provides access to user feeds.
type Repository interface {
	ValidationRepository

	// FeedCreate creates a new Feed.
	FeedCreate(feed Feed) error

	// FeedGetByURL returns the Feed for a given URL.
	FeedGetByURL(feedURL string) (Feed, error)

	// FeedGetCategories returns all categories for a giver user.
	FeedGetCategories(userUUID string) ([]Category, error)

	// FeedSubscriptionCreate creates a new Feed subscription for a given user.
	FeedSubscriptionCreate(subscription Subscription) error
}
