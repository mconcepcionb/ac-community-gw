//go:build integration

package azerothmysql

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
)

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

	username := fmt.Sprintf("ITTEST%d", time.Now().UnixNano()%1_000_000)
	_, err = client.db.Exec(`
		INSERT INTO account (username, salt, verifier, email, reg_mail, expansion)
		VALUES (?, UNHEX(REPEAT('00', 32)), UNHEX(REPEAT('00', 32)), ?, ?, 2)`,
		username, "it@example.com", "it@example.com")
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	t.Cleanup(func() {
		_, _ = client.db.Exec(`DELETE FROM account WHERE username = ?`, username)
	})

	accounts, err := client.ListAccounts(context.Background(), azerothdb.Query{Filter: username})
	if err != nil {
		t.Fatalf("ListAccounts: %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("accounts = %d", len(accounts))
	}
	account := accounts[0]
	if account.Username != username || account.Email != "it@example.com" || account.Expansion != 2 {
		t.Fatalf("account = %+v", account)
	}
}
