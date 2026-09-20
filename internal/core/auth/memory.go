package auth

import (
	"context"
	"sync"
)

// MemoryStore is a non-durable, in-memory session store.
//
// It exists so the gateway and its tests run without PostgreSQL. Production
// deployments are expected to use a PostgreSQL-backed store owned by the
// identity-discord plugin.
type MemoryStore struct {
	mu       sync.RWMutex
	sessions map[string]Session
}

// NewMemoryStore creates an empty in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{sessions: make(map[string]Session)}
}

// Create stores a session.
func (s *MemoryStore) Create(_ context.Context, session Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.ID] = session
	return nil
}

// Get returns a session by token.
func (s *MemoryStore) Get(_ context.Context, id string) (Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[id]
	if !ok {
		return Session{}, ErrSessionNotFound
	}
	return session, nil
}

// Delete removes a session by token.
func (s *MemoryStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
	return nil
}
