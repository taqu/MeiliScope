package meilisearch_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/taqu/meiliscope/internal/meilisearch"
)

// newTestClient creates a Client pointing at srv.
func newTestClient(t *testing.T, srv *httptest.Server, apiKey string) *meilisearch.Client {
	t.Helper()
	c, err := meilisearch.New(meilisearch.Config{
		BaseURL: srv.URL,
		APIKey:  apiKey,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

// --- URL validation ---

func TestNew_InvalidURL(t *testing.T) {
	cases := []struct {
		name string
		url  string
	}{
		{"empty", ""},
		{"no scheme", "localhost:7700"},
		{"ftp scheme", "ftp://localhost:7700"},
		{"userinfo", "http://user:pass@localhost:7700"},
		{"fragment", "http://localhost:7700#section"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := meilisearch.New(meilisearch.Config{BaseURL: tc.url})
			if err == nil {
				t.Fatal("expected error; got nil")
			}
		})
	}
}

func TestNew_ValidURLs(t *testing.T) {
	cases := []string{
		"http://127.0.0.1:7700",
		"http://localhost:7700",
		"https://search.example.com",
		"https://search.example.com/meilisearch",
	}
	for _, u := range cases {
		t.Run(u, func(t *testing.T) {
			_, err := meilisearch.New(meilisearch.Config{BaseURL: u})
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", u, err)
			}
		})
	}
}

// --- Bearer authentication ---

func TestGet_BearerAuth(t *testing.T) {
	const testKey = "test-api-key-abc123"
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"available"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, testKey)
	var out meilisearch.HealthResponse
	// Use a custom Get via Health
	_, err := c.Health(context.Background())
	if err != nil {
		t.Fatalf("Health: %v", err)
	}
	_ = out

	want := "Bearer " + testKey
	if gotAuth != want {
		t.Errorf("Authorization header = %q; want %q", gotAuth, want)
	}
}

func TestGet_NoAPIKey(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"available"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "") // no API key
	_, err := c.Health(context.Background())
	if err != nil {
		t.Fatalf("Health: %v", err)
	}
	if gotAuth != "" {
		t.Errorf("Authorization header present without API key: %q", gotAuth)
	}
}

// --- Request paths ---

func TestListIndexes_RequestPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[],"offset":0,"limit":20,"total":0}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "")
	_, err := c.ListIndexes(context.Background(), 0, 20)
	if err != nil {
		t.Fatalf("ListIndexes: %v", err)
	}
	if gotPath != "/indexes?offset=0&limit=20" {
		t.Errorf("path = %q; want /indexes?offset=0&limit=20", gotPath)
	}
}

func TestGetIndex_PathEscaping(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// EscapedPath preserves percent-encoding where present.
		gotPath = r.URL.EscapedPath()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"uid":"my index","primaryKey":null,"createdAt":"2024-01-01T00:00:00Z","updatedAt":"2024-01-01T00:00:00Z"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "")
	_, err := c.GetIndex(context.Background(), "my index")
	if err != nil {
		t.Fatalf("GetIndex: %v", err)
	}
	if gotPath != "/indexes/my%20index" {
		t.Errorf("path = %q; want /indexes/my%%20index", gotPath)
	}
}

// --- Custom base path ---

func TestCustomBasePath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"available"}`))
	}))
	defer srv.Close()

	c, err := meilisearch.New(meilisearch.Config{BaseURL: srv.URL + "/meilisearch"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.Health(context.Background())
	if err != nil {
		t.Fatalf("Health: %v", err)
	}
	if gotPath != "/meilisearch/health" {
		t.Errorf("path = %q; want /meilisearch/health", gotPath)
	}
}

// --- HTTP status codes ---

func TestHTTP200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"available"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "")
	h, err := c.Health(context.Background())
	if err != nil {
		t.Fatalf("Health: %v", err)
	}
	if h.Status != "available" {
		t.Errorf("status = %q; want available", h.Status)
	}
}

func TestHTTP400(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		_, _ = w.Write([]byte(`{"message":"bad request","code":"bad_request","type":"invalid_request","link":""}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "")
	_, err := c.Health(context.Background())
	if err == nil {
		t.Fatal("expected error; got nil")
	}
	if !meilisearch.IsBadRequest(err) {
		t.Errorf("expected ErrKindBadRequest; got %v", err)
	}
}

func TestHTTP401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(401)
		_, _ = w.Write([]byte(`{"message":"The Authorization header is missing.","code":"missing_authorization_header","type":"auth","link":""}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "")
	_, err := c.Health(context.Background())
	if !meilisearch.IsAuth(err) {
		t.Errorf("expected ErrKindAuth; got %v", err)
	}
}

func TestHTTP403(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(403)
		_, _ = w.Write([]byte(`{"message":"Access denied.","code":"invalid_api_key","type":"auth","link":""}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "wrong-key")
	_, err := c.ListIndexes(context.Background(), 0, 20)
	if !meilisearch.IsForbidden(err) {
		t.Errorf("expected ErrKindForbidden; got %v", err)
	}
}

func TestHTTP404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"message":"Index not found.","code":"index_not_found","type":"invalid_request","link":""}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "")
	_, err := c.GetIndex(context.Background(), "missing")
	if !meilisearch.IsNotFound(err) {
		t.Errorf("expected ErrKindNotFound; got %v", err)
	}
}

func TestHTTP429(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "")
	_, err := c.Health(context.Background())
	if !meilisearch.IsRateLimit(err) {
		t.Errorf("expected ErrKindRateLimit; got %v", err)
	}
}

func TestHTTP500(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "")
	_, err := c.Health(context.Background())
	if !meilisearch.IsServerError(err) {
		t.Errorf("expected ErrKindServer; got %v", err)
	}
}

// --- Malformed JSON ---

func TestMalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{not valid json`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "")
	_, err := c.Health(context.Background())
	if !meilisearch.IsInvalidJSON(err) {
		t.Errorf("expected ErrKindInvalidJSON; got %v", err)
	}
}

// --- Unexpected Content-Type ---

func TestUnexpectedContentType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html>Not JSON</html>`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "")
	_, err := c.Health(context.Background())
	if !meilisearch.IsUnexpectedContent(err) {
		t.Errorf("expected ErrKindUnexpectedContent; got %v", err)
	}
}

// --- Oversized response ---

func TestOversizedResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Write more than 1 MiB
		large := `{"status":"` + strings.Repeat("x", 2*1024*1024) + `"}`
		_, _ = w.Write([]byte(large))
	}))
	defer srv.Close()

	c, err := meilisearch.New(meilisearch.Config{
		BaseURL:      srv.URL,
		MaxBodyBytes: 1024 * 1024, // 1 MiB
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.Health(context.Background())
	if !meilisearch.IsResponseTooLarge(err) {
		t.Errorf("expected ErrKindResponseTooLarge; got %v", err)
	}
}

// --- Timeout ---

func TestTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Block longer than the client timeout
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c, err := meilisearch.New(meilisearch.Config{
		BaseURL: srv.URL,
		Timeout: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.Health(context.Background())
	if !meilisearch.IsTimeout(err) {
		t.Errorf("expected ErrKindTimeout; got %v", err)
	}
}

// --- Context cancellation ---

func TestContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Block until the client gives up
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "")

	// Cancel the context before making the request.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.Health(ctx)
	if !meilisearch.IsCanceled(err) {
		t.Errorf("expected ErrKindCanceled; got %v", err)
	}
}

// --- Error message sanitization ---

func TestErrorMessageDoesNotContainAPIKey(t *testing.T) {
	const sensitiveKey = "super-secret-master-key-12345"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		body, _ := json.Marshal(map[string]string{
			"message": "Invalid API key.",
			"code":    "invalid_api_key",
			"type":    "auth",
			"link":    "",
		})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	c := newTestClient(t, srv, sensitiveKey)
	_, err := c.Health(context.Background())
	if err == nil {
		t.Fatal("expected error; got nil")
	}
	if strings.Contains(err.Error(), sensitiveKey) {
		t.Errorf("error message contains API key: %v", err)
	}
}

// --- Index listing ---

func TestListIndexes_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"results": [
				{"uid":"docs","primaryKey":"id","createdAt":"2024-01-01T00:00:00Z","updatedAt":"2024-01-02T00:00:00Z"},
				{"uid":"users","primaryKey":null,"createdAt":"2024-01-01T00:00:00Z","updatedAt":"2024-01-01T00:00:00Z"}
			],
			"offset": 0,
			"limit": 20,
			"total": 2
		}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "")
	res, err := c.ListIndexes(context.Background(), 0, 20)
	if err != nil {
		t.Fatalf("ListIndexes: %v", err)
	}
	if res.Total != 2 {
		t.Errorf("Total = %d; want 2", res.Total)
	}
	if len(res.Results) != 2 {
		t.Fatalf("len(Results) = %d; want 2", len(res.Results))
	}
	if res.Results[0].UID != "docs" {
		t.Errorf("Results[0].UID = %q; want docs", res.Results[0].UID)
	}
	if res.Results[1].PrimaryKey != nil {
		t.Errorf("Results[1].PrimaryKey = %v; want nil", res.Results[1].PrimaryKey)
	}
}

func TestListIndexes_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[],"offset":0,"limit":20,"total":0}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "")
	res, err := c.ListIndexes(context.Background(), 0, 20)
	if err != nil {
		t.Fatalf("ListIndexes: %v", err)
	}
	if res.Total != 0 || len(res.Results) != 0 {
		t.Errorf("expected empty result; got %+v", res)
	}
}

// --- Index stats ---

func TestGetIndexStats_LargeDocumentCount(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"numberOfDocuments": 9999999999,
			"isIndexing": false,
			"fieldDistribution": {"id": 9999999999, "name": 5000000}
		}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "")
	stats, err := c.GetIndexStats(context.Background(), "big")
	if err != nil {
		t.Fatalf("GetIndexStats: %v", err)
	}
	if stats.NumberOfDocuments != 9999999999 {
		t.Errorf("NumberOfDocuments = %d; want 9999999999", stats.NumberOfDocuments)
	}
}

// --- Redirect security (CVE-class: cross-host credential forwarding) ---

// TestRedirect_SameHost verifies that Authorization is preserved on same-host redirects.
func TestRedirect_SameHost(t *testing.T) {
	const testKey = "secret-api-key"
	var finalAuth string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/health", http.StatusMovedPermanently)
			return
		}
		finalAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"available"}`))
	}))
	defer srv.Close()

	c, err := meilisearch.New(meilisearch.Config{BaseURL: srv.URL, APIKey: testKey})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// Issue a request that hits /redirect first; the client must follow to /health.
	if err := c.Get(context.Background(), "/redirect", new(meilisearch.HealthResponse)); err != nil {
		t.Fatalf("Get: %v", err)
	}
	want := "Bearer " + testKey
	if finalAuth != want {
		t.Errorf("Authorization after same-host redirect = %q; want %q", finalAuth, want)
	}
}

// TestRedirect_CrossHost verifies that Authorization is stripped on cross-host redirects
// to prevent credential leakage.
func TestRedirect_CrossHost(t *testing.T) {
	const testKey = "super-secret-key"
	var attackerAuth string

	// Attacker server: records whatever Authorization header it receives.
	attacker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attackerAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"available"}`))
	}))
	defer attacker.Close()

	// Origin server: immediately redirects to the attacker server.
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, attacker.URL+"/health", http.StatusMovedPermanently)
	}))
	defer origin.Close()

	c, err := meilisearch.New(meilisearch.Config{BaseURL: origin.URL, APIKey: testKey})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// The request will follow the redirect to the attacker.
	_ = c.Get(context.Background(), "/health", new(meilisearch.HealthResponse))

	// The attacker must NOT have received the Authorization header.
	if attackerAuth != "" {
		t.Errorf("credential leaked to cross-host redirect target: Authorization = %q", attackerAuth)
	}
}

// TestRedirect_TooMany verifies that the client stops after 10 redirects.
func TestRedirect_TooMany(t *testing.T) {
	count := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		http.Redirect(w, r, "/health", http.StatusMovedPermanently)
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "")
	_, err := c.Health(context.Background())
	if err == nil {
		t.Fatal("expected error after too many redirects; got nil")
	}
	// Error should indicate redirect loop, not a nil or unrelated error.
	if meilisearch.IsCanceled(err) {
		t.Errorf("unexpected canceled error: %v", err)
	}
}
