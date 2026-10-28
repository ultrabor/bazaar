package middleware

import (
	"bazaar/internal/platform/identity"
	"context"
	"log/slog"
)

// UserLoader returns the active user by id. It must return
// apperror.ErrInvalidCred when there is no such user.
type UserLoader func(ctx context.Context, userId string) (*identity.User, error)

// PermissionChecker reports whether the user holds the permission code.
type PermissionChecker func(ctx context.Context, userId, code string) (bool, error)

type Middleware struct {
	secretKey     string
	logger        *slog.Logger
	loadUser      UserLoader
	hasPermission PermissionChecker
}

func New(logger *slog.Logger, secretKey string, loadUser UserLoader, hasPermission PermissionChecker) *Middleware {
	return &Middleware{
		logger:        logger,
		secretKey:     secretKey,
		loadUser:      loadUser,
		hasPermission: hasPermission,
	}
}
