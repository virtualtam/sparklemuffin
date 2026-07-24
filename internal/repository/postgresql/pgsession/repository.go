// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package pgsession

import (
	"context"
	"errors"
	"time"

	"github.com/coder/quartz"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgbase"
	"github.com/virtualtam/sparklemuffin/pkg/session"
)

var _ session.Repository = &Repository{}

const (
	domain = "sessions"

	invalidationTaskInterval = 1 * time.Hour
	invalidationTaskTimeout  = 15 * time.Second
)

type Repository struct {
	pgbase.Repository

	clock quartz.Clock
}

// NewRepository initializes and returns a PostgreSQL Repository for the
// session domain, and starts a background task that periodically deletes
// expired Sessions.
func NewRepository(ctx context.Context, pool *pgxpool.Pool, clock quartz.Clock) *Repository {
	r := &Repository{
		Repository: *pgbase.NewRepository(pool),
		clock:      clock,
	}

	waiter := r.clock.TickerFunc(ctx, invalidationTaskInterval, r.invalidateExpiredSessions, "session-cleanup")
	go func() {
		if err := waiter.Wait(); err != nil {
			log.Error().Err(err).Msg("sessions: stopping cleanup task")
		}
	}()
	log.Info().
		Dur("interval", invalidationTaskInterval).
		Dur("timeout", invalidationTaskTimeout).
		Msgf("sessions: session expiration scheduler started")

	return r
}

// invalidateExpiredSessions deletes expired Sessions and is invoked on every cleanup interval.
func (r *Repository) invalidateExpiredSessions() error {
	ctx, cancel := context.WithTimeout(context.Background(), invalidationTaskTimeout)
	defer cancel()

	nExpired, err := r.deleteExpiredSessions(ctx)
	if err != nil {
		return err
	}

	log.Info().Int64("expired", nExpired).Msg("sessions: delete expired sessions")

	return nil
}

// deleteExpiredSessions deletes all Sessions whose RememberToken has expired
// and returns the number of deleted rows.
func (r *Repository) deleteExpiredSessions(ctx context.Context) (int64, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}

	defer r.Rollback(ctx, tx, domain, "deleteExpiredSessions")

	commandTag, err := tx.Exec(ctx, "DELETE FROM sessions WHERE remember_token_expires_at <= NOW()")
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return commandTag.RowsAffected(), nil
}

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
