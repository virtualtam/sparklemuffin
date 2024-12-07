// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package pgfeed_test

import (
	"context"
	"testing"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/jaswdr/faker"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgbase"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgfeed"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pguser"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	"github.com/virtualtam/sparklemuffin/pkg/feed/querying"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

func TestFeedQueryingService(t *testing.T) {
	ctx := context.Background()
	pool := pgbase.CreateAndMigrateTestDatabase(t, ctx)

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
			PageNumber:         1,
			PreviousPageNumber: 1,
			NextPageNumber:     1,
			TotalPages:         1,
			Offset:             1,
			Header:             querying.PageHeaderAll,
			Description:        "",
			Unread:             3,
			Categories:         wantCategories,
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
			PageNumber:         1,
			PreviousPageNumber: 1,
			NextPageNumber:     1,
			TotalPages:         1,
			Offset:             1,
			Header:             fakeData.categories[0].Name,
			Description:        "",
			Unread:             3,
			Categories:         wantCategories,
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
			PageNumber:         1,
			PreviousPageNumber: 1,
			NextPageNumber:     1,
			TotalPages:         1,
			Offset:             1,
			Header:             fakeData.feeds[1].Title,
			Description:        fakeData.feeds[1].Description,
			Unread:             3,
			Categories:         wantCategories,
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
			PageNumber:         1,
			PreviousPageNumber: 1,
			NextPageNumber:     1,
			TotalPages:         1,
			Offset:             1,
			Header:             querying.PageHeaderAll,
			Description:        "",
			Unread:             3,
			Categories:         wantCategories,
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
			PageNumber:         1,
			PreviousPageNumber: 1,
			NextPageNumber:     1,
			TotalPages:         1,
			Offset:             1,
			Header:             fakeData.categories[0].Name,
			Description:        "",
			Unread:             3,
			Categories:         wantCategories,
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
			PageNumber:         1,
			PreviousPageNumber: 1,
			NextPageNumber:     1,
			TotalPages:         1,
			Offset:             1,
			Header:             fakeData.feeds[1].Title,
			Description:        fakeData.feeds[1].Description,
			Unread:             3,
			Categories:         wantCategories,
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

type fakeData struct {
	feeds   []feed.Feed
	entries []feed.Entry

	categories      []feed.Category
	subscriptions   []feed.Subscription
	entriesMetadata []feed.EntryMetadata
}

func (fd *fakeData) insert(t *testing.T, r *pgfeed.Repository) {
	t.Helper()

	for _, feed := range fd.feeds {
		if err := r.FeedCreate(feed); err != nil {
			t.Fatalf("failed to create feed: %q", err)
		}
	}

	if _, err := r.FeedEntryCreateMany(fd.entries); err != nil {
		t.Fatalf("failed to create entries: %q", err)
	}

	for _, category := range fd.categories {
		if err := r.FeedCategoryCreate(category); err != nil {
			t.Fatalf("failed to create category: %q", err)
		}
	}

	for _, subscription := range fd.subscriptions {
		if _, err := r.FeedSubscriptionCreate(subscription); err != nil {
			t.Fatalf("failed to create subscription: %q", err)
		}
	}

	for _, entryMetadata := range fd.entriesMetadata {
		if err := r.FeedEntryMetadataCreate(entryMetadata); err != nil {
			t.Fatalf("failed to create entry metadata: %q", err)
		}
	}
}

func generateFakeData(t *testing.T, fake *faker.Faker, now time.Time, testUser user.User) fakeData {
	t.Helper()

	feed1 := generateFakeFeed(t, fake, "Local Test", "A fake feed for local testing", now)
	feed1Entries := generateFakeEntries(t, fake, now, feed1.UUID, 3)

	feed2now := now.Add(-30 * time.Minute).UTC()
	feed2 := generateFakeFeed(t, fake, "Production Feed", "This is an authentic production feed", feed2now)
	feed2Entries := generateFakeEntries(t, fake, feed2now, feed2.UUID, 2)

	category1 := generateFakeCategory(t, fake, testUser.UUID, "Run Environments")

	subscriptions := []feed.Subscription{
		{
			UUID:         fake.UUID().V4(),
			FeedUUID:     feed1.UUID,
			CategoryUUID: category1.UUID,
			UserUUID:     testUser.UUID,
		},
		{
			UUID:         fake.UUID().V4(),
			FeedUUID:     feed2.UUID,
			CategoryUUID: category1.UUID,
			UserUUID:     testUser.UUID,
		},
	}

	entriesMetadata := []feed.EntryMetadata{
		{
			UserUUID: testUser.UUID,
			EntryUID: feed1Entries[0].UID,
			Read:     true,
		},
		{
			UserUUID: testUser.UUID,
			EntryUID: feed2Entries[0].UID,
			Read:     true,
		},
	}

	data := fakeData{
		feeds:   []feed.Feed{feed1, feed2},
		entries: append(feed1Entries, feed2Entries...),
		categories: []feed.Category{
			category1,
		},
		subscriptions:   subscriptions,
		entriesMetadata: entriesMetadata,
	}

	return data
}

func generateFakeFeed(t *testing.T, fake *faker.Faker, title, description string, createdAt time.Time) feed.Feed {
	t.Helper()

	return feed.Feed{
		UUID:        fake.UUID().V4(),
		FeedURL:     fake.Internet().URL(),
		Title:       title,
		Description: description,
		Slug:        fake.Internet().Slug(),
		Hash:        xxhash.Sum64String(fake.Lorem().Text(100)),
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
		FetchedAt:   createdAt,
	}
}

func generateFakeEntries(t *testing.T, fake *faker.Faker, now time.Time, feedUUID string, n int) []feed.Entry {
	t.Helper()

	entries := make([]feed.Entry, n)

	for i := 0; i < n; i++ {
		publishedAt := now.Add(-time.Duration(i) * 12 * time.Hour)

		entry := feed.Entry{
			UID:         fake.UUID().V4(),
			FeedUUID:    feedUUID,
			URL:         fake.Internet().URL(),
			Title:       fake.Lorem().Text(10),
			Summary:     fake.Lorem().Text(100),
			PublishedAt: publishedAt,
			UpdatedAt:   publishedAt,
		}

		entries[i] = entry
	}

	return entries
}

func generateFakeCategory(t *testing.T, fake *faker.Faker, userUUID, name string) feed.Category {
	t.Helper()

	return feed.Category{
		UUID:     fake.UUID().V4(),
		UserUUID: userUUID,
		Name:     name,
		Slug:     fake.Internet().Slug(),
	}
}
