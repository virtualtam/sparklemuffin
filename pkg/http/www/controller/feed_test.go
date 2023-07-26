package controller

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/feeds"

	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	"github.com/virtualtam/sparklemuffin/pkg/querying"
)

func TestBookmarksToFeed(t *testing.T) {
	publicURL := &url.URL{
		Scheme: "http",
		Host:   "test.domain.tld:8080",
	}

	owner := querying.Owner{
		NickName:    "test-user-1",
		DisplayName: "Test User",
	}

	now := time.Now().UTC()
	yesterday := now.AddDate(0, 0, -1)

	testCases := []struct {
		tname     string
		bookmarks []bookmark.Bookmark
		want      feeds.Feed
	}{
		{
			tname: "empty feed",
			want: feeds.Feed{
				Title: fmt.Sprintf("%s's bookmarks", owner.DisplayName),
				Link: &feeds.Link{
					Href: fmt.Sprintf("http://test.domain.tld:8080/u/%s/bookmarks", owner.NickName),
				},
				Author: &feeds.Author{
					Name: owner.DisplayName,
				},
			},
		},
		{
			tname: "single bookmark with no tags nor description",
			bookmarks: []bookmark.Bookmark{
				{
					UID:       "2NpRkS46UT88WPWL6c4Ni8e1Ial",
					Title:     "Test 1",
					URL:       "https://domain.tld/path",
					CreatedAt: yesterday,
					UpdatedAt: now,
				},
			},
			want: feeds.Feed{
				Title: fmt.Sprintf("%s's bookmarks", owner.DisplayName),
				Link: &feeds.Link{
					Href: fmt.Sprintf("http://test.domain.tld:8080/u/%s/bookmarks", owner.NickName),
				},
				Author: &feeds.Author{
					Name: owner.DisplayName,
				},
				Items: []*feeds.Item{
					{
						Id:    fmt.Sprintf("%s/u/%s/bookmarks/%s", publicURL.String(), owner.NickName, "2NpRkS46UT88WPWL6c4Ni8e1Ial"),
						Title: "Test 1",
						Link: &feeds.Link{
							Href: "https://domain.tld/path",
						},
						Created: yesterday,
						Updated: now,
					},
				},
			},
		},
		{
			tname: "single bookmark with tags and Markdown description",
			bookmarks: []bookmark.Bookmark{
				{
					UID:   "2NpRkS46UT88WPWL6c4Ni8e1Ial",
					Title: "Test 1",
					URL:   "https://domain.tld/path",
					Description: `Tags:
- feed/atom
- test
`,
					Tags:      []string{"test", "feed/atom"},
					CreatedAt: yesterday,
					UpdatedAt: now,
				},
			},
			want: feeds.Feed{
				Title: fmt.Sprintf("%s's bookmarks", owner.DisplayName),
				Link: &feeds.Link{
					Href: fmt.Sprintf("http://test.domain.tld:8080/u/%s/bookmarks", owner.NickName),
				},
				Author: &feeds.Author{
					Name: owner.DisplayName,
				},
				Items: []*feeds.Item{
					{
						Id:    fmt.Sprintf("%s/u/%s/bookmarks/%s", publicURL.String(), owner.NickName, "2NpRkS46UT88WPWL6c4Ni8e1Ial"),
						Title: "Test 1",
						Link: &feeds.Link{
							Href: "https://domain.tld/path",
						},
						Content: "<p>Tags:</p>\n<ul>\n<li>feed/atom</li>\n<li>test</li>\n</ul>\n",
						Created: yesterday,
						Updated: now,
					},
				},
			},
		},
		{
			tname: "multiple bookmarks",
			bookmarks: []bookmark.Bookmark{
				{
					UID:       "2NpRkS46UT88WPWL6c4Ni8e1Ial",
					Title:     "Test 1",
					URL:       "https://domain.tld/path",
					CreatedAt: yesterday,
					UpdatedAt: now,
				},
				{
					UID:   "2NpWy2Ncn9en8R1udvJ0KWrVwlc",
					Title: "Test 1",
					URL:   "https://domain.tld/path",
					Description: `Tags:
- feed/atom
- test
`,
					Tags:      []string{"test", "feed/atom"},
					CreatedAt: yesterday,
					UpdatedAt: now,
				},
			},
			want: feeds.Feed{
				Title: fmt.Sprintf("%s's bookmarks", owner.DisplayName),
				Link: &feeds.Link{
					Href: fmt.Sprintf("http://test.domain.tld:8080/u/%s/bookmarks", owner.NickName),
				},
				Author: &feeds.Author{
					Name: owner.DisplayName,
				},
				Items: []*feeds.Item{
					{
						Id:    fmt.Sprintf("%s/u/%s/bookmarks/%s", publicURL.String(), owner.NickName, "2NpRkS46UT88WPWL6c4Ni8e1Ial"),
						Title: "Test 1",
						Link: &feeds.Link{
							Href: "https://domain.tld/path",
						},
						Created: yesterday,
						Updated: now,
					},
					{
						Id:    fmt.Sprintf("%s/u/%s/bookmarks/%s", publicURL.String(), owner.NickName, "2NpWy2Ncn9en8R1udvJ0KWrVwlc"),
						Title: "Test 1",
						Link: &feeds.Link{
							Href: "https://domain.tld/path",
						},
						Content: "<p>Tags:</p>\n<ul>\n<li>feed/atom</li>\n<li>test</li>\n</ul>\n",
						Created: yesterday,
						Updated: now,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.tname, func(t *testing.T) {
			got, err := bookmarksToFeed(publicURL, owner, tc.bookmarks)
			if err != nil {
				t.Fatalf("expected no error, got %q", err)
			}

			assertFeedsEqual(t, got, &tc.want)
		})
	}
}

func assertFeedsEqual(t *testing.T, got, want *feeds.Feed) {
	t.Helper()

	if got.Title != want.Title {
		t.Errorf("want title %q, got %q", want.Title, got.Title)
	}
	if got.Link.Href != want.Link.Href {
		t.Errorf("want link %q, got %q", want.Link.Href, got.Link.Href)
	}
	if got.Author.Name != want.Author.Name {
		t.Errorf("want author %q, got %q", want.Author.Name, got.Author.Name)
	}

	if len(got.Items) != len(want.Items) {
		t.Fatalf("want %d items, got %d", len(want.Items), len(got.Items))
	}

	for i, gotItem := range got.Items {
		wantItem := want.Items[i]

		if gotItem.Id != wantItem.Id {
			t.Errorf("want item %d id %q, got %q", i, wantItem.Id, gotItem.Id)
		}
		if gotItem.Title != wantItem.Title {
			t.Errorf("want item %d title %q, got %q", i, wantItem.Title, gotItem.Title)
		}
		if gotItem.Link.Href != wantItem.Link.Href {
			t.Errorf("want item %d link %q, got %q", i, wantItem.Link.Href, gotItem.Link.Href)
		}
		if gotItem.Content != wantItem.Content {
			t.Errorf("want item %d content %q, got %q", i, wantItem.Content, gotItem.Content)
		}
		if gotItem.Created != wantItem.Created {
			t.Errorf("want item %d created %q, got %q", i, wantItem.Created, gotItem.Created)
		}
		if gotItem.Updated != wantItem.Updated {
			t.Errorf("want item %d updated %q, got %q", i, wantItem.Updated, gotItem.Updated)
		}
	}
}
