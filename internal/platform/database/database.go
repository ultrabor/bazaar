package database

import (
	"bazaar/internal/platform/config"
	"bazaar/internal/platform/support/apperror"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Querier is the common subset of *pgxpool.Pool and pgx.Tx.
// Repositories run queries through it and do not care which one they got.
type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type beginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type txKey struct{}

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

// WithinTx runs fn inside a transaction stored in ctx.
// It commits when fn returns nil and rolls back on error or panic.
// A nested call reuses the transaction that is already in ctx.
func (d *Database) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return withinTx(ctx, d.db, fn)
}

// Conn returns the transaction from ctx, or the pool when there is none.
func (d *Database) Conn(ctx context.Context) Querier {
	if tx, ok := txFromContext(ctx); ok {
		return tx
	}
	return d.db
}

func (d *Database) Close() {
	d.db.Close()
}

func withinTx(ctx context.Context, b beginner, fn func(ctx context.Context) error) error {
	if _, ok := txFromContext(ctx); ok {
		return fn(ctx)
	}

	tx, err := b.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return errors.Join(err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}

func txFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}
