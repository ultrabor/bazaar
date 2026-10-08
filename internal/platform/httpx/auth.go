package httpx

import (
	"bazaar/internal/modules/auth"
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
)

// @Summary Register a company owner
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterOwnerRequest true "Registration data"
// @Success 201 {object} auth.RegisterResponse
// @Failure 400 {object} dto.Error "Invalid registration data"
// @Failure 409 {object} dto.Error "Phone already registered"
// @Failure 422 {object} dto.Error "Invalid input"
// @Failure 500 {object} dto.Error "Internal error"
// @Failure 503 {object} dto.Error "Service unavailable"
// @Router /auth/register-owner [post]
func (h *Handler) RegisterOwner(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var body dto.RegisterOwnerRequest

	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		dto.WriteError(w, http.StatusUnprocessableEntity, "invalid_input", "invalid input")
		h.logger.Error("invalid input", slog.Any("err", err))
		return
	}

	rq := auth.RegisterRequest{
		FirstName:   body.FirstName,
		LastName:    body.LastName,
		CompanyName: body.CompanyName,
		Phone:       body.Phone,
		Password:    body.Password,
	}

	res, err := h.sm.authService.RegisterOwner(ctx, rq)

	if err != nil {
		status := http.StatusInternalServerError
		code := "internal_error"
		msg := "internal error"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			code = "deadline_exceeded"
			msg = "deadline exceeded"
		case errors.Is(err, auth.ErrInvalidCred):
			status = http.StatusBadRequest
			code = "invalid_credential"
			msg = "invalid credential"
		case errors.Is(err, auth.ErrInvalidPhone):
			status = http.StatusBadRequest
			code = "invalid_phone"
			msg = "not valid phone number"
		case errors.Is(err, auth.ErrPhoneTaken):
			status = http.StatusConflict
			code = "phone_taken"
			msg = "phone is taken"
		case errors.Is(err, validator.ErrPasswordInvalid):
			status = http.StatusBadRequest
			code = "invalid_password"
			msg = "password is invalid"
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

// @Summary Login a user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login data"
// @Success 200 {object} auth.LoginResponse
// @Failure 400 {object} dto.Error "Invalid login data"
// @Failure 401 {object} dto.Error "Unauthorized"
// @Failure 422 {object} dto.Error "Invalid input"
// @Failure 500 {object} dto.Error "Internal error"
// @Failure 503 {object} dto.Error "Service unavailable"
// @Router /auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var body dto.LoginRequest

	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		dto.WriteError(w, http.StatusUnprocessableEntity, "invalid_input", "invalid input")
		h.logger.Error("invalid input", slog.Any("err", err))
		return
	}

	rq := auth.LoginRequest{
		Phone:    body.Phone,
		Password: body.Password,
	}

	res, err := h.sm.authService.Login(ctx, rq)

	if err != nil {
		status := http.StatusInternalServerError
		code := "internal_error"
		msg := "internal error"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			code = "deadline_exceeded"
			msg = "deadline exceeded"
		case errors.Is(err, auth.ErrInvalidCred):
			status = http.StatusUnauthorized
			code = "invalid_credential"
			msg = "invalid credential"
		case errors.Is(err, auth.ErrInvalidPhone):
			status = http.StatusBadRequest
			code = "invalid_phone"
			msg = "not valid phone number"
		case errors.Is(err, auth.ErrUnauthorized):
			status = http.StatusUnauthorized
			code = "unauthorized"
			msg = "unauthorized"
		case errors.Is(err, validator.ErrPasswordInvalid):
			status = http.StatusBadRequest
			code = "invalid_password"
			msg = "invalid password"
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

// @Summary Get current user info
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} user.User
// @Failure 401 {object} dto.Error "Unauthorized"
// @Failure 500 {object} dto.Error "Internal error"
// @Failure 503 {object} dto.Error "Service unavailable"
// @Router /auth/me [get]
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	var u *user.User
	u, ok := middleware.CurrentUser(r.Context())
	if !ok {
		dto.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(u)
}
