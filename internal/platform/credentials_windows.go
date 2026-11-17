//go:build windows

package platform

import (
	"context"
	"errors"
	"fmt"

	keyring "github.com/zalando/go-keyring"
)

const keyringSvc = "MeiliScope"

// windowsStore implements CredentialStore using Windows Credential Manager
// through the go-keyring library.
type windowsStore struct{}

func newPlatformStore() (CredentialStore, error) {
	// Probe the credential store with a no-op to verify it is accessible.
	// go-keyring on Windows uses DPAPI-protected WinCred entries; this rarely fails
	// but can when running in a restricted service account context.
	_, err := keyring.Get(keyringSvc, "__probe__")
	if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return nil, fmt.Errorf("Windows Credential Manager unavailable: %w", err)
	}
	return &windowsStore{}, nil
}

func (s *windowsStore) Save(_ context.Context, id, secret string) error {
	if err := keyring.Set(keyringSvc, id, secret); err != nil {
		return fmt.Errorf("credential store: save failed: %w", err)
	}
	return nil
}

func (s *windowsStore) Load(_ context.Context, id string) (string, error) {
	v, err := keyring.Get(keyringSvc, id)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("credential store: load failed: %w", err)
	}
	return v, nil
}

func (s *windowsStore) Delete(_ context.Context, id string) error {
	err := keyring.Delete(keyringSvc, id)
	if errors.Is(err, keyring.ErrNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("credential store: delete failed: %w", err)
	}
	return nil
}

func (s *windowsStore) IsSessionOnly() bool { return false }
