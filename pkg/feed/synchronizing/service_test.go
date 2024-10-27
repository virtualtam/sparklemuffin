// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package synchronizing

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/feeds"
	"github.com/jaswdr/faker"
	"github.com/virtualtam/sparklemuffin/internal/assert"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	"github.com/virtualtam/sparklemuffin/pkg/feed/fetching"
)

func TestServiceSynchronize(t *testing.T) {
	fake := faker.New()

	now := time.Now().UTC()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)

	repositoryFeed := feed.Feed{
		UUID:      fake.UUID().V4(),
		FeedURL:   "http://test.local",
		Title:     "Sync Test",
		Slug:      "sync-test",
		CreatedAt: yesterday,
		UpdatedAt: yesterday,
		FetchedAt: yesterday,
	}

	firstEntry := feed.Entry{
		FeedUUID:    repositoryFeed.UUID,
		URL:         "http://test.local/hello-world",
		Title:       "Hello World",
		PublishedAt: yesterday,
		UpdatedAt:   yesterday,
	}
	secondEntry := feed.Entry{
		FeedUUID:    repositoryFeed.UUID,
		URL:         "http://test.local/first-post",
		Title:       "First post!",
		PublishedAt: now,
		UpdatedAt:   now,
	}

	cases := []struct {
		tname string

		// initial repository state
		repositoryFeeds   []feed.Feed
		repositoryEntries []feed.Entry

		// remote syndication feed
		syndicationFeed feeds.Feed

		// expected repository state
		wantFeeds   []feed.Feed
		wantEntries []feed.Entry
		wantErr     error
	}{
		// nominal cases
		{
			tname: "synchronized recently, nothing to do",
			repositoryFeeds: []feed.Feed{
				{
					UUID:      repositoryFeed.UUID,
					FeedURL:   repositoryFeed.FeedURL,
					Title:     repositoryFeed.Title,
					Slug:      repositoryFeed.Slug,
					CreatedAt: yesterday,
					UpdatedAt: yesterday,
					FetchedAt: now,
				},
			},
			repositoryEntries: []feed.Entry{
				secondEntry,
				firstEntry,
			},
			wantFeeds: []feed.Feed{
				{
					UUID:      repositoryFeed.UUID,
					FeedURL:   repositoryFeed.FeedURL,
					Title:     repositoryFeed.Title,
					Slug:      repositoryFeed.Slug,
					CreatedAt: yesterday,
					UpdatedAt: yesterday,
					FetchedAt: now,
				},
			},
			wantEntries: []feed.Entry{
				secondEntry,
				firstEntry,
			},
		},
		{
			tname:           "feed updated with no changes",
			repositoryFeeds: []feed.Feed{repositoryFeed},
			repositoryEntries: []feed.Entry{
				secondEntry,
				firstEntry,
			},
			syndicationFeed: feeds.Feed{
				Title:   repositoryFeed.Title,
				Updated: repositoryFeed.FetchedAt,
				Items: []*feeds.Item{
					{
						Id:    "http://test.local/first-post",
						Title: "First post!",
						Link: &feeds.Link{
							Href: "http://test.local/first-post",
						},
						Created: now,
						Updated: now,
					},
					{
						Id:    "http://test.local/hello-world",
						Title: "Hello World",
						Link: &feeds.Link{
							Href: "http://test.local/hello-world",
						},
						Created: yesterday,
						Updated: yesterday,
					},
				},
			},
			wantFeeds: []feed.Feed{
				{
					UUID:      repositoryFeed.UUID,
					FeedURL:   repositoryFeed.FeedURL,
					Title:     repositoryFeed.Title,
					Slug:      repositoryFeed.Slug,
					CreatedAt: yesterday,
					UpdatedAt: yesterday,
					FetchedAt: now,
				},
			},
			wantEntries: []feed.Entry{
				secondEntry,
				firstEntry,
			},
		},
		{
			tname:           "feed has a new entry",
			repositoryFeeds: []feed.Feed{repositoryFeed},
			repositoryEntries: []feed.Entry{
				secondEntry,
				firstEntry,
			},
			syndicationFeed: feeds.Feed{
				Title:   repositoryFeed.Title,
				Updated: repositoryFeed.FetchedAt,
				Items: []*feeds.Item{
					{
						Id:    "http://test.local/second-post",
						Title: "Second post!",
						Link: &feeds.Link{
							Href: "http://test.local/second-post",
						},
						Created: tomorrow,
						Updated: tomorrow,
					},
					{
						Id:    "http://test.local/first-post",
						Title: "First post!",
						Link: &feeds.Link{
							Href: "http://test.local/first-post",
						},
						Created: now,
						Updated: now,
					},
					{
						Id:    "http://test.local/hello-world",
						Title: "Hello World",
						Link: &feeds.Link{
							Href: "http://test.local/hello-world",
						},
						Created: yesterday,
						Updated: yesterday,
					},
				},
			},
			wantFeeds: []feed.Feed{
				{
					UUID:      repositoryFeed.UUID,
					FeedURL:   repositoryFeed.FeedURL,
					Title:     repositoryFeed.Title,
					Slug:      repositoryFeed.Slug,
					CreatedAt: yesterday,
					UpdatedAt: yesterday,
					FetchedAt: now,
				},
			},
			wantEntries: []feed.Entry{
				secondEntry,
				firstEntry,
				{
					FeedUUID:    repositoryFeed.UUID,
					URL:         "http://test.local/second-post",
					Title:       "Second post!",
					PublishedAt: tomorrow,
					UpdatedAt:   tomorrow,
				},
			},
		},
		{
			tname:           "feed has an update for an existing entry",
			repositoryFeeds: []feed.Feed{repositoryFeed},
			repositoryEntries: []feed.Entry{
				secondEntry,
				firstEntry,
			},
			syndicationFeed: feeds.Feed{
				Title:   repositoryFeed.Title,
				Updated: repositoryFeed.FetchedAt,
				Items: []*feeds.Item{
					{
						Id:    "http://test.local/first-post",
						Title: "My Actual First post! (Updated)",
						Link: &feeds.Link{
							Href: "http://test.local/first-post",
						},
						Created: now,
						Updated: tomorrow,
					},
					{
						Id:    "http://test.local/hello-world",
						Title: "Hello World",
						Link: &feeds.Link{
							Href: "http://test.local/hello-world",
						},
						Created: yesterday,
						Updated: yesterday,
					},
				},
			},
			wantFeeds: []feed.Feed{
				{
					UUID:      repositoryFeed.UUID,
					FeedURL:   repositoryFeed.FeedURL,
					Title:     repositoryFeed.Title,
					Slug:      repositoryFeed.Slug,
					CreatedAt: yesterday,
					UpdatedAt: yesterday,
					FetchedAt: now,
				},
			},
			wantEntries: []feed.Entry{
				{
					FeedUUID:    secondEntry.FeedUUID,
					URL:         secondEntry.URL,
					Title:       "My Actual First post! (Updated)",
					PublishedAt: now,
					UpdatedAt:   tomorrow,
				},
				firstEntry,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			r := &fakeRepository{
				Feeds:   tc.repositoryFeeds,
				Entries: tc.repositoryEntries,
			}

			feedHTTPClient := &http.Client{
				Transport: &feedHTTPClient{
					syndicationFeed: &tc.syndicationFeed,
				},
			}

			feedClient := fetching.NewClient(feedHTTPClient, "sparklemuffin/test")

			s := NewService(r, feedClient)

			err := s.Synchronize(tc.tname)

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

			assertFeedsEqual(t, r.Feeds, tc.wantFeeds)
			assertEntriesEqual(t, r.Entries, tc.wantEntries)
		})
	}
}

var _ http.RoundTripper = &feedHTTPClient{}

type feedHTTPClient struct {
	syndicationFeed *feeds.Feed
}

func (c *feedHTTPClient) RoundTrip(r *http.Request) (*http.Response, error) {
	feedStr, err := c.syndicationFeed.ToAtom()
	if err != nil {
		panic(err)
	}

	resp := &http.Response{
		StatusCode: 200,
		Header: map[string][]string{
			"Content-Disposition": {"attachment; filename=test.atom"},
			"Content-Type":        {"application/atom+xml"},
		},
		Body: io.NopCloser(strings.NewReader(feedStr)),
	}

	return resp, nil
}

func assertFeedsEqual(t *testing.T, got, want []feed.Feed) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("want %d feeds, got %d", len(want), len(got))
	}

	for i, wantFeed := range want {
		gotFeed := got[i]

		if gotFeed.Slug != wantFeed.Slug {
			t.Errorf("want Slug %q, got %q", wantFeed.Slug, gotFeed.Slug)
		}
		if gotFeed.Title != wantFeed.Title {
			t.Errorf("want Title %q, got %q", wantFeed.Title, gotFeed.Title)
		}
		if gotFeed.FeedURL != wantFeed.FeedURL {
			t.Errorf("want FeedURL %q, got %q", wantFeed.FeedURL, gotFeed.FeedURL)
		}

		assert.TimeAlmostEquals(t, "CreatedAt", gotFeed.CreatedAt, wantFeed.CreatedAt, assert.TimeComparisonDelta)
		assert.TimeAlmostEquals(t, "UpdatedAt", gotFeed.UpdatedAt, wantFeed.UpdatedAt, assert.TimeComparisonDelta)
		assert.TimeAlmostEquals(t, "FetchedAt", gotFeed.FetchedAt, wantFeed.FetchedAt, assert.TimeComparisonDelta)
	}
}

func assertEntriesEqual(t *testing.T, got, want []feed.Entry) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("want %d entries, got %d", len(want), len(got))
	}

	for i, wantEntry := range want {
		gotEntry := got[i]

		if gotEntry.FeedUUID != wantEntry.FeedUUID {
			t.Errorf("want FeedUUID %q, got %q", wantEntry.FeedUUID, gotEntry.FeedUUID)
		}
		if gotEntry.Title != wantEntry.Title {
			t.Errorf("want Title %q, got %q", wantEntry.Title, gotEntry.Title)
		}
		if gotEntry.URL != wantEntry.URL {
			t.Errorf("want URL %q, got %q", wantEntry.URL, gotEntry.URL)
		}

		assert.TimeAlmostEquals(t, "PublishedAt", gotEntry.PublishedAt, wantEntry.PublishedAt, assert.TimeComparisonDelta)
		assert.TimeAlmostEquals(t, "UpdatedAt", gotEntry.UpdatedAt, wantEntry.UpdatedAt, assert.TimeComparisonDelta)
	}
}
