// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package exporting

import "github.com/virtualtam/sparklemuffin/pkg/bookmark"

type Repository interface {
	// BookmarkGetAll returns all bookmarks for a given user UUID.
	BookmarkGetAll(userUUID string) ([]bookmark.Bookmark, error)

	// BookmarkGetAllPrivate returns all private bookmarks for a given user UUID.
	BookmarkGetAllPrivate(userUUID string) ([]bookmark.Bookmark, error)

	// BookmarkGetAllPublic returns all public bookmarks for a given user UUID.
	BookmarkGetAllPublic(userUUID string) ([]bookmark.Bookmark, error)
}
