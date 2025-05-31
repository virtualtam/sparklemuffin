// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

// A FeedPage holds a set of paginated Feeds.
type FeedPage struct {
	PageNumber         uint
	PreviousPageNumber uint
	NextPageNumber     uint
	TotalPages         uint
	PagesLeft          uint
	Offset             uint

	SearchTerms string

	PageTitle       string
	Description     string
	Unread          uint
	Categories      []SubscribedFeedsByCategory
	TotalEntryCount uint
	Entries         []SubscribedFeedEntry
}

// NewFeedPage initializes and returns a new FeedPage.
func NewFeedPage(number uint, totalPages uint, pageTitle string, description string, categories []SubscribedFeedsByCategory, totalEntryCount uint, entries []SubscribedFeedEntry) FeedPage {
	var unread uint

	for _, category := range categories {
		unread += category.Unread
	}

	page := FeedPage{
		PageNumber:      number,
		TotalPages:      totalPages,
		PagesLeft:       totalPages - number,
		PageTitle:       pageTitle,
		Description:     description,
		Unread:          unread,
		Categories:      categories,
		TotalEntryCount: totalEntryCount,
		Entries:         entries,
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

// NewFeedSearchResultPage initializes and returns a new FeedPage containing search results.
func NewFeedSearchResultPage(searchTerms string, searchResultCount uint, number uint, totalPages uint, pageTitle, description string, categories []SubscribedFeedsByCategory, entries []SubscribedFeedEntry) FeedPage {
	page := NewFeedPage(number, totalPages, pageTitle, description, categories, searchResultCount, entries)
	page.SearchTerms = searchTerms

	return page
}
