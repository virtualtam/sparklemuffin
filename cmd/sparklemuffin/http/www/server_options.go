package www

import (
	"net/url"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www/csrf"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark/exporting"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark/importing"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark/querying"
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
