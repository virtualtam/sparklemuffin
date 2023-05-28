package www

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

type accountHandlerContext struct {
	userService *user.Service

	accountView *view
}

func registerAccounthandlers(
	r *mux.Router,
	userService *user.Service,
) {
	hc := accountHandlerContext{
		userService: userService,

		accountView: newView("account/account.gohtml"),
	}

	// user account
	accountRouter := r.PathPrefix("/account").Subrouter()

	accountRouter.HandleFunc("", hc.handleAccountView()).Methods(http.MethodGet)
	accountRouter.HandleFunc("/info", hc.handleAccountInfoUpdate()).Methods(http.MethodPost)
	accountRouter.HandleFunc("/password", hc.handleAccountPasswordUpdate()).Methods(http.MethodPost)

	accountRouter.Use(func(h http.Handler) http.Handler {
		return authenticatedUser(h.ServeHTTP)
	})
}

// handleAccountView renders the user account management page.
func (hc *accountHandlerContext) handleAccountView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := userValue(r.Context())
		viewData := Data{
			Content: user,
		}

		hc.accountView.render(w, r, viewData)
	}
}

// handleAccountInfoUpdate processes the account information update form.
func (hc *accountHandlerContext) handleAccountInfoUpdate() func(w http.ResponseWriter, r *http.Request) {
	type infoUpdateForm struct {
		Email       string `schema:"email"`
		NickName    string `schema:"nick_name"`
		DisplayName string `schema:"display_name"`
	}

	var form infoUpdateForm

	return func(w http.ResponseWriter, r *http.Request) {
		ctxUser := userValue(r.Context())

		if err := parseForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse account information update form")
			PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		userInfo := user.InfoUpdate{
			UUID:        ctxUser.UUID,
			Email:       form.Email,
			NickName:    form.NickName,
			DisplayName: form.DisplayName,
		}

		if err := hc.userService.UpdateInfo(userInfo); err != nil {
			log.Error().Err(err).Msg("failed to update account information")
			PutFlashError(w, "There was an error updating your information")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		PutFlashSuccess(w, "Your account information has been successfully updated")
		http.Redirect(w, r, "/account", http.StatusSeeOther)
	}
}

// handleAccountPasswordUpdate processes the user account password update form.
func (hc *accountHandlerContext) handleAccountPasswordUpdate() func(w http.ResponseWriter, r *http.Request) {
	type passwordUpdateForm struct {
		CurrentPassword         string `schema:"current_password"`
		NewPassword             string `schema:"new_password"`
		NewPasswordConfirmation string `schema:"new_password_confirmation"`
	}

	var form passwordUpdateForm

	return func(w http.ResponseWriter, r *http.Request) {
		ctxUser := userValue(r.Context())

		if err := parseForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse account password update form")
			PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, "/account", http.StatusSeeOther)
			return
		}

		userPassword := user.PasswordUpdate{
			UUID:                    ctxUser.UUID,
			CurrentPassword:         form.CurrentPassword,
			NewPassword:             form.NewPassword,
			NewPasswordConfirmation: form.NewPasswordConfirmation,
		}

		if err := hc.userService.UpdatePassword(userPassword); err != nil {
			log.Error().Err(err).Msg("failed to update account password")
			PutFlashError(w, fmt.Sprintf("There was an error updating your password: %s", err))
			http.Redirect(w, r, "/account", http.StatusSeeOther)
			return
		}

		PutFlashSuccess(w, "Your account password has been successfully updated")
		http.Redirect(w, r, "/account", http.StatusSeeOther)
	}
}
