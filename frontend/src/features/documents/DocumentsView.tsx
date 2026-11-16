import { useState, useEffect, useCallback, useRef } from "react";
import { IndexService, type IndexSummary } from "../../services/index";
import {
  DocumentService,
  type DocumentPage,
  type DocumentRecord,
} from "../../services/document";
import { useConnectionContext } from "../../context/ConnectionContext";

// ─── JSON parser (precision-preserving) ───────────────────────────────────────

type JNode =
  | { kind: "null" }
  | { kind: "bool"; value: boolean }
  | { kind: "number"; raw: string }
  | { kind: "string"; value: string }
  | { kind: "array"; items: JNode[] }
  | { kind: "object"; entries: [string, JNode][] };

interface PS {
  t: string;
  p: number;
}

function ws(s: PS) {
  while (s.p < s.t.length && " \t\r\n".includes(s.t[s.p])) s.p++;
}

function parseStr(s: PS): string {
  s.p++; // skip opening "
  let r = "";
  while (s.p < s.t.length && s.t[s.p] !== '"') {
    if (s.t[s.p] === "\\") {
      s.p++;
      const e = s.t[s.p++];
      if (e === "u") {
        r += String.fromCharCode(parseInt(s.t.slice(s.p, s.p + 4), 16));
        s.p += 4;
      } else {
        r +=
          e === "n" ? "\n" :
          e === "r" ? "\r" :
          e === "t" ? "\t" :
          e === "b" ? "\b" :
          e === "f" ? "\f" : e;
      }
    } else {
      r += s.t[s.p++];
    }
  }
  s.p++; // skip closing "
  return r;
}

function parseVal(s: PS): JNode {
  ws(s);
  const c = s.t[s.p];
  if (c === "n") { s.p += 4; return { kind: "null" }; }
  if (c === "t") { s.p += 4; return { kind: "bool", value: true }; }
  if (c === "f") { s.p += 5; return { kind: "bool", value: false }; }
  if (c === '"') return { kind: "string", value: parseStr(s) };
  if (c === "[") {
    s.p++;
    const items: JNode[] = [];
    ws(s);
    if (s.t[s.p] === "]") { s.p++; return { kind: "array", items }; }
    while (s.p < s.t.length) {
      items.push(parseVal(s));
      ws(s);
      if (s.t[s.p] === "]") { s.p++; return { kind: "array", items }; }
      s.p++;
    }
    return { kind: "array", items };
  }
  if (c === "{") {
    s.p++;
    const entries: [string, JNode][] = [];
    ws(s);
    if (s.t[s.p] === "}") { s.p++; return { kind: "object", entries }; }
    while (s.p < s.t.length) {
      ws(s);
      const key = parseStr(s);
      ws(s);
      s.p++; // skip :
      const val = parseVal(s);
      entries.push([key, val]);
      ws(s);
      if (s.t[s.p] === "}") { s.p++; return { kind: "object", entries }; }
      s.p++;
    }
    return { kind: "object", entries };
  }
  // number
  const st = s.p;
  if (s.t[s.p] === "-") s.p++;
  while (s.p < s.t.length && s.t[s.p] >= "0" && s.t[s.p] <= "9") s.p++;
  if (s.p < s.t.length && s.t[s.p] === ".") {
    s.p++;
    while (s.p < s.t.length && s.t[s.p] >= "0" && s.t[s.p] <= "9") s.p++;
  }
  if (s.p < s.t.length && (s.t[s.p] === "e" || s.t[s.p] === "E")) {
    s.p++;
    if (s.t[s.p] === "+" || s.t[s.p] === "-") s.p++;
    while (s.p < s.t.length && s.t[s.p] >= "0" && s.t[s.p] <= "9") s.p++;
  }
  return { kind: "number", raw: s.t.slice(st, s.p) };
}

function parseJsonDoc(text: string): JNode | null {
  try {
    const s: PS = { t: text, p: 0 };
    ws(s);
    return parseVal(s);
  } catch {
    return null;
  }
}

function escapePointerToken(key: string): string {
  return key.replace(/~/g, "~0").replace(/\//g, "~1");
}

function nodeToJson(node: JNode): string {
  if (node.kind === "null") return "null";
  if (node.kind === "bool") return node.value ? "true" : "false";
  if (node.kind === "number") return node.raw;
  if (node.kind === "string") return JSON.stringify(node.value);
  if (node.kind === "array") return "[" + node.items.map(nodeToJson).join(",") + "]";
  return "{" + node.entries.map(([k, v]) => JSON.stringify(k) + ":" + nodeToJson(v)).join(",") + "}";
}

function nodePreview(node: JNode): string {
  if (node.kind === "null") return "null";
  if (node.kind === "bool") return node.value ? "true" : "false";
  if (node.kind === "number") return node.raw;
  if (node.kind === "string") {
    const s = node.value.length > 40 ? node.value.slice(0, 40) + "…" : node.value;
    return `"${s}"`;
  }
  if (node.kind === "array") return `[ ${node.items.length} items ]`;
  return `{ ${node.entries.length} keys }`;
}

async function copyText(text: string) {
  try { await navigator.clipboard.writeText(text); } catch { /* ignore */ }
}

// ─── JSON Inspector ───────────────────────────────────────────────────────────

interface JsonNodeProps {
  node: JNode;
  depth: number;
  path: string;
  label?: string;
  defaultExpanded?: boolean;
}

function JsonNodeView({ node, depth, path, label, defaultExpanded = false }: JsonNodeProps) {
  const [expanded, setExpanded] = useState(defaultExpanded || depth < 2);
  const [copied, setCopied] = useState(false);

  const handleCopy = async (text: string, e: React.MouseEvent) => {
    e.stopPropagation();
    await copyText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 1200);
  };

  const isContainer = node.kind === "array" || node.kind === "object";
  const isEmpty =
    (node.kind === "array" && node.items.length === 0) ||
    (node.kind === "object" && node.entries.length === 0);

  const keyEl = label !== undefined ? (
    <span className="json-key" title={path}>{label}: </span>
  ) : null;

  if (!isContainer || isEmpty) {
    return (
      <div className="json-row" style={{ paddingLeft: depth * 14 + 4 }}>
        {keyEl}
        <span className={`json-${node.kind}`}>{nodePreview(node)}</span>
        <button className="json-copy-btn" onClick={(e) => void handleCopy(nodeToJson(node), e)}>
          {copied ? "✓" : "⎘"}
        </button>
      </div>
    );
  }

  if (!expanded) {
    return (
      <div className="json-row" style={{ paddingLeft: depth * 14 + 4 }}>
        {keyEl}
        <button className="json-toggle" onClick={() => setExpanded(true)}>▶</button>
        <span className="json-preview">{nodePreview(node)}</span>
        <button className="json-copy-btn" onClick={(e) => void handleCopy(nodeToJson(node), e)}>
          {copied ? "✓" : "⎘"}
        </button>
      </div>
    );
  }

  if (node.kind === "array") {
    return (
      <div className="json-container">
        <div className="json-row" style={{ paddingLeft: depth * 14 + 4 }}>
          {keyEl}
          <button className="json-toggle" onClick={() => setExpanded(false)}>▼</button>
          <span className="json-bracket">[</span>
          <button className="json-copy-btn" onClick={(e) => void handleCopy(nodeToJson(node), e)}>
            {copied ? "✓" : "⎘"}
          </button>
        </div>
        <div>
          {node.items.map((item, i) => (
            <JsonNodeView key={i} node={item} depth={depth + 1}
              path={`${path}[${i}]`} label={String(i)} />
          ))}
        </div>
        <div className="json-row" style={{ paddingLeft: depth * 14 + 4 }}>
          <span className="json-bracket">]</span>
        </div>
      </div>
    );
  }

  return (
    <div className="json-container">
      <div className="json-row" style={{ paddingLeft: depth * 14 + 4 }}>
        {keyEl}
        <button className="json-toggle" onClick={() => setExpanded(false)}>▼</button>
        <span className="json-bracket">{"{"}</span>
        <button className="json-copy-btn" onClick={(e) => void handleCopy(nodeToJson(node), e)}>
          {copied ? "✓" : "⎘"}
        </button>
      </div>
      <div>
        {node.entries.map(([k, v]) => (
          <JsonNodeView key={k} node={v} depth={depth + 1}
            path={`${path}/${escapePointerToken(k)}`} label={k} />
        ))}
      </div>
      <div className="json-row" style={{ paddingLeft: depth * 14 + 4 }}>
        <span className="json-bracket">{"}"}</span>
      </div>
    </div>
  );
}

function InspectorPanel({ record }: { record: DocumentRecord | null }) {
  const [copiedAll, setCopiedAll] = useState(false);

  const handleCopyAll = async () => {
    if (!record) return;
    try {
      await copyText(JSON.stringify(JSON.parse(record.rawJson), null, 2));
    } catch {
      await copyText(record.rawJson);
    }
    setCopiedAll(true);
    setTimeout(() => setCopiedAll(false), 1200);
  };

  if (!record) {
    return (
      <div className="inspector-panel inspector-panel--empty">
        <p className="inspector-empty-msg">Select a row to inspect its JSON.</p>
      </div>
    );
  }

  const root = parseJsonDoc(record.rawJson);

  return (
    <div className="inspector-panel">
      <div className="inspector-header">
        <span className="inspector-title">JSON Inspector</span>
        <button className="btn btn--ghost btn--sm" onClick={() => void handleCopyAll()}>
          {copiedAll ? "Copied ✓" : "Copy JSON"}
        </button>
      </div>
      <div className="inspector-tree">
        {root ? (
          <JsonNodeView node={root} depth={0} path="" defaultExpanded={true} />
        ) : (
          <p className="inspector-parse-error">Failed to parse document JSON.</p>
        )}
      </div>
    </div>
  );
}

// ─── Cell rendering ────────────────────────────────────────────────────────────

function CellValue({ raw }: { raw: string | undefined }) {
  if (raw === undefined) return <span className="cell-missing">—</span>;
  if (raw === "null") return <span className="cell-null">null</span>;
  if (raw === "true") return <span className="cell-bool cell-bool--true">true</span>;
  if (raw === "false") return <span className="cell-bool cell-bool--false">false</span>;
  if (raw.startsWith('"')) {
    const inner = raw.slice(1, raw.length - 1).replace(/\\"/g, '"').replace(/\\\\/g, "\\");
    const display = inner.length > 80 ? inner.slice(0, 80) + "…" : inner;
    return <span className="cell-string" title={inner.length > 80 ? inner : undefined}>{display}</span>;
  }
  if (raw.startsWith("{")) return <span className="cell-object">{"{ … }"}</span>;
  if (raw.startsWith("[")) return <span className="cell-array">{"[ … ]"}</span>;
  return <span className="cell-number">{raw}</span>;
}

// ─── Document Table ────────────────────────────────────────────────────────────

interface DocumentTableProps {
  page: DocumentPage;
  columns: string[];
  visibleColumns: Set<string>;
  selectedRawJson: string | null;
  onSelect: (record: DocumentRecord) => void;
}

function DocumentTable({ page, columns, visibleColumns, selectedRawJson, onSelect }: DocumentTableProps) {
  const displayCols = columns.filter((c) => visibleColumns.has(c));

  if (!page.items || page.items.length === 0) {
    return <div className="doc-table-empty"><p>No documents on this page.</p></div>;
  }

  return (
    <div className="doc-table-wrapper">
      <table className="doc-table">
        <thead>
          <tr>
            {displayCols.map((col) => (
              <th key={col} className="doc-th">{col}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {page.items.map((rec, i) => (
            <tr
              key={i}
              className={`doc-tr${selectedRawJson === rec.rawJson ? " doc-tr--selected" : ""}`}
              onClick={() => onSelect(rec)}
              aria-selected={selectedRawJson === rec.rawJson}
            >
              {displayCols.map((col) => (
                <td key={col} className="doc-td">
                  <CellValue raw={rec.fields?.[col]} />
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

// ─── Column Manager ────────────────────────────────────────────────────────────

function ColumnManager({
  columns, visible, onChange, onClose,
}: {
  columns: string[]; visible: Set<string>;
  onChange: (s: Set<string>) => void; onClose: () => void;
}) {
  const toggle = (col: string) => {
    const next = new Set(visible);
    if (next.has(col)) { if (next.size > 1) next.delete(col); }
    else next.add(col);
    onChange(next);
  };

  return (
    <div className="col-manager">
      <div className="col-manager-header">
        <span>Columns</span>
        <button className="btn btn--ghost btn--sm" onClick={onClose}>✕</button>
      </div>
      <div className="col-manager-list">
        {columns.map((col) => (
          <label key={col} className="col-manager-item">
            <input type="checkbox" checked={visible.has(col)} onChange={() => toggle(col)} />
            <span>{col}</span>
          </label>
        ))}
      </div>
      <div className="col-manager-footer">
        <button className="btn btn--ghost btn--sm" onClick={() => onChange(new Set(columns))}>
          Show All
        </button>
      </div>
    </div>
  );
}

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
        <DocumentTable page={page} columns={discoveredCols} visibleColumns={visibleCols}
          selectedRawJson={selectedRawJson} onSelect={onSelectRecord} />
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
          <InspectorPanel record={selectedRecord} />
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
