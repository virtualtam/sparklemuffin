// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package exporting

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/virtualtam/netscape-go/v2"

	"github.com/virtualtam/sparklemuffin/internal/test/assert"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
)

var (
	repositoryBookmarks = []bookmark.Bookmark{
		{
			UserUUID: "5d75c769-059c-4b36-9db6-1c82619e704a",
			Title:    "Bookmark 1",
			URL:      "https://example1.tld",
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
		},
		{
			UserUUID: "5d75c769-059c-4b36-9db6-1c82619e704a",
			Title:    "Bookmark 3 (private)",
			URL:      "https://example3.tld",
			Private:  true,
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
)

func TestServiceExportAsJSONDocument(t *testing.T) {
	now := time.Now().UTC()

	cases := []struct {
		tname string

		userUUID   string
		visibility Visibility

		want    *JsonDocument
		wantErr error
	}{
		// error cases
		{
			tname:   "empty visibility",
			wantErr: ErrVisibilityInvalid,
		},
		{
			tname:      "invalid visibility",
			visibility: "foggy",
			wantErr:    ErrVisibilityInvalid,
		},

		// nominal cases
		{
			tname:      "export all bookmarks, user has none",
			userUUID:   "b9e785dc-3613-4d8d-909b-31a4728b530d",
			visibility: VisibilityAll,
			want: &JsonDocument{
				Title:      "SparkleMuffin export of all bookmarks",
				ExportedAt: now,
			},
		},
		{
			tname:      "export all bookmarks",
			userUUID:   "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility: VisibilityAll,
			want: &JsonDocument{
				Title:      "SparkleMuffin export of all bookmarks",
				ExportedAt: now,
				Bookmarks: []JsonBookmark{
					{
						Title: "Bookmark 1",
						URL:   "https://example1.tld",
					},
					{
						Title:       "Bookmark 2",
						URL:         "https://example2.tld",
						Description: "Second bookmark",
						Tags:        []string{"example", "test"},
					},
					{
						Title:   "Bookmark 3 (private)",
						URL:     "https://example3.tld",
						Private: true,
					},
				},
			},
		},
		{
			tname:      "export private bookmarks",
			userUUID:   "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility: VisibilityPrivate,
			want: &JsonDocument{
				Title:      "SparkleMuffin export of private bookmarks",
				ExportedAt: now,
				Bookmarks: []JsonBookmark{
					{
						Title:   "Bookmark 3 (private)",
						URL:     "https://example3.tld",
						Private: true,
					},
				},
			},
		},
		{
			tname:      "export public bookmarks",
			userUUID:   "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility: VisibilityPublic,
			want: &JsonDocument{
				Title:      "SparkleMuffin export of public bookmarks",
				ExportedAt: now,
				Bookmarks: []JsonBookmark{
					{
						Title: "Bookmark 1",
						URL:   "https://example1.tld",
					},
					{
						Title:       "Bookmark 2",
						URL:         "https://example2.tld",
						Description: "Second bookmark",
						Tags:        []string{"example", "test"},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Bookmarks: repositoryBookmarks,
			}
			s := NewService(r)

			got, err := s.ExportAsJSONDocument(tc.userUUID, tc.visibility)

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

			if got.Title != tc.want.Title {
				t.Errorf("want Title %q, got %q", tc.want.Title, got.Title)
			}

			assert.TimeAlmostEquals(t, "ExportedAt", got.ExportedAt, tc.want.ExportedAt, assert.TimeComparisonDelta)

			if !reflect.DeepEqual(got.Bookmarks, tc.want.Bookmarks) {
				t.Errorf("want exported bookmarks %#v, got %#v", tc.want.Bookmarks, got.Bookmarks)
			}
		})
	}
}

func TestServiceExportAsNetscapeDocument(t *testing.T) {
	cases := []struct {
		tname string

		userUUID   string
		visibility Visibility

		want    *netscape.Document
		wantErr error
	}{
		// error cases
		{
			tname:   "empty visibility",
			wantErr: ErrVisibilityInvalid,
		},
		{
			tname:      "invalid visibility",
			visibility: "foggy",
			wantErr:    ErrVisibilityInvalid,
		},

		// nominal cases
		{
			tname:      "export all bookmarks, user has none",
			userUUID:   "b9e785dc-3613-4d8d-909b-31a4728b530d",
			visibility: VisibilityAll,
			want: &netscape.Document{
				Title: "SparkleMuffin export of all bookmarks",
				Root: netscape.Folder{
					Name: "SparkleMuffin export of all bookmarks",
				},
			},
		},
		{
			tname:      "export all bookmarks",
			userUUID:   "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility: VisibilityAll,
			want: &netscape.Document{
				Title: "SparkleMuffin export of all bookmarks",
				Root: netscape.Folder{
					Name: "SparkleMuffin export of all bookmarks",
					Bookmarks: []netscape.Bookmark{
						{
							Title: "Bookmark 1",
							URL:   "https://example1.tld",
						},
						{
							Title:       "Bookmark 2",
							URL:         "https://example2.tld",
							Description: "Second bookmark",
							Tags:        []string{"example", "test"},
						},
						{
							Title:   "Bookmark 3 (private)",
							URL:     "https://example3.tld",
							Private: true,
						},
					},
				},
			},
		},
		{
			tname:      "export private bookmarks",
			userUUID:   "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility: VisibilityPrivate,
			want: &netscape.Document{
				Title: "SparkleMuffin export of private bookmarks",
				Root: netscape.Folder{
					Name: "SparkleMuffin export of private bookmarks",
					Bookmarks: []netscape.Bookmark{
						{
							Title:   "Bookmark 3 (private)",
							URL:     "https://example3.tld",
							Private: true,
						},
					},
				},
			},
		},
		{
			tname:      "export public bookmarks",
			userUUID:   "5d75c769-059c-4b36-9db6-1c82619e704a",
			visibility: VisibilityPublic,
			want: &netscape.Document{
				Title: "SparkleMuffin export of public bookmarks",
				Root: netscape.Folder{
					Name: "SparkleMuffin export of public bookmarks",
					Bookmarks: []netscape.Bookmark{
						{
							Title: "Bookmark 1",
							URL:   "https://example1.tld",
						},
						{
							Title:       "Bookmark 2",
							URL:         "https://example2.tld",
							Description: "Second bookmark",
							Tags:        []string{"example", "test"},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Bookmarks: repositoryBookmarks,
			}
			s := NewService(r)

			got, err := s.ExportAsNetscapeDocument(tc.userUUID, tc.visibility)

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

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("want exported bookmarks %#v, got %#v", tc.want, got)
			}
		})
	}
}
