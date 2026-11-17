package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/taqu/meiliscope/internal/config"
	"github.com/taqu/meiliscope/internal/meilisearch"
)

// SearchError holds structured diagnostic information about a failed search.
// It is safe to expose to the frontend — credentials are never included.
type SearchError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
	Type    string `json:"type"`
	Link    string `json:"link"`
	Status  int    `json:"status"`
	Kind    string `json:"kind"` // auth | forbidden | not_found | bad_request | timeout | canceled | rate_limit | server | network | parse | response_too_large
}

// FacetStat holds min/max statistics for a numeric facet attribute.
type FacetStat struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// SearchMeta contains the parsed metadata from a Meilisearch search response.
// All pointer fields are optional because different pagination modes and server
// versions return different subsets.
type SearchMeta struct {
	Query              string                      `json:"query,omitempty"`
	ProcessingTimeMs   int                         `json:"processingTimeMs"`
	Limit              *int                        `json:"limit,omitempty"`
	Offset             *int                        `json:"offset,omitempty"`
	EstimatedTotalHits *int64                      `json:"estimatedTotalHits,omitempty"`
	Page               *int64                      `json:"page,omitempty"`
	HitsPerPage        *int                        `json:"hitsPerPage,omitempty"`
	TotalPages         *int64                      `json:"totalPages,omitempty"`
	TotalHits          *int64                      `json:"totalHits,omitempty"`
	FacetDistribution  map[string]map[string]int64 `json:"facetDistribution,omitempty"`
	FacetStats         map[string]FacetStat        `json:"facetStats,omitempty"`
	SemanticHitCount   *int64                      `json:"semanticHitCount,omitempty"`
}

// SearchResult is the complete result of a search execution.
// When Error is non-nil the search failed; Hits and Meta are empty.
// When Error is nil the search succeeded.
type SearchResult struct {
	Hits             []DocumentRecord `json:"hits"`
	Meta             SearchMeta       `json:"meta"`
	RawJSON          string           `json:"rawJson"` // complete raw response body
	ConnectionID     string           `json:"connectionId"`
	IndexUID         string           `json:"indexUid"`
	SubmittedQuery   string           `json:"submittedQuery"`
	ExecutedAt       string           `json:"executedAt"`       // RFC3339 UTC
	ClientDurationMs int64            `json:"clientDurationMs"` // wall-clock ms
	Error            *SearchError     `json:"error,omitempty"`
}

// HistoryEntry is a persisted record of a past query execution.
// It never includes API keys or authentication headers.
type HistoryEntry struct {
	ID               string `json:"id"`
	ConnectionID     string `json:"connectionId"`
	ConnectionName   string `json:"connectionName"`
	IndexUID         string `json:"indexUid"`
	QueryJSON        string `json:"queryJson"`
	ExecutedAt       string `json:"executedAt"`
	Status           string `json:"status"` // "success" | "error"
	ClientDurationMs int64  `json:"clientDurationMs,omitempty"`
	ErrorSummary     string `json:"errorSummary,omitempty"`
	ProcessingTimeMs int    `json:"processingTimeMs,omitempty"`
	HitsCount        int    `json:"hitsCount,omitempty"`
	TotalHits        *int64 `json:"totalHits,omitempty"`
}

// QueryService executes Meilisearch search requests and manages query history.
// It is safe for concurrent use.
type QueryService struct {
	connSvc *ConnectionService
	history *historyStore

	mu         sync.Mutex
	cancelFunc context.CancelFunc
	searchGen  int64 // incremented each time a search starts
}

// NewQueryService creates a QueryService.
// History is loaded from cfgManager.Dir() on creation.
func NewQueryService(connSvc *ConnectionService, cfgManager *config.Manager) *QueryService {
	qs := &QueryService{
		connSvc: connSvc,
		history: newHistoryStore(cfgManager.Dir()),
	}
	qs.history.load()
	return qs
}

// ExecuteSearch sends queryJSON to POST /indexes/{indexUID}/search on the active connection.
//
// Returns (SearchResult{}, error) for pre-condition failures:
//   - invalid or non-object queryJSON
//   - no active connection
//   - connection switched mid-flight
//
// Returns (SearchResult{Error: ...}, nil) for search API failures (HTTP 4xx/5xx,
// network errors, timeouts, cancellation).
//
// Returns (SearchResult{Hits: ..., Meta: ...}, nil) on success.
//
// Starting a new search cancels any in-flight search from the same service instance.
func (s *QueryService) ExecuteSearch(ctx context.Context, indexUID string, queryJSON string) (SearchResult, error) {
	if indexUID == "" {
		return SearchResult{}, fmt.Errorf("index UID must not be empty")
	}
	rawBody := json.RawMessage(queryJSON)
	if !json.Valid(rawBody) {
		return SearchResult{}, fmt.Errorf("invalid JSON in query")
	}
	var check map[string]json.RawMessage
	if err := json.Unmarshal(rawBody, &check); err != nil {
		return SearchResult{}, fmt.Errorf("query must be a JSON object, not an array or scalar")
	}

	client, gen, err := s.requireClient()
	if err != nil {
		return SearchResult{}, err
	}

	// Cancel any previous in-flight search and register this one.
	searchCtx, cancel := context.WithCancel(ctx)
	s.mu.Lock()
	if s.cancelFunc != nil {
		s.cancelFunc()
	}
	s.cancelFunc = cancel
	s.searchGen++
	myGen := s.searchGen
	s.mu.Unlock()
	defer func() {
		cancel()
		s.mu.Lock()
		if s.searchGen == myGen {
			s.cancelFunc = nil
		}
		s.mu.Unlock()
	}()

	connInfo := s.connSvc.GetActiveConnection()
	connID, connName := "", ""
	if connInfo != nil {
		connID = connInfo.ID
		connName = connInfo.Name
	}

	executedAt := time.Now()
	executedAtStr := executedAt.UTC().Format(time.RFC3339)

	start := time.Now()
	rawResp, searchErr := client.Search(searchCtx, indexUID, rawBody)
	clientDur := time.Since(start).Milliseconds()

	// Guard: reject results if the connection was switched while the request ran.
	if s.connSvc.currentGeneration() != gen {
		return SearchResult{}, fmt.Errorf("connection changed during request")
	}

	base := SearchResult{
		ConnectionID:     connID,
		IndexUID:         indexUID,
		SubmittedQuery:   queryJSON,
		ExecutedAt:       executedAtStr,
		ClientDurationMs: clientDur,
	}

	if searchErr != nil {
		se := classifySearchError(searchErr)
		base.Error = se
		s.history.add(HistoryEntry{
			ID:               mustGenerateID(),
			ConnectionID:     connID,
			ConnectionName:   connName,
			IndexUID:         indexUID,
			QueryJSON:        truncateQuery(queryJSON),
			ExecutedAt:       executedAtStr,
			Status:           "error",
			ClientDurationMs: clientDur,
			ErrorSummary:     se.Message,
		})
		return base, nil
	}

	parsed, parseErr := parseSearchResponse(rawResp)
	if parseErr != nil {
		base.Error = &SearchError{Message: parseErr.Error(), Kind: "parse"}
		s.history.add(HistoryEntry{
			ID:               mustGenerateID(),
			ConnectionID:     connID,
			ConnectionName:   connName,
			IndexUID:         indexUID,
			QueryJSON:        truncateQuery(queryJSON),
			ExecutedAt:       executedAtStr,
			Status:           "error",
			ClientDurationMs: clientDur,
			ErrorSummary:     parseErr.Error(),
		})
		return base, nil
	}

	base.Hits = parsed.Hits
	base.Meta = parsed.Meta
	base.RawJSON = parsed.RawJSON

	var totalHits *int64
	if parsed.Meta.TotalHits != nil {
		totalHits = parsed.Meta.TotalHits
	} else if parsed.Meta.EstimatedTotalHits != nil {
		totalHits = parsed.Meta.EstimatedTotalHits
	}

	s.history.add(HistoryEntry{
		ID:               mustGenerateID(),
		ConnectionID:     connID,
		ConnectionName:   connName,
		IndexUID:         indexUID,
		QueryJSON:        truncateQuery(queryJSON),
		ExecutedAt:       executedAtStr,
		Status:           "success",
		ClientDurationMs: clientDur,
		ProcessingTimeMs: parsed.Meta.ProcessingTimeMs,
		HitsCount:        len(parsed.Hits),
		TotalHits:        totalHits,
	})

	return base, nil
}

// CancelSearch cancels any in-flight search request.
func (s *QueryService) CancelSearch() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancelFunc != nil {
		s.cancelFunc()
		s.cancelFunc = nil
	}
}

// GetHistory returns saved query history entries, newest first.
func (s *QueryService) GetHistory() []HistoryEntry {
	return s.history.list()
}

// DeleteHistoryEntry removes a single history entry by ID.
func (s *QueryService) DeleteHistoryEntry(id string) error {
	return s.history.delete(id)
}

// ClearHistory removes all history entries.
func (s *QueryService) ClearHistory() error {
	return s.history.clear()
}

// --- internal helpers ---

func (s *QueryService) requireClient() (*meilisearch.Client, int64, error) {
	client, gen := s.connSvc.activeClient()
	if client == nil {
		return nil, 0, fmt.Errorf("no active connection")
	}
	return client, gen, nil
}

// searchResponseRaw decodes the raw search response while preserving each hit as
// json.RawMessage so that rawToDocumentRecord can handle precision-sensitive values.
type searchResponseRaw struct {
	Hits               []json.RawMessage           `json:"hits"`
	Query              string                      `json:"query,omitempty"`
	ProcessingTimeMs   int                         `json:"processingTimeMs"`
	Limit              *int                        `json:"limit,omitempty"`
	Offset             *int                        `json:"offset,omitempty"`
	EstimatedTotalHits *int64                      `json:"estimatedTotalHits,omitempty"`
	Page               *int64                      `json:"page,omitempty"`
	HitsPerPage        *int                        `json:"hitsPerPage,omitempty"`
	TotalPages         *int64                      `json:"totalPages,omitempty"`
	TotalHits          *int64                      `json:"totalHits,omitempty"`
	FacetDistribution  map[string]map[string]int64 `json:"facetDistribution,omitempty"`
	FacetStats         map[string]FacetStat        `json:"facetStats,omitempty"`
	SemanticHitCount   *int64                      `json:"semanticHitCount,omitempty"`
}

type parsedSearchResult struct {
	Hits    []DocumentRecord
	Meta    SearchMeta
	RawJSON string
}

func parseSearchResponse(rawResp json.RawMessage) (parsedSearchResult, error) {
	var parsed searchResponseRaw
	if err := json.Unmarshal(rawResp, &parsed); err != nil {
		return parsedSearchResult{}, fmt.Errorf("decode search response: %w", err)
	}

	hits := make([]DocumentRecord, 0, len(parsed.Hits))
	for _, rawHit := range parsed.Hits {
		rec, err := rawToDocumentRecord(rawHit)
		if err != nil {
			return parsedSearchResult{}, fmt.Errorf("decode search hit: %w", err)
		}
		hits = append(hits, rec)
	}

	return parsedSearchResult{
		Hits: hits,
		Meta: SearchMeta{
			Query:              parsed.Query,
			ProcessingTimeMs:   parsed.ProcessingTimeMs,
			Limit:              parsed.Limit,
			Offset:             parsed.Offset,
			EstimatedTotalHits: parsed.EstimatedTotalHits,
			Page:               parsed.Page,
			HitsPerPage:        parsed.HitsPerPage,
			TotalPages:         parsed.TotalPages,
			TotalHits:          parsed.TotalHits,
			FacetDistribution:  parsed.FacetDistribution,
			FacetStats:         parsed.FacetStats,
			SemanticHitCount:   parsed.SemanticHitCount,
		},
		RawJSON: string(rawResp),
	}, nil
}

func classifySearchError(err error) *SearchError {
	var apiErr *meilisearch.APIError
	se := &SearchError{}
	if errors.As(err, &apiErr) {
		se.Message = apiErr.Msg
		se.Code = apiErr.Code
		se.Type = apiErr.Type
		se.Link = apiErr.Link
		se.Status = apiErr.Status
	} else {
		se.Message = err.Error()
	}

	switch {
	case meilisearch.IsAuth(err):
		se.Kind = "auth"
	case meilisearch.IsForbidden(err):
		se.Kind = "forbidden"
	case meilisearch.IsNotFound(err):
		se.Kind = "not_found"
	case meilisearch.IsBadRequest(err):
		se.Kind = "bad_request"
	case meilisearch.IsCanceled(err):
		se.Kind = "canceled"
		if se.Message == "" {
			se.Message = "request canceled"
		}
	case meilisearch.IsTimeout(err):
		se.Kind = "timeout"
		if se.Message == "" {
			se.Message = "request timed out"
		}
	case meilisearch.IsRateLimit(err):
		se.Kind = "rate_limit"
	case meilisearch.IsServerError(err):
		se.Kind = "server"
	case meilisearch.IsResponseTooLarge(err):
		se.Kind = "response_too_large"
	default:
		se.Kind = "network"
		if se.Message == "" {
			se.Message = "network error"
		}
	}

	return se
}

func truncateQuery(q string) string {
	if len(q) > maxQueryBytes {
		return q[:maxQueryBytes] + "…"
	}
	return q
}

func mustGenerateID() string {
	id, err := generateID()
	if err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return id
}
