package command

import (
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/virtualtam/yawbe/pkg/bookmark"
	"github.com/virtualtam/yawbe/pkg/session"
	"github.com/virtualtam/yawbe/pkg/storage/postgresql"
	"github.com/virtualtam/yawbe/pkg/user"

	_ "github.com/jackc/pgx/v4/stdlib"
)

const (
	databaseDriver string = "pgx"

	defaultDebugMode   bool   = false
	defaultHMACKey     string = "hmac-secret-key"
	defaultDatabaseURI string = "postgres://yawbe:yawbe@localhost:15432/yawbe?sslmode=disable"
)

var (
	debugMode bool
	hmacKey   string

	db *sqlx.DB

	bookmarkService *bookmark.Service
	sessionService  *session.Service
	userService     *user.Service
)

// NewRootCommand initializes the main CLI entrypoint and common command flags.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "yawbe-srv",
		Short: "Yet Another Web Bookmarking Engine",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error

			if debugMode {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}

			// FIXME hardcoded
			db, err = sqlx.Connect(databaseDriver, defaultDatabaseURI)
			if err != nil {
				log.Error().Err(err).Msg("failed to connect to PostgresSQL")
				return err
			}
			log.Info().Msg("Successfully connected to PostgreSQL")

			repository := postgresql.NewRepository(db)

			bookmarkService = bookmark.NewService(repository)
			sessionService = session.NewService(repository, hmacKey)
			userService = user.NewService(repository)

			return nil
		},
	}

	cmd.PersistentFlags().BoolVar(
		&debugMode,
		"debug",
		defaultDebugMode,
		"Enable debugging",
	)

	cmd.PersistentFlags().StringVar(
		&hmacKey,
		"hmac-key",
		defaultHMACKey,
		"Secret HMAC key for token hashing",
	)

	return cmd
}
