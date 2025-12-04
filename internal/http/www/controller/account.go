// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package controller

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/virtualtam/sparklemuffin/internal/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/internal/http/www/middleware"
	"github.com/virtualtam/sparklemuffin/internal/http/www/view"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

// RegisterAccountHandlers registers handlers for user account management..
func RegisterAccountHandlers(
	r *chi.Mux,
	feedService *feed.Service,
	userService *user.Service,
) {
	ac := accountController{
		feedService: feedService,
		userService: userService,

		accountInfoView:        view.New("account/info.gohtml"),
		accountPasswordView:    view.New("account/password.gohtml"),
		accountPreferencesView: view.New("account/preferences.gohtml"),
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
		r.Get("/preferences", ac.handlePreferencesView())
		r.Post("/preferences", ac.handlePreferencesUpdate())
	})
}

type accountController struct {
	feedService *feed.Service
	userService *user.Service

	accountInfoView        *view.View
	accountPasswordView    *view.View
	accountPreferencesView *view.View
}

// handleInfoUpdate processes the account information update form.
func (ac *accountController) handleInfoUpdate() func(w http.ResponseWriter, r *http.Request) {
	type infoUpdateForm struct {
		Email       string `schema:"email"`
		NickName    string `schema:"nick_name"`
		DisplayName string `schema:"display_name"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)

		var form infoUpdateForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse account information update form")
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

		if err := ac.userService.UpdateInfo(ctx, userInfo); err != nil {
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

		viewData := view.Data{
			Title:   "Account Information",
			Content: ctxUser,
		}

		ac.accountInfoView.Render(w, r, viewData)
	}
}

// handlePasswordUpdate processes the user account password update form.
func (ac *accountController) handlePasswordUpdate() func(w http.ResponseWriter, r *http.Request) {
	type passwordUpdateForm struct {
		CurrentPassword         string `schema:"current_password"`
		NewPassword             string `schema:"new_password"`
		NewPasswordConfirmation string `schema:"new_password_confirmation"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)

		var form passwordUpdateForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse account password update form")
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

		if err := ac.userService.UpdatePassword(ctx, userPassword); err != nil {
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

		viewData := view.Data{
			Title:   "Account Password",
			Content: ctxUser,
		}

		ac.accountPasswordView.Render(w, r, viewData)
	}
}
func (ac *accountController) handlePreferencesUpdate() func(w http.ResponseWriter, r *http.Request) {
	type preferencesUpdateForm struct {
		ShowEntries        string `schema:"feed_show_entries"`
		ShowEntrySummaries bool   `schema:"feed_show_entry_summaries"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)

		var form preferencesUpdateForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse account preferences update form")
			view.PutFlashError(w, "There was an error processing the form")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		feedPreferences := feed.Preferences{
			UserUUID:           ctxUser.UUID,
			ShowEntries:        feed.EntryVisibility(form.ShowEntries),
			ShowEntrySummaries: form.ShowEntrySummaries,
		}

		if err := ac.feedService.UpdatePreferences(ctx, feedPreferences); err != nil {
			log.Error().Err(err).Msg("failed to update account preferences")
			view.PutFlashError(w, fmt.Sprintf("There was an error updating your preferences: %s", err))
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		view.PutFlashSuccess(w, "Your feed preferences have been successfully updated")
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
	}
}

func (ac *accountController) handlePreferencesView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctxUser := httpcontext.UserValue(ctx)

		preferences, err := ac.feedService.PreferencesByUserUUID(ctx, ctxUser.UUID)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve account preferences")
			view.PutFlashError(w, "There was an error retrieving your preferences")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		viewData := view.Data{
			Title:   "Preferences",
			Content: preferences,
		}

		ac.accountPreferencesView.Render(w, r, viewData)
	}
}
