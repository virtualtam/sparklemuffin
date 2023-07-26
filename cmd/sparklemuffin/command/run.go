package command

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/metrics"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/http/www"
)

const (
	defaultListenAddr     string = "0.0.0.0:8080"
	defaultPublicHTTPAddr string = "http://localhost:8080"

	defaultMetricsListenAddr string = "127.0.0.1:8081"

	defaultCSRFKey string = "csrf-secret-key"
)

var (
	listenAddr     string
	publicHTTPAddr string

	metricsListenAddr string

	csrfKey string
)

// NewRunCommand initializes a CLI command to start the HTTP server.
func NewRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Start the HTTP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Metrics server
			go func() {
				metricsServer := metrics.NewServer(metricsListenAddr)
				log.Info().Str("metrics_addr", metricsListenAddr).Msg("starting metrics server")

				if err := metricsServer.ListenAndServe(); err != nil {
					log.Error().Err(err).Msg("metrics server stopped")
				}
			}()

			// SparkleMuffin server
			publicURL, err := url.Parse(publicHTTPAddr)
			if err != nil {
				return fmt.Errorf("failed to parse public HTTP address: %w", err)
			}

			server := www.NewServer(
				www.WithCSRFKey(csrfKey),
				www.WithPublicURL(publicURL),
				www.WithBookmarkService(bookmarkService),
				www.WithExportingService(exportingService),
				www.WithQueryingService(queryingService),
				www.WithImportingService(importingService),
				www.WithSessionService(sessionService),
				www.WithUserService(userService),
			)

			httpServer := &http.Server{
				Addr:         listenAddr,
				Handler:      server,
				ReadTimeout:  15 * time.Second,
				WriteTimeout: 15 * time.Second,
			}

			log.Info().Str("http_addr", listenAddr).Msg("starting HTTP server")
			return httpServer.ListenAndServe()
		},
	}

	cmd.PersistentFlags().StringVar(
		&csrfKey,
		"csrf-key",
		defaultCSRFKey,
		"Secret key for CSRF token hashing",
	)

	cmd.Flags().StringVar(
		&listenAddr,
		"listen-addr",
		defaultListenAddr,
		"Listen to this address (host:port)",
	)

	cmd.Flags().StringVar(
		&metricsListenAddr,
		"metrics-listen-addr",
		defaultMetricsListenAddr,
		"Listen to this address for Prometheus metrics (host:port)",
	)

	cmd.Flags().StringVar(
		&publicHTTPAddr,
		"public-addr",
		defaultPublicHTTPAddr,
		"Public HTTP address (if behind a proxy)",
	)

	return cmd
}
