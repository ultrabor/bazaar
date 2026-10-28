package category

import (
	"bazaar/internal/platform/httpx/middleware"
	"context"
	"log/slog"

	"github.com/go-chi/chi/v5"
)

// categoryService is what the handlers call; Service implements it.
type categoryService interface {
	CreateCategory(ctx context.Context, rq *CreateCategoryRequest) (*CreateCategoryResponse, error)
	GetCompanyCategories(ctx context.Context, rq GetCompanyCategoriesRequest) ([]Category, error)
	GetCategoryById(ctx context.Context, categoryId, companyId string) (*Category, error)
	UpdateCategory(ctx context.Context, rq *UpdateCategoryRequest) (*UpdateCategoryResponse, error)
	ArchiveCategory(ctx context.Context, categoryId, companyId string) error
}

type Handler struct {
	logger *slog.Logger
	svc    categoryService
}

func NewHandler(logger *slog.Logger, svc categoryService) *Handler {
	return &Handler{logger: logger, svc: svc}
}

func (h *Handler) Routes(r chi.Router, mw *middleware.Middleware) {
	r.Group(func(r chi.Router) {
		r.Use(mw.Auth)

		r.Post("/category", mw.Permission(h.CreateCategory, "categories.create"))
		r.Get("/category", mw.Permission(h.GetAllCategories, "categories.view"))
		r.Get("/category/{id}", mw.Permission(h.GetCategoryById, "categories.view"))
		r.Put("/category/{id}", mw.Permission(h.UpdateCategory, "categories.edit"))
		r.Patch("/category/{id}/archive", mw.Permission(h.ArchiveCategory, "categories.archive"))
	})
}
