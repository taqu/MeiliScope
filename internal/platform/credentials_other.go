//go:build !windows && !linux

package platform

import "fmt"

// newPlatformStore returns ErrUnavailable on unsupported platforms.
// The caller falls back to session-only mode.
func newPlatformStore() (CredentialStore, error) {
	return nil, fmt.Errorf("no secure credential backend for this platform")
}
