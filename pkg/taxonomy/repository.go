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

	// TagRename renames the tag with a given UUID for a given user.
	TagRename(ctx context.Context, userUUID, tagUUID, newName string) error
}
