// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package paginate

import "testing"

func TestPageCount(t *testing.T) {
	itemsPerPage := uint(20)

	cases := []struct {
		tname         string
		bookmarkCount uint
		want          uint
	}{
		{
			tname:         "0 items, 1 page",
			bookmarkCount: 0,
			want:          1,
		},
		{
			tname:         "3 items, 1 page",
			bookmarkCount: 3,
			want:          1,
		},
		{
			tname:         "itemsPerPage items, 1 page",
			bookmarkCount: itemsPerPage,
			want:          1,
		},
		{
			tname:         "itemsPerPage+1 items, 2 pages",
			bookmarkCount: itemsPerPage + 1,
			want:          2,
		},
		{
			tname:         "(2*itemsPerPage - 1) items, 2 pages",
			bookmarkCount: 2*itemsPerPage - 1,
			want:          2,
		},
		{
			tname:         "(2*itemsPerPage) items, 2 pages",
			bookmarkCount: 2 * itemsPerPage,
			want:          2,
		},
		{
			tname:         "(2*itemsPerPage + 1) items, 3 pages",
			bookmarkCount: 2*itemsPerPage + 1,
			want:          3,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			got := PageCount(tc.bookmarkCount, itemsPerPage)
			if got != tc.want {
				t.Errorf("want %d pages, got %d", tc.want, got)
			}
		})
	}
}
