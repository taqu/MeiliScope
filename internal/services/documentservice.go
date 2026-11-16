package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/taqu/meiliscope/internal/meilisearch"
)

// DocumentListOptions controls document listing pagination and optional field selection.
type DocumentListOptions struct {
	Offset int      `json:"offset"`
	Limit  int      `json:"limit"`
	Fields []string `json:"fields,omitempty"`
}

// DocumentRecord is a single document from Meilisearch.
// RawJSON holds the complete document as an exact JSON string — no re-encoding.
// Fields extracts each top-level key to its raw JSON value string, enabling
// table rendering without full re-parsing and preserving numeric precision.
type DocumentRecord struct {
	RawJSON string            `json:"rawJson"`
	Fields  map[string]string `json:"fields"`
}

// DocumentPage is a paginated document list response.
type DocumentPage struct {
	Items  []DocumentRecord `json:"items"`
	Offset int              `json:"offset"`
	Limit  int              `json:"limit"`
	Total  int64            `json:"total"`
}

// DocumentService provides read-only document operations through the active connection.
type DocumentService struct {
	connSvc *ConnectionService
}

// NewDocumentService creates a DocumentService backed by the given connection service.
func NewDocumentService(connSvc *ConnectionService) *DocumentService {
	return &DocumentService{connSvc: connSvc}
}

const (
	defaultDocLimit = 50
	maxDocLimit     = 200
)

// ListDocuments returns a paginated list of documents from the given index.
func (s *DocumentService) ListDocuments(ctx context.Context, indexUID string, opts DocumentListOptions) (DocumentPage, error) {
	client, gen, err := s.requireClient()
	if err != nil {
		return DocumentPage{}, err
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = defaultDocLimit
	}
	if limit > maxDocLimit {
		limit = maxDocLimit
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	result, err := client.GetDocuments(ctx, indexUID, offset, limit, opts.Fields)
	if err != nil {
		return DocumentPage{}, wrapDocumentError("list documents", err)
	}

	// Stale-response guard: reject results if connection was switched mid-flight.
	if s.connSvc.currentGeneration() != gen {
		return DocumentPage{}, fmt.Errorf("connection changed during request")
	}

	items := make([]DocumentRecord, 0, len(result.Results))
	for _, raw := range result.Results {
		rec, parseErr := rawToDocumentRecord(raw)
		if parseErr != nil {
			return DocumentPage{}, fmt.Errorf("list documents: decode document: %w", parseErr)
		}
		items = append(items, rec)
	}

	return DocumentPage{
		Items:  items,
		Offset: result.Offset,
		Limit:  result.Limit,
		Total:  result.Total,
	}, nil
}

// GetDocument retrieves a single document by its primary-key value.
// docID must be the exact primary-key value; numeric IDs are passed as their
// string representation and are URL-encoded before the request is sent.
func (s *DocumentService) GetDocument(ctx context.Context, indexUID, docID string) (DocumentRecord, error) {
	client, gen, err := s.requireClient()
	if err != nil {
		return DocumentRecord{}, err
	}

	raw, err := client.GetDocument(ctx, indexUID, docID)
	if err != nil {
		return DocumentRecord{}, wrapDocumentError("get document", err)
	}

	if s.connSvc.currentGeneration() != gen {
		return DocumentRecord{}, fmt.Errorf("connection changed during request")
	}

	rec, parseErr := rawToDocumentRecord(raw)
	if parseErr != nil {
		return DocumentRecord{}, fmt.Errorf("get document: decode document: %w", parseErr)
	}

	return rec, nil
}

// --- helpers ---

func (s *DocumentService) requireClient() (*meilisearch.Client, int64, error) {
	client, gen := s.connSvc.activeClient()
	if client == nil {
		return nil, 0, fmt.Errorf("no active connection")
	}
	return client, gen, nil
}

// rawToDocumentRecord converts raw JSON document bytes into a DocumentRecord.
// Fields are stored as raw JSON value strings, preserving type information and
// numeric precision without converting numbers to floating-point representations.
func rawToDocumentRecord(raw json.RawMessage) (DocumentRecord, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return DocumentRecord{}, fmt.Errorf("document is not a JSON object: %w", err)
	}

	fields := make(map[string]string, len(m))
	for k, v := range m {
		fields[k] = string(v)
	}

	return DocumentRecord{
		RawJSON: string(raw),
		Fields:  fields,
	}, nil
}

func wrapDocumentError(op string, err error) error {
	if meilisearch.IsAuth(err) {
		return fmt.Errorf("%s: authentication required", op)
	}
	if meilisearch.IsForbidden(err) {
		return fmt.Errorf("%s: access denied", op)
	}
	if meilisearch.IsNotFound(err) {
		return fmt.Errorf("%s: not found", op)
	}
	if meilisearch.IsCanceled(err) {
		return fmt.Errorf("%s: canceled", op)
	}
	return fmt.Errorf("%s: %w", op, err)
}
