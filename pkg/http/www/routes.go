package www

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/virtualtam/yawbe/pkg/http/www/static"
)

func AddRoutes(r *mux.Router) {
	r.Handle("/", HomeView)
	r.HandleFunc("/static/", http.NotFound)

	r.PathPrefix("/static/").Handler(http.StripPrefix(
		"/static/",
		cacheControlWrapper(http.FileServer(http.FS(static.FS)))))
}

func cacheControlWrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=2592000") // 30 days
		h.ServeHTTP(w, r)
	})
}
