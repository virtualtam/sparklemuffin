// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package paginate_test

import (
	"testing"

	"github.com/virtualtam/sparklemuffin/internal/paginate"
)

func TestNewPage(t *testing.T) {
	itemsPerPage := uint(20)

	cases := []struct {
		tname              string
		number             uint
		totalPages         uint
		totalBookmarkCount uint
		want               paginate.Page
	}{
		{
			tname:              "page 1 of 1",
			number:             1,
			totalPages:         1,
			totalBookmarkCount: 10,
			want: paginate.Page{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     1,
				TotalPages:         1,
				PagesLeft:          0,
				ItemOffset:         1,
				ItemCount:          10,
			},
		},
		{
			tname:              "page 1 of 8",
			number:             1,
			totalPages:         8,
			totalBookmarkCount: 7*itemsPerPage + 10,
			want: paginate.Page{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     2,
				TotalPages:         8,
				PagesLeft:          7,
				ItemOffset:         1,
				ItemCount:          7*itemsPerPage + 10,
			},
		},
		{
			tname:              "page 7 of 8",
			number:             7,
			totalPages:         8,
			totalBookmarkCount: 7*itemsPerPage + 10,
			want: paginate.Page{
				PageNumber:         7,
				PreviousPageNumber: 6,
				NextPageNumber:     8,
				TotalPages:         8,
				PagesLeft:          1,
				ItemOffset:         6*itemsPerPage + 1,
				ItemCount:          7*itemsPerPage + 10,
			},
		},
		{
			tname:              "page 8 of 8",
			number:             8,
			totalPages:         8,
			totalBookmarkCount: 7*itemsPerPage + 10,
			want: paginate.Page{
				PageNumber:         8,
				PreviousPageNumber: 7,
				NextPageNumber:     8,
				TotalPages:         8,
				PagesLeft:          0,
				ItemOffset:         7*itemsPerPage + 1,
				ItemCount:          7*itemsPerPage + 10,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			got := paginate.NewPage(tc.number, tc.totalPages, itemsPerPage, tc.totalBookmarkCount)
			paginate.AssertPageEquals(t, got, tc.want)
		})
	}
}
