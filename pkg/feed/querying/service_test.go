// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import (
	"errors"
	"testing"

	"github.com/jaswdr/faker"

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

	testRepository := fakeRepository{
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
							SubscriptionAlias: user1Subscription1.Alias,
							FeedTitle:         feed1.Title,
							Read:              false,
						},
						{
							Entry:             feed1Entry2,
							SubscriptionAlias: user1Subscription1.Alias,
							FeedTitle:         feed1.Title,
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
				got, err := testService.FeedsByPage(tc.userUUID, tc.pageNumber)

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
							SubscriptionAlias: user1Subscription1.Alias,
							FeedTitle:         feed1.Title,
							Read:              false,
						},
						{
							Entry:             feed1Entry2,
							SubscriptionAlias: user1Subscription1.Alias,
							FeedTitle:         feed1.Title,
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
				got, err := testService.FeedsByCategoryAndPage(tc.userUUID, tc.category, tc.pageNumber)

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
							FeedTitle:         feed1.Title,
							SubscriptionAlias: user1Subscription1.Alias,
							Read:              false,
						},
						{
							Entry:             feed1Entry2,
							SubscriptionAlias: user1Subscription1.Alias,
							FeedTitle:         feed1.Title,
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
				got, err := testService.FeedsBySubscriptionAndPage(tc.userUUID, tc.subscription, tc.pageNumber)

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
}
