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

	SubscriptionAlias string
	FeedTitle         string
	Read              bool
}

type Subscription struct {
	UUID         string
	CategoryUUID string
	Alias        string

	FeedTitle       string
	FeedDescription string
}

type SubscriptionsByCategory struct {
	feed.Category

	Subscriptions []Subscription
}
