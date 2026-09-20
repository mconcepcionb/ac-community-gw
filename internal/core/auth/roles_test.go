package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type stubRoleSource struct {
	roles []string
	err   error
	calls int
}

func (s *stubRoleSource) RolesForUser(context.Context, uuid.UUID) ([]string, error) {
	s.calls++
	return s.roles, s.err
}

func TestManagerRefreshesRolesFromSource(t *testing.T) {
	store := &recordingStore{}
	source := &stubRoleSource{roles: []string{"member"}}
	manager := NewManager(ManagerOptions{Store: store, TTL: time.Hour, RoleSource: source})

	token, err := manager.Create(context.Background(), Principal{UserID: uuid.New(), Roles: []string{"admin"}})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	principal, err := manager.Resolve(context.Background(), token)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(principal.Roles) != 1 || principal.Roles[0] != "member" {
		t.Fatalf("roles = %v, want [member]", principal.Roles)
	}
	if source.calls != 1 {
		t.Fatalf("source calls = %d", source.calls)
	}
}

func TestManagerRoleSourceErrorFallsBackToSessionRoles(t *testing.T) {
	store := &recordingStore{}
	source := &stubRoleSource{err: errors.New("db down")}
	manager := NewManager(ManagerOptions{Store: store, TTL: time.Hour, RoleSource: source})

	token, err := manager.Create(context.Background(), Principal{UserID: uuid.New(), Roles: []string{"admin"}})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	principal, err := manager.Resolve(context.Background(), token)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(principal.Roles) != 1 || principal.Roles[0] != "admin" {
		t.Fatalf("roles = %v, want session fallback [admin]", principal.Roles)
	}
}

func TestCachingRoleSourceCachesWithinTTL(t *testing.T) {
	inner := &stubRoleSource{roles: []string{"a"}}
	source := NewCachingRoleSource(inner, time.Minute).(*cachingRoleSource)
	now := time.Now()
	source.now = func() time.Time { return now }

	ctx := context.Background()
	userID := uuid.New()
	for i := 0; i < 3; i++ {
		if _, err := source.RolesForUser(ctx, userID); err != nil {
			t.Fatalf("RolesForUser: %v", err)
		}
	}
	if inner.calls != 1 {
		t.Fatalf("inner calls = %d, want 1", inner.calls)
	}

	now = now.Add(2 * time.Minute)
	if _, err := source.RolesForUser(ctx, userID); err != nil {
		t.Fatalf("RolesForUser after ttl: %v", err)
	}
	if inner.calls != 2 {
		t.Fatalf("inner calls = %d, want 2", inner.calls)
	}
}

func TestCachingRoleSourceDoesNotCacheErrors(t *testing.T) {
	inner := &stubRoleSource{err: errors.New("boom")}
	source := NewCachingRoleSource(inner, time.Minute)
	ctx := context.Background()
	userID := uuid.New()
	for i := 0; i < 2; i++ {
		if _, err := source.RolesForUser(ctx, userID); err == nil {
			t.Fatal("expected error")
		}
	}
	if inner.calls != 2 {
		t.Fatalf("inner calls = %d, want 2", inner.calls)
	}
}
