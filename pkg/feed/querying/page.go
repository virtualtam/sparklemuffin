// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package querying

import (
	"encoding/base64"

	"github.com/virtualtam/sparklemuffin/internal/paginate"
)

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

// A Tag holds metadata for a given Feed subscription tag.
type Tag struct {
	Name        string
	EncodedName string
	Count       uint
}

// NewTag initializes and returns a new Tag.
func NewTag(name string, count uint) Tag {
	return Tag{
		Name:        name,
		EncodedName: base64.URLEncoding.EncodeToString([]byte(name)),
		Count:       count,
	}
}

// A TagPage holds a set of paginated Feed subscription tags.
type TagPage struct {
	paginate.Page

	Tags []Tag
}

// NewTagPage initializes and returns a new TagPage.
func NewTagPage(number uint, totalPages uint, tagCount uint, tags []Tag) TagPage {
	page := TagPage{
		Page: paginate.NewPage(number, totalPages, tagsPerPage, tagCount),
		Tags: tags,
	}

	return page
}

// NewTagSearchResultPage initializes and returns a new TagPage containing search results.
func NewTagSearchResultPage(searchTerms string, tagCount uint, number uint, totalPages uint, tags []Tag) TagPage {
	page := NewTagPage(number, totalPages, tagCount, tags)
	page.SearchTerms = searchTerms

	return page
}
