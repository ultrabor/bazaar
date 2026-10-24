package httpx

import (
	"bazaar/internal/modules/permission"
	"bazaar/internal/platform/httpx/dto"
	"bazaar/internal/platform/httpx/middleware"
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

	jwt, ok := middleware.CurrentUser(ctx)

	if !ok {
		dto.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	res, err := h.sm.permissionService.GetListByUserId(ctx, jwt.Id)

	if err != nil {
		status := http.StatusInternalServerError
		code := "internal_error"
		msg := "internal error"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			code = "service_unavailable"
			msg = "service unavailable"
		case errors.Is(err, permission.ErrNotFound):
			status = http.StatusNotFound
			code = "permission_not_found"
			msg = "permission_not_found"
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
	_, _ = w.Write(re)
}

func (h *Handler) SetPermission(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	jwt, ok := middleware.CurrentUser(ctx)

	if !ok {
		dto.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	var body dto.ChangePermissionRequest

	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		dto.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON")
		return
	}

	rq := permission.ChangePermissionRequest{
		GuarantorUserId: jwt.Id,
		ApplyUserId:     body.UserId,
		PermissionId:    body.PermissionId,
	}

	err = h.sm.permissionService.SetPermission(ctx, rq)

	if err != nil {
		status := http.StatusInternalServerError
		code := "internal_error"
		msg := "internal error"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			code = "service_unavailable"
			msg = "service unavailable"
		case errors.Is(err, permission.ErrNotFound):
			status = http.StatusNotFound
			code = "permission_not_found"
			msg = "permission_not_found"
		case errors.Is(err, permission.ErrInvalidCred):
			status = http.StatusNotFound
			code = "invalid_cred"
			msg = "invalid credential"
		}

		dto.WriteError(w, status, code, msg)
		h.logger.Error("service unavailable", slog.Any("err", err))
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) RemovePermission(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	jwt, ok := middleware.CurrentUser(ctx)

	if !ok {
		dto.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	var body dto.ChangePermissionRequest

	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		dto.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON")
		return
	}

	rq := permission.ChangePermissionRequest{
		GuarantorUserId: jwt.Id,
		ApplyUserId:     body.UserId,
		PermissionId:    body.PermissionId,
	}

	err = h.sm.permissionService.RemovePermission(ctx, rq)

	if err != nil {
		status := http.StatusInternalServerError
		code := "internal_error"
		msg := "internal error"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			code = "service_unavailable"
			msg = "service unavailable"
		case errors.Is(err, permission.ErrNotFound):
			status = http.StatusNotFound
			code = "permission_not_found"
			msg = "permission_not_found"
		case errors.Is(err, permission.ErrInvalidCred):
			status = http.StatusNotFound
			code = "invalid_cred"
			msg = "invalid credential"
		}

		dto.WriteError(w, status, code, msg)
		h.logger.Error("service unavailable", slog.Any("err", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
