// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package controller

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jaswdr/faker/v2"

	"github.com/virtualtam/sparklemuffin/internal/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/internal/http/www/view"
	"github.com/virtualtam/sparklemuffin/pkg/session"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

// decodedFlashLevel decodes the "flash" cookie set on the response, if any,
// and returns its level (e.g. "success", "warning", "danger").
func decodedFlashLevel(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()

	for _, cookie := range w.Result().Cookies() {
		if cookie.Name != "flash" {
			continue
		}

		raw, err := base64.URLEncoding.DecodeString(cookie.Value)
		if err != nil {
			t.Fatalf("failed to decode flash cookie: %q", err)
		}

		var decoded struct {
			Level string `json:"level"`
		}
		if err := json.Unmarshal(raw, &decoded); err != nil {
			t.Fatalf("failed to unmarshal flash cookie: %q", err)
		}

		return decoded.Level
	}

	t.Fatal("want a flash cookie to be set")
	return ""
}

func TestHandlePasswordView(t *testing.T) {
	fake := faker.New()
	ctxUser := user.User{UUID: fake.UUID().V4(), Email: fake.Internet().Email()}

	ac := accountController{
		accountPasswordView: view.New("account/password.gohtml"),
	}

	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/account/password", nil)
	r = r.WithContext(httpcontext.WithUser(r.Context(), ctxUser))
	w := httptest.NewRecorder()

	ac.handlePasswordView()(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("want status 200, got %d, body:\n%s", w.Code, w.Body.String())
	}

	body := w.Body.String()
	if !strings.Contains(body, `id="new_password" name="new_password" minlength="8"`) {
		t.Errorf("want the new password field to enforce the minimum password length, got:\n%s", body)
	}
	if !strings.Contains(body, "Must be at least 8 characters long.") {
		t.Errorf("want a hint about the minimum password length, got:\n%s", body)
	}
	if got := strings.Count(body, "minlength="); got != 1 {
		t.Errorf("want only the new password field to enforce a minimum length, got %d occurrences of minlength=:\n%s", got, body)
	}
}

func TestHandlePasswordUpdate(t *testing.T) {
	fake := faker.New()

	t.Run("successful update revokes all of the user's sessions", func(t *testing.T) {
		userRepo := &user.FakeRepository{}
		userService := user.NewService(userRepo)

		newUser := user.User{
			UUID:        fake.UUID().V4(),
			Email:       "user@example.com",
			NickName:    "user1",
			DisplayName: "User One",
			Password:    "current-password1234",
		}
		if err := userService.Add(t.Context(), newUser); err != nil {
			t.Fatalf("failed to seed user: %q", err)
		}

		ctxUser, err := userService.ByNickName(t.Context(), newUser.NickName)
		if err != nil {
			t.Fatalf("failed to retrieve seeded user: %q", err)
		}

		sessionRepo := &session.FakeRepository{
			Sessions: []session.Session{
				{UserUUID: ctxUser.UUID, RememberTokenHash: "hash-1"},
				{UserUUID: "another-user", RememberTokenHash: "hash-2"},
			},
		}
		sessionService, err := session.NewService(sessionRepo, "hmac-key")
		if err != nil {
			t.Fatal(err)
		}

		ac := accountController{
			userService:    userService,
			sessionService: sessionService,
		}

		form := url.Values{}
		form.Set("current_password", "current-password1234")
		form.Set("new_password", "new-password1234")
		form.Set("new_password_confirmation", "new-password1234")

		r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/account/password", strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		ctx := httpcontext.WithUser(r.Context(), ctxUser)
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		ac.handlePasswordUpdate()(w, r)

		if w.Code != http.StatusSeeOther {
			t.Fatalf("want status %d, got %d, body:\n%s", http.StatusSeeOther, w.Code, w.Body.String())
		}

		if len(sessionRepo.Sessions) != 1 {
			t.Fatalf("want 1 remaining session, got %d", len(sessionRepo.Sessions))
		}
		if sessionRepo.Sessions[0].UserUUID != "another-user" {
			t.Errorf("want the other user's session to remain, got session for %q", sessionRepo.Sessions[0].UserUUID)
		}

		if got := decodedFlashLevel(t, w); got != "success" {
			t.Errorf("want a success flash, got %q", got)
		}
	})

	t.Run("session revocation failure still redirects but flashes a warning instead of success", func(t *testing.T) {
		userRepo := &user.FakeRepository{}
		userService := user.NewService(userRepo)

		newUser := user.User{
			UUID:        fake.UUID().V4(),
			Email:       "user2@example.com",
			NickName:    "user2",
			DisplayName: "User Two",
			Password:    "current-password1234",
		}
		if err := userService.Add(t.Context(), newUser); err != nil {
			t.Fatalf("failed to seed user: %q", err)
		}

		ctxUser, err := userService.ByNickName(t.Context(), newUser.NickName)
		if err != nil {
			t.Fatalf("failed to retrieve seeded user: %q", err)
		}

		sessionRepo := &session.FakeRepository{
			SessionDeleteByUserUUIDErr: errors.New("connection reset"),
		}
		sessionService, err := session.NewService(sessionRepo, "hmac-key")
		if err != nil {
			t.Fatal(err)
		}

		ac := accountController{
			userService:    userService,
			sessionService: sessionService,
		}

		form := url.Values{}
		form.Set("current_password", "current-password1234")
		form.Set("new_password", "new-password1234")
		form.Set("new_password_confirmation", "new-password1234")

		r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/account/password", strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		ctx := httpcontext.WithUser(r.Context(), ctxUser)
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		ac.handlePasswordUpdate()(w, r)

		if w.Code != http.StatusSeeOther {
			t.Fatalf("want status %d, got %d, body:\n%s", http.StatusSeeOther, w.Code, w.Body.String())
		}

		if got := decodedFlashLevel(t, w); got != "warning" {
			t.Errorf("want a warning flash when session revocation fails, got %q", got)
		}
	})
}
