// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import (
	"context"

	"github.com/virtualtam/sparklemuffin/pkg/feed"
)

// Repository provides access to user feed subscriptions for querying.
type Repository interface {
	// FeedGetByUUID returns a given feed.
	FeedGetByUUID(ctx context.Context, feedUUID string) (feed.Feed, error)

	// FeedEntryGetCount returns the count of entries corresponding to a feed subscription
	// for a giver user.
	FeedEntryGetCount(ctx context.Context, userUUID string, showEntries feed.EntryVisibility) (uint, error)

	// FeedEntryGetCountByCategory returns the count of entries corresponding to a feed subscription
	// for a giver user and category.
	FeedEntryGetCountByCategory(ctx context.Context, userUUID string, showEntries feed.EntryVisibility, categoryUUID string) (uint, error)

	// FeedEntryGetCountBySubscription returns the count of entries corresponding to a feed subscription
	// for a giver user and subscription.
	FeedEntryGetCountBySubscription(ctx context.Context, userUUID string, showEntries feed.EntryVisibility, subscriptionUUID string) (uint, error)

	// FeedEntryGetCountByQuery returns the count of entries corresponding to a feed subscription
	// for a giver user, and matching a search query.
	FeedEntryGetCountByQuery(ctx context.Context, userUUID string, showEntries feed.EntryVisibility, searchTerms string) (uint, error)

	// FeedEntryGetCountByCategoryAndQuery returns the count of entries corresponding to a feed subscription
	// for a giver user and category, and matching a search query.
	FeedEntryGetCountByCategoryAndQuery(ctx context.Context, userUUID string, showEntries feed.EntryVisibility, categoryUUID string, query string) (uint, error)

	// FeedEntryGetCountBySubscriptionAndQuery returns the count of entries corresponding to a feed subscription
	// for a giver user and subscription, and matching a search query.
	FeedEntryGetCountBySubscriptionAndQuery(ctx context.Context, userUUID string, showEntries feed.EntryVisibility, subscriptionUUID string, query string) (uint, error)

	// FeedSubscriptionCategoryGetAll returns SubscribedFeeds, sorted by SubscriptionCategory.
	FeedSubscriptionCategoryGetAll(ctx context.Context, userUUID string) ([]SubscribedFeedsByCategory, error)

	// FeedSubscriptionEntryGetN returns at most n SubscriptionEntries, starting at a given offset.
	FeedSubscriptionEntryGetN(ctx context.Context, userUUID string, preferences feed.Preferences, entriesPerPage uint, offset uint) ([]SubscribedFeedEntry, error)

	// FeedSubscriptionEntryGetNByCategory returns at most n SubscriptionEntries, starting at a given offset.
	FeedSubscriptionEntryGetNByCategory(ctx context.Context, userUUID string, preferences feed.Preferences, categoryUUID string, entriesPerPage uint, offset uint) ([]SubscribedFeedEntry, error)

	// FeedSubscriptionEntryGetNBySubscription returns at most n SubscriptionEntries, starting at a given offset.
	FeedSubscriptionEntryGetNBySubscription(ctx context.Context, userUUID string, preferences feed.Preferences, subscriptionUUID string, entriesPerPage uint, offset uint) ([]SubscribedFeedEntry, error)

	// FeedSubscriptionEntryGetNByQuery returns at most n SubscriptionEntries matching a search query, starting at a given offset.
	FeedSubscriptionEntryGetNByQuery(ctx context.Context, userUUID string, preferences feed.Preferences, searchTerms string, entriesPerPage uint, offset uint) ([]SubscribedFeedEntry, error)

	// FeedSubscriptionEntryGetNByCategoryAndQuery returns at most n SubscriptionEntries matching a search query, starting at a given offset.
	FeedSubscriptionEntryGetNByCategoryAndQuery(ctx context.Context, userUUID string, preferences feed.Preferences, categoryUUID string, query string, entriesPerPage uint, offset uint) ([]SubscribedFeedEntry, error)

	// FeedSubscriptionEntryGetNBySubscriptionAndQuery returns at most n SubscriptionEntries matching a search query, starting at a given offset.
	FeedSubscriptionEntryGetNBySubscriptionAndQuery(ctx context.Context, userUUID string, preferences feed.Preferences, subscriptionUUID string, query string, entriesPerPage uint, offset uint) ([]SubscribedFeedEntry, error)

	// FeedQueryingSubscriptionByUUID returns feed subscription metadata for a given user and subscription.
	FeedQueryingSubscriptionByUUID(ctx context.Context, userUUID string, subscriptionUUID string) (Subscription, error)

	// FeedQueryingSubscriptionsByCategory returns a list of feed Subscription titles, sorted by Category.
	FeedQueryingSubscriptionsByCategory(ctx context.Context, userUUID string) ([]SubscriptionsByCategory, error)
}
