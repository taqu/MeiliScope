package services

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/taqu/meiliscope/internal/config"
	"github.com/taqu/meiliscope/internal/platform"
)

// mockCredStore is a thread-safe in-memory credential store for tests.
type mockCredStore struct {
	mu          sync.RWMutex
	data        map[string]string
	saveErr     error
	loadErr     error
	deleteErr   error
	sessionOnly bool
}

func newMockCreds() *mockCredStore {
	return &mockCredStore{data: make(map[string]string)}
}

func (m *mockCredStore) Save(_ context.Context, id, secret string) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[id] = secret
	return nil
}

func (m *mockCredStore) Load(_ context.Context, id string) (string, error) {
	if m.loadErr != nil {
		return "", m.loadErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.data[id]
	if !ok {
		return "", platform.ErrNotFound
	}
	return v, nil
}

func (m *mockCredStore) Delete(_ context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[id]; !ok {
		return platform.ErrNotFound
	}
	delete(m.data, id)
	return nil
}

func (m *mockCredStore) IsSessionOnly() bool { return m.sessionOnly }

// newTestManager creates a config.Manager backed by a temporary directory.
func newTestManager(t *testing.T) *config.Manager {
	t.Helper()
	mgr, err := config.NewManagerAt(t.TempDir())
	if err != nil {
		t.Fatalf("NewManagerAt: %v", err)
	}
	if err := mgr.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	return mgr
}

func newTestConnSvc(t *testing.T) (*ConnectionService, *mockCredStore) {
	t.Helper()
	creds := newMockCreds()
	return NewConnectionService(newTestManager(t), creds), creds
}

// --- ListConnections ---

func TestConnectionService_ListConnections_Empty(t *testing.T) {
	svc, _ := newTestConnSvc(t)
	got := svc.ListConnections()
	if len(got) != 0 {
		t.Errorf("expected empty list, got %d", len(got))
	}
}

// --- CreateConnection ---

func TestConnectionService_CreateConnection(t *testing.T) {
	svc, creds := newTestConnSvc(t)
	ctx := context.Background()

	sum, err := svc.CreateConnection(ctx, SaveConnectionRequest{
		Name:   "Local",
		URL:    "http://localhost:7700",
		APIKey: "my-key",
	})
	if err != nil {
		t.Fatalf("CreateConnection: %v", err)
	}
	if sum.ID == "" {
		t.Error("expected non-empty ID")
	}
	if sum.Name != "Local" {
		t.Errorf("Name: got %q, want %q", sum.Name, "Local")
	}
	if sum.State != StateDisconnected {
		t.Errorf("State: got %q, want disconnected", sum.State)
	}

	// Credential stored under the generated ID.
	stored, err := creds.Load(ctx, sum.ID)
	if err != nil {
		t.Fatalf("credential Load: %v", err)
	}
	if stored != "my-key" {
		t.Errorf("stored credential: got %q, want %q", stored, "my-key")
	}

	// Profile appears in listing.
	list := svc.ListConnections()
	if len(list) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(list))
	}
}

func TestConnectionService_CreateConnection_NoAPIKey(t *testing.T) {
	svc, creds := newTestConnSvc(t)
	ctx := context.Background()

	sum, err := svc.CreateConnection(ctx, SaveConnectionRequest{
		Name: "No Key",
		URL:  "http://localhost:7700",
	})
	if err != nil {
		t.Fatalf("CreateConnection: %v", err)
	}
	// No credential should be stored.
	_, loadErr := creds.Load(ctx, sum.ID)
	if !errors.Is(loadErr, platform.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", loadErr)
	}
}

func TestConnectionService_CreateConnection_InvalidName(t *testing.T) {
	svc, _ := newTestConnSvc(t)
	_, err := svc.CreateConnection(context.Background(), SaveConnectionRequest{
		URL: "http://localhost:7700",
	})
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestConnectionService_CreateConnection_InvalidURL(t *testing.T) {
	svc, _ := newTestConnSvc(t)
	_, err := svc.CreateConnection(context.Background(), SaveConnectionRequest{
		Name: "Bad URL",
		URL:  "ftp://localhost:7700",
	})
	if err == nil {
		t.Error("expected error for invalid URL scheme")
	}
}

func TestConnectionService_CreateConnection_CredSaveFailure(t *testing.T) {
	svc, creds := newTestConnSvc(t)
	creds.saveErr = errors.New("keyring unavailable")

	_, err := svc.CreateConnection(context.Background(), SaveConnectionRequest{
		Name:   "Test",
		URL:    "http://localhost:7700",
		APIKey: "key",
	})
	if err == nil {
		t.Error("expected error when credential save fails")
	}
	// Profile must not be persisted.
	if len(svc.ListConnections()) != 0 {
		t.Error("profile should not be saved when credential save fails")
	}
}

// --- UpdateConnection ---

func TestConnectionService_UpdateConnection(t *testing.T) {
	svc, creds := newTestConnSvc(t)
	ctx := context.Background()

	sum, _ := svc.CreateConnection(ctx, SaveConnectionRequest{
		Name: "Old Name", URL: "http://localhost:7700", APIKey: "old-key",
	})

	updated, err := svc.UpdateConnection(ctx, sum.ID, SaveConnectionRequest{
		Name:   "New Name",
		URL:    "http://localhost:7700",
		APIKey: "new-key",
	})
	if err != nil {
		t.Fatalf("UpdateConnection: %v", err)
	}
	if updated.Name != "New Name" {
		t.Errorf("Name: got %q, want %q", updated.Name, "New Name")
	}

	stored, _ := creds.Load(ctx, sum.ID)
	if stored != "new-key" {
		t.Errorf("credential: got %q, want new-key", stored)
	}
}

func TestConnectionService_UpdateConnection_KeepExistingKey(t *testing.T) {
	svc, creds := newTestConnSvc(t)
	ctx := context.Background()

	sum, _ := svc.CreateConnection(ctx, SaveConnectionRequest{
		Name: "Test", URL: "http://localhost:7700", APIKey: "original-key",
	})

	// Update without supplying a new APIKey.
	_, err := svc.UpdateConnection(ctx, sum.ID, SaveConnectionRequest{
		Name: "Updated Name",
		URL:  "http://localhost:7700",
		// APIKey intentionally empty
	})
	if err != nil {
		t.Fatalf("UpdateConnection: %v", err)
	}

	// Original key must be intact.
	stored, _ := creds.Load(ctx, sum.ID)
	if stored != "original-key" {
		t.Errorf("credential changed unexpectedly: got %q", stored)
	}
}

func TestConnectionService_UpdateConnection_NotFound(t *testing.T) {
	svc, _ := newTestConnSvc(t)
	_, err := svc.UpdateConnection(context.Background(), "nonexistent-id", SaveConnectionRequest{
		Name: "X", URL: "http://localhost:7700",
	})
	if err == nil {
		t.Error("expected error for unknown profile ID")
	}
}

// --- DeleteConnection ---

func TestConnectionService_DeleteConnection(t *testing.T) {
	svc, creds := newTestConnSvc(t)
	ctx := context.Background()

	sum, _ := svc.CreateConnection(ctx, SaveConnectionRequest{
		Name: "To Delete", URL: "http://localhost:7700", APIKey: "del-key",
	})

	if err := svc.DeleteConnection(ctx, sum.ID); err != nil {
		t.Fatalf("DeleteConnection: %v", err)
	}

	if len(svc.ListConnections()) != 0 {
		t.Error("connection should be removed from list")
	}

	_, credErr := creds.Load(ctx, sum.ID)
	if !errors.Is(credErr, platform.ErrNotFound) {
		t.Errorf("credential should be deleted, got: %v", credErr)
	}
}

func TestConnectionService_DeleteConnection_NotFound(t *testing.T) {
	svc, _ := newTestConnSvc(t)
	err := svc.DeleteConnection(context.Background(), "ghost")
	if err == nil {
		t.Error("expected error for unknown profile")
	}
}

// --- GetActiveConnection / IsCredentialStoreSecure ---

func TestConnectionService_GetActiveConnection_Nil(t *testing.T) {
	svc, _ := newTestConnSvc(t)
	if got := svc.GetActiveConnection(); got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestConnectionService_IsCredentialStoreSecure_Session(t *testing.T) {
	creds := &mockCredStore{data: make(map[string]string), sessionOnly: true}
	svc := NewConnectionService(newTestManager(t), creds)
	if svc.IsCredentialStoreSecure() {
		t.Error("expected false for session-only store")
	}
}

func TestConnectionService_IsCredentialStoreSecure_Persistent(t *testing.T) {
	creds := &mockCredStore{data: make(map[string]string), sessionOnly: false}
	svc := NewConnectionService(newTestManager(t), creds)
	if !svc.IsCredentialStoreSecure() {
		t.Error("expected true for persistent store")
	}
}

// --- NormaliseCredMode ---

func TestNormaliseCredMode(t *testing.T) {
	cases := []struct {
		requested   string
		sessionOnly bool
		want        string
	}{
		{"keyring", false, "keyring"},
		{"session", false, "session"},
		{"", false, "keyring"},
		{"keyring", true, "session"}, // forced session when store is session-only
		{"session", true, "session"},
	}
	for _, tc := range cases {
		got := normaliseCredMode(tc.requested, tc.sessionOnly)
		if got != tc.want {
			t.Errorf("normaliseCredMode(%q, %v) = %q, want %q", tc.requested, tc.sessionOnly, got, tc.want)
		}
	}
}

// --- Disconnect ---

func TestConnectionService_Disconnect_WhenInactive(t *testing.T) {
	svc, _ := newTestConnSvc(t)
	svc.Disconnect() // must not panic
	if svc.GetActiveConnection() != nil {
		t.Error("expected nil active connection after disconnect")
	}
}

// --- TestNewConnection ---

func TestConnectionService_TestNewConnection_Reachable(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"available"}`))
	})
	mux.HandleFunc("/version", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"pkgVersion":"1.7.0","commitSha":"abc","commitDate":"2024-01-01"}`))
	})
	mux.HandleFunc("/indexes", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[],"offset":0,"limit":1,"total":0}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	svc, _ := newTestConnSvc(t)
	res, err := svc.TestNewConnection(context.Background(), srv.URL, "test-key")
	if err != nil {
		t.Fatalf("TestNewConnection: %v", err)
	}
	if !res.ServerReachable {
		t.Error("expected ServerReachable=true")
	}
	if !res.CanListIndexes {
		t.Error("expected CanListIndexes=true")
	}
	if res.State != StateConnected {
		t.Errorf("State: got %q, want connected", res.State)
	}
	if res.ServerVersion != "1.7.0" {
		t.Errorf("ServerVersion: got %q, want 1.7.0", res.ServerVersion)
	}
}

func TestConnectionService_TestNewConnection_AuthRequired(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"available"}`))
	})
	mux.HandleFunc("/version", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"pkgVersion":"1.7.0","commitSha":"abc","commitDate":"2024-01-01"}`))
	})
	mux.HandleFunc("/indexes", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		body, _ := json.Marshal(map[string]string{
			"message": "authentication required", "code": "missing_authorization_header",
			"type": "auth_error", "link": "",
		})
		_, _ = w.Write(body)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	svc, _ := newTestConnSvc(t)
	res, err := svc.TestNewConnection(context.Background(), srv.URL, "")
	if err != nil {
		t.Fatalf("TestNewConnection: %v", err)
	}
	if res.State != StateAuthRequired {
		t.Errorf("State: got %q, want auth_required", res.State)
	}
	if res.Authenticated {
		t.Error("expected Authenticated=false")
	}
}

func TestConnectionService_TestNewConnection_LimitedAccess(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"available"}`))
	})
	mux.HandleFunc("/version", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"pkgVersion":"1.7.0","commitSha":"abc","commitDate":"2024-01-01"}`))
	})
	mux.HandleFunc("/indexes", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		body, _ := json.Marshal(map[string]string{
			"message": "access denied", "code": "invalid_api_key",
			"type": "auth_error", "link": "",
		})
		_, _ = w.Write(body)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	svc, _ := newTestConnSvc(t)
	res, err := svc.TestNewConnection(context.Background(), srv.URL, "search-only-key")
	if err != nil {
		t.Fatalf("TestNewConnection: %v", err)
	}
	if res.State != StateLimitedAccess {
		t.Errorf("State: got %q, want limited_access", res.State)
	}
	if !res.Authenticated {
		t.Error("expected Authenticated=true for search-only key")
	}
}

func TestConnectionService_TestNewConnection_ServerDown(t *testing.T) {
	svc, _ := newTestConnSvc(t)
	res, err := svc.TestNewConnection(context.Background(), "http://127.0.0.1:19999", "")
	if err != nil {
		t.Fatalf("TestNewConnection: %v", err)
	}
	if res.ServerReachable {
		t.Error("expected ServerReachable=false for refused connection")
	}
	if res.State != StateError {
		t.Errorf("State: got %q, want error", res.State)
	}
}

func TestConnectionService_TestNewConnection_InvalidURL(t *testing.T) {
	svc, _ := newTestConnSvc(t)
	res, err := svc.TestNewConnection(context.Background(), "ftp://bad.url", "")
	if err != nil {
		t.Fatalf("TestNewConnection: %v", err)
	}
	if res.State != StateError {
		t.Errorf("State: got %q, want error", res.State)
	}
}

// --- TestConnection (existing profile) ---

func TestConnectionService_TestConnection_ProfileNotFound(t *testing.T) {
	svc, _ := newTestConnSvc(t)
	_, err := svc.TestConnection(context.Background(), "nonexistent-id", "")
	if err == nil {
		t.Error("expected error for unknown profile ID")
	}
}

// --- generateID uniqueness ---

func TestGenerateID_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for range 100 {
		id, err := generateID()
		if err != nil {
			t.Fatalf("generateID: %v", err)
		}
		if seen[id] {
			t.Fatalf("duplicate ID: %s", id)
		}
		seen[id] = true
		if id == "" {
			t.Fatal("empty ID")
		}
	}
}
