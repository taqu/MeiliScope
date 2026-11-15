package platform

// WindowsCredentialStore implements CredentialStore using Windows Credential Manager.
// It is a placeholder pending Phase 1 implementation.
// The interface and planned approach are documented in docs/architecture/overview.md §2.5.
type WindowsCredentialStore struct{}

// NewCredentialStore returns the Windows Credential Manager implementation.
// In Phase 1 this will be replaced with a real implementation using
// github.com/danieljoos/wincred.
func NewCredentialStore() CredentialStore {
	return &WindowsCredentialStore{}
}

func (w *WindowsCredentialStore) Set(_, _ string) error {
	return errNotImplemented
}

func (w *WindowsCredentialStore) Get(_ string) (string, error) {
	return "", errNotImplemented
}

func (w *WindowsCredentialStore) Delete(_ string) error {
	return errNotImplemented
}
