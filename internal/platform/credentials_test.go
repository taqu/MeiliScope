package platform

import (
	"context"
	"errors"
	"sync"
	"testing"
)

func TestSessionStore_SaveAndLoad(t *testing.T) {
	s := newSessionStore()
	ctx := context.Background()

	if err := s.Save(ctx, "conn-1", "secret-key"); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := s.Load(ctx, "conn-1")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got != "secret-key" {
		t.Errorf("Load: got %q, want %q", got, "secret-key")
	}
}

func TestSessionStore_LoadNotFound(t *testing.T) {
	s := newSessionStore()
	_, err := s.Load(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSessionStore_Delete(t *testing.T) {
	s := newSessionStore()
	ctx := context.Background()

	_ = s.Save(ctx, "conn-1", "key")
	if err := s.Delete(ctx, "conn-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := s.Load(ctx, "conn-1")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestSessionStore_DeleteNotFound(t *testing.T) {
	s := newSessionStore()
	err := s.Delete(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSessionStore_Replace(t *testing.T) {
	s := newSessionStore()
	ctx := context.Background()

	_ = s.Save(ctx, "conn-1", "old-key")
	_ = s.Save(ctx, "conn-1", "new-key")

	got, err := s.Load(ctx, "conn-1")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got != "new-key" {
		t.Errorf("expected new-key, got %q", got)
	}
}

func TestSessionStore_IsSessionOnly(t *testing.T) {
	s := newSessionStore()
	if !s.IsSessionOnly() {
		t.Error("expected IsSessionOnly to be true")
	}
}

func TestSessionStore_MultipleKeys(t *testing.T) {
	s := newSessionStore()
	ctx := context.Background()

	keys := map[string]string{
		"conn-a": "key-a",
		"conn-b": "key-b",
		"conn-c": "key-c",
	}
	for id, k := range keys {
		_ = s.Save(ctx, id, k)
	}

	for id, want := range keys {
		got, err := s.Load(ctx, id)
		if err != nil {
			t.Errorf("Load(%q): %v", id, err)
			continue
		}
		if got != want {
			t.Errorf("Load(%q): got %q, want %q", id, got, want)
		}
	}
}

func TestSessionStore_ConcurrentAccess(t *testing.T) {
	s := newSessionStore()
	ctx := context.Background()
	const n = 20

	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := "conn"
			_ = s.Save(ctx, id, "key")
			_, _ = s.Load(ctx, id)
			if i%3 == 0 {
				_ = s.Delete(ctx, id)
			}
		}(i)
	}
	wg.Wait()
}
