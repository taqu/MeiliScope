import { useState, useEffect, useCallback } from "react";
import { NAV_ITEMS, type ViewId } from "./types/navigation";
import { AppService, type ApplicationInfo, type RuntimeInfo } from "./services/app";
import { ConnectionProvider, useConnectionContext } from "./context/ConnectionContext";
import { ConnectionsView } from "./features/connections/ConnectionsView";
import { IndexesView } from "./features/indexes/IndexesView";
import { DocumentsView } from "./features/documents/DocumentsView";
import { QueryView } from "./features/query/QueryView";
import { useKeyboardShortcuts } from "./hooks/useKeyboardShortcuts";

type LoadState<T> =
  | { status: "idle" }
  | { status: "loading" }
  | { status: "ok"; data: T }
  | { status: "error"; message: string };

function useBackendData<T>(fetch: () => Promise<T>): LoadState<T> {
  const [state, setState] = useState<LoadState<T>>({ status: "loading" });
  useEffect(() => {
    let cancelled = false;
    setState({ status: "loading" });
    fetch()
      .then((data) => {
        if (!cancelled) setState({ status: "ok", data });
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          const message = err instanceof Error ? err.message : String(err);
          setState({ status: "error", message });
        }
      });
    return () => {
      cancelled = true;
    };
  }, []); // eslint-disable-line react-hooks/exhaustive-deps
  return state;
}

function NavSidebar({
  active,
  onSelect,
}: {
  active: ViewId;
  onSelect: (id: ViewId) => void;
}) {
  return (
    <nav className="nav-sidebar" aria-label="Main navigation">
      {NAV_ITEMS.map((item) => (
        <button
          key={item.id}
          className={`nav-item${active === item.id ? " nav-item--active" : ""}`}
          onClick={() => onSelect(item.id)}
          title={item.title}
          aria-label={item.title}
          aria-current={active === item.id ? "page" : undefined}
        >
          <span className="nav-item__abbr">{item.label}</span>
          <span className="nav-item__title">{item.title}</span>
        </button>
      ))}
    </nav>
  );
}

// Connection state dot indicator
function ConnStateDot({ state }: { state: string }) {
  return (
    <span
      className={`conn-dot conn-dot--${state}`}
      aria-label={`Connection: ${state}`}
    />
  );
}

function Toolbar({
  appInfo,
  onOpenConnections,
}: {
  appInfo: LoadState<ApplicationInfo>;
  onOpenConnections: () => void;
}) {
  const { activeConn } = useConnectionContext();

  return (
    <header className="toolbar" role="banner">
      <span className="toolbar__brand">MeiliScope</span>
      {appInfo.status === "ok" && (
        <span className="toolbar__version">v{appInfo.data.version}</span>
      )}
      {appInfo.status === "loading" && (
        <span className="toolbar__version toolbar__version--loading">…</span>
      )}
      {appInfo.status === "error" && (
        <span
          className="toolbar__version toolbar__version--error"
          title={appInfo.message}
        >
          error
        </span>
      )}

      <div className="toolbar__spacer" />

      <button
        className="toolbar__conn-btn"
        onClick={onOpenConnections}
        aria-label="Manage connections"
      >
        {activeConn ? (
          <>
            <ConnStateDot state={activeConn.state} />
            <span className="toolbar__conn-name">{activeConn.name}</span>
          </>
        ) : (
          <span className="toolbar__conn-name toolbar__conn-name--none">
            No connection
          </span>
        )}
      </button>
    </header>
  );
}

function StatusBar({ runtimeInfo }: { runtimeInfo: LoadState<RuntimeInfo> }) {
  const { activeConn } = useConnectionContext();

  const stateLabel: Record<string, string> = {
    connected: "Connected",
    connecting: "Connecting…",
    limited_access: "Limited Access",
    auth_required: "Auth Required",
    error: "Connection Error",
    disconnected: "Disconnected",
  };

  return (
    <footer className="status-bar" role="contentinfo">
      {activeConn && (
        <>
          <span className={`status-bar__item status-bar__item--conn-${activeConn.state}`}>
            {stateLabel[activeConn.state] ?? activeConn.state}
          </span>
          <span className="status-bar__sep" aria-hidden="true">·</span>
        </>
      )}
      {runtimeInfo.status === "loading" && (
        <span className="status-bar__item">Loading…</span>
      )}
      {runtimeInfo.status === "ok" && (
        <>
          <span className="status-bar__item">{runtimeInfo.data.goVersion}</span>
          <span className="status-bar__sep" aria-hidden="true">·</span>
          <span className="status-bar__item">
            {runtimeInfo.data.os}/{runtimeInfo.data.arch}
          </span>
        </>
      )}
      {runtimeInfo.status === "error" && (
        <span
          className="status-bar__item status-bar__item--error"
          title={runtimeInfo.message}
          role="alert"
        >
          Backend unreachable
        </span>
      )}
    </footer>
  );
}

function ContentArea({ view }: { view: ViewId }) {
  switch (view) {
    case "connections":
      return <ConnectionsView />;
    case "indexes":
      return <IndexesView />;
    case "documents":
      return <DocumentsView />;
    case "query":
      return <QueryView />;
  }
}

function AppInner() {
  const [activeView, setActiveView] = useState<ViewId>("connections");
  const appInfo = useBackendData<ApplicationInfo>(AppService.getApplicationInfo);
  const runtimeInfo = useBackendData<RuntimeInfo>(AppService.getRuntimeInformation);

  const handleNavSelect = useCallback((id: ViewId) => {
    setActiveView(id);
  }, []);

  const openConnections = useCallback(() => setActiveView("connections"), []);

  useKeyboardShortcuts([
    { key: "1", ctrl: true, description: "Open Index Explorer", handler: () => setActiveView("indexes") },
    { key: "2", ctrl: true, description: "Open Document Browser", handler: () => setActiveView("documents") },
    { key: "3", ctrl: true, description: "Open Query Console", handler: () => setActiveView("query") },
    { key: "C", ctrl: true, shift: true, description: "Open Connections", handler: () => setActiveView("connections") },
    { key: "c", ctrl: true, shift: true, description: "Open Connections", handler: () => setActiveView("connections") },
  ]);

  return (
    <div className="app-shell">
      <Toolbar appInfo={appInfo} onOpenConnections={openConnections} />
      <div className="app-body">
        <NavSidebar active={activeView} onSelect={handleNavSelect} />
        <main className="content-area" role="main">
          <ContentArea view={activeView} />
        </main>
      </div>
      <StatusBar runtimeInfo={runtimeInfo} />
    </div>
  );
}

export default function App() {
  return (
    <ConnectionProvider>
      <AppInner />
    </ConnectionProvider>
  );
}
