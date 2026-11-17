import type { DocumentRecord } from "../services/document";

// ─── Cell value ───────────────────────────────────────────────────────────────

export function CellValue({ raw }: { raw: string | undefined }) {
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

// ─── Column manager ───────────────────────────────────────────────────────────

interface ColumnManagerProps {
  columns: string[];
  visible: Set<string>;
  onChange: (s: Set<string>) => void;
  onClose: () => void;
}

export function ColumnManager({ columns, visible, onChange, onClose }: ColumnManagerProps) {
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

// ─── Hit table ────────────────────────────────────────────────────────────────

interface HitTableProps {
  records: DocumentRecord[];
  columns: string[];
  visibleColumns: Set<string>;
  selectedRawJson: string | null;
  onSelect: (record: DocumentRecord) => void;
  emptyMessage?: string;
}

export function HitTable({ records, columns, visibleColumns, selectedRawJson, onSelect, emptyMessage = "No records on this page." }: HitTableProps) {
  const displayCols = columns.filter((c) => visibleColumns.has(c));

  if (records.length === 0) {
    return <div className="doc-table-empty"><p>{emptyMessage}</p></div>;
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
          {records.map((rec, i) => (
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
