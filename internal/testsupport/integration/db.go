//go:build integration

// Package integration provides shared helpers for the integration test suite.
//
// It is compiled only with the `integration` build tag. Each helper skips the
// test when the database is not configured, so the suite is safe to run
// anywhere; `task test:integration` runs it against the Compose databases.
package integration

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Postgres opens the gateway database from ACGW_DATABASE_URL.
func Postgres(t testing.TB) *sql.DB {
	t.Helper()
	return open(t, "ACGW_DATABASE_URL", "pgx")
}

// PostgresURL returns ACGW_DATABASE_URL, skipping when unset.
func PostgresURL(t testing.TB) string {
	t.Helper()
	url := os.Getenv("ACGW_DATABASE_URL")
	if url == "" {
		t.Skip("ACGW_DATABASE_URL is not set")
	}
	return url
}

// MySQL opens the database named by envVar (a go-sql-driver/mysql DSN).
func MySQL(t testing.TB, envVar string) *sql.DB {
	t.Helper()
	return open(t, envVar, "mysql")
}

func open(t testing.TB, envVar, driver string) *sql.DB {
	t.Helper()
	dsn := os.Getenv(envVar)
	if dsn == "" {
		t.Skipf("%s is not set", envVar)
	}
	db, err := sql.Open(driver, dsn)
	if err != nil {
		t.Fatalf("open %s: %v", envVar, err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("ping %s: %v", envVar, err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}
