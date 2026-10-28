package location

import (
	"bazaar/internal/platform/httpx"
	"bazaar/internal/platform/identity"
	"bazaar/internal/platform/support/apperror"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// @Summary Create a new location
// @Tags Location
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateLocationBody true "Location creation data"
// @Success 201 {object} CreateLocationResponse
// @Failure 400 {object} httpx.Error "Invalid credential data"
// @Failure 401 {object} httpx.Error "Unauthorized"
// @Failure 409 {object} httpx.Error "Location already exists"
// @Failure 500 {object} httpx.Error "Internal error"
// @Failure 503 {object} httpx.Error "Service unavailable"
// @Router /location [post]
func (h *Handler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var body CreateLocationBody

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON")
		return
	}

	u, ok := identity.FromContext(ctx)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		h.logger.Error("unauthorized")
		return
	}

	rq := CreateLocationRequest{
		CompanyId: u.CompanyId,
		Name:      body.Name,
		Address:   body.Address,
		Type:      body.Type,
	}

	res, err := h.svc.CreateLocation(ctx, &rq)
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
		case errors.Is(err, ErrLocationAlreadyExists):
			status = http.StatusConflict
			code = "location_already_exists"
			msg = "location already exists"
		case errors.Is(err, ErrInvalidLocationType):
			status = http.StatusBadRequest
			code = "invalid_location_type"
			msg = "invalid location type"
		}

		httpx.WriteError(w, status, code, msg)
		h.logger.Error("service unavailable", slog.Any("err", err))
		return
	}

	re, err := json.Marshal(res)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_, _ = w.Write(re)

}

// @Summary Get location by ID
// @Tags Location
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Location ID"
// @Success 200 {object} Location
// @Failure 400 {object} httpx.Error "Invalid credential data"
// @Failure 401 {object} httpx.Error "Unauthorized"
// @Failure 404 {object} httpx.Error "Location not found"
// @Failure 500 {object} httpx.Error "Internal error"
// @Failure 503 {object} httpx.Error "Service unavailable"
// @Router /location/{id} [get]
func (h *Handler) GetLocationById(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	locationId := chi.URLParam(r, "id")
	if locationId == "" {
		httpx.WriteError(w, http.StatusBadRequest, "location_id_required", "location id is required")
		h.logger.Error("location id is required")
		return
	}

	u, ok := identity.FromContext(ctx)

	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		h.logger.Error("unauthorized")
		return
	}

	res, err := h.svc.GetLocationById(ctx, locationId, u.CompanyId)

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
		case errors.Is(err, ErrLocationNotFound):
			status = http.StatusNotFound
			code = "location_not_found"
			msg = "location not found"
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

// @Summary Get all locations for the current user's company
// @Tags Location
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Location
// @Failure 400 {object} httpx.Error "Invalid credential data"
// @Failure 401 {object} httpx.Error "Unauthorized"
// @Failure 404 {object} httpx.Error "Location not found"
// @Failure 500 {object} httpx.Error "Internal error"
// @Failure 503 {object} httpx.Error "Service unavailable"
// @Router /location [get]
func (h *Handler) GetAllLocations(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	u, ok := identity.FromContext(ctx)

	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		h.logger.Error("unauthorized")
		return
	}

	res, err := h.svc.GetCompanyLocations(ctx, u.CompanyId)

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
		case errors.Is(err, ErrLocationNotFound):
			status = http.StatusNotFound
			code = "location_not_found"
			msg = "location not found"
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

// @Summary Update location by ID
// @Tags Location
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Location ID"
// @Param request body UpdateLocationBody true "Location update data"
// @Success 200 {object} UpdateLocationResponse
// @Failure 400 {object} httpx.Error "Invalid credential data"
// @Failure 401 {object} httpx.Error "Unauthorized"
// @Failure 404 {object} httpx.Error "Location not found"
// @Failure 409 {object} httpx.Error "Location already exists"
// @Failure 500 {object} httpx.Error "Internal error"
// @Failure 503 {object} httpx.Error "Service unavailable"
// @Router /location/{id} [put]
func (h *Handler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var body UpdateLocationBody

	var locationId = chi.URLParam(r, "id")

	if locationId == "" {
		httpx.WriteError(w, http.StatusBadRequest, "location_id_required", "location id is required")
		h.logger.Error("location id is required")
		return
	}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON")
		return
	}

	u, ok := identity.FromContext(ctx)

	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		h.logger.Error("unauthorized")
		return
	}

	rq := UpdateLocationRequest{
		LocationId: locationId,
		CompanyId:  u.CompanyId,
		Name:       body.Name,
		Address:    body.Address,
		Type:       body.Type,
	}

	res, err := h.svc.UpdateLocation(ctx, &rq)
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
		case errors.Is(err, ErrLocationNotFound):
			status = http.StatusNotFound
			code = "location_not_found"
			msg = "location not found"
		case errors.Is(err, ErrLocationAlreadyExists):
			status = http.StatusConflict
			code = "location_already_exists"
			msg = "location already exists"
		case errors.Is(err, ErrInvalidLocationType):
			status = http.StatusBadRequest
			code = "invalid_location_type"
			msg = "invalid location type"
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

// @Summary Archive location by ID
// @Tags Location
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Location ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} httpx.Error "Invalid credential data"
// @Failure 401 {object} httpx.Error "Unauthorized"
// @Failure 404 {object} httpx.Error "Location not found"
// @Failure 500 {object} httpx.Error "Internal error"
// @Failure 503 {object} httpx.Error "Service unavailable"
// @Router /location/{id}/archive [patch]
func (h *Handler) ArchiveLocation(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var locationId = chi.URLParam(r, "id")

	if locationId == "" {
		httpx.WriteError(w, http.StatusBadRequest, "location_id_required", "location id is required")
		h.logger.Error("location id is required")
		return
	}

	u, ok := identity.FromContext(ctx)

	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		h.logger.Error("unauthorized")
		return
	}

	err := h.svc.ArchiveLocation(ctx, locationId, u.CompanyId)
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
		case errors.Is(err, ErrLocationNotFound):
			status = http.StatusNotFound
			code = "location_not_found"
			msg = "location not found"
		}

		httpx.WriteError(w, status, code, msg)
		h.logger.Error("service unavailable", slog.Any("err", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Request bodies.

type CreateLocationBody struct {
	Name    string       `json:"name"`
	Address string       `json:"address"`
	Type    LocationType `json:"type"`
}

type UpdateLocationBody struct {
	Name    string       `json:"name"`
	Address string       `json:"address"`
	Type    LocationType `json:"type"`
}
