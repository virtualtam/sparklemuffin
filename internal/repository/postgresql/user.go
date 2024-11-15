// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package postgresql

import (
	"context"
	"errors"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/virtualtam/sparklemuffin/pkg/user"
)

type DBUser struct {
	UUID         string `db:"uuid"`
	Email        string `db:"email"`
	NickName     string `db:"nick_name"`
	DisplayName  string `db:"display_name"`
	PasswordHash string `db:"password_hash"`
	IsAdmin      bool   `db:"is_admin"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

var _ user.Repository = &Repository{}

func (r *Repository) UserAdd(u user.User) error {
	query := `
	INSERT INTO users(
		uuid,
		email,
		nick_name,
		display_name,
		password_hash,
		is_admin,
		created_at,
		updated_at
	)
	VALUES(
		@uuid,
		@email,
		@nick_name,
		@display_name,
		@password_hash,
		@is_admin,
		@created_at,
		@updated_at
	)`

	args := pgx.NamedArgs{
		"uuid":          u.UUID,
		"email":         u.Email,
		"nick_name":     u.NickName,
		"display_name":  u.DisplayName,
		"password_hash": u.PasswordHash,
		"is_admin":      u.IsAdmin,
		"created_at":    u.CreatedAt,
		"updated_at":    u.UpdatedAt,
	}

	return r.queryTx("users", "UserAdd", query, args)
}

func (r *Repository) UserDeleteByUUID(userUUID string) error {
	ctx := context.Background()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer r.rollback(ctx, tx, "users", "delete")

	commandTag, err := tx.Exec(
		context.Background(),
		"DELETE FROM users WHERE uuid=$1",
		userUUID,
	)
	if err != nil {
		return err
	}

	rowsAffected := commandTag.RowsAffected()

	if rowsAffected != 1 {
		return user.ErrNotFound
	}

	return tx.Commit(ctx)
}

func (r *Repository) UserGetAll() ([]user.User, error) {
	query := `
	SELECT uuid, email, nick_name, display_name, is_admin, created_at, updated_at
	FROM users`

	rows, err := r.pool.Query(context.Background(), query)
	if err != nil {
		return []user.User{}, err
	}
	defer rows.Close()

	var dbUsers []DBUser

	if err := pgxscan.ScanAll(&dbUsers, rows); err != nil {
		return []user.User{}, err
	}

	var users []user.User

	for _, dbUser := range dbUsers {
		user := user.User{
			UUID:        dbUser.UUID,
			Email:       dbUser.Email,
			NickName:    dbUser.NickName,
			DisplayName: dbUser.DisplayName,
			IsAdmin:     dbUser.IsAdmin,
			CreatedAt:   dbUser.CreatedAt,
			UpdatedAt:   dbUser.UpdatedAt,
		}

		users = append(users, user)
	}

	return users, nil
}

func (r *Repository) userGetByQuery(query string, queryParams ...any) (user.User, error) {
	rows, err := r.pool.Query(
		context.Background(),
		query,
		queryParams...,
	)
	if err != nil {
		return user.User{}, err
	}
	defer rows.Close()

	dbUser := DBUser{}

	err = pgxscan.ScanOne(&dbUser, rows)

	if errors.Is(err, pgx.ErrNoRows) {
		return user.User{}, user.ErrNotFound
	}
	if err != nil {
		return user.User{}, err
	}

	return user.User{
		UUID:         dbUser.UUID,
		Email:        dbUser.Email,
		NickName:     dbUser.NickName,
		DisplayName:  dbUser.DisplayName,
		PasswordHash: dbUser.PasswordHash,
		IsAdmin:      dbUser.IsAdmin,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
	}, nil
}

func (r *Repository) UserGetByEmail(email string) (user.User, error) {
	query := `
	SELECT uuid, email, nick_name, display_name, password_hash, is_admin, created_at, updated_at
	FROM users
	WHERE email=$1`

	return r.userGetByQuery(query, email)
}

func (r *Repository) UserGetByNickName(nick string) (user.User, error) {
	query := `
	SELECT uuid, email, nick_name, display_name, password_hash, is_admin, created_at, updated_at
	FROM users
	WHERE nick_name=$1`

	return r.userGetByQuery(query, nick)
}

func (r *Repository) UserGetByUUID(userUUID string) (user.User, error) {
	query := `
	SELECT uuid, email, nick_name, display_name, password_hash, is_admin, created_at, updated_at
	FROM users
	WHERE uuid=$1`

	return r.userGetByQuery(query, userUUID)
}

func (r *Repository) UserIsEmailRegistered(email string) (bool, error) {
	return r.rowExistsByQuery(
		"SELECT 1 FROM users WHERE email=$1",
		email,
	)
}

func (r *Repository) UserIsNickNameRegistered(nick string) (bool, error) {
	return r.rowExistsByQuery(
		"SELECT 1 FROM users WHERE nick_name=$1",
		nick,
	)
}

func (r *Repository) UserUpdate(u user.User) error {
	query := `
	UPDATE users
	SET
		email=@email,
		nick_name=@nick_name,
		display_name=@display_name,
		password_hash=@password_hash,
		is_admin=@is_admin,
		updated_at=@updated_at
	WHERE uuid=@uuid`

	args := pgx.NamedArgs{
		"uuid":          u.UUID,
		"email":         u.Email,
		"nick_name":     u.NickName,
		"display_name":  u.DisplayName,
		"password_hash": u.PasswordHash,
		"is_admin":      u.IsAdmin,
		"updated_at":    u.UpdatedAt,
	}

	return r.queryTx("users", "UserUpdate", query, args)
}

func (r *Repository) UserUpdateInfo(info user.InfoUpdate) error {
	query := `
	UPDATE users
	SET
		email=@email,
		nick_name=@nick_name,
		display_name=@display_name,
		updated_at=@updated_at
	WHERE uuid=@uuid`

	args := pgx.NamedArgs{
		"uuid":         info.UUID,
		"email":        info.Email,
		"nick_name":    info.NickName,
		"display_name": info.DisplayName,
		"updated_at":   info.UpdatedAt,
	}

	return r.queryTx("users", "UserUpdateInfo", query, args)
}

func (r *Repository) UserUpdatePasswordHash(passwordHash user.PasswordHashUpdate) error {
	query := `
	UPDATE users
	SET
		password_hash=@password_hash,
		updated_at=@updated_at
	WHERE uuid=@uuid`

	args := pgx.NamedArgs{
		"uuid":          passwordHash.UUID,
		"password_hash": passwordHash.PasswordHash,
		"updated_at":    passwordHash.UpdatedAt,
	}

	return r.queryTx("users", "UserUpdatePasswordHash", query, args)
}
