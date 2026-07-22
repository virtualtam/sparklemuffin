// Copyright VirtualTam 2022, 2026
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
		UID:       "test-uid-bookmark-1",
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
		UID:         "test-uid-bookmark-2",
		UserUUID:    "5d75c769-059c-4b36-9db6-1c82619e704a",
		Title:       "Bookmark 2",
		URL:         "https://example2.tld",
		Description: "Second bookmark",
		Tags:        []string{"example", "test"},
		CreatedAt:   time.Date(2021, 8, 17, 14, 30, 45, 100, time.Local),
	},
	{
		UID:       "test-uid-bookmark-3",
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

// testRepositoryBookmarksWithTags provides bookmarks with varied tags across
// visibility levels to exercise tag-related service methods.
var testRepositoryBookmarksWithTags = []bookmark.Bookmark{
	{
		UserUUID: "5d75c769-059c-4b36-9db6-1c82619e704a",
		Title:    "Bookmark A",
		URL:      "https://a.tld",
		Tags:     []string{"go", "programming"},
	},
	{
		UserUUID: "5d75c769-059c-4b36-9db6-1c82619e704a",
		Title:    "Bookmark B (private)",
		URL:      "https://b.tld",
		Tags:     []string{"go", "testing"},
		Private:  true,
	},
	{
		UserUUID: "5d75c769-059c-4b36-9db6-1c82619e704a",
		Title:    "Bookmark C",
		URL:      "https://c.tld",
		Tags:     []string{"go", "programming"},
	},
	{
		UserUUID: "218d03f8-976c-4387-9d74-95ed656e3921",
		Title:    "Other user bookmark",
		URL:      "https://other.tld",
		Tags:     []string{"go"},
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
			wantErr:    paginate.ErrPageNumberOutOfBounds,
		},
		{
			tname:      "page number out of bounds",
			pageNumber: 18,
			ownerUUID:  "5d75c769-059c-4b36-9db6-1c82619e704a",
			wantErr:    paginate.ErrPageNumberOutOfBounds,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Bookmarks: tc.repositoryBookmarks,
				Users:     testRepositoryUsers,
			}

			s := NewService(r)

			got, err := s.BookmarksByPage(t.Context(), tc.ownerUUID, tc.visibility, tc.pageNumber)

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

func TestServicePublicBookmarkByUID(t *testing.T) {
	cases := []struct {
		tname               string
		repositoryBookmarks []bookmark.Bookmark
		ownerUUID           string
		uid                 string
		want                BookmarkPage
		wantErr             error
	}{
		// nominal cases
		{
			tname:               "bookmark found",
			repositoryBookmarks: testRepositoryBookmarks,
			ownerUUID:           "5d75c769-059c-4b36-9db6-1c82619e704a",
			uid:                 "test-uid-bookmark-1",
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
						Title:     "Bookmark 1",
						CreatedAt: time.Date(2021, 8, 15, 14, 30, 45, 100, time.Local),
					},
				},
			},
		},
		{
			tname:               "bookmark not found (private)",
			repositoryBookmarks: testRepositoryBookmarks,
			ownerUUID:           "5d75c769-059c-4b36-9db6-1c82619e704a",
			uid:                 "test-uid-bookmark-3",
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
			tname:               "bookmark not found (unknown uid)",
			repositoryBookmarks: testRepositoryBookmarks,
			ownerUUID:           "5d75c769-059c-4b36-9db6-1c82619e704a",
			uid:                 "does-not-exist",
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

		// error cases
		{
			tname:     "owner not found",
			ownerUUID: "9681e525-f205-489d-b53e-1a858b4ca561",
			uid:       "test-uid-bookmark-1",
			wantErr:   ErrOwnerNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Bookmarks: tc.repositoryBookmarks,
				Users:     testRepositoryUsers,
			}

			s := NewService(r)

			got, err := s.PublicBookmarkByUID(t.Context(), tc.ownerUUID, tc.uid)

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

func TestServiceBookmarksBySearchQueryAndPage(t *testing.T) {
	cases := []struct {
		tname               string
		repositoryBookmarks []bookmark.Bookmark
		ownerUUID           string
		visibility          Visibility
		searchTerms         string
		pageNumber          uint
		want                BookmarkPage
		wantErr             error
	}{
		// nominal cases
		{
			tname:       "0 results",
			ownerUUID:   "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility:  VisibilityAll,
			searchTerms: "notfound",
			pageNumber:  1,
			want: BookmarkPage{
				Page: paginate.Page{
					PageNumber:         1,
					PreviousPageNumber: 1,
					NextPageNumber:     1,
					TotalPages:         1,
					ItemOffset:         1,
					SearchTerms:        "notfound",
				},
			},
		},
		{
			tname:               "1 result",
			repositoryBookmarks: testRepositoryBookmarks,
			ownerUUID:           "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility:          VisibilityAll,
			searchTerms:         "example1",
			pageNumber:          1,
			want: BookmarkPage{
				Page: paginate.Page{
					PageNumber:         1,
					PreviousPageNumber: 1,
					NextPageNumber:     1,
					TotalPages:         1,
					ItemOffset:         1,
					ItemCount:          1,
					SearchTerms:        "example1",
				},
				Bookmarks: []bookmark.Bookmark{
					{
						Title:     "Bookmark 1",
						CreatedAt: time.Date(2021, 8, 15, 14, 30, 45, 100, time.Local),
					},
				},
			},
		},
		{
			tname:               "multiple results",
			repositoryBookmarks: testRepositoryBookmarks,
			ownerUUID:           "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility:          VisibilityAll,
			searchTerms:         "example",
			pageNumber:          1,
			want: BookmarkPage{
				Page: paginate.Page{
					PageNumber:         1,
					PreviousPageNumber: 1,
					NextPageNumber:     1,
					TotalPages:         1,
					ItemOffset:         1,
					ItemCount:          3,
					SearchTerms:        "example",
				},
				Bookmarks: []bookmark.Bookmark{
					{
						Title:     "Bookmark 3 (private)",
						CreatedAt: time.Date(2021, 9, 22, 14, 30, 45, 100, time.Local),
					},
					{
						Title:     "Bookmark 2",
						CreatedAt: time.Date(2021, 8, 17, 14, 30, 45, 100, time.Local),
					},
					{
						Title:     "Bookmark 1",
						CreatedAt: time.Date(2021, 8, 15, 14, 30, 45, 100, time.Local),
					},
				},
			},
		},
		{
			tname:               "public only, multiple results",
			repositoryBookmarks: testRepositoryBookmarks,
			ownerUUID:           "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility:          VisibilityPublic,
			searchTerms:         "example",
			pageNumber:          1,
			want: BookmarkPage{
				Page: paginate.Page{
					PageNumber:         1,
					PreviousPageNumber: 1,
					NextPageNumber:     1,
					TotalPages:         1,
					ItemOffset:         1,
					ItemCount:          2,
					SearchTerms:        "example",
				},
				Bookmarks: []bookmark.Bookmark{
					{
						Title:     "Bookmark 2",
						CreatedAt: time.Date(2021, 8, 17, 14, 30, 45, 100, time.Local),
					},
					{
						Title:     "Bookmark 1",
						CreatedAt: time.Date(2021, 8, 15, 14, 30, 45, 100, time.Local),
					},
				},
			},
		},

		// error cases
		{
			tname:       "owner not found",
			ownerUUID:   "9681e525-f205-489d-b53e-1a858b4ca561",
			searchTerms: "example",
			pageNumber:  1,
			wantErr:     ErrOwnerNotFound,
		},
		{
			tname:       "zeroth page",
			ownerUUID:   "5d75c769-059c-4b36-9db6-1c82619e704a",
			searchTerms: "example",
			pageNumber:  0,
			wantErr:     paginate.ErrPageNumberOutOfBounds,
		},
		{
			tname:               "page number out of bounds",
			repositoryBookmarks: testRepositoryBookmarks,
			ownerUUID:           "5d75c769-059c-4b36-9db6-1c82619e704a",
			searchTerms:         "example",
			pageNumber:          18,
			wantErr:             paginate.ErrPageNumberOutOfBounds,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Bookmarks: tc.repositoryBookmarks,
				Users:     testRepositoryUsers,
			}

			s := NewService(r)

			got, err := s.BookmarksBySearchQueryAndPage(t.Context(), tc.ownerUUID, tc.visibility, tc.searchTerms, tc.pageNumber)

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

func TestServicePublicBookmarksBySearchQueryAndPage(t *testing.T) {
	cases := []struct {
		tname               string
		repositoryBookmarks []bookmark.Bookmark
		ownerUUID           string
		searchTerms         string
		pageNumber          uint
		want                BookmarkPage
		wantErr             error
	}{
		// nominal cases
		{
			tname:       "0 results",
			ownerUUID:   "5d75c769-059c-4b36-9db6-1c82619e704a",
			searchTerms: "notfound",
			pageNumber:  1,
			want: BookmarkPage{
				Page: paginate.Page{
					PageNumber:         1,
					PreviousPageNumber: 1,
					NextPageNumber:     1,
					TotalPages:         1,
					ItemOffset:         1,
					SearchTerms:        "notfound",
				},
			},
		},
		{
			tname:               "public bookmarks only",
			repositoryBookmarks: testRepositoryBookmarks,
			ownerUUID:           "5d75c769-059c-4b36-9db6-1c82619e704a",
			searchTerms:         "example",
			pageNumber:          1,
			want: BookmarkPage{
				Page: paginate.Page{
					PageNumber:         1,
					PreviousPageNumber: 1,
					NextPageNumber:     1,
					TotalPages:         1,
					ItemOffset:         1,
					ItemCount:          2,
					SearchTerms:        "example",
				},
				Bookmarks: []bookmark.Bookmark{
					{
						Title:     "Bookmark 2",
						CreatedAt: time.Date(2021, 8, 17, 14, 30, 45, 100, time.Local),
					},
					{
						Title:     "Bookmark 1",
						CreatedAt: time.Date(2021, 8, 15, 14, 30, 45, 100, time.Local),
					},
				},
			},
		},

		// error cases
		{
			tname:       "owner not found",
			ownerUUID:   "9681e525-f205-489d-b53e-1a858b4ca561",
			searchTerms: "example",
			pageNumber:  1,
			wantErr:     ErrOwnerNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Bookmarks: tc.repositoryBookmarks,
				Users:     testRepositoryUsers,
			}

			s := NewService(r)

			got, err := s.PublicBookmarksBySearchQueryAndPage(t.Context(), tc.ownerUUID, tc.searchTerms, tc.pageNumber)

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

func TestServicePublicBookmarksByPage(t *testing.T) {
	cases := []struct {
		tname               string
		repositoryBookmarks []bookmark.Bookmark
		ownerUUID           string
		pageNumber          uint
		want                BookmarkPage
		wantErr             error
	}{
		// nominal cases
		{
			tname:      "page 1, 0 bookmarks",
			ownerUUID:  "b8ed2e7e-a11f-42a7-ae4f-2e80485af823",
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
			tname:               "page 1, 2 public bookmarks",
			repositoryBookmarks: testRepositoryBookmarks,
			ownerUUID:           "5d75c769-059c-4b36-9db6-1c82619e704a",
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
						Title:     "Bookmark 2",
						CreatedAt: time.Date(2021, 8, 17, 14, 30, 45, 100, time.Local),
					},
					{
						Title:     "Bookmark 1",
						CreatedAt: time.Date(2021, 8, 15, 14, 30, 45, 100, time.Local),
					},
				},
			},
		},

		// error cases
		{
			tname:      "owner not found",
			ownerUUID:  "9681e525-f205-489d-b53e-1a858b4ca561",
			pageNumber: 1,
			wantErr:    ErrOwnerNotFound,
		},
		{
			tname:      "page number out of bounds",
			ownerUUID:  "5d75c769-059c-4b36-9db6-1c82619e704a",
			pageNumber: 18,
			wantErr:    paginate.ErrPageNumberOutOfBounds,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Bookmarks: tc.repositoryBookmarks,
				Users:     testRepositoryUsers,
			}

			s := NewService(r)

			got, err := s.PublicBookmarksByPage(t.Context(), tc.ownerUUID, tc.pageNumber)

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

func TestServiceTags(t *testing.T) {
	cases := []struct {
		tname               string
		repositoryBookmarks []bookmark.Bookmark
		ownerUUID           string
		visibility          Visibility
		want                []Tag
		wantErr             error
	}{
		{
			tname:      "0 tags",
			ownerUUID:  "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility: VisibilityAll,
			want:       []Tag{},
		},
		{
			tname:               "all visibility",
			repositoryBookmarks: testRepositoryBookmarksWithTags,
			ownerUUID:           "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility:          VisibilityAll,
			want: []Tag{
				{Name: "go", Count: 3},
				{Name: "programming", Count: 2},
				{Name: "testing", Count: 1},
			},
		},
		{
			tname:               "public visibility",
			repositoryBookmarks: testRepositoryBookmarksWithTags,
			ownerUUID:           "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility:          VisibilityPublic,
			want: []Tag{
				{Name: "go", Count: 2},
				{Name: "programming", Count: 2},
			},
		},
		{
			tname:               "private visibility",
			repositoryBookmarks: testRepositoryBookmarksWithTags,
			ownerUUID:           "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility:          VisibilityPrivate,
			want: []Tag{
				{Name: "go", Count: 1},
				{Name: "testing", Count: 1},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Bookmarks: tc.repositoryBookmarks,
				Users:     testRepositoryUsers,
			}

			s := NewService(r)

			got, err := s.Tags(t.Context(), tc.ownerUUID, tc.visibility)

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

			if len(got) != len(tc.want) {
				t.Fatalf("want %d tags, got %d", len(tc.want), len(got))
			}

			for i, wantTag := range tc.want {
				if got[i].Name != wantTag.Name {
					t.Errorf("want tag %d name %q, got %q", i, wantTag.Name, got[i].Name)
				}
				if got[i].Count != wantTag.Count {
					t.Errorf("want tag %d count %d, got %d", i, wantTag.Count, got[i].Count)
				}
			}
		})
	}
}

func TestServiceTagNamesByCount(t *testing.T) {
	cases := []struct {
		tname               string
		repositoryBookmarks []bookmark.Bookmark
		ownerUUID           string
		visibility          Visibility
		want                []string
		wantErr             error
	}{
		{
			tname:      "0 tags",
			ownerUUID:  "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility: VisibilityAll,
			want:       []string{},
		},
		{
			tname:               "all visibility",
			repositoryBookmarks: testRepositoryBookmarksWithTags,
			ownerUUID:           "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility:          VisibilityAll,
			want:                []string{"go", "programming", "testing"},
		},
		{
			tname:               "public visibility",
			repositoryBookmarks: testRepositoryBookmarksWithTags,
			ownerUUID:           "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility:          VisibilityPublic,
			want:                []string{"go", "programming"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Bookmarks: tc.repositoryBookmarks,
				Users:     testRepositoryUsers,
			}

			s := NewService(r)

			got, err := s.TagNamesByCount(t.Context(), tc.ownerUUID, tc.visibility)

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

			if len(got) != len(tc.want) {
				t.Fatalf("want %d tag names, got %d", len(tc.want), len(got))
			}

			for i, wantName := range tc.want {
				if got[i] != wantName {
					t.Errorf("want tag name %d %q, got %q", i, wantName, got[i])
				}
			}
		})
	}
}

func TestServiceTagsByPage(t *testing.T) {
	cases := []struct {
		tname               string
		repositoryBookmarks []bookmark.Bookmark
		ownerUUID           string
		visibility          Visibility
		pageNumber          uint
		want                TagPage
		wantErr             error
	}{
		// nominal cases
		{
			tname:      "0 tags",
			ownerUUID:  "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility: VisibilityAll,
			pageNumber: 1,
			want: TagPage{
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
			tname:               "page 1, all visibility",
			repositoryBookmarks: testRepositoryBookmarksWithTags,
			ownerUUID:           "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility:          VisibilityAll,
			pageNumber:          1,
			want: TagPage{
				Page: paginate.Page{
					PageNumber:         1,
					PreviousPageNumber: 1,
					NextPageNumber:     1,
					TotalPages:         1,
					ItemOffset:         1,
					ItemCount:          3,
				},
				Tags: []Tag{
					{Name: "go", Count: 3},
					{Name: "programming", Count: 2},
					{Name: "testing", Count: 1},
				},
			},
		},
		{
			tname:               "page 1, public visibility",
			repositoryBookmarks: testRepositoryBookmarksWithTags,
			ownerUUID:           "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility:          VisibilityPublic,
			pageNumber:          1,
			want: TagPage{
				Page: paginate.Page{
					PageNumber:         1,
					PreviousPageNumber: 1,
					NextPageNumber:     1,
					TotalPages:         1,
					ItemOffset:         1,
					ItemCount:          2,
				},
				Tags: []Tag{
					{Name: "go", Count: 2},
					{Name: "programming", Count: 2},
				},
			},
		},

		// error cases
		{
			tname:      "zeroth page",
			ownerUUID:  "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility: VisibilityAll,
			pageNumber: 0,
			wantErr:    paginate.ErrPageNumberOutOfBounds,
		},
		{
			tname:               "page number out of bounds",
			repositoryBookmarks: testRepositoryBookmarksWithTags,
			ownerUUID:           "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility:          VisibilityAll,
			pageNumber:          18,
			wantErr:             paginate.ErrPageNumberOutOfBounds,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Bookmarks: tc.repositoryBookmarks,
				Users:     testRepositoryUsers,
			}

			s := NewService(r)

			got, err := s.TagsByPage(t.Context(), tc.ownerUUID, tc.visibility, tc.pageNumber)

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

			assertTagPageEquals(t, got, tc.want)
		})
	}
}

func TestServiceTagsBySearchQueryAndPage(t *testing.T) {
	cases := []struct {
		tname               string
		repositoryBookmarks []bookmark.Bookmark
		ownerUUID           string
		visibility          Visibility
		searchTerms         string
		pageNumber          uint
		want                TagPage
		wantErr             error
	}{
		// nominal cases
		{
			tname:       "0 matches",
			ownerUUID:   "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility:  VisibilityAll,
			searchTerms: "notfound",
			pageNumber:  1,
			want: TagPage{
				Page: paginate.Page{
					PageNumber:         1,
					PreviousPageNumber: 1,
					NextPageNumber:     1,
					TotalPages:         1,
					ItemOffset:         1,
					SearchTerms:        "notfound",
				},
			},
		},
		{
			tname:               "1 match",
			repositoryBookmarks: testRepositoryBookmarksWithTags,
			ownerUUID:           "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility:          VisibilityAll,
			searchTerms:         "test",
			pageNumber:          1,
			want: TagPage{
				Page: paginate.Page{
					PageNumber:         1,
					PreviousPageNumber: 1,
					NextPageNumber:     1,
					TotalPages:         1,
					ItemOffset:         1,
					ItemCount:          1,
					SearchTerms:        "test",
				},
				Tags: []Tag{
					{Name: "testing", Count: 1},
				},
			},
		},
		{
			tname:               "multiple matches",
			repositoryBookmarks: testRepositoryBookmarksWithTags,
			ownerUUID:           "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility:          VisibilityAll,
			searchTerms:         "o",
			pageNumber:          1,
			want: TagPage{
				Page: paginate.Page{
					PageNumber:         1,
					PreviousPageNumber: 1,
					NextPageNumber:     1,
					TotalPages:         1,
					ItemOffset:         1,
					ItemCount:          2,
					SearchTerms:        "o",
				},
				Tags: []Tag{
					{Name: "go", Count: 3},
					{Name: "programming", Count: 2},
				},
			},
		},

		// error cases
		{
			tname:       "zeroth page",
			ownerUUID:   "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility:  VisibilityAll,
			searchTerms: "go",
			pageNumber:  0,
			wantErr:     paginate.ErrPageNumberOutOfBounds,
		},
		{
			tname:               "page number out of bounds",
			repositoryBookmarks: testRepositoryBookmarksWithTags,
			ownerUUID:           "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility:          VisibilityAll,
			searchTerms:         "go",
			pageNumber:          18,
			wantErr:             paginate.ErrPageNumberOutOfBounds,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Bookmarks: tc.repositoryBookmarks,
				Users:     testRepositoryUsers,
			}

			s := NewService(r)

			got, err := s.TagsBySearchQueryAndPage(t.Context(), tc.ownerUUID, tc.visibility, tc.searchTerms, tc.pageNumber)

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

			assertTagPageEquals(t, got, tc.want)
		})
	}
}
