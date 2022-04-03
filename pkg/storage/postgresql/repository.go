package postgresql

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/virtualtam/yawbe/pkg/session"
	"github.com/virtualtam/yawbe/pkg/user"
)

var _ session.Repository = &Repository{}
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

func (r *Repository) SessionAdd(sess session.Session) error {
	dbSession := Session{
		UserUUID:               sess.UserUUID,
		RememberTokenHash:      sess.RememberTokenHash,
		RememberTokenExpiresAt: sess.RememberTokenExpiresAt,
	}

	_, err := r.db.NamedExec(
		`
INSERT INTO sessions(
	user_uuid,
	remember_token_hash,
	remember_token_expires_at
)
VALUES(
	:user_uuid,
	:remember_token_hash,
	:remember_token_expires_at
)`,
		dbSession,
	)

	return err
}

func (r *Repository) SessionGetByRememberTokenHash(hash string) (session.Session, error) {
	dbSession := &Session{}

	err := r.db.QueryRowx(
		`SELECT user_uuid, remember_token_hash
FROM sessions
WHERE remember_token_hash=$1`,
		hash,
	).StructScan(dbSession)

	if errors.Is(err, sql.ErrNoRows) {
		return session.Session{}, session.ErrNotFound
	}
	if err != nil {
		return session.Session{}, err
	}

	return session.Session{
		UserUUID:          dbSession.UserUUID,
		RememberTokenHash: dbSession.RememberTokenHash,
	}, nil
}

func (r *Repository) UserAdd(u user.User) error {
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

	return err
}

func (r *Repository) UserDeleteByUUID(userUUID string) error {
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

func (r *Repository) UserGetAll() ([]user.User, error) {
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

func (r *Repository) UserGetByEmail(email string) (user.User, error) {
	dbUser := &User{}

	err := r.db.QueryRowx(
		`SELECT uuid, email, password_hash, is_admin, created_at, updated_at
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
		UUID:         dbUser.UUID,
		Email:        dbUser.Email,
		PasswordHash: dbUser.PasswordHash,
		IsAdmin:      dbUser.IsAdmin,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
	}, nil
}

func (r *Repository) UserGetByUUID(userUUID string) (user.User, error) {
	dbUser := &User{}

	err := r.db.QueryRowx(
		`SELECT uuid, email, password_hash, is_admin, created_at, updated_at
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
		UUID:         dbUser.UUID,
		Email:        dbUser.Email,
		PasswordHash: dbUser.PasswordHash,
		IsAdmin:      dbUser.IsAdmin,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
	}, nil
}

func (r *Repository) UserIsEmailRegistered(email string) (bool, error) {
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

func (r *Repository) UserUpdate(u user.User) error {
	dbUser := User{
		UUID:         u.UUID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		IsAdmin:      u.IsAdmin,
		UpdatedAt:    u.UpdatedAt,
	}

	_, err := r.db.NamedExec(`UPDATE users
SET
	email=:email,
	password_hash=:password_hash,
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

func (r *Repository) UserUpdateInfo(info user.InfoUpdate) error {
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

func (r *Repository) UserUpdatePasswordHash(passwordHash user.PasswordHashUpdate) error {
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
