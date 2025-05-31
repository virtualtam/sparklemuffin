// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import (
	"errors"
	"testing"
	"time"

	"github.com/virtualtam/sparklemuffin/internal/paginate"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

var testRepositoryUsers = []user.User{
	{
		UUID:        "b8ed2e7e-a11f-42a7-ae4f-2e80485af823",
		NickName:    "test-user-1",
		DisplayName: "Test User 1",
	},
	{
		UUID:        "5d75c769-059c-4b36-9db6-1c82619e704a",
		NickName:    "test-user-1",
		DisplayName: "Test User 1",
	},
}

var testRepositoryBookmarks = []bookmark.Bookmark{
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
}

func TestServiceByPage(t *testing.T) {
	cases := []struct {
		tname               string
		repositoryBookmarks []bookmark.Bookmark
		ownerUUID           string
		visibility          Visibility
		pageNumber          uint
		want                BookmarkPage
		wantErr             error
	}{
		// nominal cases
		{
			tname:      "page 1, 0 bookmarks",
			ownerUUID:  "b8ed2e7e-a11f-42a7-ae4f-2e80485af823",
			visibility: VisibilityAll,
			pageNumber: 1,
			want: BookmarkPage{
				Page: paginate.Page{
					PageNumber:         1,
					PreviousPageNumber: 1,
					NextPageNumber:     1,
					TotalPages:         1,
					ItemOffset:         1,
				},
			},
		},
		{
			tname:               "page 1, 3 bookmarks (2 public, 1 private)",
			repositoryBookmarks: testRepositoryBookmarks,
			ownerUUID:           "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility:          VisibilityAll,
			pageNumber:          1,
			want: BookmarkPage{
				Page: paginate.Page{
					PageNumber:         1,
					PreviousPageNumber: 1,
					NextPageNumber:     1,
					TotalPages:         1,
					ItemOffset:         1,
					ItemCount:          3,
				},
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
		{
			tname:               "page 1, 1 private bookmark",
			repositoryBookmarks: testRepositoryBookmarks,
			ownerUUID:           "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility:          VisibilityPrivate,
			pageNumber:          1,
			want: BookmarkPage{
				Page: paginate.Page{
					PageNumber:         1,
					PreviousPageNumber: 1,
					NextPageNumber:     1,
					TotalPages:         1,
					ItemOffset:         1,
					ItemCount:          1,
				},
				Bookmarks: []bookmark.Bookmark{
					{
						UserUUID:  "5d75c769-059c-4b36-9db6-1c82619e704a",
						Title:     "Bookmark 3 (private)",
						URL:       "https://example3.tld",
						Private:   true,
						CreatedAt: time.Date(2021, 9, 22, 14, 30, 45, 100, time.Local),
					},
				},
			},
		},
		{
			tname:               "page 1, 2 public bookmarks",
			repositoryBookmarks: testRepositoryBookmarks,
			ownerUUID:           "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility:          VisibilityPublic,
			pageNumber:          1,
			want: BookmarkPage{
				Page: paginate.Page{
					PageNumber:         1,
					PreviousPageNumber: 1,
					NextPageNumber:     1,
					TotalPages:         1,
					ItemOffset:         1,
					ItemCount:          2,
				},
				Bookmarks: []bookmark.Bookmark{
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
			tname:      "owner not found",
			pageNumber: 10,
			ownerUUID:  "9681e525-f205-489d-b53e-1a858b4ca561",
			wantErr:    ErrOwnerNotFound,
		},
		{
			tname:      "zeroth page",
			pageNumber: 0,
			ownerUUID:  "5d75c769-059c-4b36-9db6-1c82619e704a",
			wantErr:    ErrPageNumberOutOfBounds,
		},
		{
			tname:      "page number out of bounds",
			pageNumber: 18,
			ownerUUID:  "5d75c769-059c-4b36-9db6-1c82619e704a",
			wantErr:    ErrPageNumberOutOfBounds,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &fakeRepository{
				bookmarks: tc.repositoryBookmarks,
				users:     testRepositoryUsers,
			}

			s := NewService(r)

			got, err := s.BookmarksByPage(tc.ownerUUID, tc.visibility, tc.pageNumber)

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

			assertBookmarkPageEquals(t, got, tc.want)
		})
	}
}
