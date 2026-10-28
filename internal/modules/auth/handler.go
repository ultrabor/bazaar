package auth

import (
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
)

// @Summary Register a company owner
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterOwnerBody true "Registration data"
// @Success 201 {object} RegisterResponse
// @Failure 400 {object} httpx.Error "Invalid registration data"
// @Failure 409 {object} httpx.Error "Phone already registered"
// @Failure 500 {object} httpx.Error "Internal error"
// @Failure 503 {object} httpx.Error "Service unavailable"
// @Router /auth/register-owner [post]
func (h *Handler) RegisterOwner(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var body RegisterOwnerBody

	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON")
		h.logger.Error("invalid input", slog.Any("err", err))
		return
	}

	rq := RegisterRequest(body)

	res, err := h.svc.RegisterOwner(ctx, rq)

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
		case errors.Is(err, ErrInvalidPhone):
			status = http.StatusBadRequest
			code = "invalid_phone"
			msg = "not valid phone number"
		case errors.Is(err, ErrPhoneTaken):
			status = http.StatusConflict
			code = "phone_taken"
			msg = "phone is taken"
		case errors.Is(err, validator.ErrPasswordInvalid):
			status = http.StatusBadRequest
			code = "invalid_password"
			msg = "password is invalid"
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

// @Summary Login a user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginBody true "Login data"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} httpx.Error "Invalid login data"
// @Failure 401 {object} httpx.Error "Unauthorized"
// @Failure 500 {object} httpx.Error "Internal error"
// @Failure 503 {object} httpx.Error "Service unavailable"
// @Router /auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var body LoginBody

	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON")
		h.logger.Error("invalid input", slog.Any("err", err))
		return
	}

	rq := LoginRequest(body)

	res, err := h.svc.Login(ctx, rq)

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
			status = http.StatusUnauthorized
			code = "invalid_credential"
			msg = "invalid credential"
		case errors.Is(err, ErrInvalidPhone):
			status = http.StatusBadRequest
			code = "invalid_phone"
			msg = "not valid phone number"
		case errors.Is(err, ErrUnauthorized):
			status = http.StatusUnauthorized
			code = "unauthorized"
			msg = "unauthorized"
		case errors.Is(err, validator.ErrPasswordInvalid):
			status = http.StatusBadRequest
			code = "invalid_password"
			msg = "invalid password"
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

// @Summary Get current user info
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} identity.User
// @Failure 401 {object} httpx.Error "Unauthorized"
// @Failure 500 {object} httpx.Error "Internal error"
// @Failure 503 {object} httpx.Error "Service unavailable"
// @Router /auth/me [get]
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	u, ok := identity.FromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(u)
}

// Request bodies.

type RegisterOwnerBody struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	CompanyName string `json:"company_name"`
	Phone       string `json:"phone"`
	Password    string `json:"password"`
}

type LoginBody struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}
