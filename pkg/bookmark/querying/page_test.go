package querying

import (
	"testing"

	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
)

func assertPagesEqual(t *testing.T, got, want BookmarkPage) {
	t.Helper()

	if got.SearchTerms != want.SearchTerms {
		t.Errorf("want query %q, got %q", want.SearchTerms, got.SearchTerms)
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
	if got.Offset != want.Offset {
		t.Errorf("want offset %d, got %d", want.Offset, got.Offset)
	}

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

func TestNewPage(t *testing.T) {
	cases := []struct {
		tname      string
		number     uint
		totalPages uint
		want       BookmarkPage
	}{
		{
			tname:      "page 1 of 1",
			number:     1,
			totalPages: 1,
			want: BookmarkPage{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     1,
				TotalPages:         1,
				Offset:             1,
			},
		},
		{
			tname:      "page 1 of 8",
			number:     1,
			totalPages: 8,
			want: BookmarkPage{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     2,
				TotalPages:         8,
				Offset:             1,
			},
		},
		{
			tname:      "page 7 of 8",
			number:     7,
			totalPages: 8,
			want: BookmarkPage{
				PageNumber:         7,
				PreviousPageNumber: 6,
				NextPageNumber:     8,
				TotalPages:         8,
				Offset:             6*bookmarksPerPage + 1,
			},
		},
		{
			tname:      "page 8 of 8",
			number:     8,
			totalPages: 8,
			want: BookmarkPage{
				PageNumber:         8,
				PreviousPageNumber: 7,
				NextPageNumber:     8,
				TotalPages:         8,
				Offset:             7*bookmarksPerPage + 1,
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
			got := NewBookmarkPage(owner, tc.number, tc.totalPages, []bookmark.Bookmark{})
			assertPagesEqual(t, got, tc.want)
		})
	}
}

func TestPageCount(t *testing.T) {
	cases := []struct {
		tname         string
		bookmarkCount uint
		want          uint
	}{
		{
			tname:         "0 bookmarks, 1 page",
			bookmarkCount: 0,
			want:          1,
		},
		{
			tname:         "3 bookmarks, 1 page",
			bookmarkCount: 3,
			want:          1,
		},
		{
			tname:         "bookmarksPerPage bookmarks, 1 page",
			bookmarkCount: bookmarksPerPage,
			want:          1,
		},
		{
			tname:         "bookmarksPerPage+1 bookmarks, 2 pages",
			bookmarkCount: bookmarksPerPage + 1,
			want:          2,
		},
		{
			tname:         "(2*bookmarksPerPage - 1) bookmarks, 2 pages",
			bookmarkCount: 2*bookmarksPerPage - 1,
			want:          2,
		},
		{
			tname:         "(2*bookmarksPerPage) bookmarks, 2 pages",
			bookmarkCount: 2 * bookmarksPerPage,
			want:          2,
		},
		{
			tname:         "(2*bookmarksPerPage + 1) bookmarks, 3 pages",
			bookmarkCount: 2*bookmarksPerPage + 1,
			want:          3,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			got := pageCount(tc.bookmarkCount, bookmarksPerPage)
			if got != tc.want {
				t.Errorf("want %d pages, got %d", tc.want, got)
			}
		})
	}
}
