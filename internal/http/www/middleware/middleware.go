// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

// Package middleware provides HTTP middleware for user session management, static resource caching
// and security policy enforcement.
package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/hlog"

	"github.com/virtualtam/sparklemuffin/internal/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/internal/http/www/view"
)

var (
	errorView = view.NewError()
)

// AccessLogger logs information about incoming HTTP requests.
func AccessLogger(r *http.Request, status, size int, dur time.Duration) {
	reqID := middleware.GetReqID(r.Context())

	hlog.FromRequest(r).
		Info().
		Dur("duration_ms", dur).
		Str("host", r.Host).
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("remote_addr", r.RemoteAddr).
		Str("request_id", reqID).
		Int("size", size).
		Int("status", status).
		Msg("handle request")
}

// AdminUser requires the user to have administration privileges to
// access content.
func AdminUser(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := httpcontext.UserValue(r.Context())

		if user == nil {
			errorView.Render(w, r, http.StatusNotFound)
			return
		}

		if !user.IsAdmin {
			errorView.Render(w, r, http.StatusUnauthorized)
			return
		}

		h(w, r)
	}
}

// AuthenticatedUser requires the user to be authenticated.
func AuthenticatedUser(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := httpcontext.UserValue(r.Context())

		if user == nil {
			errorView.Render(w, r, http.StatusNotFound)
			return
		}

		h(w, r)
	}
}

// StaticCacheControl sets the Cache-Control header for static assets.
func StaticCacheControl(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=2592000") // 30 days
		h.ServeHTTP(w, r)
	}
}

// ContentSecurityPolicy sets the Content-Security-Policy header.
func ContentSecurityPolicy(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nonce, err := generateNonce()
		if err != nil {
			errorView.Render(w, r, http.StatusInternalServerError)
			return
		}

		policy := "default-src 'none'; " +
			"script-src 'self' 'unsafe-eval' 'nonce-" + nonce + "'; " +
			"style-src 'self'; " +
			"img-src 'self'; " +
			"font-src 'self'; " +
			"connect-src 'self'; " +
			"form-action 'self'; " +
			"base-uri 'self'; " +
			"frame-ancestors 'none'; " +
			"upgrade-insecure-requests"
		w.Header().Set("Content-Security-Policy", policy)

		ctx := httpcontext.WithCSPNonce(r.Context(), nonce)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// generateNonce generates a cryptographically secure 16-byte base64-encoded nonce.
func generateNonce() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
