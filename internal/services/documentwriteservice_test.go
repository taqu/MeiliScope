package services

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/taqu/meiliscope/internal/meilisearch"
)

// taskBody is a minimal valid task-enqueued JSON response.
const taskBody = `{"taskUid":42,"indexUid":"movies","status":"enqueued","type":"documentAdditionOrUpdate","enqueuedAt":"2024-01-01T00:00:00.000000Z"}`
const deleteTaskBody = `{"taskUid":43,"indexUid":"movies","status":"enqueued","type":"documentDeletion","enqueuedAt":"2024-01-01T00:00:00.000000Z"}`

func activateWrite(t *testing.T, srv *httptest.Server) (*ConnectionService, *DocumentWriteService) {
	t.Helper()
	creds := newMockCreds()
	cs := NewConnectionService(newTestManager(t), creds)
	ctx := context.Background()
	sum, err := cs.CreateConnection(ctx, SaveConnectionRequest{Name: "Test", URL: srv.URL})
	if err != nil {
		t.Fatalf("CreateConnection: %v", err)
	}
	if _, err := cs.SelectConnection(ctx, sum.ID); err != nil {
		t.Fatalf("SelectConnection: %v", err)
	}
	return cs, NewDocumentWriteService(cs)
}

// newWriteSrv builds an httptest.Server that expects the given method and path prefix
// and responds with statusCode and body.
func newWriteSrv(t *testing.T, method, pathPrefix string, statusCode int, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if !strings.HasPrefix(r.URL.Path, pathPrefix) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(body))
	}))
}

func TestDocumentWriteService_AddDocument_OK(t *testing.T) {
	srv := newWriteSrv(t, http.MethodPost, "/indexes/movies/documents", http.StatusAccepted, taskBody)
	defer srv.Close()
	_, ws := activateWrite(t, srv)

	ref, err := ws.AddDocument(context.Background(), "movies", `{"id":1,"title":"Test"}`, "id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.TaskUID != 42 {
		t.Errorf("expected taskUid=42, got %d", ref.TaskUID)
	}
	if ref.Status != "enqueued" {
		t.Errorf("expected status=enqueued, got %q", ref.Status)
	}
}

func TestDocumentWriteService_AddDocument_SentAsArray(t *testing.T) {
	var received string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		body, _ := io.ReadAll(r.Body)
		received = string(body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(taskBody))
	}))
	defer srv.Close()
	_, ws := activateWrite(t, srv)

	_, _ = ws.AddDocument(context.Background(), "movies", `{"id":1}`, "")
	if !strings.HasPrefix(received, "[") || !strings.HasSuffix(received, "]") {
		t.Errorf("expected document wrapped in array, got %q", received)
	}
}

func TestDocumentWriteService_UpdateDocument_UsesPUT(t *testing.T) {
	var method string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(taskBody))
	}))
	defer srv.Close()
	_, ws := activateWrite(t, srv)

	_, _ = ws.UpdateDocument(context.Background(), "movies", `{"id":1,"title":"Updated"}`, "id")
	if method != http.MethodPut {
		t.Errorf("expected PUT, got %s", method)
	}
}

func TestDocumentWriteService_ReplaceDocument_UsesPOST(t *testing.T) {
	var method string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(taskBody))
	}))
	defer srv.Close()
	_, ws := activateWrite(t, srv)

	_, _ = ws.ReplaceDocument(context.Background(), "movies", `{"id":1,"title":"New"}`, "id")
	if method != http.MethodPost {
		t.Errorf("expected POST, got %s", method)
	}
}

func TestDocumentWriteService_DeleteDocument_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/indexes/movies/documents/42" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(deleteTaskBody))
	}))
	defer srv.Close()
	_, ws := activateWrite(t, srv)

	ref, err := ws.DeleteDocument(context.Background(), "movies", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.TaskUID != 43 {
		t.Errorf("expected taskUid=43, got %d", ref.TaskUID)
	}
}

func TestDocumentWriteService_DeleteDocument_IDEncoded(t *testing.T) {
	var requestPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(deleteTaskBody))
	}))
	defer srv.Close()
	_, ws := activateWrite(t, srv)

	_, _ = ws.DeleteDocument(context.Background(), "movies", "hello world/slash")
	if !strings.Contains(requestPath, "hello") {
		t.Errorf("expected encoded path, got %q", requestPath)
	}
}

func TestDocumentWriteService_InvalidJSON(t *testing.T) {
	srv := newWriteSrv(t, http.MethodPost, "/indexes/", http.StatusAccepted, taskBody)
	defer srv.Close()
	_, ws := activateWrite(t, srv)

	_, err := ws.AddDocument(context.Background(), "movies", `{invalid`, "")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestDocumentWriteService_NonObjectJSON(t *testing.T) {
	srv := newWriteSrv(t, http.MethodPost, "/indexes/", http.StatusAccepted, taskBody)
	defer srv.Close()
	_, ws := activateWrite(t, srv)

	_, err := ws.AddDocument(context.Background(), "movies", `[{"id":1}]`, "")
	if err == nil {
		t.Error("expected error for array JSON")
	}
}

func TestDocumentWriteService_MissingPrimaryKey(t *testing.T) {
	srv := newWriteSrv(t, http.MethodPost, "/indexes/", http.StatusAccepted, taskBody)
	defer srv.Close()
	_, ws := activateWrite(t, srv)

	_, err := ws.AddDocument(context.Background(), "movies", `{"title":"No ID"}`, "id")
	if err == nil {
		t.Error("expected error for missing primary key")
	}
	if !strings.Contains(err.Error(), "id") {
		t.Errorf("expected error to mention field name, got %q", err.Error())
	}
}

func TestDocumentWriteService_NumericPrimaryKey(t *testing.T) {
	srv := newWriteSrv(t, http.MethodPost, "/indexes/movies/documents", http.StatusAccepted, taskBody)
	defer srv.Close()
	_, ws := activateWrite(t, srv)

	// Large integer that would lose precision through float64
	_, err := ws.AddDocument(context.Background(), "movies", `{"id":9007199254740993,"title":"BigID"}`, "id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDocumentWriteService_NestedJSON(t *testing.T) {
	srv := newWriteSrv(t, http.MethodPost, "/indexes/movies/documents", http.StatusAccepted, taskBody)
	defer srv.Close()
	_, ws := activateWrite(t, srv)

	doc := `{"id":1,"meta":{"tags":["action","drama"],"rating":{"score":8.5,"count":100}}}`
	_, err := ws.AddDocument(context.Background(), "movies", doc, "id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDocumentWriteService_UnicodeDocument(t *testing.T) {
	var received string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		received = string(body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(taskBody))
	}))
	defer srv.Close()
	_, ws := activateWrite(t, srv)

	doc := `{"id":1,"title":"日本語テスト","emoji":"🎬"}`
	_, _ = ws.AddDocument(context.Background(), "movies", doc, "id")
	if !strings.Contains(received, "日本語") || !strings.Contains(received, "🎬") {
		t.Errorf("unicode not preserved in payload, got: %q", received)
	}
}

func TestDocumentWriteService_EmptyIndexUID(t *testing.T) {
	srv := newWriteSrv(t, http.MethodPost, "/indexes/", http.StatusAccepted, taskBody)
	defer srv.Close()
	_, ws := activateWrite(t, srv)

	_, err := ws.AddDocument(context.Background(), "", `{"id":1}`, "")
	if err == nil {
		t.Error("expected error for empty index UID")
	}
}

func TestDocumentWriteService_EmptyDocumentID_Delete(t *testing.T) {
	srv := newWriteSrv(t, http.MethodDelete, "/indexes/", http.StatusAccepted, deleteTaskBody)
	defer srv.Close()
	_, ws := activateWrite(t, srv)

	_, err := ws.DeleteDocument(context.Background(), "movies", "")
	if err == nil {
		t.Error("expected error for empty document ID")
	}
}

func TestDocumentWriteService_HTTP401(t *testing.T) {
	srv := newWriteSrv(t, http.MethodPost, "/indexes/", http.StatusUnauthorized,
		`{"message":"invalid api key","code":"invalid_api_key","type":"auth_error","link":"https://docs.meilisearch.com"}`)
	defer srv.Close()
	_, ws := activateWrite(t, srv)

	_, err := ws.AddDocument(context.Background(), "movies", `{"id":1}`, "")
	if err == nil {
		t.Error("expected error for 401")
	}
	if !strings.Contains(err.Error(), "authentication") {
		t.Errorf("expected auth error message, got %q", err.Error())
	}
}

func TestDocumentWriteService_HTTP403(t *testing.T) {
	srv := newWriteSrv(t, http.MethodPost, "/indexes/", http.StatusForbidden,
		`{"message":"not allowed","code":"invalid_api_key","type":"auth_error","link":""}`)
	defer srv.Close()
	_, ws := activateWrite(t, srv)

	_, err := ws.AddDocument(context.Background(), "movies", `{"id":1}`, "")
	if err == nil {
		t.Error("expected error for 403")
	}
	if !strings.Contains(err.Error(), "access denied") {
		t.Errorf("expected access denied message, got %q", err.Error())
	}
}

func TestDocumentWriteService_HTTP404(t *testing.T) {
	srv := newWriteSrv(t, http.MethodPost, "/indexes/", http.StatusNotFound,
		`{"message":"index not found","code":"index_not_found","type":"invalid_request_error","link":""}`)
	defer srv.Close()
	_, ws := activateWrite(t, srv)

	_, err := ws.AddDocument(context.Background(), "movies", `{"id":1}`, "")
	if err == nil {
		t.Error("expected error for 404")
	}
}

func TestDocumentWriteService_HTTP409(t *testing.T) {
	srv := newWriteSrv(t, http.MethodPost, "/indexes/movies/documents", http.StatusConflict,
		`{"message":"conflict","code":"index_already_exists","type":"invalid_request_error","link":""}`)
	defer srv.Close()
	_, ws := activateWrite(t, srv)

	_, err := ws.AddDocument(context.Background(), "movies", `{"id":1}`, "")
	if err == nil {
		t.Error("expected error for 409")
	}
}

func TestDocumentWriteService_HTTP429(t *testing.T) {
	srv := newWriteSrv(t, http.MethodPost, "/indexes/", http.StatusTooManyRequests,
		`{"message":"too many requests","code":"too_many_requests","type":"server_error","link":""}`)
	defer srv.Close()
	_, ws := activateWrite(t, srv)

	_, err := ws.AddDocument(context.Background(), "movies", `{"id":1}`, "")
	if err == nil {
		t.Error("expected error for 429")
	}
}

func TestDocumentWriteService_HTTP500(t *testing.T) {
	srv := newWriteSrv(t, http.MethodPost, "/indexes/", http.StatusInternalServerError,
		`{"message":"internal error","code":"internal","type":"internal","link":""}`)
	defer srv.Close()
	_, ws := activateWrite(t, srv)

	_, err := ws.AddDocument(context.Background(), "movies", `{"id":1}`, "")
	if err == nil {
		t.Error("expected error for 500")
	}
}

func TestDocumentWriteService_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(10 * time.Second):
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	cs := newTestConnSvcWithTimeout(t, srv.URL, 100*time.Millisecond)
	ws := NewDocumentWriteService(cs)

	_, err := ws.AddDocument(context.Background(), "movies", `{"id":1}`, "")
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestDocumentWriteService_ContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(10 * time.Second):
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	_, ws := activateWrite(t, srv)

	ctx, cancel := context.WithCancel(context.Background())
	go func() { time.Sleep(50 * time.Millisecond); cancel() }()

	_, err := ws.AddDocument(ctx, "movies", `{"id":1}`, "")
	if err == nil {
		t.Error("expected cancellation error")
	}
}

func TestDocumentWriteService_MissingTaskUID(t *testing.T) {
	// Server responds without a taskUid field.
	srv := newWriteSrv(t, http.MethodPost, "/indexes/", http.StatusAccepted,
		`{"indexUid":"movies","status":"enqueued"}`)
	defer srv.Close()
	_, ws := activateWrite(t, srv)

	_, err := ws.AddDocument(context.Background(), "movies", `{"id":1}`, "")
	if err == nil {
		t.Error("expected error when taskUid is missing")
	}
}

func TestDocumentWriteService_MalformedTaskResponse(t *testing.T) {
	srv := newWriteSrv(t, http.MethodPost, "/indexes/", http.StatusAccepted, `{invalid json`)
	defer srv.Close()
	_, ws := activateWrite(t, srv)

	_, err := ws.AddDocument(context.Background(), "movies", `{"id":1}`, "")
	if err == nil {
		t.Error("expected error for malformed task response")
	}
}

func TestDocumentWriteService_NoActiveConnection(t *testing.T) {
	cs := &ConnectionService{}
	ws := NewDocumentWriteService(cs)

	_, err := ws.AddDocument(context.Background(), "movies", `{"id":1}`, "")
	if err == nil {
		t.Error("expected error with no active connection")
	}
}

func TestDocumentWriteService_BodyPreservesLargeNumbers(t *testing.T) {
	var received string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		received = string(body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(taskBody))
	}))
	defer srv.Close()
	_, ws := activateWrite(t, srv)

	largeNum := "9007199254740993"
	doc := `{"id":` + largeNum + `,"val":true}`
	_, _ = ws.AddDocument(context.Background(), "movies", doc, "")

	if !strings.Contains(received, largeNum) {
		t.Errorf("large number not preserved: got %q", received)
	}
}

func TestDocumentWriteService_TaskReferenceFields(t *testing.T) {
	body := `{"taskUid":99,"indexUid":"books","status":"enqueued","type":"documentDeletion","enqueuedAt":"2024-06-15T12:00:00.000000Z"}`
	srv := newWriteSrv(t, http.MethodDelete, "/indexes/books/documents/", http.StatusAccepted, body)
	defer srv.Close()
	_, ws := activateWrite(t, srv)

	ref, err := ws.DeleteDocument(context.Background(), "books", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.TaskUID != 99 {
		t.Errorf("TaskUID: got %d, want 99", ref.TaskUID)
	}
	if ref.IndexUID != "books" {
		t.Errorf("IndexUID: got %q, want books", ref.IndexUID)
	}
	if ref.Type != "documentDeletion" {
		t.Errorf("Type: got %q, want documentDeletion", ref.Type)
	}
	if ref.EnqueuedAt == "" {
		t.Error("EnqueuedAt should not be empty")
	}
}

// --- helper: test connection service with custom timeout ---

func newTestConnSvcWithTimeout(t *testing.T, serverURL string, timeout time.Duration) *ConnectionService {
	t.Helper()
	creds := newMockCreds()
	cs := NewConnectionService(newTestManager(t), creds)
	ctx := context.Background()
	sum, err := cs.CreateConnection(ctx, SaveConnectionRequest{Name: "Test", URL: serverURL})
	if err != nil {
		t.Fatalf("CreateConnection: %v", err)
	}
	if _, err := cs.SelectConnection(ctx, sum.ID); err != nil {
		t.Fatalf("SelectConnection: %v", err)
	}
	// Replace the client with one that has a short timeout.
	client, err := meilisearch.New(meilisearch.Config{BaseURL: serverURL, Timeout: timeout})
	if err != nil {
		t.Fatalf("build client: %v", err)
	}
	cs.mu.Lock()
	if cs.active != nil {
		cs.active.client = client
	}
	cs.mu.Unlock()
	return cs
}

// parseTaskReferenceTest is a helper for white-box parsing tests.
func parseTaskReferenceTest(raw []byte) (TaskReference, error) {
	return parseTaskReference(json.RawMessage(raw))
}

func TestParseTaskReference_Valid(t *testing.T) {
	ref, err := parseTaskReferenceTest([]byte(taskBody))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.TaskUID != 42 {
		t.Errorf("TaskUID: got %d, want 42", ref.TaskUID)
	}
}

func TestParseTaskReference_MissingUID(t *testing.T) {
	_, err := parseTaskReferenceTest([]byte(`{"indexUid":"movies","status":"enqueued"}`))
	if err == nil {
		t.Error("expected error when taskUid is 0/missing")
	}
}

func TestParseTaskReference_InvalidJSON(t *testing.T) {
	_, err := parseTaskReferenceTest([]byte(`{bad`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
