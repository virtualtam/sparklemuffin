package www

import (
	"errors"
	"net/http"

	"github.com/virtualtam/yawbe/pkg/user"
)

func login(userService *user.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var viewData Data

		form := loginForm{}
		if err := parseForm(r, &form); err != nil {
			viewData.AlertError(err)
			loginView.Render(w, r, viewData)
			return
		}

		user, err := userService.AuthenticateUser(form.email, form.password)
		if err != nil {
			viewData.AlertError(err)
			loginView.Render(w, r, viewData)
			return
		}

		if err := setUserRememberToken(w, user); err != nil {
			viewData.AlertError(err)
			loginView.Render(w, r, viewData)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func setUserRememberToken(w http.ResponseWriter, user user.User) error {
	return errors.New("not implemented")
}
