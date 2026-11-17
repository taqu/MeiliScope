import { useState } from "react";
import { Dialogs } from "@wailsio/runtime";
import { ExportService, type ExportResult } from "../../services/export";

// ─── Types ─────────────────────────────────────────────────────────────────────

export type ExportSource =
  | { kind: "documents"; indexUID: string; documentsJSON: string; rowCount: number }
  | { kind: "search"; indexUID: string; searchResponseJSON: string; hitCount: number };

export interface ExportDialogProps {
  source: ExportSource;
  connectionName: string;
  onClose: () => void;
}

type DocumentFormat = "json" | "csv";
type SearchFormat = "json_hits" | "json_full";
type ExportFormat = DocumentFormat | SearchFormat;

// ─── Helpers ───────────────────────────────────────────────────────────────────

function sanitizeUID(uid: string): string {
  return uid.replace(/[^a-zA-Z0-9]/g, "_");
}

function formatDateStamp(): string {
  const now = new Date();
  const yyyy = now.getFullYear();
  const MM = String(now.getMonth() + 1).padStart(2, "0");
  const dd = String(now.getDate()).padStart(2, "0");
  const HH = String(now.getHours()).padStart(2, "0");
  const mm = String(now.getMinutes()).padStart(2, "0");
  const ss = String(now.getSeconds()).padStart(2, "0");
  return `${yyyy}${MM}${dd}_${HH}${mm}${ss}`;
}

function suggestFilename(source: ExportSource, format: ExportFormat): string {
  const safe = sanitizeUID(source.indexUID);
  const date = formatDateStamp();

  if (source.kind === "documents") {
    const ext = format === "csv" ? "csv" : "json";
    return `documents_${safe}_${date}.${ext}`;
  }
  // search
  return `search_results_${safe}_${date}.json`;
}

function getSaveFilters(format: ExportFormat): Array<{ Pattern: string; DisplayName: string }> {
  if (format === "csv") {
    return [{ Pattern: "*.csv", DisplayName: "CSV Files" }];
  }
  return [{ Pattern: "*.json", DisplayName: "JSON Files" }];
}

// ─── Component ─────────────────────────────────────────────────────────────────

export function ExportDialog(props: ExportDialogProps) {
  const { source, connectionName, onClose } = props;

  const defaultFormat: ExportFormat =
    source.kind === "documents" ? "json" : "json_hits";

  const [format, setFormat] = useState<ExportFormat>(defaultFormat);
  const [busy, setBusy] = useState(false);
  const [result, setResult] = useState<ExportResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  const rowCount = source.kind === "documents" ? source.rowCount : source.hitCount;
  const sourceLabel = source.kind === "documents" ? "Documents" : "Search Results";

  async function handleExport() {
    setError(null);
    setResult(null);

    const filename = suggestFilename(source, format);
    const filters = getSaveFilters(format);

    let destPath: string;
    try {
      destPath = await Dialogs.SaveFile({ Filename: filename, Filters: filters });
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : String(e));
      return;
    }

    if (!destPath) {
      // User cancelled the dialog
      return;
    }

    setBusy(true);
    try {
      let res: ExportResult;
      if (source.kind === "documents") {
        res = await ExportService.exportDocuments(source.documentsJSON, format, destPath);
      } else {
        res = await ExportService.exportSearchResult(source.searchResponseJSON, format, destPath);
      }
      setResult(res);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="export-modal-overlay" role="dialog" aria-modal="true" aria-labelledby="export-dialog-title">
      <div className="export-modal">

        {/* Header */}
        <div className="export-modal-header">
          <h2 id="export-dialog-title" className="export-modal-title">
            Export — {source.indexUID}
          </h2>
          <button
            className="btn btn--ghost btn--sm export-modal-close"
            aria-label="Close"
            onClick={onClose}
            disabled={busy}
          >
            ✕
          </button>
        </div>

        {/* Body */}
        <div className="export-modal-body">

          {/* Source summary */}
          <dl className="export-source-dl">
            <dt>Connection</dt>
            <dd>{connectionName}</dd>

            <dt>Index</dt>
            <dd>{source.indexUID}</dd>

            <dt>Source</dt>
            <dd>{sourceLabel}</dd>

            <dt>Rows</dt>
            <dd>~{rowCount.toLocaleString()}</dd>
          </dl>

          {/* Format selection */}
          <fieldset className="export-format-group" disabled={busy || result !== null}>
            <legend>Format</legend>

            {source.kind === "documents" ? (
              <>
                <label className="export-format-option">
                  <input
                    type="radio"
                    name="export-format"
                    value="json"
                    checked={format === "json"}
                    onChange={() => setFormat("json")}
                  />
                  JSON Array
                </label>
                <label className="export-format-option">
                  <input
                    type="radio"
                    name="export-format"
                    value="csv"
                    checked={format === "csv"}
                    onChange={() => setFormat("csv")}
                  />
                  CSV
                </label>
              </>
            ) : (
              <>
                <label className="export-format-option">
                  <input
                    type="radio"
                    name="export-format"
                    value="json_hits"
                    checked={format === "json_hits"}
                    onChange={() => setFormat("json_hits")}
                  />
                  JSON Hits Only
                </label>
                <label className="export-format-option">
                  <input
                    type="radio"
                    name="export-format"
                    value="json_full"
                    checked={format === "json_full"}
                    onChange={() => setFormat("json_full")}
                  />
                  JSON Full Response
                </label>
              </>
            )}
          </fieldset>

          {/* Error */}
          {error && (
            <p className="export-error" role="alert">{error}</p>
          )}

          {/* Success */}
          {result && (
            <div className="export-success" role="status">
              <span className="status-badge status-badge--succeeded">Exported successfully</span>
              <p className="export-success-detail">
                {result.rowCount.toLocaleString()} row{result.rowCount !== 1 ? "s" : ""} written to:
              </p>
              <p className="export-success-path">{result.path}</p>
            </div>
          )}

          {/* Actions */}
          <div className="export-actions">
            {result ? (
              <button className="btn btn--primary" onClick={onClose}>
                Done
              </button>
            ) : (
              <button
                className="btn btn--primary"
                onClick={() => void handleExport()}
                disabled={busy}
              >
                {busy ? "Exporting…" : "Choose destination and export…"}
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

export type { ExportSource as default };
