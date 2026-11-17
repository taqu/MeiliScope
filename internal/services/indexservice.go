package services

import (
	"context"
	"fmt"
	"sort"

	"github.com/taqu/meiliscope/internal/meilisearch"
)

// IndexSummary is the frontend-safe view of a Meilisearch index.
type IndexSummary struct {
	UID        string  `json:"uid"`
	PrimaryKey *string `json:"primaryKey"`
	CreatedAt  string  `json:"createdAt"`
	UpdatedAt  string  `json:"updatedAt"`
}

// FieldDistributionEntry is a single field name with its document occurrence count.
type FieldDistributionEntry struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// IndexStatistics holds statistics for an index.
// All fields are optional — the caller should handle zero values gracefully.
type IndexStatistics struct {
	NumberOfDocuments int64                    `json:"numberOfDocuments"`
	IsIndexing        bool                     `json:"isIndexing"`
	FieldDistribution []FieldDistributionEntry `json:"fieldDistribution"`
	// StatsUnavailable is true when statistics could not be fetched (e.g. permission denied).
	StatsUnavailable bool   `json:"statsUnavailable,omitempty"`
	StatsError       string `json:"statsError,omitempty"`
}

// PaginatedIndexResult is the response for a paginated index list request.
type PaginatedIndexResult struct {
	Items  []IndexSummary `json:"items"`
	Offset int            `json:"offset"`
	Limit  int            `json:"limit"`
	Total  int            `json:"total"`
}

// IndexService provides read-only index operations through the active connection.
type IndexService struct {
	connSvc *ConnectionService
}

// NewIndexService creates an IndexService backed by the given connection service.
func NewIndexService(connSvc *ConnectionService) *IndexService {
	return &IndexService{connSvc: connSvc}
}

// ListIndexes returns a paginated list of indexes from the active connection.
func (s *IndexService) ListIndexes(ctx context.Context, offset, limit int) (PaginatedIndexResult, error) {
	client, gen, err := s.requireClient()
	if err != nil {
		return PaginatedIndexResult{}, err
	}

	if limit <= 0 {
		limit = 20
	}

	result, err := client.ListIndexes(ctx, offset, limit)
	if err != nil {
		return PaginatedIndexResult{}, wrapIndexError("list indexes", err)
	}

	// Stale-response guard: reject results if connection was switched mid-flight.
	if s.connSvc.currentGeneration() != gen {
		return PaginatedIndexResult{}, fmt.Errorf("connection changed during request")
	}

	items := make([]IndexSummary, 0, len(result.Results))
	for _, idx := range result.Results {
		items = append(items, toIndexSummary(idx))
	}

	return PaginatedIndexResult{
		Items:  items,
		Offset: result.Offset,
		Limit:  result.Limit,
		Total:  result.Total,
	}, nil
}

// GetIndex returns metadata for a single index by UID.
func (s *IndexService) GetIndex(ctx context.Context, uid string) (IndexSummary, error) {
	client, gen, err := s.requireClient()
	if err != nil {
		return IndexSummary{}, err
	}

	idx, err := client.GetIndex(ctx, uid)
	if err != nil {
		return IndexSummary{}, wrapIndexError("get index", err)
	}

	if s.connSvc.currentGeneration() != gen {
		return IndexSummary{}, fmt.Errorf("connection changed during request")
	}

	return toIndexSummary(*idx), nil
}

// GetIndexStats returns statistics for a single index by UID.
// Permission failures are not fatal: StatsUnavailable is set instead.
func (s *IndexService) GetIndexStats(ctx context.Context, uid string) (IndexStatistics, error) {
	client, gen, err := s.requireClient()
	if err != nil {
		return IndexStatistics{}, err
	}

	stats, err := client.GetIndexStats(ctx, uid)
	if s.connSvc.currentGeneration() != gen {
		return IndexStatistics{}, fmt.Errorf("connection changed during request")
	}

	if err != nil {
		if meilisearch.IsForbidden(err) || meilisearch.IsAuth(err) {
			return IndexStatistics{
				StatsUnavailable: true,
				StatsError:       "insufficient permissions to read index statistics",
			}, nil
		}
		if meilisearch.IsNotFound(err) {
			return IndexStatistics{
				StatsUnavailable: true,
				StatsError:       "index not found",
			}, nil
		}
		return IndexStatistics{}, wrapIndexError("get index stats", err)
	}

	return toIndexStatistics(stats), nil
}

// --- helpers ---

func (s *IndexService) requireClient() (*meilisearch.Client, int64, error) {
	client, gen := s.connSvc.activeClient()
	if client == nil {
		return nil, 0, fmt.Errorf("no active connection")
	}
	return client, gen, nil
}

func toIndexSummary(idx meilisearch.Index) IndexSummary {
	return IndexSummary{
		UID:        idx.UID,
		PrimaryKey: idx.PrimaryKey,
		CreatedAt:  idx.CreatedAt,
		UpdatedAt:  idx.UpdatedAt,
	}
}

func toIndexStatistics(s *meilisearch.IndexStats) IndexStatistics {
	dist := make([]FieldDistributionEntry, 0, len(s.FieldDistribution))
	for name, count := range s.FieldDistribution {
		dist = append(dist, FieldDistributionEntry{Name: name, Count: count})
	}
	// Sort descending by count for a consistent UI presentation.
	sort.Slice(dist, func(i, j int) bool {
		if dist[i].Count != dist[j].Count {
			return dist[i].Count > dist[j].Count
		}
		return dist[i].Name < dist[j].Name
	})

	return IndexStatistics{
		NumberOfDocuments: s.NumberOfDocuments,
		IsIndexing:        s.IsIndexing,
		FieldDistribution: dist,
	}
}

func wrapIndexError(op string, err error) error {
	if meilisearch.IsAuth(err) {
		return fmt.Errorf("%s: authentication required", op)
	}
	if meilisearch.IsForbidden(err) {
		return fmt.Errorf("%s: access denied", op)
	}
	if meilisearch.IsCanceled(err) {
		return fmt.Errorf("%s: canceled", op)
	}
	return fmt.Errorf("%s: %w", op, err)
}
