package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/virtualtam/venom"
	"github.com/virtualtam/yawbe/cmd/yawbe/config"
	"github.com/virtualtam/yawbe/pkg/bookmark"
	"github.com/virtualtam/yawbe/pkg/displaying"
	"github.com/virtualtam/yawbe/pkg/exporting"
	"github.com/virtualtam/yawbe/pkg/importing"
	"github.com/virtualtam/yawbe/pkg/session"
	"github.com/virtualtam/yawbe/pkg/storage/postgresql"
	"github.com/virtualtam/yawbe/pkg/user"

	_ "github.com/jackc/pgx/v4/stdlib"
)

const (
	databaseDriver string = "pgx"

	defaultDebugMode bool   = false
	defaultHMACKey   string = "hmac-secret-key"

	defaultDatabaseAddr     string = "localhost:15432"
	defaultDatabaseName     string = "yawbe"
	defaultDatabaseUser     string = "yawbe"
	defaultDatabasePassword string = "yawbe"
)

var (
	debugMode bool
	hmacKey   string

	databaseAddr     string
	databaseName     string
	databaseUser     string
	databasePassword string

	db *sqlx.DB

	bookmarkService   *bookmark.Service
	displayingService *displaying.Service
	exportingService  *exporting.Service
	importingService  *importing.Service
	sessionService    *session.Service
	userService       *user.Service
)

// NewRootCommand initializes the main CLI entrypoint and common command flags.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "yawbe",
		Short: "Yet Another Web Bookmarking Engine",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error

			if debugMode {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}

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

			databaseURI := fmt.Sprintf(
				"postgres://%s:%s@%s/%s?sslmode=disable",
				databaseUser,
				databasePassword,
				databaseAddr,
				databaseName,
			)

			db, err = sqlx.Connect(databaseDriver, databaseURI)
			if err != nil {
				log.Error().Err(err).Msg("failed to connect to PostgresSQL")
				return err
			}
			log.Info().Msg("Successfully connected to PostgreSQL")

			repository := postgresql.NewRepository(db)

			bookmarkService = bookmark.NewService(repository)
			displayingService = displaying.NewService(repository)
			exportingService = exporting.NewService(repository)
			importingService = importing.NewService(repository)
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
