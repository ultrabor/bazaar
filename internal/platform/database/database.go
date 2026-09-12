package database

import (
	"bazaar/internal/platform/config"
	"bazaar/internal/platform/support/apperror"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func New(ctx context.Context, cfg config.Config, log *slog.Logger) (*Database, error) {
	url := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresDB,
	)

	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}

	config.MaxConns = int32(cfg.PostgresMaxConn)
	config.MaxConnIdleTime = 25 * time.Second
	config.HealthCheckPeriod = 25 * time.Second
	config.MaxConnLifetime = 5 * time.Minute

	dbPool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	if err = dbPool.Ping(ctx); err != nil {
		dbPool.Close()
		return nil, err
	}

	db := Database{
		db:  dbPool,
		log: log,
	}

	return &db, nil
}

func (d *Database) Ping(ctx context.Context) error {
	if err := d.db.Ping(ctx); err != nil {
		return apperror.New("database unavailable", err)
	}
	return nil
}

func (d *Database) Close() {
	d.db.Close()
}
