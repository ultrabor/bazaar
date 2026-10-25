package middleware

import (
	"bazaar/internal/domain"
	"context"
)

type contextKey uint8

const currentUserKey contextKey = iota

// WithCurrentUser returns a copy of ctx that carries u. Auth middleware uses
// it after loading the user; tests use it to skip token checks.
func WithCurrentUser(ctx context.Context, u *domain.User) context.Context {
	return context.WithValue(ctx, currentUserKey, u)
}

func CurrentUser(ctx context.Context) (*domain.User, bool) {
	if ctx == nil {
		return nil, false
	}

	user, ok := ctx.Value(currentUserKey).(*domain.User)
	if !ok {
		return nil, false
	}

	return user, true
}
