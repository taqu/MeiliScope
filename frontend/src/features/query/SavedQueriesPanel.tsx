import { useState, useEffect, useCallback } from "react";
import {
  SavedQueryService,
  type SavedQuery,
  type UpdateSavedQueryRequest,
} from "../../services/savedquery";

// ─── Props ────────────────────────────────────────────────────────────────────

export interface SavedQueriesPanelProps {
  currentQueryText: string;
  currentIndexUID: string;
  connectionId: string;
  onLoad: (query: SavedQuery) => void;
  onClose: () => void;
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

function formatDate(iso: string): string {
  try {
    const d = new Date(iso);
    const now = new Date();
    const diffMs = now.getTime() - d.getTime();
    const diffMin = Math.floor(diffMs / 60000);
    if (diffMin < 1) return "just now";
    if (diffMin < 60) return `${diffMin}m ago`;
    const diffH = Math.floor(diffMin / 60);
    if (diffH < 24) return `${diffH}h ago`;
    const diffD = Math.floor(diffH / 24);
    if (diffD < 30) return `${diffD}d ago`;
    return d.toLocaleDateString();
  } catch {
    return iso;
  }
}

function truncate(text: string, max: number): string {
  if (text.length <= max) return text;
  return text.slice(0, max) + "…";
}

// ─── Single entry component ───────────────────────────────────────────────────

interface EntryProps {
  entry: SavedQuery;
  isSelected: boolean;
  onLoad: (entry: SavedQuery) => void;
  onDuplicate: (id: string) => void;
  onDelete: (id: string) => void;
  onRename: (id: string, newName: string) => void;
}

function SavedQueryEntry({
  entry,
  isSelected,
  onLoad,
  onDuplicate,
  onDelete,
  onRename,
}: EntryProps) {
  const [deleteConfirm, setDeleteConfirm] = useState(false);
  const [renaming, setRenaming] = useState(false);
  const [renameValue, setRenameValue] = useState(entry.name);

  function handleRenameSubmit() {
    const trimmed = renameValue.trim();
    if (trimmed && trimmed !== entry.name) {
      onRename(entry.id, trimmed);
    }
    setRenaming(false);
  }

  function handleRenameKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key === "Enter") {
      handleRenameSubmit();
    } else if (e.key === "Escape") {
      setRenameValue(entry.name);
      setRenaming(false);
    }
  }

  return (
    <div className={`sq-entry${isSelected ? " sq-entry--selected" : ""}`}>
      <div className="sq-entry-name">
        {renaming ? (
          <input
            className="form-input"
            type="text"
            value={renameValue}
            autoFocus
            onChange={(e) => setRenameValue(e.target.value)}
            onBlur={handleRenameSubmit}
            onKeyDown={handleRenameKeyDown}
          />
        ) : (
          <span title={entry.name}>{entry.name}</span>
        )}
      </div>
      <div className="sq-entry-meta">
        <span className="sq-entry-index">{entry.indexUid}</span>
        <span className="sq-entry-date">{formatDate(entry.updatedAt)}</span>
      </div>
      {entry.description && (
        <div className="sq-entry-desc">{truncate(entry.description, 80)}</div>
      )}
      <div className="sq-entry-actions">
        <button
          className="btn btn--primary btn--sm"
          onClick={() => onLoad(entry)}
          title="Load query into editor"
        >
          Load
        </button>
        <button
          className="btn btn--ghost btn--sm"
          onClick={() => onDuplicate(entry.id)}
          title="Duplicate this query"
        >
          Duplicate
        </button>
        <button
          className="btn btn--ghost btn--sm"
          onClick={() => {
            setRenameValue(entry.name);
            setRenaming(true);
          }}
          title="Rename"
        >
          Rename
        </button>
        {deleteConfirm ? (
          <>
            <span className="sq-delete-confirm-text">Delete?</span>
            <button
              className="btn btn--danger btn--sm"
              onClick={() => { onDelete(entry.id); setDeleteConfirm(false); }}
            >
              Yes
            </button>
            <button
              className="btn btn--ghost btn--sm"
              onClick={() => setDeleteConfirm(false)}
            >
              No
            </button>
          </>
        ) : (
          <button
            className="btn btn--danger btn--sm"
            onClick={() => setDeleteConfirm(true)}
            title="Delete"
          >
            Delete
          </button>
        )}
      </div>
    </div>
  );
}

// ─── Main panel ───────────────────────────────────────────────────────────────

export function SavedQueriesPanel({
  currentQueryText,
  currentIndexUID,
  connectionId,
  onLoad,
  onClose,
}: SavedQueriesPanelProps) {
  const [queries, setQueries] = useState<SavedQuery[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Save form
  const [saveName, setSaveName] = useState("");
  const [saveDesc, setSaveDesc] = useState("");
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);

  // Tracks the currently "loaded" query id for the Update button
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [updating, setUpdating] = useState(false);
  const [updateError, setUpdateError] = useState<string | null>(null);

  // ── Data loading ──────────────────────────────────────────────────────────

  const loadList = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const list = await SavedQueryService.list();
      setQueries(list);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Failed to load saved queries.");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadList();
  }, [loadList]);

  // ── Handlers ──────────────────────────────────────────────────────────────

  async function handleSave() {
    const name = saveName.trim();
    if (!name) {
      setSaveError("Name is required.");
      return;
    }
    setSaving(true);
    setSaveError(null);
    try {
      const created = await SavedQueryService.create({
        name,
        description: saveDesc.trim() || undefined,
        connectionId,
        indexUid: currentIndexUID,
        queryJson: currentQueryText,
      });
      setSaveName("");
      setSaveDesc("");
      setSelectedId(created.id);
      await loadList();
    } catch (e: unknown) {
      setSaveError(e instanceof Error ? e.message : "Save failed.");
    } finally {
      setSaving(false);
    }
  }

  async function handleUpdate() {
    if (!selectedId) return;
    const existing = queries.find((q) => q.id === selectedId);
    if (!existing) return;
    setUpdating(true);
    setUpdateError(null);
    try {
      const req: UpdateSavedQueryRequest = {
        name: existing.name,
        description: existing.description ?? null,
        connectionId,
        indexUid: currentIndexUID,
        queryJson: currentQueryText,
      };
      await SavedQueryService.update(selectedId, req);
      await loadList();
    } catch (e: unknown) {
      setUpdateError(e instanceof Error ? e.message : "Update failed.");
    } finally {
      setUpdating(false);
    }
  }

  async function handleDuplicate(id: string) {
    try {
      await SavedQueryService.duplicate(id);
      await loadList();
    } catch {
      /* ignore */
    }
  }

  async function handleDelete(id: string) {
    try {
      await SavedQueryService.delete(id);
      if (selectedId === id) setSelectedId(null);
      await loadList();
    } catch {
      /* ignore */
    }
  }

  async function handleRename(id: string, newName: string) {
    const existing = queries.find((q) => q.id === id);
    if (!existing) return;
    try {
      const req: UpdateSavedQueryRequest = {
        name: newName,
        description: existing.description ?? null,
        connectionId: existing.connectionId,
        indexUid: existing.indexUid,
        queryJson: existing.queryJson,
      };
      await SavedQueryService.update(id, req);
      await loadList();
    } catch {
      /* ignore */
    }
  }

  function handleLoad(entry: SavedQuery) {
    setSelectedId(entry.id);
    onLoad(entry);
  }

  const selectedEntry = selectedId ? queries.find((q) => q.id === selectedId) : null;

  // ── Render ────────────────────────────────────────────────────────────────

  return (
    <div className="sq-panel">
      {/* Header */}
      <div className="sq-panel-header">
        <span className="sq-panel-title">Saved Queries</span>
        <button className="btn btn--ghost btn--sm" onClick={onClose} title="Close">
          ✕
        </button>
      </div>

      {/* Save new query form */}
      <div className="sq-save-form">
        <div className="sq-save-form-title">Save current query</div>
        <div className="sb-field-row">
          <label htmlFor="sq-name">Name</label>
          <input
            id="sq-name"
            className="form-input"
            type="text"
            value={saveName}
            onChange={(e) => setSaveName(e.target.value)}
            placeholder="Query name…"
            disabled={saving}
          />
        </div>
        <div className="sb-field-row">
          <label htmlFor="sq-desc">Description</label>
          <input
            id="sq-desc"
            className="form-input"
            type="text"
            value={saveDesc}
            onChange={(e) => setSaveDesc(e.target.value)}
            placeholder="Optional description…"
            disabled={saving}
          />
        </div>
        {saveError && <div className="sb-warning">{saveError}</div>}
        <div className="sq-save-form-actions">
          <button
            className="btn btn--primary btn--sm"
            onClick={() => void handleSave()}
            disabled={saving || !saveName.trim()}
          >
            {saving ? "Saving…" : "Save"}
          </button>
        </div>
      </div>

      {/* Update existing — shown when a query is currently selected/loaded */}
      {selectedEntry && (
        <div className="sq-update-form">
          <div className="sq-update-info">
            Loaded: <strong>{selectedEntry.name}</strong>
          </div>
          {updateError && <div className="sb-warning">{updateError}</div>}
          <button
            className="btn btn--ghost btn--sm"
            onClick={() => void handleUpdate()}
            disabled={updating}
          >
            {updating ? "Updating…" : `Update "${selectedEntry.name}"`}
          </button>
        </div>
      )}

      {/* Query list */}
      <div className="sq-list">
        {loading && <div className="sq-list-status">Loading…</div>}
        {error && !loading && <div className="sb-warning">{error}</div>}
        {!loading && !error && queries.length === 0 && (
          <div className="sq-list-empty">No saved queries yet.</div>
        )}
        {!loading &&
          queries.map((entry) => (
            <SavedQueryEntry
              key={entry.id}
              entry={entry}
              isSelected={entry.id === selectedId}
              onLoad={handleLoad}
              onDuplicate={(id) => void handleDuplicate(id)}
              onDelete={(id) => void handleDelete(id)}
              onRename={(id, newName) => void handleRename(id, newName)}
            />
          ))}
      </div>
    </div>
  );
}
