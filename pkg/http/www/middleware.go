package www

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/hlog"
)

// accessLogger logs information about incoming HTTP requests.
func accessLogger(r *http.Request, status, size int, dur time.Duration) {
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

// adminUser requires the user to have administration privileges to
// access content.
func adminUser(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := userValue(r.Context())

		if user == nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		if !user.IsAdmin {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		h(w, r)
	})
}

// authenticatedUser requires the user to be authenticated.
func authenticatedUser(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := userValue(r.Context())

		if user == nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		h(w, r)
	})
}

// staticCacheControl sets the Cache-Control header for static assets.
func staticCacheControl(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=2592000") // 30 days
		h.ServeHTTP(w, r)
	})
}
