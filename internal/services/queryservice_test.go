package services

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// activateQuery creates a ConnectionService connected to srv and a QueryService on top.
func activateQuery(t *testing.T, srv *httptest.Server) (*ConnectionService, *QueryService) {
	t.Helper()
	creds := newMockCreds()
	mgr := newTestManager(t)
	connSvc := NewConnectionService(mgr, creds)
	ctx := context.Background()

	sum, err := connSvc.CreateConnection(ctx, SaveConnectionRequest{Name: "Test", URL: srv.URL})
	if err != nil {
		t.Fatalf("CreateConnection: %v", err)
	}
	if _, err := connSvc.SelectConnection(ctx, sum.ID); err != nil {
		t.Fatalf("SelectConnection: %v", err)
	}
	return connSvc, NewQueryService(connSvc, mgr)
}

// searchBody is a standard mock search response.
const searchBody = `{"hits":[{"id":1,"title":"Inception"},{"id":2,"title":"Interstellar"}],"query":"inception","processingTimeMs":4,"limit":20,"offset":0,"estimatedTotalHits":2}`
const emptyBody = `{"hits":[],"query":"","processingTimeMs":1,"limit":20,"offset":0,"estimatedTotalHits":0}`

func searchSrv(t *testing.T, uid, body string) *httptest.Server {
	t.Helper()
	return newSrv(t, handler{
		"/indexes/" + uid + "/search": func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(body))
		},
	})
}

func readBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	return io.ReadAll(r.Body)
}

// --- pre-condition errors ---

func TestQueryService_ExecuteSearch_NoConnection(t *testing.T) {
	creds := newMockCreds()
	mgr := newTestManager(t)
	connSvc := NewConnectionService(mgr, creds)
	qSvc := NewQueryService(connSvc, mgr)

	_, err := qSvc.ExecuteSearch(context.Background(), "movies", `{}`)
	if err == nil {
		t.Error("expected error when no active connection")
	}
}

func TestQueryService_ExecuteSearch_EmptyIndex(t *testing.T) {
	_, qSvc := activateQuery(t, searchSrv(t, "movies", searchBody))
	_, err := qSvc.ExecuteSearch(context.Background(), "", `{}`)
	if err == nil || !containsStr(err.Error(), "index UID") {
		t.Errorf("expected index UID error, got: %v", err)
	}
}

func TestQueryService_ExecuteSearch_InvalidJSON(t *testing.T) {
	_, qSvc := activateQuery(t, searchSrv(t, "movies", searchBody))
	_, err := qSvc.ExecuteSearch(context.Background(), "movies", `not json`)
	if err == nil || !containsStr(err.Error(), "invalid JSON") {
		t.Errorf("expected invalid JSON error, got: %v", err)
	}
}

func TestQueryService_ExecuteSearch_NonObjectJSON(t *testing.T) {
	_, qSvc := activateQuery(t, searchSrv(t, "movies", searchBody))
	_, err := qSvc.ExecuteSearch(context.Background(), "movies", `[1,2,3]`)
	if err == nil || !containsStr(err.Error(), "JSON object") {
		t.Errorf("expected object error, got: %v", err)
	}
}

// --- successful searches ---

func TestQueryService_ExecuteSearch_Basic(t *testing.T) {
	_, qSvc := activateQuery(t, searchSrv(t, "movies", searchBody))
	res, err := qSvc.ExecuteSearch(context.Background(), "movies", `{"q":"inception","limit":20}`)
	if err != nil {
		t.Fatalf("ExecuteSearch: %v", err)
	}
	if res.Error != nil {
		t.Fatalf("unexpected SearchError: %s", res.Error.Message)
	}
	if len(res.Hits) != 2 {
		t.Errorf("Hits: got %d, want 2", len(res.Hits))
	}
	if res.Meta.ProcessingTimeMs != 4 {
		t.Errorf("ProcessingTimeMs: got %d, want 4", res.Meta.ProcessingTimeMs)
	}
	if res.Meta.Query != "inception" {
		t.Errorf("Meta.Query: got %q, want %q", res.Meta.Query, "inception")
	}
	if res.SubmittedQuery == "" {
		t.Error("SubmittedQuery must be preserved")
	}
	if res.ClientDurationMs < 0 {
		t.Error("ClientDurationMs must not be negative")
	}
}

func TestQueryService_ExecuteSearch_EmptyObject(t *testing.T) {
	_, qSvc := activateQuery(t, searchSrv(t, "movies", emptyBody))
	res, err := qSvc.ExecuteSearch(context.Background(), "movies", `{}`)
	if err != nil || res.Error != nil {
		t.Fatalf("unexpected error: err=%v res=%v", err, res.Error)
	}
	if len(res.Hits) != 0 {
		t.Errorf("expected 0 hits, got %d", len(res.Hits))
	}
}

func TestQueryService_ExecuteSearch_WithFilters(t *testing.T) {
	var gotBody []byte
	srv := newSrv(t, handler{
		"/indexes/movies/search": func(w http.ResponseWriter, r *http.Request) {
			gotBody, _ = readBody(r)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(emptyBody))
		},
	})
	_, qSvc := activateQuery(t, srv)
	_, err := qSvc.ExecuteSearch(context.Background(), "movies", `{"q":"test","filter":"category = 'action'"}`)
	if err != nil {
		t.Fatalf("ExecuteSearch: %v", err)
	}
	if !containsStr(string(gotBody), "filter") {
		t.Errorf("filter not sent to server: %s", gotBody)
	}
}

func TestQueryService_ExecuteSearch_WithFacets(t *testing.T) {
	body := `{"hits":[],"query":"","processingTimeMs":2,"limit":20,"offset":0,"estimatedTotalHits":0,` +
		`"facetDistribution":{"category":{"action":10,"drama":5},"year":{"2020":3}}}`
	_, qSvc := activateQuery(t, searchSrv(t, "movies", body))
	res, err := qSvc.ExecuteSearch(context.Background(), "movies", `{"facets":["category","year"]}`)
	if err != nil || res.Error != nil {
		t.Fatalf("unexpected error: err=%v res=%v", err, res.Error)
	}
	if res.Meta.FacetDistribution == nil {
		t.Fatal("FacetDistribution must not be nil")
	}
	if len(res.Meta.FacetDistribution["category"]) != 2 {
		t.Errorf("FacetDistribution[category]: got %d, want 2", len(res.Meta.FacetDistribution["category"]))
	}
}

func TestQueryService_ExecuteSearch_WithRankingScore(t *testing.T) {
	body := `{"hits":[{"id":1,"title":"Inception","_rankingScore":0.9}],"query":"inception","processingTimeMs":3,"limit":20,"offset":0,"estimatedTotalHits":1}`
	_, qSvc := activateQuery(t, searchSrv(t, "movies", body))
	res, err := qSvc.ExecuteSearch(context.Background(), "movies", `{"q":"inception","showRankingScore":true}`)
	if err != nil || res.Error != nil {
		t.Fatalf("unexpected error: err=%v res=%v", err, res.Error)
	}
	if len(res.Hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(res.Hits))
	}
	scoreRaw, ok := res.Hits[0].Fields["_rankingScore"]
	if !ok {
		t.Error("expected _rankingScore in hit fields")
	}
	if scoreRaw != "0.9" {
		t.Errorf("_rankingScore: got %q, want %q", scoreRaw, "0.9")
	}
}

func TestQueryService_ExecuteSearch_UnknownParams(t *testing.T) {
	var gotBody []byte
	srv := newSrv(t, handler{
		"/indexes/movies/search": func(w http.ResponseWriter, r *http.Request) {
			gotBody, _ = readBody(r)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(emptyBody))
		},
	})
	_, qSvc := activateQuery(t, srv)
	_, err := qSvc.ExecuteSearch(context.Background(), "movies", `{"q":"test","unknownFutureParam":"value"}`)
	if err != nil {
		t.Fatalf("ExecuteSearch: %v", err)
	}
	if !containsStr(string(gotBody), "unknownFutureParam") {
		t.Error("unknown params must be forwarded to the server unmodified")
	}
}

func TestQueryService_ExecuteSearch_LargeNumbers(t *testing.T) {
	const largeNum = `9007199254740993`
	body := `{"hits":[{"id":` + largeNum + `}],"query":"","processingTimeMs":1,"limit":20,"offset":0,"estimatedTotalHits":1}`
	_, qSvc := activateQuery(t, searchSrv(t, "docs", body))
	res, err := qSvc.ExecuteSearch(context.Background(), "docs", `{}`)
	if err != nil || res.Error != nil {
		t.Fatalf("unexpected error: err=%v res=%v", err, res.Error)
	}
	idRaw := res.Hits[0].Fields["id"]
	if idRaw != largeNum {
		t.Errorf("large integer precision: got %q, want %q", idRaw, largeNum)
	}
}

func TestQueryService_ExecuteSearch_RawJSONPreserved(t *testing.T) {
	_, qSvc := activateQuery(t, searchSrv(t, "movies", searchBody))
	res, err := qSvc.ExecuteSearch(context.Background(), "movies", `{}`)
	if err != nil || res.Error != nil {
		t.Fatalf("unexpected error: err=%v res=%v", err, res.Error)
	}
	if res.RawJSON == "" {
		t.Error("RawJSON must not be empty")
	}
	if !containsStr(res.RawJSON, "processingTimeMs") {
		t.Error("RawJSON must contain complete response")
	}
}

func TestQueryService_ExecuteSearch_PageBasedPagination(t *testing.T) {
	body := `{"hits":[],"query":"","processingTimeMs":1,"page":2,"hitsPerPage":10,"totalPages":5,"totalHits":50}`
	_, qSvc := activateQuery(t, searchSrv(t, "movies", body))
	res, err := qSvc.ExecuteSearch(context.Background(), "movies", `{"page":2,"hitsPerPage":10}`)
	if err != nil || res.Error != nil {
		t.Fatalf("unexpected error: err=%v res=%v", err, res.Error)
	}
	if res.Meta.Page == nil || *res.Meta.Page != 2 {
		t.Errorf("Meta.Page: got %v, want 2", res.Meta.Page)
	}
	if res.Meta.TotalPages == nil || *res.Meta.TotalPages != 5 {
		t.Errorf("Meta.TotalPages: got %v, want 5", res.Meta.TotalPages)
	}
	if res.Meta.TotalHits == nil || *res.Meta.TotalHits != 50 {
		t.Errorf("Meta.TotalHits: got %v, want 50", res.Meta.TotalHits)
	}
}

func TestQueryService_ExecuteSearch_ExecutedAtSet(t *testing.T) {
	_, qSvc := activateQuery(t, searchSrv(t, "movies", emptyBody))
	res, err := qSvc.ExecuteSearch(context.Background(), "movies", `{}`)
	if err != nil || res.Error != nil {
		t.Fatalf("unexpected error: err=%v res=%v", err, res.Error)
	}
	if res.ExecutedAt == "" {
		t.Error("ExecutedAt must be set")
	}
}

func TestQueryService_ExecuteSearch_PathEscaping(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/my%20index/search": func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(emptyBody))
		},
	})
	_, qSvc := activateQuery(t, srv)
	res, err := qSvc.ExecuteSearch(context.Background(), "my index", `{}`)
	if err != nil || res.Error != nil {
		t.Fatalf("unexpected error for space in UID: err=%v res=%v", err, res.Error)
	}
}

// --- HTTP error codes ---

func TestQueryService_ExecuteSearch_HTTP400(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/movies/search": func(w http.ResponseWriter, _ *http.Request) {
			writeJSONStatus(w, http.StatusBadRequest, map[string]string{
				"message": "invalid filter expression",
				"code":    "invalid_filter_expression",
				"type":    "invalid_request_error",
				"link":    "https://docs.meilisearch.com",
			})
		},
	})
	_, qSvc := activateQuery(t, srv)
	res, err := qSvc.ExecuteSearch(context.Background(), "movies", `{"filter":"bad syntax"}`)
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if res.Error == nil {
		t.Fatal("expected SearchError for HTTP 400")
	}
	if res.Error.Kind != "bad_request" {
		t.Errorf("Kind: got %q, want bad_request", res.Error.Kind)
	}
	if res.Error.Code != "invalid_filter_expression" {
		t.Errorf("Code: got %q, want invalid_filter_expression", res.Error.Code)
	}
	if res.Error.Link != "https://docs.meilisearch.com" {
		t.Errorf("Link: got %q", res.Error.Link)
	}
}

func TestQueryService_ExecuteSearch_HTTP401(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/movies/search": func(w http.ResponseWriter, _ *http.Request) {
			writeJSONStatus(w, http.StatusUnauthorized, map[string]string{
				"message": "authentication required", "code": "missing_authorization_header", "type": "auth_error", "link": "",
			})
		},
	})
	_, qSvc := activateQuery(t, srv)
	res, err := qSvc.ExecuteSearch(context.Background(), "movies", `{}`)
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if res.Error == nil || res.Error.Kind != "auth" {
		t.Errorf("expected auth error, got: %v", res.Error)
	}
}

func TestQueryService_ExecuteSearch_HTTP403(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/secret/search": func(w http.ResponseWriter, _ *http.Request) {
			writeJSONStatus(w, http.StatusForbidden, map[string]string{
				"message": "access denied", "code": "invalid_api_key", "type": "auth_error", "link": "",
			})
		},
	})
	_, qSvc := activateQuery(t, srv)
	res, err := qSvc.ExecuteSearch(context.Background(), "secret", `{}`)
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if res.Error == nil || res.Error.Kind != "forbidden" {
		t.Errorf("expected forbidden error, got: %v", res.Error)
	}
}

func TestQueryService_ExecuteSearch_HTTP404(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/gone/search": func(w http.ResponseWriter, _ *http.Request) {
			writeJSONStatus(w, http.StatusNotFound, map[string]string{
				"message": "index not found", "code": "index_not_found", "type": "invalid_request_error", "link": "",
			})
		},
	})
	_, qSvc := activateQuery(t, srv)
	res, err := qSvc.ExecuteSearch(context.Background(), "gone", `{}`)
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if res.Error == nil || res.Error.Kind != "not_found" {
		t.Errorf("expected not_found error, got: %v", res.Error)
	}
}

func TestQueryService_ExecuteSearch_HTTP429(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/movies/search": func(w http.ResponseWriter, _ *http.Request) {
			writeJSONStatus(w, http.StatusTooManyRequests, map[string]string{
				"message": "too many requests", "code": "too_many_requests", "type": "rate_limit_error", "link": "",
			})
		},
	})
	_, qSvc := activateQuery(t, srv)
	res, err := qSvc.ExecuteSearch(context.Background(), "movies", `{}`)
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if res.Error == nil || res.Error.Kind != "rate_limit" {
		t.Errorf("expected rate_limit error, got: %v", res.Error)
	}
}

func TestQueryService_ExecuteSearch_HTTP500(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/movies/search": func(w http.ResponseWriter, _ *http.Request) {
			writeJSONStatus(w, http.StatusInternalServerError, map[string]string{
				"message": "internal error", "code": "internal", "type": "internal_error", "link": "",
			})
		},
	})
	_, qSvc := activateQuery(t, srv)
	res, err := qSvc.ExecuteSearch(context.Background(), "movies", `{}`)
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if res.Error == nil || res.Error.Kind != "server" {
		t.Errorf("expected server error, got: %v", res.Error)
	}
}

func TestQueryService_ExecuteSearch_Timeout(t *testing.T) {
	// Handler uses select so srv.Close() doesn't block if r.Context() is slow to cancel.
	srv := newSrv(t, handler{
		"/indexes/slow/search": func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
			case <-time.After(10 * time.Second):
			}
			w.WriteHeader(http.StatusServiceUnavailable)
		},
	})
	_, qSvc := activateQuery(t, srv)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	res, err := qSvc.ExecuteSearch(ctx, "slow", `{}`)
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if res.Error == nil {
		t.Error("expected SearchError for timeout")
	}
	if res.Error.Kind != "timeout" && res.Error.Kind != "canceled" {
		t.Errorf("expected timeout or canceled kind, got %q", res.Error.Kind)
	}
}

func TestQueryService_ExecuteSearch_Canceled(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/slow/search": func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
			case <-time.After(10 * time.Second):
			}
			w.WriteHeader(http.StatusServiceUnavailable)
		},
	})
	_, qSvc := activateQuery(t, srv)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	res, err := qSvc.ExecuteSearch(ctx, "slow", `{}`)
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if res.Error == nil {
		t.Error("expected SearchError for canceled context")
	}
}

func TestQueryService_CancelSearch(t *testing.T) {
	started := make(chan struct{})
	srv := newSrv(t, handler{
		"/indexes/slow/search": func(w http.ResponseWriter, r *http.Request) {
			close(started)
			select {
			case <-r.Context().Done():
			case <-time.After(10 * time.Second):
			}
			w.WriteHeader(http.StatusServiceUnavailable)
		},
	})
	_, qSvc := activateQuery(t, srv)

	done := make(chan SearchResult, 1)
	go func() {
		res, _ := qSvc.ExecuteSearch(context.Background(), "slow", `{}`)
		done <- res
	}()

	<-started
	qSvc.CancelSearch()
	res := <-done
	if res.Error == nil {
		t.Error("expected SearchError after CancelSearch")
	}
}

func TestQueryService_ExecuteSearch_MalformedResponse(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/bad/search": func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`not valid json`))
		},
	})
	_, qSvc := activateQuery(t, srv)
	res, err := qSvc.ExecuteSearch(context.Background(), "bad", `{}`)
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if res.Error == nil {
		t.Error("expected SearchError for malformed response")
	}
}

// --- history ---

func TestQueryService_History_AddedOnSuccess(t *testing.T) {
	_, qSvc := activateQuery(t, searchSrv(t, "movies", emptyBody))
	_, _ = qSvc.ExecuteSearch(context.Background(), "movies", `{}`)
	hist := qSvc.GetHistory()
	if len(hist) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(hist))
	}
	if hist[0].Status != "success" {
		t.Errorf("Status: got %q, want success", hist[0].Status)
	}
	if hist[0].IndexUID != "movies" {
		t.Errorf("IndexUID: got %q, want movies", hist[0].IndexUID)
	}
}

func TestQueryService_History_AddedOnAPIError(t *testing.T) {
	srv := newSrv(t, handler{
		"/indexes/movies/search": func(w http.ResponseWriter, _ *http.Request) {
			writeJSONStatus(w, http.StatusUnauthorized, map[string]string{
				"message": "auth failed", "code": "auth", "type": "auth_error", "link": "",
			})
		},
	})
	_, qSvc := activateQuery(t, srv)
	res, err := qSvc.ExecuteSearch(context.Background(), "movies", `{}`)
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if res.Error == nil {
		t.Fatal("expected SearchError")
	}
	hist := qSvc.GetHistory()
	if len(hist) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(hist))
	}
	if hist[0].Status != "error" {
		t.Errorf("Status: got %q, want error", hist[0].Status)
	}
}

func TestQueryService_History_NewestFirst(t *testing.T) {
	_, qSvc := activateQuery(t, searchSrv(t, "movies", emptyBody))
	for i := 0; i < 3; i++ {
		_, _ = qSvc.ExecuteSearch(context.Background(), "movies", `{}`)
	}
	if len(qSvc.GetHistory()) != 3 {
		t.Errorf("expected 3 history entries")
	}
}

func TestQueryService_History_RetentionLimit(t *testing.T) {
	_, qSvc := activateQuery(t, searchSrv(t, "movies", emptyBody))
	for i := 0; i <= maxHistoryItems+10; i++ {
		_, _ = qSvc.ExecuteSearch(context.Background(), "movies", `{}`)
	}
	hist := qSvc.GetHistory()
	if len(hist) > maxHistoryItems {
		t.Errorf("history exceeds limit: got %d, want ≤ %d", len(hist), maxHistoryItems)
	}
}

func TestQueryService_History_QueryTruncatedInHistory(t *testing.T) {
	bigQuery := `{"q":"` + string(make([]byte, maxQueryBytes+1000)) + `"}`
	_, qSvc := activateQuery(t, searchSrv(t, "movies", emptyBody))
	_, _ = qSvc.ExecuteSearch(context.Background(), "movies", bigQuery)
	hist := qSvc.GetHistory()
	if len(hist) > 0 && len(hist[0].QueryJSON) > maxQueryBytes+10 {
		t.Error("QueryJSON should be truncated in history")
	}
}

func TestQueryService_History_Delete(t *testing.T) {
	_, qSvc := activateQuery(t, searchSrv(t, "movies", emptyBody))
	_, _ = qSvc.ExecuteSearch(context.Background(), "movies", `{}`)
	hist := qSvc.GetHistory()
	if len(hist) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(hist))
	}
	if err := qSvc.DeleteHistoryEntry(hist[0].ID); err != nil {
		t.Fatalf("DeleteHistoryEntry: %v", err)
	}
	if len(qSvc.GetHistory()) != 0 {
		t.Error("expected empty history after delete")
	}
}

func TestQueryService_History_DeleteNotFound(t *testing.T) {
	_, qSvc := activateQuery(t, searchSrv(t, "movies", emptyBody))
	if err := qSvc.DeleteHistoryEntry("nonexistent"); err == nil {
		t.Error("expected error for nonexistent entry")
	}
}

func TestQueryService_History_Clear(t *testing.T) {
	_, qSvc := activateQuery(t, searchSrv(t, "movies", emptyBody))
	for i := 0; i < 3; i++ {
		_, _ = qSvc.ExecuteSearch(context.Background(), "movies", `{}`)
	}
	if err := qSvc.ClearHistory(); err != nil {
		t.Fatalf("ClearHistory: %v", err)
	}
	if len(qSvc.GetHistory()) != 0 {
		t.Error("expected empty history after clear")
	}
}

func TestQueryService_History_Persistence(t *testing.T) {
	creds := newMockCreds()
	mgr := newTestManager(t)
	connSvc := NewConnectionService(mgr, creds)
	ctx := context.Background()

	srv := searchSrv(t, "movies", emptyBody)
	sum, _ := connSvc.CreateConnection(ctx, SaveConnectionRequest{Name: "T", URL: srv.URL})
	_, _ = connSvc.SelectConnection(ctx, sum.ID)

	qSvc := NewQueryService(connSvc, mgr)
	_, _ = qSvc.ExecuteSearch(ctx, "movies", `{}`)
	_, _ = qSvc.ExecuteSearch(ctx, "movies", `{"q":"hello"}`)

	qSvc2 := NewQueryService(connSvc, mgr)
	hist := qSvc2.GetHistory()
	if len(hist) != 2 {
		t.Errorf("expected 2 history entries after reload, got %d", len(hist))
	}
}

func TestQueryService_History_CorruptedFile(t *testing.T) {
	creds := newMockCreds()
	mgr := newTestManager(t)
	connSvc := NewConnectionService(mgr, creds)

	histPath := filepath.Join(mgr.Dir(), historyFileName)
	if err := os.WriteFile(histPath, []byte("not json at all"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	qSvc := NewQueryService(connSvc, mgr)
	if len(qSvc.GetHistory()) != 0 {
		t.Error("expected empty history after corruption recovery")
	}
	if _, err := os.Stat(histPath + ".corrupted"); err != nil {
		t.Error("expected .corrupted backup file to exist")
	}
}

func TestQueryService_History_NoAPIKeys(t *testing.T) {
	creds := newMockCreds()
	mgr := newTestManager(t)
	connSvc := NewConnectionService(mgr, creds)
	ctx := context.Background()

	srv := searchSrv(t, "movies", emptyBody)
	sum, _ := connSvc.CreateConnection(ctx, SaveConnectionRequest{
		Name: "T", URL: srv.URL, APIKey: "super-secret-key",
	})
	_, _ = connSvc.SelectConnection(ctx, sum.ID)

	qSvc := NewQueryService(connSvc, mgr)
	_, _ = qSvc.ExecuteSearch(ctx, "movies", `{}`)

	b, err := os.ReadFile(filepath.Join(mgr.Dir(), historyFileName))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if containsStr(string(b), "super-secret-key") {
		t.Error("API key must not appear in the history file")
	}
}
