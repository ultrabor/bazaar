package health

import (
	"bazaar/internal/platform/database"
	"context"
)

type Service struct {
	db *database.Database
}

func New(db *database.Database) *Service {
	return &Service{db: db}
}

func (s *Service) Ready(ctx context.Context) error {
	return s.db.Ping(ctx)
}
