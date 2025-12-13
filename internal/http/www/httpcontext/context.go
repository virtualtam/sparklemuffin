// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

// Package httpcontext provides utilities to set and retrieve values from HTTP request context.
package httpcontext

import (
	"context"

	"github.com/virtualtam/sparklemuffin/pkg/user"
)

type contextKey string

const (
	cspNonceKey contextKey = "csp_nonce"
	userKey     contextKey = "user"
)

// WithCSPNonce enriches a context.Context with a nonce value used to enforce a Content Security Policy.
func WithCSPNonce(ctx context.Context, nonce string) context.Context {
	return context.WithValue(ctx, cspNonceKey, nonce)
}

// CSPNonceValue retrieves a nonce value used to enforce a Content Security Policy from a context.Context.
func CSPNonceValue(ctx context.Context) string {
	if value := ctx.Value(cspNonceKey); value != nil {
		if nonce, ok := value.(string); ok {
			return nonce
		}
	}

	return ""
}

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
