// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package paginate

import "testing"

// AssertPageEquals asserts that two pages are equal.
func AssertPageEquals(t *testing.T, got, want Page) {
	t.Helper()

	if got.SearchTerms != want.SearchTerms {
		t.Errorf("want search terms %q, got %q", want.SearchTerms, got.SearchTerms)
	}
	if got.PageNumber != want.PageNumber {
		t.Errorf("want page number %d, got %d", want.PageNumber, got.PageNumber)
	}
	if got.PreviousPageNumber != want.PreviousPageNumber {
		t.Errorf("want previous page number %d, got %d", want.PreviousPageNumber, got.PreviousPageNumber)
	}
	if got.NextPageNumber != want.NextPageNumber {
		t.Errorf("want next page number %d, got %d", want.NextPageNumber, got.NextPageNumber)
	}
	if got.TotalPages != want.TotalPages {
		t.Errorf("want %d total pages, got %d", want.TotalPages, got.TotalPages)
	}
	if got.PagesLeft != want.PagesLeft {
		t.Errorf("want %d pages left, got %d", want.PagesLeft, got.PagesLeft)
	}
	if got.ItemOffset != want.ItemOffset {
		t.Errorf("want item offset %d, got %d", want.ItemOffset, got.ItemOffset)
	}
	if got.ItemCount != want.ItemCount {
		t.Errorf("want %d total items, got %d", want.ItemCount, got.ItemCount)
	}
}
