package www

import "net/http"

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
