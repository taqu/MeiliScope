package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
)

const (
	appDirName = "MeiliScope"
	configFile = "config.json"
	schemaV1   = 1
)

// ConnectionProfile is a saved Meilisearch connection.
// It never contains the API key; credentials are managed by the credential store.
type ConnectionProfile struct {
	// ID is a stable, unique identifier used as the credential key.
	ID string `json:"id"`
	// Name is the user-visible display name.
	Name string `json:"name"`
	// URL is the base URL of the Meilisearch server.
	URL string `json:"url"`
	// CredentialMode is "keyring" for persistent storage or "session" for in-memory only.
	CredentialMode string `json:"credentialMode"`
}

// Config holds all non-secret application settings.
type Config struct {
	Version            int                 `json:"version"`
	WindowWidth        int                 `json:"windowWidth"`
	WindowHeight       int                 `json:"windowHeight"`
	ActiveView         string              `json:"activeView"`
	Connections        []ConnectionProfile `json:"connections,omitempty"`
	ActiveConnectionID string              `json:"activeConnectionID,omitempty"`
}

func defaultConfig() Config {
	return Config{
		Version:      schemaV1,
		WindowWidth:  1280,
		WindowHeight: 800,
		ActiveView:   "connections",
	}
}

// Manager loads and saves Config atomically.
type Manager struct {
	dir  string
	path string
	cfg  Config
}

// NewManager creates a Manager rooted at %AppData%\MeiliScope.
func NewManager() (*Manager, error) {
	appData, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("config: resolve AppData: %w", err)
	}
	dir := filepath.Join(appData, appDirName)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("config: create dir %s: %w", dir, err)
	}
	return &Manager{
		dir:  dir,
		path: filepath.Join(dir, configFile),
	}, nil
}

// NewManagerAt creates a Manager rooted at the given directory.
// Intended for testing; prefer NewManager in production code.
func NewManagerAt(dir string) (*Manager, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("config: create dir %s: %w", dir, err)
	}
	return &Manager{
		dir:  dir,
		path: filepath.Join(dir, configFile),
	}, nil
}

// Dir returns the application configuration directory.
func (m *Manager) Dir() string { return m.dir }

// Load reads configuration from disk. If the file is absent or corrupted,
// defaults are used and a warning is logged; the bad file is renamed for
// diagnostic purposes rather than silently deleted.
func (m *Manager) Load() error {
	data, err := os.ReadFile(m.path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			m.cfg = defaultConfig()
			return nil
		}
		return fmt.Errorf("config: read %s: %w", m.path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		slog.Warn("config: corrupted; using defaults", "path", m.path, "err", err)
		backup := m.path + ".corrupted"
		_ = os.Rename(m.path, backup)
		m.cfg = defaultConfig()
		return nil
	}

	if cfg.Version == 0 {
		cfg.Version = schemaV1
	}
	m.cfg = cfg
	return nil
}

// Save writes configuration atomically.
func (m *Manager) Save() error {
	data, err := json.MarshalIndent(m.cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("config: marshal: %w", err)
	}

	tmp := m.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return fmt.Errorf("config: write tmp %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, m.path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("config: rename to %s: %w", m.path, err)
	}
	return nil
}

// Get returns a copy of the current configuration.
func (m *Manager) Get() Config { return m.cfg }

// Update replaces the in-memory configuration. Call Save to persist.
func (m *Manager) Update(cfg Config) { m.cfg = cfg }
