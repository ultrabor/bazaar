package httpx

import (
	"bazaar/internal/modules/category"
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

// @Summary Create a category
// @Tags Category
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateCategoryRequest true "Category data"
// @Success 201 {object} category.CreateCategoryResponse
// @Failure 400 {object} dto.Error "Invalid JSON or empty name"
// @Failure 401 {object} dto.Error "Unauthorized"
// @Failure 409 {object} dto.Error "Active category name already exists"
// @Failure 500 {object} dto.Error "Internal error"
// @Failure 503 {object} dto.Error "Service unavailable"
// @Router /category [post]
func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var body dto.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		dto.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON")
		return
	}
	u, ok := middleware.CurrentUser(ctx)
	if !ok {
		dto.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	response, err := h.sm.categoryService.CreateCategory(ctx, &category.CreateCategoryRequest{
		CompanyId: u.CompanyId,
		Name:      body.Name,
	})
	if err != nil {
		h.writeCategoryError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, response)
}

// @Summary Get active categories for the current user's company
// @Tags Category
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {array} category.Category "Active categories; empty array when no categories match"
// @Failure 400 {object} dto.Error "Invalid pagination"
// @Failure 401 {object} dto.Error "Unauthorized"
// @Failure 500 {object} dto.Error "Internal error"
// @Failure 503 {object} dto.Error "Service unavailable"
// @Router /category [get]
func (h *Handler) GetAllCategories(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	u, ok := middleware.CurrentUser(ctx)
	if !ok {
		dto.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
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

	response, err := h.sm.categoryService.GetCompanyCategories(ctx, category.GetCompanyCategoriesRequest{
		CompanyId: u.CompanyId,
		Page:      page,
		Limit:     limit,
	})
	if err != nil {
		h.writeCategoryError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

// @Summary Get a category by ID
// @Tags Category
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Success 200 {object} category.Category
// @Failure 400 {object} dto.Error "Invalid category UUID"
// @Failure 401 {object} dto.Error "Unauthorized"
// @Failure 404 {object} dto.Error "Category not found"
// @Failure 500 {object} dto.Error "Internal error"
// @Failure 503 {object} dto.Error "Service unavailable"
// @Router /category/{id} [get]
func (h *Handler) GetCategoryById(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	u, ok := middleware.CurrentUser(ctx)
	if !ok {
		dto.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	response, err := h.sm.categoryService.GetCategoryById(ctx, chi.URLParam(r, "id"), u.CompanyId)
	if err != nil {
		h.writeCategoryError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

// @Summary Rename a category
// @Tags Category
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Param request body dto.UpdateCategoryRequest true "Category data"
// @Success 200 {object} category.UpdateCategoryResponse
// @Failure 400 {object} dto.Error "Invalid category UUID, JSON, or empty name"
// @Failure 401 {object} dto.Error "Unauthorized"
// @Failure 404 {object} dto.Error "Category not found"
// @Failure 409 {object} dto.Error "Active category name already exists"
// @Failure 500 {object} dto.Error "Internal error"
// @Failure 503 {object} dto.Error "Service unavailable"
// @Router /category/{id} [put]
func (h *Handler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var body dto.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		dto.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON")
		return
	}
	u, ok := middleware.CurrentUser(ctx)
	if !ok {
		dto.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	response, err := h.sm.categoryService.UpdateCategory(ctx, &category.UpdateCategoryRequest{
		CategoryId: chi.URLParam(r, "id"),
		CompanyId:  u.CompanyId,
		Name:       body.Name,
	})
	if err != nil {
		h.writeCategoryError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

// @Summary Archive a category
// @Description Products keep their category_id and remain visible. Archived categories cannot be assigned to other products.
// @Tags Category
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} dto.Error "Invalid category UUID"
// @Failure 401 {object} dto.Error "Unauthorized"
// @Failure 404 {object} dto.Error "Category not found"
// @Failure 500 {object} dto.Error "Internal error"
// @Failure 503 {object} dto.Error "Service unavailable"
// @Router /category/{id}/archive [patch]
func (h *Handler) ArchiveCategory(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	u, ok := middleware.CurrentUser(ctx)
	if !ok {
		dto.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	if err := h.sm.categoryService.ArchiveCategory(ctx, chi.URLParam(r, "id"), u.CompanyId); err != nil {
		h.writeCategoryError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) writeCategoryError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	code := "internal_error"
	message := "internal error"

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		status = http.StatusServiceUnavailable
		code = "service_unavailable"
		message = "service unavailable"
	case errors.Is(err, category.ErrInvalidCredentials):
		status = http.StatusBadRequest
		code = "invalid_credential"
		message = "invalid credential"
	case errors.Is(err, category.ErrInvalidCategoryID):
		status = http.StatusBadRequest
		code = "invalid_category_id"
		message = "category id must be a valid UUID"
	case errors.Is(err, category.ErrInvalidName):
		status = http.StatusBadRequest
		code = "validation_error"
		message = "category name must not be empty"
	case errors.Is(err, category.ErrNotFound):
		status = http.StatusNotFound
		code = "category_not_found"
		message = "category not found"
	case errors.Is(err, category.ErrAlreadyExists):
		status = http.StatusConflict
		code = "category_name_conflict"
		message = "active category name already exists"
	}

	dto.WriteError(w, status, code, message)
	if status >= http.StatusInternalServerError {
		h.logger.Error("category request failed", slog.Any("err", err))
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	payload, err := json.Marshal(value)
	if err != nil {
		dto.WriteError(w, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(payload)
}
