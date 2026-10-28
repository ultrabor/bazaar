package category

import (
	"bazaar/internal/platform/httpx"
	"bazaar/internal/platform/identity"
	"bazaar/internal/platform/support/apperror"
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
// @Param request body CreateCategoryBody true "Category data"
// @Success 201 {object} CreateCategoryResponse
// @Failure 400 {object} httpx.Error "Invalid JSON or empty name"
// @Failure 401 {object} httpx.Error "Unauthorized"
// @Failure 409 {object} httpx.Error "Active category name already exists"
// @Failure 500 {object} httpx.Error "Internal error"
// @Failure 503 {object} httpx.Error "Service unavailable"
// @Router /category [post]
func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var body CreateCategoryBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON")
		return
	}
	u, ok := identity.FromContext(ctx)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	response, err := h.svc.CreateCategory(ctx, &CreateCategoryRequest{
		CompanyId: u.CompanyId,
		Name:      body.Name,
	})
	if err != nil {
		h.writeCategoryError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, response)
}

// @Summary Get active categories for the current user's company
// @Tags Category
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {array} Category "Active categories; empty array when no categories match"
// @Failure 400 {object} httpx.Error "Invalid pagination"
// @Failure 401 {object} httpx.Error "Unauthorized"
// @Failure 500 {object} httpx.Error "Internal error"
// @Failure 503 {object} httpx.Error "Service unavailable"
// @Router /category [get]
func (h *Handler) GetAllCategories(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	u, ok := identity.FromContext(ctx)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	page, limit := 1, 20
	if raw := r.URL.Query().Get("page"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 {
			httpx.WriteError(w, http.StatusBadRequest, "invalid_pagination", "page must be greater than 0")
			return
		}
		page = value
	}
	if raw := r.URL.Query().Get("limit"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 100 {
			httpx.WriteError(w, http.StatusBadRequest, "invalid_pagination", "limit must be between 1 and 100")
			return
		}
		limit = value
	}

	response, err := h.svc.GetCompanyCategories(ctx, GetCompanyCategoriesRequest{
		CompanyId: u.CompanyId,
		Page:      page,
		Limit:     limit,
	})
	if err != nil {
		h.writeCategoryError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}

// @Summary Get a category by ID
// @Tags Category
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Success 200 {object} Category
// @Failure 400 {object} httpx.Error "Invalid category UUID"
// @Failure 401 {object} httpx.Error "Unauthorized"
// @Failure 404 {object} httpx.Error "Category not found"
// @Failure 500 {object} httpx.Error "Internal error"
// @Failure 503 {object} httpx.Error "Service unavailable"
// @Router /category/{id} [get]
func (h *Handler) GetCategoryById(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	u, ok := identity.FromContext(ctx)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	response, err := h.svc.GetCategoryById(ctx, chi.URLParam(r, "id"), u.CompanyId)
	if err != nil {
		h.writeCategoryError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}

// @Summary Rename a category
// @Tags Category
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Param request body UpdateCategoryBody true "Category data"
// @Success 200 {object} UpdateCategoryResponse
// @Failure 400 {object} httpx.Error "Invalid category UUID, JSON, or empty name"
// @Failure 401 {object} httpx.Error "Unauthorized"
// @Failure 404 {object} httpx.Error "Category not found"
// @Failure 409 {object} httpx.Error "Active category name already exists"
// @Failure 500 {object} httpx.Error "Internal error"
// @Failure 503 {object} httpx.Error "Service unavailable"
// @Router /category/{id} [put]
func (h *Handler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	var body UpdateCategoryBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "invalid JSON")
		return
	}
	u, ok := identity.FromContext(ctx)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	response, err := h.svc.UpdateCategory(ctx, &UpdateCategoryRequest{
		CategoryId: chi.URLParam(r, "id"),
		CompanyId:  u.CompanyId,
		Name:       body.Name,
	})
	if err != nil {
		h.writeCategoryError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}

// @Summary Archive a category
// @Description Products keep their category_id and remain visible. Archived categories cannot be assigned to other products.
// @Tags Category
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} httpx.Error "Invalid category UUID"
// @Failure 401 {object} httpx.Error "Unauthorized"
// @Failure 404 {object} httpx.Error "Category not found"
// @Failure 500 {object} httpx.Error "Internal error"
// @Failure 503 {object} httpx.Error "Service unavailable"
// @Router /category/{id}/archive [patch]
func (h *Handler) ArchiveCategory(w http.ResponseWriter, r *http.Request) {
	ctx, stop := context.WithTimeout(r.Context(), 2*time.Second)
	defer stop()

	u, ok := identity.FromContext(ctx)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	if err := h.svc.ArchiveCategory(ctx, chi.URLParam(r, "id"), u.CompanyId); err != nil {
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
	case errors.Is(err, apperror.ErrInvalidCred):
		status = http.StatusBadRequest
		code = "invalid_credential"
		message = "invalid credential"
	case errors.Is(err, ErrInvalidCategoryID):
		status = http.StatusBadRequest
		code = "invalid_category_id"
		message = "category id must be a valid UUID"
	case errors.Is(err, ErrInvalidCategoryName):
		status = http.StatusBadRequest
		code = "validation_error"
		message = "category name must not be empty"
	case errors.Is(err, ErrCategoryNotFound):
		status = http.StatusNotFound
		code = "category_not_found"
		message = "category not found"
	case errors.Is(err, ErrCategoryAlreadyExists):
		status = http.StatusConflict
		code = "category_name_conflict"
		message = "active category name already exists"
	}

	httpx.WriteError(w, status, code, message)
	if status >= http.StatusInternalServerError {
		h.logger.Error("category request failed", slog.Any("err", err))
	}
}

// Request bodies.

type CreateCategoryBody struct {
	Name string `json:"name"`
}

type UpdateCategoryBody struct {
	Name string `json:"name"`
}
