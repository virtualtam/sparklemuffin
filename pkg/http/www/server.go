package www

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/virtualtam/yawbe/pkg/http/www/rand"
	"github.com/virtualtam/yawbe/pkg/http/www/static"
	"github.com/virtualtam/yawbe/pkg/user"
)

var _ http.Handler = &Server{}

// Server represents the Web service.
type Server struct {
	router      *mux.Router
	userService *user.Service

	homeView      *view
	userLoginView *view
}

// NewServer initializes and returns a new Server.
func NewServer(userService *user.Service) *Server {
	s := &Server{
		router:      mux.NewRouter(),
		userService: userService,

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
	s.router.HandleFunc("/", s.homeView.handle)

	// authentication
	s.router.HandleFunc("/login", s.userLoginView.handle).Methods("GET")
	s.router.HandleFunc("/login", s.handleUserLogin()).Methods("POST")

	// static assets
	s.router.HandleFunc("/static/", http.NotFound)
	s.router.PathPrefix("/static/").Handler(http.StripPrefix(
		"/static/",
		s.staticCacheControl(http.FileServer(http.FS(static.FS)))))
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
			viewData.AlertError(err)
			s.userLoginView.render(w, r, viewData)
			return
		}

		user, err := s.userService.Authenticate(form.Email, form.Password)
		if err != nil {
			viewData.AlertError(err)
			s.userLoginView.render(w, r, viewData)
			return
		}

		if err := s.setUserRememberToken(w, &user); err != nil {
			viewData.AlertError(err)
			s.userLoginView.render(w, r, viewData)
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
			return err
		}

		user.RememberToken = token
		err = s.userService.Update(*user)
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

// staticCacheControl sets the Cache-Control header for static assets.
func (s *Server) staticCacheControl(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=2592000") // 30 days
		h.ServeHTTP(w, r)
	})
}
