package azerothcharacter

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/delivery"
	"github.com/mconcepcionb/ac-community-gw/internal/core/notice"
	"github.com/mconcepcionb/ac-community-gw/internal/core/permissions"
	"github.com/mconcepcionb/ac-community-gw/internal/core/plugins"
	"github.com/mconcepcionb/ac-community-gw/internal/core/services"
	"github.com/mconcepcionb/ac-community-gw/internal/testsupport"
)

type fakeVisibility struct {
	public      map[string]bool
	publicNames []string
	listErr     error
	publicErr   error
	setErr      error
}

func (f *fakeVisibility) SetVisibility(_ context.Context, name string, _ uuid.UUID, public bool) (bool, error) {
	if f.setErr != nil {
		return false, f.setErr
	}
	if f.public == nil {
		f.public = map[string]bool{}
	}
	f.public[name] = public
	return public, nil
}

func (f *fakeVisibility) GetVisibility(context.Context, string) (bool, error) { return false, nil }

func (f *fakeVisibility) ListByUser(context.Context, uuid.UUID) (map[string]bool, error) {
	return f.public, f.listErr
}

func (f *fakeVisibility) PublicNames(context.Context) ([]string, error) {
	return f.publicNames, f.publicErr
}

func TestHandleLeaderboard(t *testing.T) {
	characters := &fakeCharacters{characters: []azerothdb.Character{
		{Name: "Thrall", Level: 60, Class: 7, Race: 2, GuildName: "Horde"},
		{Name: "Ghost", Level: 60},
	}}
	plugin := New(Config{Characters: characters, Visibility: &fakeVisibility{publicNames: []string{"Thrall"}}})

	req := testsupport.SetPathValue(testsupport.NewRequest(http.MethodGet, "/x", ""), "board", azerothdb.BoardProgression)
	rec := testsupport.NewRecorder()
	plugin.handleLeaderboard(rec, req)
	testsupport.AssertStatus(t, rec, http.StatusOK)
	body := rec.Body.String()
	if !strings.Contains(body, "Thrall") || strings.Contains(body, "Ghost") {
		t.Fatalf("body = %s", body)
	}
}

func TestHandleLeaderboardErrors(t *testing.T) {
	req := func(board string) *http.Request {
		return testsupport.SetPathValue(testsupport.NewRequest(http.MethodGet, "/x", ""), "board", board)
	}

	t.Run("unknown board", func(t *testing.T) {
		plugin := New(Config{Characters: &fakeCharacters{}, Visibility: &fakeVisibility{}})
		rec := testsupport.NewRecorder()
		plugin.handleLeaderboard(rec, req("bogus"))
		testsupport.AssertStatus(t, rec, http.StatusNotFound)
	})

	t.Run("character db not configured", func(t *testing.T) {
		plugin := New(Config{Visibility: &fakeVisibility{}})
		rec := testsupport.NewRecorder()
		plugin.handleLeaderboard(rec, req(azerothdb.BoardWealth))
		testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)
	})

	t.Run("visibility not configured", func(t *testing.T) {
		plugin := New(Config{Characters: &fakeCharacters{}})
		rec := testsupport.NewRecorder()
		plugin.handleLeaderboard(rec, req(azerothdb.BoardWealth))
		testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)
	})

	t.Run("visibility error", func(t *testing.T) {
		plugin := New(Config{Characters: &fakeCharacters{}, Visibility: &fakeVisibility{publicErr: errors.New("db")}})
		rec := testsupport.NewRecorder()
		plugin.handleLeaderboard(rec, req(azerothdb.BoardWealth))
		testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)
	})

	t.Run("character query error", func(t *testing.T) {
		characters := &fakeCharacters{topErr: errors.New("db")}
		plugin := New(Config{Characters: characters, Visibility: &fakeVisibility{}})
		rec := testsupport.NewRecorder()
		plugin.handleLeaderboard(rec, req(azerothdb.BoardPvP))
		testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)
	})
}

func TestHandlePublicLeaderboardCaches(t *testing.T) {
	characters := &fakeCharacters{characters: []azerothdb.Character{{Name: "Thrall", Level: 60}}}
	plugin := New(Config{Characters: characters, Visibility: &fakeVisibility{publicNames: []string{"Thrall"}}})

	req := testsupport.SetPathValue(testsupport.NewRequest(http.MethodGet, "/x?limit=10&offset=0", ""), "board", azerothdb.BoardProgression)
	for i := 0; i < 2; i++ {
		rec := testsupport.NewRecorder()
		plugin.handlePublicLeaderboard(rec, req)
		testsupport.AssertStatus(t, rec, http.StatusOK)
	}
	if characters.topCalls != 1 {
		t.Fatalf("expected the second response to be cached, top calls = %d", characters.topCalls)
	}

	bad := testsupport.SetPathValue(testsupport.NewRequest(http.MethodGet, "/x", ""), "board", "nope")
	rec := testsupport.NewRecorder()
	plugin.handlePublicLeaderboard(rec, bad)
	testsupport.AssertStatus(t, rec, http.StatusNotFound)
}

func TestBoardWindow(t *testing.T) {
	cases := []struct {
		url        string
		wantLimit  int
		wantOffset int
	}{
		{"/x", 25, 0},
		{"/x?limit=0&offset=-5", 25, 0},
		{"/x?limit=500&offset=10", 100, 10},
		{"/x?limit=abc&offset=xyz", 25, 0},
	}
	for _, tc := range cases {
		limit, offset := boardWindow(testsupport.NewRequest(http.MethodGet, tc.url, ""))
		if limit != tc.wantLimit || offset != tc.wantOffset {
			t.Errorf("boardWindow(%s) = %d,%d, want %d,%d", tc.url, limit, offset, tc.wantLimit, tc.wantOffset)
		}
	}
}

func TestHandleListVisibility(t *testing.T) {
	plugin := New(Config{})
	rec := testsupport.NewRecorder()
	plugin.handleListVisibility(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodGet, "/x", ""), uuid.New()))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)

	plugin = New(Config{Visibility: &fakeVisibility{public: map[string]bool{"Thrall": true}}})
	rec = testsupport.NewRecorder()
	plugin.handleListVisibility(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodGet, "/x", ""), uuid.New()))
	testsupport.AssertStatus(t, rec, http.StatusOK)
	if !strings.Contains(rec.Body.String(), "Thrall") {
		t.Fatalf("body = %s", rec.Body.String())
	}

	plugin = New(Config{Visibility: &fakeVisibility{listErr: errors.New("db")}})
	rec = testsupport.NewRecorder()
	plugin.handleListVisibility(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodGet, "/x", ""), uuid.New()))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)
}

func TestHandleSetVisibility(t *testing.T) {
	accountID := int64(5)
	userID := uuid.New()
	body := `{"public":true}`

	t.Run("success", func(t *testing.T) {
		characters := &fakeCharacters{characters: []azerothdb.Character{{Name: "Thrall", AccountID: 5}}}
		visibility := &fakeVisibility{}
		plugin := New(Config{Characters: characters, Visibility: visibility})
		plugin.directory = &fakeDirectory{username: "ADMIN", accountID: &accountID}

		req := testsupport.SetPathValue(testsupport.NewRequest(http.MethodPut, "/x", body), "name", "Thrall")
		rec := testsupport.NewRecorder()
		plugin.handleSetVisibility(rec, testsupport.WithPrincipal(req, userID))
		testsupport.AssertStatus(t, rec, http.StatusOK)
		if !visibility.public["Thrall"] {
			t.Fatal("visibility not persisted")
		}
	})

	cases := []struct {
		name  string
		setup func(p *Plugin, r *http.Request) *http.Request
		want  int
	}{
		{"no store", func(p *Plugin, r *http.Request) *http.Request { p.visibility = nil; return r }, http.StatusServiceUnavailable},
		{"invalid name", func(p *Plugin, r *http.Request) *http.Request { return testsupport.SetPathValue(r, "name", "1") }, http.StatusUnprocessableEntity},
		{"no character db", func(p *Plugin, r *http.Request) *http.Request { p.characters = nil; return r }, http.StatusServiceUnavailable},
		{"no directory", func(p *Plugin, r *http.Request) *http.Request { p.directory = nil; return r }, http.StatusServiceUnavailable},
		{"not linked", func(p *Plugin, r *http.Request) *http.Request { p.directory = &fakeDirectory{}; return r }, http.StatusNotFound},
		{"character not found", func(p *Plugin, r *http.Request) *http.Request { p.characters = &fakeCharacters{}; return r }, http.StatusNotFound},
		{"character db error", func(p *Plugin, r *http.Request) *http.Request {
			p.characters = &fakeCharacters{findErr: errors.New("db")}
			return r
		}, http.StatusServiceUnavailable},
		{"not owner", func(p *Plugin, r *http.Request) *http.Request {
			p.characters = &fakeCharacters{characters: []azerothdb.Character{{Name: "Thrall", AccountID: 9}}}
			return r
		}, http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plugin := New(Config{Characters: &fakeCharacters{}, Visibility: &fakeVisibility{}})
			plugin.directory = &fakeDirectory{username: "ADMIN", accountID: &accountID}
			req := testsupport.SetPathValue(testsupport.NewRequest(http.MethodPut, "/x", body), "name", "Thrall")
			req = tc.setup(plugin, req)
			rec := testsupport.NewRecorder()
			plugin.handleSetVisibility(rec, testsupport.WithPrincipal(req, userID))
			testsupport.AssertStatus(t, rec, tc.want)
		})
	}
}

func TestHandleMyCharacters(t *testing.T) {
	accountID := int64(5)
	userID := uuid.New()

	plugin := New(Config{})
	rec := testsupport.NewRecorder()
	plugin.handleMyCharacters(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodGet, "/x", ""), userID))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)

	plugin = New(Config{Characters: &fakeCharacters{}})
	rec = testsupport.NewRecorder()
	plugin.handleMyCharacters(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodGet, "/x", ""), userID))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)

	plugin = New(Config{Characters: &fakeCharacters{}})
	plugin.directory = &fakeDirectory{username: "ADMIN"}
	rec = testsupport.NewRecorder()
	plugin.handleMyCharacters(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodGet, "/x", ""), userID))
	testsupport.AssertStatus(t, rec, http.StatusNotFound)

	characters := &fakeCharacters{characters: []azerothdb.Character{{Name: "Thrall", AccountID: 5}}}
	plugin = New(Config{Characters: characters})
	plugin.directory = &fakeDirectory{username: "ADMIN", accountID: &accountID}
	rec = testsupport.NewRecorder()
	plugin.handleMyCharacters(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodGet, "/x", ""), userID))
	testsupport.AssertStatus(t, rec, http.StatusOK)
	if characters.last.AccountID != accountID {
		t.Fatalf("account query = %+v", characters.last)
	}
}

func TestDirectoryCapability(t *testing.T) {
	accountID := int64(5)
	plugin := New(Config{})
	if _, err := plugin.OnlineCharacters(context.Background(), 10, 0); err == nil {
		t.Fatal("OnlineCharacters without a character DB must fail")
	}
	if _, err := plugin.CharactersByUser(context.Background(), uuid.NewString()); err == nil {
		t.Fatal("CharactersByUser without a character DB must fail")
	}

	characters := &fakeCharacters{characters: []azerothdb.Character{{Name: "Thrall"}}}
	plugin = New(Config{Characters: characters})
	if _, err := plugin.CharactersByUser(context.Background(), uuid.NewString()); err == nil {
		t.Fatal("CharactersByUser without a directory must fail")
	}

	plugin.directory = &fakeDirectory{username: "ADMIN"} // no account id
	if _, err := plugin.CharactersByUser(context.Background(), uuid.NewString()); err == nil {
		t.Fatal("CharactersByUser without a link must fail")
	}

	plugin.directory = &fakeDirectory{username: "ADMIN", accountID: &accountID}
	online, err := plugin.OnlineCharacters(context.Background(), 10, 0)
	if err != nil || len(online) != 1 || !characters.last.OnlineOnly {
		t.Fatalf("OnlineCharacters = %+v,%v", online, err)
	}
	byUser, err := plugin.CharactersByUser(context.Background(), uuid.NewString())
	if err != nil || len(byUser) != 1 || characters.last.AccountID != accountID {
		t.Fatalf("CharactersByUser = %+v,%v", byUser, err)
	}
}

func TestNoticeSend(t *testing.T) {
	if _, err := New(Config{}).Send(context.Background(), notice.Request{Character: "1"}); !errors.Is(err, notice.ErrInvalidRequest) {
		t.Fatalf("invalid recipient = %v", err)
	}
	if _, err := New(Config{NoticeItemID: 0}).Send(context.Background(), notice.Request{Character: "Thrall"}); !errors.Is(err, notice.ErrNotConfigured) {
		t.Fatalf("unconfigured = %v", err)
	}
	if _, err := New(Config{NoticeItemID: 1}).Send(context.Background(), notice.Request{Character: "Thrall"}); !errors.Is(err, azerothdb.ErrUnavailable) {
		t.Fatalf("no character db = %v", err)
	}

	characters := &fakeCharacters{characters: []azerothdb.Character{{Name: "Thrall", AccountID: 5}}}
	executor := &fakeExecutor{result: "Mail sent."}
	plugin := New(Config{NoticeItemID: 1, Characters: characters, Executor: executor})
	if _, err := plugin.Send(context.Background(), notice.Request{Character: "Ghost", AccountID: 5}); !errors.Is(err, notice.ErrCharacterNotFound) {
		t.Fatalf("missing character = %v", err)
	}
	if _, err := plugin.Send(context.Background(), notice.Request{Character: "Thrall", AccountID: 9}); !errors.Is(err, notice.ErrNotOwner) {
		t.Fatalf("not owner = %v", err)
	}
	if _, err := plugin.Send(context.Background(), notice.Request{Character: "Thrall", AccountID: 5, Subject: "Hi", Body: "Body"}); err != nil {
		t.Fatalf("Send = %v", err)
	}
	if !strings.Contains(executor.last, "Thrall") {
		t.Fatalf("command = %q", executor.last)
	}
}

func TestClassAndRaceNames(t *testing.T) {
	if className(1) != "Warrior" || className(6) != "Death Knight" || className(99) != "Unknown" {
		t.Fatal("className mismatch")
	}
	if raceName(2) != "Orc" || raceName(10) != "Blood Elf" || raceName(99) != "Unknown" {
		t.Fatal("raceName mismatch")
	}
}

func TestPermissionDefsAreNamespaced(t *testing.T) {
	for _, def := range permissionDefs() {
		if def.Owner != Name || def.Namespace != "azeroth" || !strings.HasPrefix(string(def.Name), "azeroth.") {
			t.Fatalf("definition not namespaced: %+v", def)
		}
	}
}

func newCharacterRegistry() *plugins.Registry {
	return &plugins.Registry{
		Mux:         http.NewServeMux(),
		Services:    services.NewRegistry(),
		Permissions: permissions.NewRegistry(),
		Audit:       audit.NopRecorder{},
		RequireAuth: func(next http.Handler) http.Handler { return next },
		RequirePermission: func(_ permissions.Permission, next http.Handler) http.Handler {
			return next
		},
		RateLimit: func(next http.Handler) http.Handler { return next },
	}
}

func TestRegisterRequiresAccountDirectory(t *testing.T) {
	if err := New(Config{}).Register(context.Background(), newCharacterRegistry()); err == nil {
		t.Fatal("Register must fail without the account directory capability")
	}
}

func TestRegisterWiresCharacter(t *testing.T) {
	reg := newCharacterRegistry()
	if err := services.Provide[accountDirectory](reg.Services, accountDirectoryService, &fakeDirectory{}); err != nil {
		t.Fatalf("provide directory: %v", err)
	}
	plugin := New(Config{Characters: &fakeCharacters{}, Visibility: &fakeVisibility{}})
	if err := plugin.Register(context.Background(), reg); err != nil {
		t.Fatalf("Register = %v", err)
	}
	if !reg.Permissions.Has(PermissionLeaderboardRead) {
		t.Fatal("leaderboard permission not registered")
	}
	for _, name := range []string{DirectoryService, NoticeService, DeliveryService} {
		if !reg.Services.Has(name) {
			t.Fatalf("capability %s not published", name)
		}
	}

	rec := testsupport.NewRecorder()
	reg.Mux.ServeHTTP(rec, testsupport.SetPathValue(
		testsupport.NewRequest(http.MethodGet, "/api/v1/azeroth/leaderboards/progression", ""), "board", "progression"))
	testsupport.AssertStatus(t, rec, http.StatusOK)
}

var _ delivery.Service = (*Plugin)(nil)
