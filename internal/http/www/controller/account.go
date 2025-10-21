// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package controller

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/virtualtam/sparklemuffin/internal/http/www/csrf"
	"github.com/virtualtam/sparklemuffin/internal/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/internal/http/www/middleware"
	"github.com/virtualtam/sparklemuffin/internal/http/www/view"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

const (
	actionAccountUpdate string = "account-update"
)

// RegisterAccountHandlers registers handlers for user account management..
func RegisterAccountHandlers(
	r *chi.Mux,
	csrfService *csrf.Service,
	userService *user.Service,
) {
	ac := accountController{
		csrfService: csrfService,
		userService: userService,

		accountInfoView:     view.New("account/info.gohtml"),
		accountPasswordView: view.New("account/password.gohtml"),
	}

	// user account
	r.Route("/account", func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return middleware.AuthenticatedUser(h.ServeHTTP)
		})

		r.Get("/info", ac.handleInfoView())
		r.Post("/info", ac.handleInfoUpdate())
		r.Get("/password", ac.handlePasswordView())
		r.Post("/password", ac.handlePasswordUpdate())
	})
}

type accountController struct {
	csrfService *csrf.Service
	userService *user.Service

	accountInfoView     *view.View
	accountPasswordView *view.View
}

// handleInfoUpdate processes the account information update form.
func (ac *accountController) handleInfoUpdate() func(w http.ResponseWriter, r *http.Request) {
	type infoUpdateForm struct {
		CSRFToken   string `schema:"csrf_token"`
		Email       string `schema:"email"`
		NickName    string `schema:"nick_name"`
		DisplayName string `schema:"display_name"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctxUser := httpcontext.UserValue(r.Context())

		var form infoUpdateForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse account information update form")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if !ac.csrfService.Validate(form.CSRFToken, ctxUser.UUID, actionAccountUpdate) {
			log.Warn().Msg("failed to validate CSRF token")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		userInfo := user.InfoUpdate{
			UUID:        ctxUser.UUID,
			Email:       form.Email,
			NickName:    form.NickName,
			DisplayName: form.DisplayName,
		}

		if err := ac.userService.UpdateInfo(userInfo); err != nil {
			log.Error().Err(err).Msg("failed to update account information")
			view.PutFlashError(w, "There was an error updating your information")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		view.PutFlashSuccess(w, "Your account information has been successfully updated")
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
	}
}

// handleInfoView renders the user account information page.
func (ac *accountController) handleInfoView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctxUser := httpcontext.UserValue(r.Context())
		csrfToken := ac.csrfService.Generate(ctxUser.UUID, actionAccountUpdate)

		viewData := view.Data{
			Content: view.FormContent{
				CSRFToken: csrfToken,
				Content:   ctxUser,
			},
			Title: "Account Information",
		}

		ac.accountInfoView.Render(w, r, viewData)
	}
}

// handlePasswordUpdate processes the user account password update form.
func (ac *accountController) handlePasswordUpdate() func(w http.ResponseWriter, r *http.Request) {
	type passwordUpdateForm struct {
		CSRFToken               string `schema:"csrf_token"`
		CurrentPassword         string `schema:"current_password"`
		NewPassword             string `schema:"new_password"`
		NewPasswordConfirmation string `schema:"new_password_confirmation"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctxUser := httpcontext.UserValue(r.Context())

		var form passwordUpdateForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse account password update form")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		if !ac.csrfService.Validate(form.CSRFToken, ctxUser.UUID, actionAccountUpdate) {
			log.Warn().Msg("failed to validate CSRF token")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		userPassword := user.PasswordUpdate{
			UUID:                    ctxUser.UUID,
			CurrentPassword:         form.CurrentPassword,
			NewPassword:             form.NewPassword,
			NewPasswordConfirmation: form.NewPasswordConfirmation,
		}

		if err := ac.userService.UpdatePassword(userPassword); err != nil {
			log.Error().Err(err).Msg("failed to update account password")
			view.PutFlashError(w, fmt.Sprintf("There was an error updating your password: %s", err))
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		view.PutFlashSuccess(w, "Your account password has been successfully updated")
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
	}
}

// handlePasswordView renders the user account password page.
func (ac *accountController) handlePasswordView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctxUser := httpcontext.UserValue(r.Context())
		csrfToken := ac.csrfService.Generate(ctxUser.UUID, actionAccountUpdate)

		viewData := view.Data{
			Content: view.FormContent{
				CSRFToken: csrfToken,
				Content:   ctxUser,
			},
			Title: "Account Password",
		}

		ac.accountPasswordView.Render(w, r, viewData)
	}
}
