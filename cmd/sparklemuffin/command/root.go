package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/config"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	"github.com/virtualtam/sparklemuffin/pkg/exporting"
	"github.com/virtualtam/sparklemuffin/pkg/importing"
	"github.com/virtualtam/sparklemuffin/pkg/querying"
	"github.com/virtualtam/sparklemuffin/pkg/session"
	"github.com/virtualtam/sparklemuffin/pkg/user"
	"github.com/virtualtam/venom"

	_ "github.com/jackc/pgx/v4/stdlib"
)

const (
	defaultHMACKey string = "hmac-secret-key"

	databaseDriver string = "pgx"

	defaultDatabaseAddr     string = "localhost:15432"
	defaultDatabaseName     string = "sparklemuffin"
	defaultDatabaseUser     string = "sparklemuffin"
	defaultDatabasePassword string = "sparklemuffin"
)

var (
	defaultLogLevelValue string = zerolog.LevelInfoValue
	logLevelValue        string

	hmacKey string

	databaseAddr     string
	databaseName     string
	databaseUser     string
	databasePassword string

	db *sqlx.DB

	bookmarkService  *bookmark.Service
	exportingService *exporting.Service
	importingService *importing.Service
	queryingService  *querying.Service
	sessionService   *session.Service
	userService      *user.Service
)

// NewRootCommand initializes the main CLI entrypoint and common command flags.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sparklemuffin",
		Short: "Web Bookmark Manager",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error

			// Configuration file lookup paths
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			homeConfigPath := filepath.Join(home, ".config")

			configPaths := []string{config.DefaultConfigPath, homeConfigPath, "."}

			// Inject global configuration as a pre-run hook
			//
			// This is required to let Viper load environment variables and
			// configuration entries before invoking nested commands.
			if err := venom.Inject(cmd, config.EnvPrefix, configPaths, config.ConfigName, false); err != nil {
				return err
			}

			// Global logger configuration
			var logLevel zerolog.Level

			if err := logLevel.UnmarshalText([]byte(logLevelValue)); err != nil {
				log.Error().Err(err).Msg("invalid log level")
				return err
			}

			log.Info().Str("log_level", logLevelValue).Msg("setting log level")
			zerolog.SetGlobalLevel(logLevel)

			// Database connection pool
			databaseURI := fmt.Sprintf(
				"postgres://%s:%s@%s/%s?sslmode=disable",
				databaseUser,
				databasePassword,
				databaseAddr,
				databaseName,
			)

			db, err = sqlx.Connect(databaseDriver, databaseURI)
			if err != nil {
				log.Error().
					Err(err).
					Str("database_driver", databaseDriver).
					Str("database_addr", databaseAddr).
					Str("database_name", databaseName).
					Msg("failed to connect to database")
				return err
			}
			log.Info().
				Str("database_driver", databaseDriver).
				Str("database_addr", databaseAddr).
				Str("database_name", databaseName).
				Msg("successfully connected to database")

			// Main database repository
			repository := postgresql.NewRepository(db)

			// sparklemuffin services
			bookmarkService = bookmark.NewService(repository)
			exportingService = exporting.NewService(repository)
			importingService = importing.NewService(repository)
			queryingService = querying.NewService(repository)
			sessionService = session.NewService(repository, hmacKey)
			userService = user.NewService(repository)

			return nil
		},
	}

	var logLevelValues = []string{
		zerolog.LevelTraceValue,
		zerolog.LevelDebugValue,
		zerolog.LevelInfoValue,
		zerolog.LevelWarnValue,
		zerolog.LevelErrorValue,
		zerolog.LevelFatalValue,
		zerolog.LevelPanicValue,
	}

	cmd.PersistentFlags().StringVar(
		&logLevelValue,
		"log-level",
		defaultLogLevelValue,
		fmt.Sprintf(
			"Log level (%s)",
			strings.Join(logLevelValues, ", "),
		),
	)

	cmd.PersistentFlags().StringVar(
		&databaseAddr,
		"db-addr",
		defaultDatabaseAddr,
		"Database address (host:port)",
	)
	cmd.PersistentFlags().StringVar(
		&databaseName,
		"db-name",
		defaultDatabaseName,
		"Database name",
	)
	cmd.PersistentFlags().StringVar(
		&databaseUser,
		"db-user",
		defaultDatabaseUser,
		"Database user",
	)
	cmd.PersistentFlags().StringVar(
		&databasePassword,
		"db-password",
		defaultDatabasePassword,
		"Database password",
	)

	cmd.PersistentFlags().StringVar(
		&hmacKey,
		"hmac-key",
		defaultHMACKey,
		"Secret HMAC key for token hashing",
	)

	return cmd
}