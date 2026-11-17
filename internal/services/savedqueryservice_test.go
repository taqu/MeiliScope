package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// newSavedQuerySvc creates a SavedQueryService backed by a fresh temp directory.
func newSavedQuerySvc(t *testing.T) *SavedQueryService {
	t.Helper()
	return NewSavedQueryService(newTestManager(t))
}

// createTestQuery is a convenience wrapper that calls CreateSavedQuery with
// sensible defaults and fatals on error.
func createTestQuery(t *testing.T, svc *SavedQueryService, name string) SavedQuery {
	t.Helper()
	q, err := svc.CreateSavedQuery(context.Background(), CreateSavedQueryRequest{
		Name:         name,
		ConnectionID: "conn-1",
		IndexUID:     "movies",
		QueryJSON:    `{"q":"test"}`,
	})
	if err != nil {
		t.Fatalf("CreateSavedQuery(%q): %v", name, err)
	}
	return q
}

// --- Create + List ---

func TestSavedQueryService_CreateAndList(t *testing.T) {
	svc := newSavedQuerySvc(t)
	ctx := context.Background()

	created, err := svc.CreateSavedQuery(ctx, CreateSavedQueryRequest{
		Name:         "My Query",
		ConnectionID: "conn-1",
		IndexUID:     "movies",
		QueryJSON:    `{"q":"hello"}`,
	})
	if err != nil {
		t.Fatalf("CreateSavedQuery: %v", err)
	}
	if created.ID == "" {
		t.Error("expected non-empty ID")
	}
	if created.Name != "My Query" {
		t.Errorf("Name: got %q, want %q", created.Name, "My Query")
	}
	if created.CreatedAt == "" {
		t.Error("expected non-empty CreatedAt")
	}
	if created.UpdatedAt == "" {
		t.Error("expected non-empty UpdatedAt")
	}

	list, err := svc.ListSavedQueries(ctx)
	if err != nil {
		t.Fatalf("ListSavedQueries: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("list length: got %d, want 1", len(list))
	}
	if list[0].ID != created.ID {
		t.Errorf("listed ID: got %q, want %q", list[0].ID, created.ID)
	}
}

func TestSavedQueryService_Create_EmptyName(t *testing.T) {
	svc := newSavedQuerySvc(t)
	_, err := svc.CreateSavedQuery(context.Background(), CreateSavedQueryRequest{
		Name:      "",
		QueryJSON: `{"q":"test"}`,
	})
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestSavedQueryService_Create_InvalidJSON(t *testing.T) {
	svc := newSavedQuerySvc(t)
	_, err := svc.CreateSavedQuery(context.Background(), CreateSavedQueryRequest{
		Name:      "Bad Query",
		QueryJSON: `not json at all`,
	})
	if err == nil {
		t.Error("expected error for non-JSON queryJSON")
	}
}

func TestSavedQueryService_Create_ArrayJSON(t *testing.T) {
	svc := newSavedQuerySvc(t)
	_, err := svc.CreateSavedQuery(context.Background(), CreateSavedQueryRequest{
		Name:      "Array Query",
		QueryJSON: `[1,2,3]`,
	})
	if err == nil {
		t.Error("expected error for array JSON (not an object)")
	}
}

// --- Get ---

func TestSavedQueryService_Get_Found(t *testing.T) {
	svc := newSavedQuerySvc(t)
	ctx := context.Background()

	created := createTestQuery(t, svc, "Findable")

	got, err := svc.GetSavedQuery(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetSavedQuery: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID: got %q, want %q", got.ID, created.ID)
	}
	if got.Name != created.Name {
		t.Errorf("Name: got %q, want %q", got.Name, created.Name)
	}
}

func TestSavedQueryService_Get_NotFound(t *testing.T) {
	svc := newSavedQuerySvc(t)
	_, err := svc.GetSavedQuery(context.Background(), "nonexistent-id")
	if err == nil {
		t.Error("expected error for nonexistent ID")
	}
}

// --- Update ---

func TestSavedQueryService_Update_Name(t *testing.T) {
	svc := newSavedQuerySvc(t)
	ctx := context.Background()

	created := createTestQuery(t, svc, "Original Name")

	updated, err := svc.UpdateSavedQuery(ctx, created.ID, UpdateSavedQueryRequest{
		Name:         "Renamed",
		ConnectionID: created.ConnectionID,
		IndexUID:     created.IndexUID,
		QueryJSON:    created.QueryJSON,
	})
	if err != nil {
		t.Fatalf("UpdateSavedQuery: %v", err)
	}
	if updated.Name != "Renamed" {
		t.Errorf("Name after update: got %q, want %q", updated.Name, "Renamed")
	}
	if updated.ID != created.ID {
		t.Errorf("ID must not change: got %q", updated.ID)
	}
	if updated.UpdatedAt == created.UpdatedAt {
		// UpdatedAt should change, but both are RFC3339 and may be the same
		// sub-second; we cannot guarantee they differ, so we only verify non-empty.
		_ = updated.UpdatedAt
	}
	if updated.UpdatedAt == "" {
		t.Error("expected non-empty UpdatedAt after update")
	}

	// Verify persistence via Get.
	got, err := svc.GetSavedQuery(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetSavedQuery after update: %v", err)
	}
	if got.Name != "Renamed" {
		t.Errorf("persisted Name: got %q, want %q", got.Name, "Renamed")
	}
}

func TestSavedQueryService_Update_QueryJSON(t *testing.T) {
	svc := newSavedQuerySvc(t)
	ctx := context.Background()

	created := createTestQuery(t, svc, "Query JSON Update")

	newJSON := `{"q":"updated","limit":50}`
	updated, err := svc.UpdateSavedQuery(ctx, created.ID, UpdateSavedQueryRequest{
		Name:         created.Name,
		ConnectionID: created.ConnectionID,
		IndexUID:     created.IndexUID,
		QueryJSON:    newJSON,
	})
	if err != nil {
		t.Fatalf("UpdateSavedQuery: %v", err)
	}
	if updated.QueryJSON != newJSON {
		t.Errorf("QueryJSON: got %q, want %q", updated.QueryJSON, newJSON)
	}
}

func TestSavedQueryService_Update_NotFound(t *testing.T) {
	svc := newSavedQuerySvc(t)
	_, err := svc.UpdateSavedQuery(context.Background(), "ghost-id", UpdateSavedQueryRequest{
		Name:      "Whatever",
		QueryJSON: `{}`,
	})
	if err == nil {
		t.Error("expected error for nonexistent ID")
	}
}

// --- Delete ---

func TestSavedQueryService_Delete(t *testing.T) {
	svc := newSavedQuerySvc(t)
	ctx := context.Background()

	created := createTestQuery(t, svc, "To Be Deleted")

	if err := svc.DeleteSavedQuery(ctx, created.ID); err != nil {
		t.Fatalf("DeleteSavedQuery: %v", err)
	}

	list, err := svc.ListSavedQueries(ctx)
	if err != nil {
		t.Fatalf("ListSavedQueries: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("list after delete: got %d items, want 0", len(list))
	}
}

func TestSavedQueryService_Delete_NotFound(t *testing.T) {
	svc := newSavedQuerySvc(t)
	err := svc.DeleteSavedQuery(context.Background(), "does-not-exist")
	if err == nil {
		t.Error("expected error when deleting nonexistent ID")
	}
}

// --- Duplicate ---

func TestSavedQueryService_Duplicate(t *testing.T) {
	svc := newSavedQuerySvc(t)
	ctx := context.Background()

	original := createTestQuery(t, svc, "Alpha")

	dup, err := svc.DuplicateSavedQuery(ctx, original.ID)
	if err != nil {
		t.Fatalf("DuplicateSavedQuery: %v", err)
	}

	wantName := "Alpha (copy)"
	if dup.Name != wantName {
		t.Errorf("duplicate Name: got %q, want %q", dup.Name, wantName)
	}
	if dup.ID == original.ID {
		t.Error("duplicate must have a different ID")
	}
	if dup.QueryJSON != original.QueryJSON {
		t.Errorf("duplicate QueryJSON: got %q, want %q", dup.QueryJSON, original.QueryJSON)
	}
	if dup.ConnectionID != original.ConnectionID {
		t.Errorf("duplicate ConnectionID: got %q, want %q", dup.ConnectionID, original.ConnectionID)
	}

	// Both entries should exist in the list.
	list, err := svc.ListSavedQueries(ctx)
	if err != nil {
		t.Fatalf("ListSavedQueries: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("list length after duplicate: got %d, want 2", len(list))
	}
	// Newest (duplicate) should be first.
	if list[0].ID != dup.ID {
		t.Errorf("first item should be the duplicate: got %q, want %q", list[0].ID, dup.ID)
	}
}

// --- Persistence ---

func TestSavedQueryService_Persistence(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	svc1 := NewSavedQueryService(mgr)
	q1, err := svc1.CreateSavedQuery(ctx, CreateSavedQueryRequest{
		Name:         "Persistent Query",
		ConnectionID: "conn-1",
		IndexUID:     "books",
		QueryJSON:    `{"q":"persist"}`,
	})
	if err != nil {
		t.Fatalf("CreateSavedQuery: %v", err)
	}

	// New service instance on the same directory must see the saved data.
	svc2 := NewSavedQueryService(mgr)
	list, err := svc2.ListSavedQueries(ctx)
	if err != nil {
		t.Fatalf("ListSavedQueries (svc2): %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("svc2 list length: got %d, want 1", len(list))
	}
	if list[0].ID != q1.ID {
		t.Errorf("persisted ID: got %q, want %q", list[0].ID, q1.ID)
	}
	if list[0].Name != q1.Name {
		t.Errorf("persisted Name: got %q, want %q", list[0].Name, q1.Name)
	}
}

// --- QueryJSON round-trip fidelity ---

func TestSavedQueryService_QueryJSONPreserved(t *testing.T) {
	svc := newSavedQuerySvc(t)
	ctx := context.Background()

	// Unicode, large integers, nested structures.
	original := `{"q":"日本語テスト","limit":9007199254740993,"nested":{"arr":[1,2,3],"flag":true}}`

	created, err := svc.CreateSavedQuery(ctx, CreateSavedQueryRequest{
		Name:         "Unicode/Large",
		ConnectionID: "c",
		IndexUID:     "i",
		QueryJSON:    original,
	})
	if err != nil {
		t.Fatalf("CreateSavedQuery: %v", err)
	}
	if created.QueryJSON != original {
		t.Errorf("QueryJSON not preserved:\n got: %s\nwant: %s", created.QueryJSON, original)
	}

	// Reload from disk via a fresh service on the same manager.
	mgr := newTestManager(t)
	svc2 := NewSavedQueryService(mgr)
	created2, err := svc2.CreateSavedQuery(ctx, CreateSavedQueryRequest{
		Name:         "Unicode/Large",
		ConnectionID: "c",
		IndexUID:     "i",
		QueryJSON:    original,
	})
	if err != nil {
		t.Fatalf("CreateSavedQuery (svc2): %v", err)
	}

	svc3 := NewSavedQueryService(mgr)
	got, err := svc3.GetSavedQuery(ctx, created2.ID)
	if err != nil {
		t.Fatalf("GetSavedQuery (svc3): %v", err)
	}
	if got.QueryJSON != original {
		t.Errorf("QueryJSON after reload:\n got: %s\nwant: %s", got.QueryJSON, original)
	}
}

// --- Max items cap ---

func TestSavedQueryService_MaxItems(t *testing.T) {
	svc := newSavedQuerySvc(t)
	ctx := context.Background()

	// Create one more than the limit.
	total := maxSavedQueries + 1
	for i := 0; i < total; i++ {
		_, err := svc.CreateSavedQuery(ctx, CreateSavedQueryRequest{
			Name:         fmt.Sprintf("Query %d", i),
			ConnectionID: "c",
			IndexUID:     "i",
			QueryJSON:    `{}`,
		})
		if err != nil {
			t.Fatalf("CreateSavedQuery %d: %v", i, err)
		}
	}

	list, err := svc.ListSavedQueries(ctx)
	if err != nil {
		t.Fatalf("ListSavedQueries: %v", err)
	}
	if len(list) > maxSavedQueries {
		t.Errorf("list length exceeds cap: got %d, want ≤ %d", len(list), maxSavedQueries)
	}
}

// --- Corrupted file recovery ---

func TestSavedQueryService_CorruptedFile(t *testing.T) {
	mgr := newTestManager(t)

	// Write garbage bytes to saved_queries.json before any service is created.
	filePath := filepath.Join(mgr.Dir(), savedQueriesFileName)
	if err := os.WriteFile(filePath, []byte("{{not valid json at all!"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	svc := NewSavedQueryService(mgr)
	list, err := svc.ListSavedQueries(context.Background())
	if err != nil {
		t.Fatalf("ListSavedQueries after corruption: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list after corruption recovery, got %d items", len(list))
	}

	// A .corrupted backup file should exist.
	if _, statErr := os.Stat(filePath + ".corrupted"); statErr != nil {
		t.Error("expected .corrupted backup file to exist after recovery")
	}
}
