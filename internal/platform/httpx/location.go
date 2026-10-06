package httpx

import (
	"bazaar/internal/modules/location"
	"bazaar/internal/platform/httpx/middleware"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var rq location.CreateLocationRequest

	err := json.NewDecoder(r.Body).Decode(&rq)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte("invalid input"))
		return
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
			status = http.StatusBadRequest
			msg = "location already exists"
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

	u, ok := middleware.CurrentUser(r.Context())

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

func (h *Handler) GetAllLocations(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	u, ok := middleware.CurrentUser(r.Context())

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

func (h *Handler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	var rq location.UpdateLocationRequest

	var locationId = chi.URLParam(r, "id")

	if locationId == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("location id is required"))
		h.logger.Error("location id is required")
		return
	}

	err := json.NewDecoder(r.Body).Decode(&rq)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte("invalid input"))
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

}
