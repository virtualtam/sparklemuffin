// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package controller

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/virtualtam/sparklemuffin/internal/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/internal/http/www/view"
	"github.com/virtualtam/sparklemuffin/internal/rand"
	"github.com/virtualtam/sparklemuffin/pkg/session"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

const (
	UserRememberTokenNBytes     int    = 32
	UserRememberTokenCookieName string = "remember_me"
)

// RegisterSessionHandlers registers handlers for user session management..
func RegisterSessionHandlers(
	r *chi.Mux,
	sessionService *session.Service,
	userService *user.Service,
) {
	sc := sessionController{
		sessionService: sessionService,
		userService:    userService,

		userLoginView: view.New("session/login.gohtml"),
	}

	// authentication
	r.Get("/login", sc.userLoginView.Handle)
	r.Post("/login", sc.handleUserLogin())
	r.Post("/logout", sc.handleUserLogout())
}

type sessionController struct {
	sessionService *session.Service
	userService    *user.Service

	userLoginView *view.View
}

// handleUserLogin processes data submitted through the user login form.
func (sc *sessionController) handleUserLogin() func(w http.ResponseWriter, r *http.Request) {
	type loginForm struct {
		Email    string `schema:"email"`
		Password string `schema:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var form loginForm
		if err := decodeForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse login form")
			view.PutFlashError(w, err.Error())
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		user, err := sc.userService.Authenticate(form.Email, form.Password)
		if err != nil {
			log.Error().Err(err).Msg("failed to authenticate user")
			view.PutFlashError(w, "invalid email or password")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		if err := sc.setUserRememberToken(w, user.UUID); err != nil {
			log.Error().Err(err).Msg("failed to set remember token")
			view.PutFlashError(w, "failed to save session cookie")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
	}
}

// handleUserLogout logs a user out and clears their session data.
func (sc *sessionController) handleUserLogout() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie := http.Cookie{
			Name:     UserRememberTokenCookieName,
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 1),
			HttpOnly: true,
		}
		http.SetCookie(w, &cookie)

		user := httpcontext.UserValue(r.Context())
		if user == nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		token, err := rand.RandomBase64URLString(UserRememberTokenNBytes)
		if err != nil {
			log.Error().Err(err).Msg("failed to generate a remember token")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		userSession := session.Session{
			UserUUID:      user.UUID,
			RememberToken: token,
		}

		err = sc.sessionService.Add(userSession)
		if err != nil {
			log.Error().Err(err).Msg("failed to save user session")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// setUserRememberToken creates and persists a new RememberToken if needed, and
// sets it as a session cookie.
func (sc *sessionController) setUserRememberToken(w http.ResponseWriter, userUUID string) error {
	token, err := rand.RandomBase64URLString(UserRememberTokenNBytes)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate a remember token")
		return err
	}

	// expires after one month
	expiresAt := time.Now().UTC().AddDate(0, 1, 0)

	userSession := session.Session{
		UserUUID:               userUUID,
		RememberToken:          token,
		RememberTokenExpiresAt: expiresAt,
	}

	err = sc.sessionService.Add(userSession)
	if err != nil {
		log.Error().Err(err).Msg("failed to update user")
		return err
	}

	cookie := http.Cookie{
		Name:     UserRememberTokenCookieName,
		Value:    userSession.RememberToken,
		Expires:  expiresAt,
		Path:     "/",
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)

	return nil
}
