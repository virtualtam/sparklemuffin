// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import "github.com/virtualtam/sparklemuffin/pkg/feed"

// Repository provides access to user feed subscriptions for querying.
type Repository interface {
	// FeedGetByUUID returns a given feed.
	FeedGetByUUID(feedUUID string) (feed.Feed, error)

	// FeedEntryGetCount returns the count of entries corresponding to a feed subscription
	// for a giver user.
	FeedEntryGetCount(userUUID string) (uint, error)

	// FeedEntryGetCountByCategory returns the count of entries corresponding to a feed subscription
	// for a giver user and category.
	FeedEntryGetCountByCategory(userUUID string, categoryUUID string) (uint, error)

	// FeedEntryGetCountBySubscription returns the count of entries corresponding to a feed subscription
	// for a giver user and subscription.
	FeedEntryGetCountBySubscription(userUUID string, feedUUID string) (uint, error)

	// FeedSubscriptionCategoryGetAll returns SubscribedFeeds, sorted by SubscriptionCategory.
	FeedSubscriptionCategoryGetAll(userUUID string) ([]SubscribedFeedsByCategory, error)

	// FeedSubscriptionEntryGetN returns at most n SubscriptionEntries, starting at a given offset.
	FeedSubscriptionEntryGetN(userUUID string, entriesPerPage uint, offset uint) ([]SubscribedFeedEntry, error)

	// FeedSubscriptionEntryGetNByCategory returns at most n SubscriptionEntries, starting at a given offset.
	FeedSubscriptionEntryGetNByCategory(userUUID string, categoryUUID string, entriesPerPage uint, offset uint) ([]SubscribedFeedEntry, error)

	// FeedSubscriptionEntryGetNBySubscription returns at most n SubscriptionEntries, starting at a given offset.
	FeedSubscriptionEntryGetNBySubscription(userUUID string, subscriptionUUID string, entriesPerPage uint, offset uint) ([]SubscribedFeedEntry, error)

	// FeedSubscriptionTitleByUUID returns feed subscription metadata for a given user and subscription.
	FeedSubscriptionTitleByUUID(userUUID string, subscriptionUUID string) (SubscriptionTitle, error)

	// FeedSubscriptionTitlesByCategory returns a list of feed Subscription titles, sorted by Category.
	FeedSubscriptionTitlesByCategory(userUUID string) ([]SubscriptionsTitlesByCategory, error)
}
