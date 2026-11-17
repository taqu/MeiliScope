// Package platform provides OS-specific integrations including secure credential storage.
package platform

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

// ErrNotFound is returned when a requested credential does not exist.
var ErrNotFound = errors.New("credential not found")

// ErrUnavailable is returned when the secure credential backend is not accessible.
var ErrUnavailable = errors.New("credential store unavailable")

// CredentialStore is the interface for secure API key storage.
// Implementations must never expose stored secrets through error messages or logs.
type CredentialStore interface {
	// Save stores a secret for the given id. Replaces any existing entry.
	Save(ctx context.Context, id, secret string) error
	// Load retrieves the secret for id. Returns ErrNotFound if absent.
	Load(ctx context.Context, id string) (string, error)
	// Delete removes the credential for id. Returns ErrNotFound if absent.
	Delete(ctx context.Context, id string) error
	// IsSessionOnly reports whether this store holds credentials only until exit.
	IsSessionOnly() bool
}

// NewCredentialStore returns the best available credential store for the current platform.
// If the secure persistent backend cannot be opened, a session-only in-memory store is
// returned and err is non-nil explaining why. The caller should inform the user when
// IsSessionOnly() is true.
func NewCredentialStore() (CredentialStore, error) {
	store, err := newPlatformStore()
	if err != nil {
		slog.Warn("secure credential storage unavailable; using session-only mode", "reason", err)
		return newSessionStore(), fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return store, nil
}
