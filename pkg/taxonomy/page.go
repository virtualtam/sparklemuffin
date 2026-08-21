// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package taxonomy

import "github.com/virtualtam/sparklemuffin/internal/paginate"

const tagsPerPage uint = 90

// TagPage holds a set of paginated Tags.
type TagPage struct {
	paginate.Page

	Tags []Tag
}

// NewTagPage initializes and returns a new TagPage.
func NewTagPage(number, totalPages, tagCount uint, tags []Tag) TagPage {
	return TagPage{
		Page: paginate.NewPage(number, totalPages, tagsPerPage, tagCount),
		Tags: tags,
	}
}

// NewTagSearchResultPage initializes and returns a new TagPage containing search results.
func NewTagSearchResultPage(searchTerms string, tagCount, number, totalPages uint, tags []Tag) TagPage {
	page := NewTagPage(number, totalPages, tagCount, tags)
	page.SearchTerms = searchTerms

	return page
}
