package command

import (
	"errors"

	"github.com/golang-migrate/migrate/v4"
	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/virtualtam/yawbe/internal/repository/postgresql/migrations"
)

var _ migrate.Logger = &migrateLogger{}

type migrateLogger struct {
	verbose bool
}

func (l migrateLogger) Printf(format string, v ...interface{}) {
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

			driver, err := migratepgx.WithInstance(db.DB, &migratepgx.Config{})
			if err != nil {
				log.Error().Err(err).Msg("failed to prepare the database driver")
			}

			migrater, err := migrate.NewWithInstance(
				"iofs",
				migrationsSource,
				databaseDriver,
				driver,
			)
			if err != nil {
				log.Error().Err(err).Msg("failed to load database migrations")
				return err
			}

			var verbose bool
			if logLevelValue == zerolog.LevelTraceValue || logLevelValue == zerolog.LevelDebugValue {
				verbose = true
			}
			migrater.Log = migrateLogger{verbose: verbose}

			err = migrater.Up()
			if errors.Is(err, migrate.ErrNoChange) {
				log.Info().Msg("The database schema is up to date")
			} else if err != nil {
				log.Error().Err(err).Msg("Failed to apply database migrations")
			} else {
				log.Info().Msg("All database migrations have been applied")
			}

			return nil
		},
	}

	return cmd
}
