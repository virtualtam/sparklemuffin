package querying

import (
	"math"

	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
)

// A Page holds a set of paginated bookmarks.
type Page struct {
	// Owner exposes public matadata for the User owning the bookmarks.
	Owner Owner

	PageNumber         uint
	PreviousPageNumber uint
	NextPageNumber     uint
	TotalPages         uint
	Offset             uint

	SearchTerms       string
	SearchResultCount uint

	Bookmarks []bookmark.Bookmark
}

// NewPage initializes and returns a new bookmark Page.
func NewPage(owner Owner, number uint, totalPages uint, bookmarks []bookmark.Bookmark) Page {
	page := Page{
		Owner:      owner,
		PageNumber: number,
		TotalPages: totalPages,
		Bookmarks:  bookmarks,
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

	page.Offset = (page.PageNumber-1)*bookmarksPerPage + 1

	return page
}

// NewSearchResultPage initializes and returns a new bookmark Page containing search results.
func NewSearchResultPage(owner Owner, searchTerms string, searchResultCount uint, number uint, totalPages uint, bookmarks []bookmark.Bookmark) Page {
	page := NewPage(owner, number, totalPages, bookmarks)

	page.SearchTerms = searchTerms
	page.SearchResultCount = searchResultCount

	return page
}

func pageCount(bookmarkCount, bookmarksPerPage uint) uint {
	if bookmarkCount == 0 {
		return 1
	}

	return uint(math.Ceil(float64(bookmarkCount) / float64(bookmarksPerPage)))
}
