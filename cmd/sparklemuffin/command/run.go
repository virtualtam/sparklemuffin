package command

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/justinas/alice"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/virtualtam/sparklemuffin/pkg/http/www"
)

const (
	defaultListenAddr     string = "0.0.0.0:8080"
	defaultPublicHTTPAddr string = "http://localhost:8080"
)

var (
	listenAddr     string
	publicHTTPAddr string
)

// NewRunCommand initializes a CLI command to start the HTTP server.
func NewRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Start the HTTP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			publicURL, err := url.Parse(publicHTTPAddr)
			if err != nil {
				return fmt.Errorf("failed to parse public HTTP address: %w", err)
			}

			server := www.NewServer(
				www.WithPublicURL(publicURL),
				www.WithBookmarkService(bookmarkService),
				www.WithExportingService(exportingService),
				www.WithQueryingService(queryingService),
				www.WithImportingService(importingService),
				www.WithSessionService(sessionService),
				www.WithUserService(userService),
			)

			// Structured logging
			chain := alice.New(hlog.NewHandler(log.Logger), hlog.AccessHandler(accessLogger))

			httpServer := &http.Server{
				Addr:         listenAddr,
				Handler:      chain.Then(server),
				ReadTimeout:  15 * time.Second,
				WriteTimeout: 15 * time.Second,
			}

			log.Info().Str("addr", listenAddr).Msg("starting HTTP server")
			return httpServer.ListenAndServe()
		},
	}

	cmd.Flags().StringVar(
		&listenAddr,
		"listen-addr",
		defaultListenAddr,
		"Listen to this address",
	)

	cmd.Flags().StringVar(
		&publicHTTPAddr,
		"public-addr",
		defaultPublicHTTPAddr,
		"Public HTTP address (if behind a proxy)",
	)

	return cmd
}

func accessLogger(r *http.Request, status, size int, dur time.Duration) {
	hlog.FromRequest(r).Info().
		Dur("duration_ms", dur).
		Str("host", r.Host).
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Int("size", size).
		Int("status", status).
		Msg("handle request")
}
