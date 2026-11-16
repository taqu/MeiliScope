package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// taskResultBody templates for various statuses.
const (
	taskEnqueuedBody    = `{"uid":10,"indexUid":"movies","status":"enqueued","type":"documentAdditionOrUpdate","enqueuedAt":"2024-01-01T00:00:00.000000Z"}`
	taskProcessingBody  = `{"uid":10,"indexUid":"movies","status":"processing","type":"documentAdditionOrUpdate","enqueuedAt":"2024-01-01T00:00:00.000000Z","startedAt":"2024-01-01T00:00:00.100000Z"}`
	taskSucceededBody   = `{"uid":10,"indexUid":"movies","status":"succeeded","type":"documentAdditionOrUpdate","duration":"PT0.050S","enqueuedAt":"2024-01-01T00:00:00.000000Z","startedAt":"2024-01-01T00:00:00.100000Z","finishedAt":"2024-01-01T00:00:00.150000Z","details":{"receivedDocuments":1,"indexedDocuments":1}}`
	taskFailedBody      = `{"uid":10,"indexUid":"movies","status":"failed","type":"documentAdditionOrUpdate","error":{"message":"primary key mismatch","code":"invalid_document_id","type":"invalid_request","link":"https://docs.meilisearch.com"},"duration":"PT0.010S","enqueuedAt":"2024-01-01T00:00:00.000000Z","startedAt":"2024-01-01T00:00:00.100000Z","finishedAt":"2024-01-01T00:00:00.110000Z"}`
	taskCanceledBody    = `{"uid":10,"indexUid":"movies","status":"canceled","type":"documentAdditionOrUpdate","enqueuedAt":"2024-01-01T00:00:00.000000Z"}`
)

func activateTask(t *testing.T, srv *httptest.Server) (*ConnectionService, *TaskService) {
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
	return cs, NewTaskService(cs)
}

func newTaskSrv(t *testing.T, _ int64, statusCode int, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if !strings.HasPrefix(r.URL.Path, "/tasks/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(body))
	}))
}

// --- status parsing ---

func TestTaskService_GetTask_Enqueued(t *testing.T) {
	srv := newTaskSrv(t, 10, http.StatusOK, taskEnqueuedBody)
	defer srv.Close()
	_, ts := activateTask(t, srv)

	res, err := ts.GetTask(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.UID != 10 {
		t.Errorf("UID: got %d, want 10", res.UID)
	}
	if res.Status != "enqueued" {
		t.Errorf("Status: got %q, want enqueued", res.Status)
	}
	if res.IndexUID != "movies" {
		t.Errorf("IndexUID: got %q, want movies", res.IndexUID)
	}
	if res.IsTerminal() {
		t.Error("enqueued task should not be terminal")
	}
}

func TestTaskService_GetTask_Processing(t *testing.T) {
	srv := newTaskSrv(t, 10, http.StatusOK, taskProcessingBody)
	defer srv.Close()
	_, ts := activateTask(t, srv)

	res, err := ts.GetTask(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != "processing" {
		t.Errorf("Status: got %q, want processing", res.Status)
	}
	if res.StartedAt == "" {
		t.Error("StartedAt should be set for processing task")
	}
	if res.IsTerminal() {
		t.Error("processing task should not be terminal")
	}
}

func TestTaskService_GetTask_Succeeded(t *testing.T) {
	srv := newTaskSrv(t, 10, http.StatusOK, taskSucceededBody)
	defer srv.Close()
	_, ts := activateTask(t, srv)

	res, err := ts.GetTask(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != "succeeded" {
		t.Errorf("Status: got %q, want succeeded", res.Status)
	}
	if res.Duration == "" {
		t.Error("Duration should be set for succeeded task")
	}
	if res.StartedAt == "" {
		t.Error("StartedAt should be set for succeeded task")
	}
	if res.FinishedAt == "" {
		t.Error("FinishedAt should be set for succeeded task")
	}
	if !res.IsTerminal() {
		t.Error("succeeded task should be terminal")
	}
}

func TestTaskService_GetTask_Failed(t *testing.T) {
	srv := newTaskSrv(t, 10, http.StatusOK, taskFailedBody)
	defer srv.Close()
	_, ts := activateTask(t, srv)

	res, err := ts.GetTask(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != "failed" {
		t.Errorf("Status: got %q, want failed", res.Status)
	}
	if res.Error == nil {
		t.Fatal("Error field should be populated for failed task")
	}
	if res.Error.Code != "invalid_document_id" {
		t.Errorf("Error.Code: got %q, want invalid_document_id", res.Error.Code)
	}
	if !res.IsTerminal() {
		t.Error("failed task should be terminal")
	}
}

func TestTaskService_GetTask_Canceled(t *testing.T) {
	srv := newTaskSrv(t, 10, http.StatusOK, taskCanceledBody)
	defer srv.Close()
	_, ts := activateTask(t, srv)

	res, err := ts.GetTask(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != "canceled" {
		t.Errorf("Status: got %q, want canceled", res.Status)
	}
	if !res.IsTerminal() {
		t.Error("canceled task should be terminal")
	}
}

// --- RawJSON preservation ---

func TestTaskService_GetTask_RawJSONPreserved(t *testing.T) {
	srv := newTaskSrv(t, 10, http.StatusOK, taskSucceededBody)
	defer srv.Close()
	_, ts := activateTask(t, srv)

	res, err := ts.GetTask(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RawJSON != taskSucceededBody {
		t.Errorf("RawJSON not preserved:\ngot:  %s\nwant: %s", res.RawJSON, taskSucceededBody)
	}
}

// --- HTTP error classification ---

func TestTaskService_GetTask_Forbidden(t *testing.T) {
	srv := newWriteSrv(t, http.MethodGet, "/tasks/", http.StatusForbidden,
		`{"message":"tasks.get required","code":"missing_authorization_header","type":"auth","link":""}`)
	defer srv.Close()
	_, ts := activateTask(t, srv)

	_, err := ts.GetTask(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error for 403")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("expected permission denied in error, got: %v", err)
	}
}

func TestTaskService_GetTask_NotFound(t *testing.T) {
	srv := newWriteSrv(t, http.MethodGet, "/tasks/", http.StatusNotFound,
		`{"message":"task not found","code":"task_not_found","type":"invalid_request","link":""}`)
	defer srv.Close()
	_, ts := activateTask(t, srv)

	_, err := ts.GetTask(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected not found in error, got: %v", err)
	}
}

func TestTaskService_GetTask_Unauthorized(t *testing.T) {
	srv := newWriteSrv(t, http.MethodGet, "/tasks/", http.StatusUnauthorized,
		`{"message":"The Authorization header is missing","code":"missing_authorization_header","type":"auth","link":""}`)
	defer srv.Close()
	_, ts := activateTask(t, srv)

	_, err := ts.GetTask(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error for 401")
	}
	if !strings.Contains(err.Error(), "authentication required") {
		t.Errorf("expected authentication required in error, got: %v", err)
	}
}

func TestTaskService_GetTask_ServerError(t *testing.T) {
	srv := newWriteSrv(t, http.MethodGet, "/tasks/", http.StatusInternalServerError,
		`{"message":"internal error","code":"internal","type":"internal","link":""}`)
	defer srv.Close()
	_, ts := activateTask(t, srv)

	_, err := ts.GetTask(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error for 500")
	}
}

// --- pre-condition guards ---

func TestTaskService_GetTask_InvalidUID(t *testing.T) {
	creds := newMockCreds()
	mgr := newTestManager(t)
	connSvc := NewConnectionService(mgr, creds)
	ts := NewTaskService(connSvc)

	_, err := ts.GetTask(context.Background(), 0)
	if err == nil {
		t.Fatal("expected error for UID=0")
	}
	if !strings.Contains(err.Error(), "positive integer") {
		t.Errorf("expected 'positive integer' in error, got: %v", err)
	}
}

func TestTaskService_GetTask_NegativeUID(t *testing.T) {
	creds := newMockCreds()
	mgr := newTestManager(t)
	connSvc := NewConnectionService(mgr, creds)
	ts := NewTaskService(connSvc)

	_, err := ts.GetTask(context.Background(), -5)
	if err == nil {
		t.Fatal("expected error for negative UID")
	}
}

func TestTaskService_GetTask_NoConnection(t *testing.T) {
	creds := newMockCreds()
	mgr := newTestManager(t)
	connSvc := NewConnectionService(mgr, creds)
	ts := NewTaskService(connSvc)

	_, err := ts.GetTask(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error when no connection")
	}
	if !strings.Contains(err.Error(), "no active connection") {
		t.Errorf("expected 'no active connection' in error, got: %v", err)
	}
}

// --- timeout and cancellation ---

func TestTaskService_GetTask_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(10 * time.Second):
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cs := newTestConnSvcWithTimeout(t, srv.URL, 50*time.Millisecond)
	ts := NewTaskService(cs)

	_, err := ts.GetTask(context.Background(), 10)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestTaskService_GetTask_ContextCanceled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(10 * time.Second):
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	_, ts := activateTask(t, srv)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := ts.GetTask(ctx, 10)
	if err == nil {
		t.Fatal("expected error for canceled context")
	}
}

// --- malformed responses ---

func TestTaskService_GetTask_MalformedJSON(t *testing.T) {
	srv := newWriteSrv(t, http.MethodGet, "/tasks/", http.StatusOK, `{not valid json`)
	defer srv.Close()
	_, ts := activateTask(t, srv)

	_, err := ts.GetTask(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

// --- IsTerminal ---

func TestTaskResult_IsTerminal(t *testing.T) {
	cases := []struct {
		status   string
		terminal bool
	}{
		{"enqueued", false},
		{"processing", false},
		{"succeeded", true},
		{"failed", true},
		{"canceled", true},
		{"unknown", false},
	}
	for _, tc := range cases {
		r := TaskResult{Status: tc.status}
		if got := r.IsTerminal(); got != tc.terminal {
			t.Errorf("IsTerminal(%q) = %v, want %v", tc.status, got, tc.terminal)
		}
	}
}

// --- optional fields ---

func TestTaskService_GetTask_NullOptionalFields(t *testing.T) {
	body := `{"uid":5,"indexUid":"books","status":"enqueued","type":"documentAdditionOrUpdate","enqueuedAt":"2024-06-01T00:00:00Z"}`
	srv := newWriteSrv(t, http.MethodGet, "/tasks/", http.StatusOK, body)
	defer srv.Close()
	_, ts := activateTask(t, srv)

	res, err := ts.GetTask(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Duration != "" {
		t.Errorf("Duration should be empty, got %q", res.Duration)
	}
	if res.StartedAt != "" {
		t.Errorf("StartedAt should be empty, got %q", res.StartedAt)
	}
	if res.FinishedAt != "" {
		t.Errorf("FinishedAt should be empty, got %q", res.FinishedAt)
	}
	if res.Error != nil {
		t.Errorf("Error should be nil, got %+v", res.Error)
	}
}
