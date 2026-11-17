import * as ConnectionSvc from "../../bindings/github.com/taqu/meiliscope/internal/services/connectionservice.js";

// Re-export generated types so callers don't import from deep binding paths.
export type {
  ActiveConnectionInfo,
  ConnectionSummary,
  ConnectionTestResult,
  SaveConnectionRequest,
} from "../../bindings/github.com/taqu/meiliscope/internal/services/models.js";

export type ConnectionState =
  | "disconnected"
  | "connecting"
  | "connected"
  | "limited_access"
  | "auth_required"
  | "error";

export const ConnectionService = {
  listConnections: () => ConnectionSvc.ListConnections(),
  getActiveConnection: () => ConnectionSvc.GetActiveConnection(),
  isCredentialStoreSecure: () => ConnectionSvc.IsCredentialStoreSecure(),
  createConnection: (req: {
    name: string;
    url: string;
    apiKey: string;
    credentialMode: string;
  }) => ConnectionSvc.CreateConnection(req),
  updateConnection: (
    id: string,
    req: { name: string; url: string; apiKey: string; credentialMode: string }
  ) => ConnectionSvc.UpdateConnection(id, req),
  deleteConnection: (id: string) => ConnectionSvc.DeleteConnection(id),
  testConnection: (id: string, apiKey: string) =>
    ConnectionSvc.TestConnection(id, apiKey),
  testNewConnection: (testURL: string, apiKey: string) =>
    ConnectionSvc.TestNewConnection(testURL, apiKey),
  selectConnection: (id: string) => ConnectionSvc.SelectConnection(id),
  disconnect: () => ConnectionSvc.Disconnect(),
};
