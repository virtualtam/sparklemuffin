package www

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/virtualtam/yawbe/pkg/http/www/static"
	"github.com/virtualtam/yawbe/pkg/user"
)

// AddRoutes registers all HTTP handlers for the Web interface.
func AddRoutes(r *mux.Router, userService *user.Service) {
	// static pages
	r.Handle("/", HomeView)

	// authentication
	r.Handle("/login", loginView).Methods("GET")
	r.HandleFunc("/login", handleUserLogin(userService)).Methods("POST")

	// static assets
	r.HandleFunc("/static/", http.NotFound)
	r.PathPrefix("/static/").Handler(http.StripPrefix(
		"/static/",
		cacheControlWrapper(http.FileServer(http.FS(static.FS)))))
}

// cacheControlWrapper sets the Cache-Control header.
func cacheControlWrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=2592000") // 30 days
		h.ServeHTTP(w, r)
	})
}
