// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package pgbase

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moby/moby/api/types/container"
	"github.com/testcontainers/testcontainers-go"
	testpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	// Load the pgx PostgreSQL driver.
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/migrations"
)

const (
	databaseDriver   = "pgx"
	databaseName     = "testdb"
	databaseUser     = "testuser"
	databasePassword = "testpass"
)

// CreateAndMigrateTestDatabase creates a new database and applies all SQL migrations.
func CreateAndMigrateTestDatabase(tb testing.TB) *pgxpool.Pool {
	tb.Helper()

	pool, migrater := CreateTestDatabaseAndMigrater(tb)

	if err := migrater.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		tb.Fatalf("failed to apply database migrations (up): %q", err)
	}

	return pool
}

// CreateTestDatabaseAndMigrater creates a new database and returns it alongside its Migrate
// instance, without applying any migration, so a test can step through migrations one at a
// time (e.g. to inject legacy data between two versions and exercise a backfill migration).
func CreateTestDatabaseAndMigrater(tb testing.TB) (*pgxpool.Pool, *migrate.Migrate) {
	tb.Helper()

	databaseURI, db := createTestDatabase(tb)

	migrater := getDatabaseMigrater(tb, db)

	pool, err := pgxpool.New(tb.Context(), databaseURI)
	if err != nil {
		tb.Fatalf("failed to open database connection: %q", err)
	}

	return pool, migrater
}

// createTestDatabase creates a PostgreSQL container and returns the connection string and database connection.
//
// PostgreSQL is configured for speed:
// - data is stored using a tmpfs volume;
// - WAL features are disabled.
//
// See:
// - https://www.postgresql.org/docs/15/runtime-config-wal.html
// - https://stackoverflow.com/questions/9407442/optimise-postgresql-for-fast-testing
// - https://stackoverflow.com/questions/30848670/how-to-customize-the-configuration-file-of-the-official-postgresql-docker-image
func createTestDatabase(tb testing.TB) (string, *sql.DB) {
	tb.Helper()

	ctx := tb.Context()

	pgContainer, err := testpostgres.Run(ctx,
		"postgres:17",
		testpostgres.WithDatabase(databaseName),
		testpostgres.WithUsername(databaseUser),
		testpostgres.WithPassword(databasePassword),
		testcontainers.WithHostConfigModifier(func(hostConfig *container.HostConfig) {
			hostConfig.Tmpfs = map[string]string{
				"/var/lib/postgresql/data": "rw",
			}
		}),
		testcontainers.WithConfigModifier(func(config *container.Config) {
			config.Cmd = []string{
				"postgres",
				"-c", "fsync=off",
				"-c", "synchronous_commit=off",
				"-c", "full_page_writes=off",
				"-c", "shared_buffers=512MB",
				"-c", "autovacuum=off",
			}
		}),
		testcontainers.WithWaitStrategy(
			wait.
				ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		tb.Fatalf("failed to create postgres container: %q", err)
	}

	tb.Cleanup(func() {
		// nolint: usetesting
		// tb.Context() has already been canceled, create a new context to terminate the container.
		if err := pgContainer.Terminate(context.Background()); err != nil {
			tb.Fatalf("failed to terminate postgres container: %q", err)
		}
	})

	databaseURI, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		tb.Fatalf("failed to obtain postgres connection string: %q", err)
	}

	db, err := sql.Open(databaseDriver, databaseURI)
	if err != nil {
		tb.Fatalf("failed to open database connection: %q", err)
	}

	return databaseURI, db
}

func getDatabaseMigrater(tb testing.TB, db *sql.DB) *migrate.Migrate {
	tb.Helper()

	migrationsSource, err := iofs.New(migrations.FS, ".")
	if err != nil {
		tb.Fatalf("failed to open the database migration filesystem: %q", err)
	}

	driver, err := migratepgx.WithInstance(db, &migratepgx.Config{})
	if err != nil {
		tb.Fatalf("failed to prepare the database driver: %q", err)
	}

	migrater, err := migrate.NewWithInstance(
		"iofs",
		migrationsSource,
		databaseDriver,
		driver,
	)
	if err != nil {
		tb.Fatalf("failed to load database migrations: %q", err)
	}

	return migrater
}
