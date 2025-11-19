// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package synchronizing

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/feeds"
	"github.com/jaswdr/faker/v2"

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
		Title:        "Local Test",
		Description:  "A simple syndication feed, for testing purposes.",
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
		Summary:     "First post!\n\nThis is the first post!",
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
					Description:  repositoryFeed.Description,
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
					Description:  repositoryFeed.Description,
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
					Description:  repositoryFeed.Description,
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
					Description:  repositoryFeed.Description,
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
					Description:  repositoryFeed.Description,
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
			tname:           "feed has a new title",
			repositoryFeeds: []feed.Feed{repositoryFeed},
			repositoryEntries: []feed.Entry{
				secondEntry,
				firstEntry,
			},
			atomFeed: feeds.Feed{
				Title:       "Same flavour, but blazingly faster!",
				Description: atomFeed.Description,
				Updated:     today,
				Items:       atomFeed.Items,
			},
			wantFeeds: []feed.Feed{
				{
					UUID:         repositoryFeed.UUID,
					FeedURL:      repositoryFeed.FeedURL,
					Title:        "Same flavour, but blazingly faster!",
					Description:  repositoryFeed.Description,
					Slug:         repositoryFeed.Slug,
					ETag:         `W/"d2e704d224c9df7337a8e07b6d504267a1caf6e13bc9eec82aea8cbfd55eb85b"`,
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
			tname:           "feed has a new description",
			repositoryFeeds: []feed.Feed{repositoryFeed},
			repositoryEntries: []feed.Entry{
				secondEntry,
				firstEntry,
			},
			atomFeed: feeds.Feed{
				Title:       atomFeed.Title,
				Description: "Updated description.",
				Updated:     today,
				Items:       atomFeed.Items,
			},
			wantFeeds: []feed.Feed{
				{
					UUID:         repositoryFeed.UUID,
					FeedURL:      repositoryFeed.FeedURL,
					Title:        repositoryFeed.Title,
					Description:  "Updated description.",
					Slug:         repositoryFeed.Slug,
					ETag:         `W/"bbc08d254d1abf71ac09a314d5886c45fe7bd45cd04fe4e31285f2da26b41962"`,
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
				Title:       atomFeed.Title,
				Description: atomFeed.Description,
				Updated:     tomorrow,
				Items: []*feeds.Item{
					{
						Id:    "http://test.local/second-post",
						Title: "Second post!",
						Link: &feeds.Link{
							Href: "http://test.local/second-post",
						},
						Description: "This is the second post!",
						Created:     tomorrow,
						Updated:     tomorrow,
					},
					atomFeed.Items[1],
					atomFeed.Items[0],
				},
			},
			wantFeeds: []feed.Feed{
				{
					UUID:         repositoryFeed.UUID,
					FeedURL:      repositoryFeed.FeedURL,
					Title:        repositoryFeed.Title,
					Description:  repositoryFeed.Description,
					Slug:         repositoryFeed.Slug,
					ETag:         `W/"1ae6400e4431ee18962bf860e3b3d9bc9e16bd81053d97cd57df5fa3d3313b49"`,
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
					FeedUUID:      repositoryFeed.UUID,
					URL:           "http://test.local/second-post",
					Title:         "Second post!",
					Summary:       "This is the second post!",
					TextRankTerms: []string{"post second"},
					PublishedAt:   tomorrow,
					UpdatedAt:     tomorrow,
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
				Title:       repositoryFeed.Title,
				Description: atomFeed.Description,
				Updated:     atomFeed.Updated,
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
					Description:  repositoryFeed.Description,
					Slug:         repositoryFeed.Slug,
					ETag:         `W/"f3ea5b4ab75e6a1673798ed07f90659588a51cafd06d99b597922bbfd2d9e3b8"`,
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

			err := s.Synchronize(t.Context(), tc.tname)

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

			feed.AssertFeedsEqual(t, r.Feeds, tc.wantFeeds)
			feed.AssertEntriesEqual(t, r.Entries, tc.wantEntries)
		})
	}
}
