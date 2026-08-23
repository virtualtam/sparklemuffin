// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package pgbookmark_test

import (
	"context"
	"math/rand"
	"sort"
	"testing"

	"github.com/jaswdr/faker/v2"

	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgbase"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgbookmark"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pguser"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

// generateFakeTagPool returns n unique tag names, standing in for the fixed
// vocabulary of tags a user accumulates and reuses across several imports.
func generateFakeTagPool(fake *faker.Faker, n int) []string {
	return generateUniqueSortedTags(fake, n)
}

// sampleTags picks up to n distinct tags from pool, without repetition
// (a bookmark_tags row exists at most once per bookmark/tag pair). Rejection
// sampling is used instead of rand.Perm(len(pool)): n is always small
// (<=10) while pool can hold thousands of tags, so shuffling the whole pool
// per bookmark would dominate the benchmark's own allocation profile.
func sampleTags(pool []string, n int) []string {
	if n > len(pool) {
		n = len(pool)
	}

	chosen := make(map[int]bool, n)
	tags := make([]string, 0, n)

	for len(tags) < n {
		idx := rand.Intn(len(pool))
		if chosen[idx] {
			continue
		}

		chosen[idx] = true
		tags = append(tags, pool[idx])
	}

	sort.Strings(tags)

	return tags
}

// generateFakeBookmarkFromTagPool builds a bookmark with a unique UID and
// URL, and a random subset of tags drawn from pool, so repeated imports
// mostly resolve existing tags rather than creating new ones every time.
//
// The UID must be generated here: BookmarkAddMany/BookmarkUpsertMany are
// called directly on the repository, bypassing bookmark.Service.Add (which
// would otherwise assign it) -- a batch of bookmarks left with a zero-value
// UID would all collide on the unique_user_uuid_uid constraint and get
// silently dropped by ON CONFLICT DO NOTHING.
func generateFakeBookmarkFromTagPool(fake *faker.Faker, userUUID string, pool []string) bookmark.Bookmark {
	nTags := 1 + rand.Intn(10)

	b := bookmark.NewBookmark(userUUID)
	b.URL = fake.Internet().URL()
	b.Title = fake.Lorem().Sentence(5)
	b.Description = fake.Lorem().Text(200)
	b.Tags = sampleTags(pool, nTags)
	b.Normalize()

	return *b
}

// BenchmarkBookmarkAddMany measures the cost of importing a batch of
// bookmarks, including tag resolution against the taxonomy domain.
//
//	go test -bench=BenchmarkBookmarkAddMany -benchtime=5x -benchmem ./internal/repository/postgresql/pgbookmark/...
func BenchmarkBookmarkAddMany(b *testing.B) {
	pool := pgbase.CreateAndMigrateTestDatabase(b)
	r := pgbookmark.NewRepository(pool)

	ur := pguser.NewRepository(pool)
	us := user.NewService(ur)

	fake := faker.New()

	u := user.FakeUser(b, &fake)
	if err := us.Add(context.Background(), u); err != nil {
		b.Fatalf("failed to create user: %q", err)
	}

	for _, tc := range []struct {
		name            string
		numBookmarks    int
		numDistinctTags int
	}{
		{"100_bookmarks_50_tags", 100, 50},
		{"1000_bookmarks_500_tags", 1000, 500},
		{"5000_bookmarks_4000_tags", 5000, 4000},
	} {
		b.Run(tc.name, func(b *testing.B) {
			tagPool := generateFakeTagPool(&fake, tc.numDistinctTags)

			b.ReportAllocs()

			for b.Loop() {
				bookmarks := make([]bookmark.Bookmark, tc.numBookmarks)
				for i := range bookmarks {
					bookmarks[i] = generateFakeBookmarkFromTagPool(&fake, u.UUID, tagPool)
				}

				if _, err := r.BookmarkAddMany(context.Background(), bookmarks); err != nil {
					b.Fatalf("failed to bulk add bookmarks: %q", err)
				}
			}
		})
	}
}
