// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package querying

import (
	"errors"
	"testing"

	"github.com/jaswdr/faker/v2"

	"github.com/virtualtam/sparklemuffin/internal/paginate"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

func TestService(t *testing.T) {
	// initialize test data
	fake := faker.New()

	feed1 := feed.Feed{
		UUID:    fake.UUID().V4(),
		FeedURL: "http://test.local/feed.atom",
		Title:   "Local Test",
		Slug:    "local-test",
	}

	feed1Entry1 := feed.Entry{
		UID:      "1",
		FeedUUID: feed1.UUID,
		URL:      "http://test.local/posts/1",
		Title:    "First Post",
	}

	feed1Entry2 := feed.Entry{
		UID:      "2",
		FeedUUID: feed1.UUID,
		URL:      "http://test.local/posts/2",
		Title:    "Second Post",
	}

	user1 := user.User{
		UUID: fake.UUID().V4(),
	}

	user1Category1 := feed.Category{
		UUID:     fake.UUID().V4(),
		UserUUID: user1.UUID,
		Name:     "Test Category",
		Slug:     "test-category",
	}

	user1Category2 := feed.Category{
		UUID:     fake.UUID().V4(),
		UserUUID: user1.UUID,
		Name:     "Empty Category",
		Slug:     "empty-category",
	}

	user1Subscription1 := feed.Subscription{
		UUID:         fake.UUID().V4(),
		CategoryUUID: user1Category1.UUID,
		FeedUUID:     feed1.UUID,
		UserUUID:     user1.UUID,
		Alias:        "Feed #1",
	}

	user1Feed1Entry2Metadata := feed.EntryMetadata{
		UserUUID: user1.UUID,
		EntryUID: feed1Entry2.UID,
		Read:     true,
	}

	preferences := feed.Preferences{
		ShowEntries: feed.EntryVisibilityAll,
	}

	testRepository := FakeRepository{
		Categories: []feed.Category{
			user1Category1,
			user1Category2,
		},
		Entries: []feed.Entry{
			feed1Entry1,
			feed1Entry2,
		},
		EntriesMetadata: []feed.EntryMetadata{
			user1Feed1Entry2Metadata,
		},
		Feeds: []feed.Feed{
			feed1,
		},
		Subscriptions: []feed.Subscription{
			user1Subscription1,
		},
	}

	testService := NewService(&testRepository)

	t.Run("FeedsByPage", func(t *testing.T) {
		cases := []struct {
			tname      string
			userUUID   string
			pageNumber uint
			want       FeedPage
			wantErr    error
		}{
			// nominal cases
			{
				tname:      "one page, no subscription",
				pageNumber: 1,
				want: FeedPage{
					Page: paginate.Page{
						PageNumber:         1,
						PreviousPageNumber: 1,
						NextPageNumber:     1,
						TotalPages:         1,
						ItemOffset:         1,
					},
					PageTitle: PageHeaderAll,
				},
			},
			{
				tname:      "one page, one category with one subscription, one empty category",
				userUUID:   user1.UUID,
				pageNumber: 1,
				want: FeedPage{
					Page: paginate.Page{
						PageNumber:         1,
						PreviousPageNumber: 1,
						NextPageNumber:     1,
						TotalPages:         1,
						ItemOffset:         1,
						ItemCount:          2,
					},
					PageTitle: PageHeaderAll,
					Unread:    1,

					Categories: []SubscribedFeedsByCategory{
						{
							Category: user1Category2,
						},
						{
							Category: user1Category1,
							Unread:   1,
							SubscribedFeeds: []SubscribedFeed{
								{
									Feed:   feed1,
									Unread: 1,
								},
							},
						},
					},
					Entries: []SubscribedFeedEntry{
						{
							Entry:             feed1Entry1,
							FeedSlug:          feed1.Slug,
							FeedTitle:         feed1.Title,
							SubscriptionAlias: user1Subscription1.Alias,
							Read:              false,
						},
						{
							Entry:             feed1Entry2,
							FeedSlug:          feed1.Slug,
							FeedTitle:         feed1.Title,
							SubscriptionAlias: user1Subscription1.Alias,
							Read:              true,
						},
					},
				},
			},

			// error cases
			{
				tname:      "zeroth page",
				pageNumber: 0,
				userUUID:   user1.UUID,
				wantErr:    paginate.ErrPageNumberOutOfBounds,
			},
			{
				tname:      "page number out of bounds",
				pageNumber: 18,
				userUUID:   user1.UUID,
				wantErr:    paginate.ErrPageNumberOutOfBounds,
			},
		}

		for _, tc := range cases {
			t.Run(tc.tname, func(t *testing.T) {
				got, err := testService.FeedsByPage(t.Context(), tc.userUUID, preferences, tc.pageNumber)

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

				AssertPageEquals(t, got, tc.want)
			})
		}
	})

	t.Run("FeedsByCategoryAndPage", func(t *testing.T) {
		cases := []struct {
			tname      string
			userUUID   string
			category   feed.Category
			pageNumber uint
			want       FeedPage
			wantErr    error
		}{
			// nominal cases
			{
				tname:      "one page, no subscription",
				userUUID:   user1.UUID,
				category:   user1Category2,
				pageNumber: 1,
				want: FeedPage{
					Page: paginate.Page{
						PageNumber:         1,
						PreviousPageNumber: 1,
						NextPageNumber:     1,
						TotalPages:         1,
						ItemOffset:         1,
					},

					PageTitle: "Empty Category",
					Unread:    1,

					Categories: []SubscribedFeedsByCategory{
						{
							Category: user1Category2,
						},
						{
							Category: user1Category1,
							Unread:   1,
							SubscribedFeeds: []SubscribedFeed{
								{
									Feed:   feed1,
									Unread: 1,
								},
							},
						},
					},
				},
			},
			{
				tname:      "one page, one category with one subscription, one empty category",
				userUUID:   user1.UUID,
				category:   user1Category1,
				pageNumber: 1,
				want: FeedPage{
					Page: paginate.Page{
						PageNumber:         1,
						PreviousPageNumber: 1,
						NextPageNumber:     1,
						TotalPages:         1,
						ItemOffset:         1,
						ItemCount:          2,
					},

					PageTitle: "Test Category",
					Unread:    1,

					Categories: []SubscribedFeedsByCategory{
						{
							Category: user1Category2,
						},
						{
							Category: user1Category1,
							Unread:   1,
							SubscribedFeeds: []SubscribedFeed{
								{
									Feed:   feed1,
									Unread: 1,
								},
							},
						},
					},
					Entries: []SubscribedFeedEntry{
						{
							Entry:             feed1Entry1,
							FeedSlug:          feed1.Slug,
							FeedTitle:         feed1.Title,
							SubscriptionAlias: user1Subscription1.Alias,
							Read:              false,
						},
						{
							Entry:             feed1Entry2,
							FeedSlug:          feed1.Slug,
							FeedTitle:         feed1.Title,
							SubscriptionAlias: user1Subscription1.Alias,
							Read:              true,
						},
					},
				},
			},

			// error cases
			{
				tname:      "zeroth page",
				pageNumber: 0,
				userUUID:   user1.UUID,
				wantErr:    paginate.ErrPageNumberOutOfBounds,
			},
			{
				tname:      "page number out of bounds",
				pageNumber: 18,
				userUUID:   user1.UUID,
				wantErr:    paginate.ErrPageNumberOutOfBounds,
			},
		}

		for _, tc := range cases {
			t.Run(tc.tname, func(t *testing.T) {
				got, err := testService.FeedsByCategoryAndPage(t.Context(), tc.userUUID, preferences, tc.category, tc.pageNumber)

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

				AssertPageEquals(t, got, tc.want)
			})
		}
	})

	t.Run("FeedsBySubscriptionAndPage", func(t *testing.T) {
		cases := []struct {
			tname        string
			userUUID     string
			subscription feed.Subscription
			pageNumber   uint
			want         FeedPage
			wantErr      error
		}{
			// nominal cases
			{
				tname:    "one page",
				userUUID: user1.UUID,
				subscription: feed.Subscription{
					UUID:     testRepository.Subscriptions[0].UUID,
					FeedUUID: testRepository.Feeds[0].UUID,
				},
				pageNumber: 1,
				want: FeedPage{
					Page: paginate.Page{
						PageNumber:         1,
						PreviousPageNumber: 1,
						NextPageNumber:     1,
						TotalPages:         1,
						ItemOffset:         1,
						ItemCount:          2,
					},

					PageTitle: "Local Test",
					Unread:    1,

					Categories: []SubscribedFeedsByCategory{
						{
							Category: user1Category2,
						},
						{
							Category: user1Category1,
							Unread:   1,
							SubscribedFeeds: []SubscribedFeed{
								{
									Feed:   feed1,
									Unread: 1,
								},
							},
						},
					},
					Entries: []SubscribedFeedEntry{
						{
							Entry:             feed1Entry1,
							FeedSlug:          feed1.Slug,
							FeedTitle:         feed1.Title,
							SubscriptionAlias: user1Subscription1.Alias,
							Read:              false,
						},
						{
							Entry:             feed1Entry2,
							FeedSlug:          feed1.Slug,
							FeedTitle:         feed1.Title,
							SubscriptionAlias: user1Subscription1.Alias,
							Read:              true,
						},
					},
				},
			},

			// error cases
			{
				tname:      "zeroth page",
				pageNumber: 0,
				userUUID:   user1.UUID,
				subscription: feed.Subscription{
					UUID:     testRepository.Subscriptions[0].UUID,
					FeedUUID: testRepository.Feeds[0].UUID,
				},
				wantErr: paginate.ErrPageNumberOutOfBounds,
			},
			{
				tname: "page number out of bounds",
				subscription: feed.Subscription{
					UUID:     testRepository.Subscriptions[0].UUID,
					FeedUUID: testRepository.Feeds[0].UUID,
				},
				pageNumber: 18,
				userUUID:   user1.UUID,
				wantErr:    paginate.ErrPageNumberOutOfBounds,
			},
		}

		for _, tc := range cases {
			t.Run(tc.tname, func(t *testing.T) {
				got, err := testService.FeedsBySubscriptionAndPage(t.Context(), tc.userUUID, preferences, tc.subscription, tc.pageNumber)

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

				AssertPageEquals(t, got, tc.want)
			})
		}
	})

	t.Run("SubscribedFeedEntryByUID", func(t *testing.T) {
		cases := []struct {
			tname    string
			userUUID string
			entryUID string
			want     SubscribedFeedEntry
			wantErr  error
		}{
			// nominal cases
			{
				tname:    "unread entry",
				userUUID: user1.UUID,
				entryUID: feed1Entry1.UID,
				want: SubscribedFeedEntry{
					Entry:             feed1Entry1,
					FeedSlug:          feed1.Slug,
					FeedTitle:         feed1.Title,
					SubscriptionAlias: user1Subscription1.Alias,
					Read:              false,
				},
			},
			{
				tname:    "read entry",
				userUUID: user1.UUID,
				entryUID: feed1Entry2.UID,
				want: SubscribedFeedEntry{
					Entry:             feed1Entry2,
					FeedSlug:          feed1.Slug,
					FeedTitle:         feed1.Title,
					SubscriptionAlias: user1Subscription1.Alias,
					Read:              true,
				},
			},

			// error cases
			{
				tname:    "entry not found",
				userUUID: user1.UUID,
				entryUID: "does-not-exist",
				wantErr:  feed.ErrEntryNotFound,
			},
		}

		for _, tc := range cases {
			t.Run(tc.tname, func(t *testing.T) {
				got, err := testService.SubscribedFeedEntryByUID(t.Context(), tc.userUUID, tc.entryUID)

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

				AssertSubscriptionEntriesEqual(t, []SubscribedFeedEntry{got}, []SubscribedFeedEntry{tc.want})
			})
		}
	})
}

func TestServiceEntriesCarrySubscriptionTags(t *testing.T) {
	fake := faker.New()

	userUUID := fake.UUID().V4()

	category := feed.Category{
		UUID:     fake.UUID().V4(),
		UserUUID: userUUID,
		Name:     "Test Category",
		Slug:     "test-category",
	}

	f := feed.Feed{
		UUID:    fake.UUID().V4(),
		FeedURL: "http://test.local/feed.atom",
		Title:   "Local Test",
		Slug:    "local-test",
	}

	entry := feed.Entry{
		UID:      "1",
		FeedUUID: f.UUID,
		URL:      "http://test.local/posts/1",
		Title:    "First Post",
	}

	subscription := feed.Subscription{
		UUID:         fake.UUID().V4(),
		CategoryUUID: category.UUID,
		FeedUUID:     f.UUID,
		UserUUID:     userUUID,
		Tags:         []string{"golang", "rss"},
	}

	testRepository := FakeRepository{
		Categories:    []feed.Category{category},
		Entries:       []feed.Entry{entry},
		Feeds:         []feed.Feed{f},
		Subscriptions: []feed.Subscription{subscription},
	}

	testService := NewService(&testRepository)
	preferences := feed.Preferences{ShowEntries: feed.EntryVisibilityAll}

	got, err := testService.FeedsByPage(t.Context(), userUUID, preferences, 1)
	if err != nil {
		t.Fatalf("want no error, got %q", err)
	}

	want := []SubscribedFeedEntry{
		{
			Entry:            entry,
			FeedSlug:         f.Slug,
			FeedTitle:        f.Title,
			SubscriptionTags: subscription.Tags,
		},
	}

	AssertSubscriptionEntriesEqual(t, got.Entries, want)
}

func TestServiceSubscriptionCountsByTag(t *testing.T) {
	fake := faker.New()

	userUUID := fake.UUID().V4()
	otherUserUUID := fake.UUID().V4()

	testRepository := FakeRepository{
		Subscriptions: []feed.Subscription{
			{UUID: fake.UUID().V4(), UserUUID: userUUID, Tags: []string{"golang", "rss"}},
			{UUID: fake.UUID().V4(), UserUUID: userUUID, Tags: []string{"golang"}},
			{UUID: fake.UUID().V4(), UserUUID: otherUserUUID, Tags: []string{"golang"}},
		},
	}

	testService := NewService(&testRepository)

	cases := []struct {
		tname      string
		userUUID   string
		wantCounts map[string]uint
	}{
		// nominal cases
		{
			tname:      "user has tagged subscriptions",
			userUUID:   userUUID,
			wantCounts: map[string]uint{"golang": 2, "rss": 1},
		},

		// edge cases
		{
			tname:      "user has no tagged subscriptions",
			userUUID:   fake.UUID().V4(),
			wantCounts: map[string]uint{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			got, err := testService.SubscriptionCountsByTag(t.Context(), tc.userUUID)
			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			if len(got) != len(tc.wantCounts) {
				t.Fatalf("want %d tag counts, got %d: %v", len(tc.wantCounts), len(got), got)
			}
			for name, wantCount := range tc.wantCounts {
				if got[name] != wantCount {
					t.Errorf("want count %d for tag %q, got %d", wantCount, name, got[name])
				}
			}
		})
	}
}

func TestServiceFeedsByPageFiltersByShowEntries(t *testing.T) {
	fake := faker.New()

	f := feed.Feed{
		UUID:  fake.UUID().V4(),
		Title: "Local Test",
		Slug:  "local-test",
	}

	unreadEntry := feed.Entry{UID: "unread", FeedUUID: f.UUID, URL: "http://test.local/1", Title: "Unread"}
	readEntry := feed.Entry{UID: "read", FeedUUID: f.UUID, URL: "http://test.local/2", Title: "Read"}

	userUUID := fake.UUID().V4()

	category := feed.Category{UUID: fake.UUID().V4(), UserUUID: userUUID, Name: "Category", Slug: "category"}
	subscription := feed.Subscription{UUID: fake.UUID().V4(), CategoryUUID: category.UUID, FeedUUID: f.UUID, UserUUID: userUUID}
	readMetadata := feed.EntryMetadata{UserUUID: userUUID, EntryUID: readEntry.UID, Read: true}

	testRepository := FakeRepository{
		Categories:      []feed.Category{category},
		Entries:         []feed.Entry{unreadEntry, readEntry},
		EntriesMetadata: []feed.EntryMetadata{readMetadata},
		Feeds:           []feed.Feed{f},
		Subscriptions:   []feed.Subscription{subscription},
	}

	testService := NewService(&testRepository)

	cases := []struct {
		tname       string
		showEntries feed.EntryVisibility
		wantUIDs    []string
	}{
		{tname: "all", showEntries: feed.EntryVisibilityAll, wantUIDs: []string{unreadEntry.UID, readEntry.UID}},
		{tname: "read only", showEntries: feed.EntryVisibilityRead, wantUIDs: []string{readEntry.UID}},
		{tname: "unread only", showEntries: feed.EntryVisibilityUnread, wantUIDs: []string{unreadEntry.UID}},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			preferences := feed.Preferences{ShowEntries: tc.showEntries}

			got, err := testService.FeedsByPage(t.Context(), userUUID, preferences, 1)
			if err != nil {
				t.Fatalf("want no error, got %q", err)
			}

			var gotUIDs []string
			for _, entry := range got.Entries {
				gotUIDs = append(gotUIDs, entry.UID)
			}

			if len(gotUIDs) != len(tc.wantUIDs) {
				t.Fatalf("want entries %v, got %v", tc.wantUIDs, gotUIDs)
			}
			for i, want := range tc.wantUIDs {
				if gotUIDs[i] != want {
					t.Errorf("want entries %v, got %v", tc.wantUIDs, gotUIDs)
					break
				}
			}
		})
	}
}
