package role

import (
	"bazaar/internal/platform/httpx/middleware"
	"context"
	"log/slog"

	"github.com/go-chi/chi/v5"
)

// roleService is what the handlers call; Service implements it.
type roleService interface {
	ListRoles(ctx context.Context, companyId string) ([]RoleSummary, error)
	GetRole(ctx context.Context, roleId, companyId string) (*RoleDetails, error)
	CreateRole(ctx context.Context, rq CreateRoleRequest) (*CreateRoleResponse, error)
	// RenameRole returns ErrSystemRole for the owner role.
	RenameRole(ctx context.Context, rq UpdateRoleRequest) error
	// ArchiveRole returns ErrSystemRole for the owner role and ErrRoleInUse
	// while active users hold the role.
	ArchiveRole(ctx context.Context, roleId, companyId string) error
	// AddRolePermission and RemoveRolePermission change the role package and,
	// in the same transaction, the permissions of every user of the role.
	AddRolePermission(ctx context.Context, rq RolePermissionRequest) error
	RemoveRolePermission(ctx context.Context, rq RolePermissionRequest) error
	// AssignRole checks that the user is in the company and not the owner,
	// that the role is assignable and not wider than the assigner. Then, in
	// one transaction, it takes the old package from the user, sets the role
	// and gives the new package; personal permissions outside the old
	// package stay.
	AssignRole(ctx context.Context, rq AssignRoleRequest) error
}

type Handler struct {
	logger *slog.Logger
	svc    roleService
}

func NewHandler(logger *slog.Logger, svc roleService) *Handler {
	return &Handler{logger: logger, svc: svc}
}

func (h *Handler) Routes(r chi.Router, mw *middleware.Middleware) {
	r.Group(func(r chi.Router) {
		r.Use(mw.Auth)

		r.Get("/role", mw.Permission(h.GetRoles, "roles.view"))
		r.Get("/role/{id}", mw.Permission(h.GetRoleById, "roles.view"))
		r.Post("/role", mw.Permission(h.CreateRole, "roles.create"))
		r.Put("/role/{id}", mw.Permission(h.UpdateRole, "roles.edit"))
		r.Patch("/role/{id}/archive", mw.Permission(h.ArchiveRole, "roles.archive"))
		r.Post("/role/{id}/permission", mw.Permission(h.AddRolePermission, "roles.edit"))
		r.Delete("/role/{id}/permission/{permissionId}", mw.Permission(h.RemoveRolePermission, "roles.edit"))
		r.Put("/user/{id}/role", mw.Permission(h.AssignRole, "roles.assign"))
	})
}
