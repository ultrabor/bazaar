package user

import (
	"bazaar/internal/modules/role"
	"bazaar/internal/platform/httpx"
	"bazaar/internal/platform/identity"
	"bazaar/internal/platform/support/apperror"
	"bazaar/internal/platform/support/validator"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// @Summary Create a new user
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateUserBody true "User creation data"
// @Success 201 {object} CreateUserResponse
// @Failure 400 {object} httpx.Error "Invalid credential data"
// @Failure 403 {object} httpx.Error "Role has permissions the creator lacks"
// @Failure 409 {object} httpx.Error "Phone already registered"
// @Failure 500 {object} httpx.Error "Internal error"
// @Failure 503 {object} httpx.Error "Service unavailable"
// @Router /user [post]
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var body CreateUserBody

	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON")
		h.logger.Error("invalid input", slog.Any("err", err))
		return
	}

	u, ok := identity.FromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		h.logger.Error("unauthorized")
		return
	}

	rq := CreateUserRequest{
		FirstName:  body.FirstName,
		LastName:   body.LastName,
		Phone:      body.Phone,
		Password:   body.Password,
		CompanyId:  u.CompanyId,
		UserRoleId: body.UserRoleId,
		CreatorId:  u.Id,
	}

	res, err := h.svc.CreateUser(ctx, &rq)

	if err != nil {
		status := http.StatusInternalServerError
		code := "internal_error"
		msg := "internal error"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			code = "service_unavailable"
			msg = "service unavailable"
		case errors.Is(err, ErrInvalidPhone):
			status = http.StatusBadRequest
			code = "invalid_phone"
			msg = "not valid phone number"
		case errors.Is(err, ErrPhoneTaken):
			status = http.StatusConflict
			code = "phone_taken"
			msg = "phone is taken"
		case errors.Is(err, apperror.ErrInvalidCred):
			status = http.StatusBadRequest
			code = "invalid_credential"
			msg = "invalid credential"
		case errors.Is(err, validator.ErrPasswordInvalid):
			status = http.StatusBadRequest
			code = "invalid_password"
			msg = "password is invalid"
		case errors.Is(err, ErrInvalidRole):
			status = http.StatusBadRequest
			code = "invalid_role"
			msg = "invalid role for company"
		case errors.Is(err, role.ErrRoleExceedsPermissions):
			status = http.StatusForbidden
			code = "role_exceeds_permissions"
			msg = "role has permissions you do not have"
		}

		httpx.WriteError(w, status, code, msg)
		h.logger.Error("service unavailable", slog.Any("err", err))
		return
	}

	re, err := json.Marshal(res)

	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "internal error")
		h.logger.Error("service unavailable", slog.Any("err", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_, _ = w.Write(re)
}

// @Summary Get user info
// @Tags User
// @Produce json
// @Param id path string true "User ID"
// @Security BearerAuth
// @Success 200 {object} User
// @Failure 400 {object} httpx.Error "Invalid user ID"
// @Failure 401 {object} httpx.Error "Unauthorized"
// @Failure 404 {object} httpx.Error "User not found"
// @Failure 500 {object} httpx.Error "Internal error"
// @Failure 503 {object} httpx.Error "Service unavailable"
// @Router /user/{id} [get]
func (h *Handler) GetUserById(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	userId := chi.URLParam(r, "id")
	if userId == "" {
		httpx.WriteError(w, http.StatusBadRequest, "user_id_required", "user id is required")
		h.logger.Error("user id is required")
		return
	}

	u, ok := identity.FromContext(r.Context())

	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		h.logger.Error("unauthorized")
		return
	}

	res, err := h.svc.GetUserById(ctx, userId, u.CompanyId)

	if err != nil {
		status := http.StatusInternalServerError
		code := "internal_error"
		msg := "internal error"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			code = "service_unavailable"
			msg = "service unavailable"
		case errors.Is(err, apperror.ErrInvalidCred):
			status = http.StatusBadRequest
			code = "invalid_credential"
			msg = "invalid credential"
		case errors.Is(err, ErrUserNotFound):
			status = http.StatusNotFound
			code = "user_not_found"
			msg = "user not found"
		}

		httpx.WriteError(w, status, code, msg)
		h.logger.Error("service unavailable", slog.Any("err", err))
		return
	}

	re, err := json.Marshal(res)

	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "internal error")
		h.logger.Error("service unavailable", slog.Any("err", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(re)
}

// Request bodies.

type CreateUserBody struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Phone      string `json:"phone"`
	Password   string `json:"password"`
	UserRoleId string `json:"user_role_id"`
}
