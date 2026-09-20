package fakeazerothcore

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mconcepcionb/ac-community-gw/internal/adapters/azerothsoap"
)

func TestUnauthenticatedEndpointsRequireAuth(t *testing.T) {
	_, httpServer := newRunningServer(t, Config{Username: "acgw", Password: "acgw"})

	tests := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/state"},
		{http.MethodGet, "/commands"},
		{http.MethodGet, "/commands/stream"},
		{http.MethodGet, "/"},
		{http.MethodPost, "/reset"},
		{http.MethodPost, "/"},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			httpServer.Config.Handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", rec.Code)
			}
		})
	}
}

func TestHealthzNeedsNoAuth(t *testing.T) {
	_, httpServer := newRunningServer(t, Config{Username: "acgw", Password: "acgw"})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	httpServer.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestStateRedactsPasswords(t *testing.T) {
	_, httpServer := newRunningServer(t, Config{Username: "acgw", Password: "acgw"})
	createAccount(t, httpServer.URL, ".account create Alice supersecret alice@example.com")

	resp := basicRequest(t, http.MethodGet, httpServer.URL+"/state", nil)
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	if strings.Contains(string(body), "supersecret") {
		t.Fatalf("password leaked in state: %s", body)
	}
	if strings.Contains(string(body), `"password"`) {
		t.Fatalf("password key present in state: %s", body)
	}
}

func TestJournalRedactsPasswords(t *testing.T) {
	_, httpServer := newRunningServer(t, Config{Username: "acgw", Password: "acgw"})
	createAccount(t, httpServer.URL, ".account create Alice supersecret alice@example.com")
	createAccount(t, httpServer.URL, ".account set password Alice anothersecret anothersecret")

	commands := fetchCommands(t, httpServer.URL)
	for _, entry := range commands {
		if strings.Contains(entry.Command, "supersecret") || strings.Contains(entry.Command, "anothersecret") {
			t.Fatalf("password leaked in journal: %q", entry.Command)
		}
		if !strings.Contains(entry.Command, "***") {
			t.Fatalf("password not redacted in journal: %q", entry.Command)
		}
	}
}

func TestBanRejectsUnknownDurationUnit(t *testing.T) {
	server := testServer(t)
	server.execute(".account create Bob pass")
	if got := server.execute(".ban account Bob 1x cheating"); got != msgBadValue {
		t.Fatalf("ban = %q, want %q", got, msgBadValue)
	}
	snapshot := server.Snapshot()
	if len(snapshot.Accounts) != 1 || snapshot.Accounts[0].Banned {
		t.Fatalf("account should not be banned: %+v", snapshot.Accounts)
	}
}

func TestAccountSetPasswordRejectsTooLong(t *testing.T) {
	server := testServer(t)
	server.execute(".account create Bob pass")
	long := strings.Repeat("x", maxPassword+1)
	if got := server.execute(".account set password Bob " + long + " " + long); got != msgAccountPassTooLong {
		t.Fatalf("set password = %q, want %q", got, msgAccountPassTooLong)
	}
}

func TestResetRemovesCreatedAccounts(t *testing.T) {
	server := New(Config{
		Username: "acgw",
		Password: "acgw",
		Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		Seed:     []Account{{Username: "ADMIN", Email: "admin@example.com"}},
	})
	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()

	createAccount(t, httpServer.URL, ".account create Temp pass")
	if len(server.Snapshot().Accounts) != 2 {
		t.Fatalf("accounts = %d, want 2", len(server.Snapshot().Accounts))
	}

	resp := basicRequest(t, http.MethodPost, httpServer.URL+"/reset", nil)
	defer func() { _ = resp.Body.Close() }()

	accounts := server.Snapshot().Accounts
	if len(accounts) != 1 || accounts[0].Username != "ADMIN" {
		t.Fatalf("accounts after reset = %+v", accounts)
	}
}

func TestAdapterCompatibilityStillWorks(t *testing.T) {
	server := New(Config{Username: "acgw", Password: "acgw", Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()

	client, err := azerothsoap.New(azerothsoap.Config{URL: httpServer.URL, Username: "acgw", Password: "acgw"})
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	if _, err := client.Execute(context.Background(), ".account create Dave pass"); err != nil {
		t.Fatalf("execute: %v", err)
	}
}
