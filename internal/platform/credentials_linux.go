//go:build linux

package platform

import (
	"context"
	"errors"
	"fmt"

	keyring "github.com/zalando/go-keyring"
)

const keyringSvc = "MeiliScope"

// linuxStore implements CredentialStore using the desktop Secret Service
// (GNOME Keyring, KWallet via Secret Service compatibility) through go-keyring.
type linuxStore struct{}

func newPlatformStore() (CredentialStore, error) {
	// Probe the Secret Service to verify it is reachable.
	// This fails if the D-Bus session is absent or the keyring is locked.
	_, err := keyring.Get(keyringSvc, "__probe__")
	if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return nil, fmt.Errorf("Secret Service unavailable: %w", err)
	}
	return &linuxStore{}, nil
}

func (s *linuxStore) Save(_ context.Context, id, secret string) error {
	if err := keyring.Set(keyringSvc, id, secret); err != nil {
		return fmt.Errorf("credential store: save failed: %w", err)
	}
	return nil
}

func (s *linuxStore) Load(_ context.Context, id string) (string, error) {
	v, err := keyring.Get(keyringSvc, id)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("credential store: load failed: %w", err)
	}
	return v, nil
}

func (s *linuxStore) Delete(_ context.Context, id string) error {
	err := keyring.Delete(keyringSvc, id)
	if errors.Is(err, keyring.ErrNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("credential store: delete failed: %w", err)
	}
	return nil
}

func (s *linuxStore) IsSessionOnly() bool { return false }
