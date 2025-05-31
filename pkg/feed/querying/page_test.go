// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import (
	"testing"

	"github.com/virtualtam/sparklemuffin/internal/paginate"
)

func TestNewPage(t *testing.T) {
	cases := []struct {
		tname           string
		number          uint
		totalPages      uint
		totalEntryCount uint
		want            FeedPage
	}{
		{
			tname:      "page 1 of 1",
			number:     1,
			totalPages: 1,
			want: FeedPage{
				Page: paginate.Page{
					PageNumber:         1,
					PreviousPageNumber: 1,
					NextPageNumber:     1,
					TotalPages:         1,
					PagesLeft:          0,
					ItemOffset:         1,
				},
			},
		},
		{
			tname:           "page 1 of 8",
			number:          1,
			totalPages:      8,
			totalEntryCount: 7*entriesPerPage + 10,
			want: FeedPage{
				Page: paginate.Page{
					PageNumber:         1,
					PreviousPageNumber: 1,
					NextPageNumber:     2,
					TotalPages:         8,
					PagesLeft:          7,
					ItemOffset:         1,
					ItemCount:          7*entriesPerPage + 10,
				},
			},
		},
		{
			tname:           "page 7 of 8",
			number:          7,
			totalPages:      8,
			totalEntryCount: 7*entriesPerPage + 10,
			want: FeedPage{
				Page: paginate.Page{
					PageNumber:         7,
					PreviousPageNumber: 6,
					NextPageNumber:     8,
					TotalPages:         8,
					PagesLeft:          1,
					ItemOffset:         6*entriesPerPage + 1,
					ItemCount:          7*entriesPerPage + 10,
				},
			},
		},
		{
			tname:           "page 8 of 8",
			number:          8,
			totalPages:      8,
			totalEntryCount: 7*entriesPerPage + 10,
			want: FeedPage{
				Page: paginate.Page{
					PageNumber:         8,
					PreviousPageNumber: 7,
					NextPageNumber:     8,
					TotalPages:         8,
					PagesLeft:          0,
					ItemOffset:         7*entriesPerPage + 1,
					ItemCount:          7*entriesPerPage + 10,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			got := NewFeedPage(tc.number, tc.totalPages, "", "", []SubscribedFeedsByCategory{}, tc.totalEntryCount, []SubscribedFeedEntry{})
			AssertPageEquals(t, got, tc.want)
		})
	}
}
