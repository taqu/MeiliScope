package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// activateDoc creates a ConnectionService connected to srv and a DocumentService on top.
func activateDoc(t *testing.T, srv *httptest.Server) (*ConnectionService, *DocumentService) {
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

	return connSvc, NewDocumentService(connSvc)
}

// --- no active connection ---

func TestDocumentService_NoActiveConnection(t *testing.T) {
	creds := newMockCreds()
	connSvc := NewConnectionService(newTestManager(t), creds)
	docSvc := NewDocumentService(connSvc)

	_, err := docSvc.ListDocuments(context.Background(), "movies", DocumentListOptions{})
	if err == nil {
		t.Error("expected error when no active connection")
	}
}

// --- ListDocuments ---

func TestDocumentService_ListDocuments(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/movies/documents": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, map[string]any{
				"results": []map[string]any{
					{"id": 1, "title": "Inception"},
					{"id": 2, "title": "Interstellar"},
				},
				"offset": 0, "limit": 50, "total": 2,
			})
		},
	})

	_, docSvc := activateDoc(t, srv)
	page, err := docSvc.ListDocuments(context.Background(), "movies", DocumentListOptions{Limit: 50})
	if err != nil {
		t.Fatalf("ListDocuments: %v", err)
	}
	if page.Total != 2 {
		t.Errorf("Total: got %d, want 2", page.Total)
	}
	if len(page.Items) != 2 {
		t.Fatalf("Items count: got %d, want 2", len(page.Items))
	}
	if page.Items[0].Fields["title"] != `"Inception"` {
		t.Errorf("Fields[title]: got %q, want %q", page.Items[0].Fields["title"], `"Inception"`)
	}
}

func TestDocumentService_ListDocuments_Empty(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/empty/documents": func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, map[string]any{"results": []any{}, "offset": 0, "limit": 50, "total": 0})
		},
	})

	_, docSvc := activateDoc(t, srv)
	page, err := docSvc.ListDocuments(context.Background(), "empty", DocumentListOptions{})
	if err != nil {
		t.Fatalf("ListDocuments: %v", err)
	}
	if page.Total != 0 || len(page.Items) != 0 {
		t.Errorf("expected empty result, got total=%d items=%d", page.Total, len(page.Items))
	}
}

func TestDocumentService_ListDocuments_DefaultLimit(t *testing.T) {
	var gotQuery string
	srv := newSrv(t, handler{
		"/indexes/movies/documents": func(w http.ResponseWriter, r *http.Request) {
			gotQuery = r.URL.RawQuery
			writeJSON(w, map[string]any{"results": []any{}, "offset": 0, "limit": 50, "total": 0})
		},
	})

	_, docSvc := activateDoc(t, srv)
	_, err := docSvc.ListDocuments(context.Background(), "movies", DocumentListOptions{Limit: 0})
	if err != nil {
		t.Fatalf("ListDocuments: %v", err)
	}
	if gotQuery != "limit=50&offset=0" && gotQuery != "offset=0&limit=50" {
		t.Errorf("unexpected query: %q (expected offset=0&limit=50)", gotQuery)
	}
}

func TestDocumentService_ListDocuments_LimitCapped(t *testing.T) {
	var gotQuery string
	srv := newSrv(t, handler{
		"/indexes/movies/documents": func(w http.ResponseWriter, r *http.Request) {
			gotQuery = r.URL.RawQuery
			writeJSON(w, map[string]any{"results": []any{}, "offset": 0, "limit": 200, "total": 0})
		},
	})

	_, docSvc := activateDoc(t, srv)
	_, err := docSvc.ListDocuments(context.Background(), "movies", DocumentListOptions{Limit: 9999})
	if err != nil {
		t.Fatalf("ListDocuments: %v", err)
	}
	// Limit must be capped at 200
	if !containsStr(gotQuery, "limit=200") {
		t.Errorf("expected limit=200 in query %q", gotQuery)
	}
}

func TestDocumentService_ListDocuments_Pagination(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/movies/documents": func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, map[string]any{
				"results": []map[string]any{{"id": 51, "title": "Page2"}},
				"offset":  50, "limit": 50, "total": 100,
			})
		},
	})

	_, docSvc := activateDoc(t, srv)
	page, err := docSvc.ListDocuments(context.Background(), "movies", DocumentListOptions{Offset: 50, Limit: 50})
	if err != nil {
		t.Fatalf("ListDocuments: %v", err)
	}
	if page.Offset != 50 {
		t.Errorf("Offset: got %d, want 50", page.Offset)
	}
	if page.Total != 100 {
		t.Errorf("Total: got %d, want 100", page.Total)
	}
}

func TestDocumentService_ListDocuments_DynamicFields(t *testing.T) {
	// Two documents with different fields — the union must be preserved.
	srv := newSrv(t, handler{
		"/indexes/catalog/documents": func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, map[string]any{
				"results": []map[string]any{
					{"id": 1, "title": "First", "category": "A"},
					{"id": 2, "title": "Second", "price": 100},
				},
				"offset": 0, "limit": 50, "total": 2,
			})
		},
	})

	_, docSvc := activateDoc(t, srv)
	page, err := docSvc.ListDocuments(context.Background(), "catalog", DocumentListOptions{})
	if err != nil {
		t.Fatalf("ListDocuments: %v", err)
	}
	if len(page.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(page.Items))
	}

	rec0 := page.Items[0]
	if _, ok := rec0.Fields["category"]; !ok {
		t.Error("first document: expected 'category' field")
	}
	if _, ok := rec0.Fields["price"]; ok {
		t.Error("first document: unexpected 'price' field")
	}

	rec1 := page.Items[1]
	if _, ok := rec1.Fields["price"]; !ok {
		t.Error("second document: expected 'price' field")
	}
}

func TestDocumentService_ListDocuments_NestedObjects(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/docs/documents": func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, map[string]any{
				"results": []map[string]any{
					{"id": 1, "meta": map[string]any{"author": "Alice", "year": 2024}},
				},
				"offset": 0, "limit": 50, "total": 1,
			})
		},
	})

	_, docSvc := activateDoc(t, srv)
	page, err := docSvc.ListDocuments(context.Background(), "docs", DocumentListOptions{})
	if err != nil {
		t.Fatalf("ListDocuments: %v", err)
	}
	metaRaw := page.Items[0].Fields["meta"]
	if len(metaRaw) == 0 {
		t.Fatal("expected non-empty 'meta' field")
	}
	// meta should be a JSON object string starting with {
	if metaRaw[0] != '{' {
		t.Errorf("meta: expected JSON object, got %q", metaRaw)
	}
}

func TestDocumentService_ListDocuments_Arrays(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/docs/documents": func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, map[string]any{
				"results": []map[string]any{
					{"id": 1, "tags": []string{"go", "search", "api"}},
				},
				"offset": 0, "limit": 50, "total": 1,
			})
		},
	})

	_, docSvc := activateDoc(t, srv)
	page, err := docSvc.ListDocuments(context.Background(), "docs", DocumentListOptions{})
	if err != nil {
		t.Fatalf("ListDocuments: %v", err)
	}
	tagsRaw := page.Items[0].Fields["tags"]
	if tagsRaw[0] != '[' {
		t.Errorf("tags: expected JSON array, got %q", tagsRaw)
	}
}

func TestDocumentService_ListDocuments_NullValues(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/docs/documents": func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, map[string]any{
				"results": []map[string]any{
					{"id": 1, "deleted_at": nil},
				},
				"offset": 0, "limit": 50, "total": 1,
			})
		},
	})

	_, docSvc := activateDoc(t, srv)
	page, err := docSvc.ListDocuments(context.Background(), "docs", DocumentListOptions{})
	if err != nil {
		t.Fatalf("ListDocuments: %v", err)
	}
	nullRaw := page.Items[0].Fields["deleted_at"]
	if nullRaw != "null" {
		t.Errorf("deleted_at: expected %q, got %q", "null", nullRaw)
	}
}

func TestDocumentService_ListDocuments_LargeIntegers(t *testing.T) {
	// 9007199254740993 exceeds JS Number.MAX_SAFE_INTEGER — must not be corrupted.
	const largeNum = `9007199254740993`
	srv := newSrv(t, handler{
		"/indexes/docs/documents": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"results":[{"id":` + largeNum + `,"title":"big"}],"offset":0,"limit":50,"total":1}`))
		},
	})

	_, docSvc := activateDoc(t, srv)
	page, err := docSvc.ListDocuments(context.Background(), "docs", DocumentListOptions{})
	if err != nil {
		t.Fatalf("ListDocuments: %v", err)
	}
	idRaw := page.Items[0].Fields["id"]
	if idRaw != largeNum {
		t.Errorf("large integer: got %q, want %q", idRaw, largeNum)
	}
}

func TestDocumentService_ListDocuments_UnicodeFields(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/docs/documents": func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, map[string]any{
				"results": []map[string]any{
					{"id": 1, "タイトル": "テスト", "emoji": "🚀"},
				},
				"offset": 0, "limit": 50, "total": 1,
			})
		},
	})

	_, docSvc := activateDoc(t, srv)
	page, err := docSvc.ListDocuments(context.Background(), "docs", DocumentListOptions{})
	if err != nil {
		t.Fatalf("ListDocuments: %v", err)
	}
	rec := page.Items[0]
	if _, ok := rec.Fields["タイトル"]; !ok {
		t.Error("expected Unicode field name 'タイトル'")
	}
}

func TestDocumentService_ListDocuments_NumericDocID(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/docs/documents": func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, map[string]any{
				"results": []map[string]any{{"id": 42}},
				"offset":  0, "limit": 50, "total": 1,
			})
		},
	})

	_, docSvc := activateDoc(t, srv)
	page, err := docSvc.ListDocuments(context.Background(), "docs", DocumentListOptions{})
	if err != nil {
		t.Fatalf("ListDocuments: %v", err)
	}
	idRaw := page.Items[0].Fields["id"]
	if idRaw != "42" {
		t.Errorf("numeric id: got %q, want %q", idRaw, "42")
	}
}

func TestDocumentService_ListDocuments_AuthError(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/movies/documents": func(w http.ResponseWriter, _ *http.Request) {
			writeJSONStatus(w, http.StatusUnauthorized, map[string]string{
				"message": "authentication required", "code": "missing_authorization_header",
				"type": "auth_error", "link": "",
			})
		},
	})

	_, docSvc := activateDoc(t, srv)
	_, err := docSvc.ListDocuments(context.Background(), "movies", DocumentListOptions{})
	if err == nil {
		t.Fatal("expected error for 401")
	}
	if !containsStr(err.Error(), "authentication") {
		t.Errorf("error should mention authentication: %v", err)
	}
}

func TestDocumentService_ListDocuments_ForbiddenError(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/secret/documents": func(w http.ResponseWriter, _ *http.Request) {
			writeJSONStatus(w, http.StatusForbidden, map[string]string{
				"message": "access denied", "code": "invalid_api_key",
				"type": "auth_error", "link": "",
			})
		},
	})

	_, docSvc := activateDoc(t, srv)
	_, err := docSvc.ListDocuments(context.Background(), "secret", DocumentListOptions{})
	if err == nil {
		t.Fatal("expected error for 403")
	}
	if !containsStr(err.Error(), "access denied") {
		t.Errorf("error should mention access denied: %v", err)
	}
}

func TestDocumentService_ListDocuments_NotFound(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/gone/documents": func(w http.ResponseWriter, _ *http.Request) {
			writeJSONStatus(w, http.StatusNotFound, map[string]string{
				"message": "index not found", "code": "index_not_found",
				"type": "invalid_request_error", "link": "",
			})
		},
	})

	_, docSvc := activateDoc(t, srv)
	_, err := docSvc.ListDocuments(context.Background(), "gone", DocumentListOptions{})
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !containsStr(err.Error(), "not found") {
		t.Errorf("error should mention not found: %v", err)
	}
}

func TestDocumentService_ListDocuments_ServerError(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/broken/documents": func(w http.ResponseWriter, _ *http.Request) {
			writeJSONStatus(w, http.StatusInternalServerError, map[string]string{
				"message": "internal error", "code": "internal", "type": "internal_error", "link": "",
			})
		},
	})

	_, docSvc := activateDoc(t, srv)
	_, err := docSvc.ListDocuments(context.Background(), "broken", DocumentListOptions{})
	if err == nil {
		t.Fatal("expected error for 500")
	}
}

func TestDocumentService_ListDocuments_InvalidJSON(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/bad/documents": func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`not valid json`))
		},
	})

	_, docSvc := activateDoc(t, srv)
	_, err := docSvc.ListDocuments(context.Background(), "bad", DocumentListOptions{})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestDocumentService_ListDocuments_Canceled(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/slow/documents": func(w http.ResponseWriter, r *http.Request) {
			<-r.Context().Done()
			w.WriteHeader(http.StatusServiceUnavailable)
		},
	})

	_, docSvc := activateDoc(t, srv)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := docSvc.ListDocuments(ctx, "slow", DocumentListOptions{})
	if err == nil {
		t.Fatal("expected error for canceled context")
	}
}

func TestDocumentService_ListDocuments_Timeout(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/slow/documents": func(w http.ResponseWriter, r *http.Request) {
			<-r.Context().Done()
			w.WriteHeader(http.StatusServiceUnavailable)
		},
	})

	_, docSvc := activateDoc(t, srv)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, err := docSvc.ListDocuments(ctx, "slow", DocumentListOptions{})
	if err == nil {
		t.Fatal("expected error for timeout")
	}
}

func TestDocumentService_ListDocuments_FieldSelection(t *testing.T) {
	var gotQuery string
	srv := newSrv(t, handler{
		"/indexes/movies/documents": func(w http.ResponseWriter, r *http.Request) {
			gotQuery = r.URL.RawQuery
			writeJSON(w, map[string]any{
				"results": []map[string]any{{"id": 1, "title": "Test"}},
				"offset": 0, "limit": 50, "total": 1,
			})
		},
	})

	_, docSvc := activateDoc(t, srv)
	_, err := docSvc.ListDocuments(context.Background(), "movies", DocumentListOptions{
		Fields: []string{"id", "title"},
	})
	if err != nil {
		t.Fatalf("ListDocuments: %v", err)
	}
	if !containsStr(gotQuery, "fields=") {
		t.Errorf("expected fields param in query %q", gotQuery)
	}
}

func TestDocumentService_ListDocuments_PathEscaping(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/my%20index/documents": func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, map[string]any{
				"results": []map[string]any{{"id": 1}},
				"offset": 0, "limit": 50, "total": 1,
			})
		},
	})

	_, docSvc := activateDoc(t, srv)
	_, err := docSvc.ListDocuments(context.Background(), "my index", DocumentListOptions{})
	if err != nil {
		t.Fatalf("ListDocuments with space in UID: %v", err)
	}
}

func TestDocumentService_ListDocuments_EmptyCollections(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/docs/documents": func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, map[string]any{
				"results": []map[string]any{
					{"id": 1, "emptyObj": map[string]any{}, "emptyArr": []any{}},
				},
				"offset": 0, "limit": 50, "total": 1,
			})
		},
	})

	_, docSvc := activateDoc(t, srv)
	page, err := docSvc.ListDocuments(context.Background(), "docs", DocumentListOptions{})
	if err != nil {
		t.Fatalf("ListDocuments: %v", err)
	}
	rec := page.Items[0]
	if rec.Fields["emptyObj"] != "{}" {
		t.Errorf("emptyObj: got %q, want %q", rec.Fields["emptyObj"], "{}")
	}
	if rec.Fields["emptyArr"] != "[]" {
		t.Errorf("emptyArr: got %q, want %q", rec.Fields["emptyArr"], "[]")
	}
}

// --- GetDocument ---

func TestDocumentService_GetDocument(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/movies/documents/1": func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, map[string]any{"id": 1, "title": "Inception", "year": 2010})
		},
	})

	_, docSvc := activateDoc(t, srv)
	rec, err := docSvc.GetDocument(context.Background(), "movies", "1")
	if err != nil {
		t.Fatalf("GetDocument: %v", err)
	}
	if rec.Fields["title"] != `"Inception"` {
		t.Errorf("title: got %q, want %q", rec.Fields["title"], `"Inception"`)
	}
	if len(rec.RawJSON) == 0 {
		t.Error("expected non-empty RawJSON")
	}
}

func TestDocumentService_GetDocument_StringID(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/movies/documents/abc-123": func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, map[string]any{"id": "abc-123", "title": "Test"})
		},
	})

	_, docSvc := activateDoc(t, srv)
	rec, err := docSvc.GetDocument(context.Background(), "movies", "abc-123")
	if err != nil {
		t.Fatalf("GetDocument: %v", err)
	}
	if rec.Fields["id"] != `"abc-123"` {
		t.Errorf("id: got %q, want %q", rec.Fields["id"], `"abc-123"`)
	}
}

func TestDocumentService_GetDocument_PathEscaping(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/movies/documents/hello%20world": func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, map[string]any{"id": "hello world"})
		},
	})

	_, docSvc := activateDoc(t, srv)
	rec, err := docSvc.GetDocument(context.Background(), "movies", "hello world")
	if err != nil {
		t.Fatalf("GetDocument with space in ID: %v", err)
	}
	if rec.Fields["id"] != `"hello world"` {
		t.Errorf("id: got %q, want %q", rec.Fields["id"], `"hello world"`)
	}
}

func TestDocumentService_GetDocument_NotFound(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/movies/documents/999": func(w http.ResponseWriter, _ *http.Request) {
			writeJSONStatus(w, http.StatusNotFound, map[string]string{
				"message": "document not found", "code": "document_not_found",
				"type": "invalid_request_error", "link": "",
			})
		},
	})

	_, docSvc := activateDoc(t, srv)
	_, err := docSvc.GetDocument(context.Background(), "movies", "999")
	if err == nil {
		t.Fatal("expected error for missing document")
	}
	if !containsStr(err.Error(), "not found") {
		t.Errorf("error should mention not found: %v", err)
	}
}

func TestDocumentService_GetDocument_LargeInteger(t *testing.T) {
	const largeNum = `9007199254740993`
	srv := newSrv(t, handler{
		"/indexes/docs/documents/9007199254740993": func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":` + largeNum + `}`))
		},
	})

	_, docSvc := activateDoc(t, srv)
	rec, err := docSvc.GetDocument(context.Background(), "docs", largeNum)
	if err != nil {
		t.Fatalf("GetDocument: %v", err)
	}
	if rec.Fields["id"] != largeNum {
		t.Errorf("id precision: got %q, want %q", rec.Fields["id"], largeNum)
	}
}

// --- stale-response guard ---

func TestDocumentService_StaleResponseGuard(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/movies/documents": func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, map[string]any{"results": []any{}, "offset": 0, "limit": 50, "total": 0})
		},
	})

	connSvc, docSvc := activateDoc(t, srv)

	_, err := docSvc.ListDocuments(context.Background(), "movies", DocumentListOptions{})
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	connSvc.Disconnect()

	_, err = docSvc.ListDocuments(context.Background(), "movies", DocumentListOptions{})
	if err == nil {
		t.Error("expected error after disconnect")
	}
}

// --- rawToDocumentRecord ---

func TestRawToDocumentRecord_NonObject(t *testing.T) {
	_, err := rawToDocumentRecord([]byte(`[1,2,3]`))
	if err == nil {
		t.Error("expected error for non-object JSON")
	}
}

func TestRawToDocumentRecord_RawJSONPreserved(t *testing.T) {
	input := `{"id":1,"name":"Alice"}`
	rec, err := rawToDocumentRecord([]byte(input))
	if err != nil {
		t.Fatalf("rawToDocumentRecord: %v", err)
	}
	if rec.RawJSON != input {
		t.Errorf("RawJSON: got %q, want %q", rec.RawJSON, input)
	}
}
