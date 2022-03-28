package www

import (
	"net/http"

	"github.com/virtualtam/yawbe/pkg/user"
)

func handleUserLogin(userService *user.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var viewData Data

		form := loginForm{}
		if err := parseForm(r, &form); err != nil {
			viewData.AlertError(err)
			loginView.Render(w, r, viewData)
			return
		}

		user, err := userService.Authenticate(form.Email, form.Password)
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
	return nil
}
