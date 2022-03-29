package command

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/virtualtam/yawbe/pkg/storage/postgresql/migrations"

	_ "github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/jackc/pgx/v4/stdlib"
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

			migrater, err := migrate.NewWithSourceInstance(
				"iofs",
				migrationsSource,
				fmt.Sprintf("%s://%s", defaultDatabaseDriver, defaultDatabaseURI),
			)
			if err != nil {
				log.Error().Err(err).Msg("failed to load database migrations")
				return err
			}

			migrater.Log = migrateLogger{verbose: debugMode}

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
