// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

// Package www serves the Web application.
package www

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
	slokmetrics "github.com/slok/go-http-metrics/metrics/prometheus"
	slokmiddleware "github.com/slok/go-http-metrics/middleware"
	slokstd "github.com/slok/go-http-metrics/middleware/std"

	"github.com/virtualtam/sparklemuffin/internal/http/www/controller"
	"github.com/virtualtam/sparklemuffin/internal/http/www/csrf"
	"github.com/virtualtam/sparklemuffin/internal/http/www/httpcontext"
	"github.com/virtualtam/sparklemuffin/internal/http/www/middleware"
	"github.com/virtualtam/sparklemuffin/internal/http/www/static"
	"github.com/virtualtam/sparklemuffin/internal/http/www/view"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	bookmarkexporting "github.com/virtualtam/sparklemuffin/pkg/bookmark/exporting"
	bookmarkimporting "github.com/virtualtam/sparklemuffin/pkg/bookmark/importing"
	bookmarkquerying "github.com/virtualtam/sparklemuffin/pkg/bookmark/querying"
	"github.com/virtualtam/sparklemuffin/pkg/feed"
	feedexporting "github.com/virtualtam/sparklemuffin/pkg/feed/exporting"
	feedimporting "github.com/virtualtam/sparklemuffin/pkg/feed/importing"
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
	feedService          *feed.Service
	feedExportingService *feedexporting.Service
	feedImportingService *feedimporting.Service
	feedQueryingService  *feedquerying.Service

	// User and session management services
	csrfService    *csrf.Service
	sessionService *session.Service
	userService    *user.Service

	homeView  *view.View
	errorView *view.ErrorView
}

// NewServer initializes and returns a new Server.
func NewServer(optionFuncs ...OptionFunc) (*Server, error) {
	s := &Server{
		router: chi.NewRouter(),

		homeView:  view.New("page/home.gohtml"),
		errorView: view.NewError(),
	}

	for _, optionFunc := range optionFuncs {
		if err := optionFunc(s); err != nil {
			return nil, err
		}
	}

	s.registerHandlers()

	return s, nil
}

// registerHandlers registers all HTTP handlers for the Web application.
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

	// Pages
	s.router.Get("/", s.handleHomeView())

	// Static pages
	s.router.Get("/robots.txt", s.handleRobotsTxtView())

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

	// Domain handlers
	controller.RegisterSessionHandlers(s.router, s.sessionService, s.userService)
	controller.RegisterAdminHandlers(s.router, s.csrfService, s.userService)
	controller.RegisterAccountHandlers(s.router, s.csrfService, s.feedService, s.userService)
	controller.RegisterBookmarkHandlers(s.router, s.publicURL, s.bookmarkService, s.csrfService, s.bookmarkExportingService, s.bookmarkImportingService, s.bookmarkQueryingService, s.userService)
	controller.RegisterFeedHandlers(s.router, s.csrfService, s.feedService, s.feedExportingService, s.feedImportingService, s.feedQueryingService, s.userService)

	// 404 handler
	s.router.NotFound(s.handleNotFound())
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// handleHomeView renders the application's home page.
func (s *Server) handleHomeView() func(w http.ResponseWriter, r *http.Request) {
	const title = "Home"
	defaultViewData := view.Data{
		Title:   title,
		Content: "You are currently logged out.",
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctxUser := httpcontext.UserValue(r.Context())

		if ctxUser == nil {
			s.homeView.Render(w, r, defaultViewData)
			return
		}

		viewData := view.Data{
			Title:   title,
			Content: fmt.Sprintf("Welcome back, %s!", ctxUser.NickName),
		}

		s.homeView.Render(w, r, viewData)
	}
}

// handleRobotsTxtView renders the application's robots.txt file.
//
// As user content requires authentication, we indicate that:
// - crawlers should not index the site;
// - AI bots should not index the site, nor use its content for training.
//
// See:
// - https://robotstxt.com/
// - https://robotstxt.com/ai
// - https://intoli.com/blog/analyzing-one-million-robots-txt-files/
func (s *Server) handleRobotsTxtView() func(w http.ResponseWriter, r *http.Request) {
	var robotsTxt = []byte(`User-agent: *
Disallow: /
DisallowAITraining: /
Content-Usage: ai=n`)

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		// Skip error checking as the HTTP headers have already been sent
		_, _ = w.Write(robotsTxt) // nolint:errcheck
	}
}

// handleNotFound renders an HTTP 404 Not Found error page.
func (s *Server) handleNotFound() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		s.errorView.Render(w, r, http.StatusNotFound)
	}
}

// rememberUser enriches the request context with a user.User if a valid
// remember token cookie is set.
func (s *Server) rememberUser(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" || strings.HasPrefix(r.URL.Path, "/static") {
			// Skip user session middleware for static pages and assets.
			h(w, r)
			return
		}

		cookie, err := r.Cookie(controller.UserRememberTokenCookieName)
		if err != nil {
			h(w, r)
			return
		}

		userSession, err := s.sessionService.ByRememberToken(cookie.Value)
		if err != nil {
			h(w, r)
			return
		}

		ctxUser, err := s.userService.ByUUID(userSession.UserUUID)
		if err != nil {
			h(w, r)
			return
		}

		ctx := r.Context()
		ctx = httpcontext.WithUser(ctx, ctxUser)
		r = r.WithContext(ctx)

		h(w, r)
	}
}
