// Package meilisearch provides a reusable HTTP client for the Meilisearch REST API.
// It is independent of Wails and the frontend.
package meilisearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultTimeout      = 10 * time.Second
	defaultMaxBodyBytes = 10 * 1024 * 1024 // 10 MiB
)

// Config holds the configuration for a Meilisearch HTTP client.
type Config struct {
	BaseURL      string
	APIKey       string
	Timeout      time.Duration
	MaxBodyBytes int64
}

// Client is a Meilisearch HTTP client.
// It is safe to use from multiple goroutines.
type Client struct {
	baseURL      *url.URL
	apiKey       string
	httpClient   *http.Client
	maxBodyBytes int64
}

// New creates a Client from cfg.
// It validates the base URL: HTTP or HTTPS scheme required; no userinfo; no fragment.
// Valid path prefixes (e.g. /meilisearch) are preserved.
func New(cfg Config) (*Client, error) {
	if err := validateBaseURL(cfg.BaseURL); err != nil {
		return nil, fmt.Errorf("meilisearch: invalid base URL: %w", err)
	}

	u, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("meilisearch: parse base URL: %w", err)
	}
	// Normalize: strip trailing slash from path so buildURL can always append /path.
	u.Path = strings.TrimRight(u.Path, "/")
	u.RawQuery = ""
	u.Fragment = ""

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	maxBody := cfg.MaxBodyBytes
	if maxBody <= 0 {
		maxBody = defaultMaxBodyBytes
	}

	// Use a dedicated transport (not shared DefaultTransport) to avoid global
	// state mutation. Clone preserves all TLS and proxy settings.
	transport := http.DefaultTransport.(*http.Transport).Clone()

	return &Client{
		baseURL: u,
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
			// Strip the Authorization header when following cross-origin redirects
			// to prevent credential leakage (Fetch-spec-equivalent policy).
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("meilisearch: stopped after 10 redirects")
				}
				if len(via) > 0 {
					orig := via[0].URL
					if req.URL.Host != orig.Host || req.URL.Scheme != orig.Scheme {
						req.Header.Del("Authorization")
					}
				}
				return nil
			},
		},
		maxBodyBytes: maxBody,
	}, nil
}

// validateBaseURL enforces URL safety rules described in §1.1.3.
func validateBaseURL(raw string) error {
	if raw == "" {
		return fmt.Errorf("URL must not be empty")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("malformed URL: %w", err)
	}

	switch u.Scheme {
	case "http", "https":
		// accepted
	default:
		return fmt.Errorf("scheme %q is not allowed; use http or https", u.Scheme)
	}

	if u.Host == "" {
		return fmt.Errorf("URL must have a host")
	}

	if u.User != nil {
		return fmt.Errorf("URL must not contain userinfo credentials")
	}

	if u.Fragment != "" {
		return fmt.Errorf("URL must not contain a fragment")
	}

	return nil
}

// buildURL appends an API path to the base URL.
// path must start with "/".
func (c *Client) buildURL(path string) string {
	return strings.TrimRight(c.baseURL.String(), "/") + path
}

// Get performs an HTTP GET request to path and JSON-decodes the response into out.
func (c *Client) Get(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.buildURL(path), nil)
	if err != nil {
		return &APIError{Kind: ErrKindBadRequest, Msg: "failed to build request"}
	}
	c.setHeaders(req)
	return c.do(req, out)
}

// Post performs an HTTP POST request to path with body serialized as JSON,
// then JSON-decodes the response into out. out may be nil to discard the body.
func (c *Client) Post(ctx context.Context, path string, body any, out any) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return &APIError{Kind: ErrKindBadRequest, Msg: "failed to serialize request body"}
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.buildURL(path), bodyReader)
	if err != nil {
		return &APIError{Kind: ErrKindBadRequest, Msg: "failed to build request"}
	}
	c.setHeaders(req)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.do(req, out)
}

// Put performs an HTTP PUT request to path with body serialized as JSON,
// then JSON-decodes the response into out. out may be nil to discard the body.
func (c *Client) Put(ctx context.Context, path string, body any, out any) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return &APIError{Kind: ErrKindBadRequest, Msg: "failed to serialize request body"}
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.buildURL(path), bodyReader)
	if err != nil {
		return &APIError{Kind: ErrKindBadRequest, Msg: "failed to build request"}
	}
	c.setHeaders(req)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.do(req, out)
}

// Delete performs an HTTP DELETE request to path and JSON-decodes the response into out.
// out may be nil to discard the body.
func (c *Client) Delete(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.buildURL(path), nil)
	if err != nil {
		return &APIError{Kind: ErrKindBadRequest, Msg: "failed to build request"}
	}
	c.setHeaders(req)
	return c.do(req, out)
}

// setHeaders adds authentication and Accept headers to req.
// The API key is never logged or included in errors.
func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
}

// do executes req and handles the response.
func (c *Client) do(req *http.Request, out any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return classifyTransportError(err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.readErrorBody(resp)
	}

	if out == nil {
		return nil
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		return &APIError{
			Kind:   ErrKindUnexpectedContent,
			Status: resp.StatusCode,
			Msg:    fmt.Sprintf("unexpected Content-Type %q; expected application/json", ct),
		}
	}

	limited := io.LimitReader(resp.Body, c.maxBodyBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return classifyTransportError(err)
	}
	if int64(len(data)) > c.maxBodyBytes {
		return &APIError{Kind: ErrKindResponseTooLarge, Msg: "response body exceeded size limit"}
	}

	if err := json.Unmarshal(data, out); err != nil {
		return &APIError{
			Kind:   ErrKindInvalidJSON,
			Status: resp.StatusCode,
			Msg:    "failed to decode JSON response",
		}
	}

	return nil
}

// readErrorBody reads and parses a non-2xx response into an APIError.
func (c *Client) readErrorBody(resp *http.Response) error {
	limited := io.LimitReader(resp.Body, 64*1024)
	data, _ := io.ReadAll(limited)

	var body apiErrorBody
	if json.Unmarshal(data, &body) == nil && body.Message != "" {
		return httpStatusToAPIError(resp.StatusCode, &body)
	}

	return httpStatusToAPIError(resp.StatusCode, nil)
}
