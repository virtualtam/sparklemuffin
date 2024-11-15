// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import "github.com/virtualtam/sparklemuffin/pkg/feed"

type SubscribedFeedsByCategory struct {
	feed.Category

	Unread uint

	SubscribedFeeds []SubscribedFeed
}

type SubscribedFeed struct {
	feed.Feed

	Unread uint
}

type SubscribedFeedEntry struct {
	feed.Entry

	Read bool
}

type SubscriptionTitle struct {
	SubscriptionUUID string
	FeedTitle        string
}

type SubscriptionsTitlesByCategory struct {
	feed.Category

	SubscriptionTitles []SubscriptionTitle
}
