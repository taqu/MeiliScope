package platform

import (
	"context"
	"sync"
)

// sessionStore is an in-memory credential store.
// Credentials are discarded when the application exits and are never written to disk.
type sessionStore struct {
	mu    sync.RWMutex
	store map[string]string
}

func newSessionStore() CredentialStore {
	return &sessionStore{store: make(map[string]string)}
}

func (s *sessionStore) Save(_ context.Context, id, secret string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[id] = secret
	return nil
}

func (s *sessionStore) Load(_ context.Context, id string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.store[id]
	if !ok {
		return "", ErrNotFound
	}
	return v, nil
}

func (s *sessionStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.store[id]; !ok {
		return ErrNotFound
	}
	delete(s.store, id)
	return nil
}

func (s *sessionStore) IsSessionOnly() bool { return true }
