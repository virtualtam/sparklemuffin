// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import (
	"github.com/virtualtam/sparklemuffin/pkg/feed"
)

type SubscribedFeedsByCategory struct {
	feed.Category

	Unread uint

	SubscribedFeeds []SubscribedFeed
}

type SubscribedFeed struct {
	feed.Feed

	Alias  string
	Unread uint
}

type SubscribedFeedEntry struct {
	feed.Entry

	FeedTitle string
	Read      bool
}

type SubscriptionTitle struct {
	SubscriptionUUID  string
	SubscriptionAlias string
	FeedTitle         string
	FeedDescription   string
}

type SubscriptionsTitlesByCategory struct {
	feed.Category

	SubscriptionTitles []SubscriptionTitle
}
