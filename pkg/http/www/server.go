package www

import (
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"

	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	"github.com/virtualtam/sparklemuffin/pkg/exporting"
	"github.com/virtualtam/sparklemuffin/pkg/http/www/static"
	"github.com/virtualtam/sparklemuffin/pkg/importing"
	"github.com/virtualtam/sparklemuffin/pkg/querying"
	"github.com/virtualtam/sparklemuffin/pkg/session"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

var _ http.Handler = &Server{}

// Server represents the Web service.
type Server struct {
	router    *chi.Mux
	publicURL *url.URL

	bookmarkService  *bookmark.Service
	exportingService *exporting.Service
	importingService *importing.Service
	queryingService  *querying.Service
	sessionService   *session.Service
	userService      *user.Service

	homeView *view
}

type optionFunc func(*Server)

// NewServer initializes and returns a new Server.
func NewServer(optionFuncs ...optionFunc) *Server {
	s := &Server{
		router: chi.NewRouter(),

		homeView: newView("static/home.gohtml"),
	}

	for _, optionFunc := range optionFuncs {
		optionFunc(s)
	}

	s.registerHandlers()

	return s
}

// registerHandlers registers all HTTP handlers for the Web interface.
func (s *Server) registerHandlers() {
	// Global middleware
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)

	// Structured logging
	s.router.Use(hlog.NewHandler(log.Logger), hlog.AccessHandler(accessLogger))

	s.router.Use(func(h http.Handler) http.Handler {
		return s.rememberUser(h.ServeHTTP)
	})

	// Static pages
	s.router.Get("/", s.handleHomeView())

	// Static assets
	s.router.Route("/static", func(r chi.Router) {
		r.Get("/", http.NotFound)

		r.Handle(
			"/*",
			http.StripPrefix(
				"/static/",
				staticCacheControl(
					http.FileServer(http.FS(static.FS)),
				),
			),
		)
	})

	// Register domain handlers
	registerSessionHandlers(s.router, s.sessionService, s.userService)
	registerAdminHandlers(s.router, s.userService)
	registerAccounthandlers(s.router, s.userService)
	registerBookmarkHandlers(s.router, s.publicURL, s.bookmarkService, s.queryingService, s.userService)
	registerTagHandlers(s.router, s.bookmarkService, s.queryingService)
	registerToolsHandlers(s.router, s.exportingService, s.importingService)
}

// ServeHTTP satisfies the http.Handler interface,
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// handleHomeView renders the application's home page.
func (s *Server) handleHomeView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		viewData := Data{Title: "Home"}
		s.homeView.render(w, r, viewData)
	}
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
