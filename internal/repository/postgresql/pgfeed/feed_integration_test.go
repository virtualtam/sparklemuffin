// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package pgfeed_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/jaswdr/faker"

	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgbase"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgfeed"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pguser"
	"github.com/virtualtam/sparklemuffin/internal/test/assert"
	"github.com/virtualtam/sparklemuffin/internal/test/feedtest"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	"github.com/virtualtam/sparklemuffin/pkg/feed/fetching"
	"github.com/virtualtam/sparklemuffin/pkg/feed/querying"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

func TestFeedService(t *testing.T) {
	pool := pgbase.CreateAndMigrateTestDatabase(t)

	now := time.Now().UTC()

	atomFeed := feedtest.GenerateDummyFeed(t, now)
	feedStr, err := atomFeed.ToAtom()
	if err != nil {
		t.Fatalf("failed to encode feed to Atom: %q", err)
	}
	feedETag := feedtest.HashETag(feedStr)
	feedLastModified := now
	feedHash := xxhash.Sum64String(feedStr)

	transport := feedtest.NewRoundTripper(t, atomFeed)

	testHTTPClient := &http.Client{
		Transport: transport,
	}
	feedClient := fetching.NewClient(testHTTPClient, "sparklemuffin/test")

	r := pgfeed.NewRepository(pool)
	fs := feed.NewService(r, feedClient)

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

	t.Run("create, retrieve and delete category", func(t *testing.T) {
		categoryName := "Test Category"

		// 1. Create category
		category, err := fs.CreateCategory(testUser.UUID, categoryName)
		if err != nil {
			t.Fatalf("failed to create category: %q", err)
		}

		gotCategory, err := fs.CategoryByUUID(testUser.UUID, category.UUID)
		if err != nil {
			t.Fatalf("failed to retrieve category: %q", err)
		}

		if gotCategory.Name != categoryName {
			t.Errorf("want Name %q, got %q", categoryName, gotCategory.Name)
		}
		if gotCategory.UserUUID != testUser.UUID {
			t.Errorf("want UserUUID %q, got %q", testUser.UUID, gotCategory.UserUUID)
		}
		if gotCategory.UUID != category.UUID {
			t.Errorf("want UUID %q, got %q", category.UUID, gotCategory.UUID)
		}

		// 2. Teardown
		if err := fs.DeleteCategory(testUser.UUID, category.UUID); err != nil {
			t.Fatalf("failed to delete category: %q", err)
		}

		_, err = fs.CategoryByUUID(testUser.UUID, category.UUID)
		if !errors.Is(err, feed.ErrCategoryNotFound) {
			t.Errorf("want ErrCategoryNotFound, got %q", err)
		} else if err == nil {
			t.Errorf("want ErrCategoryNotFound, got none")
		}
	})

	t.Run("create, update and delete category", func(t *testing.T) {
		categoryName := "Test Category"

		category, err := fs.CreateCategory(testUser.UUID, categoryName)
		if err != nil {
			t.Fatalf("failed to create category: %q", err)
		}

		newCategoryName := "New Test Category"
		newCategory := feed.Category{
			UUID:     category.UUID,
			UserUUID: testUser.UUID,
			Name:     newCategoryName,
		}

		if err := fs.UpdateCategory(newCategory); err != nil {
			t.Fatalf("failed to update category: %q", err)
		}

		gotCategory, err := fs.CategoryByUUID(testUser.UUID, category.UUID)
		if err != nil {
			t.Fatalf("failed to retrieve category: %q", err)
		}

		if gotCategory.Name != newCategoryName {
			t.Errorf("want Name %q, got %q", categoryName, gotCategory.Name)
		}
		if gotCategory.UserUUID != testUser.UUID {
			t.Errorf("want UserUUID %q, got %q", testUser.UUID, gotCategory.UserUUID)
		}
		if gotCategory.UUID != category.UUID {
			t.Errorf("want UUID %q, got %q", category.UUID, gotCategory.UUID)
		}

		if err := fs.DeleteCategory(testUser.UUID, category.UUID); err != nil {
			t.Fatalf("failed to delete category: %q", err)
		}
	})

	t.Run("create, retrieve and delete feed subscription", func(t *testing.T) {
		categoryName := "Subscriptions"

		// 1. Create category
		category, err := fs.CreateCategory(testUser.UUID, categoryName)
		if err != nil {
			t.Fatalf("failed to create category: %q", err)
		}

		gotCategory, err := fs.CategoryByUUID(testUser.UUID, category.UUID)
		if err != nil {
			t.Fatalf("failed to retrieve category: %q", err)
		}

		// 2. Create feed, entries and subscription
		if err := fs.Subscribe(testUser.UUID, category.UUID, "http://test.local"); err != nil {
			t.Fatalf("failed to subscribe to feed: %q", err)
		}

		wantFeed := feed.Feed{
			FeedURL:      "http://test.local",
			Title:        "Local Test",
			Description:  "A simple syndication feed, for testing purposes.",
			Slug:         "local-test",
			ETag:         feedETag,
			LastModified: feedLastModified,
			Hash:         feedHash,
			CreatedAt:    now,
			UpdatedAt:    now,
			FetchedAt:    now,
		}

		gotFeed, err := fs.FeedBySlug(wantFeed.Slug)
		if err != nil {
			t.Fatalf("failed to retrieve feed: %q", err)
		}

		feed.AssertFeedEquals(t, gotFeed, wantFeed)

		gotSubscription, err := fs.SubscriptionByFeed(testUser.UUID, gotFeed.UUID)
		if err != nil {
			t.Fatalf("failed to retrieve subscription: %q", err)
		}

		wantSubscription := feed.Subscription{
			UUID:         gotSubscription.UUID,
			CategoryUUID: gotCategory.UUID,
			FeedUUID:     gotFeed.UUID,
			UserUUID:     testUser.UUID,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		feed.AssertSubscriptionEquals(t, gotSubscription, wantSubscription)

		yesterday := now.Add(-24 * time.Hour)

		wantEntries := []querying.SubscribedFeedEntry{
			{
				Entry: feed.Entry{
					FeedUUID:    gotFeed.UUID,
					URL:         "http://test.local/first-post",
					Title:       "First post!",
					Summary:     "First post!\n\nThis is the first post!",
					PublishedAt: now,
					UpdatedAt:   now,
				},
				FeedTitle: wantFeed.Title,
			},
			{
				Entry: feed.Entry{
					FeedUUID:    gotFeed.UUID,
					URL:         "http://test.local/hello-world",
					Title:       "Hello World",
					PublishedAt: yesterday,
					UpdatedAt:   yesterday,
				},
				FeedTitle: wantFeed.Title,
			},
		}
		wantNEntries := uint(len(wantEntries))

		entryCount, err := r.FeedEntryGetCount(testUser.UUID)
		if err != nil {
			t.Fatalf("failed to retrieve entry count: %q", err)
		}

		if entryCount != wantNEntries {
			t.Errorf("want %d entries, got %d", len(wantEntries), entryCount)
		}

		gotEntries, err := r.FeedSubscriptionEntryGetN(testUser.UUID, wantNEntries, 0)
		if err != nil {
			t.Fatalf("failed to retrieve entries: %q", err)
		}

		querying.AssertSubscribedFeedEntriesEqual(t, gotEntries, wantEntries)

		// 3. Teardown
		if err := fs.DeleteSubscription(testUser.UUID, gotSubscription.UUID); err != nil {
			t.Fatalf("failed to delete subscription: %q", err)
		}

		if _, err := r.FeedGetByUUID(gotFeed.UUID); !errors.Is(err, feed.ErrFeedNotFound) {
			t.Errorf("want ErrFeedNotFound, got %q", err)
		}

		if err := fs.DeleteCategory(testUser.UUID, category.UUID); err != nil {
			t.Fatalf("failed to delete category: %q", err)
		}
	})

	t.Run("update preferences", func(t *testing.T) {
		preferences := feed.Preferences{
			UserUUID:    testUser.UUID,
			ShowEntries: feed.EntryVisibilityRead,
		}

		if err := fs.UpdatePreferences(preferences); err != nil {
			t.Fatalf("failed to update preferences: %q", err)
		}

		gotPreferences, err := fs.PreferencesByUserUUID(testUser.UUID)
		if err != nil {
			t.Fatalf("failed to retrieve preferences: %q", err)
		}

		if gotPreferences.ShowEntries != preferences.ShowEntries {
			t.Errorf("want ShowEntries %q, got %q", preferences.ShowEntries, gotPreferences.ShowEntries)
		}

		now := time.Now().UTC()
		assert.TimeAlmostEquals(t, "UpdatedAt", gotPreferences.UpdatedAt, now, assert.TimeComparisonDelta)
	})
}
