package httpx

import (
	"bazaar/internal/app/health"
	"bazaar/internal/domain"
	"bazaar/internal/modules/auth"
	"bazaar/internal/modules/permission"
	"log/slog"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	logger *slog.Logger
	router *chi.Mux
	sm     *ServiceManager
}

type ServiceManager struct {
	healthService     *health.Service
	authService       *auth.Service
	userService       domain.UserService
	locationService   domain.LocationService
	productService    domain.ProductService
	categoryService   domain.CategoryService
	permissionService *permission.Service
}

func NewServiceManager(
	healthService *health.Service,
	authService *auth.Service,
	userService domain.UserService,
	locationService domain.LocationService,
	productService domain.ProductService,
	categoryService domain.CategoryService,
	permissionService *permission.Service,
) *ServiceManager {
	return &ServiceManager{
		healthService:     healthService,
		authService:       authService,
		userService:       userService,
		locationService:   locationService,
		productService:    productService,
		categoryService:   categoryService,
		permissionService: permissionService,
	}
}

func New(logger *slog.Logger, router *chi.Mux, sm *ServiceManager) *Handler {
	return &Handler{logger: logger, sm: sm, router: router}
}
