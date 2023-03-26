package www

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/virtualtam/yawbe/internal/rand"
	"github.com/virtualtam/yawbe/pkg/session"
	"github.com/virtualtam/yawbe/pkg/user"
)

const (
	UserRememberTokenNBytes     int    = 32
	UserRememberTokenCookieName string = "remember_me"
)

type sessionHandlerContext struct {
	sessionService *session.Service
	userService    *user.Service

	userLoginView *view
}

func registerSessionHandlers(
	r *mux.Router,
	sessionService *session.Service,
	userService *user.Service,
) {
	hc := sessionHandlerContext{
		sessionService: sessionService,
		userService:    userService,

		userLoginView: newView("session/login.gohtml"),
	}

	// authentication
	r.HandleFunc("/login", hc.userLoginView.handle).Methods(http.MethodGet)
	r.HandleFunc("/login", hc.handleUserLogin()).Methods(http.MethodPost)
	r.HandleFunc("/logout", hc.handleUserLogout()).Methods(http.MethodPost)
}

// handleUserLogin processes data submitted through the user login form.
func (hc *sessionHandlerContext) handleUserLogin() func(w http.ResponseWriter, r *http.Request) {
	type loginForm struct {
		Email    string `schema:"email"`
		Password string `schema:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var form loginForm

		if err := parseForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse login form")
			PutFlashError(w, err.Error())
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		user, err := hc.userService.Authenticate(form.Email, form.Password)
		if err != nil {
			log.Error().Err(err).Msg("failed to authenticate user")
			PutFlashError(w, "invalid email or password")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		if err := hc.setUserRememberToken(w, user.UUID); err != nil {
			log.Error().Err(err).Msg("failed to set remember token")
			PutFlashError(w, "failed to save session cookie")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
	}
}

// handleUserLogout logs a user out and clears their session data.
func (hc *sessionHandlerContext) handleUserLogout() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie := http.Cookie{
			Name:     UserRememberTokenCookieName,
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 1),
			HttpOnly: true,
		}
		http.SetCookie(w, &cookie)

		user := userValue(r.Context())
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

		err = hc.sessionService.Add(userSession)
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
func (hc *sessionHandlerContext) setUserRememberToken(w http.ResponseWriter, userUUID string) error {
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

	err = hc.sessionService.Add(userSession)
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
