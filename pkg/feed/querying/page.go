// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import "github.com/virtualtam/sparklemuffin/pkg/feed"

type Feed struct {
	feed.Feed
	Unread uint
}

type Category struct {
	feed.Category
	Unread uint

	Feeds []Feed
}

type FeedEntries struct {
	// TODO: only names and slugs are needed
	// TODO: special case for all/all
	Category feed.Category
	Feed     feed.Feed

	Entries []feed.Entry
}

// A FeedPage holds a set of paginated Feeds.
type FeedPage struct {
	Categories  []Category
	FeedEntries FeedEntries
}
