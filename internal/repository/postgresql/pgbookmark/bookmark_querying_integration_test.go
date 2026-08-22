// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package pgbookmark_test

import (
	"testing"

	"github.com/jaswdr/faker/v2"

	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgbase"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgbookmark"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pguser"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	bookmarkquerying "github.com/virtualtam/sparklemuffin/pkg/bookmark/querying"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

func TestQueryingService(t *testing.T) {
	pool := pgbase.CreateAndMigrateTestDatabase(t)
	r := pgbookmark.NewRepository(pool)
	bs := bookmark.NewService(r)
	qs := bookmarkquerying.NewService(r)

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

	nBookmarks := 100
	nPrivateBookmarks := 0

	var bookmarks []bookmark.Bookmark

	for i := range nBookmarks {
		private := false
		if i%10 == 0 {
			private = true
			nPrivateBookmarks++
		}
		bookmarks = append(bookmarks, generateFakeBookmark(&fake, testUser.UUID, private))
	}

	for _, b := range bookmarks {
		if err := bs.Add(t.Context(), b); err != nil {
			t.Fatalf("failed to add bookmark: %q", err)
		}
	}

	wantBookmarksPerPage := 20

	t.Run("page 1 of all bookmarks", func(t *testing.T) {
		gotPage, err := qs.BookmarksByPage(t.Context(), testUser.UUID, bookmarkquerying.VisibilityAll, 1)
		if err != nil {
			t.Fatalf("failed to query bookmarks: %q", err)
		}

		if len(gotPage.Bookmarks) != wantBookmarksPerPage {
			t.Fatalf("want %d bookmarks, got %d", wantBookmarksPerPage, len(gotPage.Bookmarks))
		}

		for i, b := range gotPage.Bookmarks {
			bookmark.AssertBookmarkEquals(t, b, bookmarks[nBookmarks-1-i])
		}
	})

	t.Run("page 2 of all bookmarks", func(t *testing.T) {
		gotPage, err := qs.BookmarksByPage(t.Context(), testUser.UUID, bookmarkquerying.VisibilityAll, 2)
		if err != nil {
			t.Fatalf("failed to query bookmarks: %q", err)
		}

		if len(gotPage.Bookmarks) != wantBookmarksPerPage {
			t.Fatalf("want %d bookmarks, got %d", wantBookmarksPerPage, len(gotPage.Bookmarks))
		}

		for i, b := range gotPage.Bookmarks {
			bookmark.AssertBookmarkEquals(t, b, bookmarks[nBookmarks-1-wantBookmarksPerPage-i])
		}
	})

	t.Run("search matches a tag name that appears in neither the title nor the description", func(t *testing.T) {
		ctx := t.Context()

		distinctiveTag := "zzzsearchbytagonly"

		taggedBookmark := generateFakeBookmark(&fake, testUser.UUID, false)
		taggedBookmark.Tags = append(taggedBookmark.Tags, distinctiveTag)

		if err := bs.Add(ctx, taggedBookmark); err != nil {
			t.Fatalf("failed to add bookmark: %q", err)
		}

		gotPage, err := qs.BookmarksBySearchQueryAndPage(ctx, testUser.UUID, bookmarkquerying.VisibilityAll, distinctiveTag, 1)
		if err != nil {
			t.Fatalf("failed to search bookmarks: %q", err)
		}

		if len(gotPage.Bookmarks) != 1 {
			t.Fatalf("want 1 matching bookmark, got %d", len(gotPage.Bookmarks))
		}
		if gotPage.Bookmarks[0].URL != taggedBookmark.URL {
			t.Errorf("want matching bookmark %q, got %q", taggedBookmark.URL, gotPage.Bookmarks[0].URL)
		}

		if err := bs.Delete(ctx, testUser.UUID, gotPage.Bookmarks[0].UID); err != nil {
			t.Fatalf("failed to delete bookmark: %q", err)
		}
	})

	t.Run("count bookmarks by tag", func(t *testing.T) {
		ctx := t.Context()

		countTag := "zzzcountbytag"

		taggedBookmark1 := generateFakeBookmark(&fake, testUser.UUID, false)
		taggedBookmark1.Tags = append(taggedBookmark1.Tags, countTag)
		if err := bs.Add(ctx, taggedBookmark1); err != nil {
			t.Fatalf("failed to add bookmark: %q", err)
		}

		taggedBookmark2 := generateFakeBookmark(&fake, testUser.UUID, true)
		taggedBookmark2.Tags = append(taggedBookmark2.Tags, countTag)
		if err := bs.Add(ctx, taggedBookmark2); err != nil {
			t.Fatalf("failed to add bookmark: %q", err)
		}

		got, err := r.BookmarkGetCountsByTag(ctx, testUser.UUID)
		if err != nil {
			t.Fatalf("failed to count bookmarks by tag: %q", err)
		}
		if got[countTag] != 2 {
			t.Errorf("want 2 bookmarks for tag %q, got %d", countTag, got[countTag])
		}
		if got["unknown-tag"] != 0 {
			t.Errorf("want 0 bookmarks for tag %q, got %d", "unknown-tag", got["unknown-tag"])
		}

		gotBookmark1, err := bs.ByURL(ctx, testUser.UUID, taggedBookmark1.URL)
		if err != nil {
			t.Fatalf("failed to retrieve bookmark: %q", err)
		}
		gotBookmark2, err := bs.ByURL(ctx, testUser.UUID, taggedBookmark2.URL)
		if err != nil {
			t.Fatalf("failed to retrieve bookmark: %q", err)
		}

		if err := bs.Delete(ctx, testUser.UUID, gotBookmark1.UID); err != nil {
			t.Fatalf("failed to delete bookmark: %q", err)
		}
		if err := bs.Delete(ctx, testUser.UUID, gotBookmark2.UID); err != nil {
			t.Fatalf("failed to delete bookmark: %q", err)
		}
	})
}
