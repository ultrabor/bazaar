package httpx

import (
	"bazaar/internal/app/health"
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
}

func NewServiceManager(
	healthService *health.Service,
) *ServiceManager {
	return &ServiceManager{
		healthService: healthService,
	}
}

func New(logger *slog.Logger, router *chi.Mux, sm *ServiceManager) *Handler {
	return &Handler{logger: logger, sm: sm, router: router}
}
