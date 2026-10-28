package product

import (
	"bazaar/internal/platform/httpx/middleware"
	"context"
	"log/slog"

	"github.com/go-chi/chi/v5"
)

// productService is what the handlers call; Service implements it.
type productService interface {
	GetProductById(ctx context.Context, productId, companyId string) (*Product, error)
	GetCompanyProducts(ctx context.Context, rq GetCompanyProductsRequest) ([]Product, error)
	CreateProduct(ctx context.Context, rq *CreateProductRequest) (*CreateProductResponse, error)
	UpdateProduct(ctx context.Context, rq *UpdateProductRequest) (*UpdateProductResponse, error)
	ArchiveProduct(ctx context.Context, productId, companyId string) error
}

type Handler struct {
	logger *slog.Logger
	svc    productService
}

func NewHandler(logger *slog.Logger, svc productService) *Handler {
	return &Handler{logger: logger, svc: svc}
}

func (h *Handler) Routes(r chi.Router, mw *middleware.Middleware) {
	r.Group(func(r chi.Router) {
		r.Use(mw.Auth)

		r.Post("/product", mw.Permission(h.CreateProduct, "products.create"))
		r.Get("/product", mw.Permission(h.GetAllProducts, "products.view"))
		r.Get("/product/{id}", mw.Permission(h.GetProductById, "products.view"))
		r.Put("/product/{id}", mw.Permission(h.UpdateProduct, "products.edit"))
		r.Patch("/product/{id}/archive", mw.Permission(h.ArchiveProduct, "products.archive"))
	})
}
