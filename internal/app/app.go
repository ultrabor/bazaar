// Package app is the composition root of the API: it wires the modules
// together and builds the HTTP router. Adding a module means one New call
// here and one argument to NewRouter.
package app

import (
	"bazaar/internal/modules/auth"
	"bazaar/internal/modules/category"
	"bazaar/internal/modules/location"
	"bazaar/internal/modules/permission"
	"bazaar/internal/modules/product"
	"bazaar/internal/modules/role"
	"bazaar/internal/modules/user"
	"bazaar/internal/platform/config"
	"bazaar/internal/platform/database"
	"bazaar/internal/platform/health"
	"bazaar/internal/platform/httpx/middleware"
	"log/slog"
	"net/http"
)

// New wires every module on db and returns the API handler.
func New(cfg config.Config, db *database.Database, logger *slog.Logger) http.Handler {
	roles := role.New(db, logger)
	users := user.New(db, logger, roles.Service())
	permissions := permission.New(db, logger, users.Service())
	authModule := auth.New(db, logger, roles.Service(), []byte(cfg.JWTSecret))

	mw := middleware.New(logger, cfg.JWTSecret, authModule.LoadUser, permissions.HasPermission)
	probes := health.NewHandler(logger, health.New(db))

	return NewRouter(mw, probes,
		authModule,
		users,
		location.New(db, logger),
		product.New(db, logger),
		category.New(db, logger),
		permissions,
		roles,
	)
}
