// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package pgbookmark_test

import (
	"errors"
	"math/rand"
	"sort"
	"testing"

	"github.com/jaswdr/faker/v2"
	"golang.org/x/exp/slices"

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

	u := pgbase.GenerateFakeUser(t, &fake)

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

	t.Run("update tag", func(t *testing.T) {
		ctx := t.Context()

		oldTagName := "common/tag2"
		newTagName := "common/renamed"
		commonTags := []string{"common/tag1", oldTagName}
		nBookmarks := 10
		nRandomTags := 10
		nTags := nRandomTags + len(commonTags)

		for range nBookmarks {
			tags := commonTags
			tags = append(tags, generateUniqueSortedTags(&fake, nRandomTags)...)
			sort.Strings(tags)

			bkm := bookmark.Bookmark{
				UserUUID:    testUser.UUID,
				URL:         fake.Internet().URL(),
				Title:       fake.Lorem().Sentence(5),
				Description: fake.Lorem().Text(500),
				Tags:        tags,
			}

			if err := bs.Add(ctx, bkm); err != nil {
				t.Fatalf("failed to create bookmark: %q", err)
			}
		}

		uq := bookmark.TagUpdateQuery{
			UserUUID:    testUser.UUID,
			CurrentName: oldTagName,
			NewName:     newTagName,
		}

		got, err := bs.UpdateTag(ctx, uq)
		if err != nil {
			t.Fatalf("failed to update tag: %q", err)
		}

		if got != int64(nBookmarks) {
			t.Errorf("want %d updated bookmarks, got %d", nBookmarks, got)
		}

		allBookmarks, err := bs.All(ctx, testUser.UUID)
		if err != nil {
			t.Fatalf("failed to retrieve all bookmarks: %q", err)
		}

		for i, b := range allBookmarks {
			if len(b.Tags) != nTags {
				t.Errorf("want bookmark %d to have %d tags, got %d", i, nTags, len(b.Tags))
			}

			if slices.Contains(b.Tags, oldTagName) {
				t.Errorf("want bookmark %d not to have tag %s", i, oldTagName)
			}

			if !slices.Contains(b.Tags, newTagName) {
				t.Errorf("want bookmark %d to have tag %s", i, newTagName)
			}
		}

		for _, b := range allBookmarks {
			if err := bs.Delete(ctx, testUser.UUID, b.UID); err != nil {
				t.Fatalf("failed to delete bookmark: %q", err)
			}
		}
	})

	t.Run("delete tag", func(t *testing.T) {
		ctx := t.Context()

		deletedTagName := "common/tag1"
		commonTags := []string{deletedTagName, "common/tag2"}
		nBookmarks := 10
		nRandomTags := 10
		nTags := nRandomTags + len(commonTags)

		for range nBookmarks {
			tags := commonTags
			tags = append(tags, generateUniqueSortedTags(&fake, nRandomTags)...)
			sort.Strings(tags)

			bkm := bookmark.Bookmark{
				UserUUID:    testUser.UUID,
				URL:         fake.Internet().URL(),
				Title:       fake.Lorem().Sentence(5),
				Description: fake.Lorem().Text(500),
				Tags:        tags,
			}

			if err := bs.Add(ctx, bkm); err != nil {
				t.Fatalf("failed to create bookmark: %q", err)
			}
		}

		dq := bookmark.TagDeleteQuery{
			UserUUID: testUser.UUID,
			Name:     deletedTagName,
		}

		got, err := bs.DeleteTag(ctx, dq)
		if err != nil {
			t.Fatalf("failed to update tag: %q", err)
		}

		if got != int64(nBookmarks) {
			t.Errorf("want %d updated bookmarks, got %d", nBookmarks, got)
		}

		allBookmarks, err := bs.All(ctx, testUser.UUID)
		if err != nil {
			t.Fatalf("failed to retrieve all bookmarks: %q", err)
		}

		wantNTags := nTags - 1

		for i, b := range allBookmarks {
			if len(b.Tags) != wantNTags {
				t.Errorf("want bookmark %d to have %d tags, got %d", i, wantNTags, len(b.Tags))
			}

			if slices.Contains(b.Tags, deletedTagName) {
				t.Errorf("want bookmark %d not to have tag %s", i, deletedTagName)
			}
		}

		for _, b := range allBookmarks {
			if err := bs.Delete(ctx, testUser.UUID, b.UID); err != nil {
				t.Fatalf("failed to delete bookmark: %q", err)
			}
		}
	})
}
