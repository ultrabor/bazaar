package permission

import (
	"bazaar/internal/platform/httpx"
	"bazaar/internal/platform/identity"
	"bazaar/internal/platform/support/apperror"
	"bazaar/internal/user"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

func (h *Handler) GetListPermissions(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	jwt, ok := identity.FromContext(ctx)

	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	res, err := h.svc.GetUserPermissions(ctx, jwt.Id, jwt.CompanyId, r.URL.Query().Get("user_id"))

	if err != nil {
		status := http.StatusInternalServerError
		code := "internal_error"
		msg := "internal error"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			code = "service_unavailable"
			msg = "service unavailable"
		case errors.Is(err, ErrInvalidUserID):
			status = http.StatusBadRequest
			code = "invalid_user_id"
			msg = "user id must be a valid UUID"
		case errors.Is(err, apperror.ErrForbidden):
			status = http.StatusForbidden
			code = "do_not_have_enough_permission"
			msg = "dont have enough permission"
		case errors.Is(err, user.ErrUserNotFound):
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
	_, _ = w.Write(re)
}

func (h *Handler) SetPermission(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	jwt, ok := identity.FromContext(ctx)

	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	var body ChangePermissionBody

	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON")
		return
	}

	rq := ChangePermissionRequest{
		GuarantorUserId: jwt.Id,
		CompanyId:       jwt.CompanyId,
		ApplyUserId:     body.UserId,
		PermissionId:    body.PermissionId,
	}

	err = h.svc.SetPermission(ctx, rq)

	if err != nil {
		status := http.StatusInternalServerError
		code := "internal_error"
		msg := "internal error"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			code = "service_unavailable"
			msg = "service unavailable"
		case errors.Is(err, ErrPermissionNotFound):
			status = http.StatusNotFound
			code = "permission_not_found"
			msg = "permission_not_found"
		case errors.Is(err, ErrPermissionNotOwned):
			status = http.StatusForbidden
			code = "invalid_cred"
			msg = "invalid credential"
		case errors.Is(err, ErrInvalidUserID):
			status = http.StatusBadRequest
			code = "invalid_user_id"
			msg = "user id must be a valid UUID"
		case errors.Is(err, ErrPermissionAlreadyExists):
			status = http.StatusConflict
			code = "permission_already_exists"
			msg = "permission already exists"
		case errors.Is(err, user.ErrUserNotFound):
			status = http.StatusNotFound
			code = "user_not_found"
			msg = "user not found"
		}

		httpx.WriteError(w, status, code, msg)
		h.logger.Error("service unavailable", slog.Any("err", err))
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) RemovePermission(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	jwt, ok := identity.FromContext(ctx)

	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	var body ChangePermissionBody

	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON")
		return
	}

	rq := ChangePermissionRequest{
		GuarantorUserId: jwt.Id,
		CompanyId:       jwt.CompanyId,
		ApplyUserId:     body.UserId,
		PermissionId:    body.PermissionId,
	}

	err = h.svc.RemovePermission(ctx, rq)

	if err != nil {
		status := http.StatusInternalServerError
		code := "internal_error"
		msg := "internal error"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			code = "service_unavailable"
			msg = "service unavailable"
		case errors.Is(err, ErrPermissionNotFound):
			status = http.StatusNotFound
			code = "permission_not_found"
			msg = "permission_not_found"
		case errors.Is(err, ErrPermissionNotOwned):
			status = http.StatusForbidden
			code = "invalid_cred"
			msg = "invalid credential"
		case errors.Is(err, ErrInvalidUserID):
			status = http.StatusBadRequest
			code = "invalid_user_id"
			msg = "user id must be a valid UUID"
		case errors.Is(err, user.ErrUserNotFound):
			status = http.StatusNotFound
			code = "user_not_found"
			msg = "user not found"
		}

		httpx.WriteError(w, status, code, msg)
		h.logger.Error("service unavailable", slog.Any("err", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Request bodies.

type ChangePermissionBody struct {
	UserId       string `json:"user_id"`
	PermissionId string `json:"permission_id"`
}
