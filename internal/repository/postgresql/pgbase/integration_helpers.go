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
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jaswdr/faker"
	"github.com/testcontainers/testcontainers-go"
	testpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/migrations"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

const (
	databaseDriver   = "pgx"
	databaseName     = "testdb"
	databaseUser     = "testuser"
	databasePassword = "testpass"
)

func CreateAndMigrateTestDatabase(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()

	databaseURI, db := createTestDatabase(t, ctx)

	migrater := getDatabaseMigrater(t, db)
	if err := migrater.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("failed to apply database migrations (up): %q", err)
	}

	pool, err := pgxpool.New(context.Background(), databaseURI)
	if err != nil {
		t.Fatalf("failed to open database connection: %q", err)
	}

	return pool
}

func createTestDatabase(t *testing.T, ctx context.Context) (string, *sql.DB) {
	t.Helper()

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
		if err := pgContainer.Terminate(ctx); err != nil {
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
