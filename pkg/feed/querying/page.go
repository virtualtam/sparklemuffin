// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import "github.com/virtualtam/sparklemuffin/pkg/feed"

type SubscribedFeed struct {
	feed.Feed
	Unread uint
}

type Category struct {
	feed.Category
	Unread uint

	SubscribedFeeds []SubscribedFeed
}

// A FeedPage holds a set of paginated Feeds.
type FeedPage struct {
	Categories []Category
	Entries    []feed.Entry
}
