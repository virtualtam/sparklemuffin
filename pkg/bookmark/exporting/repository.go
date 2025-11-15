// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package exporting

import (
	"context"

	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
)

type Repository interface {
	// BookmarkGetAll returns all bookmarks for a given user UUID.
	BookmarkGetAll(ctx context.Context, userUUID string) ([]bookmark.Bookmark, error)

	// BookmarkGetAllPrivate returns all private bookmarks for a given user UUID.
	BookmarkGetAllPrivate(ctx context.Context, userUUID string) ([]bookmark.Bookmark, error)

	// BookmarkGetAllPublic returns all public bookmarks for a given user UUID.
	BookmarkGetAllPublic(ctx context.Context, userUUID string) ([]bookmark.Bookmark, error)
}
