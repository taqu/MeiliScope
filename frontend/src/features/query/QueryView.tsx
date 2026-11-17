import { useState, useEffect, useCallback, useRef } from "react";
import { EditorView, basicSetup } from "codemirror";
import { EditorState } from "@codemirror/state";
import { json } from "@codemirror/lang-json";
import { keymap } from "@codemirror/view";
import { oneDark } from "@codemirror/theme-one-dark";
import { IndexService, type IndexSummary } from "../../services/index";
import {
  QueryService,
  type SearchResult,
  type HistoryEntry,
  type DocumentRecord,
} from "../../services/query";
import { useConnectionContext } from "../../context/ConnectionContext";
import { InspectorPanel } from "../../components/json-tree";
import { HitTable, ColumnManager } from "../../components/doc-table";
import { SearchBuilder } from "./SearchBuilder";
import { SavedQueriesPanel } from "./SavedQueriesPanel";
import { ExportDialog, type ExportSource } from "../export/ExportDialog";
import type { SavedQuery } from "../../services/savedquery";
import { useKeyboardShortcuts } from "../../hooks/useKeyboardShortcuts";

const DEFAULT_QUERY = `{
  "q": "",
  "limit": 20
}`;

// ─── JSON editor (CodeMirror 6) ───────────────────────────────────────────────

interface JsonEditorProps {
  value: string;
  onChange: (v: string) => void;
  onExecute: () => void;
}

function JsonEditor({ value, onChange, onExecute }: JsonEditorProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const viewRef = useRef<EditorView | null>(null);
  const onExecuteRef = useRef(onExecute);
  const onChangeRef = useRef(onChange);
  const externalValueRef = useRef(value);

  onExecuteRef.current = onExecute;
  onChangeRef.current = onChange;

  useEffect(() => {
    if (!containerRef.current) return;
    const view = new EditorView({
      state: EditorState.create({
        doc: externalValueRef.current,
        extensions: [
          basicSetup,
          json(),
          oneDark,
          keymap.of([{
            key: "Ctrl-Enter",
            run: () => { onExecuteRef.current(); return true; },
          }]),
          EditorView.updateListener.of((update) => {
            if (update.docChanged) {
              const text = update.state.doc.toString();
              externalValueRef.current = text;
              onChangeRef.current(text);
            }
          }),
          EditorView.theme({
            "&": { height: "100%", fontSize: "13px" },
            ".cm-scroller": { overflow: "auto", fontFamily: "ui-monospace, 'Cascadia Code', Consolas, monospace" },
          }),
        ],
      }),
      parent: containerRef.current,
    });
    viewRef.current = view;
    return () => { view.destroy(); viewRef.current = null; };
  }, []); // only create once

  // Sync external value changes (e.g., history restore, reset)
  useEffect(() => {
    const view = viewRef.current;
    if (!view) return;
    const current = view.state.doc.toString();
    if (current === value) return;
    externalValueRef.current = value;
    view.dispatch({
      changes: { from: 0, to: view.state.doc.length, insert: value },
    });
  }, [value]);

  return <div ref={containerRef} className="cm-editor-wrapper" />;
}

// ─── Summary bar ──────────────────────────────────────────────────────────────

function SummaryBar({ result }: { result: SearchResult }) {
  if (result.error) {
    const e = result.error;
    return (
      <div className="query-summary query-summary--error">
        <span className="query-summary__badge query-summary__badge--error">{e.kind || "error"}</span>
        <span className="query-summary__msg">{e.message || "Search failed"}</span>
        {result.clientDurationMs > 0 && (
          <span className="query-summary__time">{result.clientDurationMs}ms</span>
        )}
      </div>
    );
  }

  const { meta, hits, clientDurationMs } = result;
  const hitCount = hits.length;
  const totalHits = meta.totalHits ?? meta.estimatedTotalHits;
  const isEstimated = meta.totalHits == null && meta.estimatedTotalHits != null;

  return (
    <div className="query-summary query-summary--ok">
      <span className="query-summary__badge query-summary__badge--ok">OK</span>
      <span className="query-summary__hits">
        {hitCount} returned
        {totalHits != null && (
          <> / {isEstimated ? "~" : ""}{totalHits.toLocaleString()} {isEstimated ? "estimated" : "total"}</>
        )}
      </span>
      <span className="query-summary__sep">·</span>
      <span className="query-summary__time">
        {meta.processingTimeMs}ms server / {clientDurationMs}ms client
      </span>
      {meta.query != null && meta.query !== "" && (
        <>
          <span className="query-summary__sep">·</span>
          <span className="query-summary__q" title={meta.query}>
            q: "{meta.query.length > 40 ? meta.query.slice(0, 40) + "…" : meta.query}"
          </span>
        </>
      )}
    </div>
  );
}

// ─── Error detail panel ───────────────────────────────────────────────────────

function ErrorDetailPanel({ result }: { result: SearchResult }) {
  const e = result.error!;
  return (
    <div className="query-error-panel">
      <div className="query-error-header">Search Error</div>
      <dl className="query-error-dl">
        {e.kind && <><dt>Kind</dt><dd><span className={`query-error-kind query-error-kind--${e.kind}`}>{e.kind}</span></dd></>}
        {e.message && <><dt>Message</dt><dd>{e.message}</dd></>}
        {e.code && <><dt>Code</dt><dd className="query-error-code">{e.code}</dd></>}
        {e.type && <><dt>Type</dt><dd>{e.type}</dd></>}
        {e.status > 0 && <><dt>HTTP Status</dt><dd>{e.status}</dd></>}
        {e.link && (
          <><dt>Docs</dt><dd><span className="query-error-link">{e.link}</span></dd></>
        )}
      </dl>
      <div className="query-error-meta">
        <span>Index: {result.indexUid}</span>
        {result.clientDurationMs > 0 && <span>{result.clientDurationMs}ms</span>}
      </div>
    </div>
  );
}

// ─── Facet panel ──────────────────────────────────────────────────────────────

function FacetPanel({ result }: { result: SearchResult }) {
  const { facetDistribution, facetStats } = result.meta;
  const [open, setOpen] = useState(false);

  if (!facetDistribution && !facetStats) return null;

  const facets = Object.keys(facetDistribution ?? facetStats ?? {});
  if (facets.length === 0) return null;

  return (
    <div className="query-facet-panel">
      <button className="query-facet-toggle" onClick={() => setOpen((v) => !v)}>
        {open ? "▼" : "▶"} Facets ({facets.length})
      </button>
      {open && (
        <div className="query-facet-content">
          {facets.map((attr) => (
            <div key={attr} className="query-facet-attr">
              <div className="query-facet-attr-name">{attr}</div>
              {facetStats?.[attr] && (
                <div className="query-facet-stats">
                  min: {facetStats[attr].min} · max: {facetStats[attr].max}
                </div>
              )}
              {facetDistribution?.[attr] && (
                <div className="query-facet-values">
                  {Object.entries(facetDistribution[attr])
                    .sort(([, a], [, b]) => b - a)
                    .slice(0, 20)
                    .map(([val, count]) => (
                      <div key={val} className="query-facet-value-row">
                        <span className="query-facet-value-label">{val}</span>
                        <span className="query-facet-value-count">{count}</span>
                      </div>
                    ))}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

// ─── History panel ────────────────────────────────────────────────────────────

interface HistoryPanelProps {
  entries: HistoryEntry[];
  onRestore: (entry: HistoryEntry) => void;
  onRerun: (entry: HistoryEntry) => void;
  onDelete: (id: string) => void;
  onClearAll: () => void;
}

function formatTime(iso: string): string {
  try {
    const d = new Date(iso);
    const now = new Date();
    const diffMs = now.getTime() - d.getTime();
    const diffMin = Math.floor(diffMs / 60000);
    if (diffMin < 1) return "just now";
    if (diffMin < 60) return `${diffMin}m ago`;
    const diffH = Math.floor(diffMin / 60);
    if (diffH < 24) return `${diffH}h ago`;
    return d.toLocaleDateString();
  } catch {
    return iso;
  }
}

function querySnippet(json: string): string {
  try {
    const obj = JSON.parse(json) as Record<string, unknown>;
    if (obj["q"] != null) return `q: "${String(obj["q"]).slice(0, 30)}"`;
    const keys = Object.keys(obj).slice(0, 3).join(", ");
    return keys || "{}";
  } catch {
    return json.slice(0, 40);
  }
}

function HistoryPanel({ entries, onRestore, onRerun, onDelete, onClearAll }: HistoryPanelProps) {
  const [clearConfirm, setClearConfirm] = useState(false);

  if (entries.length === 0) {
    return <div className="query-history-empty">No history yet.</div>;
  }

  return (
    <div className="query-history-list">
      <div className="query-history-actions">
        {clearConfirm ? (
          <span className="query-history-confirm">
            Clear all? {" "}
            <button className="btn btn--danger btn--sm" onClick={() => { onClearAll(); setClearConfirm(false); }}>Yes</button>
            {" "}
            <button className="btn btn--ghost btn--sm" onClick={() => setClearConfirm(false)}>No</button>
          </span>
        ) : (
          <button className="btn btn--ghost btn--sm" onClick={() => setClearConfirm(true)}>Clear All</button>
        )}
      </div>
      {entries.map((e) => (
        <div key={e.id} className={`query-history-entry query-history-entry--${e.status}`}>
          <div className="query-history-entry-meta">
            <span className={`query-history-status query-history-status--${e.status}`}>
              {e.status === "success" ? "✓" : "✗"}
            </span>
            <span className="query-history-index">{e.indexUid}</span>
            {e.connectionName && <span className="query-history-conn">{e.connectionName}</span>}
            <span className="query-history-time">{formatTime(e.executedAt)}</span>
            {e.hitsCount != null && e.status === "success" && (
              <span className="query-history-hits">{e.hitsCount} hits</span>
            )}
            {e.processingTimeMs != null && e.status === "success" && (
              <span className="query-history-proc">{e.processingTimeMs}ms</span>
            )}
          </div>
          <div className="query-history-entry-query">{querySnippet(e.queryJson)}</div>
          {e.errorSummary && (
            <div className="query-history-entry-error">{e.errorSummary}</div>
          )}
          <div className="query-history-entry-actions">
            <button className="btn btn--ghost btn--sm" onClick={() => onRestore(e)}>Restore</button>
            <button className="btn btn--primary btn--sm" onClick={() => onRerun(e)}>Rerun</button>
            <button className="btn btn--ghost btn--sm" onClick={() => onDelete(e.id)}>✕</button>
          </div>
        </div>
      ))}
    </div>
  );
}

// ─── Main QueryView ───────────────────────────────────────────────────────────

function ConnectedQueryConsole({ connectionID }: { connectionID: string }) {
  const [indexUID, setIndexUID] = useState("");
  const [indexOptions, setIndexOptions] = useState<IndexSummary[]>([]);
  const [indexLoading, setIndexLoading] = useState(false);

  const [queryText, setQueryText] = useState(DEFAULT_QUERY);
  const [validationError, setValidationError] = useState<string | null>(null);

  const [executing, setExecuting] = useState(false);
  const execIdRef = useRef(0);
  const [lastResult, setLastResult] = useState<SearchResult | null>(null);

  const [resultView, setResultView] = useState<"table" | "json">("table");
  const [selectedHit, setSelectedHit] = useState<DocumentRecord | null>(null);
  const [discoveredCols, setDiscoveredCols] = useState<string[]>([]);
  const [visibleCols, setVisibleCols] = useState<Set<string>>(new Set());
  const [showColManager, setShowColManager] = useState(false);

  const [history, setHistory] = useState<HistoryEntry[]>([]);
  const [historyOpen, setHistoryOpen] = useState(false);

  const [editorMode, setEditorMode] = useState<"json" | "builder">("json");
  const [showSavedQueries, setShowSavedQueries] = useState(false);
  const [exportSource, setExportSource] = useState<ExportSource | null>(null);

  useKeyboardShortcuts([
    {
      key: "s", ctrl: true,
      description: "Open Saved Queries",
      skipInEditor: false,
      handler: () => setShowSavedQueries((v) => !v),
    },
    {
      key: "F", shift: true, alt: true,
      description: "Format JSON",
      skipInEditor: false,
      handler: () => {
        try {
          const formatted = JSON.stringify(JSON.parse(queryText), null, 2);
          setQueryText(formatted);
          setValidationError(null);
        } catch {
          setValidationError("Cannot format: invalid JSON.");
        }
      },
    },
    {
      key: "f", shift: true, alt: true,
      description: "Format JSON",
      skipInEditor: false,
      handler: () => {
        try {
          const formatted = JSON.stringify(JSON.parse(queryText), null, 2);
          setQueryText(formatted);
          setValidationError(null);
        } catch {
          setValidationError("Cannot format: invalid JSON.");
        }
      },
    },
  ]);

  // Load indexes
  const loadIndexes = useCallback(async () => {
    setIndexLoading(true);
    try {
      const r = await IndexService.listIndexes(0, 100);
      setIndexOptions(r.items ?? []);
    } catch {
      setIndexOptions([]);
    } finally {
      setIndexLoading(false);
    }
  }, []);

  // Load history
  const loadHistory = useCallback(async () => {
    try {
      const h = await QueryService.getHistory();
      setHistory(h);
    } catch {
      /* ignore */
    }
  }, []);

  useEffect(() => {
    void loadIndexes();
    void loadHistory();
  }, [loadIndexes, loadHistory]);

  // Clear results when connection changes
  useEffect(() => {
    setLastResult(null);
    setSelectedHit(null);
    execIdRef.current++;
  }, [connectionID]);

  const validateQuery = (text: string): boolean => {
    try {
      const parsed = JSON.parse(text);
      if (typeof parsed !== "object" || Array.isArray(parsed) || parsed === null) {
        setValidationError("Query must be a JSON object, not an array or scalar.");
        return false;
      }
      setValidationError(null);
      return true;
    } catch (e: unknown) {
      setValidationError(e instanceof SyntaxError ? e.message : "Invalid JSON.");
      return false;
    }
  };

  const handleExecute = useCallback(async () => {
    if (executing) return;
    if (!indexUID.trim()) {
      setValidationError("Select or enter an index UID.");
      return;
    }
    if (!validateQuery(queryText)) return;

    const myId = ++execIdRef.current;
    setExecuting(true);
    setLastResult(null);
    setSelectedHit(null);

    try {
      const result = await QueryService.executeSearch(indexUID.trim(), queryText);
      if (execIdRef.current !== myId) return;
      setLastResult(result);

      // Discover columns from hits
      const keySet = new Set<string>();
      for (const rec of result.hits ?? [])
        for (const k of Object.keys(rec.fields ?? {})) keySet.add(k);
      const cols = Array.from(keySet);
      setDiscoveredCols(cols);
      setVisibleCols(new Set(cols));
    } catch (e: unknown) {
      if (execIdRef.current !== myId) return;
      const msg = e instanceof Error ? e.message : String(e);
      setValidationError(msg);
    } finally {
      if (execIdRef.current === myId) {
        setExecuting(false);
        void loadHistory();
      }
    }
  }, [executing, indexUID, queryText, loadHistory]);

  const handleCancel = () => {
    execIdRef.current++;
    setExecuting(false);
    void QueryService.cancelSearch();
  };

  const handleFormat = () => {
    try {
      const formatted = JSON.stringify(JSON.parse(queryText), null, 2);
      setQueryText(formatted);
      setValidationError(null);
    } catch {
      setValidationError("Cannot format: invalid JSON.");
    }
  };

  const handleClear = () => {
    setQueryText("{}");
    setValidationError(null);
  };

  const handleReset = () => {
    setQueryText(DEFAULT_QUERY);
    setValidationError(null);
  };

  const handleRestore = (entry: HistoryEntry) => {
    setIndexUID(entry.indexUid);
    setQueryText(entry.queryJson);
    setValidationError(null);
  };

  const handleRerun = async (entry: HistoryEntry) => {
    setIndexUID(entry.indexUid);
    setQueryText(entry.queryJson);
    setValidationError(null);
    // Execute after state update — use a slight delay via queueMicrotask
    // Actually, just call execute directly with the entry values
    if (executing) return;
    const myId = ++execIdRef.current;
    setExecuting(true);
    setLastResult(null);
    setSelectedHit(null);
    try {
      const result = await QueryService.executeSearch(entry.indexUid, entry.queryJson);
      if (execIdRef.current !== myId) return;
      setLastResult(result);
      const keySet = new Set<string>();
      for (const rec of result.hits ?? [])
        for (const k of Object.keys(rec.fields ?? {})) keySet.add(k);
      const cols = Array.from(keySet);
      setDiscoveredCols(cols);
      setVisibleCols(new Set(cols));
    } catch (e: unknown) {
      if (execIdRef.current !== myId) return;
      setValidationError(e instanceof Error ? e.message : String(e));
    } finally {
      if (execIdRef.current === myId) {
        setExecuting(false);
        void loadHistory();
      }
    }
  };

  const handleDeleteHistory = async (id: string) => {
    try {
      await QueryService.deleteHistoryEntry(id);
      await loadHistory();
    } catch { /* ignore */ }
  };

  const handleClearHistory = async () => {
    try {
      await QueryService.clearHistory();
      await loadHistory();
    } catch { /* ignore */ }
  };

  const hasResults = lastResult !== null;
  const hasError = lastResult?.error != null;
  const hasHits = hasResults && !hasError && (lastResult.hits?.length ?? 0) > 0;

  return (
    <div className="query-view">
      {/* Toolbar */}
      <div className="query-toolbar">
        <div className="query-index-row">
          {editorMode === "json" && (
            <>
              <input
                className="form-input query-index-input"
                list="query-index-list"
                value={indexUID}
                onChange={(e) => setIndexUID(e.target.value)}
                placeholder={indexLoading ? "Loading indexes…" : "Index UID…"}
                disabled={executing}
              />
              <datalist id="query-index-list">
                {indexOptions.map((i) => <option key={i.uid} value={i.uid} />)}
              </datalist>
            </>
          )}
          <div className="query-mode-toggle">
            <button
              className={`btn btn--sm${editorMode === "json" ? " btn--primary" : " btn--ghost"}`}
              onClick={() => setEditorMode("json")}
              title="JSON editor mode"
            >JSON</button>
            <button
              className={`btn btn--sm${editorMode === "builder" ? " btn--primary" : " btn--ghost"}`}
              onClick={() => setEditorMode("builder")}
              title="Visual query builder"
            >Builder</button>
          </div>
        </div>
        <div className="query-toolbar-actions">
          {editorMode === "json" && (
            <>
              <button
                className="btn btn--primary btn--sm"
                onClick={() => void handleExecute()}
                disabled={executing || !indexUID.trim()}
              >
                {executing ? "Running…" : "▶ Execute"}
              </button>
              {executing && (
                <button className="btn btn--ghost btn--sm" onClick={handleCancel}>
                  ✕ Cancel
                </button>
              )}
              <button className="btn btn--ghost btn--sm" onClick={handleFormat} disabled={executing}>
                {"{ }"} Format
              </button>
              <button className="btn btn--ghost btn--sm" onClick={handleClear} disabled={executing}>
                Clear
              </button>
              <button className="btn btn--ghost btn--sm" onClick={handleReset} disabled={executing}>
                Reset
              </button>
            </>
          )}
          <button
            className={`btn btn--sm${showSavedQueries ? " btn--primary" : " btn--ghost"}`}
            onClick={() => setShowSavedQueries((v) => !v)}
            title="Saved queries"
          >
            ★ Saved
          </button>
        </div>
      </div>

      {/* Saved Queries Panel */}
      {showSavedQueries && (
        <SavedQueriesPanel
          currentQueryText={queryText}
          currentIndexUID={indexUID}
          connectionId={connectionID}
          onLoad={(q: SavedQuery) => {
            setIndexUID(q.indexUid);
            setQueryText(q.queryJson);
            setValidationError(null);
            setShowSavedQueries(false);
          }}
          onClose={() => setShowSavedQueries(false)}
        />
      )}

      {/* Main split */}
      <div className="query-main">
        {/* Editor panel */}
        <div className="query-editor-panel">
          {editorMode === "json" ? (
            <>
              <JsonEditor
                value={queryText}
                onChange={(v) => { setQueryText(v); setValidationError(null); }}
                onExecute={() => void handleExecute()}
              />
              {validationError && (
                <div className="query-validation-error">{validationError}</div>
              )}
            </>
          ) : (
            <SearchBuilder
              queryText={queryText}
              onQueryChange={(v) => { setQueryText(v); setValidationError(null); }}
              indexUID={indexUID}
              onIndexChange={setIndexUID}
              indexOptions={indexOptions}
              onExecute={() => void handleExecute()}
              executing={executing}
            />
          )}
        </div>

        {/* Results panel */}
        <div className="query-results-panel">
          {/* Summary */}
          {hasResults && <SummaryBar result={lastResult!} />}

          {/* Error details */}
          {hasError && <ErrorDetailPanel result={lastResult!} />}

          {/* Success results */}
          {hasResults && !hasError && (
            <>
              {/* Tab bar + col manager */}
              <div className="query-tabs-row">
                <div className="query-tabs">
                  <button
                    className={`query-tab${resultView === "table" ? " query-tab--active" : ""}`}
                    onClick={() => setResultView("table")}
                  >Table</button>
                  <button
                    className={`query-tab${resultView === "json" ? " query-tab--active" : ""}`}
                    onClick={() => setResultView("json")}
                  >JSON</button>
                </div>
                <div className="query-tabs-actions">
                  {resultView === "table" && discoveredCols.length > 0 && (
                    <button className="btn btn--ghost btn--sm" onClick={() => setShowColManager((v) => !v)}>
                      Columns
                    </button>
                  )}
                  <button
                    className="btn btn--ghost btn--sm"
                    onClick={() => setExportSource({
                      kind: "search",
                      indexUID,
                      searchResponseJSON: lastResult!.rawJson,
                      hitCount: lastResult!.hits?.length ?? 0,
                    })}
                    title="Export results"
                  >
                    ↓ Export
                  </button>
                </div>
              </div>

              {resultView === "table" && (
                <div className="query-table-area">
                  {showColManager && (
                    <ColumnManager
                      columns={discoveredCols}
                      visible={visibleCols}
                      onChange={setVisibleCols}
                      onClose={() => setShowColManager(false)}
                    />
                  )}
                  <div className="query-table-with-inspector">
                    <HitTable
                      records={lastResult!.hits ?? []}
                      columns={discoveredCols}
                      visibleColumns={visibleCols}
                      selectedRawJson={selectedHit?.rawJson ?? null}
                      onSelect={setSelectedHit}
                      emptyMessage="No hits returned."
                    />
                    <InspectorPanel
                      rawJson={selectedHit?.rawJson ?? null}
                      emptyMessage="Select a row to inspect."
                    />
                  </div>
                  <FacetPanel result={lastResult!} />
                </div>
              )}

              {resultView === "json" && (
                <div className="query-json-area">
                  <InspectorPanel
                    rawJson={lastResult!.rawJson}
                    title="Response JSON"
                  />
                </div>
              )}
            </>
          )}

          {/* Empty / loading states */}
          {!hasResults && !executing && (
            <div className="query-idle">
              <p>Enter a query and press <kbd>Ctrl+Enter</kbd> or click <strong>Execute</strong>.</p>
            </div>
          )}
          {executing && (
            <div className="query-running">
              <div className="query-running__spinner" />
              <p>Executing search…</p>
            </div>
          )}

          {/* Hits count hint when hits exist */}
          {hasHits && resultView === "table" && (
            <div className="query-hits-hint">
              {lastResult!.hits.length} hit{lastResult!.hits.length !== 1 ? "s" : ""} returned
              {discoveredCols.length > 0 && ` · ${discoveredCols.length} fields`}
            </div>
          )}
        </div>
      </div>

      {/* History panel */}
      <div className="query-history-panel">
        <button
          className="query-history-toggle"
          onClick={() => setHistoryOpen((v) => !v)}
        >
          {historyOpen ? "▼" : "▶"} History ({history.length})
        </button>
        {historyOpen && (
          <HistoryPanel
            entries={history}
            onRestore={handleRestore}
            onRerun={(e) => void handleRerun(e)}
            onDelete={(id) => void handleDeleteHistory(id)}
            onClearAll={() => void handleClearHistory()}
          />
        )}
      </div>

      {/* Export dialog */}
      {exportSource && (
        <ExportDialog
          source={exportSource}
          connectionName={connectionID}
          onClose={() => setExportSource(null)}
        />
      )}
    </div>
  );
}

export function QueryView() {
  const { activeConn } = useConnectionContext();
  const isConnected = activeConn?.state === "connected" || activeConn?.state === "limited_access";

  if (!activeConn || !isConnected) {
    return (
      <div className="placeholder-view">
        <h2 className="placeholder-title">Query Console</h2>
        <p className="placeholder-desc">
          {activeConn
            ? `Connection state: ${activeConn.state}. Try reconnecting.`
            : "Select a connection to use the Query Console."}
        </p>
      </div>
    );
  }

  return <ConnectedQueryConsole key={activeConn.id} connectionID={activeConn.id} />;
}
