package middleware

import (
	"bazaar/internal/modules/permission"
	"bazaar/internal/modules/user"
	"context"
	"log/slog"
)

type UserLoader func(context.Context, string) (*user.User, error)
type PermissionLoader func(context.Context, string, string) (*permission.Permission, error)

type Middleware struct {
	secretKey     string
	logger        *slog.Logger
	loadUser      UserLoader
	loadPermision PermissionLoader
}

func New(logger *slog.Logger, secretKey string, loadUser UserLoader, loadPermision PermissionLoader) *Middleware {
	return &Middleware{
		logger:        logger,
		secretKey:     secretKey,
		loadUser:      loadUser,
		loadPermision: loadPermision,
	}
}
