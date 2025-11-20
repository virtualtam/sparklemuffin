// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package pgbase

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/golang-migrate/migrate/v4"
	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jaswdr/faker/v2"
	"github.com/testcontainers/testcontainers-go"
	testpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	// Load the pgx PostgreSQL driver.
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/migrations"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

const (
	databaseDriver   = "pgx"
	databaseName     = "testdb"
	databaseUser     = "testuser"
	databasePassword = "testpass"
)

// CreateAndMigrateTestDatabase creates a new database and applies all SQL migrations.
func CreateAndMigrateTestDatabase(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURI, db := createTestDatabase(t)

	migrater := getDatabaseMigrater(t, db)
	if err := migrater.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("failed to apply database migrations (up): %q", err)
	}

	pool, err := pgxpool.New(t.Context(), databaseURI)
	if err != nil {
		t.Fatalf("failed to open database connection: %q", err)
	}

	return pool
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
func createTestDatabase(t *testing.T) (string, *sql.DB) {
	t.Helper()

	ctx := t.Context()

	pgContainer, err := testpostgres.Run(ctx,
		"postgres:15",
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
		t.Fatalf("failed to create postgres container: %q", err)
	}

	t.Cleanup(func() {
		// nolint: usetesting
		// t.Context() has already been canceled, create a new context to terminate the container.
		if err := pgContainer.Terminate(context.Background()); err != nil {
			t.Fatalf("failed to terminate postgres container: %q", err)
		}
	})

	databaseURI, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to obtain postgres connection string: %q", err)
	}

	db, err := sql.Open(databaseDriver, databaseURI)
	if err != nil {
		t.Fatalf("failed to open database connection: %q", err)
	}

	return databaseURI, db
}

func getDatabaseMigrater(t *testing.T, db *sql.DB) *migrate.Migrate {
	t.Helper()

	migrationsSource, err := iofs.New(migrations.FS, ".")
	if err != nil {
		t.Fatalf("failed to open the database migration filesystem: %q", err)
	}

	driver, err := migratepgx.WithInstance(db, &migratepgx.Config{})
	if err != nil {
		t.Fatalf("failed to prepare the database driver: %q", err)
	}

	migrater, err := migrate.NewWithInstance(
		"iofs",
		migrationsSource,
		databaseDriver,
		driver,
	)
	if err != nil {
		t.Fatalf("failed to load database migrations: %q", err)
	}

	return migrater
}

// GenerateFakeUser generates a new user for testing.
func GenerateFakeUser(t *testing.T, fake *faker.Faker) user.User {
	t.Helper()

	person := fake.Person()
	internet := fake.Internet()

	// Nicknames must match user.nickNameRegex
	nick := strings.ReplaceAll(internet.User(), ".", "")

	return user.User{
		Email:       person.Contact().Email,
		NickName:    nick,
		DisplayName: person.Name(),
		Password:    internet.Password(),
	}
}
