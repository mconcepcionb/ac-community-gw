package identitydiscord

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMemoryStateStoreConsume(t *testing.T) {
	store := NewMemoryStateStore()
	ctx := context.Background()

	if err := store.CreateOAuthState(ctx, "s1", "verifier-1", "/store", time.Now().Add(time.Minute)); err != nil {
		t.Fatalf("create: %v", err)
	}
	verifier, returnTo, err := store.ConsumeOAuthState(ctx, "s1")
	if err != nil {
		t.Fatalf("consume: %v", err)
	}
	if verifier != "verifier-1" || returnTo != "/store" {
		t.Fatalf("got %q %q", verifier, returnTo)
	}

	if _, _, err := store.ConsumeOAuthState(ctx, "s1"); !errors.Is(err, ErrStateNotFound) {
		t.Fatalf("replay error = %v, want ErrStateNotFound", err)
	}

	if _, _, err := store.ConsumeOAuthState(ctx, "unknown"); !errors.Is(err, ErrStateNotFound) {
		t.Fatalf("unknown error = %v, want ErrStateNotFound", err)
	}

	if err := store.CreateOAuthState(ctx, "s2", "v", "", time.Now().Add(-time.Second)); err != nil {
		t.Fatalf("create expired: %v", err)
	}
	if _, _, err := store.ConsumeOAuthState(ctx, "s2"); !errors.Is(err, ErrStateNotFound) {
		t.Fatalf("expired error = %v, want ErrStateNotFound", err)
	}
}
