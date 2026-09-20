package azerothmysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

const (
	defaultLimit = 100
	maxLimit     = 500
	// maxOffset bounds deep pagination so a request cannot force an unbounded
	// table scan.
	maxOffset    = 100_000
	pingTimeout  = 10 * time.Second
	connMaxIdle  = 5 * time.Minute
	dsnParseTime = true
)

// openPool parses and normalizes the DSN, then opens and verifies a read-only
// pool. Time parsing is always enabled so timestamp columns scan correctly, and
// multi-statement execution is rejected.
func openPool(dsn string, cfg Config, label string) (*sql.DB, error) {
	trimmed := strings.TrimSpace(dsn)
	if trimmed == "" {
		return nil, fmt.Errorf("azerothmysql: %s dsn must not be empty", label)
	}
	parsed, err := mysql.ParseDSN(trimmed)
	if err != nil {
		return nil, fmt.Errorf("azerothmysql: parse %s dsn: %w", label, err)
	}
	if parsed.MultiStatements {
		return nil, fmt.Errorf("azerothmysql: %s dsn must not enable multiStatements", label)
	}
	parsed.ParseTime = dsnParseTime
	parsed.InterpolateParams = false
	if parsed.Loc == nil {
		parsed.Loc = time.UTC
	}

	db, err := sql.Open("mysql", parsed.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("azerothmysql: open %s: %w", label, err)
	}
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}
	db.SetConnMaxIdleTime(connMaxIdle)

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("azerothmysql: ping %s: %w", label, err)
	}
	return db, nil
}

func clampLimit(limit int) int {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}

func clampOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	if offset > maxOffset {
		return maxOffset
	}
	return offset
}

// escapeLike escapes LIKE metacharacters so user input is matched literally.
// MySQL uses backslash as the default LIKE escape character.
func escapeLike(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(value)
}
