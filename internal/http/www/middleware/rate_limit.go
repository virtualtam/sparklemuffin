// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package middleware

import (
	"net/http"
	"strings"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/rs/zerolog/log"
)

const (
	loginRateLimitPerIPRequests = 60
	loginRateLimitPerIPWindow   = 1 * time.Minute

	loginRateLimitPerAccountRequests = 5
	loginRateLimitPerAccountWindow   = 1 * time.Minute

	missingEmailRateLimitKey = "missing-email"
)

// RateLimitLogin prevents brute-force login attacks by limiting login attempts by IP address
// and by user email.
//
// Responds with http.StatusTooManyRequests when triggered.
func RateLimitLogin(h http.Handler) http.Handler {
	return httprate.LimitBy(
		loginRateLimitPerIPRequests,
		loginRateLimitPerIPWindow,
		loginIPKeyFunc,
		httprate.WithLimitHandler(onLoginRateLimitExceeded),
	)(
		httprate.LimitBy(
			loginRateLimitPerAccountRequests,
			loginRateLimitPerAccountWindow,
			loginEmailKeyFunc,
			httprate.WithLimitHandler(onLoginRateLimitExceeded),
		)(h),
	)
}

func loginIPKeyFunc(r *http.Request) (string, error) {
	return httprate.CanonicalizeIP(chimiddleware.GetClientIP(r.Context())), nil
}

func loginEmailKeyFunc(r *http.Request) (string, error) {
	if err := r.ParseForm(); err != nil {
		return missingEmailRateLimitKey, nil // nolint: nilerr
	}

	email := strings.ToLower(strings.TrimSpace(r.PostFormValue("email")))
	if email == "" {
		return missingEmailRateLimitKey, nil
	}

	return email, nil
}

func onLoginRateLimitExceeded(w http.ResponseWriter, r *http.Request) {
	email, err := loginEmailKeyFunc(r)
	if err != nil {
		email = missingEmailRateLimitKey
	}

	log.Warn().
		Str("client_ip", chimiddleware.GetClientIP(r.Context())).
		Str("email", email).
		Msg("login: rate limit exceeded")

	http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
}
