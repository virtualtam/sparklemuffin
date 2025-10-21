// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package pgfeed_test

import (
	"testing"
	"time"

	"github.com/jaswdr/faker"

	"github.com/virtualtam/sparklemuffin/internal/paginate"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgbase"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgfeed"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pguser"
	"github.com/virtualtam/sparklemuffin/pkg/feed/querying"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

func TestFeedQueryingService(t *testing.T) {
	pool := pgbase.CreateAndMigrateTestDatabase(t)

	r := pgfeed.NewRepository(pool)
	qs := querying.NewService(r)

	ur := pguser.NewRepository(pool)
	us := user.NewService(ur)

	fake := faker.New()

	u := pgbase.GenerateFakeUser(t, &fake)

	if err := us.Add(u); err != nil {
		t.Fatalf("failed to create user: %q", err)
	}

	testUser, err := us.ByNickName(u.NickName)
	if err != nil {
		t.Fatalf("failed to retrieve user: %q", err)
	}

	now := time.Now().UTC()
	fakeData := generateFakeData(t, &fake, now, testUser)
	fakeData.insert(t, r)

	wantCategories := []querying.SubscribedFeedsByCategory{
		{
			Category: fakeData.categories[0],
			Unread:   3,
			SubscribedFeeds: []querying.SubscribedFeed{
				{
					Feed:   fakeData.feeds[0],
					Unread: 2,
				},
				{
					Feed:   fakeData.feeds[1],
					Unread: 1,
				},
			},
		},
	}

	t.Run("FeedsByPage", func(t *testing.T) {
		wantPage := querying.FeedPage{
			Page: paginate.Page{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     1,
				TotalPages:         1,
				ItemOffset:         1,
				ItemCount:          5,
			},

			PageTitle:   querying.PageHeaderAll,
			Description: "",
			Unread:      3,
			Categories:  wantCategories,
			Entries: []querying.SubscribedFeedEntry{
				{
					Entry:     fakeData.entries[0],
					FeedTitle: fakeData.feeds[0].Title,
					Read:      true,
				},
				{
					Entry:     fakeData.entries[3],
					FeedTitle: fakeData.feeds[1].Title,
					Read:      true,
				},
				{
					Entry:     fakeData.entries[1],
					FeedTitle: fakeData.feeds[0].Title,
				},
				{
					Entry:     fakeData.entries[4],
					FeedTitle: fakeData.feeds[1].Title,
				},
				{
					Entry:     fakeData.entries[2],
					FeedTitle: fakeData.feeds[0].Title,
				},
			},
		}

		gotPage, err := qs.FeedsByPage(testUser.UUID, 1)
		if err != nil {
			t.Fatalf("failed to retrieve feeds by page: %q", err)
		}

		querying.AssertPageEquals(t, gotPage, wantPage)
	})

	t.Run("FeedsByCategoryAndPage", func(t *testing.T) {
		wantPage := querying.FeedPage{
			Page: paginate.Page{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     1,
				TotalPages:         1,
				ItemOffset:         1,
				ItemCount:          5,
			},

			PageTitle:   fakeData.categories[0].Name,
			Description: "",
			Unread:      3,
			Categories:  wantCategories,
			Entries: []querying.SubscribedFeedEntry{
				{
					Entry:     fakeData.entries[0],
					FeedTitle: fakeData.feeds[0].Title,
					Read:      true,
				},
				{
					Entry:     fakeData.entries[3],
					FeedTitle: fakeData.feeds[1].Title,
					Read:      true,
				},
				{
					Entry:     fakeData.entries[1],
					FeedTitle: fakeData.feeds[0].Title,
				},
				{
					Entry:     fakeData.entries[4],
					FeedTitle: fakeData.feeds[1].Title,
				},
				{
					Entry:     fakeData.entries[2],
					FeedTitle: fakeData.feeds[0].Title,
				},
			},
		}

		gotPage, err := qs.FeedsByCategoryAndPage(testUser.UUID, fakeData.categories[0], 1)
		if err != nil {
			t.Fatalf("failed to retrieve feeds by category and page: %q", err)
		}

		querying.AssertPageEquals(t, gotPage, wantPage)
	})

	t.Run("FeedsBySubscriptionAndPage", func(t *testing.T) {
		wantPage := querying.FeedPage{
			Page: paginate.Page{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     1,
				TotalPages:         1,
				ItemOffset:         1,
				ItemCount:          2,
			},

			PageTitle:   fakeData.feeds[1].Title,
			Description: fakeData.feeds[1].Description,
			Unread:      3,
			Categories:  wantCategories,
			Entries: []querying.SubscribedFeedEntry{
				{
					Entry:     fakeData.entries[3],
					FeedTitle: fakeData.feeds[1].Title,
					Read:      true,
				},
				{
					Entry:     fakeData.entries[4],
					FeedTitle: fakeData.feeds[1].Title,
				},
			},
		}

		gotPage, err := qs.FeedsBySubscriptionAndPage(testUser.UUID, fakeData.subscriptions[1], 1)
		if err != nil {
			t.Fatalf("failed to retrieve feeds by subscription and page: %q", err)
		}

		querying.AssertPageEquals(t, gotPage, wantPage)
	})

	t.Run("FeedsByQueryAndPage", func(t *testing.T) {
		wantPage := querying.FeedPage{
			Page: paginate.Page{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     1,
				TotalPages:         1,
				ItemOffset:         1,
				ItemCount:          2,
				SearchTerms:        "authentic production",
			},

			PageTitle:  querying.PageHeaderAll,
			Unread:     3,
			Categories: wantCategories,
			Entries: []querying.SubscribedFeedEntry{
				{
					Entry:     fakeData.entries[3],
					FeedTitle: fakeData.feeds[1].Title,
					Read:      true,
				},
				{
					Entry:     fakeData.entries[4],
					FeedTitle: fakeData.feeds[1].Title,
				},
			},
		}

		gotPage, err := qs.FeedsByQueryAndPage(testUser.UUID, "authentic production", 1)
		if err != nil {
			t.Fatalf("failed to retrieve feeds by query and page: %q", err)
		}

		querying.AssertPageEquals(t, gotPage, wantPage)
	})

	t.Run("FeedsByCategoryAndQueryAndPage", func(t *testing.T) {
		wantPage := querying.FeedPage{
			Page: paginate.Page{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     1,
				TotalPages:         1,
				ItemOffset:         1,
				ItemCount:          2,
				SearchTerms:        "authentic production",
			},

			PageTitle:   fakeData.categories[0].Name,
			Description: "",
			Unread:      3,
			Categories:  wantCategories,
			Entries: []querying.SubscribedFeedEntry{
				{
					Entry:     fakeData.entries[3],
					FeedTitle: fakeData.feeds[1].Title,
					Read:      true,
				},
				{
					Entry:     fakeData.entries[4],
					FeedTitle: fakeData.feeds[1].Title,
				},
			},
		}

		gotPage, err := qs.FeedsByCategoryAndQueryAndPage(testUser.UUID, fakeData.categories[0], "authentic production", 1)
		if err != nil {
			t.Fatalf("failed to retrieve feeds by category and query and page: %q", err)
		}

		querying.AssertPageEquals(t, gotPage, wantPage)
	})

	t.Run("FeedsBySubscriptionAndQueryAndPage", func(t *testing.T) {
		wantPage := querying.FeedPage{
			Page: paginate.Page{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     1,
				TotalPages:         1,
				ItemOffset:         1,
				ItemCount:          2,
				SearchTerms:        "authentic production",
			},

			PageTitle:   fakeData.feeds[1].Title,
			Description: fakeData.feeds[1].Description,
			Unread:      3,
			Categories:  wantCategories,
			Entries: []querying.SubscribedFeedEntry{
				{
					Entry:     fakeData.entries[3],
					FeedTitle: fakeData.feeds[1].Title,
					Read:      true,
				},
				{
					Entry:     fakeData.entries[4],
					FeedTitle: fakeData.feeds[1].Title,
				},
			},
		}

		gotPage, err := qs.FeedsBySubscriptionAndQueryAndPage(testUser.UUID, fakeData.subscriptions[1], "authentic production", 1)
		if err != nil {
			t.Fatalf("failed to retrieve feeds by subscription and query and page: %q", err)
		}

		querying.AssertPageEquals(t, gotPage, wantPage)
	})
}
