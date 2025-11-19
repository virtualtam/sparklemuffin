// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package pguser_test

import (
	"errors"
	"testing"

	"github.com/jaswdr/faker/v2"

	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgbase"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgfeed"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pguser"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

func TestUserService(t *testing.T) {
	pool := pgbase.CreateAndMigrateTestDatabase(t)
	r := pguser.NewRepository(pool)

	s := user.NewService(r)

	fr := pgfeed.NewRepository(pool)
	fs := feed.NewService(fr, nil)

	fake := faker.New()

	t.Run("create, retrieve and delete user", func(t *testing.T) {
		ctx := t.Context()
		u := pgbase.GenerateFakeUser(t, &fake)

		// 1. Create user
		if err := s.Add(ctx, u); err != nil {
			t.Fatalf("failed to create user: %q", err)
		}

		// 2. Retrieve user
		gotUser, err := s.ByNickName(ctx, u.NickName)
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

		// 3. Retrieve feed preferences
		gotPreferences, err := fs.PreferencesByUserUUID(ctx, gotUser.UUID)
		if err != nil {
			t.Fatalf("failed to retrieve feed preferences: %q", err)
		}

		if gotPreferences.ShowEntries != feed.EntryVisibilityAll {
			t.Errorf("want feed preference %q, got %q", feed.EntryVisibilityAll, gotPreferences.ShowEntries)
		}

		// 4. Delete user
		if err := s.DeleteByUUID(ctx, gotUser.UUID); err != nil {
			t.Fatalf("failed to delete user by UUID: %q", err)
		}

		_, err = s.ByNickName(ctx, u.NickName)
		if !errors.Is(err, user.ErrNotFound) {
			t.Fatalf("want %q, got %q", user.ErrNotFound, err)
		}
	})

	t.Run("update user", func(t *testing.T) {
		ctx := t.Context()
		u := pgbase.GenerateFakeUser(t, &fake)

		if err := s.Add(ctx, u); err != nil {
			t.Fatalf("failed to create user: %q", err)
		}

		gotUser, err := s.ByNickName(ctx, u.NickName)
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

		if err := s.Update(ctx, updatedUser); err != nil {
			t.Fatalf("want no error, got %q", err)
		}

		gotUpdatedUser, err := s.ByUUID(ctx, gotUser.UUID)
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
		ctx := t.Context()
		u := pgbase.GenerateFakeUser(t, &fake)

		if err := s.Add(ctx, u); err != nil {
			t.Fatalf("failed to create user: %q", err)
		}

		gotUser, err := s.ByNickName(ctx, u.NickName)
		if err != nil {
			t.Fatalf("failed to retrieve user: %q", err)
		}

		info := user.InfoUpdate{
			UUID:        gotUser.UUID,
			Email:       gotUser.Email,
			NickName:    gotUser.NickName,
			DisplayName: gotUser.DisplayName,
		}

		if err := s.UpdateInfo(ctx, info); err != nil {
			t.Fatalf("want no error, got %q", err)
		}

		gotUpdatedUser, err := s.ByUUID(ctx, gotUser.UUID)
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
		ctx := t.Context()
		u := pgbase.GenerateFakeUser(t, &fake)

		if err := s.Add(ctx, u); err != nil {
			t.Fatalf("failed to create user: %q", err)
		}

		gotUser, err := s.ByNickName(ctx, u.NickName)
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

		if err := s.UpdateInfo(ctx, info); err != nil {
			t.Fatalf("want no error, got %q", err)
		}

		gotUpdatedUser, err := s.ByUUID(ctx, gotUser.UUID)
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
		ctx := t.Context()
		u := pgbase.GenerateFakeUser(t, &fake)

		if err := s.Add(ctx, u); err != nil {
			t.Fatalf("failed to create user: %q", err)
		}

		gotUser, err := s.ByNickName(ctx, u.NickName)
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

		if err := s.UpdatePassword(ctx, passwordUpdate); err != nil {
			t.Fatalf("want no error, got %q", err)
		}

		gotUpdatedUser, err := s.ByUUID(ctx, gotUser.UUID)
		if err != nil {
			t.Fatalf("failed to retrieve user: %q", err)
		}

		if gotUpdatedUser.PasswordHash == u.PasswordHash {
			t.Error("password hash was not updated")
		}
	})

	t.Run("authenticate user", func(t *testing.T) {
		ctx := t.Context()
		u := pgbase.GenerateFakeUser(t, &fake)

		if err := s.Add(ctx, u); err != nil {
			t.Fatalf("failed to create user: %q", err)
		}

		authenticatedUser, err := s.Authenticate(ctx, u.Email, u.Password)
		if err != nil {
			t.Fatalf("want no error, got %q", err)
		}

		if authenticatedUser.Email != u.Email {
			t.Errorf("want email %q, got %q", u.Email, authenticatedUser.Email)
		}
	})
}
