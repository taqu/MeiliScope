import { useState, useCallback, type FormEvent } from "react";
import {
  ConnectionService,
  type ConnectionSummary,
  type ConnectionTestResult,
} from "../../services/connection";
import { useConnectionContext } from "../../context/ConnectionContext";

// --- Connection form ---

interface ConnFormData {
  name: string;
  url: string;
  apiKey: string;
  credentialMode: "keyring" | "session";
}

const emptyForm = (): ConnFormData => ({
  name: "",
  url: "",
  apiKey: "",
  credentialMode: "keyring",
});

function profileToForm(p: ConnectionSummary): ConnFormData {
  return {
    name: p.name,
    url: p.url,
    apiKey: "",
    credentialMode: p.credentialMode === "session" ? "session" : "keyring",
  };
}

type ModalMode = { kind: "create" } | { kind: "edit"; id: string };

// --- State indicator ---

function StateBadge({ state }: { state: string }) {
  const labels: Record<string, string> = {
    connected: "Connected",
    connecting: "Connecting…",
    limited_access: "Limited",
    auth_required: "Auth Required",
    error: "Error",
    disconnected: "Disconnected",
  };
  return (
    <span className={`state-badge state-badge--${state}`}>
      {labels[state] ?? state}
    </span>
  );
}

// --- Connection form modal ---

interface ConnFormProps {
  mode: ModalMode;
  initial: ConnFormData;
  isSecureStore: boolean;
  onClose: () => void;
  onSaved: () => void;
}

function ConnectionFormModal({
  mode,
  initial,
  isSecureStore,
  onClose,
  onSaved,
}: ConnFormProps) {
  const [form, setForm] = useState<ConnFormData>(initial);
  const [showKey, setShowKey] = useState(false);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState<ConnectionTestResult | null>(
    null
  );
  const [error, setError] = useState<string | null>(null);

  function set<K extends keyof ConnFormData>(key: K, value: ConnFormData[K]) {
    setForm((f) => ({ ...f, [key]: value }));
    setTestResult(null);
  }

  const handleTest = useCallback(async () => {
    if (!form.url) {
      setError("Server URL is required to test the connection.");
      return;
    }
    setTesting(true);
    setError(null);
    setTestResult(null);
    try {
      let result: ConnectionTestResult;
      if (mode.kind === "edit") {
        result = await ConnectionService.testConnection(mode.id, form.apiKey);
      } else {
        result = await ConnectionService.testNewConnection(
          form.url,
          form.apiKey
        );
      }
      setTestResult(result);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setTesting(false);
    }
  }, [mode, form.url, form.apiKey]);

  const handleSubmit = useCallback(
    async (e: FormEvent) => {
      e.preventDefault();
      if (!form.name.trim()) {
        setError("Name is required.");
        return;
      }
      if (!form.url.trim()) {
        setError("Server URL is required.");
        return;
      }
      setSaving(true);
      setError(null);
      try {
        const credMode =
          !isSecureStore || form.credentialMode === "session"
            ? "session"
            : "keyring";
        const req = {
          name: form.name.trim(),
          url: form.url.trim(),
          apiKey: form.apiKey,
          credentialMode: credMode,
        };
        if (mode.kind === "create") {
          await ConnectionService.createConnection(req);
        } else {
          await ConnectionService.updateConnection(mode.id, req);
        }
        onSaved();
      } catch (e: unknown) {
        setError(e instanceof Error ? e.message : String(e));
      } finally {
        setSaving(false);
      }
    },
    [mode, form, isSecureStore, onSaved]
  );

  return (
    <div className="modal-overlay" role="dialog" aria-modal="true">
      <div className="modal">
        <h2 className="modal__title">
          {mode.kind === "create" ? "New Connection" : "Edit Connection"}
        </h2>

        <form onSubmit={handleSubmit} noValidate>
          <label className="form-label">
            Display Name
            <input
              className="form-input"
              type="text"
              value={form.name}
              onChange={(e) => set("name", e.target.value)}
              placeholder="My Meilisearch"
              autoFocus
              required
            />
          </label>

          <label className="form-label">
            Server URL
            <input
              className="form-input"
              type="url"
              value={form.url}
              onChange={(e) => set("url", e.target.value)}
              placeholder="http://localhost:7700"
              required
            />
          </label>

          <label className="form-label">
            API Key
            {mode.kind === "edit" && (
              <span className="form-label__hint">
                Leave blank to keep existing key
              </span>
            )}
            <div className="form-input-row">
              <input
                className="form-input"
                type={showKey ? "text" : "password"}
                value={form.apiKey}
                onChange={(e) => set("apiKey", e.target.value)}
                placeholder={
                  mode.kind === "edit"
                    ? "Enter new key to replace"
                    : "Optional"
                }
                autoComplete="off"
              />
              <button
                type="button"
                className="btn btn--ghost btn--sm"
                onClick={() => setShowKey((v) => !v)}
                aria-label={showKey ? "Hide API key" : "Show API key"}
              >
                {showKey ? "Hide" : "Show"}
              </button>
            </div>
          </label>

          {isSecureStore && (
            <label className="form-check">
              <input
                type="checkbox"
                checked={form.credentialMode === "keyring"}
                onChange={(e) =>
                  set("credentialMode", e.target.checked ? "keyring" : "session")
                }
              />
              Save API key securely (OS credential store)
            </label>
          )}

          {!isSecureStore && (
            <p className="form-hint form-hint--warn">
              Secure credential storage is unavailable on this system. The API
              key will be held in memory only and not saved between sessions.
            </p>
          )}

          {testResult && (
            <div
              className={`test-result test-result--${testResult.serverReachable ? "ok" : "fail"}`}
            >
              <strong>Test result:</strong> {testResult.message}
              {testResult.serverVersion && (
                <span className="test-result__version">
                  {" "}
                  · v{testResult.serverVersion}
                </span>
              )}
            </div>
          )}

          {error && <p className="form-error">{error}</p>}

          <div className="modal__actions">
            <button
              type="button"
              className="btn btn--ghost"
              onClick={handleTest}
              disabled={testing || saving}
            >
              {testing ? "Testing…" : "Test Connection"}
            </button>
            <div className="modal__actions-right">
              <button
                type="button"
                className="btn btn--ghost"
                onClick={onClose}
                disabled={saving}
              >
                Cancel
              </button>
              <button
                type="submit"
                className="btn btn--primary"
                disabled={saving}
              >
                {saving ? "Saving…" : "Save"}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
}

// --- Connection row ---

interface ConnRowProps {
  conn: ConnectionSummary;
  isActive: boolean;
  onConnect: (id: string) => void;
  onEdit: (id: string) => void;
  onDelete: (id: string) => void;
  onDisconnect: () => void;
}

function ConnectionRow({
  conn,
  isActive,
  onConnect,
  onEdit,
  onDelete,
  onDisconnect,
}: ConnRowProps) {
  const [deleting, setDeleting] = useState(false);

  const handleDelete = useCallback(async () => {
    if (!confirm(`Delete connection "${conn.name}"?`)) return;
    setDeleting(true);
    try {
      await ConnectionService.deleteConnection(conn.id);
      onDelete(conn.id);
    } catch (e: unknown) {
      alert(e instanceof Error ? e.message : String(e));
    } finally {
      setDeleting(false);
    }
  }, [conn.id, conn.name, onDelete]);

  return (
    <div className={`conn-row${isActive ? " conn-row--active" : ""}`}>
      <div className="conn-row__info">
        <span className="conn-row__name">{conn.name}</span>
        <span className="conn-row__url">{conn.url}</span>
      </div>
      <div className="conn-row__meta">
        <StateBadge state={isActive ? conn.state : "disconnected"} />
      </div>
      <div className="conn-row__actions">
        {isActive && conn.state === "connected" ? (
          <button
            className="btn btn--ghost btn--sm"
            onClick={onDisconnect}
          >
            Disconnect
          </button>
        ) : (
          <button
            className="btn btn--primary btn--sm"
            onClick={() => onConnect(conn.id)}
          >
            Connect
          </button>
        )}
        <button
          className="btn btn--ghost btn--sm"
          onClick={() => onEdit(conn.id)}
        >
          Edit
        </button>
        <button
          className="btn btn--danger btn--sm"
          onClick={handleDelete}
          disabled={deleting}
        >
          {deleting ? "…" : "Delete"}
        </button>
      </div>
    </div>
  );
}

// --- Main ConnectionsView ---

export function ConnectionsView() {
  const {
    connections,
    activeConn,
    isCredentialStoreSecure,
    refreshConnections,
    selectConnection,
    disconnect,
  } = useConnectionContext();

  const [modal, setModal] = useState<ModalMode | null>(null);
  const [connecting, setConnecting] = useState<string | null>(null);
  const [connectError, setConnectError] = useState<string | null>(null);

  const handleConnect = useCallback(
    async (id: string) => {
      setConnecting(id);
      setConnectError(null);
      try {
        await selectConnection(id);
      } catch (e: unknown) {
        setConnectError(e instanceof Error ? e.message : String(e));
      } finally {
        setConnecting(null);
      }
    },
    [selectConnection]
  );

  const handleDisconnect = useCallback(async () => {
    await disconnect();
  }, [disconnect]);

  const openCreate = useCallback(() => setModal({ kind: "create" }), []);
  const openEdit = useCallback(
    (id: string) => setModal({ kind: "edit", id }),
    []
  );
  const closeModal = useCallback(() => setModal(null), []);
  const handleSaved = useCallback(() => {
    setModal(null);
    refreshConnections();
  }, [refreshConnections]);

  const modalInitial = (): ConnFormData => {
    if (!modal) return emptyForm();
    if (modal.kind === "create") return emptyForm();
    const p = connections.find((c) => c.id === modal.id);
    return p ? profileToForm(p) : emptyForm();
  };

  return (
    <div className="connections-view">
      <div className="connections-header">
        <h1 className="connections-header__title">Connections</h1>
        <button className="btn btn--primary btn--sm" onClick={openCreate}>
          + New Connection
        </button>
      </div>

      {connectError && (
        <p className="form-error connections-error">{connectError}</p>
      )}

      {connections.length === 0 ? (
        <div className="empty-state">
          <p className="empty-state__msg">No connections saved.</p>
          <p className="empty-state__hint">
            Click &ldquo;New Connection&rdquo; to add your first Meilisearch
            server.
          </p>
        </div>
      ) : (
        <div className="conn-list">
          {connections.map((c) => (
            <ConnectionRow
              key={c.id}
              conn={c}
              isActive={activeConn?.id === c.id}
              onConnect={connecting ? () => {} : handleConnect}
              onEdit={openEdit}
              onDelete={() => refreshConnections()}
              onDisconnect={handleDisconnect}
            />
          ))}
        </div>
      )}

      {modal && (
        <ConnectionFormModal
          mode={modal}
          initial={modalInitial()}
          isSecureStore={isCredentialStoreSecure}
          onClose={closeModal}
          onSaved={handleSaved}
        />
      )}
    </div>
  );
}
