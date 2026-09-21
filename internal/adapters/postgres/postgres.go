// Package postgres implements the gateway's PostgreSQL persistence adapter.
//
// It uses database/sql with the pgx driver and applies Goose migrations. It is
// the only place that knows how to open and migrate the gateway-owned schema.
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	// Register the pgx stdlib driver with database/sql.
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/mconcepcionb/ac-community-gw/internal/core/config"
)

// DB wraps the database/sql connection pool.
type DB struct {
	db *sql.DB
}

// Open opens and verifies a PostgreSQL connection pool.
func Open(ctx context.Context, cfg config.Database) (*DB, error) {
	if cfg.URL == "" {
		return nil, errors.New("postgres: database URL is empty")
	}
	pool, err := sql.Open("pgx", cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("postgres: open: %w", err)
	}
	if cfg.MaxOpenConns > 0 {
		pool.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		pool.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		pool.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	if err := pool.PingContext(ctx); err != nil {
		_ = pool.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	return &DB{db: pool}, nil
}

// SQL exposes the underlying pool for repositories.
func (d *DB) SQL() *sql.DB { return d.db }

// Close closes the pool.
func (d *DB) Close() error { return d.db.Close() }

// Name implements persistence.Check.
func (d *DB) Name() string { return "postgres" }

// Check implements persistence.Check.
func (d *DB) Check(ctx context.Context) error { return d.db.PingContext(ctx) }

// Unavailable is a readiness check used when the configured database could not
// be reached at startup.
type Unavailable struct {
	Err error
}

// Name implements persistence.Check.
func (Unavailable) Name() string { return "postgres" }

// Check implements persistence.Check.
func (u Unavailable) Check(context.Context) error { return u.Err }
