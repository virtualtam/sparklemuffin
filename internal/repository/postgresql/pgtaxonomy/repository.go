// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package pgtaxonomy

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/virtualtam/sparklemuffin/internal/repository/postgresql/pgbase"
	"github.com/virtualtam/sparklemuffin/pkg/taxonomy"
)

var _ taxonomy.Repository = &Repository{}

const domain = "taxonomy"

// OnTagMergeFn reassigns a domain's tag associations from one tag UUID to
// another, using q so the reassignment commits or rolls back together with
// MergeTag's own writes.
type OnTagMergeFn func(ctx context.Context, q pgbase.Querier, userUUID, oldTagUUID, newTagUUID string) error

type Repository struct {
	*pgbase.Repository

	onTagMergeFns []OnTagMergeFn
}

func NewRepository(pool *pgxpool.Pool, onTagMergeFns ...OnTagMergeFn) *Repository {
	return &Repository{
		Repository:    pgbase.NewRepository(pool),
		onTagMergeFns: onTagMergeFns,
	}
}

// TagAdd adds a new tag.
func (r *Repository) TagAdd(ctx context.Context, tag taxonomy.Tag) error {
	return r.TagAddTx(ctx, r.Pool, tag)
}

// TagAddTx adds a new tag using q, so callers can participate in an
// existing transaction (e.g. alongside another domain's own writes).
func (r *Repository) TagAddTx(ctx context.Context, q pgbase.Querier, tag taxonomy.Tag) error {
	query := `
	INSERT INTO taxonomy_tags(tag_uuid, user_uuid, tag_name, tag_name_tsv, created_at, updated_at)
	VALUES(@tag_uuid, @user_uuid, @tag_name, TO_TSVECTOR(@tag_name), @created_at, @updated_at)`

	args := pgx.NamedArgs{
		"tag_uuid":   tag.UUID,
		"user_uuid":  tag.UserUUID,
		"tag_name":   tag.Name,
		"created_at": tag.CreatedAt,
		"updated_at": tag.UpdatedAt,
	}

	_, err := q.Exec(ctx, query, args)
	return err
}

// TagDelete deletes the tag with a given name for a given user, if it exists.
func (r *Repository) TagDelete(ctx context.Context, userUUID, name string) error {
	_, err := r.Pool.Exec(
		ctx,
		`
		DELETE FROM taxonomy_tags
		WHERE user_uuid=$1
		AND   tag_name=$2`,
		userUUID,
		name,
	)

	return err
}

// TagGetByName returns the tag with a given name for a given user.
func (r *Repository) TagGetByName(ctx context.Context, userUUID, name string) (taxonomy.Tag, error) {
	return r.TagGetByNameTx(ctx, r.Pool, userUUID, name)
}

// TagGetByNameTx returns the tag with a given name for a given user, using q.
func (r *Repository) TagGetByNameTx(ctx context.Context, q pgbase.Querier, userUUID, name string) (taxonomy.Tag, error) {
	query := `
	SELECT tag_uuid, user_uuid, tag_name, created_at, updated_at
	FROM taxonomy_tags
	WHERE user_uuid=$1
	AND   tag_name=$2`

	return tagGetQuery(ctx, q, query, userUUID, name)
}

// TagGetByUUID returns the tag with a given UUID for a given user.
func (r *Repository) TagGetByUUID(ctx context.Context, userUUID, tagUUID string) (taxonomy.Tag, error) {
	query := `
	SELECT tag_uuid, user_uuid, tag_name, created_at, updated_at
	FROM taxonomy_tags
	WHERE user_uuid=$1
	AND   tag_uuid=$2`

	return tagGetQuery(ctx, r.Pool, query, userUUID, tagUUID)
}

// TagGetCount returns the number of tags for a given user.
func (r *Repository) TagGetCount(ctx context.Context, userUUID string) (uint, error) {
	var count uint

	err := r.Pool.QueryRow(
		ctx,
		"SELECT COUNT(*) FROM taxonomy_tags WHERE user_uuid=$1",
		userUUID,
	).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// MergeTag merges the tag with the given old UUID into the tag with the
// given new UUID for a given user, atomically: every registered
// OnTagMergeFn reassigns its domain's references to the new tag, then the
// old tag is deleted.
func (r *Repository) MergeTag(ctx context.Context, userUUID, oldTagUUID, newTagUUID string) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer r.Rollback(ctx, tx, domain, "MergeTag")

	for _, onTagMergeFn := range r.onTagMergeFns {
		if err := onTagMergeFn(ctx, tx, userUUID, oldTagUUID, newTagUUID); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(
		ctx,
		"DELETE FROM taxonomy_tags WHERE user_uuid=$1 AND tag_uuid=$2",
		userUUID,
		oldTagUUID,
	); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// TagGetN returns at most n tags for a given user, starting at a given offset.
func (r *Repository) TagGetN(ctx context.Context, userUUID string, n, offset uint) ([]taxonomy.Tag, error) {
	query := `
	SELECT tag_uuid, user_uuid, tag_name, created_at, updated_at
	FROM taxonomy_tags
	WHERE user_uuid=$1
	ORDER BY tag_name
	LIMIT $2 OFFSET $3`

	return tagGetManyQuery(ctx, r.Pool, query, userUUID, n, offset)
}

// TagNameIsRegistered returns whether a user already has a tag with a given name.
func (r *Repository) TagNameIsRegistered(ctx context.Context, userUUID, name string) (bool, error) {
	return r.RowExistsByQuery(
		ctx,
		"SELECT 1 FROM taxonomy_tags WHERE user_uuid=$1 AND tag_name=$2",
		userUUID,
		name,
	)
}

// TagRename renames the tag with a given UUID for a given user.
func (r *Repository) TagRename(ctx context.Context, userUUID, tagUUID, newName string) error {
	_, err := r.Pool.Exec(
		ctx,
		`
		UPDATE taxonomy_tags
		SET tag_name=$1, tag_name_tsv=TO_TSVECTOR($1), updated_at=NOW()
		WHERE user_uuid=$2
		AND   tag_uuid=$3`,
		newName,
		userUUID,
		tagUUID,
	)

	return err
}

// TagSearchCount returns the number of tags for a given user and search terms.
func (r *Repository) TagSearchCount(ctx context.Context, userUUID, searchTerms string) (uint, error) {
	var count uint

	err := r.Pool.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM taxonomy_tags
		WHERE user_uuid=$1
		AND   tag_name ILIKE $2`,
		userUUID,
		"%"+searchTerms+"%",
	).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// TagSearchN returns at most n tags for a given user and search terms, starting at a given offset.
func (r *Repository) TagSearchN(ctx context.Context, userUUID, searchTerms string, n, offset uint) ([]taxonomy.Tag, error) {
	query := `
	SELECT tag_uuid, user_uuid, tag_name, created_at, updated_at
	FROM taxonomy_tags
	WHERE user_uuid=$1
	AND   tag_name ILIKE $2
	ORDER BY tag_name
	LIMIT $3 OFFSET $4`

	return tagGetManyQuery(ctx, r.Pool, query, userUUID, "%"+searchTerms+"%", n, offset)
}
