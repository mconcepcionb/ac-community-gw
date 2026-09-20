package azerothcharacter

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
)

// newMailPlugin wires a plugin whose caller is linked to account 5 and whose
// recipient Thrall belongs to account 5.
func newMailPlugin(executor *fakeExecutor) *Plugin {
	accountID := int64(5)
	plugin := New(Config{
		Executor:   executor,
		Characters: &fakeCharacters{characters: []azerothdb.Character{{Name: "Thrall", AccountID: 5}}},
	})
	plugin.directory = &fakeDirectory{username: "ADMIN", accountID: &accountID}
	return plugin
}

func TestSendMailRejectsNonOwner(t *testing.T) {
	accountID := int64(6)
	executor := &fakeExecutor{result: "should not run"}
	plugin := New(Config{
		Executor:   executor,
		Characters: &fakeCharacters{characters: []azerothdb.Character{{Name: "Thrall", AccountID: 5}}},
	})
	plugin.directory = &fakeDirectory{username: "OTHER", accountID: &accountID}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/mail",
		strings.NewReader(`{"character":"Thrall","money":100}`))
	rec := httptest.NewRecorder()
	plugin.handleSendMail(rec, req)

	if rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "not_owner") {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if executor.last != "" {
		t.Fatalf("mail reached executor: %q", executor.last)
	}
}

func TestSendMailRejectsNoLinkedAccount(t *testing.T) {
	executor := &fakeExecutor{result: "should not run"}
	plugin := New(Config{
		Executor:   executor,
		Characters: &fakeCharacters{characters: []azerothdb.Character{{Name: "Thrall", AccountID: 5}}},
	})
	plugin.directory = &fakeDirectory{}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/mail",
		strings.NewReader(`{"character":"Thrall","money":100}`))
	rec := httptest.NewRecorder()
	plugin.handleSendMail(rec, req)

	if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), "account_not_linked") {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if executor.last != "" {
		t.Fatalf("mail reached executor: %q", executor.last)
	}
}

func TestAdminMailBypassesOwnership(t *testing.T) {
	executor := &fakeExecutor{result: "Mail sent"}
	plugin := New(Config{
		Executor:   executor,
		Characters: &fakeCharacters{characters: []azerothdb.Character{{Name: "Thrall", AccountID: 5}}},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/admin/characters/Thrall/mail",
		strings.NewReader(`{"money":100}`))
	req.SetPathValue("name", "Thrall")
	rec := httptest.NewRecorder()
	plugin.handleAdminMail(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(executor.last, ".send money Thrall") {
		t.Fatalf("command = %q", executor.last)
	}
}

func TestAdminMailRejectsUnknownCharacter(t *testing.T) {
	executor := &fakeExecutor{result: "should not run"}
	plugin := New(Config{Executor: executor, Characters: &fakeCharacters{}})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/admin/characters/Ghost/mail",
		strings.NewReader(`{"money":100}`))
	req.SetPathValue("name", "Ghost")
	rec := httptest.NewRecorder()
	plugin.handleAdminMail(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if executor.last != "" {
		t.Fatalf("mail reached executor: %q", executor.last)
	}
}

func TestSendMailFailsClosedWithoutCharacterDB(t *testing.T) {
	accountID := int64(5)
	executor := &fakeExecutor{result: "should not run"}
	plugin := New(Config{Executor: executor})
	plugin.directory = &fakeDirectory{username: "ADMIN", accountID: &accountID}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/mail",
		strings.NewReader(`{"character":"Thrall","money":100}`))
	rec := httptest.NewRecorder()
	plugin.handleSendMail(rec, req)

	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), "character_db_not_configured") {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if executor.last != "" {
		t.Fatalf("mail reached executor: %q", executor.last)
	}
}

func TestSendMailFailsClosedWithoutDirectory(t *testing.T) {
	executor := &fakeExecutor{result: "should not run"}
	plugin := New(Config{
		Executor:   executor,
		Characters: &fakeCharacters{characters: []azerothdb.Character{{Name: "Thrall", AccountID: 5}}},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/mail",
		strings.NewReader(`{"character":"Thrall","money":100}`))
	rec := httptest.NewRecorder()
	plugin.handleSendMail(rec, req)

	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), "identity_storage_unavailable") {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if executor.last != "" {
		t.Fatalf("mail reached executor: %q", executor.last)
	}
}
