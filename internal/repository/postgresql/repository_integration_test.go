package postgresql_test

import (
	"context"
	"errors"
	"math/rand"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jaswdr/faker"
	"github.com/jmoiron/sqlx"
	"github.com/testcontainers/testcontainers-go"
	testpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/exp/slices"

	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/migrations"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	"github.com/virtualtam/sparklemuffin/pkg/querying"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

const (
	databaseDriver   = "pgx"
	databaseName     = "testdb"
	databaseUser     = "testuser"
	databasePassword = "testpass"
)

func createTestDatabase(t *testing.T, ctx context.Context) *sqlx.DB {
	t.Helper()

	pgContainer, err := testpostgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15"),
		testpostgres.WithDatabase(databaseName),
		testpostgres.WithUsername(databaseUser),
		testpostgres.WithPassword(databasePassword),
		testcontainers.WithWaitStrategy(
			wait.
				ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to create postgres container: %q", err)
	}

	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate postgres container: %q", err)
		}
	})

	databaseURI, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to obtain postgres connection string: %q", err)
	}

	db, err := sqlx.Connect(databaseDriver, databaseURI)
	if err != nil {
		t.Fatalf("failed to open database connection: %q", err)
	}

	migrateTestDatabase(t, db)

	return db
}

func migrateTestDatabase(t *testing.T, db *sqlx.DB) {
	migrationsSource, err := iofs.New(migrations.FS, ".")
	if err != nil {
		t.Fatalf("failed to open the database migration filesystem: %q", err)
	}

	driver, err := migratepgx.WithInstance(db.DB, &migratepgx.Config{})
	if err != nil {
		t.Fatalf("failed to prepare the database driver: %q", err)
	}

	migrater, err := migrate.NewWithInstance(
		"iofs",
		migrationsSource,
		databaseDriver,
		driver,
	)
	if err != nil {
		t.Fatalf("failed to load database migrations: %q", err)
	}

	if err := migrater.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("failed to apply database migrations: %q", err)
	}
}

func generateFakeUser(t *testing.T, fake *faker.Faker) user.User {
	person := fake.Person()
	internet := fake.Internet()

	// Nicknames must match user.nickNameRegex
	nick := strings.ReplaceAll(internet.User(), ".", "")

	return user.User{
		Email:       person.Contact().Email,
		NickName:    nick,
		DisplayName: person.Name(),
		Password:    internet.Password(),
	}
}

func generateUniqueSortedTags(fake *faker.Faker, nTags int) []string {
	tags := []string{}
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

func assertBookmarksEqual(t *testing.T, got, want bookmark.Bookmark) {
	t.Helper()

	if got.URL != want.URL {
		t.Errorf("want URL %q, got %q", want.URL, got.URL)
	}

	if got.Title != want.Title {
		t.Errorf("want Title %q, got %q", want.Title, got.Title)
	}

	if got.Description != want.Description {
		t.Errorf("want Description %q, got %q", want.Description, got.Description)
	}

	if got.Private != want.Private {
		t.Errorf("want Private %t, got %t", want.Private, got.Private)
	}

	if len(got.Tags) != len(want.Tags) {
		t.Fatalf("want %d tags, got %d", len(want.Tags), len(got.Tags))
	}

	for i, wantTag := range want.Tags {
		if got.Tags[i] != wantTag {
			t.Errorf("want tag %d Name %q, got %q", i, wantTag, got.Tags[i])
		}
	}
}

func TestBookmarkService(t *testing.T) {
	ctx := context.Background()
	db := createTestDatabase(t, ctx)
	r := postgresql.NewRepository(db)
	bs := bookmark.NewService(r)
	us := user.NewService(r)

	fake := faker.New()

	u := generateFakeUser(t, &fake)

	if err := us.Add(u); err != nil {
		t.Fatalf("failed to create user: %q", err)
	}

	testUser, err := us.ByNickName(u.NickName)
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
				if err := bs.Add(tc.bkm); err != nil {
					t.Fatalf("failed to create bookmark: %q", err)
				}

				gotBookmark, err := bs.ByURL(testUser.UUID, tc.bkm.URL)
				if err != nil {
					t.Fatalf("failed to retrieve bookmark: %q", err)
				}

				if gotBookmark.UserUUID != testUser.UUID {
					t.Errorf("want UserUUID %q, got %q", testUser.UUID, tc.bkm.UserUUID)
				}

				assertBookmarksEqual(t, gotBookmark, tc.bkm)

				if err := bs.Delete(testUser.UUID, gotBookmark.UID); err != nil {
					t.Fatalf("failed to delete bookmark: %q", err)
				}

				_, err = bs.ByUID(testUser.UUID, gotBookmark.UID)
				if !errors.Is(err, bookmark.ErrNotFound) {
					t.Fatalf("want %q, got %q", bookmark.ErrNotFound, err)
				}
			})
		}
	})

	t.Run("create, update and delete bookmark", func(t *testing.T) {
		bkm := bookmark.Bookmark{
			UserUUID:    testUser.UUID,
			URL:         fake.Internet().URL(),
			Title:       fake.Lorem().Sentence(5),
			Description: fake.Lorem().Text(500),
			Tags:        generateUniqueSortedTags(&fake, 10),
		}

		if err := bs.Add(bkm); err != nil {
			t.Fatalf("failed to create bookmark: %q", err)
		}

		gotBookmark, err := bs.ByURL(testUser.UUID, bkm.URL)
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

		if err := bs.Update(updatedBookmark); err != nil {
			t.Fatalf("failed to update bookmark: %q", err)
		}

		gotUpdatedBookmark, err := bs.ByUID(testUser.UUID, gotBookmark.UID)
		if err != nil {
			t.Fatalf("failed to retrieve bookmark: %q", err)
		}

		assertBookmarksEqual(t, gotUpdatedBookmark, updatedBookmark)

		if err := bs.Delete(testUser.UUID, gotBookmark.UID); err != nil {
			t.Fatalf("failed to delete bookmark: %q", err)
		}

		_, err = bs.ByUID(testUser.UUID, gotBookmark.UID)
		if !errors.Is(err, bookmark.ErrNotFound) {
			t.Fatalf("want %q, got %q", bookmark.ErrNotFound, err)
		}
	})

	t.Run("update tag", func(t *testing.T) {
		oldTagName := "common/tag2"
		newTagName := "common/renamed"
		commonTags := []string{"common/tag1", oldTagName}
		nBookmarks := 10
		nRandomTags := 10
		nTags := nRandomTags + len(commonTags)

		for i := 0; i < nBookmarks; i++ {
			tags := append(commonTags, generateUniqueSortedTags(&fake, nRandomTags)...)
			sort.Strings(tags)

			bkm := bookmark.Bookmark{
				UserUUID:    testUser.UUID,
				URL:         fake.Internet().URL(),
				Title:       fake.Lorem().Sentence(5),
				Description: fake.Lorem().Text(500),
				Tags:        tags,
			}

			if err := bs.Add(bkm); err != nil {
				t.Fatalf("failed to create bookmark: %q", err)
			}
		}

		uq := bookmark.TagUpdateQuery{
			UserUUID:    testUser.UUID,
			CurrentName: oldTagName,
			NewName:     newTagName,
		}

		got, err := bs.UpdateTag(uq)
		if err != nil {
			t.Fatalf("failed to update tag: %q", err)
		}

		if got != int64(nBookmarks) {
			t.Errorf("want %d updated bookmarks, got %d", nBookmarks, got)
		}

		allBookmarks, err := bs.All(testUser.UUID)
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
			if err := bs.Delete(testUser.UUID, b.UID); err != nil {
				t.Fatalf("failed to delete bookmark: %q", err)
			}
		}
	})

	t.Run("delete tag", func(t *testing.T) {
		deletedTagName := "common/tag1"
		commonTags := []string{deletedTagName, "common/tag2"}
		nBookmarks := 10
		nRandomTags := 10
		nTags := nRandomTags + len(commonTags)

		for i := 0; i < nBookmarks; i++ {
			tags := append(commonTags, generateUniqueSortedTags(&fake, nRandomTags)...)
			sort.Strings(tags)

			bkm := bookmark.Bookmark{
				UserUUID:    testUser.UUID,
				URL:         fake.Internet().URL(),
				Title:       fake.Lorem().Sentence(5),
				Description: fake.Lorem().Text(500),
				Tags:        tags,
			}

			if err := bs.Add(bkm); err != nil {
				t.Fatalf("failed to create bookmark: %q", err)
			}
		}

		dq := bookmark.TagDeleteQuery{
			UserUUID: testUser.UUID,
			Name:     deletedTagName,
		}

		got, err := bs.DeleteTag(dq)
		if err != nil {
			t.Fatalf("failed to update tag: %q", err)
		}

		if got != int64(nBookmarks) {
			t.Errorf("want %d updated bookmarks, got %d", nBookmarks, got)
		}

		allBookmarks, err := bs.All(testUser.UUID)
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
			if err := bs.Delete(testUser.UUID, b.UID); err != nil {
				t.Fatalf("failed to delete bookmark: %q", err)
			}
		}
	})
}

func TestQueryingService(t *testing.T) {
	ctx := context.Background()
	db := createTestDatabase(t, ctx)
	r := postgresql.NewRepository(db)
	bs := bookmark.NewService(r)
	qs := querying.NewService(r)
	us := user.NewService(r)

	fake := faker.New()

	u := generateFakeUser(t, &fake)

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

	for i := 0; i < nBookmarks; i++ {
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
		gotPage, err := qs.BookmarksByPage(testUser.UUID, querying.VisibilityAll, 1)
		if err != nil {
			t.Fatalf("failed to query bookmarks: %q", err)
		}

		if len(gotPage.Bookmarks) != wantBookmarksPerPage {
			t.Fatalf("want %d bookmarks, got %d", wantBookmarksPerPage, len(gotPage.Bookmarks))
		}

		for i, b := range gotPage.Bookmarks {
			assertBookmarksEqual(t, b, bookmarks[nBookmarks-1-i])
		}
	})

	t.Run("page 2 of all bookmarks", func(t *testing.T) {
		gotPage, err := qs.BookmarksByPage(testUser.UUID, querying.VisibilityAll, 2)
		if err != nil {
			t.Fatalf("failed to query bookmarks: %q", err)
		}

		if len(gotPage.Bookmarks) != wantBookmarksPerPage {
			t.Fatalf("want %d bookmarks, got %d", wantBookmarksPerPage, len(gotPage.Bookmarks))
		}

		for i, b := range gotPage.Bookmarks {
			assertBookmarksEqual(t, b, bookmarks[nBookmarks-1-wantBookmarksPerPage-i])
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

		tags := []querying.Tag{}
		for name, count := range tagMap {
			tag := querying.NewTag(name, count)
			tags = append(tags, tag)
		}

		sort.Slice(tags, func(i, j int) bool {
			if tags[i].Count != tags[j].Count {
				return tags[i].Count > tags[j].Count
			}
			return tags[i].Name < tags[j].Name
		})

		gotTags, err := qs.Tags(testUser.UUID, querying.VisibilityAll)
		if err != nil {
			t.Fatalf("failed to get tags: %q", err)
		}

		gotTagNames, err := qs.TagNamesByCount(testUser.UUID, querying.VisibilityAll)
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

func TestUserService(t *testing.T) {
	ctx := context.Background()
	db := createTestDatabase(t, ctx)
	r := postgresql.NewRepository(db)
	s := user.NewService(r)

	fake := faker.New()

	t.Run("create, retrieve and delete user", func(t *testing.T) {
		u := generateFakeUser(t, &fake)

		if err := s.Add(u); err != nil {
			t.Fatalf("failed to create user: %q", err)
		}

		gotUser, err := s.ByNickName(u.NickName)
		if err != nil {
			t.Fatalf("failed to retrieve user: %q", err)
		}

		if gotUser.Email != u.Email {
			t.Errorf("want email %q, got %q", u.Email, gotUser.Email)
		}
		if gotUser.IsAdmin != u.IsAdmin {
			t.Errorf("want admin %t, got %t", u.IsAdmin, gotUser.IsAdmin)
		}
		if gotUser.UUID == "" {
			t.Error("want UUID to be set")
		}

		if err := s.DeleteByUUID(gotUser.UUID); err != nil {
			t.Fatalf("failed to delete user by UUID: %q", err)
		}

		_, err = s.ByNickName(u.NickName)
		if !errors.Is(err, user.ErrNotFound) {
			t.Fatalf("want %q, got %q", user.ErrNotFound, err)
		}
	})

	t.Run("update user", func(t *testing.T) {
		u := generateFakeUser(t, &fake)

		if err := s.Add(u); err != nil {
			t.Fatalf("failed to create user: %q", err)
		}

		gotUser, err := s.ByNickName(u.NickName)
		if err != nil {
			t.Fatalf("failed to retrieve user: %q", err)
		}

		updatedPerson := fake.Person()

		updatedUser := user.User{
			UUID:        gotUser.UUID,
			Email:       updatedPerson.Contact().Email,
			NickName:    gotUser.NickName,
			DisplayName: updatedPerson.Name(),
			Password:    fake.Internet().Password(),
		}

		if err := s.Update(updatedUser); err != nil {
			t.Fatalf("want no error, got %q", err)
		}

		gotUpdatedUser, err := s.ByUUID(gotUser.UUID)
		if err != nil {
			t.Fatalf("failed to retrieve user: %q", err)
		}

		if gotUpdatedUser.Email != updatedUser.Email {
			t.Errorf("want email %q, got %q", updatedUser.Email, gotUpdatedUser.Email)
		}
		if gotUpdatedUser.DisplayName != updatedUser.DisplayName {
			t.Errorf("want display name %q, got %q", updatedUser.DisplayName, gotUpdatedUser.DisplayName)
		}
	})

	t.Run("update user info with no change", func(t *testing.T) {
		u := generateFakeUser(t, &fake)

		if err := s.Add(u); err != nil {
			t.Fatalf("failed to create user: %q", err)
		}

		gotUser, err := s.ByNickName(u.NickName)
		if err != nil {
			t.Fatalf("failed to retrieve user: %q", err)
		}

		info := user.InfoUpdate{
			UUID:        gotUser.UUID,
			Email:       gotUser.Email,
			NickName:    gotUser.NickName,
			DisplayName: gotUser.DisplayName,
		}

		if err := s.UpdateInfo(info); err != nil {
			t.Fatalf("want no error, got %q", err)
		}

		gotUpdatedUser, err := s.ByUUID(gotUser.UUID)
		if err != nil {
			t.Fatalf("failed to retrieve user: %q", err)
		}

		if gotUpdatedUser.Email != u.Email {
			t.Errorf("want email %q, got %q", u.Email, gotUpdatedUser.Email)
		}
		if gotUpdatedUser.DisplayName != u.DisplayName {
			t.Errorf("want display name %q, got %q", u.DisplayName, gotUpdatedUser.DisplayName)
		}
	})

	t.Run("update user info", func(t *testing.T) {
		u := generateFakeUser(t, &fake)

		if err := s.Add(u); err != nil {
			t.Fatalf("failed to create user: %q", err)
		}

		gotUser, err := s.ByNickName(u.NickName)
		if err != nil {
			t.Fatalf("failed to retrieve user: %q", err)
		}

		newPerson := fake.Person()

		info := user.InfoUpdate{
			UUID:        gotUser.UUID,
			Email:       newPerson.Contact().Email,
			NickName:    gotUser.NickName,
			DisplayName: newPerson.Name(),
		}

		if err := s.UpdateInfo(info); err != nil {
			t.Fatalf("want no error, got %q", err)
		}

		gotUpdatedUser, err := s.ByUUID(gotUser.UUID)
		if err != nil {
			t.Fatalf("failed to retrieve user: %q", err)
		}

		if gotUpdatedUser.Email != info.Email {
			t.Errorf("want email %q, got %q", info.Email, gotUpdatedUser.Email)
		}
		if gotUpdatedUser.DisplayName != info.DisplayName {
			t.Errorf("want display name %q, got %q", info.DisplayName, gotUpdatedUser.DisplayName)
		}
	})

	t.Run("update user password", func(t *testing.T) {
		u := generateFakeUser(t, &fake)

		if err := s.Add(u); err != nil {
			t.Fatalf("failed to create user: %q", err)
		}

		gotUser, err := s.ByNickName(u.NickName)
		if err != nil {
			t.Fatalf("failed to retrieve user: %q", err)
		}

		newPassword := fake.Internet().Password()

		passwordUpdate := user.PasswordUpdate{
			UUID:                    gotUser.UUID,
			CurrentPassword:         u.Password,
			NewPassword:             newPassword,
			NewPasswordConfirmation: newPassword,
		}

		if err := s.UpdatePassword(passwordUpdate); err != nil {
			t.Fatalf("want no error, got %q", err)
		}

		gotUpdatedUser, err := s.ByUUID(gotUser.UUID)
		if err != nil {
			t.Fatalf("failed to retrieve user: %q", err)
		}

		if gotUpdatedUser.PasswordHash == u.PasswordHash {
			t.Error("password hash was not updated")
		}
	})

	t.Run("authenticate user", func(t *testing.T) {
		u := generateFakeUser(t, &fake)

		if err := s.Add(u); err != nil {
			t.Fatalf("failed to create user: %q", err)
		}

		authenticatedUser, err := s.Authenticate(u.Email, u.Password)
		if err != nil {
			t.Fatalf("want no error, got %q", err)
		}

		if authenticatedUser.Email != u.Email {
			t.Errorf("want email %q, got %q", u.Email, authenticatedUser.Email)
		}
	})
}
