import { useState, useEffect, useCallback, useRef } from "react";
import {
  IndexService,
  type IndexSummary,
  type IndexStatistics,
} from "../../services/index";
import { useConnectionContext } from "../../context/ConnectionContext";

const PAGE_SIZE = 20;

function formatDate(iso: string): string {
  if (!iso) return "—";
  try {
    return new Date(iso).toLocaleString();
  } catch {
    return iso;
  }
}

function formatNumber(n: number): string {
  return n.toLocaleString();
}

// --- Index list panel ---

interface IndexListProps {
  selectedUID: string | null;
  onSelect: (uid: string) => void;
}

function IndexListPanel({ selectedUID, onSelect }: IndexListProps) {
  const [items, setItems] = useState<IndexSummary[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const activeRef = useRef(true);

  const load = useCallback(async (off: number) => {
    setLoading(true);
    setError(null);
    try {
      const result = await IndexService.listIndexes(off, PAGE_SIZE);
      if (!activeRef.current) return;
      setItems(result.items ?? []);
      setTotal(result.total);
      setOffset(result.offset);
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
    return () => {
      activeRef.current = false;
    };
  }, [load]);

  const handleRefresh = useCallback(() => void load(offset), [load, offset]);
  const handlePrev = useCallback(
    () => void load(Math.max(0, offset - PAGE_SIZE)),
    [load, offset]
  );
  const handleNext = useCallback(
    () => void load(offset + PAGE_SIZE),
    [load, offset]
  );

  const hasMore = offset + PAGE_SIZE < total;
  const hasPrev = offset > 0;

  return (
    <div className="idx-list-panel">
      <div className="idx-list-header">
        <span className="idx-list-header__title">Indexes</span>
        {total > 0 && (
          <span className="idx-list-header__count">{formatNumber(total)}</span>
        )}
        <button
          className="btn btn--ghost btn--sm idx-list-header__refresh"
          onClick={handleRefresh}
          disabled={loading}
          aria-label="Refresh index list"
        >
          {loading ? "…" : "↺"}
        </button>
      </div>

      {error && <p className="idx-list-error">{error}</p>}

      {!loading && !error && items.length === 0 && (
        <div className="empty-state empty-state--sm">
          <p className="empty-state__msg">No indexes found.</p>
        </div>
      )}

      <ul className="idx-list" role="listbox" aria-label="Index list">
        {items.map((idx) => (
          <li
            key={idx.uid}
            role="option"
            aria-selected={selectedUID === idx.uid}
            className={`idx-item${selectedUID === idx.uid ? " idx-item--selected" : ""}`}
            onClick={() => onSelect(idx.uid)}
          >
            <span className="idx-item__uid">{idx.uid}</span>
          </li>
        ))}
      </ul>

      {(hasPrev || hasMore) && (
        <div className="idx-list-pagination">
          <button
            className="btn btn--ghost btn--sm"
            onClick={handlePrev}
            disabled={!hasPrev || loading}
          >
            ← Prev
          </button>
          <span className="idx-list-pagination__info">
            {offset + 1}–{Math.min(offset + PAGE_SIZE, total)} of{" "}
            {formatNumber(total)}
          </span>
          <button
            className="btn btn--ghost btn--sm"
            onClick={handleNext}
            disabled={!hasMore || loading}
          >
            Next →
          </button>
        </div>
      )}
    </div>
  );
}

// --- Field distribution table ---

function FieldDistributionTable({ stats }: { stats: IndexStatistics }) {
  if (stats.statsUnavailable) {
    return (
      <p className="idx-detail-unavailable">
        {stats.statsError ?? "Statistics unavailable."}
      </p>
    );
  }
  if (!stats.fieldDistribution || stats.fieldDistribution.length === 0) {
    return <p className="idx-detail-muted">No field data available.</p>;
  }
  return (
    <table className="field-dist-table">
      <thead>
        <tr>
          <th>Field</th>
          <th>Documents</th>
        </tr>
      </thead>
      <tbody>
        {stats.fieldDistribution.map((entry) => (
          <tr key={entry.name}>
            <td>{entry.name}</td>
            <td>{formatNumber(Number(entry.count))}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

// --- Index detail panel ---

function IndexDetailPanel({ uid }: { uid: string }) {
  const [index, setIndex] = useState<IndexSummary | null>(null);
  const [stats, setStats] = useState<IndexStatistics | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setIndex(null);
    setStats(null);
    setError(null);
    setLoading(true);

    let active = true;
    Promise.all([IndexService.getIndex(uid), IndexService.getIndexStats(uid)])
      .then(([idx, st]) => {
        if (!active) return;
        setIndex(idx);
        setStats(st);
      })
      .catch((e: unknown) => {
        if (!active) return;
        setError(e instanceof Error ? e.message : String(e));
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
  }, [uid]);

  if (loading) {
    return (
      <div className="idx-detail-panel idx-detail-panel--loading">
        <p className="idx-detail-muted">Loading…</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="idx-detail-panel">
        <p className="form-error">{error}</p>
      </div>
    );
  }

  if (!index) return null;

  return (
    <div className="idx-detail-panel">
      <h2 className="idx-detail-title">{index.uid}</h2>

      <dl className="idx-detail-dl">
        <dt>Primary Key</dt>
        <dd>{index.primaryKey ?? <em>Not set</em>}</dd>

        <dt>Created</dt>
        <dd>{formatDate(index.createdAt)}</dd>

        <dt>Updated</dt>
        <dd>{formatDate(index.updatedAt)}</dd>

        {stats && !stats.statsUnavailable && (
          <>
            <dt>Documents</dt>
            <dd>{formatNumber(Number(stats.numberOfDocuments))}</dd>

            <dt>Indexing</dt>
            <dd>{stats.isIndexing ? "In progress" : "Idle"}</dd>
          </>
        )}
      </dl>

      {stats && (
        <>
          <h3 className="idx-detail-section-title">Field Distribution</h3>
          <FieldDistributionTable stats={stats} />
        </>
      )}
    </div>
  );
}

// --- Main IndexesView ---

// ConnectedExplorer is only rendered when a connection is active.
// It uses the connection ID as a React key so panels are remounted on switch,
// preventing stale data from appearing under a new connection.
function ConnectedExplorer({ connectionID }: { connectionID: string }) {
  const [selectedUID, setSelectedUID] = useState<string | null>(null);

  return (
    <div className="idx-explorer" key={connectionID}>
      <IndexListPanel selectedUID={selectedUID} onSelect={setSelectedUID} />

      <div className="idx-detail-pane">
        {selectedUID ? (
          <IndexDetailPanel key={selectedUID} uid={selectedUID} />
        ) : (
          <div className="idx-detail-panel idx-detail-panel--empty">
            <p className="idx-detail-muted">Select an index to view details.</p>
          </div>
        )}
      </div>
    </div>
  );
}

export function IndexesView() {
  const { activeConn } = useConnectionContext();

  const isConnected =
    activeConn?.state === "connected" ||
    activeConn?.state === "limited_access";

  if (!activeConn || !isConnected) {
    return (
      <div className="placeholder-view">
        <h2 className="placeholder-title">Index Explorer</h2>
        <p className="placeholder-desc">
          {activeConn
            ? `Connection state: ${activeConn.state}. Try reconnecting from the Connections view.`
            : "Select a connection to browse indexes."}
        </p>
      </div>
    );
  }

  return <ConnectedExplorer connectionID={activeConn.id} />;
}
