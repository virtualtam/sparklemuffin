package postgresql

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/virtualtam/yawbe/pkg/user"
)

var _ user.Repository = &Repository{}

// Repository provides a PostgreSQL persistence layer.
type Repository struct {
	db *sqlx.DB
}

// NewRepository initializes and returns a PostgreSQL Repository.
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) AddUser(u user.User) error {
	dbUser := User{
		UUID:              u.UUID,
		Email:             u.Email,
		PasswordHash:      u.PasswordHash,
		RememberTokenHash: u.RememberTokenHash,
		IsAdmin:           u.IsAdmin,
	}

	_, err := r.db.NamedExec(
		"INSERT INTO users(uuid, email, password_hash, remember_token_hash, is_admin) VALUES(:uuid, :email, :password_hash, :remember_token_hash, :is_admin)",
		dbUser,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetUserByEmail(email string) (user.User, error) {
	dbUser := &User{}

	err := r.db.QueryRowx(
		"SELECT uuid, email, password_hash, remember_token_hash, is_admin FROM users WHERE email=$1",
		email,
	).StructScan(dbUser)

	if errors.Is(err, sql.ErrNoRows) {
		return user.User{}, user.ErrNotFound
	}
	if err != nil {
		return user.User{}, err
	}

	return user.User{
		UUID:              dbUser.UUID,
		Email:             dbUser.Email,
		PasswordHash:      dbUser.PasswordHash,
		RememberTokenHash: dbUser.RememberTokenHash,
		IsAdmin:           dbUser.IsAdmin,
	}, nil
}

func (r *Repository) GetUserByRememberTokenHash(rememberTokenHash string) (user.User, error) {
	dbUser := &User{}

	err := r.db.QueryRowx(
		"SELECT uuid, email, password_hash, remember_token_hash, is_admin FROM users WHERE remember_token_hash=$1",
		rememberTokenHash,
	).StructScan(dbUser)

	if errors.Is(err, sql.ErrNoRows) {
		return user.User{}, user.ErrNotFound
	}
	if err != nil {
		return user.User{}, err
	}

	return user.User{
		UUID:              dbUser.UUID,
		Email:             dbUser.Email,
		PasswordHash:      dbUser.PasswordHash,
		RememberTokenHash: dbUser.RememberTokenHash,
		IsAdmin:           dbUser.IsAdmin,
	}, nil
}

func (r *Repository) UpdateUser(u user.User) error {
	dbUser := User{
		UUID:              u.UUID,
		Email:             u.Email,
		PasswordHash:      u.PasswordHash,
		RememberTokenHash: u.RememberTokenHash,
		IsAdmin:           u.IsAdmin,
	}

	_, err := r.db.NamedExec(`UPDATE users
SET email=:email, password_hash=:password_hash, remember_token_hash=:remember_token_hash, is_admin=:is_admin
WHERE uuid=:uuid`,
		dbUser,
	)

	if err != nil {
		return err
	}

	return nil
}
