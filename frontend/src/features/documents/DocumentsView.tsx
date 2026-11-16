import { useState, useEffect, useCallback, useRef } from "react";
import { IndexService, type IndexSummary } from "../../services/index";
import {
  DocumentService,
  type DocumentPage,
  type DocumentRecord,
} from "../../services/document";
import { useConnectionContext } from "../../context/ConnectionContext";
import { InspectorPanel } from "../../components/json-tree";
import { HitTable, ColumnManager } from "../../components/doc-table";
import { DocumentEditor, type EditorMode } from "./DocumentEditor";

// ─── Index List Panel ──────────────────────────────────────────────────────────

function DocIndexPanel({
  selectedUID, onSelect,
}: {
  selectedUID: string | null;
  onSelect: (uid: string, primaryKey: string | null) => void;
}) {
  const [items, setItems] = useState<IndexSummary[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const activeRef = useRef(true);
  const PAGE = 20;

  const load = useCallback(async (off: number) => {
    setLoading(true);
    setError(null);
    try {
      const r = await IndexService.listIndexes(off, PAGE);
      if (!activeRef.current) return;
      setItems(r.items ?? []);
      setTotal(r.total);
      setOffset(r.offset);
    } catch (e: unknown) {
      if (!activeRef.current) return;
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      if (activeRef.current) setLoading(false);
    }
  }, []);

  useEffect(() => {
    activeRef.current = true;
    void load(0);
    return () => { activeRef.current = false; };
  }, [load]);

  return (
    <div className="doc-idx-panel">
      <div className="doc-idx-header">
        <span className="doc-idx-header__title">Indexes</span>
        <button className="btn btn--ghost btn--sm" onClick={() => void load(offset)} disabled={loading}>
          {loading ? "…" : "↺"}
        </button>
      </div>
      {error && <p className="doc-idx-error">{error}</p>}
      {!loading && !error && items.length === 0 && <p className="doc-idx-empty">No indexes.</p>}
      <ul className="doc-idx-list" role="listbox">
        {items.map((idx) => (
          <li key={idx.uid} role="option" aria-selected={selectedUID === idx.uid}
            className={`doc-idx-item${selectedUID === idx.uid ? " doc-idx-item--selected" : ""}`}
            onClick={() => onSelect(idx.uid, idx.primaryKey ?? null)}>
            {idx.uid}
          </li>
        ))}
      </ul>
      {(offset > 0 || offset + PAGE < total) && (
        <div className="doc-idx-pagination">
          <button className="btn btn--ghost btn--sm"
            onClick={() => void load(Math.max(0, offset - PAGE))} disabled={offset <= 0 || loading}>←</button>
          <button className="btn btn--ghost btn--sm"
            onClick={() => void load(offset + PAGE)} disabled={offset + PAGE >= total || loading}>→</button>
        </div>
      )}
    </div>
  );
}

// ─── Browser Panel ─────────────────────────────────────────────────────────────

const PAGE_SIZES = [25, 50, 100, 200];

interface BrowserPanelProps {
  indexUID: string;
  primaryKey: string | null;
  selectedRawJson: string | null;
  selectedRecord: DocumentRecord | null;
  onSelectRecord: (rec: DocumentRecord) => void;
  onNew: () => void;
  onEdit: () => void;
  onDelete: () => void;
  onRegisterRefresh: (fn: () => void) => void;
}

function BrowserPanelWithActions({
  indexUID,
  selectedRawJson,
  selectedRecord,
  onSelectRecord,
  onNew,
  onEdit,
  onDelete,
  onRegisterRefresh,
}: BrowserPanelProps) {
  const [page, setPage] = useState<DocumentPage | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [offset, setOffset] = useState(0);
  const [limit, setLimit] = useState(50);
  const [discoveredCols, setDiscoveredCols] = useState<string[]>([]);
  const [visibleCols, setVisibleCols] = useState<Set<string>>(new Set());
  const [showColManager, setShowColManager] = useState(false);
  const loadGen = useRef(0);

  const load = useCallback(async (off: number, lim: number) => {
    const gen = ++loadGen.current;
    setLoading(true);
    setError(null);
    try {
      const result = await DocumentService.listDocuments(indexUID, { offset: off, limit: lim });
      if (loadGen.current !== gen) return;
      setPage(result);
      const keySet = new Set<string>();
      for (const rec of result.items ?? [])
        for (const k of Object.keys(rec.fields ?? {})) keySet.add(k);
      setDiscoveredCols((prev) => Array.from(new Set([...prev, ...keySet])));
      setVisibleCols((prev) => {
        if (prev.size === 0) return new Set(keySet);
        const next = new Set(prev);
        for (const k of keySet) next.add(k);
        return next;
      });
    } catch (e: unknown) {
      if (loadGen.current !== gen) return;
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      if (loadGen.current === gen) setLoading(false);
    }
  }, [indexUID]);

  useEffect(() => {
    setPage(null);
    setOffset(0);
    setDiscoveredCols([]);
    setVisibleCols(new Set());
    void load(0, limit);
  }, [indexUID, load, limit]);

  const goPage = (newOff: number) => {
    setOffset(newOff);
    void load(newOff, limit);
  };

  const changeLimit = (newLim: number) => {
    setLimit(newLim);
    setOffset(0);
    void load(0, newLim);
  };

  const refresh = () => void load(offset, limit);

  useEffect(() => {
    onRegisterRefresh(refresh);
  }, [onRegisterRefresh, refresh]); // eslint-disable-line react-hooks/exhaustive-deps

  const total = Number(page?.total ?? 0);
  const hasPrev = offset > 0;
  const hasMore = offset + limit < total;

  return (
    <div className="browser-panel">
      <div className="browser-header">
        <span className="browser-index-name">{indexUID}</span>
        <div className="browser-header-actions">
          <button className="btn btn--primary btn--sm" onClick={onNew}>
            + New
          </button>
          <button className="btn btn--ghost btn--sm" onClick={onEdit} disabled={!selectedRecord}>
            Edit
          </button>
          <button className="btn btn--danger btn--sm" onClick={onDelete} disabled={!selectedRecord}>
            Delete
          </button>
          {discoveredCols.length > 0 && (
            <button className="btn btn--ghost btn--sm" onClick={() => setShowColManager((v) => !v)}>
              Columns
            </button>
          )}
          <button className="btn btn--ghost btn--sm" onClick={refresh} disabled={loading}>
            {loading ? "…" : "↺ Refresh"}
          </button>
        </div>
      </div>

      {showColManager && (
        <ColumnManager columns={discoveredCols} visible={visibleCols}
          onChange={setVisibleCols} onClose={() => setShowColManager(false)} />
      )}

      {discoveredCols.length > 0 && (
        <p className="browser-col-hint">
          Columns discovered from this page. Some fields may not appear until browsed.
        </p>
      )}

      {error && <p className="form-error browser-error">{error}</p>}
      {loading && !page && <div className="browser-loading"><p>Loading documents…</p></div>}

      {page && !error && (
        <HitTable
          records={page.items ?? []}
          columns={discoveredCols}
          visibleColumns={visibleCols}
          selectedRawJson={selectedRawJson}
          onSelect={onSelectRecord}
          emptyMessage="No documents on this page."
        />
      )}

      {page && (
        <div className="browser-pagination">
          <button className="btn btn--ghost btn--sm" onClick={() => goPage(0)} disabled={!hasPrev || loading}>«</button>
          <button className="btn btn--ghost btn--sm" onClick={() => goPage(Math.max(0, offset - limit))} disabled={!hasPrev || loading}>‹ Prev</button>
          <span className="browser-page-info">
            {total === 0
              ? "No documents"
              : `${offset + 1}–${Math.min(offset + limit, total)} of ${total.toLocaleString()}`}
          </span>
          <button className="btn btn--ghost btn--sm" onClick={() => goPage(offset + limit)} disabled={!hasMore || loading}>Next ›</button>
          <select className="browser-page-size" value={limit}
            onChange={(e) => changeLimit(Number(e.target.value))}>
            {PAGE_SIZES.map((s) => <option key={s} value={s}>{s} / page</option>)}
          </select>
        </div>
      )}
    </div>
  );
}

// ─── Main DocumentsView ────────────────────────────────────────────────────────

interface SelectedIndex { uid: string; primaryKey: string | null; }

interface EditorState {
  mode: EditorMode;
  initialDocJson?: string;
  documentID?: string;
}

function getDocumentID(rec: DocumentRecord, primaryKeyField: string | null): string | null {
  if (!primaryKeyField) return null;
  const val = (rec.fields as Record<string, unknown>)?.[primaryKeyField];
  return val !== undefined ? String(val) : null;
}

function ConnectedDocumentBrowser({ connectionID: _cid }: { connectionID: string }) {
  const [selected, setSelected] = useState<SelectedIndex | null>(null);
  const [selectedRecord, setSelectedRecord] = useState<DocumentRecord | null>(null);
  const [editorState, setEditorState] = useState<EditorState | null>(null);
  const browserRefreshRef = useRef<(() => void) | null>(null);

  const handleIndexSelect = (uid: string, primaryKey: string | null) => {
    setSelected({ uid, primaryKey });
    setSelectedRecord(null);
  };

  const handleNewDocument = () => {
    setEditorState({ mode: "create" });
  };

  const handleEditDocument = () => {
    if (!selectedRecord) return;
    const docID = selected?.primaryKey ? getDocumentID(selectedRecord, selected.primaryKey) : undefined;
    setEditorState({
      mode: "edit",
      initialDocJson: selectedRecord.rawJson,
      documentID: docID ?? undefined,
    });
  };

  const handleDeleteDocument = () => {
    if (!selectedRecord) return;
    const docID = selected?.primaryKey ? getDocumentID(selectedRecord, selected.primaryKey) : undefined;
    setEditorState({
      mode: "delete",
      initialDocJson: selectedRecord.rawJson,
      documentID: docID ?? undefined,
    });
  };

  const handleEditorClose = () => setEditorState(null);

  const handleEditorSuccess = (_taskUid: number) => {
    // Editor transitions to task panel; keep editor open
  };

  const handleRefreshNeeded = () => {
    browserRefreshRef.current?.();
    setSelectedRecord(null);
  };

  return (
    <div className="doc-browser">
      <DocIndexPanel selectedUID={selected?.uid ?? null} onSelect={handleIndexSelect} />

      {selected ? (
        <div className="browser-with-inspector">
          <BrowserPanelWithActions
            key={selected.uid}
            indexUID={selected.uid}
            primaryKey={selected.primaryKey}
            selectedRawJson={selectedRecord?.rawJson ?? null}
            onSelectRecord={setSelectedRecord}
            onNew={handleNewDocument}
            onEdit={handleEditDocument}
            onDelete={handleDeleteDocument}
            selectedRecord={selectedRecord}
            onRegisterRefresh={(fn) => { browserRefreshRef.current = fn; }}
          />
          <InspectorPanel rawJson={selectedRecord?.rawJson ?? null} />
        </div>
      ) : (
        <div className="browser-no-index">
          <p className="browser-no-index-msg">Select an index to browse documents.</p>
        </div>
      )}

      {editorState && selected && (
        <DocumentEditor
          mode={editorState.mode}
          indexUID={selected.uid}
          primaryKeyField={selected.primaryKey}
          initialDocJson={editorState.initialDocJson}
          documentID={editorState.documentID}
          onClose={handleEditorClose}
          onSuccess={handleEditorSuccess}
          onRefreshNeeded={handleRefreshNeeded}
        />
      )}
    </div>
  );
}

export function DocumentsView() {
  const { activeConn } = useConnectionContext();
  const isConnected = activeConn?.state === "connected" || activeConn?.state === "limited_access";

  if (!activeConn || !isConnected) {
    return (
      <div className="placeholder-view">
        <h2 className="placeholder-title">Document Browser</h2>
        <p className="placeholder-desc">
          {activeConn
            ? `Connection state: ${activeConn.state}. Try reconnecting.`
            : "Select a connection to browse documents."}
        </p>
      </div>
    );
  }

  return <ConnectedDocumentBrowser key={activeConn.id} connectionID={activeConn.id} />;
}
