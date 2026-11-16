package services

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/taqu/meiliscope/internal/config"
	"github.com/taqu/meiliscope/internal/meilisearch"
	"github.com/taqu/meiliscope/internal/platform"
)

// ConnectionState describes the current state of a connection.
type ConnectionState string

const (
	StateDisconnected  ConnectionState = "disconnected"
	StateConnecting    ConnectionState = "connecting"
	StateConnected     ConnectionState = "connected"
	StateLimitedAccess ConnectionState = "limited_access"
	StateAuthRequired  ConnectionState = "auth_required"
	StateError         ConnectionState = "error"
)

// ConnectionSummary is the frontend-safe view of a connection.
// It never includes the API key or any credential material.
type ConnectionSummary struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	URL            string          `json:"url"`
	CredentialMode string          `json:"credentialMode"`
	State          ConnectionState `json:"state"`
}

// ActiveConnectionInfo is returned by GetActiveConnection.
type ActiveConnectionInfo struct {
	ConnectionSummary
	IsCredentialStoreSecure bool `json:"isCredentialStoreSecure"`
}

// SaveConnectionRequest is used to create or update a connection profile.
type SaveConnectionRequest struct {
	Name           string `json:"name"`
	URL            string `json:"url"`
	APIKey         string `json:"apiKey"`         // empty = keep existing stored key
	CredentialMode string `json:"credentialMode"` // "keyring" | "session"
}

// ConnectionTestResult reports the outcome of a connection test.
type ConnectionTestResult struct {
	ServerReachable bool            `json:"serverReachable"`
	Authenticated   bool            `json:"authenticated"`
	CanListIndexes  bool            `json:"canListIndexes"`
	State           ConnectionState `json:"state"`
	Message         string          `json:"message"`
	ServerVersion   string          `json:"serverVersion,omitempty"`
}

// activeConn holds the in-flight Meilisearch client and connection metadata.
type activeConn struct {
	id         string
	client     *meilisearch.Client
	state      ConnectionState
	generation int64
}

// ConnectionService manages connection profiles and the active Meilisearch client.
// It is safe for concurrent use.
type ConnectionService struct {
	cfgManager *config.Manager
	creds      platform.CredentialStore

	mu         sync.RWMutex
	active     *activeConn
	generation atomic.Int64
}

// NewConnectionService creates a ConnectionService.
// The credential store must already be initialised by the caller.
func NewConnectionService(cfgManager *config.Manager, creds platform.CredentialStore) *ConnectionService {
	return &ConnectionService{
		cfgManager: cfgManager,
		creds:      creds,
	}
}

// --- Wails-exported methods ---

// ListConnections returns all saved connection profiles without credentials.
func (s *ConnectionService) ListConnections() []ConnectionSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cfg := s.cfgManager.Get()
	out := make([]ConnectionSummary, 0, len(cfg.Connections))
	for _, p := range cfg.Connections {
		cs := ConnectionSummary{
			ID:             p.ID,
			Name:           p.Name,
			URL:            p.URL,
			CredentialMode: p.CredentialMode,
			State:          StateDisconnected,
		}
		if s.active != nil && s.active.id == p.ID {
			cs.State = s.active.state
		}
		out = append(out, cs)
	}
	return out
}

// GetActiveConnection returns the currently active connection, or nil if disconnected.
func (s *ConnectionService) GetActiveConnection() *ActiveConnectionInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.active == nil {
		return nil
	}

	cfg := s.cfgManager.Get()
	for _, p := range cfg.Connections {
		if p.ID == s.active.id {
			return &ActiveConnectionInfo{
				ConnectionSummary: ConnectionSummary{
					ID:             p.ID,
					Name:           p.Name,
					URL:            p.URL,
					CredentialMode: p.CredentialMode,
					State:          s.active.state,
				},
				IsCredentialStoreSecure: !s.creds.IsSessionOnly(),
			}
		}
	}
	return nil
}

// IsCredentialStoreSecure reports whether the credential store uses persistent OS storage.
func (s *ConnectionService) IsCredentialStoreSecure() bool {
	return !s.creds.IsSessionOnly()
}

// CreateConnection adds a new connection profile.
// If APIKey is non-empty, it is stored in the credential store under the new profile ID.
func (s *ConnectionService) CreateConnection(ctx context.Context, req SaveConnectionRequest) (ConnectionSummary, error) {
	if err := validateSaveRequest(req); err != nil {
		return ConnectionSummary{}, err
	}

	id, err := generateID()
	if err != nil {
		return ConnectionSummary{}, fmt.Errorf("connection: generate ID: %w", err)
	}

	mode := normaliseCredMode(req.CredentialMode, s.creds.IsSessionOnly())

	// 1. Store credential first (before updating config) to avoid orphaned profiles.
	if req.APIKey != "" {
		if err := s.creds.Save(ctx, id, req.APIKey); err != nil {
			return ConnectionSummary{}, fmt.Errorf("connection: store credential: %w", err)
		}
	}

	profile := config.ConnectionProfile{
		ID:             id,
		Name:           req.Name,
		URL:            req.URL,
		CredentialMode: mode,
	}

	// 2. Persist the profile. On failure, attempt to clean up the stored credential.
	if err := s.addProfile(profile); err != nil {
		if req.APIKey != "" {
			if delErr := s.creds.Delete(ctx, id); delErr != nil && !errors.Is(delErr, platform.ErrNotFound) {
				slog.Warn("connection: credential cleanup after failed profile save", "id", id, "err", delErr)
			}
		}
		return ConnectionSummary{}, fmt.Errorf("connection: save profile: %w", err)
	}

	slog.Info("connection created", "id", id, "name", req.Name)
	return ConnectionSummary{
		ID:             id,
		Name:           req.Name,
		URL:            req.URL,
		CredentialMode: mode,
		State:          StateDisconnected,
	}, nil
}

// UpdateConnection edits an existing connection profile.
// If APIKey is non-empty, the stored credential is replaced.
// If APIKey is empty, the existing stored credential is kept unchanged.
func (s *ConnectionService) UpdateConnection(ctx context.Context, id string, req SaveConnectionRequest) (ConnectionSummary, error) {
	if err := validateSaveRequest(req); err != nil {
		return ConnectionSummary{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cfg := s.cfgManager.Get()
	idx := findProfile(cfg.Connections, id)
	if idx < 0 {
		return ConnectionSummary{}, fmt.Errorf("connection: profile %q not found", id)
	}

	mode := normaliseCredMode(req.CredentialMode, s.creds.IsSessionOnly())

	// Replace credential only when a new key is provided.
	if req.APIKey != "" {
		if err := s.creds.Save(ctx, id, req.APIKey); err != nil {
			return ConnectionSummary{}, fmt.Errorf("connection: update credential: %w", err)
		}
	}

	cfg.Connections[idx].Name = req.Name
	cfg.Connections[idx].URL = req.URL
	cfg.Connections[idx].CredentialMode = mode

	s.cfgManager.Update(cfg)
	if err := s.cfgManager.Save(); err != nil {
		return ConnectionSummary{}, fmt.Errorf("connection: persist update: %w", err)
	}

	// If this connection is currently active, update its in-memory URL reference.
	if s.active != nil && s.active.id == id {
		// The client will be re-created on the next SelectConnection call.
		s.active.state = StateDisconnected
		s.active = nil
	}

	slog.Info("connection updated", "id", id, "name", req.Name)
	return ConnectionSummary{
		ID:             id,
		Name:           req.Name,
		URL:            req.URL,
		CredentialMode: mode,
		State:          StateDisconnected,
	}, nil
}

// DeleteConnection removes a connection profile and its stored credential.
// If credential deletion fails, the error is reported but the profile is still removed.
func (s *ConnectionService) DeleteConnection(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg := s.cfgManager.Get()
	idx := findProfile(cfg.Connections, id)
	if idx < 0 {
		return fmt.Errorf("connection: profile %q not found", id)
	}

	// If this is the active connection, disconnect first.
	if s.active != nil && s.active.id == id {
		s.generation.Add(1)
		s.active = nil
	}

	// Remove profile from config.
	cfg.Connections = append(cfg.Connections[:idx], cfg.Connections[idx+1:]...)
	if cfg.ActiveConnectionID == id {
		cfg.ActiveConnectionID = ""
	}
	s.cfgManager.Update(cfg)
	if err := s.cfgManager.Save(); err != nil {
		return fmt.Errorf("connection: persist deletion: %w", err)
	}

	// Best-effort credential cleanup; report but do not block on failure.
	if err := s.creds.Delete(ctx, id); err != nil && !errors.Is(err, platform.ErrNotFound) {
		return fmt.Errorf("connection: profile deleted but credential cleanup failed: %w", err)
	}

	slog.Info("connection deleted", "id", id)
	return nil
}

// TestConnection tests reachability and authentication for a profile.
// It performs only read-only requests and never modifies Meilisearch data.
// If apiKey is non-empty it is used instead of the stored credential (useful before saving).
func (s *ConnectionService) TestConnection(ctx context.Context, id string, apiKey string) (ConnectionTestResult, error) {
	cfg := s.cfgManager.Get()
	idx := findProfile(cfg.Connections, id)
	if idx < 0 {
		return ConnectionTestResult{}, fmt.Errorf("connection: profile %q not found", id)
	}
	profile := cfg.Connections[idx]

	key := apiKey
	if key == "" {
		var err error
		key, err = s.creds.Load(ctx, id)
		if err != nil && !errors.Is(err, platform.ErrNotFound) {
			return ConnectionTestResult{}, fmt.Errorf("connection: load credential: %w", err)
		}
	}

	return testConnection(ctx, profile.URL, key)
}

// TestNewConnection tests a connection that has not yet been saved.
// url and apiKey are provided directly.
func (s *ConnectionService) TestNewConnection(ctx context.Context, testURL, apiKey string) (ConnectionTestResult, error) {
	return testConnection(ctx, testURL, apiKey)
}

// SelectConnection activates a saved connection by ID.
// In-flight requests on the previous connection are invalidated by incrementing the generation counter.
func (s *ConnectionService) SelectConnection(ctx context.Context, id string) (ActiveConnectionInfo, error) {
	cfg := s.cfgManager.Get()
	idx := findProfile(cfg.Connections, id)
	if idx < 0 {
		return ActiveConnectionInfo{}, fmt.Errorf("connection: profile %q not found", id)
	}
	profile := cfg.Connections[idx]

	apiKey, err := s.creds.Load(ctx, id)
	if err != nil && !errors.Is(err, platform.ErrNotFound) {
		return ActiveConnectionInfo{}, fmt.Errorf("connection: load credential: %w", err)
	}

	client, err := meilisearch.New(meilisearch.Config{
		BaseURL: profile.URL,
		APIKey:  apiKey,
	})
	if err != nil {
		return ActiveConnectionInfo{}, fmt.Errorf("connection: build client: %w", err)
	}

	gen := s.generation.Add(1)

	s.mu.Lock()
	s.active = &activeConn{
		id:         id,
		client:     client,
		state:      StateConnecting,
		generation: gen,
	}
	s.mu.Unlock()

	// Persist the active connection ID for restart recovery.
	cfg.ActiveConnectionID = id
	s.cfgManager.Update(cfg)
	if err := s.cfgManager.Save(); err != nil {
		slog.Warn("connection: failed to persist active connection", "err", err)
	}

	// Probe the connection to establish the final state.
	result, _ := testConnection(ctx, profile.URL, apiKey)

	s.mu.Lock()
	// Only update if generation is still current (no concurrent switch).
	if s.active != nil && s.active.generation == gen {
		s.active.state = result.State
	}
	s.mu.Unlock()

	slog.Info("connection selected", "id", id, "state", result.State)
	return ActiveConnectionInfo{
		ConnectionSummary: ConnectionSummary{
			ID:             profile.ID,
			Name:           profile.Name,
			URL:            profile.URL,
			CredentialMode: profile.CredentialMode,
			State:          result.State,
		},
		IsCredentialStoreSecure: !s.creds.IsSessionOnly(),
	}, nil
}

// Disconnect clears the active connection.
func (s *ConnectionService) Disconnect() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.generation.Add(1)
	s.active = nil
}

// activeClient returns the Meilisearch client and generation counter for the
// current active connection. Returns (nil, 0) when disconnected.
// Called by IndexService; not exposed to the frontend.
func (s *ConnectionService) activeClient() (*meilisearch.Client, int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.active == nil {
		return nil, 0
	}
	return s.active.client, s.active.generation
}

// currentGeneration returns the current connection generation counter.
func (s *ConnectionService) currentGeneration() int64 {
	return s.generation.Load()
}

// --- helpers ---

func validateSaveRequest(req SaveConnectionRequest) error {
	if req.Name == "" {
		return fmt.Errorf("connection: name must not be empty")
	}
	if req.URL == "" {
		return fmt.Errorf("connection: URL must not be empty")
	}
	// Validate URL by attempting to build a client (dry-run).
	_, err := meilisearch.New(meilisearch.Config{BaseURL: req.URL})
	if err != nil {
		return fmt.Errorf("connection: invalid URL: %w", err)
	}
	return nil
}

func normaliseCredMode(requested string, sessionOnly bool) string {
	if sessionOnly || requested == "session" {
		return "session"
	}
	return "keyring"
}

func (s *ConnectionService) addProfile(p config.ConnectionProfile) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg := s.cfgManager.Get()
	cfg.Connections = append(cfg.Connections, p)
	s.cfgManager.Update(cfg)
	return s.cfgManager.Save()
}

func findProfile(profiles []config.ConnectionProfile, id string) int {
	for i, p := range profiles {
		if p.ID == id {
			return i
		}
	}
	return -1
}

// generateID produces a random UUID v4 string without external dependencies.
func generateID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant RFC 4122
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}

// testConnection performs read-only probes to determine reachability and auth state.
func testConnection(ctx context.Context, baseURL, apiKey string) (ConnectionTestResult, error) {
	client, err := meilisearch.New(meilisearch.Config{
		BaseURL: baseURL,
		APIKey:  apiKey,
	})
	if err != nil {
		return ConnectionTestResult{
			State:   StateError,
			Message: "invalid server URL",
		}, nil
	}

	res := ConnectionTestResult{}

	// 1. Health check — tests basic reachability.
	_, err = client.Health(ctx)
	if err != nil {
		if meilisearch.IsCanceled(err) {
			return res, nil
		}
		res.State = StateError
		res.Message = friendlyTransportError(err)
		return res, nil
	}
	res.ServerReachable = true

	// 2. Version — read-only, available without auth on most deployments.
	version, err := client.Version(ctx)
	if err == nil && version != nil {
		res.ServerVersion = version.PkgVersion
	}

	// 3. List indexes — tests that the API key can access indexes.
	_, err = client.ListIndexes(ctx, 0, 1)
	if err != nil {
		switch {
		case meilisearch.IsCanceled(err):
			return res, nil
		case meilisearch.IsAuth(err):
			res.State = StateAuthRequired
			res.Message = "authentication required — please provide an API key"
			res.Authenticated = false
		case meilisearch.IsForbidden(err):
			res.Authenticated = true
			res.State = StateLimitedAccess
			res.Message = "API key does not have permission to list indexes"
		default:
			res.Authenticated = apiKey != ""
			res.State = StateError
			res.Message = "failed to list indexes: " + sanitiseErrorMessage(err)
		}
		return res, nil
	}

	res.Authenticated = true
	res.CanListIndexes = true
	res.State = StateConnected
	res.Message = "connection successful"
	return res, nil
}

// friendlyTransportError returns a user-friendly message for transport errors,
// ensuring no credentials are included.
func friendlyTransportError(err error) string {
	switch {
	case meilisearch.IsTimeout(err):
		return "connection timed out — check the server address and network"
	case meilisearch.IsCanceled(err):
		return "request canceled"
	default:
		return sanitiseErrorMessage(err)
	}
}

// sanitiseErrorMessage extracts a safe message from err (no credentials).
func sanitiseErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	var apiErr *meilisearch.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Msg
	}
	return "unexpected error"
}
