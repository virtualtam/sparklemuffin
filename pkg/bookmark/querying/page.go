// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package querying

import (
	"github.com/virtualtam/sparklemuffin/internal/paginate"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
)

// A BookmarkPage holds a set of paginated bookmarks.
type BookmarkPage struct {
	paginate.Page

	// Owner exposes public metadata for the User owning the bookmarks.
	Owner Owner

	Bookmarks []bookmark.Bookmark
}

// NewBookmarkPage initializes and returns a new BookmarkPage.
func NewBookmarkPage(owner Owner, number uint, totalPages uint, totalBookmarkCount uint, bookmarks []bookmark.Bookmark) BookmarkPage {
	page := BookmarkPage{
		Page:      paginate.NewPage(number, totalPages, bookmarksPerPage, totalBookmarkCount),
		Owner:     owner,
		Bookmarks: bookmarks,
	}

	return page
}

// NewBookmarkSearchResultPage initializes and returns a new BookmarkPage containing search results.
func NewBookmarkSearchResultPage(owner Owner, searchTerms string, searchResultCount uint, number uint, totalPages uint, bookmarks []bookmark.Bookmark) BookmarkPage {
	page := NewBookmarkPage(owner, number, totalPages, searchResultCount, bookmarks)
	page.SearchTerms = searchTerms

	return page
}
