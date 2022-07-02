package command

import (
	"net/http"
	"time"

	"github.com/justinas/alice"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/virtualtam/yawbe/pkg/http/www"
)

const (
	defaultListenAddr string = "0.0.0.0:8080"
)

var (
	listenAddr string
)

// NewRunCommand initializes a CLI command to start the HTTP server.
func NewRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Start the HTTP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			server := www.NewServer().
				WithBookmarkService(bookmarkService).
				WithExportingService(exportingService).
				WithImportingService(importingService).
				WithSessionService(sessionService).
				WithUserService(userService)

			// Structured loging
			chain := alice.New(hlog.NewHandler(log.Logger), hlog.AccessHandler(accessLogger))

			httpServer := &http.Server{
				Addr:         listenAddr,
				Handler:      chain.Then(server),
				ReadTimeout:  15 * time.Second,
				WriteTimeout: 15 * time.Second,
			}

			log.Info().Msgf("Listening to http://%s", listenAddr)
			return httpServer.ListenAndServe()
		},
	}

	cmd.Flags().StringVar(
		&listenAddr,
		"listen-addr",
		defaultListenAddr,
		"Listen to this address",
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
		Msg("Request")
}
