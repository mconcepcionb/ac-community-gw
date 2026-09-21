package azerothaccount

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/commands"
	"github.com/mconcepcionb/ac-community-gw/internal/core/notice"
	"github.com/mconcepcionb/ac-community-gw/internal/core/permissions"
	"github.com/mconcepcionb/ac-community-gw/internal/core/plugins"
	"github.com/mconcepcionb/ac-community-gw/internal/core/services"
	"github.com/mconcepcionb/ac-community-gw/internal/core/userdir"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothaccount/domain"
	"github.com/mconcepcionb/ac-community-gw/internal/testsupport"
)

type fakeClaims struct {
	claim       domain.Claim
	hasClaim    bool
	list        []domain.Claim
	upsertErr   error
	listErr     error
	incremented int
	deleted     int
	upserted    domain.Claim
}

func (f *fakeClaims) UpsertClaim(_ context.Context, userID uuid.UUID, username, hash string, expires time.Time) (domain.Claim, error) {
	if f.upsertErr != nil {
		return domain.Claim{}, f.upsertErr
	}
	claim := domain.Claim{UserID: userID, AccountUsername: username, CodeHash: hash, ExpiresAt: expires}
	f.claim, f.hasClaim, f.upserted = claim, true, claim
	return claim, nil
}

func (f *fakeClaims) GetClaim(context.Context, uuid.UUID) (domain.Claim, error) {
	if !f.hasClaim {
		return domain.Claim{}, domain.ErrClaimNotFound
	}
	return f.claim, nil
}

func (f *fakeClaims) IncrementClaimAttempts(context.Context, uuid.UUID) (domain.Claim, error) {
	f.incremented++
	f.claim.Attempts++
	return f.claim, nil
}

func (f *fakeClaims) DeleteClaim(context.Context, uuid.UUID) error {
	f.deleted++
	f.hasClaim = false
	return nil
}

func (f *fakeClaims) ListClaims(context.Context, int, int) ([]domain.Claim, error) {
	return f.list, f.listErr
}

type fakeNotice struct {
	last notice.Request
	out  string
	err  error
}

func (f *fakeNotice) Send(_ context.Context, req notice.Request) (string, error) {
	f.last = req
	return f.out, f.err
}

func withNotice(plugin *Plugin, svc notice.Service) {
	registry := services.NewRegistry()
	_ = services.Provide[notice.Service](registry, noticeService, svc)
	plugin.registry = registry
}

func futureClaim(username, code string) fakeClaims {
	return fakeClaims{hasClaim: true, claim: domain.Claim{
		UserID: uuid.New(), AccountUsername: username,
		CodeHash: hashCode(code), ExpiresAt: time.Now().Add(time.Minute),
	}}
}

func TestHandleStartClaimSuccess(t *testing.T) {
	reader := &fakeReader{accounts: []azerothdb.Account{{ID: 5, Username: "ADMIN"}}}
	links := newFakeLinks()
	claims := &fakeClaims{}
	notif := &fakeNotice{}
	plugin := New(Config{Accounts: reader, Links: links, Claims: claims})
	withNotice(plugin, notif)

	rec := testsupport.NewRecorder()
	plugin.handleStartClaim(rec, testsupport.WithPrincipal(
		testsupport.NewRequest(http.MethodPost, "/x", `{"account_username":"ADMIN","character":"Thrall"}`), uuid.New()))

	testsupport.AssertStatus(t, rec, http.StatusOK)
	match := regexp.MustCompile(`code is (\d{6})`).FindStringSubmatch(notif.last.Body)
	if match == nil {
		t.Fatalf("notice body has no code: %q", notif.last.Body)
	}
	if claims.upserted.CodeHash != hashCode(match[1]) {
		t.Fatal("stored hash does not match the delivered code")
	}
	if notif.last.AccountID != 5 || notif.last.Character != "Thrall" {
		t.Fatalf("notice = %+v", notif.last)
	}
}

func TestHandleStartClaimErrors(t *testing.T) {
	userID := uuid.New()
	body := `{"account_username":"ADMIN","character":"Thrall"}`

	t.Run("not configured", func(t *testing.T) {
		plugin := New(Config{
			Accounts: &fakeReader{accounts: []azerothdb.Account{{ID: 5, Username: "ADMIN"}}},
			Links:    newFakeLinks(), Claims: &fakeClaims{},
		})
		rec := testsupport.NewRecorder()
		plugin.handleStartClaim(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodPost, "/x", body), userID))
		testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)
	})

	t.Run("already linked", func(t *testing.T) {
		links := newFakeLinks()
		links.links[userID] = domain.Link{UserID: userID, AccountUsername: "ADMIN"}
		plugin := New(Config{Accounts: &fakeReader{}, Links: links, Claims: &fakeClaims{}})
		rec := testsupport.NewRecorder()
		plugin.handleStartClaim(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodPost, "/x", body), userID))
		testsupport.AssertStatus(t, rec, http.StatusConflict)
	})

	t.Run("link lookup error", func(t *testing.T) {
		links := newFakeLinks()
		links.getErr = errors.New("db down")
		plugin := New(Config{Accounts: &fakeReader{}, Links: links, Claims: &fakeClaims{}})
		rec := testsupport.NewRecorder()
		plugin.handleStartClaim(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodPost, "/x", body), userID))
		testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)
	})

	t.Run("account not found", func(t *testing.T) {
		plugin := New(Config{Accounts: &fakeReader{}, Links: newFakeLinks(), Claims: &fakeClaims{}})
		rec := testsupport.NewRecorder()
		plugin.handleStartClaim(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodPost, "/x", body), userID))
		testsupport.AssertStatus(t, rec, http.StatusUnprocessableEntity)
	})

	t.Run("claim store unavailable", func(t *testing.T) {
		plugin := New(Config{Accounts: &fakeReader{accounts: []azerothdb.Account{{ID: 5, Username: "ADMIN"}}}, Links: newFakeLinks(), Claims: &fakeClaims{upsertErr: errors.New("db")}})
		rec := testsupport.NewRecorder()
		plugin.handleStartClaim(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodPost, "/x", body), userID))
		testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)
	})

	t.Run("no links or claims", func(t *testing.T) {
		plugin := New(Config{Accounts: &fakeReader{}})
		rec := testsupport.NewRecorder()
		plugin.handleStartClaim(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodPost, "/x", body), userID))
		testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)
	})

	for name, noticeErr := range map[string]struct {
		err  error
		want int
	}{
		"not owner":           {notice.ErrNotOwner, http.StatusForbidden},
		"character not found": {notice.ErrCharacterNotFound, http.StatusUnprocessableEntity},
		"invalid request":     {notice.ErrInvalidRequest, http.StatusUnprocessableEntity},
		"transport":           {errors.New("smtp"), http.StatusBadGateway},
	} {
		t.Run(name, func(t *testing.T) {
			plugin := New(Config{
				Accounts: &fakeReader{accounts: []azerothdb.Account{{ID: 5, Username: "ADMIN"}}},
				Links:    newFakeLinks(), Claims: &fakeClaims{},
			})
			withNotice(plugin, &fakeNotice{err: noticeErr.err})
			rec := testsupport.NewRecorder()
			plugin.handleStartClaim(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodPost, "/x", body), userID))
			testsupport.AssertStatus(t, rec, noticeErr.want)
		})
	}
}

func TestHandleVerifyClaim(t *testing.T) {
	userID := uuid.New()
	body := `{"account_username":"ADMIN","code":"123456"}`

	t.Run("success links the account", func(t *testing.T) {
		links := newFakeLinks()
		claims := futureClaim("ADMIN", "123456")
		claims.claim.UserID = userID
		reader := &fakeReader{accounts: []azerothdb.Account{{ID: 7, Username: "ADMIN"}}}
		plugin := New(Config{Accounts: reader, Links: links, Claims: &claims})

		rec := testsupport.NewRecorder()
		plugin.handleVerifyClaim(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodPost, "/x", body), userID))
		testsupport.AssertStatus(t, rec, http.StatusOK)

		link := links.links[userID]
		if link.AccountID == nil || *link.AccountID != 7 {
			t.Fatalf("link = %+v", link)
		}
		if claims.deleted == 0 {
			t.Fatal("the claim should be consumed")
		}
	})

	t.Run("success without account reader", func(t *testing.T) {
		links := newFakeLinks()
		claims := futureClaim("ADMIN", "123456")
		claims.claim.UserID = userID
		plugin := New(Config{Links: links, Claims: &claims})

		rec := testsupport.NewRecorder()
		plugin.handleVerifyClaim(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodPost, "/x", body), userID))
		testsupport.AssertStatus(t, rec, http.StatusOK)
		if links.links[userID].AccountID != nil {
			t.Fatal("account id should be nil without a reader")
		}
	})

	t.Run("basic guards", func(t *testing.T) {
		cases := []struct {
			name  string
			setup func(*Plugin)
			body  string
			want  int
		}{
			{"missing stores", func(p *Plugin) { p.links, p.claims = nil, nil }, body, http.StatusServiceUnavailable},
			{"already linked", func(p *Plugin) { p.links.(*fakeLinks).links[userID] = domain.Link{UserID: userID} }, body, http.StatusConflict},
			{"link lookup error", func(p *Plugin) { p.links.(*fakeLinks).getErr = errors.New("db") }, body, http.StatusServiceUnavailable},
			{"no pending", func(p *Plugin) {}, body, http.StatusUnprocessableEntity},
			{"mismatch", func(p *Plugin) {
				c := futureClaim("OTHER", "123456")
				c.claim.UserID = userID
				p.claims = &c
			}, body, http.StatusUnprocessableEntity},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				plugin := New(Config{Links: newFakeLinks(), Claims: &fakeClaims{}})
				tc.setup(plugin)
				rec := testsupport.NewRecorder()
				plugin.handleVerifyClaim(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodPost, "/x", tc.body), userID))
				testsupport.AssertStatus(t, rec, tc.want)
			})
		}
	})

	t.Run("expired", func(t *testing.T) {
		claims := futureClaim("ADMIN", "123456")
		claims.claim.UserID = userID
		claims.claim.ExpiresAt = time.Now().Add(-time.Minute)
		plugin := New(Config{Links: newFakeLinks(), Claims: &claims})
		rec := testsupport.NewRecorder()
		plugin.handleVerifyClaim(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodPost, "/x", body), userID))
		testsupport.AssertStatus(t, rec, http.StatusUnprocessableEntity)
		if claims.deleted == 0 {
			t.Fatal("an expired claim should be deleted")
		}
	})

	t.Run("too many attempts", func(t *testing.T) {
		claims := futureClaim("ADMIN", "123456")
		claims.claim.UserID = userID
		claims.claim.Attempts = claimMaxAttempts
		plugin := New(Config{Links: newFakeLinks(), Claims: &claims})
		rec := testsupport.NewRecorder()
		plugin.handleVerifyClaim(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodPost, "/x", body), userID))
		testsupport.AssertStatus(t, rec, http.StatusTooManyRequests)
	})

	t.Run("invalid code counts an attempt", func(t *testing.T) {
		claims := futureClaim("ADMIN", "123456")
		claims.claim.UserID = userID
		plugin := New(Config{Links: newFakeLinks(), Claims: &claims})
		rec := testsupport.NewRecorder()
		plugin.handleVerifyClaim(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodPost, "/x", `{"account_username":"ADMIN","code":"000000"}`), userID))
		testsupport.AssertStatus(t, rec, http.StatusUnprocessableEntity)
		if claims.incremented != 1 {
			t.Fatalf("incremented = %d", claims.incremented)
		}
	})
}

func TestHandleListClaims(t *testing.T) {
	plugin := New(Config{})
	rec := testsupport.NewRecorder()
	plugin.handleListClaims(rec, testsupport.NewRequest(http.MethodGet, "/x", ""))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)

	claims := &fakeClaims{list: []domain.Claim{{UserID: uuid.New(), AccountUsername: "ADMIN", ExpiresAt: time.Now(), Attempts: 2}}}
	plugin = New(Config{Claims: claims})
	rec = testsupport.NewRecorder()
	plugin.handleListClaims(rec, testsupport.NewRequest(http.MethodGet, "/x", ""))
	testsupport.AssertStatus(t, rec, http.StatusOK)
	if !strings.Contains(rec.Body.String(), "ADMIN") {
		t.Fatalf("body = %s", rec.Body.String())
	}

	claims.listErr = errors.New("db")
	rec = testsupport.NewRecorder()
	plugin.handleListClaims(rec, testsupport.NewRequest(http.MethodGet, "/x", ""))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)
}

func TestHandleMyAccount(t *testing.T) {
	plugin := New(Config{})
	rec := testsupport.NewRecorder()
	plugin.handleMyAccount(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodGet, "/x", ""), uuid.New()))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)

	plugin = New(Config{Links: newFakeLinks()})
	rec = testsupport.NewRecorder()
	plugin.handleMyAccount(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodGet, "/x", ""), uuid.New()))
	testsupport.AssertStatus(t, rec, http.StatusOK)
	if !strings.Contains(rec.Body.String(), `"linked":false`) {
		t.Fatalf("body = %s", rec.Body.String())
	}

	userID := uuid.New()
	links := newFakeLinks()
	links.links[userID] = domain.Link{UserID: userID, AccountUsername: "ADMIN"}
	plugin = New(Config{Links: links})
	rec = testsupport.NewRecorder()
	plugin.handleMyAccount(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodGet, "/x", ""), userID))
	testsupport.AssertStatus(t, rec, http.StatusOK)
	if !strings.Contains(rec.Body.String(), `"linked":true`) {
		t.Fatalf("body = %s", rec.Body.String())
	}

	links.getErr = errors.New("db")
	rec = testsupport.NewRecorder()
	plugin.handleMyAccount(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodGet, "/x", ""), userID))
	testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)
}

func TestHandleCreateMyAccount(t *testing.T) {
	userID := uuid.New()
	body := `{"username":"Alice","password":"secret"}`

	t.Run("missing links", func(t *testing.T) {
		plugin := New(Config{})
		rec := testsupport.NewRecorder()
		plugin.handleCreateMyAccount(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodPost, "/x", body), userID))
		testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)
	})

	t.Run("already linked", func(t *testing.T) {
		links := newFakeLinks()
		links.links[userID] = domain.Link{UserID: userID}
		plugin := New(Config{Links: links})
		rec := testsupport.NewRecorder()
		plugin.handleCreateMyAccount(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodPost, "/x", body), userID))
		testsupport.AssertStatus(t, rec, http.StatusConflict)
	})

	t.Run("link lookup error", func(t *testing.T) {
		links := newFakeLinks()
		links.getErr = errors.New("db")
		plugin := New(Config{Links: links})
		rec := testsupport.NewRecorder()
		plugin.handleCreateMyAccount(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodPost, "/x", body), userID))
		testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)
	})

	t.Run("invalid username", func(t *testing.T) {
		plugin := New(Config{Links: newFakeLinks(), Executor: &fakeExecutor{}})
		rec := testsupport.NewRecorder()
		plugin.handleCreateMyAccount(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodPost, "/x", `{"username":"bad name","password":"secret"}`), userID))
		testsupport.AssertStatus(t, rec, http.StatusUnprocessableEntity)
	})

	t.Run("command failure", func(t *testing.T) {
		plugin := New(Config{Links: newFakeLinks(), Executor: &fakeExecutor{err: errors.New("boom")}})
		rec := testsupport.NewRecorder()
		plugin.handleCreateMyAccount(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodPost, "/x", body), userID))
		testsupport.AssertStatus(t, rec, http.StatusBadGateway)
	})

	t.Run("success links the account", func(t *testing.T) {
		reader := &fakeReader{accounts: []azerothdb.Account{{ID: 3, Username: "Alice"}}}
		links := newFakeLinks()
		plugin := New(Config{Links: links, Executor: &fakeExecutor{result: "ok"}, Accounts: reader})
		rec := testsupport.NewRecorder()
		plugin.handleCreateMyAccount(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodPost, "/x", body), userID))
		testsupport.AssertStatus(t, rec, http.StatusCreated)
		if links.links[userID].AccountID == nil || *links.links[userID].AccountID != 3 {
			t.Fatalf("link = %+v", links.links[userID])
		}
	})

	t.Run("link store failure", func(t *testing.T) {
		links := newFakeLinks()
		links.upsertErr = domain.ErrAccountAlreadyLinked
		plugin := New(Config{Links: links, Executor: &fakeExecutor{result: "ok"}})
		rec := testsupport.NewRecorder()
		plugin.handleCreateMyAccount(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodPost, "/x", body), userID))
		testsupport.AssertStatus(t, rec, http.StatusServiceUnavailable)
	})
}

func TestLinkedAccount(t *testing.T) {
	if _, _, err := New(Config{}).LinkedAccount(context.Background(), uuid.NewString()); err == nil {
		t.Fatal("LinkedAccount without a store must fail")
	}
	plugin := New(Config{Links: newFakeLinks()})
	if _, _, err := plugin.LinkedAccount(context.Background(), "not-a-uuid"); err == nil {
		t.Fatal("invalid user id must fail")
	}
	if _, _, err := plugin.LinkedAccount(context.Background(), uuid.NewString()); err == nil {
		t.Fatal("missing link must fail")
	}

	userID := uuid.New()
	accountID := int64(4)
	links := newFakeLinks()
	links.links[userID] = domain.Link{UserID: userID, AccountUsername: "ADMIN", AccountID: &accountID}
	plugin = New(Config{Links: links})
	username, got, err := plugin.LinkedAccount(context.Background(), userID.String())
	if err != nil || username != "ADMIN" || got == nil || *got != accountID {
		t.Fatalf("LinkedAccount = %q,%v,%v", username, got, err)
	}
}

func TestPermissionDefsAreNamespaced(t *testing.T) {
	defs := permissionDefs()
	for _, def := range defs {
		if def.Owner != Name || def.Namespace != "azeroth" || !strings.HasPrefix(string(def.Name), "azeroth.") {
			t.Fatalf("definition not namespaced: %+v", def)
		}
	}
}

func newAccountRegistry() *plugins.Registry {
	return &plugins.Registry{
		Mux:         http.NewServeMux(),
		Services:    services.NewRegistry(),
		Permissions: permissions.NewRegistry(),
		Commands:    commands.NewRegistry(),
		Audit:       audit.NopRecorder{},
		RequireAuth: func(next http.Handler) http.Handler { return next },
		RequirePermission: func(_ permissions.Permission, next http.Handler) http.Handler {
			return next
		},
		RateLimit: func(next http.Handler) http.Handler { return next },
	}
}

func TestRegisterRequiresUserDirectory(t *testing.T) {
	if err := New(Config{Executor: &fakeExecutor{}}).Register(context.Background(), newAccountRegistry()); err == nil {
		t.Fatal("Register must fail without the user directory capability")
	}
}

func TestRegisterWiresAccount(t *testing.T) {
	reg := newAccountRegistry()
	if err := services.Provide[userdir.Directory](reg.Services, identityUserDirectory, &fakeDirectory{}); err != nil {
		t.Fatalf("provide directory: %v", err)
	}
	plugin := New(Config{Executor: &fakeExecutor{result: "ok"}, Links: newFakeLinks(), Claims: &fakeClaims{}})
	if err := plugin.Register(context.Background(), reg); err != nil {
		t.Fatalf("Register = %v", err)
	}

	if !reg.Permissions.Has(PermissionAccountSelf) {
		t.Fatal("account permission not registered")
	}
	if !reg.Services.Has(ServiceAccountDirectory) {
		t.Fatal("account directory capability not published")
	}
	if !reg.Commands.Has(CommandCreateAccount) || !reg.Commands.Has(CommandSetEmail) {
		t.Fatal("account commands not registered")
	}

	rec := testsupport.NewRecorder()
	reg.Mux.ServeHTTP(rec, testsupport.WithPrincipal(testsupport.NewRequest(http.MethodGet, "/api/v1/azeroth/me/account", ""), uuid.New()))
	testsupport.AssertStatus(t, rec, http.StatusOK)
}
