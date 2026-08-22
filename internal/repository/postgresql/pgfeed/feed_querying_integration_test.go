// Copyright VirtualTam 2022, 2026
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

	u := user.FakeUser(t, &fake)

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
					FeedSlug:  fakeData.feeds[0].Slug,
					FeedTitle: fakeData.feeds[0].Title,
					Read:      true,
				},
				{
					Entry:     fakeData.entries[3],
					FeedSlug:  fakeData.feeds[1].Slug,
					FeedTitle: fakeData.feeds[1].Title,
					Read:      true,
				},
				{
					Entry:     fakeData.entries[1],
					FeedSlug:  fakeData.feeds[0].Slug,
					FeedTitle: fakeData.feeds[0].Title,
				},
				{
					Entry:     fakeData.entries[4],
					FeedSlug:  fakeData.feeds[1].Slug,
					FeedTitle: fakeData.feeds[1].Title,
				},
				{
					Entry:     fakeData.entries[2],
					FeedSlug:  fakeData.feeds[0].Slug,
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
					FeedSlug:  fakeData.feeds[0].Slug,
					FeedTitle: fakeData.feeds[0].Title,
					Read:      true,
				},
				{
					Entry:     fakeData.entries[3],
					FeedSlug:  fakeData.feeds[1].Slug,
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
					FeedSlug:  fakeData.feeds[0].Slug,
					FeedTitle: fakeData.feeds[0].Title,
				},
				{
					Entry:     fakeData.entries[4],
					FeedSlug:  fakeData.feeds[1].Slug,
					FeedTitle: fakeData.feeds[1].Title,
				},
				{
					Entry:     fakeData.entries[2],
					FeedSlug:  fakeData.feeds[0].Slug,
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
					FeedSlug:  fakeData.feeds[0].Slug,
					FeedTitle: fakeData.feeds[0].Title,
					Read:      true,
				},
				{
					Entry:     fakeData.entries[3],
					FeedSlug:  fakeData.feeds[1].Slug,
					FeedTitle: fakeData.feeds[1].Title,
					Read:      true,
				},
				{
					Entry:     fakeData.entries[1],
					FeedSlug:  fakeData.feeds[0].Slug,
					FeedTitle: fakeData.feeds[0].Title,
				},
				{
					Entry:     fakeData.entries[4],
					FeedSlug:  fakeData.feeds[1].Slug,
					FeedTitle: fakeData.feeds[1].Title,
				},
				{
					Entry:     fakeData.entries[2],
					FeedSlug:  fakeData.feeds[0].Slug,
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
					FeedSlug:  fakeData.feeds[0].Slug,
					FeedTitle: fakeData.feeds[0].Title,
					Read:      true,
				},
				{
					Entry:     fakeData.entries[3],
					FeedSlug:  fakeData.feeds[1].Slug,
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
					FeedSlug:  fakeData.feeds[0].Slug,
					FeedTitle: fakeData.feeds[0].Title,
				},
				{
					Entry:     fakeData.entries[4],
					FeedSlug:  fakeData.feeds[1].Slug,
					FeedTitle: fakeData.feeds[1].Title,
				},
				{
					Entry:     fakeData.entries[2],
					FeedSlug:  fakeData.feeds[0].Slug,
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
					FeedSlug:  fakeData.feeds[1].Slug,
					FeedTitle: fakeData.feeds[1].Title,
					Read:      true,
				},
				{
					Entry:     fakeData.entries[4],
					FeedSlug:  fakeData.feeds[1].Slug,
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
					FeedSlug:  fakeData.feeds[1].Slug,
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
					FeedSlug:  fakeData.feeds[1].Slug,
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
					FeedSlug:  fakeData.feeds[1].Slug,
					FeedTitle: fakeData.feeds[1].Title,
					Read:      true,
				},
				{
					Entry:     fakeData.entries[4],
					FeedSlug:  fakeData.feeds[1].Slug,
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
					FeedSlug:  fakeData.feeds[1].Slug,
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
					FeedSlug:  fakeData.feeds[1].Slug,
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
					FeedSlug:  fakeData.feeds[1].Slug,
					FeedTitle: fakeData.feeds[1].Title,
					Read:      true,
				},
				{
					Entry:     fakeData.entries[4],
					FeedSlug:  fakeData.feeds[1].Slug,
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
					FeedSlug:  fakeData.feeds[1].Slug,
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
					FeedSlug:  fakeData.feeds[1].Slug,
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
					FeedSlug:  fakeData.feeds[1].Slug,
					FeedTitle: fakeData.feeds[1].Title,
					Read:      true,
				},
				{
					Entry:     fakeData.entries[4],
					FeedSlug:  fakeData.feeds[1].Slug,
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
					FeedSlug:  fakeData.feeds[1].Slug,
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
					FeedSlug:  fakeData.feeds[1].Slug,
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

// TestFeedQueryingServiceSearchMatchesSubscriptionTag verifies that entry
// search matches on a subscription's tags, not just on entry/feed text: the
// search term below ("golang") appears nowhere in the feed's title,
// description, or entries, only on the tag of the subscription itself.
func TestFeedQueryingServiceSearchMatchesSubscriptionTag(t *testing.T) {
	pool := pgbase.CreateAndMigrateTestDatabase(t)

	r := pgfeed.NewRepository(pool)
	qs := querying.NewService(r)

	ur := pguser.NewRepository(pool)
	us := user.NewService(ur)

	fake := faker.New()

	u := user.FakeUser(t, &fake)
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

	now := time.Now().UTC()

	taggedFeed := generateFakeFeed(t, &fake, "Local Test", "A fake feed for local testing", now)
	if err := r.FeedCreate(t.Context(), taggedFeed); err != nil {
		t.Fatalf("failed to create feed: %q", err)
	}

	taggedFeedEntries := generateFakeEntries(t, &fake, now, taggedFeed.UUID, 1)
	if _, err := r.FeedEntryCreateMany(t.Context(), taggedFeedEntries); err != nil {
		t.Fatalf("failed to create entries: %q", err)
	}

	category := generateFakeCategory(t, &fake, testUser.UUID, "Run Environments")
	if err := r.FeedCategoryCreate(t.Context(), category); err != nil {
		t.Fatalf("failed to create category: %q", err)
	}

	taggedSubscription := feed.Subscription{
		UUID:         fake.UUID().V4(),
		FeedUUID:     taggedFeed.UUID,
		CategoryUUID: category.UUID,
		UserUUID:     testUser.UUID,
		Tags:         []string{"golang"},
	}
	if _, err := r.FeedSubscriptionCreate(t.Context(), taggedSubscription); err != nil {
		t.Fatalf("failed to create subscription: %q", err)
	}

	wantPage := querying.FeedPage{
		Page: paginate.Page{
			PageNumber:         1,
			PreviousPageNumber: 1,
			NextPageNumber:     1,
			TotalPages:         1,
			ItemOffset:         1,
			ItemCount:          1,
			SearchTerms:        "golang",
		},

		PageTitle: querying.PageHeaderAll,
		Unread:    1,
		Categories: []querying.SubscribedFeedsByCategory{
			{
				Category: category,
				Unread:   1,
				SubscribedFeeds: []querying.SubscribedFeed{
					{Feed: taggedFeed, Unread: 1},
				},
			},
		},
		Entries: []querying.SubscribedFeedEntry{
			{
				Entry:            taggedFeedEntries[0],
				FeedSlug:         taggedFeed.Slug,
				FeedTitle:        taggedFeed.Title,
				SubscriptionTags: taggedSubscription.Tags,
			},
		},
	}

	gotPage, err := qs.FeedsByQueryAndPage(t.Context(), testUser.UUID, preferences, "golang", 1)
	if err != nil {
		t.Fatalf("failed to retrieve feeds by query and page: %q", err)
	}

	querying.AssertPageEquals(t, gotPage, wantPage)
}

func TestFeedQueryingServiceSubscriptionCountsByTag(t *testing.T) {
	pool := pgbase.CreateAndMigrateTestDatabase(t)

	r := pgfeed.NewRepository(pool)
	qs := querying.NewService(r)

	ur := pguser.NewRepository(pool)
	us := user.NewService(ur)

	fake := faker.New()

	u := user.FakeUser(t, &fake)
	if err := us.Add(t.Context(), u); err != nil {
		t.Fatalf("failed to create user: %q", err)
	}

	testUser, err := us.ByNickName(t.Context(), u.NickName)
	if err != nil {
		t.Fatalf("failed to retrieve user: %q", err)
	}

	now := time.Now().UTC()
	countTag := "zzzcountbytag"

	category := generateFakeCategory(t, &fake, testUser.UUID, "Count Test")
	if err := r.FeedCategoryCreate(t.Context(), category); err != nil {
		t.Fatalf("failed to create category: %q", err)
	}

	feed1 := generateFakeFeed(t, &fake, "Feed 1", "", now)
	if err := r.FeedCreate(t.Context(), feed1); err != nil {
		t.Fatalf("failed to create feed: %q", err)
	}

	subscription1 := feed.Subscription{
		UUID:         fake.UUID().V4(),
		FeedUUID:     feed1.UUID,
		CategoryUUID: category.UUID,
		UserUUID:     testUser.UUID,
		Tags:         []string{countTag},
	}
	if _, err := r.FeedSubscriptionCreate(t.Context(), subscription1); err != nil {
		t.Fatalf("failed to create subscription: %q", err)
	}

	feed2 := generateFakeFeed(t, &fake, "Feed 2", "", now)
	if err := r.FeedCreate(t.Context(), feed2); err != nil {
		t.Fatalf("failed to create feed: %q", err)
	}

	subscription2 := feed.Subscription{
		UUID:         fake.UUID().V4(),
		FeedUUID:     feed2.UUID,
		CategoryUUID: category.UUID,
		UserUUID:     testUser.UUID,
		Tags:         []string{countTag, "other"},
	}
	if _, err := r.FeedSubscriptionCreate(t.Context(), subscription2); err != nil {
		t.Fatalf("failed to create subscription: %q", err)
	}

	got, err := qs.SubscriptionCountsByTag(t.Context(), testUser.UUID)
	if err != nil {
		t.Fatalf("failed to count subscriptions by tag: %q", err)
	}
	if got[countTag] != 2 {
		t.Errorf("want 2 subscriptions for tag %q, got %d", countTag, got[countTag])
	}
	if got["unknown-tag"] != 0 {
		t.Errorf("want 0 subscriptions for tag %q, got %d", "unknown-tag", got["unknown-tag"])
	}
}
