package httpx

import (
	"bazaar/internal/modules/location"
	"bazaar/internal/platform/httpx/dto"
	"bazaar/internal/platform/httpx/middleware"
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
// @Param request body dto.CreateLocationRequest true "Location creation data"
// @Success 201 {object} location.CreateLocationResponse
// @Failure 400 {string} string "Invalid credential data"
// @Failure 401 {string} string "Unauthorized"
// @Failure 409 {string} string "Location already exists"
// @Failure 422 {string} string "Invalid input"
// @Failure 500 {string} string "Internal error"
// @Failure 503 {string} string "Service unavailable"
// @Router /location [post]
func (h *Handler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var body dto.CreateLocationRequest

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte("invalid input"))
		return
	}

	u, ok := middleware.CurrentUser(ctx)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
		h.logger.Error("unauthorized")
		return
	}

	rq := location.CreateLocationRequest{
		CompanyId: u.CompanyId,
		Name:      body.Name,
		Address:   body.Address,
		Type:      body.Type,
	}

	res, err := h.sm.locationService.CreateLocation(ctx, &rq)
	if err != nil {
		status := http.StatusInternalServerError
		msg := "DB closed"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			msg = "deadline exceeded"
		case errors.Is(err, location.ErrInvalidCred):
			status = http.StatusBadRequest
			msg = "invalid credential"
		case errors.Is(err, location.ErrLocationAlreadyExists):
			status = http.StatusConflict
			msg = "location already exists"
		case errors.Is(err, location.ErrInvalidType):
			status = http.StatusBadRequest
			msg = "invalid location type"
		}

		w.WriteHeader(status)
		_, _ = w.Write([]byte(msg))
		h.logger.Error("service unavailable", slog.Any("err", err))
		return
	}

	re, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
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
// @Success 200 {object} location.Location
// @Failure 400 {string} string "Invalid credential data"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Location not found"
// @Failure 500 {string} string "Internal error"
// @Failure 503 {string} string "Service unavailable"
// @Router /location/{id} [get]
func (h *Handler) GetLocationById(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	locationId := chi.URLParam(r, "id")
	if locationId == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("location id is required"))
		h.logger.Error("location id is required")
		return
	}

	u, ok := middleware.CurrentUser(ctx)

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
		h.logger.Error("unauthorized")
		return
	}

	res, err := h.sm.locationService.GetLocationById(ctx, locationId, u.CompanyId)

	if err != nil {
		status := http.StatusInternalServerError
		msg := "DB closed"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			msg = "deadline exceeded"
		case errors.Is(err, location.ErrInvalidCred):
			status = http.StatusBadRequest
			msg = "invalid credential"
		case errors.Is(err, location.ErrNotFound):
			status = http.StatusNotFound
			msg = "location not found"
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

// @Summary Get all locations for the current user's company
// @Tags Location
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} location.Location
// @Failure 400 {string} string "Invalid credential data"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Location not found"
// @Failure 500 {string} string "Internal error"
// @Failure 503 {string} string "Service unavailable"
// @Router /location [get]
func (h *Handler) GetAllLocations(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	u, ok := middleware.CurrentUser(ctx)

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
		h.logger.Error("unauthorized")
		return
	}

	res, err := h.sm.locationService.GetCompanyLocations(ctx, u.CompanyId)

	if err != nil {
		status := http.StatusInternalServerError
		msg := "DB closed"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			msg = "deadline exceeded"
		case errors.Is(err, location.ErrInvalidCred):
			status = http.StatusBadRequest
			msg = "invalid credential"
		case errors.Is(err, location.ErrNotFound):
			status = http.StatusNotFound
			msg = "location not found"
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
	_, _ = w.Write(re)
}

// @Summary Update location by ID
// @Tags Location
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Location ID"
// @Param request body dto.UpdateLocationRequest true "Location update data"
// @Success 200 {object} location.UpdateLocationResponse
// @Failure 400 {string} string "Invalid credential data"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Location not found"
// @Failure 409 {string} string "Location already exists"
// @Failure 422 {string} string "Invalid input"
// @Failure 500 {string} string "Internal error"
// @Failure 503 {string} string "Service unavailable"
// @Router /location/{id} [put]
func (h *Handler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var body dto.UpdateLocationRequest

	var locationId = chi.URLParam(r, "id")

	if locationId == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("location id is required"))
		h.logger.Error("location id is required")
		return
	}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte("invalid input"))
		return
	}

	u, ok := middleware.CurrentUser(ctx)

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
		h.logger.Error("unauthorized")
		return
	}

	rq := location.UpdateLocationRequest{
		LocationId: locationId,
		CompanyId:  u.CompanyId,
		Name:       body.Name,
		Address:    body.Address,
		Type:       body.Type,
	}

	res, err := h.sm.locationService.UpdateLocation(ctx, &rq)
	if err != nil {
		status := http.StatusInternalServerError
		msg := "DB closed"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			msg = "deadline exceeded"
		case errors.Is(err, location.ErrInvalidCred):
			status = http.StatusBadRequest
			msg = "invalid credential"
		case errors.Is(err, location.ErrNotFound):
			status = http.StatusNotFound
			msg = "location not found"
		case errors.Is(err, location.ErrLocationAlreadyExists):
			status = http.StatusConflict
			msg = "location already exists"
		case errors.Is(err, location.ErrInvalidType):
			status = http.StatusBadRequest
			msg = "invalid location type"
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

// @Summary Archive location by ID
// @Tags Location
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Location ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {string} string "Invalid credential data"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Location not found"
// @Failure 500 {string} string "Internal error"
// @Failure 503 {string} string "Service unavailable"
// @Router /location/{id}/archive [patch]
func (h *Handler) ArchiveLocation(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var locationId = chi.URLParam(r, "id")

	if locationId == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("location id is required"))
		h.logger.Error("location id is required")
		return
	}

	u, ok := middleware.CurrentUser(ctx)

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
		h.logger.Error("unauthorized")
		return
	}

	err := h.sm.locationService.ArchiveLocation(ctx, locationId, u.CompanyId)
	if err != nil {
		status := http.StatusInternalServerError
		msg := "DB closed"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			msg = "deadline exceeded"
		case errors.Is(err, location.ErrInvalidCred):
			status = http.StatusBadRequest
			msg = "invalid credential"
		case errors.Is(err, location.ErrNotFound):
			status = http.StatusNotFound
			msg = "location not found"
		}

		w.WriteHeader(status)
		_, _ = w.Write([]byte(msg))
		h.logger.Error("service unavailable", slog.Any("err", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
