// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package www

import (
	"net/url"

	"github.com/prometheus/client_golang/prometheus"

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

// OptionFunc represents a function that configures a set of options for a Server.
type OptionFunc func(*Server) error

// WithMetricsRegistry sets the Prometheus metrics registry used to expose application metrics.
func WithMetricsRegistry(prefix string, registry *prometheus.Registry) OptionFunc {
	return func(s *Server) error {
		if prefix == "" {
			return ErrServerMetricsPrefixRequired
		}
		if registry == nil {
			return ErrServerMetricsRegistryRequired
		}

		s.metricsPrefix = prefix
		s.metricsRegistry = registry
		return nil
	}
}

// WithPublicURL sets the public URL the application is accessible at.
//
// This URL is used to generate absolute URLs for permalinks and feeds.
func WithPublicURL(publicURL *url.URL) OptionFunc {
	return func(s *Server) error {
		if publicURL == nil {
			return ErrServerPublicURLRequired
		}

		s.publicURL = publicURL
		return nil
	}
}

// WithBookmarkServices sets the bookmark management services.
func WithBookmarkServices(
	bookmarkService *bookmark.Service,
	exportingService *bookmarkexporting.Service,
	importingService *bookmarkimporting.Service,
	queryingService *bookmarkquerying.Service,
) OptionFunc {
	return func(s *Server) error {
		if bookmarkService == nil {
			return ErrServerBookmarkServiceRequired
		}
		if exportingService == nil {
			return ErrServerBookmarkExportingServiceRequired
		}
		if importingService == nil {
			return ErrServerBookmarkImportingServiceRequired
		}
		if queryingService == nil {
			return ErrServerBookmarkQueryingServiceRequired
		}

		s.bookmarkService = bookmarkService
		s.bookmarkExportingService = exportingService
		s.bookmarkImportingService = importingService
		s.bookmarkQueryingService = queryingService
		return nil
	}
}

// WithFeedServices sets the feed management services.
func WithFeedServices(
	feedService *feed.Service,
	feedExportingService *feedexporting.Service,
	feedImportingService *feedimporting.Service,
	feedQueryingService *feedquerying.Service,
) OptionFunc {
	return func(s *Server) error {
		if feedService == nil {
			return ErrServerFeedServiceRequired
		}
		if feedExportingService == nil {
			return ErrServerFeedExportingServiceRequired
		}
		if feedImportingService == nil {
			return ErrServerFeedImportingServiceRequired
		}
		if feedQueryingService == nil {
			return ErrServerFeedQueryingServiceRequired
		}

		s.feedService = feedService
		s.feedExportingService = feedExportingService
		s.feedImportingService = feedImportingService
		s.feedQueryingService = feedQueryingService
		return nil
	}
}

// WithSessionService sets the user session management service.
func WithSessionService(sessionService *session.Service) OptionFunc {
	return func(s *Server) error {
		if sessionService == nil {
			return ErrServerSessionServiceRequired
		}

		s.sessionService = sessionService
		return nil
	}
}

// WithUserService sets the user management service.
func WithUserService(userService *user.Service) OptionFunc {
	return func(s *Server) error {
		if userService == nil {
			return ErrServerUserServiceRequired
		}

		s.userService = userService
		return nil
	}
}
