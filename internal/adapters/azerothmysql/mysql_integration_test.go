//go:build integration

package azerothmysql

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
)

// TestListAccounts reads the compose MariaDB fixture, which is read-only for the
// gateway user: the assertions target the seeded ADMIN/PLAYER/BANNED accounts.
func TestListAccounts(t *testing.T) {
	dsn := os.Getenv("ACGW_AZEROTH_LOGIN_DB_DSN")
	if dsn == "" {
		t.Skip("ACGW_AZEROTH_LOGIN_DB_DSN is not set")
	}
	client, err := New(Config{DSN: dsn})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = client.Close() }()
	ctx := context.Background()

	accounts, err := client.ListAccounts(ctx, azerothdb.Query{Filter: "ADMIN"})
	if err != nil {
		t.Fatalf("ListAccounts: %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("accounts = %d, want 1", len(accounts))
	}
	admin := accounts[0]
	if admin.Username != "ADMIN" || admin.Email != "admin@example.com" || admin.GMLevel != 3 || admin.Expansion != 2 {
		t.Fatalf("admin = %+v", admin)
	}

	found, err := client.FindAccountByUsername(ctx, "ADMIN")
	if err != nil || found.Username != "ADMIN" || found.GMLevel != 3 {
		t.Fatalf("FindAccountByUsername = %+v, %v", found, err)
	}

	banned, err := client.FindAccountByUsername(ctx, "BANNED")
	if err != nil || !banned.Banned || banned.BanReason == "" {
		t.Fatalf("banned = %+v, %v", banned, err)
	}

	if _, err := client.FindAccountByUsername(ctx, "NOSUCHACCOUNT"); !errors.Is(err, azerothdb.ErrAccountNotFound) {
		t.Fatalf("FindAccountByUsername missing = %v, want ErrAccountNotFound", err)
	}
}

func TestListAccountsPaginationAndEscaping(t *testing.T) {
	dsn := os.Getenv("ACGW_AZEROTH_LOGIN_DB_DSN")
	if dsn == "" {
		t.Skip("ACGW_AZEROTH_LOGIN_DB_DSN is not set")
	}
	client, err := New(Config{DSN: dsn})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = client.Close() }()
	ctx := context.Background()

	page, err := client.ListAccounts(ctx, azerothdb.Query{Limit: 1, Offset: 0})
	if err != nil {
		t.Fatalf("ListAccounts page: %v", err)
	}
	if len(page) != 1 {
		t.Fatalf("page = %d, want 1", len(page))
	}

	// A user-supplied LIKE wildcard must be treated literally, so '%' matches
	// no seeded username instead of matching everything.
	wildcard, err := client.ListAccounts(ctx, azerothdb.Query{Filter: "%"})
	if err != nil {
		t.Fatalf("ListAccounts wildcard: %v", err)
	}
	if len(wildcard) != 0 {
		t.Fatalf("wildcard filter leaked: %+v", wildcard)
	}

	none, err := client.ListAccounts(ctx, azerothdb.Query{Filter: "no-such-account"})
	if err != nil || len(none) != 0 {
		t.Fatalf("unexpected matches: %+v, %v", none, err)
	}
}
