package azerothaccount

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/userdir"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothaccount/domain"
)

type fakeExecutor struct {
	last   string
	result string
	err    error
}

func (f *fakeExecutor) Execute(_ context.Context, command string) (string, error) {
	f.last = command
	return f.result, f.err
}

type fakeReader struct {
	accounts []azerothdb.Account
	last     azerothdb.Query
	err      error
	findErr  error
}

func (f *fakeReader) ListAccounts(_ context.Context, query azerothdb.Query) ([]azerothdb.Account, error) {
	f.last = query
	return f.accounts, f.err
}

func (f *fakeReader) FindAccountByUsername(_ context.Context, username string) (azerothdb.Account, error) {
	if f.findErr != nil {
		return azerothdb.Account{}, f.findErr
	}
	for _, account := range f.accounts {
		if account.Username == username {
			return account, nil
		}
	}
	return azerothdb.Account{}, azerothdb.ErrAccountNotFound
}

type fakeLinks struct {
	links     map[uuid.UUID]domain.Link
	byUser    map[string]uuid.UUID
	upsertErr error
}

func newFakeLinks() *fakeLinks {
	return &fakeLinks{links: map[uuid.UUID]domain.Link{}, byUser: map[string]uuid.UUID{}}
}

func (f *fakeLinks) Upsert(_ context.Context, userID uuid.UUID, username string, accountID *int64) (domain.Link, error) {
	if f.upsertErr != nil {
		return domain.Link{}, f.upsertErr
	}
	if owner, ok := f.byUser[username]; ok && owner != userID {
		return domain.Link{}, domain.ErrAccountAlreadyLinked
	}
	link := domain.Link{UserID: userID, AccountUsername: username, AccountID: accountID, LinkedAt: time.Now(), UpdatedAt: time.Now()}
	f.links[userID] = link
	f.byUser[username] = userID
	return link, nil
}

func (f *fakeLinks) Get(_ context.Context, userID uuid.UUID) (domain.Link, error) {
	link, ok := f.links[userID]
	if !ok {
		return domain.Link{}, domain.ErrLinkNotFound
	}
	return link, nil
}

func (f *fakeLinks) Delete(_ context.Context, userID uuid.UUID) error {
	delete(f.links, userID)
	return nil
}

func (f *fakeLinks) List(_ context.Context) ([]domain.Link, error) {
	out := make([]domain.Link, 0, len(f.links))
	for _, link := range f.links {
		out = append(out, link)
	}
	return out, nil
}

type fakeDirectory struct {
	ids    map[string]uuid.UUID
	byName map[string][]userdir.User
	err    error
}

func (f *fakeDirectory) ResolveUser(_ context.Context, discordID string) (uuid.UUID, bool, error) {
	if f.err != nil {
		return uuid.Nil, false, f.err
	}
	id, ok := f.ids[discordID]
	return id, ok, nil
}

func (f *fakeDirectory) ResolveUserByName(_ context.Context, name string) ([]userdir.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.byName[name], nil
}

func (f *fakeDirectory) ListUsers(_ context.Context, _ string, _, _ int) ([]userdir.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	out := make([]userdir.User, 0)
	for _, users := range f.byName {
		out = append(out, users...)
	}
	return out, nil
}

func TestHandleCreateAccount(t *testing.T) {
	executor := &fakeExecutor{result: "Account created: Alice"}
	plugin := New(Config{Executor: executor})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/accounts",
		strings.NewReader(`{"username":"Alice","password":"secret","email":"alice@example.com"}`))
	rec := httptest.NewRecorder()
	plugin.handleCreateAccount(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if executor.last != ".account create Alice secret alice@example.com" {
		t.Fatalf("command = %q", executor.last)
	}
	if !strings.Contains(rec.Body.String(), "Account created: Alice") {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestHandleCreateAccountValidation(t *testing.T) {
	plugin := New(Config{Executor: &fakeExecutor{}})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/accounts",
		strings.NewReader(`{"password":"secret"}`))
	rec := httptest.NewRecorder()
	plugin.handleCreateAccount(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestHandleChangePasswordUsesPathUsername(t *testing.T) {
	executor := &fakeExecutor{result: "The password was changed"}
	plugin := New(Config{Executor: executor})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/accounts/Alice/password",
		strings.NewReader(`{"password":"newpass"}`))
	req.SetPathValue("username", "Alice")
	rec := httptest.NewRecorder()
	plugin.handleChangePassword(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if executor.last != ".account set password Alice newpass newpass" {
		t.Fatalf("command = %q", executor.last)
	}
}

func TestHandleSetEmailUsesPathUsername(t *testing.T) {
	executor := &fakeExecutor{result: "The email was changed"}
	plugin := New(Config{Executor: executor})

	req := httptest.NewRequest(http.MethodPut, "/api/v1/azeroth/accounts/Alice/email",
		strings.NewReader(`{"email":"new@example.com"}`))
	req.SetPathValue("username", "Alice")
	rec := httptest.NewRecorder()
	plugin.handleSetEmail(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if executor.last != ".account set email Alice new@example.com" {
		t.Fatalf("command = %q", executor.last)
	}
}

func TestHandleCreateAccountUpstreamFailure(t *testing.T) {
	plugin := New(Config{Executor: &fakeExecutor{err: errors.New("boom")}})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/accounts",
		strings.NewReader(`{"username":"Alice","password":"secret"}`))
	rec := httptest.NewRecorder()
	plugin.handleCreateAccount(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestHandleListAccounts(t *testing.T) {
	reader := &fakeReader{accounts: []azerothdb.Account{
		{ID: 1, Username: "ADMIN", Email: "admin@example.com", GMLevel: 3, Online: true},
	}}
	plugin := New(Config{Executor: &fakeExecutor{}, Accounts: reader})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/azeroth/accounts?filter=ad&limit=10", nil)
	rec := httptest.NewRecorder()
	plugin.handleListAccounts(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if reader.last.Filter != "ad" || reader.last.Limit != 10 {
		t.Fatalf("query = %+v", reader.last)
	}
	if !strings.Contains(rec.Body.String(), "ADMIN") {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestHandleListAccountsUnavailable(t *testing.T) {
	plugin := New(Config{Executor: &fakeExecutor{}})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/azeroth/accounts", nil)
	rec := httptest.NewRecorder()
	plugin.handleListAccounts(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "login_db_not_configured") {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestHandleListAccountsDatabaseDown(t *testing.T) {
	plugin := New(Config{Executor: &fakeExecutor{}, Accounts: azerothdb.Unavailable{Err: errors.New("dial tcp: refused")}})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/azeroth/accounts", nil)
	rec := httptest.NewRecorder()
	plugin.handleListAccounts(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "login_db_unavailable") {
		t.Fatalf("body = %s", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "refused") {
		t.Fatalf("internal error leaked: %s", rec.Body.String())
	}
}

func TestHandleCreateLinkByDiscordID(t *testing.T) {
	userID := uuid.New()
	reader := &fakeReader{accounts: []azerothdb.Account{{ID: 7, Username: "ADMIN"}}}
	links := newFakeLinks()
	plugin := New(Config{Executor: &fakeExecutor{}, Accounts: reader, Links: links})
	plugin.users = &fakeDirectory{ids: map[string]uuid.UUID{"42": userID}}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/account-links",
		strings.NewReader(`{"discord_id":"42","account_username":"ADMIN"}`))
	rec := httptest.NewRecorder()
	plugin.handleCreateLink(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	link, ok := links.links[userID]
	if !ok || link.AccountUsername != "ADMIN" || link.AccountID == nil || *link.AccountID != 7 {
		t.Fatalf("link = %+v", link)
	}
}

func TestHandleCreateLinkByDiscordUsername(t *testing.T) {
	userID := uuid.New()
	reader := &fakeReader{accounts: []azerothdb.Account{{ID: 9, Username: "ADMIN"}}}
	links := newFakeLinks()
	plugin := New(Config{Executor: &fakeExecutor{}, Accounts: reader, Links: links})
	plugin.users = &fakeDirectory{byName: map[string][]userdir.User{
		"alice": {{ID: userID, DiscordID: "42", Username: "alice"}},
	}}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/account-links",
		strings.NewReader(`{"discord_username":"alice","account_username":"ADMIN"}`))
	rec := httptest.NewRecorder()
	plugin.handleCreateLink(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if _, ok := links.links[userID]; !ok {
		t.Fatal("link not created")
	}
}

func TestHandleCreateLinkAmbiguousUsername(t *testing.T) {
	reader := &fakeReader{accounts: []azerothdb.Account{{ID: 9, Username: "ADMIN"}}}
	plugin := New(Config{Executor: &fakeExecutor{}, Accounts: reader, Links: newFakeLinks()})
	plugin.users = &fakeDirectory{byName: map[string][]userdir.User{
		"bob": {{ID: uuid.New(), Username: "bob"}, {ID: uuid.New(), Username: "bob"}},
	}}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/account-links",
		strings.NewReader(`{"discord_username":"bob","account_username":"ADMIN"}`))
	rec := httptest.NewRecorder()
	plugin.handleCreateLink(rec, req)

	if rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), "ambiguous_user") {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandleCreateLinkMissingTarget(t *testing.T) {
	plugin := New(Config{Executor: &fakeExecutor{}, Accounts: &fakeReader{}, Links: newFakeLinks()})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/account-links",
		strings.NewReader(`{"account_username":"ADMIN"}`))
	rec := httptest.NewRecorder()
	plugin.handleCreateLink(rec, req)

	if rec.Code != http.StatusUnprocessableEntity || !strings.Contains(rec.Body.String(), "missing_user") {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandleCreateLinkUnknownUser(t *testing.T) {
	plugin := New(Config{Executor: &fakeExecutor{}, Accounts: &fakeReader{}, Links: newFakeLinks()})
	plugin.users = &fakeDirectory{ids: map[string]uuid.UUID{}}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/account-links",
		strings.NewReader(`{"discord_id":"nope","account_username":"ADMIN"}`))
	rec := httptest.NewRecorder()
	plugin.handleCreateLink(rec, req)

	if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), "user_not_found") {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandleCreateLinkAccountNotFound(t *testing.T) {
	plugin := New(Config{Executor: &fakeExecutor{}, Accounts: &fakeReader{}, Links: newFakeLinks()})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/account-links",
		strings.NewReader(`{"user_id":"`+uuid.NewString()+`","account_username":"GHOST"}`))
	rec := httptest.NewRecorder()
	plugin.handleCreateLink(rec, req)

	if rec.Code != http.StatusUnprocessableEntity || !strings.Contains(rec.Body.String(), "account_not_found") {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandleCreateLinkAlreadyLinked(t *testing.T) {
	links := newFakeLinks()
	links.upsertErr = domain.ErrAccountAlreadyLinked
	plugin := New(Config{Executor: &fakeExecutor{}, Accounts: &fakeReader{accounts: []azerothdb.Account{{ID: 1, Username: "ADMIN"}}}, Links: links})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/account-links",
		strings.NewReader(`{"user_id":"`+uuid.NewString()+`","account_username":"ADMIN"}`))
	rec := httptest.NewRecorder()
	plugin.handleCreateLink(rec, req)

	if rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), "account_already_linked") {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandleGetAndDeleteLink(t *testing.T) {
	userID := uuid.New()
	links := newFakeLinks()
	links.links[userID] = domain.Link{UserID: userID, AccountUsername: "ADMIN", LinkedAt: time.Now(), UpdatedAt: time.Now()}
	plugin := New(Config{Executor: &fakeExecutor{}, Links: links})

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/azeroth/account-links/"+userID.String(), nil)
	getReq.SetPathValue("user_id", userID.String())
	getRec := httptest.NewRecorder()
	plugin.handleGetLink(getRec, getReq)
	if getRec.Code != http.StatusOK || !strings.Contains(getRec.Body.String(), "ADMIN") {
		t.Fatalf("get status = %d body=%s", getRec.Code, getRec.Body.String())
	}

	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/azeroth/account-links/"+userID.String(), nil)
	delReq.SetPathValue("user_id", userID.String())
	delRec := httptest.NewRecorder()
	plugin.handleDeleteLink(delRec, delReq)
	if delRec.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d", delRec.Code)
	}
}
