package azerothcharacter

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
)

func TestHandleListCharactersByAccount(t *testing.T) {
	characters := &fakeCharacters{characters: []azerothdb.Character{
		{GUID: 1, Name: "Thrall", Race: 2, Class: 7, Level: 80, Online: true},
	}}
	accounts := &fakeAccounts{found: true, account: azerothdb.Account{ID: 5, Username: "ADMIN"}}
	plugin := New(Config{Characters: characters, Accounts: accounts})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/azeroth/characters?account=ADMIN", nil)
	rec := httptest.NewRecorder()
	plugin.handleListCharacters(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if characters.last.AccountID != 5 {
		t.Fatalf("account id = %d", characters.last.AccountID)
	}
	if !strings.Contains(rec.Body.String(), "Thrall") || !strings.Contains(rec.Body.String(), "Shaman") {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestHandleListUserCharacters(t *testing.T) {
	accountID := int64(5)
	characters := &fakeCharacters{characters: []azerothdb.Character{{GUID: 1, Name: "Thrall", Level: 80}}}
	plugin := New(Config{Characters: characters})
	plugin.directory = &fakeDirectory{username: "ADMIN", accountID: &accountID}

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/azeroth/users/"+userID.String()+"/characters", nil)
	req.SetPathValue("user_id", userID.String())
	rec := httptest.NewRecorder()
	plugin.handleListUserCharacters(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if characters.last.AccountID != 5 {
		t.Fatalf("account id = %d", characters.last.AccountID)
	}
}

func TestHandleListUserCharactersNoLink(t *testing.T) {
	plugin := New(Config{Characters: &fakeCharacters{}})
	plugin.directory = &fakeDirectory{}
	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/azeroth/users/"+userID.String()+"/characters", nil)
	req.SetPathValue("user_id", userID.String())
	rec := httptest.NewRecorder()
	plugin.handleListUserCharacters(rec, req)

	if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), "account_not_linked") {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandleGetCharacterNotFound(t *testing.T) {
	plugin := New(Config{Characters: &fakeCharacters{}})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/azeroth/characters/Ghost", nil)
	req.SetPathValue("name", "Ghost")
	rec := httptest.NewRecorder()
	plugin.handleGetCharacter(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListCharactersExactJSON(t *testing.T) {
	characters := &fakeCharacters{characters: []azerothdb.Character{{
		GUID: 7, AccountID: 5, Name: "Thrall", Race: 2, Class: 7,
		Level: 80, Online: true, LogoutTime: 1704164645, TotalTime: 1234,
		Money: 999, GuildName: "Horde",
	}}}
	plugin := New(Config{Characters: characters})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/azeroth/characters?account_id=5", nil)
	rec := httptest.NewRecorder()
	plugin.handleListCharacters(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	want := `{"characters":[{"guid":7,"name":"Thrall","race":2,"race_name":"Orc","class":7,"class_name":"Shaman","gender":0,"level":80,"online":true,"guild":"Horde","money":999,"total_time":1234,"logout_time":"2024-01-02T03:04:05Z"}]}` + "\n"
	if rec.Body.String() != want {
		t.Fatalf("body = %s\nwant = %s", rec.Body.String(), want)
	}
}

func TestGetCharacterExactJSONNullLogout(t *testing.T) {
	characters := &fakeCharacters{characters: []azerothdb.Character{{GUID: 1, Name: "Thrall"}}}
	plugin := New(Config{Characters: characters})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/azeroth/characters/Thrall", nil)
	req.SetPathValue("name", "Thrall")
	rec := httptest.NewRecorder()
	plugin.handleGetCharacter(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	want := `{"guid":1,"name":"Thrall","race":0,"race_name":"Unknown","class":0,"class_name":"Unknown","gender":0,"level":0,"online":false,"guild":"","money":0,"total_time":0,"logout_time":null}` + "\n"
	if rec.Body.String() != want {
		t.Fatalf("body = %s\nwant = %s", rec.Body.String(), want)
	}
}

func TestHandleSendItems(t *testing.T) {
	executor := &fakeExecutor{result: "Mail sent"}
	plugin := newMailPlugin(executor)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/mail",
		strings.NewReader(`{"character":"Thrall","subject":"Reward","body":"hi","items":[{"id":1234,"count":2}]}`))
	rec := httptest.NewRecorder()
	plugin.handleSendMail(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if executor.last != `.send items Thrall "Reward" "hi" 1234:2` {
		t.Fatalf("command = %q", executor.last)
	}
}

func TestHandleSendMoney(t *testing.T) {
	executor := &fakeExecutor{result: "Mail sent"}
	plugin := newMailPlugin(executor)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/mail",
		strings.NewReader(`{"character":"Thrall","money":10000}`))
	rec := httptest.NewRecorder()
	plugin.handleSendMail(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if executor.last != `.send money Thrall "Reward" " " 10000` {
		t.Fatalf("command = %q", executor.last)
	}
}

func TestHandleSendMailSanitizesSubject(t *testing.T) {
	executor := &fakeExecutor{result: "ok"}
	plugin := newMailPlugin(executor)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/mail",
		strings.NewReader(`{"character":"Thrall","subject":"a\" b\nc","items":[{"id":1,"count":1}]}`))
	rec := httptest.NewRecorder()
	plugin.handleSendMail(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	command := executor.last
	if strings.Contains(command, "\n") {
		t.Fatalf("newline not sanitized: %q", command)
	}
	if strings.Count(command, `"`) != 4 {
		t.Fatalf("unexpected quotes in command: %q", command)
	}
}

func TestHandleSendMailRejectsBadRecipient(t *testing.T) {
	plugin := New(Config{Executor: &fakeExecutor{}, Characters: &fakeCharacters{}})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/mail",
		strings.NewReader(`{"character":"Thrall; .server shutdown","items":[{"id":1,"count":1}]}`))
	rec := httptest.NewRecorder()
	plugin.handleSendMail(rec, req)

	if rec.Code != http.StatusUnprocessableEntity || !strings.Contains(rec.Body.String(), "invalid_recipient") {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandleSendMailEmptyDelivery(t *testing.T) {
	plugin := New(Config{Executor: &fakeExecutor{}, Characters: &fakeCharacters{characters: []azerothdb.Character{{Name: "Thrall"}}}})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/mail",
		strings.NewReader(`{"character":"Thrall"}`))
	rec := httptest.NewRecorder()
	plugin.handleSendMail(rec, req)

	if rec.Code != http.StatusUnprocessableEntity || !strings.Contains(rec.Body.String(), "empty_delivery") {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestSendMailExactJSON(t *testing.T) {
	executor := &fakeExecutor{result: "Mail sent"}
	plugin := newMailPlugin(executor)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/azeroth/mail",
		strings.NewReader(`{"character":"Thrall","subject":"Reward","body":"hi","items":[{"id":1234,"count":2}]}`))
	rec := httptest.NewRecorder()
	plugin.handleSendMail(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	want := `{"recipient":"Thrall","results":["Mail sent"]}` + "\n"
	if rec.Body.String() != want {
		t.Fatalf("body = %s\nwant = %s", rec.Body.String(), want)
	}
}
