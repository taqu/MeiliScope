import * as QuerySvc from "../../bindings/github.com/taqu/meiliscope/internal/services/queryservice.js";
import type { DocumentRecord } from "./document";

export type { DocumentRecord };

export interface FacetStat {
  min: number;
  max: number;
}

export interface SearchError {
  message: string;
  code: string;
  type: string;
  link: string;
  status: number;
  kind: string;
}

export interface SearchMeta {
  query?: string;
  processingTimeMs: number;
  limit?: number;
  offset?: number;
  estimatedTotalHits?: number;
  page?: number;
  hitsPerPage?: number;
  totalPages?: number;
  totalHits?: number;
  facetDistribution?: Record<string, Record<string, number>>;
  facetStats?: Record<string, FacetStat>;
  semanticHitCount?: number;
}

export interface SearchResult {
  hits: DocumentRecord[];
  meta: SearchMeta;
  rawJson: string;
  connectionId: string;
  indexUid: string;
  submittedQuery: string;
  executedAt: string;
  clientDurationMs: number;
  error?: SearchError | null;
}

export interface HistoryEntry {
  id: string;
  connectionId: string;
  connectionName: string;
  indexUid: string;
  queryJson: string;
  executedAt: string;
  status: string;
  clientDurationMs?: number;
  errorSummary?: string;
  processingTimeMs?: number;
  hitsCount?: number;
  totalHits?: number;
}

export const QueryService = {
  executeSearch: (indexUID: string, queryJSON: string): Promise<SearchResult> =>
    QuerySvc.ExecuteSearch(indexUID, queryJSON) as Promise<SearchResult>,
  cancelSearch: (): Promise<void> =>
    QuerySvc.CancelSearch() as Promise<void>,
  getHistory: (): Promise<HistoryEntry[]> =>
    QuerySvc.GetHistory() as Promise<HistoryEntry[]>,
  deleteHistoryEntry: (id: string): Promise<void> =>
    QuerySvc.DeleteHistoryEntry(id) as Promise<void>,
  clearHistory: (): Promise<void> =>
    QuerySvc.ClearHistory() as Promise<void>,
};
