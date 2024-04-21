// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package postgresql

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// Repository provides a PostgreSQL persistence layer.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository initializes and returns a PostgreSQL Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) rollback(ctx context.Context, tx pgx.Tx, domain string, operation string) {
	err := tx.Rollback(ctx)
	if errors.Is(err, pgx.ErrTxClosed) {
		return
	}
	if err != nil {
		log.Error().
			Err(err).
			Str("domain", domain).
			Str("operation", operation).
			Msg("failed to rollback transaction")
		return
	}

	log.Warn().
		Err(err).
		Str("domain", domain).
		Str("operation", operation).
		Msg("transaction rolled back")
}

func (r *Repository) rowExistsByQuery(query string, queryParams ...any) (bool, error) {
	var exists int64

	err := r.pool.QueryRow(
		context.Background(),
		query,
		queryParams...,
	).Scan(&exists)

	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *Repository) add(domain string, operation string, query string, args pgx.NamedArgs) error {
	ctx := context.Background()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer r.rollback(ctx, tx, domain, operation)

	_, err = tx.Exec(ctx, query, args)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
