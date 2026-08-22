// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package pgbookmark_test

import (
	"errors"
	"math/rand"
	"slices"
	"sort"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jaswdr/faker/v2"

	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgbase"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgbookmark"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pguser"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

func generateFakeBookmark(fake *faker.Faker, userUUID string, private bool) bookmark.Bookmark {
	nTags := rand.Intn(10)
	tags := generateUniqueSortedTags(fake, nTags)

	return bookmark.Bookmark{
		UserUUID:    userUUID,
		URL:         fake.Internet().URL(),
		Title:       fake.Lorem().Sentence(5),
		Description: fake.Lorem().Text(500),
		Tags:        tags,
		Private:     private,
	}
}

func generateUniqueSortedTags(fake *faker.Faker, nTags int) []string {
	var tags []string
	tagMap := map[string]bool{}

	for len(tags) < nTags {
		tag := fake.Lorem().Word()
		if tag == "" || tagMap[tag] {
			continue
		}

		tags = append(tags, tag)
		tagMap[tag] = true
	}

	sort.Strings(tags)

	return tags
}

func TestBookmarkService(t *testing.T) {
	pool := pgbase.CreateAndMigrateTestDatabase(t)
	r := pgbookmark.NewRepository(pool)
	bs := bookmark.NewService(r)

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

	t.Run("create, retrieve and delete bookmark", func(t *testing.T) {
		testCases := []struct {
			tname string
			bkm   bookmark.Bookmark
		}{
			{
				tname: "simple bookmark",
				bkm: bookmark.Bookmark{
					UserUUID: testUser.UUID,
					URL:      fake.Internet().URL(),
					Title:    fake.Lorem().Sentence(5),
				},
			},
			{
				tname: "bookmark with description",
				bkm: bookmark.Bookmark{
					UserUUID:    testUser.UUID,
					URL:         fake.Internet().URL(),
					Title:       fake.Lorem().Sentence(5),
					Description: fake.Lorem().Text(500),
				},
			},
			{
				tname: "bookmark with tags",
				bkm: bookmark.Bookmark{
					UserUUID: testUser.UUID,
					URL:      fake.Internet().URL(),
					Title:    fake.Lorem().Sentence(5),
					Tags:     generateUniqueSortedTags(&fake, 10),
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.tname, func(t *testing.T) {
				ctx := t.Context()

				if err := bs.Add(ctx, tc.bkm); err != nil {
					t.Fatalf("failed to create bookmark: %q", err)
				}

				gotBookmark, err := bs.ByURL(ctx, testUser.UUID, tc.bkm.URL)
				if err != nil {
					t.Fatalf("failed to retrieve bookmark: %q", err)
				}

				if gotBookmark.UserUUID != testUser.UUID {
					t.Errorf("want UserUUID %q, got %q", testUser.UUID, tc.bkm.UserUUID)
				}

				bookmark.AssertBookmarkEquals(t, gotBookmark, tc.bkm)

				if err := bs.Delete(ctx, testUser.UUID, gotBookmark.UID); err != nil {
					t.Fatalf("failed to delete bookmark: %q", err)
				}

				_, err = bs.ByUID(ctx, testUser.UUID, gotBookmark.UID)
				if !errors.Is(err, bookmark.ErrNotFound) {
					t.Fatalf("want %q, got %q", bookmark.ErrNotFound, err)
				}
			})
		}
	})

	t.Run("create, update and delete bookmark", func(t *testing.T) {
		ctx := t.Context()
		bkm := bookmark.Bookmark{
			UserUUID:    testUser.UUID,
			URL:         fake.Internet().URL(),
			Title:       fake.Lorem().Sentence(5),
			Description: fake.Lorem().Text(500),
			Tags:        generateUniqueSortedTags(&fake, 10),
		}

		if err := bs.Add(ctx, bkm); err != nil {
			t.Fatalf("failed to create bookmark: %q", err)
		}

		gotBookmark, err := bs.ByURL(ctx, testUser.UUID, bkm.URL)
		if err != nil {
			t.Fatalf("failed to retrieve bookmark: %q", err)
		}

		updatedBookmark := bookmark.Bookmark{
			UserUUID:    gotBookmark.UserUUID,
			UID:         gotBookmark.UID,
			URL:         gotBookmark.URL,
			Title:       fake.Lorem().Sentence(5),
			Description: fake.Lorem().Text(500),
			Tags:        generateUniqueSortedTags(&fake, 10),
		}

		if err := bs.Update(ctx, updatedBookmark); err != nil {
			t.Fatalf("failed to update bookmark: %q", err)
		}

		gotUpdatedBookmark, err := bs.ByUID(ctx, testUser.UUID, gotBookmark.UID)
		if err != nil {
			t.Fatalf("failed to retrieve bookmark: %q", err)
		}

		bookmark.AssertBookmarkEquals(t, gotUpdatedBookmark, updatedBookmark)

		if err := bs.Delete(ctx, testUser.UUID, gotBookmark.UID); err != nil {
			t.Fatalf("failed to delete bookmark: %q", err)
		}

		_, err = bs.ByUID(ctx, testUser.UUID, gotBookmark.UID)
		if !errors.Is(err, bookmark.ErrNotFound) {
			t.Fatalf("want %q, got %q", bookmark.ErrNotFound, err)
		}
	})

	t.Run("bookmark_tags and taxonomy_tags are populated on add and update", func(t *testing.T) {
		ctx := t.Context()

		tagA := "integration/tag-a"
		tagB := "integration/tag-b"

		bkm := bookmark.Bookmark{
			UserUUID: testUser.UUID,
			URL:      fake.Internet().URL(),
			Title:    fake.Lorem().Sentence(5),
			Tags:     []string{tagA, tagB},
		}

		if err := bs.Add(ctx, bkm); err != nil {
			t.Fatalf("failed to create bookmark: %q", err)
		}

		gotBookmark, err := bs.ByURL(ctx, testUser.UUID, bkm.URL)
		if err != nil {
			t.Fatalf("failed to retrieve bookmark: %q", err)
		}

		assertBookmarkTagNames(t, pool, testUser.UUID, gotBookmark.UID, []string{tagA, tagB})

		// A second bookmark reusing tagA must reuse the same taxonomy_tags row.
		bkm2 := bookmark.Bookmark{
			UserUUID: testUser.UUID,
			URL:      fake.Internet().URL(),
			Title:    fake.Lorem().Sentence(5),
			Tags:     []string{tagA},
		}

		if err := bs.Add(ctx, bkm2); err != nil {
			t.Fatalf("failed to create second bookmark: %q", err)
		}

		gotBookmark2, err := bs.ByURL(ctx, testUser.UUID, bkm2.URL)
		if err != nil {
			t.Fatalf("failed to retrieve second bookmark: %q", err)
		}

		if got := countTaxonomyTagsByName(t, pool, testUser.UUID, tagA); got != 1 {
			t.Errorf("want exactly 1 taxonomy_tags row for %q, got %d", tagA, got)
		}

		// Updating the first bookmark replaces its tag set.
		updatedBookmark := bookmark.Bookmark{
			UserUUID: gotBookmark.UserUUID,
			UID:      gotBookmark.UID,
			URL:      gotBookmark.URL,
			Title:    gotBookmark.Title,
			Tags:     []string{tagB},
		}

		if err := bs.Update(ctx, updatedBookmark); err != nil {
			t.Fatalf("failed to update bookmark: %q", err)
		}

		assertBookmarkTagNames(t, pool, testUser.UUID, gotBookmark.UID, []string{tagB})

		if err := bs.Delete(ctx, testUser.UUID, gotBookmark.UID); err != nil {
			t.Fatalf("failed to delete bookmark: %q", err)
		}
		if err := bs.Delete(ctx, testUser.UUID, gotBookmark2.UID); err != nil {
			t.Fatalf("failed to delete second bookmark: %q", err)
		}
	})

	t.Run("adding a bookmark rejects a tag name containing whitespace", func(t *testing.T) {
		ctx := t.Context()

		bkm := bookmark.Bookmark{
			UserUUID: testUser.UUID,
			URL:      fake.Internet().URL(),
			Title:    fake.Lorem().Sentence(5),
			Tags:     []string{"not a valid tag"},
		}

		err := bs.Add(ctx, bkm)
		if !errors.Is(err, taxonomy.ErrTagNameContainsWhitespace) {
			t.Fatalf("want %q, got %q", taxonomy.ErrTagNameContainsWhitespace, err)
		}

		if _, err := bs.ByURL(ctx, testUser.UUID, bkm.URL); !errors.Is(err, bookmark.ErrNotFound) {
			t.Fatalf("want the bookmark not to have been created, got %q", err)
		}
	})

	t.Run("bulk add populates bookmark_tags and skips conflicting URLs untouched", func(t *testing.T) {
		ctx := t.Context()

		sharedTag := "bulk/shared"

		existing := bookmark.NewBookmark(testUser.UUID)
		existing.URL = fake.Internet().URL()
		existing.Title = fake.Lorem().Sentence(5)
		existing.Tags = []string{"bulk/existing-only"}
		existing.Normalize()

		if err := bs.Add(ctx, *existing); err != nil {
			t.Fatalf("failed to create existing bookmark: %q", err)
		}

		gotExisting, err := bs.ByURL(ctx, testUser.UUID, existing.URL)
		if err != nil {
			t.Fatalf("failed to retrieve existing bookmark: %q", err)
		}

		conflicting := bookmark.NewBookmark(testUser.UUID)
		conflicting.URL = existing.URL
		conflicting.Title = fake.Lorem().Sentence(5)
		conflicting.Tags = []string{sharedTag, "bulk/should-not-apply"}
		conflicting.Normalize()

		fresh := bookmark.NewBookmark(testUser.UUID)
		fresh.URL = fake.Internet().URL()
		fresh.Title = fake.Lorem().Sentence(5)
		fresh.Tags = []string{sharedTag}
		fresh.Normalize()

		rowsAffected, err := r.BookmarkAddMany(ctx, []bookmark.Bookmark{*conflicting, *fresh})
		if err != nil {
			t.Fatalf("failed to bulk add bookmarks: %q", err)
		}
		if rowsAffected != 1 {
			t.Errorf("want 1 row affected (fresh only), got %d", rowsAffected)
		}

		assertBookmarkTagNames(t, pool, testUser.UUID, gotExisting.UID, []string{"bulk/existing-only"})

		gotFresh, err := bs.ByURL(ctx, testUser.UUID, fresh.URL)
		if err != nil {
			t.Fatalf("failed to retrieve fresh bookmark: %q", err)
		}

		assertBookmarkTagNames(t, pool, testUser.UUID, gotFresh.UID, []string{sharedTag})

		if got := countTaxonomyTagsByName(t, pool, testUser.UUID, "bulk/should-not-apply"); got != 0 {
			t.Errorf("want no orphaned taxonomy_tags row for a tag exclusive to the skipped bookmark, got %d", got)
		}

		if err := bs.Delete(ctx, testUser.UUID, gotExisting.UID); err != nil {
			t.Fatalf("failed to delete existing bookmark: %q", err)
		}
		if err := bs.Delete(ctx, testUser.UUID, gotFresh.UID); err != nil {
			t.Fatalf("failed to delete fresh bookmark: %q", err)
		}
	})

	t.Run("bulk upsert replaces bookmark_tags using the existing row's UID", func(t *testing.T) {
		ctx := t.Context()

		oldTag := "bulk/old"
		newTag := "bulk/new"

		existing := bookmark.NewBookmark(testUser.UUID)
		existing.URL = fake.Internet().URL()
		existing.Title = fake.Lorem().Sentence(5)
		existing.Tags = []string{oldTag}
		existing.Normalize()

		if err := bs.Add(ctx, *existing); err != nil {
			t.Fatalf("failed to create bookmark: %q", err)
		}

		gotExisting, err := bs.ByURL(ctx, testUser.UUID, existing.URL)
		if err != nil {
			t.Fatalf("failed to retrieve bookmark: %q", err)
		}

		// A freshly generated bookmark deliberately carries a different UID
		// than the existing row: on conflict, the update must keep the
		// existing row's UID rather than the payload's.
		updated := bookmark.NewBookmark(testUser.UUID)
		updated.URL = existing.URL
		updated.Title = fake.Lorem().Sentence(5)
		updated.Tags = []string{newTag}
		updated.Normalize()

		if updated.UID == gotExisting.UID {
			t.Fatalf("test setup invariant violated: expected different generated UIDs")
		}

		rowsAffected, err := r.BookmarkUpsertMany(ctx, []bookmark.Bookmark{*updated})
		if err != nil {
			t.Fatalf("failed to bulk upsert bookmark: %q", err)
		}
		if rowsAffected != 1 {
			t.Errorf("want 1 row affected, got %d", rowsAffected)
		}

		gotUpdated, err := bs.ByURL(ctx, testUser.UUID, existing.URL)
		if err != nil {
			t.Fatalf("failed to retrieve updated bookmark: %q", err)
		}

		if gotUpdated.UID != gotExisting.UID {
			t.Fatalf("want the pre-existing UID %q to be preserved, got %q", gotExisting.UID, gotUpdated.UID)
		}

		assertBookmarkTagNames(t, pool, testUser.UUID, gotUpdated.UID, []string{newTag})

		if err := bs.Delete(ctx, testUser.UUID, gotUpdated.UID); err != nil {
			t.Fatalf("failed to delete bookmark: %q", err)
		}
	})
}

func assertBookmarkTagNames(t *testing.T, pool *pgxpool.Pool, userUUID, bookmarkUID string, want []string) {
	t.Helper()

	rows, err := pool.Query(
		t.Context(),
		`
		SELECT tt.tag_name
		FROM bookmark_tags bt
		JOIN taxonomy_tags tt ON tt.tag_uuid = bt.tag_uuid
		WHERE bt.user_uuid=$1
		AND   bt.bookmark_uid=$2
		ORDER BY tt.tag_name`,
		userUUID,
		bookmarkUID,
	)
	if err != nil {
		t.Fatalf("failed to query bookmark_tags: %q", err)
	}
	defer rows.Close()

	var got []string

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("failed to scan tag name: %q", err)
		}
		got = append(got, name)
	}

	wantSorted := slices.Clone(want)
	sort.Strings(wantSorted)

	if !slices.Equal(got, wantSorted) {
		t.Errorf("want bookmark_tags names %v, got %v", wantSorted, got)
	}
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
