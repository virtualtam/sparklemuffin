// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package www

import (
	"net/url"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/csrf"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	bookmarkexporting "github.com/virtualtam/sparklemuffin/pkg/bookmark/exporting"
	bookmarkimporting "github.com/virtualtam/sparklemuffin/pkg/bookmark/importing"
	bookmarkquerying "github.com/virtualtam/sparklemuffin/pkg/bookmark/querying"
	"github.com/virtualtam/sparklemuffin/pkg/session"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

func WithCSRFKey(csrfKey string) optionFunc {
	return func(s *Server) {
		s.csrfService = csrf.NewService(csrfKey)
	}
}

func WithMetricsRegistry(prefix string, registry *prometheus.Registry) optionFunc {
	return func(s *Server) {
		s.metricsPrefix = prefix
		s.metricsRegistry = registry
	}
}

func WithPublicURL(publicURL *url.URL) optionFunc {
	return func(s *Server) {
		s.publicURL = publicURL
	}
}

func WithBookmarkServices(
	bookmarkService *bookmark.Service,
	exportingService *bookmarkexporting.Service,
	importingService *bookmarkimporting.Service,
	queryingService *bookmarkquerying.Service,
) optionFunc {
	return func(s *Server) {
		s.bookmarkService = bookmarkService
		s.bookmarkExportingService = exportingService
		s.bookmarkImportingService = importingService
		s.bookmarkQueryingService = queryingService
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
