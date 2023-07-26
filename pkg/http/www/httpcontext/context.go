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

// httpcontext.UserValue retrieves a user.User from a context.Context.
func UserValue(ctx context.Context) *user.User {
	if value := ctx.Value(userKey); value != nil {
		if user, ok := value.(user.User); ok {
			return &user
		}
	}

	return nil
}
