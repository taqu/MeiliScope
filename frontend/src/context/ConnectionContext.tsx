import {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
  type ReactNode,
} from "react";
import {
  ConnectionService,
  type ActiveConnectionInfo,
  type ConnectionSummary,
} from "../services/connection";

interface ConnectionContextValue {
  connections: ConnectionSummary[];
  activeConn: ActiveConnectionInfo | null;
  isCredentialStoreSecure: boolean;
  refreshConnections: () => void;
  selectConnection: (id: string) => Promise<ActiveConnectionInfo>;
  disconnect: () => Promise<void>;
}

const ConnectionContext = createContext<ConnectionContextValue | null>(null);

export function useConnectionContext(): ConnectionContextValue {
  const ctx = useContext(ConnectionContext);
  if (!ctx) throw new Error("useConnectionContext used outside ConnectionProvider");
  return ctx;
}

export function ConnectionProvider({ children }: { children: ReactNode }) {
  const [connections, setConnections] = useState<ConnectionSummary[]>([]);
  const [activeConn, setActiveConn] = useState<ActiveConnectionInfo | null>(null);
  const [isCredentialStoreSecure, setIsCredentialStoreSecure] = useState(true);

  const refresh = useCallback(async () => {
    try {
      const [conns, active, secure] = await Promise.all([
        ConnectionService.listConnections(),
        ConnectionService.getActiveConnection(),
        ConnectionService.isCredentialStoreSecure(),
      ]);
      setConnections(conns ?? []);
      setActiveConn(active ?? null);
      setIsCredentialStoreSecure(secure);
    } catch {
      // Non-fatal: keep existing state
    }
  }, []);

  // Load initial state on mount.
  useEffect(() => {
    void refresh();
  }, [refresh]);

  const selectConnection = useCallback(
    async (id: string): Promise<ActiveConnectionInfo> => {
      const info = await ConnectionService.selectConnection(id);
      setActiveConn(info);
      // Refresh list to update state badges.
      const conns = await ConnectionService.listConnections();
      setConnections(conns ?? []);
      return info;
    },
    []
  );

  const disconnect = useCallback(async () => {
    await ConnectionService.disconnect();
    setActiveConn(null);
    const conns = await ConnectionService.listConnections();
    setConnections(conns ?? []);
  }, []);

  return (
    <ConnectionContext.Provider
      value={{
        connections,
        activeConn,
        isCredentialStoreSecure,
        refreshConnections: refresh,
        selectConnection,
        disconnect,
      }}
    >
      {children}
    </ConnectionContext.Provider>
  );
}
