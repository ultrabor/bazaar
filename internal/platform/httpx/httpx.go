package httpx

import (
	"bazaar/internal/app/health"
	"bazaar/internal/modules/auth"
	"log/slog"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	logger *slog.Logger
	router *chi.Mux
	sm     *ServiceManager
}

type ServiceManager struct {
	healthService *health.Service
	authService   *auth.Service
}

func NewServiceManager(
	healthService *health.Service,
	authService *auth.Service,
) *ServiceManager {
	return &ServiceManager{
		healthService: healthService,
		authService:   authService,
	}
}

func New(logger *slog.Logger, router *chi.Mux, sm *ServiceManager) *Handler {
	return &Handler{logger: logger, sm: sm, router: router}
}
