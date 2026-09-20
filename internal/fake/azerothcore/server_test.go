package fakeazerothcore

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mconcepcionb/ac-community-gw/internal/adapters/azerothsoap"
)

func testServer(t *testing.T) *Server {
	t.Helper()
	return New(Config{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
}

func TestAccountLifecycle(t *testing.T) {
	server := testServer(t)

	cases := []struct {
		command string
		want    string
	}{
		{".account create Alice secret alice@example.com", "Account created: Alice"},
		{".account create Alice secret", msgAccountAlreadyExists},
		{".account set password Alice newpass newpass", msgPasswordChanged},
		{".account set password Alice one two", msgPasswordsDoNotMatch},
		{".account set password Ghost pass pass", "Account not exist: Ghost"},
		{".account set email Alice bob@example.com", msgEmailChanged},
		{".account set gmlevel Alice 3 -1", "You change security level of account Alice to 3."},
	}
	for _, testCase := range cases {
		if got := server.execute(testCase.command); got != testCase.want {
			t.Fatalf("%q = %q, want %q", testCase.command, got, testCase.want)
		}
	}

	snapshot := server.Snapshot()
	if len(snapshot.Accounts) != 1 {
		t.Fatalf("accounts = %d", len(snapshot.Accounts))
	}
	account := snapshot.Accounts[0]
	if account.Password != "newpass" || account.Email != "bob@example.com" || account.GMLevel != 3 {
		t.Fatalf("account = %+v", account)
	}
}

func TestBanAndUnban(t *testing.T) {
	server := testServer(t)
	server.execute(".account create Bob pass")

	if got := server.execute(".ban account Bob 1d cheating"); got != "Bob is banned for 1 Day(s). Reason: cheating." {
		t.Fatalf("ban = %q", got)
	}
	if got := server.execute(".unban account Bob"); got != "Bob unbanned." {
		t.Fatalf("unban = %q", got)
	}
	if got := server.execute(".unban account Bob"); got != "There was an error removing the ban on Bob." {
		t.Fatalf("second unban = %q", got)
	}
	if got := server.execute(".ban account Ghost 1d x"); got != "account Ghost not found" {
		t.Fatalf("missing ban = %q", got)
	}
}

func TestIncorrectSyntaxAndUnknownCommand(t *testing.T) {
	server := testServer(t)
	if got := server.execute(".account create OnlyName"); got != msgIncorrectSyntax {
		t.Fatalf("syntax = %q", got)
	}
	if got := server.execute(".unknown thing"); got != "Command '.unknown' does not exist" {
		t.Fatalf("unknown = %q", got)
	}
}

func TestServerInfo(t *testing.T) {
	server := testServer(t)
	output := server.execute(".server info")
	for _, want := range []string{"Connected players:", "Server uptime:", "Update time diff:"} {
		if !strings.Contains(output, want) {
			t.Fatalf("server info missing %q:\n%s", want, output)
		}
	}
}

func soapRequest(command string) string {
	return `<?xml version="1.0"?><soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">` +
		`<soap:Body><executeCommand xmlns="urn:AC"><commandString>` + command + `</commandString></executeCommand></soap:Body></soap:Envelope>`
}

func TestSOAPRequiresAuth(t *testing.T) {
	server := New(Config{Username: "acgw", Password: "acgw", Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()

	resp, err := http.Post(httpServer.URL, "text/xml", strings.NewReader(soapRequest(".server info")))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

// TestAdapterCompatibility proves the fake speaks the same wire protocol as the
// real gateway SOAP adapter.
func TestAdapterCompatibility(t *testing.T) {
	server := New(Config{Username: "acgw", Password: "acgw", Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()

	client, err := azerothsoap.New(azerothsoap.Config{
		URL:      httpServer.URL,
		Username: "acgw",
		Password: "acgw",
	})
	if err != nil {
		t.Fatalf("client: %v", err)
	}

	result, err := client.Execute(context.Background(), ".account create Carol pass carol@example.com")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if result != "Account created: Carol" {
		t.Fatalf("result = %q", result)
	}

	resp := basicRequest(t, http.MethodGet, httpServer.URL+"/state", nil)
	defer func() { _ = resp.Body.Close() }()
	var seed Seed
	if err := json.NewDecoder(resp.Body).Decode(&seed); err != nil {
		t.Fatalf("decode state: %v", err)
	}
	if len(seed.Accounts) != 1 || seed.Accounts[0].Username != "CAROL" {
		t.Fatalf("state = %+v", seed)
	}
}
