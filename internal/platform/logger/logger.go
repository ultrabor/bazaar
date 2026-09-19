package logger

import (
	"bazaar/internal/platform/config"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
)

const (
	LevelDebug = -4
	LevelInfo  = 0
	LevelWarn  = 4
	LevelError = 8
)

func LogLevel(level string) slog.Leveler {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// type Logger struct {
// 	h Handler
// }

type Handler struct {
	h slog.Handler
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.h.Enabled(ctx, level)
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{h: h.h.WithAttrs(attrs)}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{h: h.h.WithGroup(name)}
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {

	attrs := make(map[string]any, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})

	b, err := json.MarshalIndent(attrs, "", "  ")
	if err != nil {
		return fmt.Errorf("error when marshaling attrs: %w", err)
	}

	// h.h.
	fmt.Println(
		r.Time.Format(timeFormat),
		r.Level.String()+":",
		r.Message,
		string(b),
	)

	return nil
}

const (
	timeFormat = "[15:04:05.000]"
)

func New(cfg config.Config, serviceName string) *slog.Logger {

	b := &bytes.Buffer{}

	return slog.New(&Handler{
		h: slog.NewJSONHandler(b, &slog.HandlerOptions{
			Level: LogLevel(cfg.LogLevel),
		}),
	}).With(
		slog.String("service", serviceName),
		slog.String("env", cfg.AppEnv),
	)
}
