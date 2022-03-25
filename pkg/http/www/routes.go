package www

import "github.com/gorilla/mux"

func AddRoutes(r *mux.Router) {
	r.Handle("/", HomeView)
}
