package meilisearch

import (
	"context"
	"fmt"
	"net/url"
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
