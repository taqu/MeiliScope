package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/taqu/meiliscope/internal/config"
)

const (
	savedQueriesFileName = "saved_queries.json"
	maxSavedQueries      = 500
	savedQueriesVersion  = 1
)

// SavedQuery is a named, persisted search query attached to a connection and index.
type SavedQuery struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	ConnectionID string `json:"connectionId"`
	IndexUID     string `json:"indexUid"`
	QueryJSON    string `json:"queryJson"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

// CreateSavedQueryRequest carries the fields needed to create a new saved query.
type CreateSavedQueryRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	ConnectionID string `json:"connectionId"`
	IndexUID     string `json:"indexUid"`
	QueryJSON    string `json:"queryJson"`
}

// UpdateSavedQueryRequest carries the fields that may be updated on an existing
// saved query. Description is a pointer so that nil means "do not change".
type UpdateSavedQueryRequest struct {
	Name         string  `json:"name"`
	Description  *string `json:"description,omitempty"`
	ConnectionID string  `json:"connectionId"`
	IndexUID     string  `json:"indexUid"`
	QueryJSON    string  `json:"queryJson"`
}

// savedQueriesFile is the on-disk envelope written to saved_queries.json.
type savedQueriesFile struct {
	Version int          `json:"version"`
	Items   []SavedQuery `json:"items"`
}

// SavedQueryService persists named search queries to saved_queries.json inside
// the application config directory. It is safe for concurrent use.
type SavedQueryService struct {
	cfgManager *config.Manager

	mu     sync.Mutex
	loaded bool
	data   savedQueriesFile
}

// NewSavedQueryService creates a SavedQueryService backed by cfgManager.
func NewSavedQueryService(cfgManager *config.Manager) *SavedQueryService {
	return &SavedQueryService{
		cfgManager: cfgManager,
		data:       savedQueriesFile{Version: savedQueriesVersion},
	}
}

// --- Wails-exported methods ---

// CreateSavedQuery creates a new saved query. Name must be non-empty and
// QueryJSON must be a valid JSON object. Returns the newly created SavedQuery.
func (s *SavedQueryService) CreateSavedQuery(_ context.Context, req CreateSavedQueryRequest) (SavedQuery, error) {
	if req.Name == "" {
		return SavedQuery{}, fmt.Errorf("savedquery: name must not be empty")
	}
	if err := validateQueryJSON(req.QueryJSON); err != nil {
		return SavedQuery{}, err
	}

	id, err := generateID()
	if err != nil {
		return SavedQuery{}, fmt.Errorf("savedquery: generate ID: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	q := SavedQuery{
		ID:           id,
		Name:         req.Name,
		Description:  req.Description,
		ConnectionID: req.ConnectionID,
		IndexUID:     req.IndexUID,
		QueryJSON:    req.QueryJSON,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.ensureLoadedNoLock()

	// Prepend so the list is newest-first.
	s.data.Items = append([]SavedQuery{q}, s.data.Items...)
	if len(s.data.Items) > maxSavedQueries {
		s.data.Items = s.data.Items[:maxSavedQueries]
	}

	if err := s.saveNoLock(); err != nil {
		return SavedQuery{}, err
	}
	return q, nil
}

// ListSavedQueries returns all saved queries, newest first.
func (s *SavedQueryService) ListSavedQueries(_ context.Context) ([]SavedQuery, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ensureLoadedNoLock()

	out := make([]SavedQuery, len(s.data.Items))
	copy(out, s.data.Items)
	return out, nil
}

// GetSavedQuery returns a single saved query by ID.
func (s *SavedQueryService) GetSavedQuery(_ context.Context, id string) (SavedQuery, error) {
	if id == "" {
		return SavedQuery{}, fmt.Errorf("savedquery: id must not be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.ensureLoadedNoLock()

	for _, q := range s.data.Items {
		if q.ID == id {
			return q, nil
		}
	}
	return SavedQuery{}, fmt.Errorf("savedquery: %q not found", id)
}

// UpdateSavedQuery updates name, description, connectionID, indexUID, or queryJSON
// for the query identified by id. Fields absent from the request are preserved.
func (s *SavedQueryService) UpdateSavedQuery(_ context.Context, id string, req UpdateSavedQueryRequest) (SavedQuery, error) {
	if id == "" {
		return SavedQuery{}, fmt.Errorf("savedquery: id must not be empty")
	}
	if req.Name == "" {
		return SavedQuery{}, fmt.Errorf("savedquery: name must not be empty")
	}
	if err := validateQueryJSON(req.QueryJSON); err != nil {
		return SavedQuery{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.ensureLoadedNoLock()

	for i, q := range s.data.Items {
		if q.ID != id {
			continue
		}
		q.Name = req.Name
		if req.Description != nil {
			q.Description = *req.Description
		}
		q.ConnectionID = req.ConnectionID
		q.IndexUID = req.IndexUID
		q.QueryJSON = req.QueryJSON
		q.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
		s.data.Items[i] = q

		if err := s.saveNoLock(); err != nil {
			return SavedQuery{}, err
		}
		return q, nil
	}
	return SavedQuery{}, fmt.Errorf("savedquery: %q not found", id)
}

// DeleteSavedQuery removes a saved query by ID.
func (s *SavedQueryService) DeleteSavedQuery(_ context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("savedquery: id must not be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.ensureLoadedNoLock()

	for i, q := range s.data.Items {
		if q.ID == id {
			s.data.Items = append(s.data.Items[:i], s.data.Items[i+1:]...)
			return s.saveNoLock()
		}
	}
	return fmt.Errorf("savedquery: %q not found", id)
}

// DuplicateSavedQuery creates a copy of an existing saved query with " (copy)"
// appended to the name. Returns the newly created duplicate.
func (s *SavedQueryService) DuplicateSavedQuery(_ context.Context, id string) (SavedQuery, error) {
	if id == "" {
		return SavedQuery{}, fmt.Errorf("savedquery: id must not be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.ensureLoadedNoLock()

	var src SavedQuery
	found := false
	for _, q := range s.data.Items {
		if q.ID == id {
			src = q
			found = true
			break
		}
	}
	if !found {
		return SavedQuery{}, fmt.Errorf("savedquery: %q not found", id)
	}

	newID, err := generateID()
	if err != nil {
		return SavedQuery{}, fmt.Errorf("savedquery: generate ID: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	dup := SavedQuery{
		ID:           newID,
		Name:         src.Name + " (copy)",
		Description:  src.Description,
		ConnectionID: src.ConnectionID,
		IndexUID:     src.IndexUID,
		QueryJSON:    src.QueryJSON,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Prepend so the duplicate appears first.
	s.data.Items = append([]SavedQuery{dup}, s.data.Items...)
	if len(s.data.Items) > maxSavedQueries {
		s.data.Items = s.data.Items[:maxSavedQueries]
	}

	if err := s.saveNoLock(); err != nil {
		return SavedQuery{}, err
	}
	return dup, nil
}

// --- internal helpers ---

// ensureLoadedNoLock loads the JSON file on first access. The caller must hold
// s.mu before calling this method.
func (s *SavedQueryService) ensureLoadedNoLock() {
	if s.loaded {
		return
	}
	s.loaded = true

	path := s.filePath()
	b, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			slog.Warn("savedquery: failed to read file", "path", path, "err", err)
		}
		s.data = savedQueriesFile{Version: savedQueriesVersion}
		return
	}

	var fd savedQueriesFile
	if err := json.Unmarshal(b, &fd); err != nil {
		slog.Warn("savedquery: corrupted file, resetting", "path", path, "err", err)
		backup := path + ".corrupted"
		_ = os.Rename(path, backup)
		s.data = savedQueriesFile{Version: savedQueriesVersion}
		return
	}
	s.data = fd
}

// saveNoLock marshals s.data and writes it atomically. The caller must hold
// s.mu.
func (s *SavedQueryService) saveNoLock() error {
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("savedquery: marshal: %w", err)
	}

	path := s.filePath()
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0600); err != nil {
		return fmt.Errorf("savedquery: write tmp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("savedquery: rename: %w", err)
	}
	return nil
}

func (s *SavedQueryService) filePath() string {
	return filepath.Join(s.cfgManager.Dir(), savedQueriesFileName)
}

// validateQueryJSON returns an error if queryJSON is not a valid JSON object.
func validateQueryJSON(queryJSON string) error {
	if !json.Valid([]byte(queryJSON)) {
		return fmt.Errorf("savedquery: queryJson is not valid JSON")
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal([]byte(queryJSON), &obj); err != nil {
		return fmt.Errorf("savedquery: queryJson must be a JSON object")
	}
	return nil
}
