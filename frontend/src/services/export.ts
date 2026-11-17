import * as ExSvc from "../../bindings/github.com/taqu/meiliscope/internal/services/exportservice.js";

export interface ExportResult {
  path: string;
  rowCount: number;
}

export const ExportService = {
  exportDocuments: (documentsJSON: string, format: string, destPath: string): Promise<ExportResult> =>
    ExSvc.ExportDocuments(documentsJSON, format, destPath) as Promise<ExportResult>,
  exportSearchResult: (searchResponseJSON: string, format: string, destPath: string): Promise<ExportResult> =>
    ExSvc.ExportSearchResult(searchResponseJSON, format, destPath) as Promise<ExportResult>,
};
