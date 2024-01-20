// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package bookmark

import (
	"errors"
	"testing"
)

func TestServiceAdd(t *testing.T) {
	cases := []struct {
		tname               string
		repositoryBookmarks []Bookmark
		bookmark            Bookmark
		want                Bookmark
		wantErr             error
	}{
		// error cases
		{
			tname:   "empty bookmark",
			wantErr: ErrUserUUIDRequired,
		},
		{
			tname: "missing user UUID",
			bookmark: Bookmark{
				URL:   "https://domain.tld",
				Title: "Example Domain",
			},
			wantErr: ErrUserUUIDRequired,
		},
		{
			tname: "empty URL",
			bookmark: Bookmark{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
			},
			wantErr: ErrURLRequired,
		},
		{
			tname: "empty (whitespace) URL",
			bookmark: Bookmark{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "   ",
			},
			wantErr: ErrURLRequired,
		},
		{
			tname: "unparseable URL",
			bookmark: Bookmark{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      ":/dmn",
			},
			wantErr: ErrURLInvalid,
		},
		{
			tname: "duplicate URL",
			repositoryBookmarks: []Bookmark{
				{
					UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
					URL:      "https://duplicate.domain.tld",
				},
			},
			bookmark: Bookmark{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "https://duplicate.domain.tld",
			},
			wantErr: ErrURLAlreadyRegistered,
		},
		{
			tname: "empty title",
			bookmark: Bookmark{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "https://domain.tld",
			},
			wantErr: ErrTitleRequired,
		},
		{
			tname: "empty (whitespace) title",
			bookmark: Bookmark{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "https://domain.tld",
				Title:    "    ",
			},
			wantErr: ErrTitleRequired,
		},

		// nominal cases
		{
			tname: "add bookmark",
			bookmark: Bookmark{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "https://domain.tld",
				Title:    "Example Domain",
			},
			want: Bookmark{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "https://domain.tld",
				Title:    "Example Domain",
			},
		},
		{
			tname: "add bookmark with description",
			bookmark: Bookmark{
				UserUUID:    "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:         "https://domain.tld",
				Title:       "Example Domain",
				Description: "Hello,\nThis bookmark has a longer description!",
			},
			want: Bookmark{
				UserUUID:    "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:         "https://domain.tld",
				Title:       "Example Domain",
				Description: "Hello,\nThis bookmark has a longer description!",
			},
		},
		{
			tname: "add bookmark with tags",
			bookmark: Bookmark{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "https://domain.tld",
				Title:    "Example Domain",
				Tags: []string{
					"example",
					"test",
				},
			},
			want: Bookmark{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "https://domain.tld",
				Title:    "Example Domain",
				Tags: []string{
					"example",
					"test",
				},
			},
		},
		{
			tname: "add bookmark with unsorted tags",
			bookmark: Bookmark{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "https://domain.tld",
				Title:    "Example Domain",
				Tags: []string{
					"euphonium",
					"xylophone",
					"aulos",
				},
			},
			want: Bookmark{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "https://domain.tld",
				Title:    "Example Domain",
				Tags: []string{
					"aulos",
					"euphonium",
					"xylophone",
				},
			},
		},
		{
			tname: "add bookmark with empty (whitespace) tags",
			bookmark: Bookmark{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "https://domain.tld",
				Title:    "Example Domain",
				Tags: []string{
					"   ",  // spaces
					"	",    // tab
					" 	  ", // spaces, tab, spaces
				},
			},
			want: Bookmark{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "https://domain.tld",
				Title:    "Example Domain",
			},
		},
		{
			tname: "add bookmark with duplicate tags",
			bookmark: Bookmark{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "https://domain.tld",
				Title:    "Example Domain",
				Tags: []string{
					"dupe",
					"dupe",
				},
			},
			want: Bookmark{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "https://domain.tld",
				Title:    "Example Domain",
				Tags: []string{
					"dupe",
				},
			},
		},
		{
			tname: "add bookmark with duplicate tags containing spaces",
			bookmark: Bookmark{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "https://domain.tld",
				Title:    "Example Domain",
				Tags: []string{
					"   dupe",
					"dupe   ",
				},
			},
			want: Bookmark{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "https://domain.tld",
				Title:    "Example Domain",
				Tags: []string{
					"dupe",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Bookmarks: tc.repositoryBookmarks,
			}
			s := NewService(r)

			err := s.Add(tc.bookmark)

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

			got := r.Bookmarks[len(r.Bookmarks)-1]

			assertBookmarksEqual(t, got, tc.want)
		})
	}
}

func TestServiceByUID(t *testing.T) {
	cases := []struct {
		tname               string
		repositoryBookmarks []Bookmark
		userUUID            string
		uid                 string
		want                Bookmark
		wantErr             error
	}{
		{
			tname:   "empty UID",
			wantErr: ErrUIDRequired,
		},
		{
			tname:   "invalid UID",
			uid:     "invalid",
			wantErr: ErrUIDInvalid,
		},
		{
			tname:   "empty user UUID",
			uid:     "27L5erU5VNJzIGY1uPUqzLkc9zV",
			wantErr: ErrUserUUIDRequired,
		},
		{
			tname:    "not found",
			uid:      "27L5pr0PGGF6YTV7ULLu2K1x4xe",
			userUUID: "f0127fa0-722d-458d-9f3c-31823c42e2b7",
			wantErr:  ErrNotFound,
		},
		{
			tname: "get bookmark",
			repositoryBookmarks: []Bookmark{
				{
					UserUUID:    "f0127fa0-722d-458d-9f3c-31823c42e2b7",
					UID:         "27L5pr0PGGF6YTV7ULLu2K1x4xe",
					URL:         "https://domain.tld",
					Title:       "Test Domain",
					Description: "This is useful for tests!",
				},
			},
			uid:      "27L5pr0PGGF6YTV7ULLu2K1x4xe",
			userUUID: "f0127fa0-722d-458d-9f3c-31823c42e2b7",
			want: Bookmark{
				UserUUID:    "f0127fa0-722d-458d-9f3c-31823c42e2b7",
				UID:         "27L5pr0PGGF6YTV7ULLu2K1x4xe",
				URL:         "https://domain.tld",
				Title:       "Test Domain",
				Description: "This is useful for tests!",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Bookmarks: tc.repositoryBookmarks,
			}
			s := NewService(r)

			got, err := s.ByUID(tc.userUUID, tc.uid)

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

			if got.URL != tc.want.URL {
				t.Errorf("want URL %q, got %q", tc.want.URL, got.URL)
			}
			if got.Title != tc.want.Title {
				t.Errorf("want Title %q, got %q", tc.want.Title, got.Title)
			}
			if got.Description != tc.want.Description {
				t.Errorf("want Description %q, got %q", tc.want.Description, got.Description)
			}
		})
	}
}

func TestServiceDelete(t *testing.T) {
	cases := []struct {
		tname               string
		repositoryBookmarks []Bookmark
		userUUID            string
		uid                 string
		wantErr             error
	}{
		{
			tname:   "empty UID",
			wantErr: ErrUIDRequired,
		},
		{
			tname:   "invalid UID",
			uid:     "invalid",
			wantErr: ErrUIDInvalid,
		},
		{
			tname:   "empty user UUID",
			uid:     "27L5erU5VNJzIGY1uPUqzLkc9zV",
			wantErr: ErrUserUUIDRequired,
		},
		{
			tname:    "not found",
			uid:      "27L5pr0PGGF6YTV7ULLu2K1x4xe",
			userUUID: "f0127fa0-722d-458d-9f3c-31823c42e2b7",
			wantErr:  ErrNotFound,
		},
		{
			tname: "delete bookmark",
			repositoryBookmarks: []Bookmark{
				{
					UserUUID: "f0127fa0-722d-458d-9f3c-31823c42e2b7",
					UID:      "27L5pr0PGGF6YTV7ULLu2K1x4xe",
					URL:      "https://domain.tld",
					Title:    "Test Domain",
				},
			},
			uid:      "27L5pr0PGGF6YTV7ULLu2K1x4xe",
			userUUID: "f0127fa0-722d-458d-9f3c-31823c42e2b7",
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Bookmarks: tc.repositoryBookmarks,
			}
			s := NewService(r)

			err := s.Delete(tc.userUUID, tc.uid)

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
		})
	}
}

func TestServiceUpdate(t *testing.T) {
	cases := []struct {
		tname               string
		repositoryBookmarks []Bookmark
		bookmark            Bookmark
		want                Bookmark
		wantErr             error
	}{
		// error cases
		{
			tname:   "empty bookmark",
			wantErr: ErrUIDRequired,
		},
		{
			tname: "invalid UID",
			bookmark: Bookmark{
				UID: "12345",
			},
			wantErr: ErrUIDInvalid,
		},
		{
			tname: "missing user UUID",
			bookmark: Bookmark{
				UID:   "27L4DoEZaRASKhQKygRCrvVAwkr",
				URL:   "https://domain.tld",
				Title: "Example Domain",
			},
			wantErr: ErrUserUUIDRequired,
		},
		{
			tname: "empty URL",
			bookmark: Bookmark{
				UID:      "27L4DoEZaRASKhQKygRCrvVAwkr",
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
			},
			wantErr: ErrURLRequired,
		},
		{
			tname: "empty (whitespace) URL",
			bookmark: Bookmark{
				UID:      "27L4DoEZaRASKhQKygRCrvVAwkr",
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "   ",
			},
			wantErr: ErrURLRequired,
		},
		{
			tname: "unparseable URL",
			bookmark: Bookmark{
				UID:      "27L4DoEZaRASKhQKygRCrvVAwkr",
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      ":/dmn",
			},
			wantErr: ErrURLInvalid,
		},
		{
			tname: "empty title",
			bookmark: Bookmark{
				UID:      "27L4DoEZaRASKhQKygRCrvVAwkr",
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "https://domain.tld",
			},
			wantErr: ErrTitleRequired,
		},
		{
			tname: "empty (whitespace) title",
			bookmark: Bookmark{
				UID:      "27L4DoEZaRASKhQKygRCrvVAwkr",
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "https://domain.tld",
				Title:    "    ",
			},
			wantErr: ErrTitleRequired,
		},
		{
			tname: "not found",
			bookmark: Bookmark{
				UID:      "27L4DoEZaRASKhQKygRCrvVAwkr",
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "https://domain.tld",
				Title:    "Example Domain",
			},
			wantErr: ErrNotFound,
		},

		// nominal cases
		{
			tname: "update bookmark",
			repositoryBookmarks: []Bookmark{
				{
					UID:      "27L4DoEZaRASKhQKygRCrvVAwkr",
					UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
					URL:      "https://domain.tld",
					Title:    "Example Doma",
				},
			},
			bookmark: Bookmark{
				UID:      "27L4DoEZaRASKhQKygRCrvVAwkr",
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "https://domain.tld",
				Title:    "Example Domain",
			},
			want: Bookmark{
				URL:   "https://domain.tld",
				Title: "Example Domain",
			},
		},
		{
			tname: "update bookmark with description",
			repositoryBookmarks: []Bookmark{
				{
					UID:      "27L4DoEZaRASKhQKygRCrvVAwkr",
					UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
					URL:      "https://domain.tld",
					Title:    "Example Doma",
				},
			},
			bookmark: Bookmark{
				UID:         "27L4DoEZaRASKhQKygRCrvVAwkr",
				UserUUID:    "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:         "https://domain.tld",
				Title:       "Example Domain",
				Description: "Hello,\nThis bookmark has a longer description!",
			},
			want: Bookmark{
				URL:         "https://domain.tld",
				Title:       "Example Domain",
				Description: "Hello,\nThis bookmark has a longer description!",
			},
		},
		{
			tname: "update bookmark with tags",
			repositoryBookmarks: []Bookmark{
				{
					UID:      "27L4DoEZaRASKhQKygRCrvVAwkr",
					UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
					URL:      "https://domain.tld",
					Title:    "Example Doma",
				},
			},
			bookmark: Bookmark{
				UID:      "27L4DoEZaRASKhQKygRCrvVAwkr",
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				URL:      "https://domain.tld",
				Title:    "Example Domain",
				Tags: []string{
					"example",
					"  dupe",
					"  ", // spaces
					"	 ", // tab and spaces
					"test",
					"dupe",
					"dupe  ",
				},
			},
			want: Bookmark{
				URL:   "https://domain.tld",
				Title: "Example Domain",
				Tags: []string{
					"dupe",
					"example",
					"test",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Bookmarks: tc.repositoryBookmarks,
			}
			s := NewService(r)

			err := s.Update(tc.bookmark)

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

			got, err := r.BookmarkGetByUID(tc.bookmark.UserUUID, tc.bookmark.UID)
			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			assertBookmarksEqual(t, got, tc.want)
		})
	}
}

func TestServiceDeleteTag(t *testing.T) {
	cases := []struct {
		tname                   string
		repositoryBookmarks     []Bookmark
		tagDeleteQuery          TagDeleteQuery
		want                    int64
		wantErr                 error
		wantRepositoryBookmarks []Bookmark
	}{
		// error cases
		{
			tname: "tag is empty",
			tagDeleteQuery: TagDeleteQuery{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
			},
			wantErr: ErrTagNameRequired,
		},
		{
			tname: "tag is empty (whitespace)",
			tagDeleteQuery: TagDeleteQuery{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				Name:     "     ",
			},
			wantErr: ErrTagNameRequired,
		},
		{
			tname: "tag contains whitespace (multiple values)",
			tagDeleteQuery: TagDeleteQuery{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				Name:     "tag1   tag2",
			},
			wantErr: ErrTagNameContainsWhitespace,
		},

		// nominal cases
		{
			tname: "no bookmark with this tag",
			tagDeleteQuery: TagDeleteQuery{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				Name:     "tag1",
			},
		},
		{
			tname: "update bookmark with tags",
			repositoryBookmarks: []Bookmark{
				{
					UID:      "27L4DoEZaRASKhQKygRCrvVAwkr",
					UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
					URL:      "https://domain.tld",
					Tags:     []string{"a", "c", "delete-me", "z"},
					Title:    "Example Domain",
				},
			},
			tagDeleteQuery: TagDeleteQuery{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				Name:     "delete-me",
			},
			want: 1,
			wantRepositoryBookmarks: []Bookmark{
				{
					UID:      "27L4DoEZaRASKhQKygRCrvVAwkr",
					UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
					URL:      "https://domain.tld",
					Tags:     []string{"a", "c", "z"},
					Title:    "Example Domain",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Bookmarks: tc.repositoryBookmarks,
			}
			s := NewService(r)

			got, err := s.DeleteTag(tc.tagDeleteQuery)

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

			if got != tc.want {
				t.Errorf("want %d updated bookmarks, got %d", tc.want, got)
			}

			for index, bookmark := range r.Bookmarks {
				assertBookmarksEqual(t, bookmark, tc.wantRepositoryBookmarks[index])
			}
		})
	}
}

func TestServiceUpdateTag(t *testing.T) {
	cases := []struct {
		tname                   string
		repositoryBookmarks     []Bookmark
		tagNameUpdate           TagUpdateQuery
		want                    int64
		wantErr                 error
		wantRepositoryBookmarks []Bookmark
	}{
		// error cases
		{
			tname: "current tag is empty",
			tagNameUpdate: TagUpdateQuery{
				UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
			},
			wantErr: ErrTagNameRequired,
		},
		{
			tname: "current tag is empty (whitespace)",
			tagNameUpdate: TagUpdateQuery{
				UserUUID:    "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				CurrentName: "     ",
			},
			wantErr: ErrTagNameRequired,
		},
		{
			tname: "current tag contains whitespace (multiple values)",
			tagNameUpdate: TagUpdateQuery{
				UserUUID:    "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				CurrentName: "tag1   tag2",
			},
			wantErr: ErrTagNameContainsWhitespace,
		},
		{
			tname: "new tag is empty",
			tagNameUpdate: TagUpdateQuery{
				UserUUID:    "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				CurrentName: "test",
			},
			wantErr: ErrTagNameRequired,
		},
		{
			tname: "new tag is empty (whitespace)",
			tagNameUpdate: TagUpdateQuery{
				UserUUID:    "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				CurrentName: "test",
				NewName:     "     ",
			},
			wantErr: ErrTagNameRequired,
		},
		{
			tname: "new tag contains whitespace (multiple values)",
			tagNameUpdate: TagUpdateQuery{
				UserUUID:    "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				CurrentName: "tag1",
				NewName:     "tag2 tag3   tag4",
			},
			wantErr: ErrTagNameContainsWhitespace,
		},
		{
			tname: "new tag equals current tag",
			tagNameUpdate: TagUpdateQuery{
				UserUUID:    "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				CurrentName: "tag1",
				NewName:     "tag1",
			},
			wantErr: ErrTagNewNameEqualsCurrentName,
		},

		// nominal cases
		{
			tname: "no bookmark with this tag",
			tagNameUpdate: TagUpdateQuery{
				UserUUID:    "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				CurrentName: "tag1",
				NewName:     "tag2",
			},
		},
		{
			tname: "update bookmark with tags",
			repositoryBookmarks: []Bookmark{
				{
					UID:      "27L4DoEZaRASKhQKygRCrvVAwkr",
					UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
					URL:      "https://domain.tld",
					Tags:     []string{"a", "c", "replace-me", "z"},
					Title:    "Example Domain",
				},
			},
			tagNameUpdate: TagUpdateQuery{
				UserUUID:    "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
				CurrentName: "replace-me",
				NewName:     "b",
			},
			want: 1,
			wantRepositoryBookmarks: []Bookmark{
				{
					UID:      "27L4DoEZaRASKhQKygRCrvVAwkr",
					UserUUID: "6fe6a0c6-62da-4d05-b0c5-dc9d6ef58096",
					URL:      "https://domain.tld",
					Tags:     []string{"a", "b", "c", "z"},
					Title:    "Example Domain",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Bookmarks: tc.repositoryBookmarks,
			}
			s := NewService(r)

			got, err := s.UpdateTag(tc.tagNameUpdate)

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

			if got != tc.want {
				t.Errorf("want %d updated bookmarks, got %d", tc.want, got)
			}

			for index, bookmark := range r.Bookmarks {
				assertBookmarksEqual(t, bookmark, tc.wantRepositoryBookmarks[index])
			}
		})
	}
}

func assertBookmarksEqual(t *testing.T, got, want Bookmark) {
	t.Helper()

	if got.URL != want.URL {
		t.Errorf("want URL %q, got %q", want.URL, got.URL)
	}

	if got.Title != want.Title {
		t.Errorf("want Title %q, got %q", want.Title, got.Title)
	}

	if got.Description != want.Description {
		t.Errorf("want Description %q, got %q", want.Description, got.Description)
	}

	if got.Private != want.Private {
		t.Errorf("want Private %t, got %t", want.Private, got.Private)
	}

	if len(got.Tags) != len(want.Tags) {
		t.Fatalf("want %d tags, got %d", len(want.Tags), len(got.Tags))
	}

	for i, wantTag := range want.Tags {
		if got.Tags[i] != wantTag {
			t.Errorf("want tag %d Name %q, got %q", i, wantTag, got.Tags[i])
		}
	}
}
