package httpx

import (
	"bazaar/internal/app/health"
	"bazaar/internal/domain"
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
	authService       domain.AuthService
	userService       domain.UserService
	locationService   domain.LocationService
	productService    domain.ProductService
	categoryService   domain.CategoryService
	permissionService domain.PermissionService
}

func NewServiceManager(
	healthService *health.Service,
	authService domain.AuthService,
	userService domain.UserService,
	locationService domain.LocationService,
	productService domain.ProductService,
	categoryService domain.CategoryService,
	permissionService domain.PermissionService,
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
