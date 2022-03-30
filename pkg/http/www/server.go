package www

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/virtualtam/yawbe/pkg/http/www/rand"
	"github.com/virtualtam/yawbe/pkg/http/www/static"
	"github.com/virtualtam/yawbe/pkg/user"
)

var _ http.Handler = &Server{}

// Server represents the Web service.
type Server struct {
	router      *mux.Router
	userService *user.Service

	adminView     *view
	homeView      *view
	userLoginView *view
}

// NewServer initializes and returns a new Server.
func NewServer(userService *user.Service) *Server {
	s := &Server{
		router:      mux.NewRouter(),
		userService: userService,

		adminView:     newView("admin/admin.gohtml"),
		homeView:      newView("static/home.gohtml"),
		userLoginView: newView("user/login.gohtml"),
	}

	s.addRoutes()

	return s
}

// ServeHTTP satisfies the http.Handler interface,
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// addRoutes registers all HTTP handlers for the Web interface.
func (s *Server) addRoutes() {
	// static pages
	s.router.HandleFunc("/", s.rememberUser(s.homeView.handle))

	// administration
	s.router.HandleFunc("/admin", s.rememberUser(s.requireAdminUser(s.handleAdmin()))).Methods("GET")

	// authentication
	s.router.HandleFunc("/login", s.rememberUser(s.userLoginView.handle)).Methods("GET")
	s.router.HandleFunc("/login", s.rememberUser(s.handleUserLogin())).Methods("POST")
	s.router.HandleFunc("/logout", s.rememberUser(s.handleUserLogout())).Methods("POST")

	// static assets
	s.router.HandleFunc("/static/", http.NotFound)
	s.router.PathPrefix("/static/").Handler(http.StripPrefix(
		"/static/",
		s.staticCacheControl(http.FileServer(http.FS(static.FS)))))
}

// handleAdmin displays the main administration page.
func (s *Server) handleAdmin() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		s.adminView.handle(w, r)
	}
}

// handleUserLogin processes data submitted through the user login form.
func (s *Server) handleUserLogin() func(w http.ResponseWriter, r *http.Request) {
	var viewData Data

	type loginForm struct {
		Email    string `schema:"email"`
		Password string `schema:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var form loginForm

		if err := parseForm(r, &form); err != nil {
			log.Error().Err(err).Msg("failed to parse login form")
			viewData.AlertError(err)
			s.userLoginView.render(w, r, viewData)
			return
		}

		user, err := s.userService.Authenticate(form.Email, form.Password)
		if err != nil {
			log.Error().Err(err).Msg("failed to authenticate user")
			viewData.AlertError(err)
			s.userLoginView.render(w, r, viewData)
			return
		}

		if err := s.setUserRememberToken(w, &user); err != nil {
			log.Error().Err(err).Msg("failed to set remember token")
			viewData.AlertError(err)
			s.userLoginView.render(w, r, viewData)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

// handleUserLogout logs a user out and clears their session data.
func (s *Server) handleUserLogout() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie := http.Cookie{
			Name:     UserRememberTokenCookieName,
			Value:    "",
			Expires:  time.Now(),
			HttpOnly: true,
		}
		http.SetCookie(w, &cookie)

		user := userValue(r.Context())
		if user == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		token, err := rand.RandomBase64URLString(UserRememberTokenNBytes)
		if err != nil {
			log.Error().Err(err).Msg("failed to generate a remember token")
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		user.RememberToken = token
		err = s.userService.Update(*user)
		if err != nil {
			log.Error().Err(err).Msg("failed to update user")
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

const (
	UserRememberTokenNBytes     int    = 32
	UserRememberTokenCookieName string = "remember_me"
)

// setUserRememberToken creates and persists a new RememberToken if needed, and
// sets it as a session cookie.
func (s *Server) setUserRememberToken(w http.ResponseWriter, user *user.User) error {
	if user.RememberToken == "" {
		token, err := rand.RandomBase64URLString(UserRememberTokenNBytes)
		if err != nil {
			log.Error().Err(err).Msg("failed to generate a remember token")
			return err
		}

		user.RememberToken = token
		err = s.userService.Update(*user)
		if err != nil {
			log.Error().Err(err).Msg("failed to update user")
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

		user, err := s.userService.ByRememberToken(cookie.Value)
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

// requireAdminUser requires the user to have administration privileges to
// access content.
func (s *Server) requireAdminUser(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := userValue(r.Context())

		if user == nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		if !user.IsAdmin {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		h(w, r)
	})
}
