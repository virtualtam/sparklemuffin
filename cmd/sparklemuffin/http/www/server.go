// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package www

import (
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
	slokmetrics "github.com/slok/go-http-metrics/metrics/prometheus"
	slokmiddleware "github.com/slok/go-http-metrics/middleware"
	slokstd "github.com/slok/go-http-metrics/middleware/std"

	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/controller"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/csrf"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/middleware"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/static"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/view"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	bookmarkexporting "github.com/virtualtam/sparklemuffin/pkg/bookmark/exporting"
	bookmarkimporting "github.com/virtualtam/sparklemuffin/pkg/bookmark/importing"
	bookmarkquerying "github.com/virtualtam/sparklemuffin/pkg/bookmark/querying"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	feedquerying "github.com/virtualtam/sparklemuffin/pkg/feed/querying"
	"github.com/virtualtam/sparklemuffin/pkg/session"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

var _ http.Handler = &Server{}

// Server represents the Web service.
type Server struct {
	router    *chi.Mux
	publicURL *url.URL

	metricsPrefix   string
	metricsRegistry *prometheus.Registry

	// Bookmark services
	bookmarkService          *bookmark.Service
	bookmarkExportingService *bookmarkexporting.Service
	bookmarkImportingService *bookmarkimporting.Service
	bookmarkQueryingService  *bookmarkquerying.Service

	// Feed services
	feedService         *feed.Service
	feedQueryingService *feedquerying.Service

	// User and session management services
	csrfService    *csrf.Service
	sessionService *session.Service
	userService    *user.Service

	homeView *view.View
}

type optionFunc func(*Server)

// NewServer initializes and returns a new Server.
func NewServer(optionFuncs ...optionFunc) *Server {
	s := &Server{
		router: chi.NewRouter(),

		homeView: view.New("static/home.gohtml"),
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
	s.router.Use(chimiddleware.RequestID)
	s.router.Use(chimiddleware.RealIP)

	if s.metricsRegistry != nil {
		prometheusMiddleware := slokmiddleware.New(
			slokmiddleware.Config{
				Recorder: slokmetrics.NewRecorder(
					slokmetrics.Config{
						Prefix:   s.metricsPrefix,
						Registry: s.metricsRegistry,
					},
				),
			},
		)
		s.router.Use(slokstd.HandlerProvider("", prometheusMiddleware))
	}

	// Structured logging
	s.router.Use(hlog.NewHandler(log.Logger), hlog.AccessHandler(middleware.AccessLogger))

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
				middleware.StaticCacheControl(
					http.FileServer(http.FS(static.FS)),
				),
			),
		)
	})

	// Register domain handlers
	controller.RegisterSessionHandlers(s.router, s.sessionService, s.userService)
	controller.RegisterAdminHandlers(s.router, s.csrfService, s.userService)
	controller.RegisterAccounthandlers(s.router, s.csrfService, s.userService)
	controller.RegisterBookmarkHandlers(s.router, s.publicURL, s.bookmarkService, s.csrfService, s.bookmarkQueryingService, s.userService)
	controller.RegisterFeedHandlers(s.router, s.feedService, s.feedQueryingService, s.userService)
	controller.RegisterToolsHandlers(s.router, s.bookmarkExportingService, s.bookmarkImportingService)
}

// ServeHTTP satisfies the http.Handler interface,
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// handleHomeView renders the application's home page.
func (s *Server) handleHomeView() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		viewData := view.Data{Title: "Home"}
		s.homeView.Render(w, r, viewData)
	}
}

// rememberUser enriches the request context with a user.User if a valid
// remember token cookie is set.
func (s *Server) rememberUser(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(controller.UserRememberTokenCookieName)
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
		ctx = httpcontext.WithUser(ctx, user)
		r = r.WithContext(ctx)

		h(w, r)
	})
}
