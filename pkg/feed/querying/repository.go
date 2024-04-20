// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import "github.com/virtualtam/sparklemuffin/pkg/feed"

// Repository provides access to user feed subscriptions for querying.
type Repository interface {
	// FeedCategorySubscribedFeedGetMany returns SubscribedFeeds, sorted by Category.
	FeedCategorySubscribedFeedGetMany(userUUID string) ([]Category, error)

	// FeedEntryGetManyByPage returns a paginated list of Entries.
	FeedEntryGetManyByPage(userUUID string) ([]feed.Entry, error)
}
