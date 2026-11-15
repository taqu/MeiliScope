// Package platform provides Windows-specific operating system integrations.
package platform

import "errors"

// ErrNotFound is returned when a credential does not exist.
var ErrNotFound = errors.New("credential not found")

// CredentialStore is the interface for secure credential storage.
// The Windows implementation uses the Windows Credential Manager.
// A mock implementation is used in tests.
type CredentialStore interface {
	// Set stores a credential. Replaces any existing entry with the same key.
	Set(key, value string) error
	// Get retrieves a credential by key. Returns ErrNotFound if absent.
	Get(key string) (string, error)
	// Delete removes a credential by key. Returns ErrNotFound if absent.
	Delete(key string) error
}
