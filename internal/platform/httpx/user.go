package httpx

import (
	"bazaar/internal/modules/user"
	"bazaar/internal/platform/httpx/dto"
	"bazaar/internal/platform/httpx/middleware"
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
// @Param request body dto.CreateUserRequest true "User creation data"
// @Success 201 {object} user.CreateUserResponse
// @Failure 400 {object} dto.Error "Invalid credential data"
// @Failure 409 {object} dto.Error "Phone already registered"
// @Failure 500 {object} dto.Error "Internal error"
// @Failure 503 {object} dto.Error "Service unavailable"
// @Router /user [post]
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var body dto.CreateUserRequest

	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		dto.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON")
		h.logger.Error("invalid input", slog.Any("err", err))
		return
	}

	u, ok := middleware.CurrentUser(r.Context())
	if !ok {
		dto.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		h.logger.Error("unauthorized")
		return
	}

	rq := user.CreateUserRequest{
		FirstName:  body.FirstName,
		LastName:   body.LastName,
		Phone:      body.Phone,
		Password:   body.Password,
		CompanyId:  u.CompanyId,
		UserRoleId: body.UserRoleId,
	}

	res, err := h.sm.userService.CreateUser(ctx, &rq)

	if err != nil {
		status := http.StatusInternalServerError
		code := "internal_error"
		msg := "internal error"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			code = "service_unavailable"
			msg = "service unavailable"
		case errors.Is(err, user.ErrInvalidPhone):
			status = http.StatusBadRequest
			code = "invalid_phone"
			msg = "not valid phone number"
		case errors.Is(err, user.ErrPhoneTaken):
			status = http.StatusConflict
			code = "phone_taken"
			msg = "phone is taken"
		case errors.Is(err, user.ErrInvalidCred):
			status = http.StatusBadRequest
			code = "invalid_credential"
			msg = "invalid credential"
		case errors.Is(err, validator.ErrPasswordInvalid):
			status = http.StatusBadRequest
			code = "invalid_password"
			msg = "password is invalid"
		case errors.Is(err, user.ErrInvalidRole):
			status = http.StatusBadRequest
			code = "invalid_role"
			msg = "invalid role for company"
		}

		dto.WriteError(w, status, code, msg)
		h.logger.Error("service unavailable", slog.Any("err", err))
		return
	}

	re, err := json.Marshal(res)

	if err != nil {
		dto.WriteError(w, http.StatusInternalServerError, "internal_error", "internal error")
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
// @Success 200 {object} user.User
// @Failure 400 {object} dto.Error "Invalid user ID"
// @Failure 401 {object} dto.Error "Unauthorized"
// @Failure 404 {object} dto.Error "User not found"
// @Failure 500 {object} dto.Error "Internal error"
// @Failure 503 {object} dto.Error "Service unavailable"
// @Router /user/{id} [get]
func (h *Handler) GetUserById(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	userId := chi.URLParam(r, "id")
	if userId == "" {
		dto.WriteError(w, http.StatusBadRequest, "user_id_required", "user id is required")
		h.logger.Error("user id is required")
		return
	}

	u, ok := middleware.CurrentUser(r.Context())

	if !ok {
		dto.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		h.logger.Error("unauthorized")
		return
	}

	res, err := h.sm.userService.GetUserById(ctx, userId, u.CompanyId)

	if err != nil {
		status := http.StatusInternalServerError
		code := "internal_error"
		msg := "internal error"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			code = "service_unavailable"
			msg = "service unavailable"
		case errors.Is(err, user.ErrInvalidCred):
			status = http.StatusBadRequest
			code = "invalid_credential"
			msg = "invalid credential"
		case errors.Is(err, user.ErrNotFound):
			status = http.StatusNotFound
			code = "user_not_found"
			msg = "user not found"
		}

		dto.WriteError(w, status, code, msg)
		h.logger.Error("service unavailable", slog.Any("err", err))
		return
	}

	re, err := json.Marshal(res)

	if err != nil {
		dto.WriteError(w, http.StatusInternalServerError, "internal_error", "internal error")
		h.logger.Error("service unavailable", slog.Any("err", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(re)
}
