// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package pgtaxonomy

import (
	"context"
	"errors"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"

	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgbase"
	"github.com/virtualtam/sparklemuffin/pkg/taxonomy"
)

func dbTagToTag(dbTag DBTag) taxonomy.Tag {
	return taxonomy.Tag{
		UUID:      dbTag.TagUUID,
		UserUUID:  dbTag.UserUUID,
		Name:      dbTag.TagName,
		CreatedAt: dbTag.CreatedAt,
		UpdatedAt: dbTag.UpdatedAt,
	}
}

func tagGetQuery(ctx context.Context, q pgbase.Querier, query string, queryParams ...any) (taxonomy.Tag, error) {
	rows, err := q.Query(ctx, query, queryParams...)
	if err != nil {
		return taxonomy.Tag{}, err
	}
	defer rows.Close()

	dbTag := &DBTag{}
	err = pgxscan.ScanOne(dbTag, rows)

	if errors.Is(err, pgx.ErrNoRows) {
		return taxonomy.Tag{}, taxonomy.ErrNotFound
	}
	if err != nil {
		return taxonomy.Tag{}, err
	}

	return dbTagToTag(*dbTag), nil
}

func tagGetManyQuery(ctx context.Context, q pgbase.Querier, query string, queryParams ...any) ([]taxonomy.Tag, error) {
	rows, err := q.Query(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dbTags []DBTag

	if err := pgxscan.ScanAll(&dbTags, rows); err != nil {
		return nil, err
	}

	tags := make([]taxonomy.Tag, len(dbTags))

	for i, dbTag := range dbTags {
		tags[i] = dbTagToTag(dbTag)
	}

	return tags, nil
}
