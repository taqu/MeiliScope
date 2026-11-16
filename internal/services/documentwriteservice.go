package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/taqu/meiliscope/internal/meilisearch"
)

// TaskReference is the synchronous response returned when a document mutation is accepted.
// It references an asynchronous task that must be polled to determine completion.
type TaskReference struct {
	TaskUID    int64  `json:"taskUid"`
	IndexUID   string `json:"indexUid"`
	Status     string `json:"status"`
	Type       string `json:"type"`
	EnqueuedAt string `json:"enqueuedAt"`
}

// DocumentWriteService provides controlled document creation, update, and deletion
// through the active connection. It is safe for concurrent use.
//
// All methods accept only the operations they describe — no arbitrary HTTP
// mutations or administrative endpoints are exposed.
type DocumentWriteService struct {
	connSvc *ConnectionService
}

// NewDocumentWriteService creates a DocumentWriteService backed by connSvc.
func NewDocumentWriteService(connSvc *ConnectionService) *DocumentWriteService {
	return &DocumentWriteService{connSvc: connSvc}
}

// AddDocument sends documentJSON to POST /indexes/{indexUID}/documents.
//
// Meilisearch semantics: if a document with the same primary key already exists, it is
// completely replaced. If no match is found, a new document is created. This is an
// add-or-replace operation, not a create-only operation.
//
// If primaryKeyField is non-empty, the document is validated to contain that field
// before the request is sent.
//
// documentJSON must be a valid JSON object.
func (s *DocumentWriteService) AddDocument(ctx context.Context, indexUID, documentJSON, primaryKeyField string) (TaskReference, error) {
	raw, err := s.validateDocumentJSON(documentJSON, primaryKeyField)
	if err != nil {
		return TaskReference{}, err
	}
	return s.writeDocuments(ctx, indexUID, "add", raw, false)
}

// UpdateDocument sends documentJSON to PUT /indexes/{indexUID}/documents.
//
// Meilisearch semantics: only the fields present in documentJSON are updated.
// All other fields of the matching document are preserved by the server.
// This is a partial-update operation.
//
// If primaryKeyField is non-empty, the document is validated to contain that field.
//
// documentJSON must be a valid JSON object.
func (s *DocumentWriteService) UpdateDocument(ctx context.Context, indexUID, documentJSON, primaryKeyField string) (TaskReference, error) {
	raw, err := s.validateDocumentJSON(documentJSON, primaryKeyField)
	if err != nil {
		return TaskReference{}, err
	}
	return s.writeDocuments(ctx, indexUID, "update", raw, true)
}

// ReplaceDocument sends documentJSON to POST /indexes/{indexUID}/documents.
//
// Meilisearch semantics: identical to AddDocument — the existing document is completely
// replaced. Fields present in the original but absent from documentJSON will be removed.
//
// If primaryKeyField is non-empty, the document is validated to contain that field.
//
// documentJSON must be a valid JSON object.
func (s *DocumentWriteService) ReplaceDocument(ctx context.Context, indexUID, documentJSON, primaryKeyField string) (TaskReference, error) {
	raw, err := s.validateDocumentJSON(documentJSON, primaryKeyField)
	if err != nil {
		return TaskReference{}, err
	}
	return s.writeDocuments(ctx, indexUID, "replace", raw, false)
}

// DeleteDocument sends DELETE /indexes/{indexUID}/documents/{documentID}.
//
// documentID must be the exact primary-key value. It is URL-encoded before dispatch.
// An empty documentID is rejected before any request is made.
func (s *DocumentWriteService) DeleteDocument(ctx context.Context, indexUID, documentID string) (TaskReference, error) {
	if indexUID == "" {
		return TaskReference{}, fmt.Errorf("index UID must not be empty")
	}
	if documentID == "" {
		return TaskReference{}, fmt.Errorf("document ID must not be empty")
	}

	client, gen, err := s.requireClient()
	if err != nil {
		return TaskReference{}, err
	}

	rawResp, apiErr := client.DeleteDocument(ctx, indexUID, documentID)
	if apiErr != nil {
		return TaskReference{}, classifyWriteError("delete document", apiErr)
	}

	if s.connSvc.currentGeneration() != gen {
		return TaskReference{}, fmt.Errorf("connection changed during request")
	}

	return parseTaskReference(rawResp)
}

// --- internal helpers ---

func (s *DocumentWriteService) requireClient() (*meilisearch.Client, int64, error) {
	client, gen := s.connSvc.activeClient()
	if client == nil {
		return nil, 0, fmt.Errorf("no active connection")
	}
	return client, gen, nil
}

// validateDocumentJSON checks that documentJSON is a valid JSON object and, if
// primaryKeyField is non-empty, that the object contains that field.
func (s *DocumentWriteService) validateDocumentJSON(documentJSON, primaryKeyField string) (json.RawMessage, error) {
	raw := json.RawMessage(documentJSON)
	if !json.Valid(raw) {
		return nil, fmt.Errorf("invalid JSON in document")
	}
	var check map[string]json.RawMessage
	if err := json.Unmarshal(raw, &check); err != nil {
		return nil, fmt.Errorf("document must be a JSON object, not an array or scalar")
	}
	if primaryKeyField != "" {
		if _, ok := check[primaryKeyField]; !ok {
			return nil, fmt.Errorf("document is missing required primary-key field %q", primaryKeyField)
		}
	}
	return raw, nil
}

// writeDocuments wraps the validated single document in an array and dispatches the
// appropriate HTTP method (POST for add/replace, PUT for update).
func (s *DocumentWriteService) writeDocuments(ctx context.Context, indexUID, op string, doc json.RawMessage, usePut bool) (TaskReference, error) {
	if indexUID == "" {
		return TaskReference{}, fmt.Errorf("index UID must not be empty")
	}

	client, gen, err := s.requireClient()
	if err != nil {
		return TaskReference{}, err
	}

	// Always send as an array so the API behaviour is unambiguous.
	batch := json.RawMessage("[" + string(doc) + "]")

	var rawResp json.RawMessage
	var apiErr error
	if usePut {
		rawResp, apiErr = client.UpdateDocuments(ctx, indexUID, batch)
	} else {
		rawResp, apiErr = client.AddDocuments(ctx, indexUID, batch)
	}
	if apiErr != nil {
		return TaskReference{}, classifyWriteError(op+" document", apiErr)
	}

	if s.connSvc.currentGeneration() != gen {
		return TaskReference{}, fmt.Errorf("connection changed during request")
	}

	return parseTaskReference(rawResp)
}

// parseTaskReference decodes a Meilisearch task-enqueued response.
func parseTaskReference(raw json.RawMessage) (TaskReference, error) {
	var t struct {
		TaskUID    int64  `json:"taskUid"`
		IndexUID   string `json:"indexUid"`
		Status     string `json:"status"`
		Type       string `json:"type"`
		EnqueuedAt string `json:"enqueuedAt"`
	}
	if err := json.Unmarshal(raw, &t); err != nil {
		return TaskReference{}, fmt.Errorf("decode task reference: %w", err)
	}
	if t.TaskUID == 0 {
		return TaskReference{}, fmt.Errorf("server did not return a task UID")
	}
	return TaskReference{
		TaskUID:    t.TaskUID,
		IndexUID:   t.IndexUID,
		Status:     t.Status,
		Type:       t.Type,
		EnqueuedAt: t.EnqueuedAt,
	}, nil
}

// classifyWriteError maps meilisearch API errors to user-readable messages,
// keeping the structured error accessible for callers.
func classifyWriteError(op string, err error) error {
	switch {
	case meilisearch.IsAuth(err):
		return fmt.Errorf("%s: authentication required: %w", op, err)
	case meilisearch.IsForbidden(err):
		return fmt.Errorf("%s: access denied (write permission required): %w", op, err)
	case meilisearch.IsNotFound(err):
		return fmt.Errorf("%s: index or document not found: %w", op, err)
	case meilisearch.IsBadRequest(err):
		return fmt.Errorf("%s: invalid request: %w", op, err)
	case meilisearch.IsRateLimit(err):
		return fmt.Errorf("%s: rate limit exceeded: %w", op, err)
	case meilisearch.IsCanceled(err):
		return fmt.Errorf("%s: canceled: %w", op, err)
	case meilisearch.IsTimeout(err):
		return fmt.Errorf("%s: request timed out — submission status unknown: %w", op, err)
	default:
		return fmt.Errorf("%s: %w", op, err)
	}
}
