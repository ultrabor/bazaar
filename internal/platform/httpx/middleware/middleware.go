package middleware

import "log/slog"

type Middleware struct {
	secretKey string
	logger    *slog.Logger
}

func New(logger *slog.Logger, secretKey string) *Middleware {
	return &Middleware{logger: logger, secretKey: secretKey}
}
