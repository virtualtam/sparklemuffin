// Copyright (c) VirtualTam
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
}
