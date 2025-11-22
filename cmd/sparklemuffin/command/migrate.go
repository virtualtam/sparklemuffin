// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package command

import (
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/migrations"
)

// NewMigrateCommand initializes a CLI command to create database tables and run
// SQL migrations.
func NewMigrateCommand() *cobra.Command {
	var dbVersion uint

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Initialize database and run migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := log.With().
				Str("database_driver", databaseDriver).
				Str("database_addr", databaseAddr).
				Str("database_name", databaseName).
				Logger()

			migrationsSource, err := iofs.New(migrations.FS, ".")
			if err != nil {
				logger.Error().Err(err).Msg("failed to open the database migration filesystem")
			}

			db, err := sql.Open(databaseDriver, databaseURI)
			if err != nil {
				logger.Error().Err(err).Msg("failed to open database connection")
				return err
			}
			logger.Info().Msg("successfully opened database connection")

			driver, err := migratepgx.WithInstance(db, &migratepgx.Config{})
			if err != nil {
				logger.Error().Err(err).Msg("migrate: failed to prepare the database driver")
			}

			migrater, err := migrate.NewWithInstance(
				"iofs",
				migrationsSource,
				databaseDriver,
				driver,
			)
			if err != nil {
				logger.Error().Err(err).Msg("migrate: failed to load database migrations")
				return err
			}

			migrater.Log = newMigrateLogger(logger, logLevelValue)

			previousVersion, _, err := migrater.Version()
			if err != nil {
				logger.Error().Err(err).Msg("migrate: failed to retrieve database migration version")
				return err
			}

			if dbVersion != 0 {
				// Migrate to a specific database schema version.
				err = migrater.Migrate(dbVersion)
			} else {
				// Migrate to the latest database schema version.
				err = migrater.Up()
			}

			if errors.Is(err, migrate.ErrNoChange) {
				currentVersion, dirty, err := migrater.Version()
				if err != nil {
					logger.Error().Err(err).Msg("migrate: failed to retrieve database migration version")
					return err
				}

				logger.Info().
					Uint("current_version", currentVersion).
					Bool("dirty", dirty).
					Msg("migrate: nothing to do, the database schema is up to date")
				return nil
			}

			if err != nil {
				logger.Error().Err(err).Msg("migrate: failed to apply database migrations")
				return err
			}

			currentVersion, dirty, err := migrater.Version()
			if err != nil {
				logger.Error().Err(err).Msg("migrate: failed to retrieve database migration version")
				return err
			}

			logger.Info().
				Uint("current_version", currentVersion).
				Uint("previous_version", previousVersion).
				Bool("dirty", dirty).
				Msg("migrate: all database migrations have been applied")

			return nil
		},
	}

	cmd.Flags().UintVar(
		&dbVersion,
		"db-version",
		0,
		"Migrate the database schema to this version",
	)

	return cmd
}

var _ migrate.Logger = &migrateLogger{}

type migrateLogger struct {
	logger  zerolog.Logger
	verbose bool
}

func newMigrateLogger(logger zerolog.Logger, logLevelValue string) *migrateLogger {
	return &migrateLogger{
		logger:  logger,
		verbose: logLevelValue == zerolog.LevelTraceValue || logLevelValue == zerolog.LevelDebugValue,
	}
}

func (l *migrateLogger) Printf(format string, v ...any) {
	l.logger.Printf(format, v...)
}

func (l *migrateLogger) Verbose() bool {
	return l.verbose
}
