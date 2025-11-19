// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package pgfeed_test

import (
	"testing"
	"time"

	"github.com/jaswdr/faker/v2"

	"github.com/virtualtam/sparklemuffin/internal/paginate"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgbase"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgfeed"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pguser"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
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

	if err := us.Add(t.Context(), u); err != nil {
		t.Fatalf("failed to create user: %q", err)
	}

	testUser, err := us.ByNickName(t.Context(), u.NickName)
	if err != nil {
		t.Fatalf("failed to retrieve user: %q", err)
	}

	preferences, err := r.FeedPreferencesGetByUserUUID(t.Context(), testUser.UUID)
	if err != nil {
		t.Fatalf("failed to retrieve preferences: %q", err)
	}

	preferencesRead := feed.Preferences{
		UserUUID:    preferences.UserUUID,
		ShowEntries: feed.EntryVisibilityRead,
	}

	preferencesUnread := feed.Preferences{
		UserUUID:    preferences.UserUUID,
		ShowEntries: feed.EntryVisibilityUnread,
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

	t.Run("FeedsByPage - All", func(t *testing.T) {
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

		gotPage, err := qs.FeedsByPage(t.Context(), testUser.UUID, preferences, 1)
		if err != nil {
			t.Fatalf("failed to retrieve feeds by page: %q", err)
		}

		querying.AssertPageEquals(t, gotPage, wantPage)
	})

	t.Run("FeedsByPage - Read", func(t *testing.T) {
		wantPage := querying.FeedPage{
			Page: paginate.Page{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     1,
				TotalPages:         1,
				ItemOffset:         1,
				ItemCount:          2,
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
			},
		}

		gotPage, err := qs.FeedsByPage(t.Context(), testUser.UUID, preferencesRead, 1)
		if err != nil {
			t.Fatalf("failed to retrieve feeds by page: %q", err)
		}

		querying.AssertPageEquals(t, gotPage, wantPage)
	})

	t.Run("FeedsByPage - Unread", func(t *testing.T) {
		wantPage := querying.FeedPage{
			Page: paginate.Page{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     1,
				TotalPages:         1,
				ItemOffset:         1,
				ItemCount:          3,
			},

			PageTitle:   querying.PageHeaderAll,
			Description: "",
			Unread:      3,
			Categories:  wantCategories,
			Entries: []querying.SubscribedFeedEntry{
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

		gotPage, err := qs.FeedsByPage(t.Context(), testUser.UUID, preferencesUnread, 1)
		if err != nil {
			t.Fatalf("failed to retrieve feeds by page: %q", err)
		}

		querying.AssertPageEquals(t, gotPage, wantPage)
	})

	t.Run("FeedsByCategoryAndPage - All", func(t *testing.T) {
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

		gotPage, err := qs.FeedsByCategoryAndPage(t.Context(), testUser.UUID, preferences, fakeData.categories[0], 1)
		if err != nil {
			t.Fatalf("failed to retrieve feeds by category and page: %q", err)
		}

		querying.AssertPageEquals(t, gotPage, wantPage)
	})

	t.Run("FeedsByCategoryAndPage - Read", func(t *testing.T) {
		wantPage := querying.FeedPage{
			Page: paginate.Page{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     1,
				TotalPages:         1,
				ItemOffset:         1,
				ItemCount:          2,
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
			},
		}

		gotPage, err := qs.FeedsByCategoryAndPage(t.Context(), testUser.UUID, preferencesRead, fakeData.categories[0], 1)
		if err != nil {
			t.Fatalf("failed to retrieve feeds by category and page: %q", err)
		}

		querying.AssertPageEquals(t, gotPage, wantPage)
	})

	t.Run("FeedsByCategoryAndPage - Unread", func(t *testing.T) {
		wantPage := querying.FeedPage{
			Page: paginate.Page{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     1,
				TotalPages:         1,
				ItemOffset:         1,
				ItemCount:          3,
			},

			PageTitle:   fakeData.categories[0].Name,
			Description: "",
			Unread:      3,
			Categories:  wantCategories,
			Entries: []querying.SubscribedFeedEntry{
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

		gotPage, err := qs.FeedsByCategoryAndPage(t.Context(), testUser.UUID, preferencesUnread, fakeData.categories[0], 1)
		if err != nil {
			t.Fatalf("failed to retrieve feeds by category and page: %q", err)
		}

		querying.AssertPageEquals(t, gotPage, wantPage)
	})

	t.Run("FeedsBySubscriptionAndPage - All", func(t *testing.T) {
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

		gotPage, err := qs.FeedsBySubscriptionAndPage(t.Context(), testUser.UUID, preferences, fakeData.subscriptions[1], 1)
		if err != nil {
			t.Fatalf("failed to retrieve feeds by subscription and page: %q", err)
		}

		querying.AssertPageEquals(t, gotPage, wantPage)
	})

	t.Run("FeedsBySubscriptionAndPage - Read", func(t *testing.T) {
		wantPage := querying.FeedPage{
			Page: paginate.Page{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     1,
				TotalPages:         1,
				ItemOffset:         1,
				ItemCount:          1,
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
			},
		}

		gotPage, err := qs.FeedsBySubscriptionAndPage(t.Context(), testUser.UUID, preferencesRead, fakeData.subscriptions[1], 1)
		if err != nil {
			t.Fatalf("failed to retrieve feeds by subscription and page: %q", err)
		}

		querying.AssertPageEquals(t, gotPage, wantPage)
	})

	t.Run("FeedsBySubscriptionAndPage - Unread", func(t *testing.T) {
		wantPage := querying.FeedPage{
			Page: paginate.Page{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     1,
				TotalPages:         1,
				ItemOffset:         1,
				ItemCount:          1,
			},

			PageTitle:   fakeData.feeds[1].Title,
			Description: fakeData.feeds[1].Description,
			Unread:      3,
			Categories:  wantCategories,
			Entries: []querying.SubscribedFeedEntry{
				{
					Entry:     fakeData.entries[4],
					FeedTitle: fakeData.feeds[1].Title,
				},
			},
		}

		gotPage, err := qs.FeedsBySubscriptionAndPage(t.Context(), testUser.UUID, preferencesUnread, fakeData.subscriptions[1], 1)
		if err != nil {
			t.Fatalf("failed to retrieve feeds by subscription and page: %q", err)
		}

		querying.AssertPageEquals(t, gotPage, wantPage)
	})

	t.Run("FeedsByQueryAndPage - All", func(t *testing.T) {
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

		gotPage, err := qs.FeedsByQueryAndPage(t.Context(), testUser.UUID, preferences, "authentic production", 1)
		if err != nil {
			t.Fatalf("failed to retrieve feeds by query and page: %q", err)
		}

		querying.AssertPageEquals(t, gotPage, wantPage)
	})

	t.Run("FeedsByQueryAndPage - Read", func(t *testing.T) {
		wantPage := querying.FeedPage{
			Page: paginate.Page{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     1,
				TotalPages:         1,
				ItemOffset:         1,
				ItemCount:          1,
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
			},
		}

		gotPage, err := qs.FeedsByQueryAndPage(t.Context(), testUser.UUID, preferencesRead, "authentic production", 1)
		if err != nil {
			t.Fatalf("failed to retrieve feeds by query and page: %q", err)
		}

		querying.AssertPageEquals(t, gotPage, wantPage)
	})

	t.Run("FeedsByQueryAndPage - Unread", func(t *testing.T) {
		wantPage := querying.FeedPage{
			Page: paginate.Page{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     1,
				TotalPages:         1,
				ItemOffset:         1,
				ItemCount:          1,
				SearchTerms:        "authentic production",
			},

			PageTitle:  querying.PageHeaderAll,
			Unread:     3,
			Categories: wantCategories,
			Entries: []querying.SubscribedFeedEntry{
				{
					Entry:     fakeData.entries[4],
					FeedTitle: fakeData.feeds[1].Title,
				},
			},
		}

		gotPage, err := qs.FeedsByQueryAndPage(t.Context(), testUser.UUID, preferencesUnread, "authentic production", 1)
		if err != nil {
			t.Fatalf("failed to retrieve feeds by query and page: %q", err)
		}

		querying.AssertPageEquals(t, gotPage, wantPage)
	})

	t.Run("FeedsByCategoryAndQueryAndPage - All", func(t *testing.T) {
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

		gotPage, err := qs.FeedsByCategoryAndQueryAndPage(t.Context(), testUser.UUID, preferences, fakeData.categories[0], "authentic production", 1)
		if err != nil {
			t.Fatalf("failed to retrieve feeds by category and query and page: %q", err)
		}

		querying.AssertPageEquals(t, gotPage, wantPage)
	})

	t.Run("FeedsByCategoryAndQueryAndPage - Read", func(t *testing.T) {
		wantPage := querying.FeedPage{
			Page: paginate.Page{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     1,
				TotalPages:         1,
				ItemOffset:         1,
				ItemCount:          1,
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
			},
		}

		gotPage, err := qs.FeedsByCategoryAndQueryAndPage(t.Context(), testUser.UUID, preferencesRead, fakeData.categories[0], "authentic production", 1)
		if err != nil {
			t.Fatalf("failed to retrieve feeds by category and query and page: %q", err)
		}

		querying.AssertPageEquals(t, gotPage, wantPage)
	})

	t.Run("FeedsByCategoryAndQueryAndPage - Unread", func(t *testing.T) {
		wantPage := querying.FeedPage{
			Page: paginate.Page{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     1,
				TotalPages:         1,
				ItemOffset:         1,
				ItemCount:          1,
				SearchTerms:        "authentic production",
			},

			PageTitle:   fakeData.categories[0].Name,
			Description: "",
			Unread:      3,
			Categories:  wantCategories,
			Entries: []querying.SubscribedFeedEntry{
				{
					Entry:     fakeData.entries[4],
					FeedTitle: fakeData.feeds[1].Title,
				},
			},
		}

		gotPage, err := qs.FeedsByCategoryAndQueryAndPage(t.Context(), testUser.UUID, preferencesUnread, fakeData.categories[0], "authentic production", 1)
		if err != nil {
			t.Fatalf("failed to retrieve feeds by category and query and page: %q", err)
		}

		querying.AssertPageEquals(t, gotPage, wantPage)
	})

	t.Run("FeedsBySubscriptionAndQueryAndPage - All", func(t *testing.T) {
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

		gotPage, err := qs.FeedsBySubscriptionAndQueryAndPage(t.Context(), testUser.UUID, preferences, fakeData.subscriptions[1], "authentic production", 1)
		if err != nil {
			t.Fatalf("failed to retrieve feeds by subscription and query and page: %q", err)
		}

		querying.AssertPageEquals(t, gotPage, wantPage)
	})

	t.Run("FeedsBySubscriptionAndQueryAndPage - Read", func(t *testing.T) {
		wantPage := querying.FeedPage{
			Page: paginate.Page{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     1,
				TotalPages:         1,
				ItemOffset:         1,
				ItemCount:          1,
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
			},
		}

		gotPage, err := qs.FeedsBySubscriptionAndQueryAndPage(t.Context(), testUser.UUID, preferencesRead, fakeData.subscriptions[1], "authentic production", 1)
		if err != nil {
			t.Fatalf("failed to retrieve feeds by subscription and query and page: %q", err)
		}

		querying.AssertPageEquals(t, gotPage, wantPage)
	})

	t.Run("FeedsBySubscriptionAndQueryAndPage - Unread", func(t *testing.T) {
		wantPage := querying.FeedPage{
			Page: paginate.Page{
				PageNumber:         1,
				PreviousPageNumber: 1,
				NextPageNumber:     1,
				TotalPages:         1,
				ItemOffset:         1,
				ItemCount:          1,
				SearchTerms:        "authentic production",
			},

			PageTitle:   fakeData.feeds[1].Title,
			Description: fakeData.feeds[1].Description,
			Unread:      3,
			Categories:  wantCategories,
			Entries: []querying.SubscribedFeedEntry{
				{
					Entry:     fakeData.entries[4],
					FeedTitle: fakeData.feeds[1].Title,
				},
			},
		}

		gotPage, err := qs.FeedsBySubscriptionAndQueryAndPage(t.Context(), testUser.UUID, preferencesUnread, fakeData.subscriptions[1], "authentic production", 1)
		if err != nil {
			t.Fatalf("failed to retrieve feeds by subscription and query and page: %q", err)
		}

		querying.AssertPageEquals(t, gotPage, wantPage)
	})
}
