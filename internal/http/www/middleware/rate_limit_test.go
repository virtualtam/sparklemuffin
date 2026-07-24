// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/virtualtam/sparklemuffin/internal/http/www/controller"
	"github.com/virtualtam/sparklemuffin/pkg/session"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

func TestRateLimit_PerAccount(t *testing.T) {
	sessionRepo := &session.FakeRepository{}
	sessionService, err := session.NewService(sessionRepo, "hmac-key")
	if err != nil {
		t.Fatal(err)
	}
	userService := user.NewService(&user.FakeRepository{})

	mux := chi.NewMux()
	controller.RegisterSessionHandlers(mux, sessionService, userService)

	var lastCode int
	for range 6 {
		lastCode = postLoginForm(t, mux, "victim@example.com", "wrong-password")
	}

	if lastCode != http.StatusTooManyRequests {
		t.Errorf("want status %d after exceeding the per-account limit, got %d", http.StatusTooManyRequests, lastCode)
	}
}

func TestRateLimit_PerIP(t *testing.T) {
	sessionRepo := &session.FakeRepository{}
	sessionService, err := session.NewService(sessionRepo, "hmac-key")
	if err != nil {
		t.Fatal(err)
	}
	userService := user.NewService(&user.FakeRepository{})

	mux := chi.NewMux()
	controller.RegisterSessionHandlers(mux, sessionService, userService)

	// distinct email per request keeps the per-account limiter from tripping first
	var lastCode int
	for i := range 61 {
		lastCode = postLoginForm(t, mux, fmt.Sprintf("victim-%d@example.com", i), "wrong-password")
	}

	if lastCode != http.StatusTooManyRequests {
		t.Errorf("want status %d after exceeding the per-IP limit, got %d", http.StatusTooManyRequests, lastCode)
	}
}

func postLoginForm(t *testing.T, h http.Handler, email string, password string) int {
	t.Helper()

	form := url.Values{"email": {email}, "password": {password}}
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w.Code
}
