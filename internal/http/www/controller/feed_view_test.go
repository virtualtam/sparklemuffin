// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package controller

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/virtualtam/sparklemuffin/internal/http/www/view"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	feedquerying "github.com/virtualtam/sparklemuffin/pkg/feed/querying"
)

func TestFeedEntryTemplate(t *testing.T) {
	v := view.New("feed/feed_list.gohtml")

	unreadEntry := feedquerying.SubscribedFeedEntry{
		Entry: feed.Entry{
			UID:         "entry-uid-1",
			URL:         "https://example.com/posts/1",
			Title:       "First Post",
			Summary:     "A short summary",
			PublishedAt: time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC),
		},
		FeedSlug:  "example-feed",
		FeedTitle: "Example Feed",
		Read:      false,
	}

	readEntry := unreadEntry
	readEntry.Read = true

	cases := []struct {
		tname              string
		entry              feedquerying.SubscribedFeedEntry
		showEntrySummaries bool
		wantContains       []string
		wantNotContains    []string
	}{
		{
			tname:              "unread entry, summaries shown",
			entry:              unreadEntry,
			showEntrySummaries: true,
			wantContains: []string{
				`id="feed-entry-entry-uid-1"`,
				`href="https://example.com/posts/1"`,
				"First Post",
				"A short summary",
				"Mark as read",
				`hx-post="/feeds/entries/entry-uid-1/toggle-read"`,
				`hx-target="#feed-entry-entry-uid-1"`,
				`hx-swap="outerHTML"`,
				`urlPath&#34;:&#34;/feeds`,
				`search&#34;:&#34;term`,
				`page&#34;:2`,
			},
			wantNotContains: []string{
				"has-text-grey-light",
				"<form",
			},
		},
		{
			tname:              "read entry, summaries hidden",
			entry:              readEntry,
			showEntrySummaries: false,
			wantContains: []string{
				`id="feed-entry-entry-uid-1"`,
				"has-text-grey-light",
				"Mark as unread",
				`hx-post="/feeds/entries/entry-uid-1/toggle-read"`,
			},
			wantNotContains: []string{
				"A short summary",
				"<form",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			data := map[string]any{
				"Entry":              tc.entry,
				"ShowEntrySummaries": tc.showEntrySummaries,
				"URLPath":            "/feeds",
				"SearchTerms":        "term",
				"PageNumber":         uint(2),
			}

			w := httptest.NewRecorder()

			if err := v.RenderTemplate(w, "feedEntry", data); err != nil {
				t.Fatalf("failed to render feedEntry template: %s", err)
			}

			body := w.Body.String()

			for _, want := range tc.wantContains {
				if !strings.Contains(body, want) {
					t.Errorf("want body to contain %q, got:\n%s", want, body)
				}
			}
			for _, notWant := range tc.wantNotContains {
				if strings.Contains(body, notWant) {
					t.Errorf("want body to NOT contain %q, got:\n%s", notWant, body)
				}
			}
		})
	}
}

func TestEntryListTemplate(t *testing.T) {
	v := view.New("feed/feed_list.gohtml")

	entry1 := feedquerying.SubscribedFeedEntry{
		Entry: feed.Entry{
			UID:         "entry-uid-1",
			URL:         "https://example.com/posts/1",
			Title:       "First Post",
			PublishedAt: time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC),
		},
		FeedSlug:  "example-feed",
		FeedTitle: "Example Feed",
	}

	entry2 := feedquerying.SubscribedFeedEntry{
		Entry: feed.Entry{
			UID:         "entry-uid-2",
			URL:         "https://example.com/posts/2",
			Title:       "Second Post",
			PublishedAt: time.Date(2026, time.July, 2, 0, 0, 0, 0, time.UTC),
		},
		FeedSlug:  "example-feed",
		FeedTitle: "Example Feed",
	}

	data := map[string]any{
		"Entries":            []feedquerying.SubscribedFeedEntry{entry1, entry2},
		"ItemOffset":         uint(1),
		"ShowEntrySummaries": true,
		"URLPath":            "/feeds",
		"SearchTerms":        "",
		"PageNumber":         uint(1),
	}

	w := httptest.NewRecorder()

	if err := v.RenderTemplate(w, "entryList", data); err != nil {
		t.Fatalf("failed to render entryList template: %s", err)
	}

	body := w.Body.String()

	wantContains := []string{
		`<ol id="entry-list" start="1">`,
		`id="feed-entry-entry-uid-1"`,
		`id="feed-entry-entry-uid-2"`,
		"First Post",
		"Second Post",
		`hx-post="/feeds/entries/entry-uid-1/toggle-read"`,
		`hx-post="/feeds/entries/entry-uid-2/toggle-read"`,
	}
	for _, want := range wantContains {
		if !strings.Contains(body, want) {
			t.Errorf("want body to contain %q, got:\n%s", want, body)
		}
	}
}
