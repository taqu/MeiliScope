package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

const (
	historyFileName = "query_history.json"
	maxHistoryItems = 100
	maxQueryBytes   = 64 * 1024 // 64 KiB
	historyVersion  = 1
)

type historyFileData struct {
	Version int            `json:"version"`
	Entries []HistoryEntry `json:"entries"`
}

type historyStore struct {
	path string
	mu   sync.Mutex
	data historyFileData
}

func newHistoryStore(dir string) *historyStore {
	return &historyStore{
		path: filepath.Join(dir, historyFileName),
		data: historyFileData{Version: historyVersion},
	}
}

func (h *historyStore) load() {
	h.mu.Lock()
	defer h.mu.Unlock()

	b, err := os.ReadFile(h.path)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			slog.Warn("history: failed to read file", "path", h.path, "err", err)
		}
		h.data = historyFileData{Version: historyVersion}
		return
	}

	var hf historyFileData
	if err := json.Unmarshal(b, &hf); err != nil {
		slog.Warn("history: corrupted file, resetting", "path", h.path, "err", err)
		backup := h.path + ".corrupted"
		_ = os.Rename(h.path, backup)
		h.data = historyFileData{Version: historyVersion}
		return
	}
	h.data = hf
}

func (h *historyStore) add(entry HistoryEntry) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Prepend so newest entries come first.
	h.data.Entries = append([]HistoryEntry{entry}, h.data.Entries...)
	if len(h.data.Entries) > maxHistoryItems {
		h.data.Entries = h.data.Entries[:maxHistoryItems]
	}
	if err := h.saveNoLock(); err != nil {
		slog.Warn("history: failed to save", "err", err)
	}
}

func (h *historyStore) list() []HistoryEntry {
	h.mu.Lock()
	defer h.mu.Unlock()

	out := make([]HistoryEntry, len(h.data.Entries))
	copy(out, h.data.Entries)
	return out
}

func (h *historyStore) delete(id string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i, e := range h.data.Entries {
		if e.ID == id {
			h.data.Entries = append(h.data.Entries[:i], h.data.Entries[i+1:]...)
			return h.saveNoLock()
		}
	}
	return fmt.Errorf("history: entry %q not found", id)
}

func (h *historyStore) clear() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.data.Entries = nil
	return h.saveNoLock()
}

func (h *historyStore) saveNoLock() error {
	b, err := json.MarshalIndent(h.data, "", "  ")
	if err != nil {
		return fmt.Errorf("history: marshal: %w", err)
	}

	tmp := h.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0600); err != nil {
		return fmt.Errorf("history: write tmp: %w", err)
	}
	if err := os.Rename(tmp, h.path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("history: rename: %w", err)
	}
	return nil
}
