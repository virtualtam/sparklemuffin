// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package importing

import (
	"context"

	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
)

var _ bookmark.ValidationRepository = &validationRepository{}

type validationRepository struct{}

func (r *validationRepository) BookmarkIsURLRegistered(_ context.Context, userUUID, url string) (bool, error) {
	// Unicity checks for bulk operations must be handled by the persistence layer.
	return false, nil
}

func (r *validationRepository) BookmarkIsURLRegisteredToAnotherUID(_ context.Context, userUUID, url, uid string) (bool, error) {
	// Unicity checks for bulk operations must be handled by the persistence layer.
	return false, nil
}

type Repository interface {
	// BookmarkAddMany adds a collection of new bookmarks.
	BookmarkAddMany(ctx context.Context, bookmarks []bookmark.Bookmark) (int64, error)

	// BookmarkUpsertMany adds a collection of new bookmarks and updates
	// existing bookmarks in case of conflict.
	BookmarkUpsertMany(ctx context.Context, bookmarks []bookmark.Bookmark) (int64, error)
}
