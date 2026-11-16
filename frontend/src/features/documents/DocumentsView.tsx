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
  selectedRawJson: string | null;
  onSelectRecord: (rec: DocumentRecord) => void;
}

function BrowserPanel({ indexUID, selectedRawJson, onSelectRecord }: BrowserPanelProps) {
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

  const total = Number(page?.total ?? 0);
  const hasPrev = offset > 0;
  const hasMore = offset + limit < total;

  return (
    <div className="browser-panel">
      <div className="browser-header">
        <span className="browser-index-name">{indexUID}</span>
        <div className="browser-header-actions">
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

function ConnectedDocumentBrowser({ connectionID: _cid }: { connectionID: string }) {
  const [selected, setSelected] = useState<SelectedIndex | null>(null);
  const [selectedRecord, setSelectedRecord] = useState<DocumentRecord | null>(null);

  const handleIndexSelect = (uid: string, primaryKey: string | null) => {
    setSelected({ uid, primaryKey });
    setSelectedRecord(null);
  };

  return (
    <div className="doc-browser">
      <DocIndexPanel selectedUID={selected?.uid ?? null} onSelect={handleIndexSelect} />

      {selected ? (
        <div className="browser-with-inspector">
          <BrowserPanel
            key={selected.uid}
            indexUID={selected.uid}
            selectedRawJson={selectedRecord?.rawJson ?? null}
            onSelectRecord={setSelectedRecord}
          />
          <InspectorPanel rawJson={selectedRecord?.rawJson ?? null} />
        </div>
      ) : (
        <div className="browser-no-index">
          <p className="browser-no-index-msg">Select an index to browse documents.</p>
        </div>
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
