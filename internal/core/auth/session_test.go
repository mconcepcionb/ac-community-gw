package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type recordingStore struct {
	created Session
	getKey  string
	deleted string
}

func (s *recordingStore) Create(_ context.Context, session Session) error {
	s.created = session
	return nil
}

func (s *recordingStore) Get(_ context.Context, id string) (Session, error) {
	s.getKey = id
	return s.created, nil
}

func (s *recordingStore) Delete(_ context.Context, id string) error {
	s.deleted = id
	return nil
}

func TestManagerHashesSessionTokens(t *testing.T) {
	store := &recordingStore{}
	manager := NewManager(ManagerOptions{Store: store, TTL: time.Hour})

	principal := Principal{UserID: uuid.New(), DiscordID: "42", Roles: []string{"member"}}
	token, err := manager.Create(context.Background(), principal)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	hash := sha256Hex(token)
	if store.created.ID != hash {
		t.Fatalf("stored id = %q, want hash %q", store.created.ID, hash)
	}
	if store.created.ID == token {
		t.Fatal("raw token must never be stored")
	}
	if len(store.created.Roles) != 1 || store.created.Roles[0] != "member" {
		t.Fatalf("roles = %v", store.created.Roles)
	}

	resolved, err := manager.Resolve(context.Background(), token)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if store.getKey != hash {
		t.Fatalf("lookup key = %q, want hash %q", store.getKey, hash)
	}
	if resolved.DiscordID != "42" {
		t.Fatalf("resolved = %+v", resolved)
	}

	if err := manager.Revoke(context.Background(), token); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
	if store.deleted != hash {
		t.Fatalf("deleted key = %q, want hash %q", store.deleted, hash)
	}
}

func TestManagerRejectsExpiredSession(t *testing.T) {
	store := &recordingStore{}
	now := time.Now()
	manager := NewManager(ManagerOptions{
		Store: store,
		TTL:   time.Hour,
		Now:   func() time.Time { return now },
	})

	token, err := manager.Create(context.Background(), Principal{UserID: uuid.New()})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	now = now.Add(2 * time.Hour)
	if _, err := manager.Resolve(context.Background(), token); !errors.Is(err, ErrSessionExpired) {
		t.Fatalf("expected ErrSessionExpired, got %v", err)
	}
	if store.deleted != sha256Hex(token) {
		t.Fatal("expired session should be deleted")
	}
}

func TestParseSameSite(t *testing.T) {
	cases := map[string]http.SameSite{
		"":       http.SameSiteLaxMode,
		"lax":    http.SameSiteLaxMode,
		"LAX":    http.SameSiteLaxMode,
		"strict": http.SameSiteStrictMode,
		"none":   http.SameSiteNoneMode,
	}
	for input, want := range cases {
		got, err := ParseSameSite(input)
		if err != nil {
			t.Fatalf("ParseSameSite(%q): %v", input, err)
		}
		if got != want {
			t.Fatalf("ParseSameSite(%q) = %v, want %v", input, got, want)
		}
	}
	if _, err := ParseSameSite("bogus"); err == nil {
		t.Fatal("expected error for invalid value")
	}
}

func TestManagerUsesConfiguredSameSite(t *testing.T) {
	manager := NewManager(ManagerOptions{
		Store:    &recordingStore{},
		TTL:      time.Hour,
		SameSite: http.SameSiteStrictMode,
	})
	rec := httptest.NewRecorder()
	manager.SetCookie(rec, "token", time.Now().Add(time.Hour))

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %v", cookies)
	}
	if cookies[0].SameSite != http.SameSiteStrictMode {
		t.Fatalf("SameSite = %v", cookies[0].SameSite)
	}
	if !cookies[0].HttpOnly {
		t.Fatal("cookie should be HttpOnly")
	}
}

func TestSetCookieMaxAgeUsesManagerClock(t *testing.T) {
	now := time.Now()
	manager := NewManager(ManagerOptions{
		Store: &recordingStore{},
		TTL:   time.Hour,
		Now:   func() time.Time { return now },
	})
	rec := httptest.NewRecorder()
	manager.SetCookie(rec, "token", now.Add(time.Hour))

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %v", cookies)
	}
	if cookies[0].MaxAge != 3600 {
		t.Fatalf("MaxAge = %d, want 3600", cookies[0].MaxAge)
	}
}

type failingDeleteStore struct {
	recordingStore
}

func (s *failingDeleteStore) Delete(context.Context, string) error {
	return errors.New("delete failed")
}

func TestResolveExpiredSessionSurfacesDeleteError(t *testing.T) {
	store := &failingDeleteStore{}
	now := time.Now()
	manager := NewManager(ManagerOptions{
		Store: store,
		TTL:   time.Hour,
		Now:   func() time.Time { return now },
	})
	token, err := manager.Create(context.Background(), Principal{UserID: uuid.New()})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	now = now.Add(2 * time.Hour)
	_, err = manager.Resolve(context.Background(), token)
	if !errors.Is(err, ErrSessionExpired) {
		t.Fatalf("error = %v, want ErrSessionExpired", err)
	}
	if !strings.Contains(err.Error(), "cleanup failed") {
		t.Fatalf("cleanup failure not surfaced: %v", err)
	}
}
