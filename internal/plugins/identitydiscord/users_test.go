package identitydiscord

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/userdir"
)

func TestListUsersReturnsTypedResponse(t *testing.T) {
	repo := &fakeRepo{users: []userdir.User{{
		ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		DiscordID:   "42",
		Username:    "user",
		GlobalName:  "User",
		DisplayName: "User#42",
		CreatedAt:   time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
	}}}
	plugin, _, _ := newTestPlugin(t, &fakeProvider{}, repo)

	rec := httptest.NewRecorder()
	plugin.handleListUsers(rec, httptest.NewRequest(http.MethodGet, "/api/v1/identity/users", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	want := `{"users":[{"user_id":"11111111-1111-1111-1111-111111111111","discord_id":"42","username":"user","global_name":"User","display_name":"User#42","created_at":"2024-01-02T03:04:05Z"}]}` + "\n"
	if rec.Body.String() != want {
		t.Fatalf("body = %s\nwant = %s", rec.Body.String(), want)
	}
}

func TestListUsersEmptyReturnsArray(t *testing.T) {
	plugin, _, _ := newTestPlugin(t, &fakeProvider{}, &fakeRepo{})

	rec := httptest.NewRecorder()
	plugin.handleListUsers(rec, httptest.NewRequest(http.MethodGet, "/api/v1/identity/users", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Body.String(); got != `{"users":[]}`+"\n" {
		t.Fatalf("body = %q", got)
	}
}

func TestListUsersUnavailableWithoutRepository(t *testing.T) {
	plugin := New(Config{})

	rec := httptest.NewRecorder()
	plugin.handleListUsers(rec, httptest.NewRequest(http.MethodGet, "/api/v1/identity/users", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
}
