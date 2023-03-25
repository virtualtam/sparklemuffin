package www

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/virtualtam/yawbe/internal/rand"
	"github.com/virtualtam/yawbe/pkg/bookmark"
	"github.com/virtualtam/yawbe/pkg/exporting"
	"github.com/virtualtam/yawbe/pkg/http/www/static"
	"github.com/virtualtam/yawbe/pkg/importing"
	"github.com/virtualtam/yawbe/pkg/querying"
	"github.com/virtualtam/yawbe/pkg/session"
	"github.com/virtualtam/yawbe/pkg/user"
)

var _ http.Handler = &Server{}

// Server represents the Web service.
type Server struct {
	router *mux.Router

	bookmarkService  *bookmark.Service
	exportingService *exporting.Service
	importingService *importing.Service
	queryingService  *querying.Service
	sessionService   *session.Service
	userService      *user.Service

	homeView      *view
	userLoginView *view
}

type optionFunc func(*Server)

// NewServer initializes and returns a new Server.
func NewServer(optionFuncs ...optionFunc) *Server {
	s := &Server{
		router: mux.NewRouter(),

		homeView:      newView("static/home.gohtml"),
		userLoginView: newView("user/login.gohtml"),
	}

	for _, optionFunc := range optionFuncs {
		optionFunc(s)
	}

	s.addRoutes()

	return s
}

func WithBookmarkService(bookmarkService *bookmark.Service) optionFunc {
	return func(s *Server) {
		s.bookmarkService = bookmarkService
	}
}

func WithExportingService(exportingService *exporting.Service) optionFunc {
	return func(s *Server) {
		s.exportingService = exportingService
	}
}

func WithImportingService(importingService *importing.Service) optionFunc {
	return func(s *Server) {
		s.importingService = importingService
	}
}

func WithQueryingService(queryingService *querying.Service) optionFunc {
	return func(s *Server) {
		s.queryingService = queryingService
	}
}

func WithSessionService(sessionService *session.Service) optionFunc {
	return func(s *Server) {
		s.sessionService = sessionService
	}
}

func WithUserService(userService *user.Service) optionFunc {
	return func(s *Server) {
		s.userService = userService
	}
}

// ServeHTTP satisfies the http.Handler interface,
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// addRoutes registers all HTTP handlers for the Web interface.
func (s *Server) addRoutes() {
	// static pages
	s.router.HandleFunc("/", s.homeView.handle)

	// authentication
	s.router.HandleFunc("/login", s.userLoginView.handle).Methods(http.MethodGet)
	s.router.HandleFunc("/login", s.handleUserLogin()).Methods(http.MethodPost)
	s.router.HandleFunc("/logout", s.handleUserLogout()).Methods(http.MethodPost)

	setupAdminHandlers(s.router, s.userService)
	setupAccounthandlers(s.router, s.userService)
	setupBookmarkHandlers(s.router, s.bookmarkService, s.queryingService, s.userService)
	setupToolsHandlers(s.router, s.exportingService, s.importingService)

	// static assets
	s.router.HandleFunc("/static/", http.NotFound)
	s.router.PathPrefix("/static/").Handler(http.StripPrefix(
		"/static/",
		s.staticCacheControl(http.FileServer(http.FS(static.FS)))))

	// global middleware
	s.router.Use(func(h http.Handler) http.Handler {
		return s.rememberUser(h.ServeHTTP)
	})
}

// handleUserLogin processes data submitted through the user login form.
func (s *Server) handleUserLogin() func(w http.ResponseWriter, r *http.Request) {
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

		user, err := s.userService.Authenticate(form.Email, form.Password)
		if err != nil {
			log.Error().Err(err).Msg("failed to authenticate user")
			PutFlashError(w, "invalid email or password")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		if err := s.setUserRememberToken(w, user.UUID); err != nil {
			log.Error().Err(err).Msg("failed to set remember token")
			PutFlashError(w, "failed to save session cookie")
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/bookmarks", http.StatusSeeOther)
	}
}

// handleUserLogout logs a user out and clears their session data.
func (s *Server) handleUserLogout() func(w http.ResponseWriter, r *http.Request) {
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

		err = s.sessionService.Add(userSession)
		if err != nil {
			log.Error().Err(err).Msg("failed to save user session")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

const (
	UserRememberTokenNBytes     int    = 32
	UserRememberTokenCookieName string = "remember_me"
)

// setUserRememberToken creates and persists a new RememberToken if needed, and
// sets it as a session cookie.
func (s *Server) setUserRememberToken(w http.ResponseWriter, userUUID string) error {
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

	err = s.sessionService.Add(userSession)
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

// staticCacheControl sets the Cache-Control header for static assets.
func (s *Server) staticCacheControl(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=2592000") // 30 days
		h.ServeHTTP(w, r)
	})
}

// rememberUser enriches the request context with a user.User if a valid
// remember token cookie is set.
func (s *Server) rememberUser(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(UserRememberTokenCookieName)
		if err != nil {
			h(w, r)
			return
		}

		session, err := s.sessionService.ByRememberToken(cookie.Value)
		if err != nil {
			h(w, r)
			return
		}

		user, err := s.userService.ByUUID(session.UserUUID)
		if err != nil {
			h(w, r)
			return
		}

		ctx := r.Context()
		ctx = withUser(ctx, user)
		r = r.WithContext(ctx)

		h(w, r)
	})
}
