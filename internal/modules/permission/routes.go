package permission

import (
	"bazaar/internal/platform/httpx/middleware"
	"context"
	"log/slog"

	"github.com/go-chi/chi/v5"
)

// permissionService is what the handlers call; Service implements it.
type permissionService interface {
	// GetUserPermissions returns the viewer's own permissions when userId is
	// empty or the viewer's id. For another user the viewer needs
	// permissions.view (else ErrForbidden) and the user must belong to
	// companyId (else ErrUserNotFound).
	GetUserPermissions(ctx context.Context, viewerId, companyId, userId string) ([]Permission, error)
	// SetPermission and RemovePermission return ErrPermissionNotOwned when
	// the guarantor does not hold the permission.
	SetPermission(ctx context.Context, rq ChangePermissionRequest) error
	RemovePermission(ctx context.Context, rq ChangePermissionRequest) error
}

type Handler struct {
	logger *slog.Logger
	svc    permissionService
}

func NewHandler(logger *slog.Logger, svc permissionService) *Handler {
	return &Handler{logger: logger, svc: svc}
}

func (h *Handler) Routes(r chi.Router, mw *middleware.Middleware) {
	r.Group(func(r chi.Router) {
		r.Use(mw.Auth)

		// Own permissions are always visible; the service checks
		// permissions.view when ?user_id points to another user.
		r.Get("/permissions", h.GetListPermissions)
		r.Post("/permission/set", mw.Permission(h.SetPermission, "permissions.grant"))
		r.Post("/permission/remove", mw.Permission(h.RemovePermission, "permissions.revoke"))
	})
}
