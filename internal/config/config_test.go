package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()
	if cfg.Version != schemaV1 {
		t.Errorf("expected version %d, got %d", schemaV1, cfg.Version)
	}
	if cfg.ActiveView == "" {
		t.Error("expected non-empty ActiveView")
	}
	if cfg.WindowWidth <= 0 || cfg.WindowHeight <= 0 {
		t.Error("expected positive default window dimensions")
	}
}

func TestManagerLoadMissing(t *testing.T) {
	dir := t.TempDir()
	m := &Manager{dir: dir, path: filepath.Join(dir, configFile)}
	if err := m.Load(); err != nil {
		t.Fatalf("Load missing file: %v", err)
	}
	cfg := m.Get()
	if cfg.Version != schemaV1 {
		t.Errorf("expected default version %d, got %d", schemaV1, cfg.Version)
	}
}

func TestManagerSaveLoad(t *testing.T) {
	dir := t.TempDir()
	m := &Manager{dir: dir, path: filepath.Join(dir, configFile)}
	m.Update(Config{
		Version:      schemaV1,
		WindowWidth:  1920,
		WindowHeight: 1080,
		ActiveView:   "indexes",
	})
	if err := m.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	m2 := &Manager{dir: dir, path: filepath.Join(dir, configFile)}
	if err := m2.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	got := m2.Get()
	if got.WindowWidth != 1920 || got.WindowHeight != 1080 {
		t.Errorf("unexpected dimensions: %dx%d", got.WindowWidth, got.WindowHeight)
	}
	if got.ActiveView != "indexes" {
		t.Errorf("expected activeView 'indexes', got %q", got.ActiveView)
	}
}

func TestManagerCorrupted(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, configFile)
	if err := os.WriteFile(path, []byte("not json {{"), 0600); err != nil {
		t.Fatal(err)
	}
	m := &Manager{dir: dir, path: path}
	if err := m.Load(); err != nil {
		t.Fatalf("Load corrupted: %v", err)
	}
	cfg := m.Get()
	if cfg.Version != schemaV1 {
		t.Error("expected fallback to default on corrupted config")
	}
	if _, err := os.Stat(path + ".corrupted"); err != nil {
		t.Error("expected .corrupted backup file")
	}
}
