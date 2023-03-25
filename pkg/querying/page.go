package querying

import (
	"math"

	"github.com/virtualtam/yawbe/pkg/bookmark"
)

// A Page holds a set of paginated bookmarks.
type Page struct {
	// Owner exposes public matadata for the User owning the bookmarks.
	Owner Owner

	PageNumber         int
	PreviousPageNumber int
	NextPageNumber     int
	TotalPages         int
	Offset             int

	SearchTerms       string
	SearchResultCount int

	Bookmarks []bookmark.Bookmark
}

// NewPage initializes and returns a new bookmark Page.
func NewPage(owner Owner, number int, totalPages int, bookmarks []bookmark.Bookmark) Page {
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
func NewSearchResultPage(owner Owner, searchTerms string, searchResultCount int, number int, totalPages int, bookmarks []bookmark.Bookmark) Page {
	page := NewPage(owner, number, totalPages, bookmarks)

	page.SearchTerms = searchTerms
	page.SearchResultCount = searchResultCount

	return page
}

func pageCount(bookmarkCount, bookmarksPerPage int) int {
	if bookmarkCount == 0 {
		return 1
	}

	return int(math.Ceil(float64(bookmarkCount) / float64(bookmarksPerPage)))
}
