package postgres

import (
	"context"
	"testing"

	"github.com/mconcepcionb/ac-community-gw/internal/core/config"
)

func TestOpenRequiresURL(t *testing.T) {
	if _, err := Open(context.Background(), config.Database{}); err == nil {
		t.Fatal("expected error for empty database URL")
	}
}

func TestUnavailableCheckReportsError(t *testing.T) {
	check := Unavailable{Err: context.DeadlineExceeded}
	if check.Name() != "postgres" {
		t.Fatalf("name = %q", check.Name())
	}
	if err := check.Check(context.Background()); err == nil {
		t.Fatal("expected the stored error")
	}
}
