package postgresql_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jaswdr/faker"
	"github.com/jmoiron/sqlx"
	"github.com/testcontainers/testcontainers-go"
	testpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql"
	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/migrations"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

const (
	databaseDriver   = "pgx"
	databaseName     = "testdb"
	databaseUser     = "testuser"
	databasePassword = "testpass"
)

func createTestDatabase(t *testing.T, ctx context.Context) *sqlx.DB {
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

	db, err := sqlx.Connect(databaseDriver, databaseURI)
	if err != nil {
		t.Fatalf("failed to open database connection: %q", err)
	}

	migrateTestDatabase(t, db)

	return db
}

func migrateTestDatabase(t *testing.T, db *sqlx.DB) {
	migrationsSource, err := iofs.New(migrations.FS, ".")
	if err != nil {
		t.Fatalf("failed to open the database migration filesystem: %q", err)
	}

	driver, err := migratepgx.WithInstance(db.DB, &migratepgx.Config{})
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

func TestUserService(t *testing.T) {
	ctx := context.Background()
	db := createTestDatabase(t, ctx)
	r := postgresql.NewRepository(db)
	s := user.NewService(r)

	fake := faker.New()

	t.Run("create, retrieve and delete user", func(t *testing.T) {
		u := generateFakeUser(t, &fake)

		if err := s.Add(u); err != nil {
			t.Fatalf("failed to create user: %q", err)
		}

		gotUser, err := s.ByNickName(u.NickName)
		if err != nil {
			t.Fatalf("failed to retrieve user: %q", err)
		}

		if gotUser.Email != u.Email {
			t.Errorf("want email %q, got %q", u.Email, gotUser.Email)
		}
		if gotUser.IsAdmin != u.IsAdmin {
			t.Errorf("want admin %t, got %t", u.IsAdmin, gotUser.IsAdmin)
		}
		if gotUser.UUID == "" {
			t.Error("want UUID to be set")
		}

		if err := s.DeleteByUUID(gotUser.UUID); err != nil {
			t.Fatalf("failed to delete user by UUID: %q", err)
		}

		_, err = s.ByNickName(u.NickName)
		if !errors.Is(err, user.ErrNotFound) {
			t.Fatalf("want %q, got %q", user.ErrNotFound, err)
		}
	})

	t.Run("update user", func(t *testing.T) {
		u := generateFakeUser(t, &fake)

		if err := s.Add(u); err != nil {
			t.Fatalf("failed to create user: %q", err)
		}

		gotUser, err := s.ByNickName(u.NickName)
		if err != nil {
			t.Fatalf("failed to retrieve user: %q", err)
		}

		updatedPerson := fake.Person()

		updatedUser := user.User{
			UUID:        gotUser.UUID,
			Email:       updatedPerson.Contact().Email,
			NickName:    gotUser.NickName,
			DisplayName: updatedPerson.Name(),
			Password:    fake.Internet().Password(),
		}

		if err := s.Update(updatedUser); err != nil {
			t.Fatalf("want no error, got %q", err)
		}

		gotUpdatedUser, err := s.ByUUID(gotUser.UUID)
		if err != nil {
			t.Fatalf("failed to retrieve user: %q", err)
		}

		if gotUpdatedUser.Email != updatedUser.Email {
			t.Errorf("want email %q, got %q", updatedUser.Email, gotUpdatedUser.Email)
		}
		if gotUpdatedUser.DisplayName != updatedUser.DisplayName {
			t.Errorf("want display name %q, got %q", updatedUser.DisplayName, gotUpdatedUser.DisplayName)
		}
	})

	t.Run("update user info with no change", func(t *testing.T) {
		u := generateFakeUser(t, &fake)

		if err := s.Add(u); err != nil {
			t.Fatalf("failed to create user: %q", err)
		}

		gotUser, err := s.ByNickName(u.NickName)
		if err != nil {
			t.Fatalf("failed to retrieve user: %q", err)
		}

		info := user.InfoUpdate{
			UUID:        gotUser.UUID,
			Email:       gotUser.Email,
			NickName:    gotUser.NickName,
			DisplayName: gotUser.DisplayName,
		}

		if err := s.UpdateInfo(info); err != nil {
			t.Fatalf("want no error, got %q", err)
		}

		gotUpdatedUser, err := s.ByUUID(gotUser.UUID)
		if err != nil {
			t.Fatalf("failed to retrieve user: %q", err)
		}

		if gotUpdatedUser.Email != u.Email {
			t.Errorf("want email %q, got %q", u.Email, gotUpdatedUser.Email)
		}
		if gotUpdatedUser.DisplayName != u.DisplayName {
			t.Errorf("want display name %q, got %q", u.DisplayName, gotUpdatedUser.DisplayName)
		}
	})

	t.Run("update user info", func(t *testing.T) {
		u := generateFakeUser(t, &fake)

		if err := s.Add(u); err != nil {
			t.Fatalf("failed to create user: %q", err)
		}

		gotUser, err := s.ByNickName(u.NickName)
		if err != nil {
			t.Fatalf("failed to retrieve user: %q", err)
		}

		newPerson := fake.Person()

		info := user.InfoUpdate{
			UUID:        gotUser.UUID,
			Email:       newPerson.Contact().Email,
			NickName:    gotUser.NickName,
			DisplayName: newPerson.Name(),
		}

		if err := s.UpdateInfo(info); err != nil {
			t.Fatalf("want no error, got %q", err)
		}

		gotUpdatedUser, err := s.ByUUID(gotUser.UUID)
		if err != nil {
			t.Fatalf("failed to retrieve user: %q", err)
		}

		if gotUpdatedUser.Email != info.Email {
			t.Errorf("want email %q, got %q", info.Email, gotUpdatedUser.Email)
		}
		if gotUpdatedUser.DisplayName != info.DisplayName {
			t.Errorf("want display name %q, got %q", info.DisplayName, gotUpdatedUser.DisplayName)
		}
	})

	t.Run("update user password", func(t *testing.T) {
		u := generateFakeUser(t, &fake)

		if err := s.Add(u); err != nil {
			t.Fatalf("failed to create user: %q", err)
		}

		gotUser, err := s.ByNickName(u.NickName)
		if err != nil {
			t.Fatalf("failed to retrieve user: %q", err)
		}

		newPassword := fake.Internet().Password()

		passwordUpdate := user.PasswordUpdate{
			UUID:                    gotUser.UUID,
			CurrentPassword:         u.Password,
			NewPassword:             newPassword,
			NewPasswordConfirmation: newPassword,
		}

		if err := s.UpdatePassword(passwordUpdate); err != nil {
			t.Fatalf("want no error, got %q", err)
		}

		gotUpdatedUser, err := s.ByUUID(gotUser.UUID)
		if err != nil {
			t.Fatalf("failed to retrieve user: %q", err)
		}

		if gotUpdatedUser.PasswordHash == u.PasswordHash {
			t.Error("password hash was not updated")
		}
	})

	t.Run("authenticate user", func(t *testing.T) {
		u := generateFakeUser(t, &fake)

		if err := s.Add(u); err != nil {
			t.Fatalf("failed to create user: %q", err)
		}

		authenticatedUser, err := s.Authenticate(u.Email, u.Password)
		if err != nil {
			t.Fatalf("want no error, got %q", err)
		}

		if authenticatedUser.Email != u.Email {
			t.Errorf("want email %q, got %q", u.Email, authenticatedUser.Email)
		}
	})
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
