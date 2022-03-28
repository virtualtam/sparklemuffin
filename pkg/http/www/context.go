package www

import (
	"context"

	"github.com/virtualtam/yawbe/pkg/user"
)

type userContextKey string

const (
	userKey userContextKey = "user"
)

// withUser enriches a context.Context with a User object.
func withUser(ctx context.Context, user user.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}
