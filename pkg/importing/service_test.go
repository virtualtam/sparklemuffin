package importing

import (
	"errors"
	"testing"

	"github.com/virtualtam/netscape-go/v2"
	"github.com/virtualtam/yawbe/pkg/bookmark"
)

func TestServiceImportFromNetscapeDocument(t *testing.T) {
	cases := []struct {
		tname               string
		repositoryBookmarks []bookmark.Bookmark

		userUUID           string
		document           netscape.Document
		onConflictStrategy OnConflictStrategy
		visibility         Visibility

		want       []bookmark.Bookmark
		wantErr    error
		wantStatus Status
	}{
		// nominal cases
		{
			tname:              "empty repository, empty document",
			onConflictStrategy: OnConflictKeepExisting,
		},
		{
			tname:              "flat document with new bookmarks",
			userUUID:           "1632e701-e153-4f43-87ab-7fecacf8763f",
			onConflictStrategy: OnConflictKeepExisting,
			visibility:         VisibilityDefault,
			document: netscape.Document{
				Root: netscape.Folder{
					Bookmarks: []netscape.Bookmark{
						{
							Title: "Flat 1",
							URL:   "https://flat1.domain.tld",
							Tags:  []string{"flat", "test"},
						},
					},
				},
			},
			want: []bookmark.Bookmark{
				{
					UserUUID: "1632e701-e153-4f43-87ab-7fecacf8763f",
					Title:    "Flat 1",
					URL:      "https://flat1.domain.tld",
					Tags:     []string{"flat", "test"},
				},
			},
			wantStatus: Status{
				NewOrUpdated: 1,
			},
		},
		{
			tname: "flat document with new and conflicting bookmarks (keep existing)",
			repositoryBookmarks: []bookmark.Bookmark{
				{
					UserUUID: "1632e701-e153-4f43-87ab-7fecacf8763f",
					Title:    "Flat 1",
					URL:      "https://flat1.domain.tld",
					Tags:     []string{"flat", "test"},
				},
			},
			userUUID:           "1632e701-e153-4f43-87ab-7fecacf8763f",
			onConflictStrategy: OnConflictKeepExisting,
			visibility:         VisibilityDefault,
			document: netscape.Document{
				Root: netscape.Folder{
					Bookmarks: []netscape.Bookmark{
						{
							Title: "Flat 1",
							URL:   "https://flat1.domain.tld",
							Tags:  []string{"flat", "test"},
						},
						{
							Title: "Flat 2",
							URL:   "https://flat2.domain.tld",
							Tags:  []string{"flat", "test"},
						},
					},
				},
			},
			want: []bookmark.Bookmark{
				{
					UserUUID: "1632e701-e153-4f43-87ab-7fecacf8763f",
					Title:    "Flat 1",
					URL:      "https://flat1.domain.tld",
					Tags:     []string{"flat", "test"},
				},
				{
					UserUUID: "1632e701-e153-4f43-87ab-7fecacf8763f",
					Title:    "Flat 2",
					URL:      "https://flat2.domain.tld",
					Tags:     []string{"flat", "test"},
				},
			},
			wantStatus: Status{
				NewOrUpdated: 1,
				Skipped:      1,
			},
		},
		{
			tname: "flat document with new and conflicting bookmarks (overwrite)",
			repositoryBookmarks: []bookmark.Bookmark{
				{
					UserUUID: "1632e701-e153-4f43-87ab-7fecacf8763f",
					Title:    "Update Me!",
					URL:      "https://flat1.domain.tld",
					Tags:     []string{},
				},
			},
			userUUID:           "1632e701-e153-4f43-87ab-7fecacf8763f",
			onConflictStrategy: OnConflictOverwrite,
			visibility:         VisibilityDefault,
			document: netscape.Document{
				Root: netscape.Folder{
					Bookmarks: []netscape.Bookmark{
						{
							Title: "Flat 1",
							URL:   "https://flat1.domain.tld",
							Tags:  []string{"flat", "test"},
						},
						{
							Title:   "Flat 2",
							URL:     "https://flat2.domain.tld",
							Private: true,
							Tags:    []string{"flat", "test"},
						},
					},
				},
			},
			want: []bookmark.Bookmark{
				{
					UserUUID: "1632e701-e153-4f43-87ab-7fecacf8763f",
					Title:    "Flat 1",
					URL:      "https://flat1.domain.tld",
					Tags:     []string{"flat", "test"},
				},
				{
					UserUUID: "1632e701-e153-4f43-87ab-7fecacf8763f",
					Title:    "Flat 2",
					URL:      "https://flat2.domain.tld",
					Private:  true,
					Tags:     []string{"flat", "test"},
				},
			},
			wantStatus: Status{
				NewOrUpdated: 2,
			},
		},

		// error cases
		{
			tname:              "invalid on-conflict strategy",
			onConflictStrategy: "flee",
			wantErr:            ErrOnConflictStrategyInvalid,
		},
		{
			tname:              "invalid visibility",
			onConflictStrategy: OnConflictKeepExisting,
			visibility:         "foggy",
			document: netscape.Document{
				Root: netscape.Folder{
					Bookmarks: []netscape.Bookmark{
						{
							Title: "Flat 1",
							URL:   "https://flat1.domain.tld",
							Tags:  []string{"flat", "test"},
						},
					},
				},
			},
			wantErr: ErrVisibilityInvalid,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Bookmarks: tc.repositoryBookmarks,
			}

			s := NewService(r)

			status, err := s.ImportFromNetscapeDocument(tc.userUUID, &tc.document, tc.visibility, tc.onConflictStrategy)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("want error %q, got %q", tc.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			if status.NewOrUpdated != tc.wantStatus.NewOrUpdated {
				t.Errorf("want %d new bookmark(s), got %d", tc.wantStatus.NewOrUpdated, status.NewOrUpdated)
			}
			if status.Skipped != tc.wantStatus.Skipped {
				t.Errorf("want %d skipped bookmark(s), got %d", tc.wantStatus.Skipped, status.Skipped)
			}
			if status.Invalid != tc.wantStatus.Invalid {
				t.Errorf("want %d invalid bookmark(s), got %d", tc.wantStatus.Invalid, status.Invalid)
			}

			if len(r.Bookmarks) != len(tc.want) {
				t.Fatalf("want %d bookmark(s), got %d", len(tc.want), len(r.Bookmarks))
			}

			for index, wantBookmark := range tc.want {
				assertBookmarksEqual(t, r.Bookmarks[index], wantBookmark)
			}
		})
	}
}

func assertBookmarksEqual(t *testing.T, got, want bookmark.Bookmark) {
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
