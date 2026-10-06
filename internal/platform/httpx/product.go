package httpx

import (
	"bazaar/internal/modules/product"
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

// @Summary Create a new Product
// @Tags Product
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateProductRequest true "Product creation data"
// @Success 201 {object} product.CreateProductResponse
// @Failure 400 {string} string "Invalid credential data"
// @Failure 401 {string} string "Unauthorized"
// @Failure 409 {string} string "Product already exists"
// @Failure 422 {string} string "Invalid input"
// @Failure 500 {string} string "Internal error"
// @Failure 503 {string} string "Service unavailable"
// @Router /product [post]
func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var body dto.CreateProductRequest

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

	rq := product.CreateProductRequest{
		CompanyId: u.CompanyId,
		Name:      body.Name,
		SKU:       body.SKU,
		Unit:      body.Unit,
	}

	res, err := h.sm.productService.CreateProduct(ctx, &rq)
	if err != nil {
		status := http.StatusInternalServerError
		msg := "DB closed"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			msg = "deadline exceeded"
		case errors.Is(err, product.ErrInvalidCred):
			status = http.StatusBadRequest
			msg = "invalid credential"
		case errors.Is(err, product.ErrProductAlreadyExists):
			status = http.StatusConflict
			msg = "Product already exists"
		case errors.Is(err, product.ErrInvalidUnitType):
			status = http.StatusBadRequest
			msg = "invalid product unit type"
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

// @Summary Get Product by ID
// @Tags Product
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Success 200 {object} product.Product
// @Failure 400 {string} string "Invalid credential data"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Product not found"
// @Failure 500 {string} string "Internal error"
// @Failure 503 {string} string "Service unavailable"
// @Router /product/{id} [get]
func (h *Handler) GetProductById(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	ProductId := chi.URLParam(r, "id")
	if ProductId == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Product id is required"))
		h.logger.Error("Product id is required")
		return
	}

	u, ok := middleware.CurrentUser(ctx)

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
		h.logger.Error("unauthorized")
		return
	}

	res, err := h.sm.productService.GetProductById(ctx, ProductId, u.CompanyId)

	if err != nil {
		status := http.StatusInternalServerError
		msg := "DB closed"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			msg = "deadline exceeded"
		case errors.Is(err, product.ErrInvalidCred):
			status = http.StatusBadRequest
			msg = "invalid credential"
		case errors.Is(err, product.ErrNotFound):
			status = http.StatusNotFound
			msg = "Product not found"
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

// @Summary Get all Products for the current user's company
// @Tags Product
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} product.Product
// @Failure 400 {string} string "Invalid credential data"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Product not found"
// @Failure 500 {string} string "Internal error"
// @Failure 503 {string} string "Service unavailable"
// @Router /product [get]
func (h *Handler) GetAllProducts(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	u, ok := middleware.CurrentUser(ctx)

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
		h.logger.Error("unauthorized")
		return
	}

	res, err := h.sm.productService.GetCompanyProducts(ctx, u.CompanyId)

	if err != nil {
		status := http.StatusInternalServerError
		msg := "DB closed"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			msg = "deadline exceeded"
		case errors.Is(err, product.ErrInvalidCred):
			status = http.StatusBadRequest
			msg = "invalid credential"
		case errors.Is(err, product.ErrNotFound):
			status = http.StatusNotFound
			msg = "Product not found"
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

// @Summary Update Product by ID
// @Tags Product
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Param request body dto.UpdateProductRequest true "Product update data"
// @Success 200 {object} product.UpdateProductResponse
// @Failure 400 {string} string "Invalid credential data"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Product not found"
// @Failure 409 {string} string "Product already exists"
// @Failure 422 {string} string "Invalid input"
// @Failure 500 {string} string "Internal error"
// @Failure 503 {string} string "Service unavailable"
// @Router /product/{id} [put]
func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var body dto.UpdateProductRequest

	var ProductId = chi.URLParam(r, "id")

	if ProductId == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Product id is required"))
		h.logger.Error("Product id is required")
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

	rq := product.UpdateProductRequest{
		ProductId: ProductId,
		CompanyId: u.CompanyId,
		Name:      body.Name,
		SKU:       body.SKU,
		Unit:      body.Unit,
	}

	res, err := h.sm.productService.UpdateProduct(ctx, &rq)
	if err != nil {
		status := http.StatusInternalServerError
		msg := "DB closed"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			msg = "deadline exceeded"
		case errors.Is(err, product.ErrInvalidCred):
			status = http.StatusBadRequest
			msg = "invalid credential"
		case errors.Is(err, product.ErrNotFound):
			status = http.StatusNotFound
			msg = "Product not found"
		case errors.Is(err, product.ErrProductAlreadyExists):
			status = http.StatusConflict
			msg = "Product already exists"
		case errors.Is(err, product.ErrInvalidUnitType):
			status = http.StatusBadRequest
			msg = "invalid product unit type"
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

// @Summary Archive Product by ID
// @Tags Product
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {string} string "Invalid credential data"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Product not found"
// @Failure 500 {string} string "Internal error"
// @Failure 503 {string} string "Service unavailable"
// @Router /product/{id}/archive [patch]
func (h *Handler) ArchiveProduct(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var ProductId = chi.URLParam(r, "id")

	if ProductId == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Product id is required"))
		h.logger.Error("Product id is required")
		return
	}

	u, ok := middleware.CurrentUser(ctx)

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
		h.logger.Error("unauthorized")
		return
	}

	err := h.sm.productService.ArchiveProduct(ctx, ProductId, u.CompanyId)
	if err != nil {
		status := http.StatusInternalServerError
		msg := "DB closed"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			msg = "deadline exceeded"
		case errors.Is(err, product.ErrInvalidCred):
			status = http.StatusBadRequest
			msg = "invalid credential"
		case errors.Is(err, product.ErrNotFound):
			status = http.StatusNotFound
			msg = "Product not found"
		}

		w.WriteHeader(status)
		_, _ = w.Write([]byte(msg))
		h.logger.Error("service unavailable", slog.Any("err", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
