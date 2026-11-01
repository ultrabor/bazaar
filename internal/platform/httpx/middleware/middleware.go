package middleware

import (
	"bazaar/internal/platform/identity"
	"context"
	"log/slog"
)

// UserLoader returns the active user of a bearer token. It must return
// apperror.ErrInvalidCred when the session is missing, expired or revoked.
type UserLoader func(ctx context.Context, token string) (*identity.User, error)

// PermissionChecker reports whether the user holds the permission code.
type PermissionChecker func(ctx context.Context, userId, code string) (bool, error)

type Middleware struct {
	logger        *slog.Logger
	loadUser      UserLoader
	hasPermission PermissionChecker
}

func New(logger *slog.Logger, loadUser UserLoader, hasPermission PermissionChecker) *Middleware {
	return &Middleware{
		logger:        logger,
		loadUser:      loadUser,
		hasPermission: hasPermission,
	}
}
