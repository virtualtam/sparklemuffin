// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package httpcontext

import (
	"context"

	"github.com/virtualtam/sparklemuffin/pkg/user"
)

type userContextKey string

const (
	userKey userContextKey = "user"
)

// WithUser enriches a context.Context with a user.User.
func WithUser(ctx context.Context, user user.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// UserValue retrieves a user.User from a context.Context.
func UserValue(ctx context.Context) *user.User {
	if value := ctx.Value(userKey); value != nil {
		if ctxUser, ok := value.(user.User); ok {
			return &ctxUser
		}
	}

	return nil
}
