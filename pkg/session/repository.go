// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

package session

import (
	"context"
)

// Repository provides access to users' Web Session.
type Repository interface {
	// SessionAdd saves a new user Session.
	SessionAdd(ctx context.Context, s Session) error

	// SessionGetByRememberTokenHash returns the Session corresponding to a
	// given remember token hash.
	SessionGetByRememberTokenHash(ctx context.Context, hash string) (Session, error)

	// SessionDeleteByRememberTokenHash deletes the Session corresponding to a
	// given remember token hash.
	SessionDeleteByRememberTokenHash(ctx context.Context, hash string) error

	// SessionDeleteByUserUUID deletes all Sessions belonging to a given user.
	SessionDeleteByUserUUID(ctx context.Context, userUUID string) error
}
