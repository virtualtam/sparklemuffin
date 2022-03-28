package www

import (
	"net/http"

	"github.com/virtualtam/yawbe/pkg/http/www/rand"
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

		if err := setUserRememberToken(userService, w, &user); err != nil {
			viewData.AlertError(err)
			loginView.Render(w, r, viewData)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

const (
	UserRememberTokenNBytes     int    = 32
	UserRememberTokenCookieName string = "remember_me"
)

func setUserRememberToken(userService *user.Service, w http.ResponseWriter, user *user.User) error {
	if user.RememberToken == "" {
		token, err := rand.RandomBase64URLString(UserRememberTokenNBytes)
		if err != nil {
			return err
		}

		user.RememberToken = token
		err = userService.Update(*user)
		if err != nil {
			return err
		}
	}

	cookie := http.Cookie{
		Name:     UserRememberTokenCookieName,
		Value:    user.RememberToken,
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)

	return nil
}
