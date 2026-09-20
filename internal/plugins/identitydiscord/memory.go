package identitydiscord

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrStateNotFound is returned when an OAuth state is unknown, already used or
// expired. Both state stores return it so callers cannot tell them apart.
var ErrStateNotFound = errors.New("identitydiscord: oauth state not found")

// MemoryStateStore is a non-durable OAuth state store used for tests and when
// the gateway runs without PostgreSQL.
type MemoryStateStore struct {
	mu     sync.Mutex
	states map[string]memoryState
}

type memoryState struct {
	codeVerifier string
	returnTo     string
	expiresAt    time.Time
}

// NewMemoryStateStore creates an empty in-memory state store.
func NewMemoryStateStore() *MemoryStateStore {
	return &MemoryStateStore{states: make(map[string]memoryState)}
}

// CreateOAuthState implements StateStore.
func (s *MemoryStateStore) CreateOAuthState(_ context.Context, state, codeVerifier, returnTo string, expiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[state] = memoryState{codeVerifier: codeVerifier, returnTo: returnTo, expiresAt: expiresAt}
	return nil
}

// ConsumeOAuthState implements StateStore. It deletes the state on use and
// rejects expired states.
func (s *MemoryStateStore) ConsumeOAuthState(_ context.Context, state string) (string, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.states[state]
	if !ok {
		return "", "", ErrStateNotFound
	}
	delete(s.states, state)
	if !time.Now().Before(record.expiresAt) {
		return "", "", ErrStateNotFound
	}
	return record.codeVerifier, record.returnTo, nil
}
