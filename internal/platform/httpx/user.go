package httpx

import (
	"bazaar/internal/modules/user"
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
// @Param request body user.CreateUserRequest true "User creation data"
// @Success 201 {object} user.CreateUserResponse
// @Failure 400 {string} string "Invalid credential data"
// @Failure 409 {string} string "Phone already registered"
// @Failure 422 {string} string "Invalid input"
// @Failure 500 {string} string "Internal error"
// @Failure 503 {string} string "Service unavailable"
// @Router /user [post]
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var rq user.CreateUserRequest

	err := json.NewDecoder(r.Body).Decode(&rq)

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte("invalid input"))
		h.logger.Error("invalid input", slog.Any("err", err))
		return
	}

	u, ok := middleware.CurrentUser(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
		h.logger.Error("unauthorized")
		return
	}

	rq.CompanyId = u.CompanyId

	res, err := h.sm.userService.CreateUser(ctx, &rq)

	if err != nil {
		status := http.StatusInternalServerError
		msg := "DB closed"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			msg = "deadline exceeded"
		case errors.Is(err, user.ErrInvalidPhone):
			status = http.StatusBadRequest
			msg = "not valid phone number"
		case errors.Is(err, user.ErrPhoneTaken):
			status = http.StatusConflict
			msg = "phone is taken"
		case errors.Is(err, user.ErrInvalidCred):
			status = http.StatusBadRequest
			msg = "invalid credential"
		case errors.Is(err, validator.ErrPasswordInvalid):
			status = http.StatusBadRequest
			msg = "password is invalid"
		case errors.Is(err, user.ErrInvalidRole):
			status = http.StatusBadRequest
			msg = "invalid role for company"
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

// @Summary Get user info
// @Tags User
// @Produce json
// @Param id path string true "User ID"
// @Security BearerAuth
// @Success 200 {object} user.User
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal error"
// @Failure 503 {string} string "Service unavailable"
// @Router /user/{id} [get]
func (h *Handler) GetUserById(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	userId := chi.URLParam(r, "id")
	if userId == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("user id is required"))
		h.logger.Error("user id is required")
		return
	}

	u, ok := middleware.CurrentUser(r.Context())

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
		h.logger.Error("unauthorized")
		return
	}

	res, err := h.sm.userService.GetUserById(ctx, userId, u.CompanyId)

	if err != nil {
		status := http.StatusInternalServerError
		msg := "DB closed"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			msg = "deadline exceeded"
		case errors.Is(err, user.ErrInvalidCred):
			status = http.StatusBadRequest
			msg = "invalid credential"
		case errors.Is(err, user.ErrNotFound):
			status = http.StatusNotFound
			msg = "user not found"
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
