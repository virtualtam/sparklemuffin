// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package feed

import (
	"testing"
	"time"

	"github.com/jaswdr/faker"
	"github.com/mmcdole/gofeed"
	"github.com/segmentio/ksuid"
)

func TestEntryNormalize(t *testing.T) {
	cases := []struct {
		tname           string
		description     string
		content         string
		wantDescription string
		wantContent     string
	}{
		{
			tname: "empty content and description",
		},
		{
			tname:           "plain text content and description",
			description:     "A short description",
			content:         "Some plain text content",
			wantDescription: "A short description",
			wantContent:     "Some plain text content",
		},
		{
			tname:           "HTML content and description",
			description:     `A <strong>short</strong> description with <a href="https://example.com">link</a>`,
			content:         `<p>Some <em>formatted</em> content with a <a href="https://example.com">link</a></p>`,
			wantDescription: "A short description with link",
			wantContent:     "Some formatted content with a link",
		},
		{
			tname:       "multiline HTML content",
			description: "<p>Description</p>",
			content: `<article>
	<h1>Title</h1>
	<p>First paragraph</p>
	<p>Second paragraph</p>
	<ul>
		<li>Item 1</li>
		<li>Item 2</li>
	</ul>
</article>`,
			wantDescription: "Description",
			wantContent:     "Title\n\n First paragraph\n\n Second paragraph\n\n \n - Item 1 \n - Item 2",
		},
		{
			tname:           "invalid HTML",
			description:     "<p>Unclosed paragraph",
			content:         "<div>Unclosed div",
			wantDescription: "Unclosed paragraph",
			wantContent:     "Unclosed div",
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			e := Entry{
				description: tc.description,
				content:     tc.content,
			}

			e.Normalize()

			if e.description != tc.wantDescription {
				t.Errorf("want %q, got %q", tc.wantDescription, e.description)
			}

			if e.content != tc.wantContent {
				t.Errorf("want %q, got %q", tc.wantContent, e.content)
			}
		})
	}
}

func TestEntrySummarize(t *testing.T) {
	cases := []struct {
		tname       string
		description string
		content     string
		want        string
	}{
		{
			tname: "empty content and description",
		},
		{
			tname:       "short content",
			description: "A description",
			content:     "A very short content that should be kept as is",
			want:        "A very short content that should be kept as is",
		},
		{
			tname: "long content",
			content: `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do
eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad
minim veniam, quis nostrud exercitation ullamco laboris nisi ut
aliquip ex ea commodo consequat. Duis aute irure dolor in
reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla
pariatur. Excepteur sint occaecat cupidatat non proident, sunt in
culpa qui officia deserunt mollit anim id est laborum.`,
			want: "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qu…",
		},
		{
			tname:       "fallback to description when no content",
			description: "A short description that should be used as summary",
			want:        "A short description that should be used as summary",
		},
		{
			tname: "long description fallback",
			description: `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do
eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad
minim veniam, quis nostrud exercitation ullamco laboris nisi ut
aliquip ex ea commodo consequat. Duis aute irure dolor in
reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla
pariatur. Excepteur sint occaecat cupidatat non proident, sunt in
culpa qui officia deserunt mollit anim id est laborum.`,
			want: "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qu…",
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			e := Entry{
				description: tc.description,
				content:     tc.content,
			}

			e.Normalize()
			e.summarize()

			if got := e.Summary; got != tc.want {
				t.Errorf("want %q, got %q", tc.want, got)
			}
		})
	}
}

func TestEntryValidateForAddition(t *testing.T) {
	now := time.Now().UTC()
	fake := faker.New()
	feedUUID := fake.UUID().V4()

	cases := []struct {
		tname   string
		entry   Entry
		wantErr error
	}{
		// error cases
		{
			tname: "empty URL",
			entry: Entry{
				URL: "",
			},
			wantErr: ErrEntryURLRequired,
		},
		{
			tname: "invalid URL (no scheme)",
			entry: Entry{
				URL: "invalid",
			},
			wantErr: ErrEntryURLNoScheme,
		},
		{
			tname: "invalid URL (unsupported scheme)",
			entry: Entry{
				URL: "ftp://example.com",
			},
			wantErr: ErrEntryURLUnsupportedScheme,
		},
		{
			tname: "invalid URL (no host)",
			entry: Entry{
				URL: "https://",
			},
			wantErr: ErrEntryURLNoHost,
		},
		{
			tname: "empty title",
			entry: Entry{
				URL:   "https://example.com",
				Title: "",
			},
			wantErr: ErrEntryTitleRequired,
		},
		{
			tname: "publication date in the future",
			entry: Entry{
				UID:         ksuid.New().String(),
				FeedUUID:    feedUUID,
				URL:         "https://example.com",
				Title:       "Example Post",
				PublishedAt: now.Add(7 * 24 * time.Hour).UTC(),
			},
			wantErr: ErrEntryPublishedAtInTheFuture,
		},
		{
			tname: "update date in the future",
			entry: Entry{
				UID:       ksuid.New().String(),
				FeedUUID:  feedUUID,
				URL:       "https://example.com",
				Title:     "Example Post",
				UpdatedAt: now.Add(7 * 24 * time.Hour).UTC(),
			},
			wantErr: ErrEntryUpdatedAtInTheFuture,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			if err := tc.entry.ValidateForAddition(now); err != tc.wantErr {
				t.Errorf("want %q, got %q", tc.wantErr, err)
			}
		})
	}
}

func TestNewEntryFromItem(t *testing.T) {
	now := time.Now().UTC()
	fake := faker.New()
	feedUUID := fake.UUID().V4()

	cases := []struct {
		name        string
		item        *gofeed.Item
		wantContent string
		wantDesc    string
	}{
		{
			name: "item with content and description",
			item: &gofeed.Item{
				Title:       "Test Title",
				Link:        "https://example.com/post",
				Content:     "<p>Full content</p>",
				Description: "<p>Short description</p>",
			},
			wantContent: "Full content",
			wantDesc:    "Short description",
		},
		{
			name: "item with description only",
			item: &gofeed.Item{
				Title:       "Test Title",
				Link:        "https://example.com/post",
				Description: "<p>Short description</p>",
			},
			wantDesc: "Short description",
		},
		{
			name: "item with content only",
			item: &gofeed.Item{
				Title:   "Test Title",
				Link:    "https://example.com/post",
				Content: "<p>Full content</p>",
			},
			wantContent: "Full content",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			entry := NewEntryFromItem(feedUUID, now, tt.item)

			if entry.content != tt.wantContent {
				t.Errorf("want %q, got %q", tt.wantContent, entry.content)
			}

			if entry.description != tt.wantDesc {
				t.Errorf("want %q, got %q", tt.wantDesc, entry.description)
			}

			if entry.FeedUUID != feedUUID {
				t.Errorf("want %q, got %q", feedUUID, entry.FeedUUID)
			}

			if entry.Title != tt.item.Title {
				t.Errorf("want %q, got %q", tt.item.Title, entry.Title)
			}

			if entry.URL != tt.item.Link {
				t.Errorf("want %q, got %q", tt.item.Link, entry.URL)
			}
		})
	}
}
