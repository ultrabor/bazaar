package httpx

import (
	"bazaar/internal/domain"
	"bazaar/internal/platform/httpx/dto"
	"bazaar/internal/platform/httpx/middleware"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// @Summary List roles of the current user's company
// @Tags Role
// @Produce json
// @Security BearerAuth
// @Success 200 {array} domain.RoleSummary
// @Failure 401 {object} dto.Error "Unauthorized"
// @Failure 500 {object} dto.Error "Internal error"
// @Failure 503 {object} dto.Error "Service unavailable"
// @Router /role [get]
func (h *Handler) GetRoles(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	u, ok := currentUser(w, ctx)
	if !ok {
		return
	}

	res, err := h.sm.roleService.ListRoles(ctx, u.CompanyId)
	if err != nil {
		h.writeRoleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, res)
}

// @Summary Get a role with its permissions
// @Tags Role
// @Produce json
// @Security BearerAuth
// @Param id path string true "Role ID"
// @Success 200 {object} domain.RoleDetails
// @Failure 401 {object} dto.Error "Unauthorized"
// @Failure 404 {object} dto.Error "Role not found"
// @Failure 500 {object} dto.Error "Internal error"
// @Failure 503 {object} dto.Error "Service unavailable"
// @Router /role/{id} [get]
func (h *Handler) GetRoleById(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	u, ok := currentUser(w, ctx)
	if !ok {
		return
	}

	res, err := h.sm.roleService.GetRole(ctx, chi.URLParam(r, "id"), u.CompanyId)
	if err != nil {
		h.writeRoleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, res)
}

// @Summary Create a role without permissions
// @Tags Role
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateRoleRequest true "Role data"
// @Success 201 {object} domain.CreateRoleResponse
// @Failure 400 {object} dto.Error "Invalid JSON or role name"
// @Failure 401 {object} dto.Error "Unauthorized"
// @Failure 409 {object} dto.Error "Active role name already exists"
// @Failure 500 {object} dto.Error "Internal error"
// @Failure 503 {object} dto.Error "Service unavailable"
// @Router /role [post]
func (h *Handler) CreateRole(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var body dto.CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		dto.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON")
		return
	}
	u, ok := currentUser(w, ctx)
	if !ok {
		return
	}

	res, err := h.sm.roleService.CreateRole(ctx, domain.CreateRoleRequest{CompanyId: u.CompanyId, Name: body.Name})
	if err != nil {
		h.writeRoleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, res)
}

// @Summary Rename a role
// @Tags Role
// @Accept json
// @Security BearerAuth
// @Param id path string true "Role ID"
// @Param request body dto.UpdateRoleRequest true "Role data"
// @Success 204
// @Failure 400 {object} dto.Error "Invalid JSON or role name"
// @Failure 401 {object} dto.Error "Unauthorized"
// @Failure 403 {object} dto.Error "System role"
// @Failure 404 {object} dto.Error "Role not found"
// @Failure 409 {object} dto.Error "Active role name already exists"
// @Failure 500 {object} dto.Error "Internal error"
// @Failure 503 {object} dto.Error "Service unavailable"
// @Router /role/{id} [put]
func (h *Handler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var body dto.UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		dto.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON")
		return
	}
	u, ok := currentUser(w, ctx)
	if !ok {
		return
	}

	err := h.sm.roleService.RenameRole(ctx, domain.UpdateRoleRequest{
		RoleId:    chi.URLParam(r, "id"),
		CompanyId: u.CompanyId,
		Name:      body.Name,
	})
	if err != nil {
		h.writeRoleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Archive a role
// @Tags Role
// @Security BearerAuth
// @Param id path string true "Role ID"
// @Success 204
// @Failure 401 {object} dto.Error "Unauthorized"
// @Failure 403 {object} dto.Error "System role"
// @Failure 404 {object} dto.Error "Role not found"
// @Failure 409 {object} dto.Error "Role is assigned to active users"
// @Failure 500 {object} dto.Error "Internal error"
// @Failure 503 {object} dto.Error "Service unavailable"
// @Router /role/{id}/archive [patch]
func (h *Handler) ArchiveRole(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	u, ok := currentUser(w, ctx)
	if !ok {
		return
	}

	if err := h.sm.roleService.ArchiveRole(ctx, chi.URLParam(r, "id"), u.CompanyId); err != nil {
		h.writeRoleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Add a permission to a role package
// @Description Users of the role get the permission too.
// @Tags Role
// @Accept json
// @Security BearerAuth
// @Param id path string true "Role ID"
// @Param request body dto.AddRolePermissionRequest true "Permission"
// @Success 204
// @Failure 400 {object} dto.Error "Invalid JSON"
// @Failure 401 {object} dto.Error "Unauthorized"
// @Failure 403 {object} dto.Error "System role or permission not held"
// @Failure 404 {object} dto.Error "Role not found"
// @Failure 500 {object} dto.Error "Internal error"
// @Failure 503 {object} dto.Error "Service unavailable"
// @Router /role/{id}/permission [post]
func (h *Handler) AddRolePermission(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var body dto.AddRolePermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		dto.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON")
		return
	}
	u, ok := currentUser(w, ctx)
	if !ok {
		return
	}

	err := h.sm.roleService.AddRolePermission(ctx, domain.RolePermissionRequest{
		RoleId:       chi.URLParam(r, "id"),
		CompanyId:    u.CompanyId,
		PermissionId: body.PermissionId,
		GrantorId:    u.Id,
	})
	if err != nil {
		h.writeRoleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Remove a permission from a role package
// @Description The permission is taken from users of the role too.
// @Tags Role
// @Security BearerAuth
// @Param id path string true "Role ID"
// @Param permissionId path string true "Permission ID"
// @Success 204
// @Failure 401 {object} dto.Error "Unauthorized"
// @Failure 403 {object} dto.Error "System role or permission not held"
// @Failure 404 {object} dto.Error "Role or permission not found"
// @Failure 500 {object} dto.Error "Internal error"
// @Failure 503 {object} dto.Error "Service unavailable"
// @Router /role/{id}/permission/{permissionId} [delete]
func (h *Handler) RemoveRolePermission(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	u, ok := currentUser(w, ctx)
	if !ok {
		return
	}

	err := h.sm.roleService.RemoveRolePermission(ctx, domain.RolePermissionRequest{
		RoleId:       chi.URLParam(r, "id"),
		CompanyId:    u.CompanyId,
		PermissionId: chi.URLParam(r, "permissionId"),
		GrantorId:    u.Id,
	})
	if err != nil {
		h.writeRoleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Assign a role to a user
// @Description Takes the old role package from the user and gives the new one; personal permissions outside the old package stay.
// @Tags Role
// @Accept json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body dto.AssignRoleRequest true "Role"
// @Success 204
// @Failure 400 {object} dto.Error "Invalid JSON, user id or role"
// @Failure 401 {object} dto.Error "Unauthorized"
// @Failure 403 {object} dto.Error "Owner or role wider than the assigner"
// @Failure 404 {object} dto.Error "User not found"
// @Failure 500 {object} dto.Error "Internal error"
// @Failure 503 {object} dto.Error "Service unavailable"
// @Router /user/{id}/role [put]
func (h *Handler) AssignRole(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var body dto.AssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		dto.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON")
		return
	}
	u, ok := currentUser(w, ctx)
	if !ok {
		return
	}

	err := h.sm.roleService.AssignRole(ctx, domain.AssignRoleRequest{
		UserId:     chi.URLParam(r, "id"),
		CompanyId:  u.CompanyId,
		RoleId:     body.RoleId,
		AssignerId: u.Id,
	})
	if err != nil {
		h.writeRoleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// currentUser writes 401 and returns false when the request has no user.
func currentUser(w http.ResponseWriter, ctx context.Context) (*domain.User, bool) {
	u, ok := middleware.CurrentUser(ctx)
	if !ok {
		dto.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
	}
	return u, ok
}

func (h *Handler) writeRoleError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	code := "internal_error"
	message := "internal error"

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		status = http.StatusServiceUnavailable
		code = "service_unavailable"
		message = "service unavailable"
	case errors.Is(err, domain.ErrInvalidCred):
		status = http.StatusBadRequest
		code = "invalid_credential"
		message = "invalid credential"
	case errors.Is(err, domain.ErrInvalidRoleName):
		status = http.StatusBadRequest
		code = "validation_error"
		message = "role name must be 1 to 100 characters"
	case errors.Is(err, domain.ErrInvalidUserID):
		status = http.StatusBadRequest
		code = "invalid_user_id"
		message = "user id must be a valid UUID"
	case errors.Is(err, domain.ErrRoleNotAssignable):
		status = http.StatusBadRequest
		code = "invalid_role"
		message = "invalid role for company"
	case errors.Is(err, domain.ErrSystemRole):
		status = http.StatusForbidden
		code = "system_role"
		message = "system role cannot be changed"
	case errors.Is(err, domain.ErrPermissionNotOwned):
		status = http.StatusForbidden
		code = "permission_not_owned"
		message = "you do not have this permission"
	case errors.Is(err, domain.ErrRoleExceedsPermissions):
		status = http.StatusForbidden
		code = "role_exceeds_permissions"
		message = "role has permissions you do not have"
	case errors.Is(err, domain.ErrRoleNotFound):
		status = http.StatusNotFound
		code = "role_not_found"
		message = "role not found"
	case errors.Is(err, domain.ErrPermissionNotFound):
		status = http.StatusNotFound
		code = "permission_not_found"
		message = "role does not have this permission"
	case errors.Is(err, domain.ErrUserNotFound):
		status = http.StatusNotFound
		code = "user_not_found"
		message = "user not found"
	case errors.Is(err, domain.ErrRoleAlreadyExists):
		status = http.StatusConflict
		code = "role_name_conflict"
		message = "active role name already exists"
	case errors.Is(err, domain.ErrRoleInUse):
		status = http.StatusConflict
		code = "role_in_use"
		message = "role is assigned to active users"
	}

	dto.WriteError(w, status, code, message)
	if status >= http.StatusInternalServerError {
		h.logger.Error("role request failed", slog.Any("err", err))
	}
}
