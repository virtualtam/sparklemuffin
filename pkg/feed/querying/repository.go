// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

// Repository provides access to user feed subscriptions for querying.
type Repository interface {
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
	FeedSubscriptionCategoryGetAll(userUUID string) ([]SubscriptionCategory, error)

	// FeedSubscriptionEntryGetN returns at most n SubscriptionEntries, starting at a given offset.
	FeedSubscriptionEntryGetN(userUUID string, entriesPerPage uint, offset uint) ([]SubscriptionEntry, error)

	// FeedSubscriptionEntryGetNByCategory returns at most n SubscriptionEntries, starting at a given offset.
	FeedSubscriptionEntryGetNByCategory(userUUID string, categoryUUID string, entriesPerPage uint, offset uint) ([]SubscriptionEntry, error)

	// FeedSubscriptionEntryGetNBySubscription returns at most n SubscriptionEntries, starting at a given offset.
	FeedSubscriptionEntryGetNBySubscription(userUUID string, subscriptionUUID string, entriesPerPage uint, offset uint) ([]SubscriptionEntry, error)
}
