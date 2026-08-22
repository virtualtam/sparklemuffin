// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package pgbase

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// A Querier allows repository helpers to run a query against either a pooled connection
// or an existing transaction, so writes spanning more than one domain's tables can commit
// atomically.
type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

var (
	_ Querier = (pgx.Tx)(nil)
	_ Querier = (*pgxpool.Pool)(nil)
	_ Querier = (*pgxpool.Tx)(nil)
)
