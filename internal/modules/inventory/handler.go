package inventory

import (
	"bazaar/internal/platform/httpx"
	"bazaar/internal/platform/support/apperror"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

const requestTimeout = 2 * time.Second

// @Summary List the stock of a location
// @Description Quantities are integers in thousandths of the product unit: 1500 is 1.5 kg. Archived locations stay readable.
// @Tags Inventory
// @Produce json
// @Security BearerAuth
// @Param id path string true "Location ID"
// @Param limit query int false "Page size, default 50, max 200"
// @Param offset query int false "Rows to skip"
// @Success 200 {array} Balance
// @Failure 400 {object} httpx.Error "Invalid pagination"
// @Failure 401 {object} httpx.Error "Unauthorized"
// @Failure 403 {object} httpx.Error "Forbidden"
// @Failure 404 {object} httpx.Error "Location not found"
// @Failure 500 {object} httpx.Error "Internal error"
// @Failure 503 {object} httpx.Error "Service unavailable"
// @Router /location/{id}/stock [get]
func (h *Handler) ListStock(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), requestTimeout)
	defer stop()

	u, ok := httpx.CurrentUser(w, ctx)
	if !ok {
		return
	}
	limit, offset, ok := parsePage(w, r)
	if !ok {
		return
	}

	res, err := h.svc.ListBalances(ctx, BalanceFilter{
		CompanyID:  u.CompanyId,
		LocationID: chi.URLParam(r, "id"),
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, res)
}

// @Summary List stock movements of a product at a location
// @Description Newest first. Qty is signed thousandths: positive for stock in, negative for stock out.
// @Tags Inventory
// @Produce json
// @Security BearerAuth
// @Param id path string true "Location ID"
// @Param productId path string true "Product ID"
// @Param limit query int false "Page size, default 50, max 200"
// @Param offset query int false "Rows to skip"
// @Success 200 {array} Movement
// @Failure 400 {object} httpx.Error "Invalid pagination"
// @Failure 401 {object} httpx.Error "Unauthorized"
// @Failure 403 {object} httpx.Error "Forbidden"
// @Failure 404 {object} httpx.Error "Location not found"
// @Failure 500 {object} httpx.Error "Internal error"
// @Failure 503 {object} httpx.Error "Service unavailable"
// @Router /location/{id}/stock/{productId}/movements [get]
func (h *Handler) ListMovements(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), requestTimeout)
	defer stop()

	u, ok := httpx.CurrentUser(w, ctx)
	if !ok {
		return
	}
	limit, offset, ok := parsePage(w, r)
	if !ok {
		return
	}

	res, err := h.svc.ListMovements(ctx, MovementFilter{
		CompanyID:  u.CompanyId,
		LocationID: chi.URLParam(r, "id"),
		ProductID:  chi.URLParam(r, "productId"),
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, res)
}

// parsePage reads optional non-negative limit and offset; on bad input it
// writes 400 and returns false.
func parsePage(w http.ResponseWriter, r *http.Request) (limit, offset int, ok bool) {
	q := r.URL.Query()
	limit, okLimit := nonNegative(q.Get("limit"))
	offset, okOffset := nonNegative(q.Get("offset"))
	if !okLimit || !okOffset {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_pagination", "limit and offset must be non-negative integers")
		return 0, 0, false
	}
	return limit, offset, true
}

func nonNegative(s string) (int, bool) {
	if s == "" {
		return 0, true
	}
	n, err := strconv.Atoi(s)
	return n, err == nil && n >= 0
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, apperror.ErrInvalidCred):
		httpx.WriteError(w, http.StatusBadRequest, "invalid_credential", "invalid credential")
	case errors.Is(err, ErrLocationNotFound):
		httpx.WriteError(w, http.StatusNotFound, "location_not_found", "location not found")
	case errors.Is(err, context.DeadlineExceeded):
		httpx.WriteError(w, http.StatusServiceUnavailable, "service_unavailable", "service unavailable")
		h.logger.Error("inventory request timed out", slog.Any("err", err))
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "internal error")
		h.logger.Error("inventory request failed", slog.Any("err", err))
	}
}
