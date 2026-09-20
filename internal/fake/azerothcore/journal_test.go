package fakeazerothcore

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mconcepcionb/ac-community-gw/internal/adapters/azerothsoap"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
)

func newRunningServer(t *testing.T, cfg Config) (*Server, *httptest.Server) {
	t.Helper()
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	server := New(cfg)
	httpServer := httptest.NewServer(server.Handler())
	t.Cleanup(httpServer.Close)
	return server, httpServer
}

func createAccount(t *testing.T, baseURL, command string) {
	t.Helper()
	client, err := azerothsoap.New(azerothsoap.Config{URL: baseURL, Username: "acgw", Password: "acgw"})
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	if _, err := client.Execute(context.Background(), command); err != nil {
		t.Fatalf("execute %q: %v", command, err)
	}
}

func basicRequest(t *testing.T, method, url string, body io.Reader) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	req.SetBasicAuth("acgw", "acgw")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	return resp
}

func fetchCommands(t *testing.T, baseURL string) []CommandEntry {
	t.Helper()
	resp := basicRequest(t, http.MethodGet, baseURL+"/commands", nil)
	defer func() { _ = resp.Body.Close() }()
	var payload struct {
		Commands []CommandEntry `json:"commands"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode commands: %v", err)
	}
	return payload.Commands
}

func TestCommandJournalRecordsRequestID(t *testing.T) {
	_, httpServer := newRunningServer(t, Config{Username: "acgw", Password: "acgw"})

	client, err := azerothsoap.New(azerothsoap.Config{URL: httpServer.URL, Username: "acgw", Password: "acgw"})
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	ctx := httpapi.ContextWithRequestID(context.Background(), "req-42")
	if _, err := client.Execute(ctx, ".server info"); err != nil {
		t.Fatalf("execute: %v", err)
	}

	commands := fetchCommands(t, httpServer.URL)
	if len(commands) != 1 {
		t.Fatalf("commands = %d", len(commands))
	}
	entry := commands[0]
	if entry.RequestID != "req-42" {
		t.Fatalf("request id = %q", entry.RequestID)
	}
	if entry.Command != ".server info" {
		t.Fatalf("command = %q", entry.Command)
	}
	if entry.Result == "" {
		t.Fatal("result is empty")
	}
}

func TestResetRestoresSeedAndClearsJournal(t *testing.T) {
	_, httpServer := newRunningServer(t, Config{
		Username: "acgw",
		Password: "acgw",
		Seed:     []Account{{Username: "ADMIN", Password: "ADMIN", GMLevel: 3}},
	})

	createAccount(t, httpServer.URL, ".account create Temp pass")
	if len(fetchCommands(t, httpServer.URL)) != 1 {
		t.Fatal("expected one journal entry")
	}

	resp := basicRequest(t, http.MethodPost, httpServer.URL+"/reset", nil)
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("reset status = %d", resp.StatusCode)
	}

	if len(fetchCommands(t, httpServer.URL)) != 0 {
		t.Fatal("journal should be empty after reset")
	}

	stateResp := basicRequest(t, http.MethodGet, httpServer.URL+"/state", nil)
	defer func() { _ = stateResp.Body.Close() }()
	var seed Seed
	if err := json.NewDecoder(stateResp.Body).Decode(&seed); err != nil {
		t.Fatalf("decode state: %v", err)
	}
	if len(seed.Accounts) != 1 || seed.Accounts[0].Username != "ADMIN" {
		t.Fatalf("state = %+v", seed)
	}
}

func TestJournalRingCapacity(t *testing.T) {
	server := New(Config{JournalCapacity: 2, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	for _, command := range []string{".server info", ".server info", ".server info"} {
		server.journal.append(CommandEntry{Command: command})
	}
	entries := server.journal.list(10)
	if len(entries) != 2 {
		t.Fatalf("entries = %d", len(entries))
	}
	if entries[0].Seq != 3 || entries[1].Seq != 2 {
		t.Fatalf("seq = %d, %d", entries[0].Seq, entries[1].Seq)
	}
}

func TestCORSHeaders(t *testing.T) {
	_, httpServer := newRunningServer(t, Config{AllowedOrigins: []string{"http://localhost:5173"}})

	req := httptest.NewRequest(http.MethodOptions, "/commands", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	httpServer.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatalf("allow origin = %q", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSRejectsUnknownOrigin(t *testing.T) {
	_, httpServer := newRunningServer(t, Config{AllowedOrigins: []string{"http://localhost:5173"}})

	req := httptest.NewRequest(http.MethodGet, "/commands", nil)
	req.Header.Set("Origin", "https://evil.test")
	rec := httptest.NewRecorder()
	httpServer.Config.Handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("allow origin = %q, want empty", got)
	}
}
