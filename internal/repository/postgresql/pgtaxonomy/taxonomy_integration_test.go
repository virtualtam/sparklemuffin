// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package pgtaxonomy_test

import (
	"errors"
	"testing"

	"github.com/jaswdr/faker/v2"

	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgbase"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgtaxonomy"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pguser"
	"github.com/virtualtam/sparklemuffin/pkg/taxonomy"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

func TestTaxonomyService(t *testing.T) {
	pool := pgbase.CreateAndMigrateTestDatabase(t)
	r := pgtaxonomy.NewRepository(pool)
	ts := taxonomy.NewService(r)

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

	t.Run("add, get and delete a tag", func(t *testing.T) {
		ctx := t.Context()
		name := fake.Lorem().Word()

		tags, err := ts.GetOrCreateTags(ctx, testUser.UUID, []string{name})
		if err != nil {
			t.Fatalf("failed to create tag: %q", err)
		}
		if len(tags) != 1 {
			t.Fatalf("want 1 tag, got %d", len(tags))
		}

		got, err := r.TagGetByName(ctx, testUser.UUID, name)
		if err != nil {
			t.Fatalf("failed to retrieve tag: %q", err)
		}
		if got.UUID != tags[0].UUID || got.Name != name {
			t.Errorf("want tag %+v, got %+v", tags[0], got)
		}

		if err := ts.DeleteTag(ctx, taxonomy.TagDeleteQuery{UserUUID: testUser.UUID, Name: name}); err != nil {
			t.Fatalf("failed to delete tag: %q", err)
		}

		_, err = r.TagGetByName(ctx, testUser.UUID, name)
		if !errors.Is(err, taxonomy.ErrNotFound) {
			t.Fatalf("want %q, got %q", taxonomy.ErrNotFound, err)
		}
	})

	t.Run("get a tag by UUID", func(t *testing.T) {
		ctx := t.Context()
		name := fake.Lorem().Word()

		tags, err := ts.GetOrCreateTags(ctx, testUser.UUID, []string{name})
		if err != nil {
			t.Fatalf("failed to create tag: %q", err)
		}

		got, err := ts.TagByUUID(ctx, testUser.UUID, tags[0].UUID)
		if err != nil {
			t.Fatalf("failed to retrieve tag: %q", err)
		}
		if got.Name != name {
			t.Errorf("want tag name %q, got %q", name, got.Name)
		}

		if err := ts.DeleteTag(ctx, taxonomy.TagDeleteQuery{UserUUID: testUser.UUID, Name: name}); err != nil {
			t.Fatalf("failed to delete tag: %q", err)
		}

		_, err = ts.TagByUUID(ctx, testUser.UUID, tags[0].UUID)
		if !errors.Is(err, taxonomy.ErrNotFound) {
			t.Fatalf("want %q, got %q", taxonomy.ErrNotFound, err)
		}
	})

	t.Run("rename a tag", func(t *testing.T) {
		ctx := t.Context()
		currentName := fake.Lorem().Word()
		newName := fake.Lorem().Word()

		if _, err := ts.GetOrCreateTags(ctx, testUser.UUID, []string{currentName}); err != nil {
			t.Fatalf("failed to create tag: %q", err)
		}

		if _, err := ts.RenameTag(ctx, taxonomy.TagUpdateQuery{UserUUID: testUser.UUID, CurrentName: currentName, NewName: newName}); err != nil {
			t.Fatalf("failed to rename tag: %q", err)
		}

		_, err := r.TagGetByName(ctx, testUser.UUID, currentName)
		if !errors.Is(err, taxonomy.ErrNotFound) {
			t.Fatalf("want %q for the old name, got %q", taxonomy.ErrNotFound, err)
		}

		got, err := r.TagGetByName(ctx, testUser.UUID, newName)
		if err != nil {
			t.Fatalf("failed to retrieve renamed tag: %q", err)
		}
		if got.Name != newName {
			t.Errorf("want tag name %q, got %q", newName, got.Name)
		}
	})

	t.Run("list and search tags", func(t *testing.T) {
		ctx := t.Context()

		fakeListUser := user.FakeUser(t, &fake)
		if err := us.Add(ctx, fakeListUser); err != nil {
			t.Fatalf("failed to create user: %q", err)
		}

		listUserUUID := fakeListUser.UUID

		names := []string{"golang", "gopher", "python"}
		if _, err := ts.GetOrCreateTags(ctx, listUserUUID, names); err != nil {
			t.Fatalf("failed to create tags: %q", err)
		}

		listPage, err := ts.ListTags(ctx, listUserUUID, 1)
		if err != nil {
			t.Fatalf("failed to list tags: %q", err)
		}
		if len(listPage.Tags) != len(names) {
			t.Fatalf("want %d tags, got %d", len(names), len(listPage.Tags))
		}

		searchPage, err := ts.SearchTags(ctx, listUserUUID, "go", 1)
		if err != nil {
			t.Fatalf("failed to search tags: %q", err)
		}
		if len(searchPage.Tags) != 2 {
			t.Fatalf("want 2 matching tags, got %d", len(searchPage.Tags))
		}
	})

	t.Run("TagAddTx and TagGetByNameTx participate in an external transaction", func(t *testing.T) {
		ctx := t.Context()
		name := fake.Lorem().Word()

		tag, err := taxonomy.NewTag(testUser.UUID, name)
		if err != nil {
			t.Fatalf("failed to build tag: %q", err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatalf("failed to begin transaction: %q", err)
		}

		if err := r.TagAddTx(ctx, tx, tag); err != nil {
			t.Fatalf("failed to add tag within the transaction: %q", err)
		}

		gotInTx, err := r.TagGetByNameTx(ctx, tx, testUser.UUID, name)
		if err != nil {
			t.Fatalf("failed to retrieve tag within the transaction: %q", err)
		}
		if gotInTx.UUID != tag.UUID {
			t.Errorf("want tag UUID %q within the transaction, got %q", tag.UUID, gotInTx.UUID)
		}

		if _, err := r.TagGetByName(ctx, testUser.UUID, name); !errors.Is(err, taxonomy.ErrNotFound) {
			t.Fatalf("want %q before commit, got %q", taxonomy.ErrNotFound, err)
		}

		if err := tx.Commit(ctx); err != nil {
			t.Fatalf("failed to commit transaction: %q", err)
		}

		gotAfterCommit, err := r.TagGetByName(ctx, testUser.UUID, name)
		if err != nil {
			t.Fatalf("failed to retrieve tag after commit: %q", err)
		}
		if gotAfterCommit.UUID != tag.UUID {
			t.Errorf("want tag UUID %q after commit, got %q", tag.UUID, gotAfterCommit.UUID)
		}
	})

	t.Run("TagAddManyTx and TagGetManyByNameTx resolve a mix of existing and new tags", func(t *testing.T) {
		ctx := t.Context()

		existingName := fake.Lorem().Word()
		newName1 := fake.Lorem().Word()
		newName2 := fake.Lorem().Word()

		existingTags, err := ts.GetOrCreateTags(ctx, testUser.UUID, []string{existingName})
		if err != nil {
			t.Fatalf("failed to create existing tag: %q", err)
		}
		existingTag := existingTags[0]

		names := []string{existingName, newName1, newName2}

		var tags []taxonomy.Tag
		for _, name := range names {
			tag, err := taxonomy.NewTag(testUser.UUID, name)
			if err != nil {
				t.Fatalf("failed to build tag: %q", err)
			}
			tags = append(tags, tag)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatalf("failed to begin transaction: %q", err)
		}

		if err := r.TagAddManyTx(ctx, tx, tags); err != nil {
			t.Fatalf("failed to add tags within the transaction: %q", err)
		}

		got, err := r.TagGetManyByNameTx(ctx, tx, testUser.UUID, names)
		if err != nil {
			t.Fatalf("failed to retrieve tags within the transaction: %q", err)
		}
		if len(got) != len(names) {
			t.Fatalf("want %d tags, got %d", len(names), len(got))
		}

		gotByName := make(map[string]taxonomy.Tag, len(got))
		for _, tag := range got {
			gotByName[tag.Name] = tag
		}

		if gotByName[existingName].UUID != existingTag.UUID {
			t.Errorf("want existing tag %q to keep UUID %q, got %q", existingName, existingTag.UUID, gotByName[existingName].UUID)
		}

		if err := tx.Commit(ctx); err != nil {
			t.Fatalf("failed to commit transaction: %q", err)
		}

		for _, name := range []string{newName1, newName2} {
			gotAfterCommit, err := r.TagGetByName(ctx, testUser.UUID, name)
			if err != nil {
				t.Fatalf("failed to retrieve tag %q after commit: %q", name, err)
			}
			if gotAfterCommit.Name != name {
				t.Errorf("want tag name %q after commit, got %q", name, gotAfterCommit.Name)
			}
		}
	})
}
