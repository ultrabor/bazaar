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
)

func (h *Handler) GetListPermissions(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	jwt, ok := middleware.CurrentUser(ctx)

	if !ok {
		dto.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	res, err := h.sm.permissionService.GetUserPermissions(ctx, jwt.Id, jwt.CompanyId, r.URL.Query().Get("user_id"))

	if err != nil {
		status := http.StatusInternalServerError
		code := "internal_error"
		msg := "internal error"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			code = "service_unavailable"
			msg = "service unavailable"
		case errors.Is(err, domain.ErrInvalidUserID):
			status = http.StatusBadRequest
			code = "invalid_user_id"
			msg = "user id must be a valid UUID"
		case errors.Is(err, domain.ErrForbidden):
			status = http.StatusForbidden
			code = "do_not_have_enough_permission"
			msg = "dont have enough permission"
		case errors.Is(err, domain.ErrUserNotFound):
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

	rq := domain.ChangePermissionRequest{
		GuarantorUserId: jwt.Id,
		CompanyId:       jwt.CompanyId,
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
		case errors.Is(err, domain.ErrPermissionNotFound):
			status = http.StatusNotFound
			code = "permission_not_found"
			msg = "permission_not_found"
		case errors.Is(err, domain.ErrPermissionNotOwned):
			status = http.StatusForbidden
			code = "invalid_cred"
			msg = "invalid credential"
		case errors.Is(err, domain.ErrInvalidUserID):
			status = http.StatusBadRequest
			code = "invalid_user_id"
			msg = "user id must be a valid UUID"
		case errors.Is(err, domain.ErrPermissionAlreadyExists):
			status = http.StatusConflict
			code = "permission_already_exists"
			msg = "permission already exists"
		case errors.Is(err, domain.ErrUserNotFound):
			status = http.StatusNotFound
			code = "user_not_found"
			msg = "user not found"
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

	rq := domain.ChangePermissionRequest{
		GuarantorUserId: jwt.Id,
		CompanyId:       jwt.CompanyId,
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
		case errors.Is(err, domain.ErrPermissionNotFound):
			status = http.StatusNotFound
			code = "permission_not_found"
			msg = "permission_not_found"
		case errors.Is(err, domain.ErrPermissionNotOwned):
			status = http.StatusForbidden
			code = "invalid_cred"
			msg = "invalid credential"
		case errors.Is(err, domain.ErrInvalidUserID):
			status = http.StatusBadRequest
			code = "invalid_user_id"
			msg = "user id must be a valid UUID"
		case errors.Is(err, domain.ErrUserNotFound):
			status = http.StatusNotFound
			code = "user_not_found"
			msg = "user not found"
		}

		dto.WriteError(w, status, code, msg)
		h.logger.Error("service unavailable", slog.Any("err", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
