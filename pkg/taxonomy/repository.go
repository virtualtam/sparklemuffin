// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package taxonomy

import "context"

// ValidationRepository provides methods for Tag validation.
type ValidationRepository interface {
	// TagNameIsRegistered returns whether a user already has a tag with a given name.
	TagNameIsRegistered(ctx context.Context, userUUID, name string) (bool, error)
}

// Repository provides access to persisted Tags.
type Repository interface {
	ValidationRepository

	// TagAdd adds a new tag.
	TagAdd(ctx context.Context, tag Tag) error

	// TagDelete deletes the tag with a given name for a given user, if it exists.
	TagDelete(ctx context.Context, userUUID, name string) error

	// TagGetByName returns the tag with a given name for a given user.
	TagGetByName(ctx context.Context, userUUID, name string) (Tag, error)

	// TagGetCount returns the number of tags for a given user.
	TagGetCount(ctx context.Context, userUUID string) (uint, error)

	// TagGetN returns at most n tags for a given user, starting at a given offset.
	TagGetN(ctx context.Context, userUUID string, n, offset uint) ([]Tag, error)

	// TagRename renames the tag with a given UUID for a given user.
	TagRename(ctx context.Context, userUUID, tagUUID, newName string) error

	// TagSearchCount returns the number of tags for a given user and search terms.
	TagSearchCount(ctx context.Context, userUUID, searchTerms string) (uint, error)

	// TagSearchN returns at most n tags for a given user and search terms, starting at a given offset.
	TagSearchN(ctx context.Context, userUUID, searchTerms string, n, offset uint) ([]Tag, error)
}
