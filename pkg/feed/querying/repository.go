// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

// Repository provides access to user feed subscriptions for querying.
type Repository interface {
	// FeedSubscriptionCategoryGetAll returns SubscribedFeeds, sorted by SubscriptionCategory.
	FeedSubscriptionCategoryGetAll(userUUID string) ([]SubscriptionCategory, error)

	// FeedSubscriptionEntryGetCount returns the count of entries corresponding to a feed subscription
	// for a giver user.
	FeedSubscriptionEntryGetCount(userUUID string) (uint, error)

	// FeedSubscriptionEntryGetN returns at most n SubscriptionEntries, starting at a given offset.
	FeedSubscriptionEntryGetN(userUUID string, entriesPerPage uint, offset uint) ([]SubscriptionEntry, error)
}
