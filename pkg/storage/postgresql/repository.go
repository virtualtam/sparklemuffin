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
		UUID:         u.UUID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		IsAdmin:      u.IsAdmin,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}

	_, err := r.db.NamedExec(
		`
INSERT INTO users(
	uuid,
	email,
	password_hash,
	is_admin,
	created_at,
	updated_at
)
VALUES(
	:uuid,
	:email,
	:password_hash,
	:is_admin,
	:created_at,
	:updated_at
)`,
		dbUser,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) DeleteUserByUUID(userUUID string) error {
	result, err := r.db.Exec("DELETE FROM users WHERE uuid=$1", userUUID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return user.ErrNotFound
	}

	return nil
}

func (r *Repository) GetAllUsers() ([]user.User, error) {
	rows, err := r.db.Queryx("SELECT uuid, email, is_admin, created_at, updated_at FROM users")
	if err != nil {
		return []user.User{}, err
	}

	users := []user.User{}

	for rows.Next() {
		dbUser := User{}

		if err := rows.StructScan(&dbUser); err != nil {
			return []user.User{}, err
		}

		user := user.User{
			UUID:      dbUser.UUID,
			Email:     dbUser.Email,
			IsAdmin:   dbUser.IsAdmin,
			CreatedAt: dbUser.CreatedAt,
			UpdatedAt: dbUser.UpdatedAt,
		}

		users = append(users, user)
	}

	return users, nil
}

func (r *Repository) GetUserByEmail(email string) (user.User, error) {
	dbUser := &User{}

	err := r.db.QueryRowx(
		`SELECT uuid, email, password_hash, remember_token_hash, is_admin, created_at, updated_at
FROM users
WHERE email=$1`,
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
		CreatedAt:         dbUser.CreatedAt,
		UpdatedAt:         dbUser.UpdatedAt,
	}, nil
}

func (r *Repository) GetUserByRememberTokenHash(rememberTokenHash string) (user.User, error) {
	dbUser := &User{}

	err := r.db.QueryRowx(
		`SELECT uuid, email, password_hash, remember_token_hash, is_admin, created_at, updated_at
FROM users
WHERE remember_token_hash=$1`,
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
		CreatedAt:         dbUser.CreatedAt,
		UpdatedAt:         dbUser.UpdatedAt,
	}, nil
}

func (r *Repository) GetUserByUUID(userUUID string) (user.User, error) {
	dbUser := &User{}

	err := r.db.QueryRowx(
		`SELECT uuid, email, password_hash, remember_token_hash, is_admin, created_at, updated_at
FROM users
WHERE uuid=$1`,
		userUUID,
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
		CreatedAt:         dbUser.CreatedAt,
		UpdatedAt:         dbUser.UpdatedAt,
	}, nil
}

func (r *Repository) IsUserEmailRegistered(email string) (bool, error) {
	dbUser := &User{}

	err := r.db.QueryRowx("SELECT email FROM users WHERE email=$1", email).StructScan(dbUser)

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *Repository) UpdateUser(u user.User) error {
	dbUser := User{
		UUID:              u.UUID,
		Email:             u.Email,
		PasswordHash:      u.PasswordHash,
		RememberTokenHash: u.RememberTokenHash,
		IsAdmin:           u.IsAdmin,
		UpdatedAt:         u.UpdatedAt,
	}

	_, err := r.db.NamedExec(`UPDATE users
SET
	email=:email,
	password_hash=:password_hash,
	remember_token_hash=:remember_token_hash,
	is_admin=:is_admin,
	updated_at=:updated_at
WHERE uuid=:uuid`,
		dbUser,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) UpdateUserInfo(info user.InfoUpdate) error {
	dbUser := User{
		UUID:      info.UUID,
		Email:     info.Email,
		UpdatedAt: info.UpdatedAt,
	}

	_, err := r.db.NamedExec(`UPDATE users
SET
	email=:email,
	updated_at=:updated_at
WHERE uuid=:uuid`,
		dbUser,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) UpdateUserPasswordHash(passwordHash user.PasswordHashUpdate) error {
	dbUser := User{
		UUID:         passwordHash.UUID,
		PasswordHash: passwordHash.PasswordHash,
		UpdatedAt:    passwordHash.UpdatedAt,
	}

	_, err := r.db.NamedExec(`UPDATE users
SET
	password_hash=:password_hash,
	updated_at=:updated_at
WHERE uuid=:uuid`,
		dbUser,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) UpdateUserRememberTokenHash(u user.User) error {
	dbUser := User{
		UUID:              u.UUID,
		RememberTokenHash: u.RememberTokenHash,
	}

	_, err := r.db.NamedExec(`UPDATE users
SET	remember_token_hash=:remember_token_hash
WHERE uuid=:uuid`,
		dbUser,
	)

	if err != nil {
		return err
	}

	return nil
}
