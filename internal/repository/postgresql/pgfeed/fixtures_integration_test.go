// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package pgfeed_test

import (
	"testing"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/jaswdr/faker"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgfeed"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

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
