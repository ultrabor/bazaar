package middleware

import (
	"bazaar/internal/modules/user"
	"context"
)

type contextKey uint8

const currentUserKey contextKey = iota

func CurrentUser(ctx context.Context) (*user.User, bool) {
	if ctx == nil {
		return nil, false
	}

	user, ok := ctx.Value(currentUserKey).(*user.User)
	if !ok {
		return nil, false
	}

	return user, true
}
