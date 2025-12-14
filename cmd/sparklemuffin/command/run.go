// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package command

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/earthboundkid/versioninfo/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/virtualtam/sparklemuffin/internal/http/monitoring"
	"github.com/virtualtam/sparklemuffin/internal/http/www"
	feedsynchronizing "github.com/virtualtam/sparklemuffin/pkg/feed/synchronizing"
)

// NewRunCommand initializes a CLI command to start the HTTP servers.
func NewRunCommand() *cobra.Command {
	const (
		defaultWebListenAddr        string = "0.0.0.0:8080"
		defaultPublicWebAddr        string = "http://localhost:8080"
		defaultMonitoringListenAddr string = "0.0.0.0:8090"
	)

	var (
		webListenAddr        string
		publicWebAddr        string
		monitoringListenAddr string
	)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Start the HTTP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info().
				Str("log_level", logLevelValue).
				Str("version", versioninfo.Short()).
				Msg("global: setting up services")

			// Periodic tasks
			var feedSynchronizingLocker sync.Mutex
			feedSynchronizingScheduler := feedsynchronizing.NewScheduler(
				feedSynchronizingService,
				&feedSynchronizingLocker,
			)
			go feedSynchronizingScheduler.Run(context.Background())

			// HTTP - Monitoring server
			monitoringServer, metricsRegistry := monitoring.NewServer(rootCmdName, monitoringListenAddr, versionDetails)

			go func() {
				log.Info().Str("addr", monitoringListenAddr).Msg("monitoring: listening for HTTP requests")

				if err := monitoringServer.ListenAndServe(); err != nil {
					log.Error().Err(err).Msg("monitoring: server stopped")
				}
			}()

			// HTTP - SparkleMuffin server
			publicURL, err := url.Parse(publicWebAddr)
			if err != nil {
				return fmt.Errorf("%s: failed to parse public HTTP address: %w", rootCmdName, err)
			}

			server, err := www.NewServer(
				www.WithMetricsRegistry(rootCmdName, metricsRegistry),
				www.WithPublicURL(publicURL),
				www.WithBookmarkServices(
					bookmarkService,
					bookmarkExportingService,
					bookmarkImportingService,
					bookmarkQueryingService,
				),
				www.WithFeedServices(
					feedService,
					feedExportingService,
					feedImportingService,
					feedQueryingService,
				),
				www.WithSessionService(sessionService),
				www.WithUserService(userService),
			)
			if err != nil {
				return fmt.Errorf("%s: failed to create server: %w", rootCmdName, err)
			}

			httpServer := &http.Server{
				Addr:         webListenAddr,
				Handler:      server,
				ReadTimeout:  15 * time.Second,
				WriteTimeout: 15 * time.Second,
			}

			log.Info().Str("addr", webListenAddr).Msgf("%s: listening for HTTP requests", rootCmdName)
			return httpServer.ListenAndServe()
		},
	}

	cmd.Flags().StringVar(
		&webListenAddr,
		"listen-addr",
		defaultWebListenAddr,
		"Listen to this address (host:port)",
	)

	cmd.Flags().StringVar(
		&monitoringListenAddr,
		"monitoring-listen-addr",
		defaultMonitoringListenAddr,
		"Listen to this address for monitoring (host:port)",
	)

	cmd.Flags().StringVar(
		&publicWebAddr,
		"public-addr",
		defaultPublicWebAddr,
		"Public HTTP address (if behind a proxy)",
	)

	return cmd
}
