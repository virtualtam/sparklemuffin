// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package command

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/virtualtam/venom"

	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/config"
	"github.com/virtualtam/sparklemuffin/cmd/sparklemuffin/version"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark/exporting"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark/importing"
	"github.com/virtualtam/sparklemuffin/pkg/bookmark/querying"
	"github.com/virtualtam/sparklemuffin/pkg/session"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

const (
	rootCmdName string = "sparklemuffin"

	defaultHMACKey string = "hmac-secret-key"

	databaseDriver string = "pgx"

	defaultDatabaseAddr     string = "localhost:15432"
	defaultDatabaseSSLMode  string = "disable"
	defaultDatabaseName     string = "sparklemuffin"
	defaultDatabaseUser     string = "sparklemuffin"
	defaultDatabasePassword string = "sparklemuffin"
)

var (
	defaultLogLevelValue string = zerolog.LevelInfoValue
	logLevelValue        string

	versionDetails *version.Details

	hmacKey string

	databaseAddr     string
	databaseSSLMode  string
	databaseName     string
	databaseUser     string
	databasePassword string

	// Database connection URI. Populated by the root command.
	databaseURI string

	pgxPool *pgxpool.Pool

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
		Use:   rootCmdName,
		Short: "SparkleMuffin - Web Bookmark Manager",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			versionDetails = version.NewDetails()

			if cmd.Name() == versionCmdName {
				// Do not setup the service stack for these commands
				return nil
			}

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
			v := viper.New()
			if err := venom.InjectTo(v, cmd, config.EnvPrefix, configPaths, config.ConfigName, false); err != nil {
				return err
			}

			// Global logger configuration
			var logLevel zerolog.Level

			if err := logLevel.UnmarshalText([]byte(logLevelValue)); err != nil {
				log.Error().Err(err).Msg("invalid log level")
				return err
			}

			zerolog.SetGlobalLevel(logLevel)

			if configFileUsed := v.ConfigFileUsed(); configFileUsed != "" {
				log.Info().Str("config_file", v.ConfigFileUsed()).Msg("configuration: using file")
			} else {
				log.Info().Strs("config_paths", configPaths).Msg("configuration: no file found")
			}

			// Encode the database password with percent encoding in case it contains special characters.
			// https://www.postgresql.org/docs/current/libpq-connect.html
			// https://datatracker.ietf.org/doc/html/rfc3986#section-2.1
			databasePassword = url.QueryEscape(databasePassword)
			databaseURI = fmt.Sprintf(
				"postgres://%s:%s@%s/%s?sslmode=%s",
				databaseUser,
				databasePassword,
				databaseAddr,
				databaseName,
				databaseSSLMode,
			)

			pgxPool, err = pgxpool.New(context.Background(), databaseURI)
			if err != nil {
				log.Error().
					Err(err).
					Str("database_driver", databaseDriver).
					Str("database_addr", databaseAddr).
					Str("database_name", databaseName).
					Msg("database: failed to create connection pool")
				return err
			}

			if err := pgxPool.Ping(context.Background()); err != nil {
				log.Error().
					Err(err).
					Str("database_driver", databaseDriver).
					Str("database_addr", databaseAddr).
					Str("database_name", databaseName).
					Msg("database: failed to ping")
				return err
			}

			log.Info().
				Str("database_driver", databaseDriver).
				Str("database_addr", databaseAddr).
				Str("database_name", databaseName).
				Msg("database: successfully created connection pool")

			// Main database repository
			repository := postgresql.NewRepository(pgxPool)

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
		&databaseSSLMode,
		"db-sslmode",
		defaultDatabaseSSLMode,
		"Database sslmode",
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
		"Secret key for HMAC session token hashing",
	)

	return cmd
}
