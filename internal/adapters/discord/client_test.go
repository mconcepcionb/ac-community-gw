package discord

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func testClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	client, err := New(Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "https://gateway.test/callback",
		AuthorizeURL: "https://discord.test/authorize",
		TokenURL:     server.URL + "/token",
		APIBaseURL:   server.URL,
		Scopes:       []string{"identify", "guilds.members.read"},
		HTTPClient:   server.Client(),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return client
}

func TestAuthorizeURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer server.Close()

	got, err := url.Parse(testClient(t, server).AuthorizeURL("state-1", "challenge-1"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	query := got.Query()
	if query.Get("client_id") != "client-id" {
		t.Fatalf("client_id = %q", query.Get("client_id"))
	}
	if query.Get("response_type") != "code" {
		t.Fatalf("response_type = %q", query.Get("response_type"))
	}
	if query.Get("scope") != "identify guilds.members.read" {
		t.Fatalf("scope = %q", query.Get("scope"))
	}
	if query.Get("state") != "state-1" {
		t.Fatalf("state = %q", query.Get("state"))
	}
	if query.Get("code_challenge") != "challenge-1" || query.Get("code_challenge_method") != "S256" {
		t.Fatalf("pkce = %q / %q", query.Get("code_challenge"), query.Get("code_challenge_method"))
	}
}

func TestExchange(t *testing.T) {
	var form url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		form = r.PostForm
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"token-1","token_type":"Bearer","expires_in":3600}`))
	}))
	defer server.Close()

	token, err := testClient(t, server).Exchange(context.Background(), "code-1", "verifier-1")
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	if token.AccessToken != "token-1" {
		t.Fatalf("access token = %q", token.AccessToken)
	}
	if form.Get("grant_type") != "authorization_code" || form.Get("code") != "code-1" {
		t.Fatalf("form = %v", form)
	}
	if form.Get("code_verifier") != "verifier-1" || form.Get("client_secret") != "client-secret" {
		t.Fatalf("form = %v", form)
	}
}

func TestCurrentUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/@me" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer token-1" {
			t.Errorf("authorization = %q", r.Header.Get("Authorization"))
		}
		_, _ = w.Write([]byte(`{"id":"42","username":"user","global_name":"User","avatar":"hash"}`))
	}))
	defer server.Close()

	user, err := testClient(t, server).CurrentUser(context.Background(), "token-1")
	if err != nil {
		t.Fatalf("CurrentUser: %v", err)
	}
	if user.ID != "42" || user.Username != "user" || user.GlobalName != "User" || user.Avatar != "hash" {
		t.Fatalf("user = %+v", user)
	}
}

func TestGuildMemberRoleIDs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/@me/guilds/guild-1/member" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"roles":["role-a","role-b"]}`))
	}))
	defer server.Close()

	roles, err := testClient(t, server).GuildMemberRoleIDs(context.Background(), "token-1", "guild-1")
	if err != nil {
		t.Fatalf("GuildMemberRoleIDs: %v", err)
	}
	if len(roles) != 2 || roles[0] != "role-a" || roles[1] != "role-b" {
		t.Fatalf("roles = %v", roles)
	}
}

func TestUpstreamErrorMapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_client","error_description":"bad secret","access_token":"should-not-leak"}`))
	}))
	defer server.Close()

	_, err := testClient(t, server).Exchange(context.Background(), "code-1", "")
	if !errors.Is(err, ErrUpstream) {
		t.Fatalf("expected ErrUpstream, got %v", err)
	}
	if !strings.Contains(err.Error(), "invalid_client") {
		t.Fatalf("expected the OAuth error code to be surfaced: %v", err)
	}
	if strings.Contains(err.Error(), "should-not-leak") {
		t.Fatalf("error leaks a token value: %v", err)
	}
}
