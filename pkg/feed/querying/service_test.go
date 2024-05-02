// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import (
	"errors"
	"testing"

	"github.com/virtualtam/sparklemuffin/pkg/feed"
)

func TestServiceByPage(t *testing.T) {
	userUUID := "8fbf7b71-97a0-42e8-8d9c-1a6fac6fa7a3"

	r := &fakeRepository{
		Categories: []feed.Category{
			{
				UUID:     "8f041d0f-8f49-4ffa-99a3-896ea372bfc",
				UserUUID: userUUID,
				Name:     "Test Category",
				Slug:     "test-category",
			},
		},
		Entries: []feed.Entry{
			{
				UID:      "1",
				FeedUUID: "04f7dcbc-7080-4ca9-9000-aeac3f62dfb5",
				URL:      "http://test.local/posts/1",
				Title:    "First Post",
			},
		},
		Feeds: []feed.Feed{
			{
				UUID:  "04f7dcbc-7080-4ca9-9000-aeac3f62dfb5",
				URL:   "http://test.local/feed.atom",
				Title: "Local Test",
				Slug:  "local-test",
			},
		},
		Subscriptions: []feed.Subscription{
			{
				UUID:         "72261134-e4df-4472-87ae-097e6433a438",
				CategoryUUID: "8f041d0f-8f49-4ffa-99a3-896ea372bfc",
				FeedUUID:     "04f7dcbc-7080-4ca9-9000-aeac3f62dfb5",
				UserUUID:     userUUID,
			},
		},
	}

	s := NewService(r)

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
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     1,
				TotalPages:         1,
				Offset:             1,
			},
		},
		{
			tname:      "one page, one category with one subscription",
			userUUID:   userUUID,
			pageNumber: 1,
			want: FeedPage{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     1,
				TotalPages:         1,
				Offset:             1,

				Unread: 1,

				Categories: []SubscriptionCategory{
					{
						Category: feed.Category{
							UUID:     "8f041d0f-8f49-4ffa-99a3-896ea372bfc2",
							UserUUID: userUUID,
							Name:     "Test Category",
							Slug:     "test-category",
						},
						Unread: 1,
						SubscribedFeeds: []SubscribedFeed{
							{
								Feed: feed.Feed{
									UUID:  "04f7dcbc-7080-4ca9-9000-aeac3f62dfb5",
									URL:   "http://test.local/feed.atom",
									Title: "Local Test",
									Slug:  "local-test",
								},
								Unread: 1,
							},
						},
					},
				},
				Entries: []SubscriptionEntry{
					{
						Entry: feed.Entry{
							UID:      "1",
							FeedUUID: "04f7dcbc-7080-4ca9-9000-aeac3f62dfb5",
							URL:      "http://test.local/posts/1",
							Title:    "First Post",
						},
						Read: false,
					},
				},
			},
		},

		// error cases
		{
			tname:      "zeroth page",
			pageNumber: 0,
			userUUID:   userUUID,
			wantErr:    ErrPageNumberOutOfBounds,
		},
		{
			tname:      "page number out of bounds",
			pageNumber: 18,
			userUUID:   userUUID,
			wantErr:    ErrPageNumberOutOfBounds,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			got, err := s.FeedsByPage(tc.userUUID, tc.pageNumber)

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

			assertPagesEqual(t, got, tc.want)
		})
	}
}

func TestServiceByCategoryAndPage(t *testing.T) {
	userUUID := "8fbf7b71-97a0-42e8-8d9c-1a6fac6fa7a3"

	r := &fakeRepository{
		Categories: []feed.Category{
			{
				UUID:     "8f041d0f-8f49-4ffa-99a3-896ea372bfc",
				UserUUID: userUUID,
				Name:     "Test Category",
				Slug:     "test-category",
			},
			{
				UUID:     "13326cd8-98a0-4cba-a8fc-4c28c1ffc462",
				UserUUID: userUUID,
				Name:     "Empty Category",
				Slug:     "empty-category",
			},
		},
		Entries: []feed.Entry{
			{
				UID:      "1",
				FeedUUID: "04f7dcbc-7080-4ca9-9000-aeac3f62dfb5",
				URL:      "http://test.local/posts/1",
				Title:    "First Post",
			},
		},
		Feeds: []feed.Feed{
			{
				UUID:  "04f7dcbc-7080-4ca9-9000-aeac3f62dfb5",
				URL:   "http://test.local/feed.atom",
				Title: "Local Test",
				Slug:  "local-test",
			},
		},
		Subscriptions: []feed.Subscription{
			{
				UUID:         "72261134-e4df-4472-87ae-097e6433a438",
				CategoryUUID: "8f041d0f-8f49-4ffa-99a3-896ea372bfc",
				FeedUUID:     "04f7dcbc-7080-4ca9-9000-aeac3f62dfb5",
				UserUUID:     userUUID,
			},
		},
	}

	wantCategories := []SubscriptionCategory{
		{
			Category: feed.Category{
				UUID:     "13326cd8-98a0-4cba-a8fc-4c28c1ffc462",
				UserUUID: userUUID,
				Name:     "Empty Category",
				Slug:     "empty-category",
			},
		},
		{
			Category: feed.Category{
				UUID:     "8f041d0f-8f49-4ffa-99a3-896ea372bfc2",
				UserUUID: userUUID,
				Name:     "Test Category",
				Slug:     "test-category",
			},
			Unread: 1,
			SubscribedFeeds: []SubscribedFeed{
				{
					Feed: feed.Feed{
						UUID:  "04f7dcbc-7080-4ca9-9000-aeac3f62dfb5",
						URL:   "http://test.local/feed.atom",
						Title: "Local Test",
						Slug:  "local-test",
					},
					Unread: 1,
				},
			},
		},
	}

	s := NewService(r)

	cases := []struct {
		tname        string
		userUUID     string
		categoryUUID string
		pageNumber   uint
		want         FeedPage
		wantErr      error
	}{
		// nominal cases
		{
			tname:        "one page, no subscription",
			userUUID:     userUUID,
			categoryUUID: "13326cd8-98a0-4cba-a8fc-4c28c1ffc462",
			pageNumber:   1,
			want: FeedPage{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     1,
				TotalPages:         1,
				Offset:             1,

				Unread: 1,

				Categories: wantCategories,
			},
		},
		{
			tname:        "one page, one category with one subscription, one empty category",
			userUUID:     userUUID,
			categoryUUID: "8f041d0f-8f49-4ffa-99a3-896ea372bfc",
			pageNumber:   1,
			want: FeedPage{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     1,
				TotalPages:         1,
				Offset:             1,

				Unread: 1,

				Categories: wantCategories,
				Entries: []SubscriptionEntry{
					{
						Entry: feed.Entry{
							UID:      "1",
							FeedUUID: "04f7dcbc-7080-4ca9-9000-aeac3f62dfb5",
							URL:      "http://test.local/posts/1",
							Title:    "First Post",
						},
						Read: false,
					},
				},
			},
		},

		// error cases
		{
			tname:      "zeroth page",
			pageNumber: 0,
			userUUID:   userUUID,
			wantErr:    ErrPageNumberOutOfBounds,
		},
		{
			tname:      "page number out of bounds",
			pageNumber: 18,
			userUUID:   userUUID,
			wantErr:    ErrPageNumberOutOfBounds,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			got, err := s.FeedsByCategoryAndPage(tc.userUUID, tc.categoryUUID, tc.pageNumber)

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

			assertPagesEqual(t, got, tc.want)
		})
	}
}
