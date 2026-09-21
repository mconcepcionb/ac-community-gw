package identitydiscord

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
	"github.com/mconcepcionb/ac-community-gw/internal/core/metrics"
	"github.com/mconcepcionb/ac-community-gw/internal/core/permissions"
	"github.com/mconcepcionb/ac-community-gw/internal/core/userdir"
	identitydiscordrepo "github.com/mconcepcionb/ac-community-gw/internal/plugins/identitydiscord/repository/generated"
)

var errFake = errors.New("fake upstream failure")

type fakeProvider struct {
	state       string
	challenge   string
	exchangeErr error
	userErr     error
	roleIDs     []string
	roleErr     error
}

func (f *fakeProvider) AuthorizeURL(state, codeChallenge string) string {
	f.state = state
	f.challenge = codeChallenge
	return "https://discord.test/authorize?state=" + url.QueryEscape(state)
}

func (f *fakeProvider) Exchange(context.Context, string, string) (auth.DiscordToken, error) {
	if f.exchangeErr != nil {
		return auth.DiscordToken{}, f.exchangeErr
	}
	return auth.DiscordToken{AccessToken: "token"}, nil
}

func (f *fakeProvider) CurrentUser(context.Context, string) (auth.DiscordUser, error) {
	if f.userErr != nil {
		return auth.DiscordUser{}, f.userErr
	}
	return auth.DiscordUser{ID: "42", Username: "user", GlobalName: "User"}, nil
}

func (f *fakeProvider) GuildMemberRoleIDs(context.Context, string, string) ([]string, error) {
	if f.roleErr != nil {
		return nil, f.roleErr
	}
	return f.roleIDs, nil
}

type fakeRepo struct {
	userID    uuid.UUID
	created   bool
	mapped    []string
	roles     []string
	byDiscord map[string]uuid.UUID
	users     []userdir.User
	profile   userdir.User

	rolesList     []string
	roleGrants    []identitydiscordrepo.RolePermission
	mappings      []identitydiscordrepo.DiscordRoleMapping
	listRolesErr  error
	replaceErr    error
	upsertErr     error
	upsertRoleErr error
	deleteErr     error

	lastReplaceRole  string
	lastReplacePerms []string
	lastMappingID    string
	lastMappingRole  string
}

func (f *fakeRepo) ListUsers(_ context.Context, _ string, _, _ int) ([]userdir.User, error) {
	return f.users, nil
}

func (f *fakeRepo) ResolveUserByName(_ context.Context, _ string) ([]userdir.User, error) {
	return f.users, nil
}

func (f *fakeRepo) UserProfile(_ context.Context, _ uuid.UUID) (userdir.User, error) {
	return f.profile, nil
}

func (f *fakeRepo) UserIDByDiscordID(_ context.Context, discordID string) (uuid.UUID, bool, error) {
	id, ok := f.byDiscord[discordID]
	if !ok {
		return uuid.Nil, false, nil
	}
	return id, true, nil
}

func (f *fakeRepo) Provision(context.Context, auth.DiscordUser) (uuid.UUID, bool, error) {
	if f.userID == uuid.Nil {
		f.userID = uuid.New()
	}
	return f.userID, f.created, nil
}

func (f *fakeRepo) MappedRoles(context.Context, []string) ([]string, error) {
	return f.mapped, nil
}

func (f *fakeRepo) UserRoles(context.Context, uuid.UUID) ([]string, error) {
	return f.roles, nil
}

func (f *fakeRepo) UpdateUserRoles(_ context.Context, _ uuid.UUID, roles []string) error {
	f.roles = roles
	return nil
}

func (f *fakeRepo) SyncPermissions(context.Context, []permissions.Definition) error { return nil }

func (f *fakeRepo) ListRolePermissions(context.Context) ([]identitydiscordrepo.RolePermission, error) {
	return f.roleGrants, nil
}

func (f *fakeRepo) ListRoles(context.Context) ([]string, error) {
	return f.rolesList, f.listRolesErr
}

func (f *fakeRepo) UpsertRole(context.Context, string) error { return f.upsertRoleErr }

func (f *fakeRepo) GrantRolePermission(context.Context, string, string) error { return nil }

func (f *fakeRepo) RevokeRolePermission(context.Context, string, string) error { return nil }

func (f *fakeRepo) ReplaceRolePermissions(_ context.Context, role string, perms []string) error {
	f.lastReplaceRole, f.lastReplacePerms = role, perms
	return f.replaceErr
}

func (f *fakeRepo) ListDiscordRoleMappings(context.Context) ([]identitydiscordrepo.DiscordRoleMapping, error) {
	return f.mappings, nil
}

func (f *fakeRepo) UpsertDiscordRoleMapping(_ context.Context, discordRoleID, role string) error {
	f.lastMappingID, f.lastMappingRole = discordRoleID, role
	return f.upsertErr
}

func (f *fakeRepo) DeleteDiscordRoleMapping(context.Context, string) error { return f.deleteErr }

func newTestPlugin(t *testing.T, provider auth.DiscordProvider, repo Repository) (*Plugin, *auth.Manager, *MemoryStateStore) {
	t.Helper()
	sessions := auth.NewManager(auth.ManagerOptions{
		Store:      auth.NewMemoryStore(),
		TTL:        time.Hour,
		Secure:     false,
		CookieName: "acgw_session",
	})
	states := NewMemoryStateStore()
	plugin := New(Config{
		Sessions:   sessions,
		Provider:   provider,
		States:     states,
		Repository: repo,
		Authorizer: permissions.NewAuthorizer(),
		GuildID:    "guild-1",
	})
	return plugin, sessions, states
}

func login(t *testing.T, plugin *Plugin, target string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	plugin.handleLogin(rec, httptest.NewRequest(http.MethodGet, target, nil))
	return rec
}

func TestHandleLoginRedirectsWithStateAndPKCE(t *testing.T) {
	provider := &fakeProvider{}
	plugin, _, _ := newTestPlugin(t, provider, &fakeRepo{})

	rec := login(t, plugin, "/api/v1/auth/discord/login")
	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Header().Get("Location"), "state=") {
		t.Fatalf("location = %q", rec.Header().Get("Location"))
	}
	if provider.state == "" || provider.challenge == "" {
		t.Fatalf("missing state/challenge: %q / %q", provider.state, provider.challenge)
	}
}

func TestHandleCallbackSuccess(t *testing.T) {
	provider := &fakeProvider{roleIDs: []string{"role-1"}}
	repo := &fakeRepo{mapped: []string{"member"}}
	plugin, sessions, _ := newTestPlugin(t, provider, repo)

	login(t, plugin, "/api/v1/auth/discord/login")

	rec := httptest.NewRecorder()
	plugin.handleCallback(rec, httptest.NewRequest(http.MethodGet, "/api/v1/auth/discord/callback?code=abc&state="+provider.state, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	var cookie *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == "acgw_session" {
			cookie = c
		}
	}
	if cookie == nil || !cookie.HttpOnly {
		t.Fatalf("session cookie missing or not HttpOnly: %+v", cookie)
	}

	principal, err := sessions.Resolve(context.Background(), cookie.Value)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if principal.DiscordID != "42" || len(principal.Roles) != 1 || principal.Roles[0] != "member" {
		t.Fatalf("principal = %+v", principal)
	}

	meRec := httptest.NewRecorder()
	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil).WithContext(auth.WithPrincipal(context.Background(), principal))
	plugin.handleMe(meRec, meReq)
	if meRec.Code != http.StatusOK {
		t.Fatalf("me status = %d", meRec.Code)
	}
	var payload map[string]any
	if err := json.Unmarshal(meRec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("me decode: %v", err)
	}
	if payload["discord_id"] != "42" {
		t.Fatalf("me payload = %v", payload)
	}
}

func TestHandleCallbackRejectsUnknownState(t *testing.T) {
	plugin, _, _ := newTestPlugin(t, &fakeProvider{}, &fakeRepo{})
	rec := httptest.NewRecorder()
	plugin.handleCallback(rec, httptest.NewRequest(http.MethodGet, "/api/v1/auth/discord/callback?code=abc&state=nope", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestHandleCallbackRejectsMissingState(t *testing.T) {
	plugin, _, _ := newTestPlugin(t, &fakeProvider{}, &fakeRepo{})
	rec := httptest.NewRecorder()
	plugin.handleCallback(rec, httptest.NewRequest(http.MethodGet, "/api/v1/auth/discord/callback?code=abc", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestHandleCallbackRejectsReplayedState(t *testing.T) {
	provider := &fakeProvider{}
	plugin, _, _ := newTestPlugin(t, provider, &fakeRepo{})
	login(t, plugin, "/api/v1/auth/discord/login")
	state := provider.state

	plugin.handleCallback(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/cb?code=abc&state="+state, nil))
	rec := httptest.NewRecorder()
	plugin.handleCallback(rec, httptest.NewRequest(http.MethodGet, "/cb?code=abc&state="+state, nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestHandleCallbackMapsUpstreamFailures(t *testing.T) {
	cases := map[string]*fakeProvider{
		"exchange": {exchangeErr: errFake},
		"user":     {userErr: errFake},
		"roles":    {roleIDs: []string{"r"}, roleErr: errFake},
	}
	for name, provider := range cases {
		t.Run(name, func(t *testing.T) {
			plugin, _, _ := newTestPlugin(t, provider, &fakeRepo{})
			login(t, plugin, "/api/v1/auth/discord/login")
			rec := httptest.NewRecorder()
			plugin.handleCallback(rec, httptest.NewRequest(http.MethodGet, "/cb?code=abc&state="+provider.state, nil))
			if rec.Code != http.StatusBadGateway {
				t.Fatalf("status = %d", rec.Code)
			}
		})
	}
}

func TestHandleCallbackRedirectsToReturnTo(t *testing.T) {
	provider := &fakeProvider{}
	plugin, _, _ := newTestPlugin(t, provider, &fakeRepo{})
	login(t, plugin, "/api/v1/auth/discord/login?return_to=/dashboard")

	rec := httptest.NewRecorder()
	plugin.handleCallback(rec, httptest.NewRequest(http.MethodGet, "/cb?code=abc&state="+provider.state, nil))
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/dashboard" {
		t.Fatalf("status = %d location = %q", rec.Code, rec.Header().Get("Location"))
	}
}

func TestHandleCallbackIgnoresUnsafeReturnTo(t *testing.T) {
	provider := &fakeProvider{}
	plugin, _, _ := newTestPlugin(t, provider, &fakeRepo{})
	login(t, plugin, "/api/v1/auth/discord/login?return_to=https://evil.test")

	rec := httptest.NewRecorder()
	plugin.handleCallback(rec, httptest.NewRequest(http.MethodGet, "/cb?code=abc&state="+provider.state, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	roles, ok := payload["roles"].([]any)
	if !ok || len(roles) != 0 {
		t.Fatalf("roles = %#v, want []", payload["roles"])
	}
}

func TestSanitizeReturnTo(t *testing.T) {
	valid := map[string]string{
		"/dashboard":        "/dashboard",
		"/store/wallet?x=1": "/store/wallet?x=1",
	}
	for input, want := range valid {
		if got := sanitizeReturnTo(input); got != want {
			t.Fatalf("sanitizeReturnTo(%q) = %q, want %q", input, got, want)
		}
	}

	invalid := []string{
		"https://evil.test",
		"http://x/y",
		"//evil.test",
		`/\evil.test`,
		`/%5Cevil.test`,
		`\evil.test`,
		"javascript:alert(1)",
		"/a\r\nb",
	}
	for _, input := range invalid {
		if got := sanitizeReturnTo(input); got != "" {
			t.Fatalf("sanitizeReturnTo(%q) = %q, want empty", input, got)
		}
	}
}

func TestHandleMeIncludesRolesAndPermissions(t *testing.T) {
	sessions := auth.NewManager(auth.ManagerOptions{Store: auth.NewMemoryStore(), TTL: time.Hour})
	authorizer := permissions.NewAuthorizer()
	authorizer.Grant("member", "store.products.write", "azeroth.items.read", "azeroth.items.read")
	plugin := New(Config{
		Sessions:   sessions,
		Provider:   &fakeProvider{},
		States:     NewMemoryStateStore(),
		Repository: &fakeRepo{},
		Authorizer: authorizer,
	})

	principal := auth.Principal{UserID: uuid.New(), DiscordID: "42", Roles: []string{"member"}}
	rec := httptest.NewRecorder()
	plugin.handleMe(rec, httptest.NewRequest(http.MethodGet, "/api/v1/me", nil).WithContext(auth.WithPrincipal(context.Background(), principal)))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var payload MeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload.UserID != principal.UserID.String() || payload.DiscordID != "42" {
		t.Fatalf("payload = %+v", payload)
	}
	if len(payload.Roles) != 1 || payload.Roles[0] != "member" {
		t.Fatalf("roles = %v", payload.Roles)
	}
	want := []string{"azeroth.items.read", "store.products.write"}
	if len(payload.Permissions) != len(want) {
		t.Fatalf("permissions = %v, want %v", payload.Permissions, want)
	}
	for i := range want {
		if payload.Permissions[i] != want[i] {
			t.Fatalf("permissions = %v, want %v", payload.Permissions, want)
		}
	}
}

func TestHandleMeIncludesDiscordProfile(t *testing.T) {
	userID := uuid.New()
	created := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	plugin := New(Config{
		Sessions: auth.NewManager(auth.ManagerOptions{Store: auth.NewMemoryStore(), TTL: time.Hour}),
		Repository: &fakeRepo{profile: userdir.User{
			ID:          userID,
			DiscordID:   "42",
			Username:    "alice",
			GlobalName:  "Alice",
			DisplayName: "Alice",
			Avatar:      "abc123",
			CreatedAt:   created,
		}},
	})
	principal := auth.Principal{UserID: userID, DiscordID: "42"}
	rec := httptest.NewRecorder()
	plugin.handleMe(rec, httptest.NewRequest(http.MethodGet, "/api/v1/me", nil).WithContext(auth.WithPrincipal(context.Background(), principal)))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var payload MeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload.Username != "alice" || payload.GlobalName != "Alice" ||
		payload.DisplayName != "Alice" || payload.Avatar != "abc123" {
		t.Fatalf("profile = %+v", payload)
	}
	if payload.CreatedAt != created.Format(time.RFC3339) {
		t.Fatalf("created_at = %q", payload.CreatedAt)
	}
}

func TestHandleMeWithoutAuthorizerReturnsEmptyArrays(t *testing.T) {
	plugin := New(Config{
		Sessions:   auth.NewManager(auth.ManagerOptions{Store: auth.NewMemoryStore(), TTL: time.Hour}),
		Repository: &fakeRepo{},
	})
	principal := auth.Principal{UserID: uuid.New(), DiscordID: "42"}
	rec := httptest.NewRecorder()
	plugin.handleMe(rec, httptest.NewRequest(http.MethodGet, "/api/v1/me", nil).WithContext(auth.WithPrincipal(context.Background(), principal)))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var payload MeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload.Roles == nil || payload.Permissions == nil {
		t.Fatalf("roles/permissions must be arrays, got roles=%v permissions=%v", payload.Roles, payload.Permissions)
	}
	if len(payload.Permissions) != 0 {
		t.Fatalf("permissions = %v, want empty", payload.Permissions)
	}
}

func TestPluginMetricsCounters(t *testing.T) {
	registry := metrics.NewRegistry()
	provider := &fakeProvider{roleIDs: []string{"role-1"}}
	repo := &fakeRepo{mapped: []string{"member"}}
	plugin := New(Config{
		Sessions:   auth.NewManager(auth.ManagerOptions{Store: auth.NewMemoryStore(), TTL: time.Hour}),
		Provider:   provider,
		States:     NewMemoryStateStore(),
		Repository: repo,
		GuildID:    "guild-1",
		Metrics:    registry,
	})

	login(t, plugin, "/api/v1/auth/discord/login")

	success := httptest.NewRecorder()
	plugin.handleCallback(success, httptest.NewRequest(http.MethodGet, "/cb?code=abc&state="+provider.state, nil))
	if success.Code != http.StatusOK {
		t.Fatalf("success status = %d", success.Code)
	}

	failure := httptest.NewRecorder()
	plugin.handleCallback(failure, httptest.NewRequest(http.MethodGet, "/cb?code=abc&state=bogus", nil))
	if failure.Code != http.StatusBadRequest {
		t.Fatalf("failure status = %d", failure.Code)
	}

	snapshot := registry.Snapshot()
	if snapshot[metricLoginStarted] != 1 {
		t.Fatalf("started = %d", snapshot[metricLoginStarted])
	}
	if snapshot[metricLoginCompleted] != 1 || snapshot[metricSessionCreated] != 1 {
		t.Fatalf("completed = %d, session = %d", snapshot[metricLoginCompleted], snapshot[metricSessionCreated])
	}
	if snapshot[metricLoginFailed] != 1 {
		t.Fatalf("failed = %d", snapshot[metricLoginFailed])
	}
	if snapshot[metricRolesChanged] != 1 {
		t.Fatalf("roles changed = %d", snapshot[metricRolesChanged])
	}
}
