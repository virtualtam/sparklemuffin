// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package migrations_test

import (
	"errors"
	"testing"

	"github.com/golang-migrate/migrate/v4"

	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgbase"
)

// TestMigration000014BackfillHandlesDuplicateBookmarkTags verifies that
// migration 000014's backfill of bookmark_tags does not abort when a
// bookmark's legacy tags array contains a literal duplicate (e.g. the
// pre-normalization app-level dedup was bypassed, or the array was edited
// directly).
func TestMigration000014BackfillHandlesDuplicateBookmarkTags(t *testing.T) {
	ctx := t.Context()
	pool, migrater := pgbase.CreateTestDatabaseAndMigrater(t)

	if err := migrater.Migrate(13); err != nil {
		t.Fatalf("failed to migrate to version 13: %q", err)
	}

	userUUID := "11111111-1111-1111-1111-111111111111"
	bookmarkUID := "22222222-2222-2222-2222-222222222222"

	if _, err := pool.Exec(
		ctx,
		`INSERT INTO users(uuid, email, nick_name, display_name, password_hash)
		 VALUES($1, 'dup-tags@example.com', 'dup-tags', 'Dup Tags', 'hash')`,
		userUUID,
	); err != nil {
		t.Fatalf("failed to seed user: %q", err)
	}

	if _, err := pool.Exec(
		ctx,
		`INSERT INTO bookmarks(uid, user_uuid, url, title, description, private, tags)
		 VALUES($1, $2, 'https://example.com/dup-tags', 'Duplicate tags', '', FALSE, ARRAY['dup', 'dup'])`,
		bookmarkUID,
		userUUID,
	); err != nil {
		t.Fatalf("failed to seed bookmark with a duplicate tag: %q", err)
	}

	if err := migrater.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migration backfill failed on a bookmark with a duplicate tag: %q", err)
	}

	var tagRowCount int
	if err := pool.QueryRow(
		ctx,
		`SELECT COUNT(*)
		 FROM bookmark_tags bt
		 JOIN taxonomy_tags tt ON tt.tag_uuid = bt.tag_uuid
		 WHERE bt.user_uuid=$1 AND bt.bookmark_uid=$2 AND tt.tag_name='dup'`,
		userUUID,
		bookmarkUID,
	).Scan(&tagRowCount); err != nil {
		t.Fatalf("failed to count bookmark_tags rows: %q", err)
	}

	if tagRowCount != 1 {
		t.Errorf("want exactly 1 bookmark_tags row for the duplicate tag, got %d", tagRowCount)
	}
}

// TestMigration000015BackfillHandlesDuplicateSubscriptionTags mirrors
// TestMigration000014BackfillHandlesDuplicateBookmarkTags for feed
// subscriptions: migration 000015's backfill of feed_subscription_tags must
// not abort when a subscription's legacy tags array contains a literal
// duplicate.
func TestMigration000015BackfillHandlesDuplicateSubscriptionTags(t *testing.T) {
	ctx := t.Context()
	pool, migrater := pgbase.CreateTestDatabaseAndMigrater(t)

	if err := migrater.Migrate(14); err != nil {
		t.Fatalf("failed to migrate to version 14: %q", err)
	}

	userUUID := "33333333-3333-3333-3333-333333333333"
	categoryUUID := "44444444-4444-4444-4444-444444444444"
	feedUUID := "55555555-5555-5555-5555-555555555555"
	subscriptionUUID := "66666666-6666-6666-6666-666666666666"

	if _, err := pool.Exec(
		ctx,
		`INSERT INTO users(uuid, email, nick_name, display_name, password_hash)
		 VALUES($1, 'dup-sub-tags@example.com', 'dup-sub-tags', 'Dup Sub Tags', 'hash')`,
		userUUID,
	); err != nil {
		t.Fatalf("failed to seed user: %q", err)
	}

	if _, err := pool.Exec(
		ctx,
		`INSERT INTO feed_categories(uuid, user_uuid, name, slug)
		 VALUES($1, $2, 'Category', 'category')`,
		categoryUUID,
		userUUID,
	); err != nil {
		t.Fatalf("failed to seed feed category: %q", err)
	}

	if _, err := pool.Exec(
		ctx,
		`INSERT INTO feed_feeds(uuid, feed_url, title, slug, etag, last_modified)
		 VALUES($1, 'https://example.com/feed.xml', 'Feed', 'feed', '', NOW())`,
		feedUUID,
	); err != nil {
		t.Fatalf("failed to seed feed: %q", err)
	}

	if _, err := pool.Exec(
		ctx,
		`INSERT INTO feed_subscriptions(uuid, category_uuid, feed_uuid, user_uuid, tags)
		 VALUES($1, $2, $3, $4, ARRAY['dup', 'dup'])`,
		subscriptionUUID,
		categoryUUID,
		feedUUID,
		userUUID,
	); err != nil {
		t.Fatalf("failed to seed subscription with a duplicate tag: %q", err)
	}

	if err := migrater.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migration backfill failed on a subscription with a duplicate tag: %q", err)
	}

	var tagRowCount int
	if err := pool.QueryRow(
		ctx,
		`SELECT COUNT(*)
		 FROM feed_subscription_tags fst
		 JOIN taxonomy_tags tt ON tt.tag_uuid = fst.tag_uuid
		 WHERE fst.user_uuid=$1 AND fst.subscription_uuid=$2 AND tt.tag_name='dup'`,
		userUUID,
		subscriptionUUID,
	).Scan(&tagRowCount); err != nil {
		t.Fatalf("failed to count feed_subscription_tags rows: %q", err)
	}

	if tagRowCount != 1 {
		t.Errorf("want exactly 1 feed_subscription_tags row for the duplicate tag, got %d", tagRowCount)
	}
}
