package querying

import (
	"errors"
	"testing"
	"time"

	"github.com/virtualtam/yawbe/pkg/bookmark"
)

func TestServiceByPage(t *testing.T) {
	cases := []struct {
		tname               string
		repositoryBookmarks []bookmark.Bookmark
		userUUID            string
		pageNumber          int
		want                Page
		wantErr             error
	}{
		// nominal cases
		{
			tname:      "page 1, 0 bookmarks",
			userUUID:   "b8ed2e7e-a11f-42a7-ae4f-2e80485af823",
			pageNumber: 1,
			want: Page{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     1,
				TotalPages:         1,
				Offset:             1,
			},
		},
		{
			tname: "page 1, 3 bookmarks",
			repositoryBookmarks: []bookmark.Bookmark{
				{
					UserUUID:  "5d75c769-059c-4b36-9db6-1c82619e704a",
					Title:     "Bookmark 1",
					URL:       "https://example1.tld",
					CreatedAt: time.Date(2021, 8, 15, 14, 30, 45, 100, time.Local),
				},
				{
					UserUUID: "218d03f8-976c-4387-9d74-95ed656e3921",
					Title:    "Other user's bookmark 1",
					URL:      "https://test.co.uk",
				},
				{
					UserUUID:    "5d75c769-059c-4b36-9db6-1c82619e704a",
					Title:       "Bookmark 2",
					URL:         "https://example2.tld",
					Description: "Second bookmark",
					Tags:        []string{"example", "test"},
					CreatedAt:   time.Date(2021, 8, 17, 14, 30, 45, 100, time.Local),
				},
				{
					UserUUID:  "5d75c769-059c-4b36-9db6-1c82619e704a",
					Title:     "Bookmark 3 (private)",
					URL:       "https://example3.tld",
					Private:   true,
					CreatedAt: time.Date(2021, 9, 22, 14, 30, 45, 100, time.Local),
				},
				{
					UserUUID: "218d03f8-976c-4387-9d74-95ed656e3921",
					Title:    "Other user's bookmark 2 (private)",
					URL:      "https://test-private.co.uk",
					Private:  true,
				},
				{
					UserUUID: "218d03f8-976c-4387-9d74-95ed656e3921",
					Title:    "Other user's bookmark 3",
					URL:      "https://other.co.uk",
				},
			},
			userUUID:   "5d75c769-059c-4b36-9db6-1c82619e704a",
			pageNumber: 1,
			want: Page{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     1,
				TotalPages:         1,
				Offset:             1,
				Bookmarks: []bookmark.Bookmark{
					{
						UserUUID:  "5d75c769-059c-4b36-9db6-1c82619e704a",
						Title:     "Bookmark 3 (private)",
						URL:       "https://example3.tld",
						Private:   true,
						CreatedAt: time.Date(2021, 9, 22, 14, 30, 45, 100, time.Local),
					},
					{
						UserUUID:    "5d75c769-059c-4b36-9db6-1c82619e704a",
						Title:       "Bookmark 2",
						URL:         "https://example2.tld",
						Description: "Second bookmark",
						Tags:        []string{"example", "test"},
						CreatedAt:   time.Date(2021, 8, 17, 14, 30, 45, 100, time.Local),
					},
					{
						UserUUID:  "5d75c769-059c-4b36-9db6-1c82619e704a",
						Title:     "Bookmark 1",
						URL:       "https://example1.tld",
						CreatedAt: time.Date(2021, 8, 15, 14, 30, 45, 100, time.Local),
					},
				},
			},
		},

		// error cases
		{
			tname:      "negative page number",
			pageNumber: -12,
			wantErr:    ErrPageNumberOutOfBounds,
		},
		{
			tname:      "zeroth page",
			pageNumber: 0,
			wantErr:    ErrPageNumberOutOfBounds,
		},
		{
			tname:      "page number out of bounds",
			pageNumber: 18,
			wantErr:    ErrPageNumberOutOfBounds,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &fakeRepository{
				bookmarks: tc.repositoryBookmarks,
			}

			s := NewService(r)

			got, err := s.ByPage(tc.userUUID, tc.pageNumber)

			if tc.wantErr != nil {
				if errors.Is(err, tc.wantErr) {
					return
				}
				if err == nil {
					t.Fatalf("want error %q, got nil", tc.wantErr)
				}
				t.Fatalf("want error %q, got %q", tc.wantErr, err)
			}

			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			assertPagesEqual(t, got, tc.want)
		})
	}
}
