package www

import (
	"context"

	"github.com/virtualtam/yawbe/pkg/user"
)

type userContextKey string

const (
	userKey userContextKey = "user"
)

// withUser enriches a context.Context with a user.User.
func withUser(ctx context.Context, user user.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// userValue retrieves a user.User from a context.Context.
func userValue(ctx context.Context) *user.User {
	if value := ctx.Value(userKey); value != nil {
		if user, ok := value.(user.User); ok {
			return &user
		}
	}

	return nil
}
