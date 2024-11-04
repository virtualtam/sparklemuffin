// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package synchronizing

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/feeds"
	"github.com/jaswdr/faker"
	"github.com/virtualtam/sparklemuffin/internal/assert"
	"github.com/virtualtam/sparklemuffin/internal/test/feedtest"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	"github.com/virtualtam/sparklemuffin/pkg/feed/fetching"
)

func TestServiceSynchronize(t *testing.T) {
	fake := faker.New()

	now := time.Now().UTC()

	// hardcode dates for feed data to ease reproducibility (e.g. for the ETag header)
	today, err := time.Parse(time.DateTime, "2024-10-30 20:54:16")
	if err != nil {
		t.Fatalf("failed to parse date: %q", err)
	}
	yesterday := today.Add(-24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	atomFeed := feedtest.GenerateDummyFeed(t, today)

	feedStr, err := atomFeed.ToAtom()
	if err != nil {
		t.Fatalf("failed to encode feed to Atom: %q", err)
	}

	feedETag := feedtest.HashETag(feedStr)
	feedLastModified := today

	repositoryFeed := feed.Feed{
		UUID:         fake.UUID().V4(),
		FeedURL:      "http://test.local",
		Title:        "Sync Test",
		Slug:         "sync-test",
		ETag:         feedETag,
		LastModified: feedLastModified,
		CreatedAt:    yesterday,
		UpdatedAt:    yesterday,
		FetchedAt:    yesterday,
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
		PublishedAt: today,
		UpdatedAt:   today,
	}

	cases := []struct {
		tname string

		// initial repository state
		repositoryFeeds   []feed.Feed
		repositoryEntries []feed.Entry

		// remote syndication feed
		atomFeed feeds.Feed

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
					UUID:         repositoryFeed.UUID,
					FeedURL:      repositoryFeed.FeedURL,
					Title:        repositoryFeed.Title,
					Slug:         repositoryFeed.Slug,
					ETag:         repositoryFeed.ETag,
					LastModified: repositoryFeed.LastModified,
					CreatedAt:    yesterday,
					UpdatedAt:    yesterday,
					FetchedAt:    now, // -> skip synchronization
				},
			},
			repositoryEntries: []feed.Entry{
				secondEntry,
				firstEntry,
			},
			wantFeeds: []feed.Feed{
				{
					UUID:         repositoryFeed.UUID,
					FeedURL:      repositoryFeed.FeedURL,
					Title:        repositoryFeed.Title,
					Slug:         repositoryFeed.Slug,
					ETag:         repositoryFeed.ETag,
					LastModified: repositoryFeed.LastModified,
					CreatedAt:    yesterday,
					UpdatedAt:    yesterday,
					FetchedAt:    now,
				},
			},
			wantEntries: []feed.Entry{
				secondEntry,
				firstEntry,
			},
		},
		{
			tname:           "ETag and Last-Modified match: feed metadata updated, feed entry update skipped",
			repositoryFeeds: []feed.Feed{repositoryFeed},
			repositoryEntries: []feed.Entry{
				secondEntry,
				firstEntry,
			},
			atomFeed: atomFeed,
			wantFeeds: []feed.Feed{
				{
					UUID:         repositoryFeed.UUID,
					FeedURL:      repositoryFeed.FeedURL,
					Title:        repositoryFeed.Title,
					Slug:         repositoryFeed.Slug,
					ETag:         repositoryFeed.ETag,
					LastModified: repositoryFeed.LastModified,
					CreatedAt:    repositoryFeed.CreatedAt,
					UpdatedAt:    now,
					FetchedAt:    now,
				},
			},
			wantEntries: []feed.Entry{
				secondEntry,
				firstEntry,
			},
		},
		{
			tname: "ETag does not match: feed metadata updated, feed entries updated (no change)",
			repositoryFeeds: []feed.Feed{
				{
					UUID:         repositoryFeed.UUID,
					FeedURL:      repositoryFeed.FeedURL,
					Title:        repositoryFeed.Title,
					Slug:         repositoryFeed.Slug,
					ETag:         feedtest.HashETag("does-not-match"),
					LastModified: repositoryFeed.LastModified,
					CreatedAt:    repositoryFeed.CreatedAt,
					UpdatedAt:    repositoryFeed.UpdatedAt,
					FetchedAt:    repositoryFeed.FetchedAt,
				},
			},
			repositoryEntries: []feed.Entry{
				secondEntry,
				firstEntry,
			},
			atomFeed: atomFeed,
			wantFeeds: []feed.Feed{
				{
					UUID:         repositoryFeed.UUID,
					FeedURL:      repositoryFeed.FeedURL,
					Title:        repositoryFeed.Title,
					Slug:         repositoryFeed.Slug,
					ETag:         repositoryFeed.ETag,
					LastModified: repositoryFeed.LastModified,
					CreatedAt:    repositoryFeed.CreatedAt,
					UpdatedAt:    now,
					FetchedAt:    now,
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
			atomFeed: feeds.Feed{
				Title:   atomFeed.Title,
				Updated: tomorrow,
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
						Created: today,
						Updated: today,
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
					UUID:         repositoryFeed.UUID,
					FeedURL:      repositoryFeed.FeedURL,
					Title:        repositoryFeed.Title,
					Slug:         repositoryFeed.Slug,
					ETag:         `W/"e13c781ba03006c00fc2de9a6aefd364c391bf8790b81bb90c4088d30c9ab0c0"`,
					LastModified: tomorrow,
					CreatedAt:    yesterday,
					UpdatedAt:    now,
					FetchedAt:    now,
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
			atomFeed: feeds.Feed{
				Title:   repositoryFeed.Title,
				Updated: atomFeed.Updated,
				Items: []*feeds.Item{
					{
						Id:    "http://test.local/first-post",
						Title: "My Actual First post! (Updated)",
						Link: &feeds.Link{
							Href: "http://test.local/first-post",
						},
						Created: today,
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
					UUID:         repositoryFeed.UUID,
					FeedURL:      repositoryFeed.FeedURL,
					Title:        repositoryFeed.Title,
					Slug:         repositoryFeed.Slug,
					ETag:         `W/"0d3624583a1afa409004483fd625d4b9c4a2151c420eb56673bff1d562e8d3d2"`,
					LastModified: feedLastModified,
					CreatedAt:    yesterday,
					UpdatedAt:    now,
					FetchedAt:    now,
				},
			},
			wantEntries: []feed.Entry{
				{
					FeedUUID:    secondEntry.FeedUUID,
					URL:         secondEntry.URL,
					Title:       "My Actual First post! (Updated)",
					PublishedAt: today,
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

			transport := feedtest.NewRoundTripper(t, tc.atomFeed)
			feedHTTPClient := &http.Client{
				Transport: transport,
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

		if gotFeed.ETag != wantFeed.ETag {
			t.Errorf("want ETag %q, got %q", wantFeed.ETag, gotFeed.ETag)
		}

		assert.TimeAlmostEquals(t, "LastModified", gotFeed.LastModified, wantFeed.LastModified, assert.TimeComparisonDelta)
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
			t.Errorf("want entry FeedUUID %q, got %q", wantEntry.FeedUUID, gotEntry.FeedUUID)
		}
		if gotEntry.Title != wantEntry.Title {
			t.Errorf("want entry Title %q, got %q", wantEntry.Title, gotEntry.Title)
		}
		if gotEntry.URL != wantEntry.URL {
			t.Errorf("want entry URL %q, got %q", wantEntry.URL, gotEntry.URL)
		}

		assert.TimeAlmostEquals(t, "entry PublishedAt", gotEntry.PublishedAt, wantEntry.PublishedAt, assert.TimeComparisonDelta)
		assert.TimeAlmostEquals(t, "entry UpdatedAt", gotEntry.UpdatedAt, wantEntry.UpdatedAt, assert.TimeComparisonDelta)
	}
}
