package httpx

import (
	"bazaar/internal/modules/auth"
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
// @Param request body auth.RegisterRequest true "Registration data"
// @Success 201 {object} auth.RegisterResponse
// @Failure 400 {string} string "Invalid registration data"
// @Failure 409 {string} string "Phone already registered"
// @Failure 422 {string} string "Invalid input"
// @Failure 500 {string} string "Internal error"
// @Failure 503 {string} string "Service unavailable"
// @Router /auth/register-owner [post]
func (h *Handler) RegisterOwner(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var rq auth.RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&rq)

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte("invalid input"))
		h.logger.Error("invalid input", slog.Any("err", err))
		return
	}

	res, err := h.sm.authService.RegisterOwner(ctx, rq)

	if err != nil {
		status := http.StatusInternalServerError
		msg := "DB closed"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			msg = "deadline exceeded"
		case errors.Is(err, auth.ErrInvalidCred):
			status = http.StatusBadRequest
			msg = "invalid credential"
		case errors.Is(err, auth.ErrInvalidPhone):
			status = http.StatusBadRequest
			msg = "not valid phone number"
		case errors.Is(err, auth.ErrPhoneTaken):
			status = http.StatusConflict
			msg = "phone is taken"
		case errors.Is(err, validator.ErrPasswordInvalid):
			status = http.StatusBadRequest
			msg = "password is invalid"
		}

		w.WriteHeader(status)
		_, _ = w.Write([]byte(msg))
		h.logger.Error("service unavailable", slog.Any("err", err))
		return
	}

	re, err := json.Marshal(res)

	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("service unavailable"))
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
// @Param request body auth.LoginRequest true "Login data"
// @Success 200 {object} auth.LoginResponse
// @Failure 400 {string} string "Invalid login data"
// @Failure 401 {string} string "Unauthorized"
// @Failure 422 {string} string "Invalid input"
// @Failure 500 {string} string "Internal error"
// @Failure 503 {string} string "Service unavailable"
// @Router /auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var rq auth.LoginRequest

	err := json.NewDecoder(r.Body).Decode(&rq)

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte("invalid input"))
		h.logger.Error("invalid input", slog.Any("err", err))
		return
	}

	res, err := h.sm.authService.Login(ctx, rq)

	if err != nil {
		status := http.StatusInternalServerError
		msg := "DB closed"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			msg = "deadline exceeded"
		case errors.Is(err, auth.ErrInvalidCred):
			status = http.StatusUnauthorized
			msg = "invalid credential"
		case errors.Is(err, auth.ErrInvalidPhone):
			status = http.StatusBadRequest
			msg = "not valid phone number"
		case errors.Is(err, auth.ErrUnauthorized):
			status = http.StatusUnauthorized
			msg = "unauthorized"
		case errors.Is(err, validator.ErrPasswordInvalid):
			status = http.StatusBadRequest
			msg = "invalid password"
		}

		w.WriteHeader(status)
		_, _ = w.Write([]byte(msg))
		h.logger.Error("service unavailable", slog.Any("err", err))
		return
	}

	re, err := json.Marshal(res)

	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("service unavailable"))
		h.logger.Error("service unavailable", slog.Any("err", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(re)
}
