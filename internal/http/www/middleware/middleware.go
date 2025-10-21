// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package middleware

import (
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
