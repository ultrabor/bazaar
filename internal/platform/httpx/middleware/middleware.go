package middleware

import (
	"bazaar/internal/modules/user"
	"context"
	"log/slog"
)

type UserLoader func(context.Context, string) (*user.User, error)

type Middleware struct {
	secretKey string
	logger    *slog.Logger
	loadUser  UserLoader
}

func New(logger *slog.Logger, secretKey string, loadUser UserLoader) *Middleware {
	return &Middleware{
		logger:    logger,
		secretKey: secretKey,
		loadUser:  loadUser,
	}
}
