package pgbookmark_test

import (
	"sort"
	"testing"

	"github.com/jaswdr/faker"
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

	u := pgbase.GenerateFakeUser(t, &fake)

	if err := us.Add(u); err != nil {
		t.Fatalf("failed to create user: %q", err)
	}

	testUser, err := us.ByNickName(u.NickName)
	if err != nil {
		t.Fatalf("failed to retrieve user: %q", err)
	}

	nBookmarks := 100
	nPrivateBookmarks := 0

	bookmarks := []bookmark.Bookmark{}

	for i := range nBookmarks {
		private := false
		if i%10 == 0 {
			private = true
			nPrivateBookmarks++
		}
		bookmarks = append(bookmarks, generateFakeBookmark(&fake, testUser.UUID, private))
	}

	for _, b := range bookmarks {
		if err := bs.Add(b); err != nil {
			t.Fatalf("failed to add bookmark: %q", err)
		}
	}

	wantBookmarksPerPage := 20

	t.Run("page 1 of all bookmarks", func(t *testing.T) {
		gotPage, err := qs.BookmarksByPage(testUser.UUID, bookmarkquerying.VisibilityAll, 1)
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
		gotPage, err := qs.BookmarksByPage(testUser.UUID, bookmarkquerying.VisibilityAll, 2)
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

	t.Run("all tags", func(t *testing.T) {
		tagMap := make(map[string]uint)

		for _, b := range bookmarks {
			for _, tag := range b.Tags {
				_, ok := tagMap[tag]
				if !ok {
					tagMap[tag] = 1
					continue
				}

				tagMap[tag]++
			}
		}

		tags := []bookmarkquerying.Tag{}
		for name, count := range tagMap {
			tag := bookmarkquerying.NewTag(name, count)
			tags = append(tags, tag)
		}

		sort.Slice(tags, func(i, j int) bool {
			if tags[i].Count != tags[j].Count {
				return tags[i].Count > tags[j].Count
			}
			return tags[i].Name < tags[j].Name
		})

		gotTags, err := qs.Tags(testUser.UUID, bookmarkquerying.VisibilityAll)
		if err != nil {
			t.Fatalf("failed to get tags: %q", err)
		}

		gotTagNames, err := qs.TagNamesByCount(testUser.UUID, bookmarkquerying.VisibilityAll)
		if err != nil {
			t.Fatalf("failed to get tag names: %q", err)
		}

		if len(gotTags) != len(tags) {
			t.Fatalf("want %d tags, got %d", len(tags), len(gotTags))
		}
		if len(gotTagNames) != len(tags) {
			t.Fatalf("want %d tag names, got %d", len(tags), len(gotTagNames))
		}

		for i, wantTag := range tags {
			if gotTags[i].Name != wantTag.Name {
				t.Errorf("want tag %d name %q, got %q", i, wantTag.Name, gotTags[i].Name)
			}
			if gotTags[i].Count != wantTag.Count {
				t.Errorf("want tag %d count %d, got %d", i, wantTag.Count, gotTags[i].Count)
			}

			if gotTagNames[i] != wantTag.Name {
				t.Errorf("want tagname %d value %q, got %q", i, wantTag.Name, gotTagNames[i])
			}
		}
	})
}
