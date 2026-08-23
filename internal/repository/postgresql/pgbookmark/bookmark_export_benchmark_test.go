// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package pgbookmark_test

import (
	"context"
	"testing"

	"github.com/jaswdr/faker/v2"

	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgbase"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgbookmark"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pguser"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

// BenchmarkBookmarkGetAll measures the cost of reading back a user's full
// bookmark set (with tags joined and aggregated), as used by
// bookmark/exporting.Service.
//
//	go test -bench=BenchmarkBookmarkGetAll -benchtime=5x -benchmem ./internal/repository/postgresql/pgbookmark/...
func BenchmarkBookmarkGetAll(b *testing.B) {
	pool := pgbase.CreateAndMigrateTestDatabase(b)
	r := pgbookmark.NewRepository(pool)

	ur := pguser.NewRepository(pool)
	us := user.NewService(ur)

	fake := faker.New()

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
			u := user.FakeUser(b, &fake)
			if err := us.Add(context.Background(), u); err != nil {
				b.Fatalf("failed to create user: %q", err)
			}

			tagPool := generateFakeTagPool(&fake, tc.numDistinctTags)

			bookmarks := make([]bookmark.Bookmark, tc.numBookmarks)
			for i := range bookmarks {
				bookmarks[i] = generateFakeBookmarkFromTagPool(&fake, u.UUID, tagPool)
			}

			if _, err := r.BookmarkAddMany(context.Background(), bookmarks); err != nil {
				b.Fatalf("failed to seed bookmarks: %q", err)
			}

			b.ReportAllocs()

			for b.Loop() {
				if _, err := r.BookmarkGetAll(context.Background(), u.UUID); err != nil {
					b.Fatalf("failed to get bookmarks: %q", err)
				}
			}
		})
	}
}
