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
	FeedEntryGetCountBySubscription(userUUID string, subscriptionUUID string) (uint, error)

	// FeedEntryGetCountByQuery returns the count of entries corresponding to a feed subscription
	// for a giver user, and matching a search query.
	FeedEntryGetCountByQuery(userUUID string, searchTerms string) (uint, error)

	// FeedEntryGetCountByCategoryAndQuery returns the count of entries corresponding to a feed subscription
	// for a giver user and category, and matching a search query.
	FeedEntryGetCountByCategoryAndQuery(userUUID string, categoryUUID string, query string) (uint, error)

	// FeedEntryGetCountBySubscriptionAndQuery returns the count of entries corresponding to a feed subscription
	// for a giver user and subscription, and matching a search query.
	FeedEntryGetCountBySubscriptionAndQuery(userUUID string, subscriptionUUID string, query string) (uint, error)

	// FeedSubscriptionCategoryGetAll returns SubscribedFeeds, sorted by SubscriptionCategory.
	FeedSubscriptionCategoryGetAll(userUUID string) ([]SubscribedFeedsByCategory, error)

	// FeedSubscriptionEntryGetN returns at most n SubscriptionEntries, starting at a given offset.
	FeedSubscriptionEntryGetN(userUUID string, entriesPerPage uint, offset uint) ([]SubscribedFeedEntry, error)

	// FeedSubscriptionEntryGetNByCategory returns at most n SubscriptionEntries, starting at a given offset.
	FeedSubscriptionEntryGetNByCategory(userUUID string, categoryUUID string, entriesPerPage uint, offset uint) ([]SubscribedFeedEntry, error)

	// FeedSubscriptionEntryGetNBySubscription returns at most n SubscriptionEntries, starting at a given offset.
	FeedSubscriptionEntryGetNBySubscription(userUUID string, subscriptionUUID string, entriesPerPage uint, offset uint) ([]SubscribedFeedEntry, error)

	// FeedSubscriptionEntryGetNByQuery returns at most n SubscriptionEntries matching a search query, starting at a given offset.
	FeedSubscriptionEntryGetNByQuery(userUUID string, searchTerms string, entriesPerPage uint, offset uint) ([]SubscribedFeedEntry, error)

	// FeedSubscriptionEntryGetNByCategoryAndQuery returns at most n SubscriptionEntries matching a search query, starting at a given offset.
	FeedSubscriptionEntryGetNByCategoryAndQuery(userUUID string, categoryUUID string, query string, entriesPerPage uint, offset uint) ([]SubscribedFeedEntry, error)

	// FeedSubscriptionEntryGetNBySubscriptionAndQuery returns at most n SubscriptionEntries matching a search query, starting at a given offset.
	FeedSubscriptionEntryGetNBySubscriptionAndQuery(userUUID string, subscriptionUUID string, query string, entriesPerPage uint, offset uint) ([]SubscribedFeedEntry, error)

	// FeedQueryingSubscriptionByUUID returns feed subscription metadata for a given user and subscription.
	FeedQueryingSubscriptionByUUID(userUUID string, subscriptionUUID string) (Subscription, error)

	// FeedQueryingSubscriptionsByCategory returns a list of feed Subscription titles, sorted by Category.
	FeedQueryingSubscriptionsByCategory(userUUID string) ([]SubscriptionsByCategory, error)
}
