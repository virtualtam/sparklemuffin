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

var _ migrate.Logger = &migrateLogger{}

type migrateLogger struct {
	verbose bool
}

func (l migrateLogger) Printf(format string, v ...any) {
	log.Printf(format, v...)
}

func (l migrateLogger) Verbose() bool {
	return l.verbose
}

// NewMigrateCommand initializes a CLI command to create database tables and run
// SQL migrations.
func NewMigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Initialize database and run migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			migrationsSource, err := iofs.New(migrations.FS, ".")
			if err != nil {
				log.Error().Err(err).Msg("failed to open the database migration filesystem")
			}

			db, err := sql.Open(databaseDriver, databaseURI)
			if err != nil {
				log.Error().
					Err(err).
					Str("database_driver", databaseDriver).
					Str("database_addr", databaseAddr).
					Str("database_name", databaseName).
					Msg("failed to open database connection")
				return err
			}
			log.Info().
				Str("database_driver", databaseDriver).
				Str("database_addr", databaseAddr).
				Str("database_name", databaseName).
				Msg("successfully opened database connection")

			driver, err := migratepgx.WithInstance(db, &migratepgx.Config{})
			if err != nil {
				log.Error().
					Err(err).
					Str("database_driver", databaseDriver).
					Str("database_addr", databaseAddr).
					Msg("migrate: failed to prepare the database driver")
			}

			migrater, err := migrate.NewWithInstance(
				"iofs",
				migrationsSource,
				databaseDriver,
				driver,
			)
			if err != nil {
				log.Error().
					Err(err).
					Str("database_driver", databaseDriver).
					Str("database_addr", databaseAddr).
					Msg("migrate: failed to load database migrations")
				return err
			}

			var verbose bool
			if logLevelValue == zerolog.LevelTraceValue || logLevelValue == zerolog.LevelDebugValue {
				verbose = true
			}
			migrater.Log = migrateLogger{verbose: verbose}

			err = migrater.Up()
			if errors.Is(err, migrate.ErrNoChange) {
				log.Info().
					Str("database_driver", databaseDriver).
					Str("database_addr", databaseAddr).
					Msg("migrate: the database schema is up to date")
			} else if err != nil {
				log.Error().
					Err(err).
					Str("database_driver", databaseDriver).
					Str("database_addr", databaseAddr).
					Msg("migrate: failed to apply database migrations")
			} else {
				log.Info().
					Str("database_driver", databaseDriver).
					Str("database_addr", databaseAddr).
					Msg("migrate: all database migrations have been applied")
			}

			return nil
		},
	}

	return cmd
}
