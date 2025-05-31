// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package paginate

// A Page holds pagination metadata.
type Page struct {
	// The number of the current page.
	PageNumber uint

	// The number of the previous page.
	PreviousPageNumber uint

	// The number of the next page.
	NextPageNumber uint

	// The total number of pages.
	TotalPages uint

	// The number of pages left.
	PagesLeft uint

	// The total number of items.
	ItemCount uint

	// The offset of the first item on the current page.
	ItemOffset uint

	// Terms used in a search query.
	SearchTerms string
}

// NewPage initializes and returns a new Page.
func NewPage(pageNumber uint, totalPages uint, itemsPerPage uint, itemCount uint) Page {
	page := Page{
		PageNumber: pageNumber,
		TotalPages: totalPages,
		PagesLeft:  totalPages - pageNumber,
		ItemCount:  itemCount,
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

	page.ItemOffset = (page.PageNumber-1)*itemsPerPage + 1

	return page
}
