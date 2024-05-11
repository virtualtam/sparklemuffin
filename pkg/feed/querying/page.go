// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import (
	"github.com/virtualtam/sparklemuffin/pkg/feed"
)

// A FeedPage holds a set of paginated Feeds.
type FeedPage struct {
	PageNumber         uint
	PreviousPageNumber uint
	NextPageNumber     uint
	TotalPages         uint
	Offset             uint

	Header     string
	Unread     uint
	Categories []SubscriptionCategory
	Entries    []SubscriptionEntry
}

func NewFeedPage(number uint, totalPages uint, header string, categories []SubscriptionCategory, entries []SubscriptionEntry) FeedPage {
	var unread uint

	for _, category := range categories {
		unread += category.Unread
	}

	page := FeedPage{
		PageNumber: number,
		TotalPages: totalPages,
		Header:     header,
		Unread:     unread,
		Categories: categories,
		Entries:    entries,
	}

	if page.PageNumber == 1 {
		page.PreviousPageNumber = 1
	} else {
		page.PreviousPageNumber = page.PageNumber - 1
	}

	if page.PageNumber == page.TotalPages {
		page.NextPageNumber = page.PageNumber
	} else {
		page.NextPageNumber = page.PageNumber + 1
	}

	page.Offset = (page.PageNumber-1)*entriesPerPage + 1

	return page
}

type SubscriptionCategory struct {
	feed.Category

	Unread uint

	SubscribedFeeds []SubscribedFeed
}

type SubscribedFeed struct {
	feed.Feed

	Unread uint
}

type SubscriptionEntry struct {
	feed.Entry

	Read bool
}
