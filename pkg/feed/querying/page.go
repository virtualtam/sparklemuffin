// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import "github.com/virtualtam/sparklemuffin/internal/paginate"

// A FeedPage holds a set of paginated Feeds.
type FeedPage struct {
	paginate.Page

	PageTitle   string
	Description string
	Unread      uint
	Categories  []SubscribedFeedsByCategory
	Entries     []SubscribedFeedEntry
}

// NewFeedPage initializes and returns a new FeedPage.
func NewFeedPage(number uint, totalPages uint, pageTitle string, description string, categories []SubscribedFeedsByCategory, totalEntryCount uint, entries []SubscribedFeedEntry) FeedPage {
	var unread uint

	for _, category := range categories {
		unread += category.Unread
	}

	page := FeedPage{
		Page:        paginate.NewPage(number, totalPages, entriesPerPage, totalEntryCount),
		PageTitle:   pageTitle,
		Description: description,
		Unread:      unread,
		Categories:  categories,
		Entries:     entries,
	}

	return page
}

// NewFeedSearchResultPage initializes and returns a new FeedPage containing search results.
func NewFeedSearchResultPage(searchTerms string, searchResultCount uint, number uint, totalPages uint, pageTitle, description string, categories []SubscribedFeedsByCategory, entries []SubscribedFeedEntry) FeedPage {
	page := NewFeedPage(number, totalPages, pageTitle, description, categories, searchResultCount, entries)
	page.SearchTerms = searchTerms

	return page
}
