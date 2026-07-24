// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package pgsession_test

import (
	"errors"
	"testing"
	"time"

	"github.com/coder/quartz"
	"github.com/jaswdr/faker/v2"

	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgbase"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgsession"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pguser"
	"github.com/virtualtam/sparklemuffin/pkg/session"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

func TestRepository(t *testing.T) {
	pool := pgbase.CreateAndMigrateTestDatabase(t)

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

	r := pgsession.NewRepository(t.Context(), pool, quartz.NewReal())

	t.Run("expired sessions are not returned", func(t *testing.T) {
		sess := session.Session{
			UserUUID:               testUser.UUID,
			RememberTokenHash:      "expired-hash",
			RememberTokenExpiresAt: time.Now().UTC().Add(-1 * time.Hour),
		}
		if err := r.SessionAdd(t.Context(), sess); err != nil {
			t.Fatalf("failed to add session: %q", err)
		}

		_, err := r.SessionGetByRememberTokenHash(t.Context(), "expired-hash")
		if !errors.Is(err, session.ErrNotFound) {
			t.Fatalf("want %q, got %q", session.ErrNotFound, err)
		}
	})

	t.Run("non-expired sessions are returned", func(t *testing.T) {
		sess := session.Session{
			UserUUID:               testUser.UUID,
			RememberTokenHash:      "valid-hash",
			RememberTokenExpiresAt: time.Now().UTC().Add(1 * time.Hour),
		}
		if err := r.SessionAdd(t.Context(), sess); err != nil {
			t.Fatalf("failed to add session: %q", err)
		}

		got, err := r.SessionGetByRememberTokenHash(t.Context(), "valid-hash")
		if err != nil {
			t.Fatalf("failed to retrieve session: %q", err)
		}
		if got.UserUUID != testUser.UUID {
			t.Errorf("want user UUID %q, got %q", testUser.UUID, got.UserUUID)
		}
	})

	t.Run("delete by remember token hash", func(t *testing.T) {
		sess := session.Session{
			UserUUID:               testUser.UUID,
			RememberTokenHash:      "delete-by-hash",
			RememberTokenExpiresAt: time.Now().UTC().Add(1 * time.Hour),
		}
		if err := r.SessionAdd(t.Context(), sess); err != nil {
			t.Fatalf("failed to add session: %q", err)
		}

		if err := r.SessionDeleteByRememberTokenHash(t.Context(), "delete-by-hash"); err != nil {
			t.Fatalf("failed to delete session: %q", err)
		}

		_, err := r.SessionGetByRememberTokenHash(t.Context(), "delete-by-hash")
		if !errors.Is(err, session.ErrNotFound) {
			t.Fatalf("want %q, got %q", session.ErrNotFound, err)
		}
	})

	t.Run("delete by user UUID removes all sessions for that user", func(t *testing.T) {
		otherUser := pgbase.GenerateFakeUser(t, &fake)
		if err := us.Add(t.Context(), otherUser); err != nil {
			t.Fatalf("failed to create user: %q", err)
		}
		retrievedOtherUser, err := us.ByNickName(t.Context(), otherUser.NickName)
		if err != nil {
			t.Fatalf("failed to retrieve user: %q", err)
		}

		for _, hash := range []string{"multi-1", "multi-2"} {
			sess := session.Session{
				UserUUID:               testUser.UUID,
				RememberTokenHash:      hash,
				RememberTokenExpiresAt: time.Now().UTC().Add(1 * time.Hour),
			}
			if err := r.SessionAdd(t.Context(), sess); err != nil {
				t.Fatalf("failed to add session: %q", err)
			}
		}

		otherSess := session.Session{
			UserUUID:               retrievedOtherUser.UUID,
			RememberTokenHash:      "other-user-session",
			RememberTokenExpiresAt: time.Now().UTC().Add(1 * time.Hour),
		}
		if err := r.SessionAdd(t.Context(), otherSess); err != nil {
			t.Fatalf("failed to add session: %q", err)
		}

		if err := r.SessionDeleteByUserUUID(t.Context(), testUser.UUID); err != nil {
			t.Fatalf("failed to delete sessions: %q", err)
		}

		for _, hash := range []string{"multi-1", "multi-2"} {
			_, err := r.SessionGetByRememberTokenHash(t.Context(), hash)
			if !errors.Is(err, session.ErrNotFound) {
				t.Fatalf("want %q, got %q", session.ErrNotFound, err)
			}
		}

		got, err := r.SessionGetByRememberTokenHash(t.Context(), "other-user-session")
		if err != nil {
			t.Fatalf("failed to retrieve session: %q", err)
		}
		if got.UserUUID != retrievedOtherUser.UUID {
			t.Errorf("want user UUID %q, got %q", retrievedOtherUser.UUID, got.UserUUID)
		}
	})
}

func TestRepository_InvalidateSessions(t *testing.T) {
	pool := pgbase.CreateAndMigrateTestDatabase(t)

	ur := pguser.NewRepository(pool)
	us := user.NewService(ur)

	u := pgbase.GenerateFakeUser(t, new(faker.New()))
	if err := us.Add(t.Context(), u); err != nil {
		t.Fatalf("failed to create user: %q", err)
	}

	testUser, err := us.ByNickName(t.Context(), u.NickName)
	if err != nil {
		t.Fatalf("failed to retrieve user: %q", err)
	}

	mClock := quartz.NewMock(t)
	r := pgsession.NewRepository(t.Context(), pool, mClock)

	expiredSess := session.Session{
		UserUUID:               testUser.UUID,
		RememberTokenHash:      "periodic-cleanup-expired",
		RememberTokenExpiresAt: time.Now().UTC().Add(-1 * time.Hour),
	}
	if err := r.SessionAdd(t.Context(), expiredSess); err != nil {
		t.Fatalf("failed to add session: %q", err)
	}

	validSess := session.Session{
		UserUUID:               testUser.UUID,
		RememberTokenHash:      "periodic-cleanup-valid",
		RememberTokenExpiresAt: time.Now().UTC().Add(1 * time.Hour),
	}
	if err := r.SessionAdd(t.Context(), validSess); err != nil {
		t.Fatalf("failed to add session: %q", err)
	}

	_, waiter := mClock.AdvanceNext()
	waiter.MustWait(t.Context())

	const (
		countSessionsQuery = "SELECT COUNT(*) FROM sessions WHERE remember_token_hash=$1"
	)

	var expiredCount int
	expiredRow := pool.QueryRow(t.Context(), countSessionsQuery, "periodic-cleanup-expired")
	if err := expiredRow.Scan(&expiredCount); err != nil {
		t.Fatalf("failed to count sessions: %q", err)
	}
	if expiredCount != 0 {
		t.Errorf("want expired session to be physically deleted, got %d remaining", expiredCount)
	}

	var validCount int
	validRow := pool.QueryRow(t.Context(), countSessionsQuery, "periodic-cleanup-valid")
	if err := validRow.Scan(&validCount); err != nil {
		t.Fatalf("failed to count sessions: %q", err)
	}
	if validCount != 1 {
		t.Errorf("want non-expired session to remain, got %d remaining", validCount)
	}
}
