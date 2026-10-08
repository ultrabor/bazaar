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
	"strconv"
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
// @Failure 400 {object} dto.Error "Invalid JSON or product fields"
// @Failure 401 {object} dto.Error "Unauthorized"
// @Failure 409 {object} dto.Error "Duplicate SKU"
// @Failure 500 {object} dto.Error "Internal error"
// @Failure 503 {object} dto.Error "Service unavailable"
// @Router /product [post]
func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var body dto.CreateProductRequest

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		dto.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON")
		return
	}

	u, ok := middleware.CurrentUser(ctx)
	if !ok {
		dto.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
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
		code := "internal_error"
		msg := "internal error"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			code = "service_unavailable"
			msg = "service unavailable"
		case errors.Is(err, product.ErrInvalidCred):
			status = http.StatusBadRequest
			code = "validation_error"
			msg = "invalid product fields"
		case errors.Is(err, product.ErrProductAlreadyExists):
			status = http.StatusConflict
			code = "sku_conflict"
			msg = "SKU already exists"
		case errors.Is(err, product.ErrInvalidUnitType):
			status = http.StatusBadRequest
			code = "validation_error"
			msg = "invalid product unit type"
		}

		dto.WriteError(w, status, code, msg)
		h.logger.Error("service unavailable", slog.Any("err", err))
		return
	}

	re, err := json.Marshal(res)
	if err != nil {
		dto.WriteError(w, http.StatusInternalServerError, "internal_error", "internal error")
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
// @Failure 400 {object} dto.Error "Invalid product ID or credential"
// @Failure 401 {object} dto.Error "Unauthorized"
// @Failure 404 {object} dto.Error "Product not found"
// @Failure 500 {object} dto.Error "Internal error"
// @Failure 503 {object} dto.Error "Service unavailable"
// @Router /product/{id} [get]
func (h *Handler) GetProductById(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	ProductId := chi.URLParam(r, "id")
	if ProductId == "" {
		dto.WriteError(w, http.StatusBadRequest, "product_id_required", "product id is required")
		h.logger.Error("Product id is required")
		return
	}

	u, ok := middleware.CurrentUser(ctx)

	if !ok {
		dto.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		h.logger.Error("unauthorized")
		return
	}

	res, err := h.sm.productService.GetProductById(ctx, ProductId, u.CompanyId)

	if err != nil {
		status := http.StatusInternalServerError
		code := "internal_error"
		msg := "internal error"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			code = "service_unavailable"
			msg = "service unavailable"
		case errors.Is(err, product.ErrInvalidCred):
			status = http.StatusBadRequest
			code = "invalid_credential"
			msg = "invalid credential"
		case errors.Is(err, product.ErrNotFound):
			status = http.StatusNotFound
			code = "product_not_found"
			msg = "product not found"
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
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(re)
}

// @Summary Get all Products for the current user's company
// @Tags Product
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {array} product.Product
// @Failure 400 {object} dto.Error "Invalid credential or pagination"
// @Failure 401 {object} dto.Error "Unauthorized"
// @Failure 404 {object} dto.Error "Product not found"
// @Failure 500 {object} dto.Error "Internal error"
// @Failure 503 {object} dto.Error "Service unavailable"
// @Router /product [get]
func (h *Handler) GetAllProducts(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	u, ok := middleware.CurrentUser(ctx)

	if !ok {
		dto.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		h.logger.Error("unauthorized")
		return
	}

	page, limit := 1, 20

	if raw := r.URL.Query().Get("page"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 {
			dto.WriteError(w, http.StatusBadRequest, "invalid_pagination", "page must be greater than 0")
			return
		}
		page = value
	}

	if raw := r.URL.Query().Get("limit"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 100 {
			dto.WriteError(w, http.StatusBadRequest, "invalid_pagination", "limit must be between 1 and 100")
			return
		}
		limit = value
	}

	rq := product.GetCompanyProductsRequest{
		CompanyId: u.CompanyId,
		Page:      page,
		Limit:     limit,
	}

	res, err := h.sm.productService.GetCompanyProducts(ctx, rq)

	if err != nil {
		status := http.StatusInternalServerError
		code := "internal_error"
		msg := "internal error"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			code = "service_unavailable"
			msg = "service unavailable"
		case errors.Is(err, product.ErrInvalidCred):
			status = http.StatusBadRequest
			code = "invalid_credential"
			msg = "invalid credential"
		case errors.Is(err, product.ErrNotFound):
			status = http.StatusNotFound
			code = "product_not_found"
			msg = "product not found"
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

// @Summary Update Product by ID
// @Tags Product
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Param request body dto.UpdateProductRequest true "Product update data"
// @Success 200 {object} product.UpdateProductResponse
// @Failure 400 {object} dto.Error "Invalid JSON or product fields"
// @Failure 401 {object} dto.Error "Unauthorized"
// @Failure 404 {object} dto.Error "Product not found"
// @Failure 409 {object} dto.Error "Duplicate SKU"
// @Failure 500 {object} dto.Error "Internal error"
// @Failure 503 {object} dto.Error "Service unavailable"
// @Router /product/{id} [put]
func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var body dto.UpdateProductRequest

	var ProductId = chi.URLParam(r, "id")

	if ProductId == "" {
		dto.WriteError(w, http.StatusBadRequest, "product_id_required", "product id is required")
		h.logger.Error("Product id is required")
		return
	}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		dto.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON")
		return
	}

	u, ok := middleware.CurrentUser(ctx)

	if !ok {
		dto.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
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
		code := "internal_error"
		msg := "internal error"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			code = "service_unavailable"
			msg = "service unavailable"
		case errors.Is(err, product.ErrInvalidCred):
			status = http.StatusBadRequest
			code = "validation_error"
			msg = "invalid product fields"
		case errors.Is(err, product.ErrNotFound):
			status = http.StatusNotFound
			code = "product_not_found"
			msg = "product not found"
		case errors.Is(err, product.ErrProductAlreadyExists):
			status = http.StatusConflict
			code = "sku_conflict"
			msg = "SKU already exists"
		case errors.Is(err, product.ErrInvalidUnitType):
			status = http.StatusBadRequest
			code = "validation_error"
			msg = "invalid product unit type"
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
// @Failure 400 {object} dto.Error "Invalid credential data"
// @Failure 401 {object} dto.Error "Unauthorized"
// @Failure 404 {object} dto.Error "Product not found"
// @Failure 500 {object} dto.Error "Internal error"
// @Failure 503 {object} dto.Error "Service unavailable"
// @Router /product/{id}/archive [patch]
func (h *Handler) ArchiveProduct(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var ProductId = chi.URLParam(r, "id")

	if ProductId == "" {
		dto.WriteError(w, http.StatusBadRequest, "product_id_required", "product id is required")
		h.logger.Error("Product id is required")
		return
	}

	u, ok := middleware.CurrentUser(ctx)

	if !ok {
		dto.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		h.logger.Error("unauthorized")
		return
	}

	err := h.sm.productService.ArchiveProduct(ctx, ProductId, u.CompanyId)
	if err != nil {
		status := http.StatusInternalServerError
		code := "internal_error"
		msg := "internal error"
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			status = http.StatusServiceUnavailable
			code = "service_unavailable"
			msg = "service unavailable"
		case errors.Is(err, product.ErrInvalidCred):
			status = http.StatusBadRequest
			code = "invalid_credential"
			msg = "invalid credential"
		case errors.Is(err, product.ErrNotFound):
			status = http.StatusNotFound
			code = "product_not_found"
			msg = "product not found"
		}

		dto.WriteError(w, status, code, msg)
		h.logger.Error("service unavailable", slog.Any("err", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
