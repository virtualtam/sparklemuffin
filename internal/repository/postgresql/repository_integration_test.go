package postgresql_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

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

func createTestDatabase(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()

	pgContainer, err := testpostgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15"),
		testpostgres.WithDatabase(databaseName),
		testpostgres.WithUsername(databaseUser),
		testpostgres.WithPassword(databasePassword),
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
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close database connection used for migrations: %q", err)
		}
	}()

	migrateTestDatabase(t, db)

	pool, err := pgxpool.New(context.Background(), databaseURI)
	if err != nil {
		t.Fatalf("failed to open database connection: %q", err)
	}

	return pool
}

func migrateTestDatabase(t *testing.T, db *sql.DB) {
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

	if err := migrater.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("failed to apply database migrations: %q", err)
	}
}

func generateFakeUser(t *testing.T, fake *faker.Faker) user.User {
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
