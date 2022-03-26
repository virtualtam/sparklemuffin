package www

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/virtualtam/yawbe/pkg/http/www/static"
)

func AddRoutes(r *mux.Router) {
	r.Handle("/", HomeView)
	r.HandleFunc("/static/", http.NotFound)

	// TODO Cache-Control
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.FS(static.FS))))
}
