// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package querying

import (
	"context"

	"github.com/virtualtam/sparklemuffin/pkg/bookmark"
)

// Repository provides access to query user bookmarks.
type Repository interface {
	// BookmarkGetCount returns the number of bookmarks for a given user.
	BookmarkGetCount(ctx context.Context, userUUID string, visibility Visibility) (uint, error)

	// BookmarkGetN returns at most n bookmarks for a given user, starting at
	// a given offset.
	BookmarkGetN(ctx context.Context, userUUID string, visibility Visibility, n uint, offset uint) ([]bookmark.Bookmark, error)

	// BookmarkGetPublicByUID returns the bookmark for a given user and UID, provided the bookmark is public.
	BookmarkGetPublicByUID(ctx context.Context, userUUID, uid string) (bookmark.Bookmark, error)

	// BookmarkSearchCount returns the number of bookmarks for a given user and
	// search terms.
	BookmarkSearchCount(ctx context.Context, userUUID string, visibility Visibility, searchTerms string) (uint, error)

	// BookmarkSearchN returns at most n bookmarks for a given user and search
	// terms, starting at a given offset.
	BookmarkSearchN(ctx context.Context, userUUID string, visibility Visibility, searchTerms string, n uint, offset uint) ([]bookmark.Bookmark, error)

	// OwnerGetByUUID returns the Owner corresponding to a given UUID.
	OwnerGetByUUID(ctx context.Context, uuid string) (Owner, error)

	// BookmarkTagGetAll returns all tags for a given user.
	BookmarkTagGetAll(ctx context.Context, userUUID string, visibility Visibility) ([]Tag, error)

	// BookmarkTagGetCount returns the number of tags for a given user.
	BookmarkTagGetCount(ctx context.Context, userUUID string, visibility Visibility) (uint, error)

	// BookmarkTagGetN returns at most n tags for a given user, starting at
	// a given offset.
	BookmarkTagGetN(ctx context.Context, userUUID string, visibility Visibility, n uint, offset uint) ([]Tag, error)

	// BookmarkTagFilterCount returns the number of tags for a given user and
	// filter term.
	BookmarkTagFilterCount(ctx context.Context, userUUID string, visibility Visibility, filterTerm string) (uint, error)

	// BookmarkTagFilterN returns at most n tags for a given user and filter term, starting at a given offset.
	BookmarkTagFilterN(ctx context.Context, userUUID string, visibility Visibility, filterTerm string, n uint, offset uint) ([]Tag, error)
}
