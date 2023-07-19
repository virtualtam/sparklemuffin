package www

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"

	"github.com/virtualtam/sparklemuffin/pkg/http/www/csrf"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

const (
	actionAccountUpdate string = "account-update"
)

type accountHandlerContext struct {
	csrfService *csrf.Service
	userService *user.Service

	accountView *view
}

func registerAccounthandlers(
	r *chi.Mux,
	csrfService *csrf.Service,
	userService *user.Service,
) {
	hc := accountHandlerContext{
		csrfService: csrfService,
		userService: userService,

		accountView: newView("account/account.gohtml"),
	}

	// user account
	r.Route("/account", func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return authenticatedUser(h.ServeHTTP)
		})

		r.Get("/", hc.handleAccountView())
		r.Post("/info", hc.handleAccountInfoUpdate())
		r.Post("/password", hc.handleAccountPasswordUpdate())
	})
}

// handleAccountView renders the user account management page.
func (hc *accountHandlerContext) handleAccountView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := userValue(r.Context())
		csrfToken := hc.csrfService.Generate(user.UUID, actionAccountUpdate)

		viewData := Data{
			Content: FormContent{
				CSRFToken: csrfToken,
				Content:   user,
			},
			Title: "Account",
		}

		hc.accountView.render(w, r, viewData)
	}
}

// handleAccountInfoUpdate processes the account information update form.
func (hc *accountHandlerContext) handleAccountInfoUpdate() func(w http.ResponseWriter, r *http.Request) {
	type infoUpdateForm struct {
		CSRFToken   string `form:"csrf_token"`
		Email       string `form:"email"`
		NickName    string `form:"nick_name"`
		DisplayName string `form:"display_name"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctxUser := userValue(r.Context())

		var form infoUpdateForm
		if err := render.DecodeForm(r.Body, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse account information update form")
			PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if !hc.csrfService.Validate(form.CSRFToken, ctxUser.UUID, actionAccountUpdate) {
			log.Warn().Msg("failed to validate CSRF token")
			PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
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
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		PutFlashSuccess(w, "Your account information has been successfully updated")
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
	}
}

// handleAccountPasswordUpdate processes the user account password update form.
func (hc *accountHandlerContext) handleAccountPasswordUpdate() func(w http.ResponseWriter, r *http.Request) {
	type passwordUpdateForm struct {
		CSRFToken               string `form:"csrf_token"`
		CurrentPassword         string `form:"current_password"`
		NewPassword             string `form:"new_password"`
		NewPasswordConfirmation string `form:"new_password_confirmation"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctxUser := userValue(r.Context())

		var form passwordUpdateForm
		if err := render.DecodeForm(r.Body, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse account password update form")
			PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if !hc.csrfService.Validate(form.CSRFToken, ctxUser.UUID, actionAccountUpdate) {
			log.Warn().Msg("failed to validate CSRF token")
			PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
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
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		PutFlashSuccess(w, "Your account password has been successfully updated")
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
	}
}
