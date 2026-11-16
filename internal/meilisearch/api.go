package meilisearch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// HealthResponse is the response from GET /health.
type HealthResponse struct {
	Status string `json:"status"`
}

// VersionResponse is the response from GET /version.
type VersionResponse struct {
	CommitSHA  string `json:"commitSha"`
	CommitDate string `json:"commitDate"`
	PkgVersion string `json:"pkgVersion"`
}

// Index is an individual Meilisearch index as returned by the API.
type Index struct {
	UID        string  `json:"uid"`
	PrimaryKey *string `json:"primaryKey"`
	CreatedAt  string  `json:"createdAt"`
	UpdatedAt  string  `json:"updatedAt"`
}

// IndexesResult is the paginated response from GET /indexes.
type IndexesResult struct {
	Results []Index `json:"results"`
	Offset  int     `json:"offset"`
	Limit   int     `json:"limit"`
	Total   int     `json:"total"`
}

// IndexStats is the response from GET /indexes/{uid}/stats.
type IndexStats struct {
	NumberOfDocuments int64            `json:"numberOfDocuments"`
	IsIndexing        bool             `json:"isIndexing"`
	FieldDistribution map[string]int64 `json:"fieldDistribution"`
}

// Health calls GET /health and returns the server health status.
func (c *Client) Health(ctx context.Context) (*HealthResponse, error) {
	var out HealthResponse
	if err := c.Get(ctx, "/health", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Version calls GET /version and returns server version information.
func (c *Client) Version(ctx context.Context) (*VersionResponse, error) {
	var out VersionResponse
	if err := c.Get(ctx, "/version", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListIndexes calls GET /indexes with offset and limit pagination.
func (c *Client) ListIndexes(ctx context.Context, offset, limit int) (*IndexesResult, error) {
	path := fmt.Sprintf("/indexes?offset=%d&limit=%d", offset, limit)
	var out IndexesResult
	if err := c.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetIndex calls GET /indexes/{uid} for a single index.
func (c *Client) GetIndex(ctx context.Context, uid string) (*Index, error) {
	path := "/indexes/" + url.PathEscape(uid)
	var out Index
	if err := c.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetIndexStats calls GET /indexes/{uid}/stats.
func (c *Client) GetIndexStats(ctx context.Context, uid string) (*IndexStats, error) {
	path := "/indexes/" + url.PathEscape(uid) + "/stats"
	var out IndexStats
	if err := c.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DocumentsResult is the paginated response from GET /indexes/{uid}/documents.
// Each element of Results is a raw JSON object, preserving all types and numeric precision.
type DocumentsResult struct {
	Results []json.RawMessage `json:"results"`
	Offset  int               `json:"offset"`
	Limit   int               `json:"limit"`
	Total   int64             `json:"total"`
}

// GetDocuments calls GET /indexes/{uid}/documents with offset/limit pagination.
// If fields is non-empty, only the specified field names are returned by the server.
func (c *Client) GetDocuments(ctx context.Context, uid string, offset, limit int, fields []string) (*DocumentsResult, error) {
	params := url.Values{}
	params.Set("offset", strconv.Itoa(offset))
	params.Set("limit", strconv.Itoa(limit))
	if len(fields) > 0 {
		params.Set("fields", strings.Join(fields, ","))
	}
	path := "/indexes/" + url.PathEscape(uid) + "/documents?" + params.Encode()
	var out DocumentsResult
	if err := c.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetDocument calls GET /indexes/{uid}/documents/{id} and returns the raw document JSON.
func (c *Client) GetDocument(ctx context.Context, uid, docID string) (json.RawMessage, error) {
	path := "/indexes/" + url.PathEscape(uid) + "/documents/" + url.PathEscape(docID)
	var out json.RawMessage
	if err := c.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// TaskResponse is the asynchronous task reference returned by document mutation endpoints.
type TaskResponse struct {
	TaskUID    int64  `json:"taskUid"`
	IndexUID   string `json:"indexUid"`
	Status     string `json:"status"`
	Type       string `json:"type"`
	EnqueuedAt string `json:"enqueuedAt"`
}

// AddDocuments calls POST /indexes/{uid}/documents with docs as the request body.
// docs must be a JSON array or object. If the index already has a document with
// the same primary-key value, it is completely replaced. Returns a task reference.
func (c *Client) AddDocuments(ctx context.Context, uid string, docs json.RawMessage) (json.RawMessage, error) {
	path := "/indexes/" + url.PathEscape(uid) + "/documents"
	var raw json.RawMessage
	if err := c.Post(ctx, path, docs, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// UpdateDocuments calls PUT /indexes/{uid}/documents with docs as the request body.
// Only the fields present in docs are updated; all other fields of matching documents
// are preserved. Returns a task reference.
func (c *Client) UpdateDocuments(ctx context.Context, uid string, docs json.RawMessage) (json.RawMessage, error) {
	path := "/indexes/" + url.PathEscape(uid) + "/documents"
	var raw json.RawMessage
	if err := c.Put(ctx, path, docs, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// DeleteDocument calls DELETE /indexes/{uid}/documents/{docID} and returns a task reference.
// docID must be the exact primary-key value; it is URL-encoded before the request is sent.
func (c *Client) DeleteDocument(ctx context.Context, uid string, docID string) (json.RawMessage, error) {
	path := "/indexes/" + url.PathEscape(uid) + "/documents/" + url.PathEscape(docID)
	var raw json.RawMessage
	if err := c.Delete(ctx, path, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// GetTask calls GET /tasks/{taskUID} and returns the raw task JSON response.
func (c *Client) GetTask(ctx context.Context, taskUID int64) (json.RawMessage, error) {
	path := fmt.Sprintf("/tasks/%d", taskUID)
	var raw json.RawMessage
	if err := c.Get(ctx, path, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// Search calls POST /indexes/{uid}/search with requestJSON as the body and returns
// the complete raw response bytes, preserving all fields and numeric precision.
// requestJSON must be a valid JSON object.
func (c *Client) Search(ctx context.Context, uid string, requestJSON json.RawMessage) (json.RawMessage, error) {
	path := "/indexes/" + url.PathEscape(uid) + "/search"
	var raw json.RawMessage
	if err := c.Post(ctx, path, requestJSON, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}
