import * as IndexSvc from "../../bindings/github.com/taqu/meiliscope/internal/services/indexservice.js";

export type {
  FieldDistributionEntry,
  IndexStatistics,
  IndexSummary,
  PaginatedIndexResult,
} from "../../bindings/github.com/taqu/meiliscope/internal/services/models.js";

export const IndexService = {
  listIndexes: (offset: number, limit: number) =>
    IndexSvc.ListIndexes(offset, limit),
  getIndex: (uid: string) => IndexSvc.GetIndex(uid),
  getIndexStats: (uid: string) => IndexSvc.GetIndexStats(uid),
};
