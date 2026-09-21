//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/mconcepcionb/ac-community-gw/internal/core/config"
	"github.com/mconcepcionb/ac-community-gw/internal/testsupport/integration"
)

func TestOpenAndCheck(t *testing.T) {
	url := integration.PostgresURL(t)
	db, err := Open(context.Background(), config.Database{
		URL:             url,
		MaxOpenConns:    2,
		MaxIdleConns:    1,
		ConnMaxLifetime: time.Minute,
	})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if db.Name() != "postgres" {
		t.Fatalf("name = %q", db.Name())
	}
	if err := db.Check(context.Background()); err != nil {
		t.Fatalf("Check: %v", err)
	}
	if err := db.SQL().Ping(); err != nil {
		t.Fatalf("SQL ping: %v", err)
	}
}

func TestOpenFailsWhenUnreachable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := Open(ctx, config.Database{URL: "postgres://nobody:nobody@127.0.0.1:1/none?sslmode=disable"}); err == nil {
		t.Fatal("expected Open to fail for an unreachable database")
	}
}
