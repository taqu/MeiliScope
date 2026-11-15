import { useState, useEffect, useCallback } from "react";
import { NAV_ITEMS, type ViewId } from "./types/navigation";
import { AppService, type ApplicationInfo, type RuntimeInfo } from "./services/app";
import { ConnectionsView } from "./features/connections/ConnectionsView";
import { IndexesView } from "./features/indexes/IndexesView";
import { DocumentsView } from "./features/documents/DocumentsView";
import { QueryView } from "./features/query/QueryView";

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

function Toolbar({ appInfo }: { appInfo: LoadState<ApplicationInfo> }) {
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
    </header>
  );
}

function StatusBar({ runtimeInfo }: { runtimeInfo: LoadState<RuntimeInfo> }) {
  return (
    <footer className="status-bar" role="contentinfo">
      {runtimeInfo.status === "loading" && (
        <span className="status-bar__item">Loading runtime info…</span>
      )}
      {runtimeInfo.status === "ok" && (
        <>
          <span className="status-bar__item">{runtimeInfo.data.goVersion}</span>
          <span className="status-bar__sep" aria-hidden="true">
            ·
          </span>
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

export default function App() {
  const [activeView, setActiveView] = useState<ViewId>("connections");
  const appInfo = useBackendData<ApplicationInfo>(AppService.getApplicationInfo);
  const runtimeInfo = useBackendData<RuntimeInfo>(
    AppService.getRuntimeInformation
  );

  const handleNavSelect = useCallback((id: ViewId) => {
    setActiveView(id);
  }, []);

  return (
    <div className="app-shell">
      <Toolbar appInfo={appInfo} />
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
