// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package pgfeed_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jaswdr/faker/v2"

	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgbase"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgfeed"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgtaxonomy"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pguser"
	"github.com/virtualtam/sparklemuffin/internal/test/assert"
	"github.com/virtualtam/sparklemuffin/internal/test/feedtest"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	"github.com/virtualtam/sparklemuffin/pkg/feed/fetching"
	"github.com/virtualtam/sparklemuffin/pkg/feed/querying"
	"github.com/virtualtam/sparklemuffin/pkg/taxonomy"
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

	transport := feedtest.NewRoundTripperFromFeed(t, atomFeed)

	testHTTPClient := &http.Client{
		Transport: transport,
	}
	feedClient := fetching.NewClient(testHTTPClient, "sparklemuffin/test")

	r := pgfeed.NewRepository(pool)

	// avoid a real DNS lookup for these tests' test-only hostnames
	noopURLValidator := func(_ context.Context, _ string) error { return nil }
	fs := feed.NewService(r, feedClient, noopURLValidator)

	ur := pguser.NewRepository(pool)
	us := user.NewService(ur)

	u := user.FakeUser(t, new(faker.New()))

	if err := us.Add(t.Context(), u); err != nil {
		t.Fatalf("failed to create user: %q", err)
	}

	testUser, err := us.ByNickName(t.Context(), u.NickName)
	if err != nil {
		t.Fatalf("failed to retrieve user: %q", err)
	}

	t.Run("create, retrieve and delete category", func(t *testing.T) {
		ctx := t.Context()
		categoryName := "Test Category"

		// 1. Create category
		category, err := fs.CreateCategory(ctx, testUser.UUID, categoryName)
		if err != nil {
			t.Fatalf("failed to create category: %q", err)
		}

		gotCategory, err := fs.CategoryByUUID(ctx, testUser.UUID, category.UUID)
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
		if err := fs.DeleteCategory(ctx, testUser.UUID, category.UUID); err != nil {
			t.Fatalf("failed to delete category: %q", err)
		}

		_, err = fs.CategoryByUUID(ctx, testUser.UUID, category.UUID)
		if !errors.Is(err, feed.ErrCategoryNotFound) {
			t.Errorf("want ErrCategoryNotFound, got %q", err)
		} else if err == nil {
			t.Errorf("want ErrCategoryNotFound, got none")
		}
	})

	t.Run("create, update and delete category", func(t *testing.T) {
		ctx := t.Context()
		categoryName := "Test Category"

		category, err := fs.CreateCategory(ctx, testUser.UUID, categoryName)
		if err != nil {
			t.Fatalf("failed to create category: %q", err)
		}

		newCategoryName := "New Test Category"
		newCategory := feed.Category{
			UUID:     category.UUID,
			UserUUID: testUser.UUID,
			Name:     newCategoryName,
		}

		if err := fs.UpdateCategory(ctx, newCategory); err != nil {
			t.Fatalf("failed to update category: %q", err)
		}

		gotCategory, err := fs.CategoryByUUID(ctx, testUser.UUID, category.UUID)
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

		if err := fs.DeleteCategory(ctx, testUser.UUID, category.UUID); err != nil {
			t.Fatalf("failed to delete category: %q", err)
		}
	})

	t.Run("create, retrieve and delete feed subscription", func(t *testing.T) {
		ctx := t.Context()
		categoryName := "Subscriptions"

		// 1. Create category
		category, err := fs.CreateCategory(ctx, testUser.UUID, categoryName)
		if err != nil {
			t.Fatalf("failed to create category: %q", err)
		}

		gotCategory, err := fs.CategoryByUUID(ctx, testUser.UUID, category.UUID)
		if err != nil {
			t.Fatalf("failed to retrieve category: %q", err)
		}

		// 2. Create feed, entries and subscription
		if err := fs.Subscribe(ctx, testUser.UUID, category.UUID, "http://test.local", nil); err != nil {
			t.Fatalf("failed to subscribe to feed: %q", err)
		}

		// Retrieve the feed by URL to obtain its actual slug (which now includes a UUID suffix).
		gotFeed, err := r.FeedGetByURL(ctx, "http://test.local")
		if err != nil {
			t.Fatalf("failed to retrieve feed by URL: %q", err)
		}

		wantSlugPrefix := "local-test-"
		if !strings.HasPrefix(gotFeed.Slug, wantSlugPrefix) {
			t.Errorf("want Slug with prefix %q, got %q", wantSlugPrefix, gotFeed.Slug)
		}

		wantFeed := feed.Feed{
			FeedURL:      "http://test.local",
			Title:        "Local Test",
			Description:  "A simple syndication feed, for testing purposes.",
			Slug:         gotFeed.Slug,
			ETag:         feedETag,
			LastModified: feedLastModified,
			Hash:         feedHash,
			CreatedAt:    now,
			UpdatedAt:    now,
			FetchedAt:    now,
		}

		gotFeedBySlug, err := fs.FeedBySlug(ctx, gotFeed.Slug)
		if err != nil {
			t.Fatalf("failed to retrieve feed by slug: %q", err)
		}

		feed.AssertFeedEquals(t, gotFeedBySlug, wantFeed)

		gotSubscription, err := fs.SubscriptionByFeed(ctx, testUser.UUID, gotFeed.UUID)
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
				FeedUUID:    gotFeed.UUID,
				URL:         "http://test.local/first-post",
				Title:       "First post!",
				Summary:     "First post!\n\nThis is the first post!",
				PublishedAt: now,
				UpdatedAt:   now,
				FeedTitle:   wantFeed.Title,
			},
			{
				FeedUUID:    gotFeed.UUID,
				URL:         "http://test.local/hello-world",
				Title:       "Hello World",
				PublishedAt: yesterday,
				UpdatedAt:   yesterday,
				FeedTitle:   wantFeed.Title,
			},
		}
		wantNEntries := uint(len(wantEntries))

		entryCount, err := r.FeedEntryGetCount(ctx, testUser.UUID, feed.EntryVisibilityAll)
		if err != nil {
			t.Fatalf("failed to retrieve entry count: %q", err)
		}

		if entryCount != wantNEntries {
			t.Errorf("want %d entries, got %d", len(wantEntries), entryCount)
		}

		preferences, err := r.FeedPreferencesGetByUserUUID(ctx, testUser.UUID)
		if err != nil {
			t.Fatalf("failed to retrieve preferences: %q", err)
		}

		gotEntries, err := r.FeedSubscriptionEntryGetN(ctx, testUser.UUID, preferences, wantNEntries, 0)
		if err != nil {
			t.Fatalf("failed to retrieve entries: %q", err)
		}

		querying.AssertSubscribedFeedEntriesEqual(t, gotEntries, wantEntries)

		// 3. Teardown
		if err := fs.DeleteSubscription(ctx, testUser.UUID, gotSubscription.UUID); err != nil {
			t.Fatalf("failed to delete subscription: %q", err)
		}

		if _, err := r.FeedGetByUUID(ctx, gotFeed.UUID); !errors.Is(err, feed.ErrFeedNotFound) {
			t.Errorf("want ErrFeedNotFound, got %q", err)
		}

		if err := fs.DeleteCategory(ctx, testUser.UUID, category.UUID); err != nil {
			t.Fatalf("failed to delete category: %q", err)
		}
	})

	t.Run("two feeds with the same title get unique slugs", func(t *testing.T) {
		ctx := t.Context()

		category, err := fs.CreateCategory(ctx, testUser.UUID, "Same Title Test")
		if err != nil {
			t.Fatalf("failed to create category: %q", err)
		}

		// 1. Subscribe to two feeds with the same title.
		if err := fs.Subscribe(ctx, testUser.UUID, category.UUID, "http://feed1.test.local", nil); err != nil {
			t.Fatalf("failed to subscribe to feed1: %q", err)
		}
		if err := fs.Subscribe(ctx, testUser.UUID, category.UUID, "http://feed2.test.local", nil); err != nil {
			t.Fatalf("failed to subscribe to feed2: %q", err)
		}

		// 2. Retrieve feeds.
		feed1, err := r.FeedGetByURL(ctx, "http://feed1.test.local")
		if err != nil {
			t.Fatalf("failed to retrieve feed1: %q", err)
		}
		feed2, err := r.FeedGetByURL(ctx, "http://feed2.test.local")
		if err != nil {
			t.Fatalf("failed to retrieve feed2: %q", err)
		}

		if feed1.Slug == feed2.Slug {
			t.Errorf("want different slugs for feeds with the same title, got %q for both", feed1.Slug)
		}
		if !strings.HasPrefix(feed1.Slug, "local-test-") {
			t.Errorf("want feed1 Slug with prefix %q, got %q", "local-test-", feed1.Slug)
		}
		if !strings.HasPrefix(feed2.Slug, "local-test-") {
			t.Errorf("want feed2 Slug with prefix %q, got %q", "local-test-", feed2.Slug)
		}

		// 3. Teardown
		sub1, err := fs.SubscriptionByFeed(ctx, testUser.UUID, feed1.UUID)
		if err != nil {
			t.Fatalf("failed to retrieve subscription1: %q", err)
		}
		sub2, err := fs.SubscriptionByFeed(ctx, testUser.UUID, feed2.UUID)
		if err != nil {
			t.Fatalf("failed to retrieve subscription2: %q", err)
		}
		if err := fs.DeleteSubscription(ctx, testUser.UUID, sub1.UUID); err != nil {
			t.Fatalf("failed to delete subscription1: %q", err)
		}
		if err := fs.DeleteSubscription(ctx, testUser.UUID, sub2.UUID); err != nil {
			t.Fatalf("failed to delete subscription2: %q", err)
		}
		if err := fs.DeleteCategory(ctx, testUser.UUID, category.UUID); err != nil {
			t.Fatalf("failed to delete category: %q", err)
		}
	})

	t.Run("update preferences", func(t *testing.T) {
		ctx := t.Context()
		preferences := feed.Preferences{
			UserUUID:    testUser.UUID,
			ShowEntries: feed.EntryVisibilityRead,
		}

		if err := fs.UpdatePreferences(ctx, preferences); err != nil {
			t.Fatalf("failed to update preferences: %q", err)
		}

		gotPreferences, err := fs.PreferencesByUserUUID(ctx, testUser.UUID)
		if err != nil {
			t.Fatalf("failed to retrieve preferences: %q", err)
		}

		if gotPreferences.ShowEntries != preferences.ShowEntries {
			t.Errorf("want ShowEntries %q, got %q", preferences.ShowEntries, gotPreferences.ShowEntries)
		}

		now := time.Now().UTC()
		assert.TimeAlmostEquals(t, "UpdatedAt", gotPreferences.UpdatedAt, now, assert.TimeComparisonDelta)
	})

	t.Run("renaming a tag into an existing one reassigns feed_subscription_tags instead of losing them", func(t *testing.T) {
		ctx := t.Context()
		fake := faker.New()

		oldTag := "merge/old"
		newTag := "merge/new"

		category := generateFakeCategory(t, &fake, testUser.UUID, "Merge Test")
		if err := r.FeedCategoryCreate(ctx, category); err != nil {
			t.Fatalf("failed to create category: %q", err)
		}

		onlyOldFeed := generateFakeFeed(t, &fake, "Only Old", "", now)
		if err := r.FeedCreate(ctx, onlyOldFeed); err != nil {
			t.Fatalf("failed to create feed: %q", err)
		}

		onlyOld := feed.Subscription{
			UUID:         fake.UUID().V4(),
			FeedUUID:     onlyOldFeed.UUID,
			CategoryUUID: category.UUID,
			UserUUID:     testUser.UUID,
			Tags:         []string{oldTag},
		}
		if _, err := r.FeedSubscriptionCreate(ctx, onlyOld); err != nil {
			t.Fatalf("failed to create subscription: %q", err)
		}

		bothFeed := generateFakeFeed(t, &fake, "Both", "", now)
		if err := r.FeedCreate(ctx, bothFeed); err != nil {
			t.Fatalf("failed to create feed: %q", err)
		}

		both := feed.Subscription{
			UUID:         fake.UUID().V4(),
			FeedUUID:     bothFeed.UUID,
			CategoryUUID: category.UUID,
			UserUUID:     testUser.UUID,
			Tags:         []string{oldTag, newTag},
		}
		if _, err := r.FeedSubscriptionCreate(ctx, both); err != nil {
			t.Fatalf("failed to create subscription: %q", err)
		}

		taxonomyRepo := pgtaxonomy.NewRepository(pool, r.OnTagMerge)
		taxonomyService := taxonomy.NewService(taxonomyRepo)

		if _, err := taxonomyService.RenameTag(ctx, taxonomy.TagUpdateQuery{
			UserUUID:    testUser.UUID,
			CurrentName: oldTag,
			NewName:     newTag,
		}); err != nil {
			t.Fatalf("failed to rename tag: %q", err)
		}

		gotOnlyOld, err := r.FeedSubscriptionGetByUUID(ctx, testUser.UUID, onlyOld.UUID)
		if err != nil {
			t.Fatalf("failed to retrieve subscription: %q", err)
		}
		if len(gotOnlyOld.Tags) != 1 || gotOnlyOld.Tags[0] != newTag {
			t.Errorf("want tags [%s], got %v", newTag, gotOnlyOld.Tags)
		}

		gotBoth, err := r.FeedSubscriptionGetByUUID(ctx, testUser.UUID, both.UUID)
		if err != nil {
			t.Fatalf("failed to retrieve subscription: %q", err)
		}
		if len(gotBoth.Tags) != 1 || gotBoth.Tags[0] != newTag {
			t.Errorf("want tags [%s], got %v", newTag, gotBoth.Tags)
		}

		if got := countTaxonomyTagsByName(t, pool, testUser.UUID, oldTag); got != 0 {
			t.Errorf("want the merged-away tag %q gone from taxonomy_tags, got %d rows", oldTag, got)
		}
		if got := countTaxonomyTagsByName(t, pool, testUser.UUID, newTag); got != 1 {
			t.Errorf("want exactly 1 taxonomy_tags row for %q, got %d", newTag, got)
		}

		if err := r.FeedSubscriptionDelete(ctx, testUser.UUID, onlyOld.UUID); err != nil {
			t.Fatalf("failed to delete subscription: %q", err)
		}
		if err := r.FeedSubscriptionDelete(ctx, testUser.UUID, both.UUID); err != nil {
			t.Fatalf("failed to delete subscription: %q", err)
		}
	})

	t.Run("creating a subscription rejects a tag name containing whitespace", func(t *testing.T) {
		ctx := t.Context()
		fake := faker.New()

		category := generateFakeCategory(t, &fake, testUser.UUID, "Invalid Tag Test")
		if err := r.FeedCategoryCreate(ctx, category); err != nil {
			t.Fatalf("failed to create category: %q", err)
		}

		invalidTagFeed := generateFakeFeed(t, &fake, "Invalid Tag Feed", "", now)
		if err := r.FeedCreate(ctx, invalidTagFeed); err != nil {
			t.Fatalf("failed to create feed: %q", err)
		}

		subscription := feed.Subscription{
			UUID:         fake.UUID().V4(),
			FeedUUID:     invalidTagFeed.UUID,
			CategoryUUID: category.UUID,
			UserUUID:     testUser.UUID,
			Tags:         []string{"not a valid tag"},
		}

		_, err := r.FeedSubscriptionCreate(ctx, subscription)
		if !errors.Is(err, taxonomy.ErrTagNameContainsWhitespace) {
			t.Fatalf("want %q, got %q", taxonomy.ErrTagNameContainsWhitespace, err)
		}

		if _, err := r.FeedSubscriptionGetByUUID(ctx, testUser.UUID, subscription.UUID); !errors.Is(err, feed.ErrSubscriptionNotFound) {
			t.Fatalf("want the subscription not to have been created, got %q", err)
		}
	})
}

func countTaxonomyTagsByName(t *testing.T, pool *pgxpool.Pool, userUUID, name string) int {
	t.Helper()

	var count int

	err := pool.QueryRow(
		t.Context(),
		"SELECT COUNT(*) FROM taxonomy_tags WHERE user_uuid=$1 AND tag_name=$2",
		userUUID,
		name,
	).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count taxonomy_tags: %q", err)
	}

	return count
}
