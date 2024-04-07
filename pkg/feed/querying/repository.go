// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import "github.com/virtualtam/sparklemuffin/pkg/feed"

type Repository interface {
	FeedGetSubscriptionsByCategories(userUUID string) ([]Category, error)

	FeedGetEntriesByPage(userUUID string) ([]feed.Entry, error)
}
