// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import (
	"testing"

	"github.com/virtualtam/sparklemuffin/internal/paginate"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
)

func TestNewPage(t *testing.T) {
	cases := []struct {
		tname              string
		number             uint
		totalPages         uint
		totalBookmarkCount uint
		want               BookmarkPage
	}{
		{
			tname:              "page 1 of 1",
			number:             1,
			totalPages:         1,
			totalBookmarkCount: 10,
			want: BookmarkPage{
				Page: paginate.Page{
					PageNumber:         1,
					PreviousPageNumber: 1,
					NextPageNumber:     1,
					TotalPages:         1,
					PagesLeft:          0,
					ItemOffset:         1,
					ItemCount:          10,
				},
			},
		},
		{
			tname:              "page 1 of 8",
			number:             1,
			totalPages:         8,
			totalBookmarkCount: 7*bookmarksPerPage + 10,
			want: BookmarkPage{
				Page: paginate.Page{
					PageNumber:         1,
					PreviousPageNumber: 1,
					NextPageNumber:     2,
					TotalPages:         8,
					PagesLeft:          7,
					ItemOffset:         1,
					ItemCount:          7*bookmarksPerPage + 10,
				},
			},
		},
		{
			tname:              "page 7 of 8",
			number:             7,
			totalPages:         8,
			totalBookmarkCount: 7*bookmarksPerPage + 10,
			want: BookmarkPage{
				Page: paginate.Page{
					PageNumber:         7,
					PreviousPageNumber: 6,
					NextPageNumber:     8,
					TotalPages:         8,
					PagesLeft:          1,
					ItemOffset:         6*bookmarksPerPage + 1,
					ItemCount:          7*bookmarksPerPage + 10,
				},
			},
		},
		{
			tname:              "page 8 of 8",
			number:             8,
			totalPages:         8,
			totalBookmarkCount: 7*bookmarksPerPage + 10,
			want: BookmarkPage{
				Page: paginate.Page{
					PageNumber:         8,
					PreviousPageNumber: 7,
					NextPageNumber:     8,
					TotalPages:         8,
					PagesLeft:          0,
					ItemOffset:         7*bookmarksPerPage + 1,
					ItemCount:          7*bookmarksPerPage + 10,
				},
			},
		},
	}

	owner := Owner{
		UUID:        "13faccd4-8b67-46cf-823a-87e1fa0f7e62",
		NickName:    "test-user",
		DisplayName: "Test User",
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			got := NewBookmarkPage(owner, tc.number, tc.totalPages, tc.totalBookmarkCount, []bookmark.Bookmark{})
			assertBookmarkPageEquals(t, got, tc.want)
		})
	}
}

func assertBookmarkPageEquals(t *testing.T, got, want BookmarkPage) {
	t.Helper()

	paginate.AssertPageEquals(t, got.Page, want.Page)

	if len(got.Bookmarks) != len(want.Bookmarks) {
		t.Fatalf("want %d bookmarks, got %d", len(want.Bookmarks), len(got.Bookmarks))
	}

	for i, wantBookmark := range want.Bookmarks {
		if got.Bookmarks[i].Title != wantBookmark.Title {
			t.Errorf("want bookmark %d title %q, got %q", i, wantBookmark.Title, got.Bookmarks[i].Title)
		}

		if !got.Bookmarks[i].CreatedAt.Equal(wantBookmark.CreatedAt) {
			t.Errorf("want bookmark %d created at %q, got %q", i, wantBookmark.CreatedAt, got.Bookmarks[i].CreatedAt)
		}
	}
}
