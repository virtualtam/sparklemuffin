package importing

import (
	"testing"

	"github.com/virtualtam/netscape-go/v2"
	"github.com/virtualtam/yawbe/pkg/bookmark"
)

func TestServiceImportFromNetscapeDocument(t *testing.T) {
	cases := []struct {
		tname               string
		repositoryBookmarks []bookmark.Bookmark

		userUUID   string
		document   netscape.Document
		visibility Visibility

		want       []bookmark.Bookmark
		wantStatus Status
	}{
		{
			tname: "empty repository, empty document",
		},
		{
			tname:      "flat document with new bookmarks",
			userUUID:   "1632e701-e153-4f43-87ab-7fecacf8763f",
			visibility: VisibilityDefault,
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
				New: 1,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &FakeRepository{
				Bookmarks: tc.repositoryBookmarks,
			}

			s := NewService(r)

			status, err := s.ImportFromNetscapeDocument(tc.userUUID, &tc.document, tc.visibility)

			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			if status.New != tc.wantStatus.New {
				t.Errorf("want %d new bookmark(s), got %d", tc.wantStatus.New, status.New)
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
		})
	}
}
