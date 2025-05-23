// Copyright (c) VirtualTam
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

func (r *Repository) SessionAdd(sess session.Session) error {
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

	return r.QueryTx("sessions", "SessionAdd", query, args)
}

func (r *Repository) SessionGetByRememberTokenHash(hash string) (session.Session, error) {
	query := `
	SELECT user_uuid, remember_token_hash
	FROM sessions
	WHERE remember_token_hash=$1`

	dbSession := &DBSession{}

	rows, err := r.Pool.Query(
		context.Background(),
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
		UserUUID:          dbSession.UserUUID,
		RememberTokenHash: dbSession.RememberTokenHash,
	}, nil
}
