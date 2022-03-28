package www

import (
	"net/http"

	"github.com/gorilla/schema"
)

type loginForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

func parseForm(r *http.Request, dst interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	dec := schema.NewDecoder()

	if err := dec.Decode(dst, r.PostForm); err != nil {
		return err
	}

	return nil
}
