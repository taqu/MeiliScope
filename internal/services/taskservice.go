package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/taqu/meiliscope/internal/meilisearch"
)

// TaskError holds structured error information from a failed Meilisearch task.
// It is safe to expose to the frontend — credentials are never included.
type TaskError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
	Type    string `json:"type"`
	Link    string `json:"link"`
}

// TaskResult is the parsed status of a single Meilisearch task.
// Status is one of: "enqueued", "processing", "succeeded", "failed", "canceled".
// RawJSON holds the complete unmodified task response for detailed inspection.
type TaskResult struct {
	UID        int64      `json:"uid"`
	IndexUID   string     `json:"indexUid"`
	Status     string     `json:"status"`
	Type       string     `json:"type"`
	Error      *TaskError `json:"error,omitempty"`
	Duration   string     `json:"duration,omitempty"`
	EnqueuedAt string     `json:"enqueuedAt"`
	StartedAt  string     `json:"startedAt,omitempty"`
	FinishedAt string     `json:"finishedAt,omitempty"`
	RawJSON    string     `json:"rawJson"`
}

// IsTerminal reports whether the task has reached a terminal state.
func (t TaskResult) IsTerminal() bool {
	switch t.Status {
	case "succeeded", "failed", "canceled":
		return true
	default:
		return false
	}
}

// TaskService retrieves asynchronous Meilisearch task status through the active connection.
// The frontend is responsible for polling — this service provides a single-shot retrieval.
type TaskService struct {
	connSvc *ConnectionService
}

// NewTaskService creates a TaskService backed by connSvc.
func NewTaskService(connSvc *ConnectionService) *TaskService {
	return &TaskService{connSvc: connSvc}
}

// GetTask retrieves the current status of a Meilisearch task by its UID.
//
// Returns a TaskResult with the parsed status and the complete raw JSON response.
// If the task cannot be retrieved (e.g. HTTP 403 when task.get permission is missing),
// the error message explains the limitation without disclosing credentials.
func (s *TaskService) GetTask(ctx context.Context, taskUID int64) (TaskResult, error) {
	if taskUID <= 0 {
		return TaskResult{}, fmt.Errorf("task UID must be a positive integer")
	}

	client, gen, err := s.requireClient()
	if err != nil {
		return TaskResult{}, err
	}

	rawResp, apiErr := client.GetTask(ctx, taskUID)
	if apiErr != nil {
		return TaskResult{}, classifyTaskError(apiErr)
	}

	if s.connSvc.currentGeneration() != gen {
		return TaskResult{}, fmt.Errorf("connection changed during request")
	}

	return parseTaskResult(rawResp)
}

// --- internal helpers ---

func (s *TaskService) requireClient() (*meilisearch.Client, int64, error) {
	client, gen := s.connSvc.activeClient()
	if client == nil {
		return nil, 0, fmt.Errorf("no active connection")
	}
	return client, gen, nil
}

// taskResponseRaw decodes the full task JSON while preserving optional string fields
// that may be absent (null in JSON rather than missing).
type taskResponseRaw struct {
	UID        int64           `json:"uid"`
	IndexUID   string          `json:"indexUid"`
	Status     string          `json:"status"`
	Type       string          `json:"type"`
	Error      *TaskError      `json:"error"`
	Duration   *string         `json:"duration"`
	EnqueuedAt string          `json:"enqueuedAt"`
	StartedAt  *string         `json:"startedAt"`
	FinishedAt *string         `json:"finishedAt"`
	Details    json.RawMessage `json:"details,omitempty"`
}

func parseTaskResult(raw json.RawMessage) (TaskResult, error) {
	var t taskResponseRaw
	if err := json.Unmarshal(raw, &t); err != nil {
		return TaskResult{}, fmt.Errorf("decode task response: %w", err)
	}

	result := TaskResult{
		UID:        t.UID,
		IndexUID:   t.IndexUID,
		Status:     t.Status,
		Type:       t.Type,
		Error:      t.Error,
		EnqueuedAt: t.EnqueuedAt,
		RawJSON:    string(raw),
	}
	if t.Duration != nil {
		result.Duration = *t.Duration
	}
	if t.StartedAt != nil {
		result.StartedAt = *t.StartedAt
	}
	if t.FinishedAt != nil {
		result.FinishedAt = *t.FinishedAt
	}
	return result, nil
}

func classifyTaskError(err error) error {
	switch {
	case meilisearch.IsAuth(err):
		return fmt.Errorf("get task: authentication required: %w", err)
	case meilisearch.IsForbidden(err):
		return fmt.Errorf("get task: permission denied (tasks.get required): %w", err)
	case meilisearch.IsNotFound(err):
		return fmt.Errorf("get task: task not found: %w", err)
	case meilisearch.IsCanceled(err):
		return fmt.Errorf("get task: canceled: %w", err)
	case meilisearch.IsTimeout(err):
		return fmt.Errorf("get task: request timed out: %w", err)
	default:
		return fmt.Errorf("get task: %w", err)
	}
}
