// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package controller

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/virtualtam/sparklemuffin/internal/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/pkg/session"
)

func TestHandleUserLogout(t *testing.T) {
	t.Run("no session cookie set", func(t *testing.T) {
		sessionRepo := &session.FakeRepository{}
		sessionService, err := session.NewService(sessionRepo, "hmac-key")
		if err != nil {
			t.Fatal(err)
		}

		sc := sessionController{sessionService: sessionService}

		r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/logout", nil)
		w := httptest.NewRecorder()

		sc.handleUserLogout()(w, r)

		if w.Code != http.StatusSeeOther {
			t.Fatalf("want status %d, got %d", http.StatusSeeOther, w.Code)
		}
	})

	t.Run("revokes the session bound to the request's remember token", func(t *testing.T) {
		sessionRepo := &session.FakeRepository{}
		sessionService, err := session.NewService(sessionRepo, "hmac-key")
		if err != nil {
			t.Fatal(err)
		}

		rememberToken := "test-remember-token"
		if err := sessionService.Add(t.Context(), session.Session{
			UserUUID:      testCtxUser.UUID,
			RememberToken: rememberToken,
		}); err != nil {
			t.Fatalf("failed to seed session: %q", err)
		}

		sc := sessionController{sessionService: sessionService}

		r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/logout", nil)
		r.AddCookie(&http.Cookie{Name: UserRememberTokenCookieName, Value: rememberToken})
		ctx := httpcontext.WithUser(r.Context(), testCtxUser)
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		sc.handleUserLogout()(w, r)

		if w.Code != http.StatusSeeOther {
			t.Fatalf("want status %d, got %d", http.StatusSeeOther, w.Code)
		}

		if len(sessionRepo.Sessions) != 0 {
			t.Errorf("want the session to be deleted, %d session(s) remain", len(sessionRepo.Sessions))
		}

		_, err = sessionService.ByRememberToken(t.Context(), rememberToken)
		if !errors.Is(err, session.ErrNotFound) {
			t.Errorf("want %q, got %q", session.ErrNotFound, err)
		}
	})
}
