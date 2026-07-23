// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package pgsession

import (
	"context"
	"errors"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgbase"
	"github.com/virtualtam/sparklemuffin/pkg/session"
)

var _ session.Repository = &Repository{}

type Repository struct {
	pgbase.Repository
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		Repository: *pgbase.NewRepository(pool),
	}
}

const (
	domain = "sessions"
)

func (r *Repository) SessionAdd(ctx context.Context, sess session.Session) error {
	query := `
	INSERT INTO sessions(
		user_uuid,
		remember_token_hash,
		remember_token_expires_at
	)
	VALUES(
		@user_uuid,
		@remember_token_hash,
		@remember_token_expires_at
	)`

	args := pgx.NamedArgs{
		"user_uuid":                 sess.UserUUID,
		"remember_token_hash":       sess.RememberTokenHash,
		"remember_token_expires_at": sess.RememberTokenExpiresAt,
	}

	return r.QueryTx(ctx, domain, "SessionAdd", query, args)
}

func (r *Repository) SessionGetByRememberTokenHash(ctx context.Context, hash string) (session.Session, error) {
	query := `
	SELECT user_uuid, remember_token_hash, remember_token_expires_at
	FROM sessions
	WHERE remember_token_hash=$1
	AND remember_token_expires_at > NOW()`

	dbSession := &DBSession{}

	rows, err := r.Pool.Query(
		ctx,
		query,
		hash,
	)
	if err != nil {
		return session.Session{}, err
	}
	defer rows.Close()

	err = pgxscan.ScanOne(dbSession, rows)

	if errors.Is(err, pgx.ErrNoRows) {
		return session.Session{}, session.ErrNotFound
	}
	if err != nil {
		return session.Session{}, err
	}

	return session.Session{
		UserUUID:               dbSession.UserUUID,
		RememberTokenHash:      dbSession.RememberTokenHash,
		RememberTokenExpiresAt: dbSession.RememberTokenExpiresAt,
	}, nil
}

func (r *Repository) SessionDeleteByRememberTokenHash(ctx context.Context, hash string) error {
	query := `DELETE FROM sessions WHERE remember_token_hash=@remember_token_hash`

	args := pgx.NamedArgs{
		"remember_token_hash": hash,
	}

	return r.QueryTx(ctx, domain, "SessionDeleteByRememberTokenHash", query, args)
}

func (r *Repository) SessionDeleteByUserUUID(ctx context.Context, userUUID string) error {
	query := `DELETE FROM sessions WHERE user_uuid=@user_uuid`

	args := pgx.NamedArgs{
		"user_uuid": userUUID,
	}

	return r.QueryTx(ctx, domain, "SessionDeleteByUserUUID", query, args)
}
