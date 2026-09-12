package goose

import (
	"bazaar/internal/platform/config"
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

type Goose struct {
	pool *sql.DB
}

func New(cfg config.Config) (*Goose, error) {
	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", cfg.PostgresHost, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDB)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return &Goose{pool: db}, nil
}

func (g *Goose) Up(ctx context.Context) error {
	if err := goose.UpContext(ctx, g.pool, "migrations"); err != nil {
		return err
	}

	return nil
}

func (g *Goose) Close() {
	g.pool.Close()
}
