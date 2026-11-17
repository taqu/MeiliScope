package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/taqu/meiliscope/internal/meilisearch"
)

// handler is a map of path → handler function used to build test servers.
type handler map[string]http.HandlerFunc

// defaultHandlers returns the standard probe endpoints required by SelectConnection.
func defaultHandlers() handler {
	return handler{
		"/health":  jsonHandler(`{"status":"available"}`),
		"/version": jsonHandler(`{"pkgVersion":"1.7.0","commitSha":"abc","commitDate":"2024-01-01"}`),
		"/indexes": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, map[string]any{"results": []any{}, "offset": 0, "limit": 20, "total": 0})
		},
	}
}

// newSrv creates a test server from the given handlers.
// Handlers in h override defaults for matching paths.
func newSrv(t *testing.T, h handler) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	defaults := defaultHandlers()
	// Apply test-specific handlers first so they take precedence.
	registered := map[string]bool{}
	for path, fn := range h {
		mux.HandleFunc(path, fn)
		registered[path] = true
	}
	// Apply defaults for paths not overridden.
	for path, fn := range defaults {
		if !registered[path] {
			mux.HandleFunc(path, fn)
		}
	}
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func jsonHandler(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, body)
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONStatus(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// activate creates a ConnectionService connected to srv and an IndexService on top.
func activate(t *testing.T, srv *httptest.Server) (*ConnectionService, *IndexService) {
	t.Helper()
	creds := newMockCreds()
	connSvc := NewConnectionService(newTestManager(t), creds)
	ctx := context.Background()

	sum, err := connSvc.CreateConnection(ctx, SaveConnectionRequest{
		Name: "Test", URL: srv.URL,
	})
	if err != nil {
		t.Fatalf("CreateConnection: %v", err)
	}
	if _, err := connSvc.SelectConnection(ctx, sum.ID); err != nil {
		t.Fatalf("SelectConnection: %v", err)
	}

	return connSvc, NewIndexService(connSvc)
}

// --- no active connection ---

func TestIndexService_NoActiveConnection(t *testing.T) {
	creds := newMockCreds()
	connSvc := NewConnectionService(newTestManager(t), creds)
	idxSvc := NewIndexService(connSvc)

	_, err := idxSvc.ListIndexes(context.Background(), 0, 20)
	if err == nil {
		t.Error("expected error when no active connection")
	}
}

// --- ListIndexes ---

func TestIndexService_ListIndexes(t *testing.T) {
	pk := "id"
	srv := newSrv(t, handler{
		"/indexes": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.RawQuery == "offset=0&limit=20" {
				writeJSON(w, map[string]any{
					"results": []map[string]any{
						{"uid": "movies", "primaryKey": pk, "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-02T00:00:00Z"},
						{"uid": "books", "primaryKey": nil, "createdAt": "2024-01-03T00:00:00Z", "updatedAt": "2024-01-04T00:00:00Z"},
					},
					"offset": 0, "limit": 20, "total": 2,
				})
			}
		},
	})

	_, idxSvc := activate(t, srv)

	result, err := idxSvc.ListIndexes(context.Background(), 0, 20)
	if err != nil {
		t.Fatalf("ListIndexes: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("Total: got %d, want 2", result.Total)
	}
	if len(result.Items) != 2 {
		t.Fatalf("Items count: got %d, want 2", len(result.Items))
	}
	if result.Items[0].UID != "movies" {
		t.Errorf("Items[0].UID: got %q, want movies", result.Items[0].UID)
	}
	if result.Items[0].PrimaryKey == nil || *result.Items[0].PrimaryKey != "id" {
		t.Error("Items[0].PrimaryKey: expected 'id'")
	}
	if result.Items[1].PrimaryKey != nil {
		t.Error("Items[1].PrimaryKey: expected nil")
	}
}

func TestIndexService_ListIndexes_DefaultLimit(t *testing.T) {
	var gotQuery string
	srv := newSrv(t, handler{
		"/indexes": func(w http.ResponseWriter, r *http.Request) {
			gotQuery = r.URL.RawQuery
			writeJSON(w, map[string]any{"results": []any{}, "offset": 0, "limit": 20, "total": 0})
		},
	})

	_, idxSvc := activate(t, srv)
	_, err := idxSvc.ListIndexes(context.Background(), 0, 0) // 0 → default 20
	if err != nil {
		t.Fatalf("ListIndexes: %v", err)
	}
	if gotQuery != "offset=0&limit=20" {
		t.Errorf("unexpected query: %q (expected limit=20)", gotQuery)
	}
}

func TestIndexService_ListIndexes_Empty(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes": func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, map[string]any{"results": []any{}, "offset": 0, "limit": 20, "total": 0})
		},
	})

	_, idxSvc := activate(t, srv)
	result, err := idxSvc.ListIndexes(context.Background(), 0, 20)
	if err != nil {
		t.Fatalf("ListIndexes: %v", err)
	}
	if result.Total != 0 || len(result.Items) != 0 {
		t.Errorf("expected empty result, got total=%d items=%d", result.Total, len(result.Items))
	}
}

func TestIndexService_ListIndexes_Pagination(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes": func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, map[string]any{
				"results": []map[string]string{{"uid": "page2", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}},
				"offset":  20, "limit": 20, "total": 25,
			})
		},
	})

	_, idxSvc := activate(t, srv)
	result, err := idxSvc.ListIndexes(context.Background(), 20, 20)
	if err != nil {
		t.Fatalf("ListIndexes: %v", err)
	}
	if result.Offset != 20 {
		t.Errorf("Offset: got %d, want 20", result.Offset)
	}
	if result.Total != 25 {
		t.Errorf("Total: got %d, want 25", result.Total)
	}
}

func TestIndexService_ListIndexes_AuthError(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes": func(w http.ResponseWriter, _ *http.Request) {
			writeJSONStatus(w, http.StatusUnauthorized, map[string]string{
				"message": "authentication required", "code": "missing_authorization_header",
				"type": "auth_error", "link": "",
			})
		},
	})

	_, idxSvc := activate(t, srv)
	_, err := idxSvc.ListIndexes(context.Background(), 0, 20)
	if err == nil {
		t.Fatal("expected error for 401")
	}
}

// --- GetIndex ---

func TestIndexService_GetIndex(t *testing.T) {
	pk := "id"
	srv := newSrv(t, handler{
		"/indexes/movies": func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, map[string]any{
				"uid": "movies", "primaryKey": pk,
				"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-02T00:00:00Z",
			})
		},
	})

	_, idxSvc := activate(t, srv)
	idx, err := idxSvc.GetIndex(context.Background(), "movies")
	if err != nil {
		t.Fatalf("GetIndex: %v", err)
	}
	if idx.UID != "movies" {
		t.Errorf("UID: got %q, want movies", idx.UID)
	}
	if idx.PrimaryKey == nil || *idx.PrimaryKey != "id" {
		t.Error("PrimaryKey: expected 'id'")
	}
}

func TestIndexService_GetIndex_NoPrimaryKey(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/empty-idx": func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, map[string]any{
				"uid": "empty-idx", "primaryKey": nil,
				"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
			})
		},
	})

	_, idxSvc := activate(t, srv)
	idx, err := idxSvc.GetIndex(context.Background(), "empty-idx")
	if err != nil {
		t.Fatalf("GetIndex: %v", err)
	}
	if idx.PrimaryKey != nil {
		t.Errorf("PrimaryKey: expected nil, got %q", *idx.PrimaryKey)
	}
}

func TestIndexService_GetIndex_NotFound(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/missing": func(w http.ResponseWriter, _ *http.Request) {
			writeJSONStatus(w, http.StatusNotFound, map[string]string{
				"message": "index not found", "code": "index_not_found",
				"type": "invalid_request_error", "link": "",
			})
		},
	})

	_, idxSvc := activate(t, srv)
	_, err := idxSvc.GetIndex(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for 404")
	}
}

// --- GetIndexStats ---

func TestIndexService_GetIndexStats(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/movies/stats": func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, map[string]any{
				"numberOfDocuments": 125000,
				"isIndexing":        false,
				"fieldDistribution": map[string]int64{
					"title":    125000,
					"overview": 124800,
					"id":       125000,
				},
			})
		},
	})

	_, idxSvc := activate(t, srv)
	stats, err := idxSvc.GetIndexStats(context.Background(), "movies")
	if err != nil {
		t.Fatalf("GetIndexStats: %v", err)
	}
	if stats.NumberOfDocuments != 125000 {
		t.Errorf("NumberOfDocuments: got %d, want 125000", stats.NumberOfDocuments)
	}
	if stats.IsIndexing {
		t.Error("expected IsIndexing=false")
	}
	if stats.StatsUnavailable {
		t.Error("expected StatsUnavailable=false")
	}
	if len(stats.FieldDistribution) != 3 {
		t.Errorf("FieldDistribution length: got %d, want 3", len(stats.FieldDistribution))
	}
}

func TestIndexService_GetIndexStats_SortedDescending(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/test/stats": func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, map[string]any{
				"numberOfDocuments": 10,
				"isIndexing":        false,
				"fieldDistribution": map[string]int64{
					"rare":   2,
					"common": 10,
					"medium": 5,
				},
			})
		},
	})

	_, idxSvc := activate(t, srv)
	stats, err := idxSvc.GetIndexStats(context.Background(), "test")
	if err != nil {
		t.Fatalf("GetIndexStats: %v", err)
	}
	if stats.FieldDistribution[0].Name != "common" {
		t.Errorf("expected 'common' first, got %q", stats.FieldDistribution[0].Name)
	}
	if stats.FieldDistribution[2].Name != "rare" {
		t.Errorf("expected 'rare' last, got %q", stats.FieldDistribution[2].Name)
	}
}

func TestIndexService_GetIndexStats_PermissionDenied(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/restricted/stats": func(w http.ResponseWriter, _ *http.Request) {
			writeJSONStatus(w, http.StatusForbidden, map[string]string{
				"message": "access denied", "code": "invalid_api_key",
				"type": "auth_error", "link": "",
			})
		},
	})

	_, idxSvc := activate(t, srv)
	stats, err := idxSvc.GetIndexStats(context.Background(), "restricted")
	if err != nil {
		t.Fatalf("GetIndexStats should not return error for 403, got: %v", err)
	}
	if !stats.StatsUnavailable {
		t.Error("expected StatsUnavailable=true")
	}
	if stats.StatsError == "" {
		t.Error("expected non-empty StatsError")
	}
}

func TestIndexService_GetIndexStats_AuthRequired(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/secured/stats": func(w http.ResponseWriter, _ *http.Request) {
			writeJSONStatus(w, http.StatusUnauthorized, map[string]string{
				"message": "authentication required", "code": "missing_authorization_header",
				"type": "auth_error", "link": "",
			})
		},
	})

	_, idxSvc := activate(t, srv)
	stats, err := idxSvc.GetIndexStats(context.Background(), "secured")
	if err != nil {
		t.Fatalf("GetIndexStats should not return error for 401, got: %v", err)
	}
	if !stats.StatsUnavailable {
		t.Error("expected StatsUnavailable=true")
	}
}

func TestIndexService_GetIndexStats_IndexNotFound(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/gone/stats": func(w http.ResponseWriter, _ *http.Request) {
			writeJSONStatus(w, http.StatusNotFound, map[string]string{
				"message": "index not found", "code": "index_not_found",
				"type": "invalid_request_error", "link": "",
			})
		},
	})

	_, idxSvc := activate(t, srv)
	stats, err := idxSvc.GetIndexStats(context.Background(), "gone")
	if err != nil {
		t.Fatalf("GetIndexStats should not return error for 404, got: %v", err)
	}
	if !stats.StatsUnavailable {
		t.Error("expected StatsUnavailable=true for missing index stats")
	}
}

// --- stale-response guard ---

func TestIndexService_StaleResponseGuard(t *testing.T) {
	srv := newSrv(t, nil) // default handlers only

	connSvc, idxSvc := activate(t, srv)

	_, err := idxSvc.ListIndexes(context.Background(), 0, 20)
	if err != nil {
		t.Fatalf("first ListIndexes: %v", err)
	}

	// Disconnecting invalidates the generation; subsequent calls must fail.
	connSvc.Disconnect()

	_, err = idxSvc.ListIndexes(context.Background(), 0, 20)
	if err == nil {
		t.Error("expected error after disconnect")
	}
}

// --- context cancellation ---

func TestIndexService_ListIndexes_Canceled(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes": func(w http.ResponseWriter, r *http.Request) {
			<-r.Context().Done()
			w.WriteHeader(http.StatusServiceUnavailable)
		},
	})

	_, idxSvc := activate(t, srv)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := idxSvc.ListIndexes(ctx, 0, 20)
	if err == nil {
		t.Error("expected error for canceled context")
	}
}

// --- large document counts ---

func TestIndexService_GetIndexStats_LargeCount(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/large/stats": func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, map[string]any{
				"numberOfDocuments": int64(10_000_000),
				"isIndexing":        true,
				"fieldDistribution": map[string]int64{},
			})
		},
	})

	_, idxSvc := activate(t, srv)
	stats, err := idxSvc.GetIndexStats(context.Background(), "large")
	if err != nil {
		t.Fatalf("GetIndexStats: %v", err)
	}
	if stats.NumberOfDocuments != 10_000_000 {
		t.Errorf("NumberOfDocuments: got %d, want 10000000", stats.NumberOfDocuments)
	}
	if !stats.IsIndexing {
		t.Error("expected IsIndexing=true")
	}
}

// --- path escaping ---

func TestIndexService_GetIndex_PathEscaping(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/my%20index": func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, map[string]any{
				"uid": "my index", "primaryKey": nil,
				"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
			})
		},
	})

	_, idxSvc := activate(t, srv)
	idx, err := idxSvc.GetIndex(context.Background(), "my index")
	if err != nil {
		t.Fatalf("GetIndex with space in UID: %v", err)
	}
	if idx.UID != "my index" {
		t.Errorf("UID: got %q, want %q", idx.UID, "my index")
	}
}

// --- toIndexStatistics field distribution sort (alphabetic tie-break) ---

func TestToIndexStatistics_FieldDistribution_AlphabeticTieBreak(t *testing.T) {
	stats := &meilisearch.IndexStats{
		NumberOfDocuments: 10,
		IsIndexing:        false,
		FieldDistribution: map[string]int64{
			"z-field": 10,
			"a-field": 10,
			"m-field": 10,
		},
	}
	out := toIndexStatistics(stats)
	if len(out.FieldDistribution) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(out.FieldDistribution))
	}
	names := []string{out.FieldDistribution[0].Name, out.FieldDistribution[1].Name, out.FieldDistribution[2].Name}
	if names[0] != "a-field" || names[1] != "m-field" || names[2] != "z-field" {
		t.Errorf("unexpected order: %v", names)
	}
}

// --- wrapIndexError ---

func TestWrapIndexError(t *testing.T) {
	cases := []struct {
		name    string
		err     error
		wantStr string
	}{
		{"auth", &meilisearch.APIError{Kind: meilisearch.ErrKindAuth, Msg: "unauth"}, "authentication required"},
		{"forbidden", &meilisearch.APIError{Kind: meilisearch.ErrKindForbidden, Msg: "denied"}, "access denied"},
		{"canceled", &meilisearch.APIError{Kind: meilisearch.ErrKindCanceled, Msg: "canceled"}, "canceled"},
		{"other", fmt.Errorf("some error"), "some error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := wrapIndexError("op", tc.err)
			if got == nil {
				t.Fatal("expected non-nil error")
			}
			if !containsStr(got.Error(), tc.wantStr) {
				t.Errorf("wrapIndexError(%q): got %q, want to contain %q", tc.name, got.Error(), tc.wantStr)
			}
		})
	}
}

func containsStr(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	for i := range len(s) - len(sub) + 1 {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
