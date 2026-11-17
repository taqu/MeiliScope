# Connections

MeiliScope can maintain multiple saved connections to different Meilisearch servers.

---

## Opening the Connections Panel

Press **Ctrl+Shift+C** or click the **Connections** button in the top bar.

---

## Adding a Connection

1. Click **Add connection**.
2. Fill in:
   - **Name** — a display label for this connection (e.g. "Local Dev", "Production").
   - **URL** — the Meilisearch server URL including scheme and port (e.g. `http://localhost:7700` or `https://search.example.com`).
   - **API Key** — leave blank for open instances. Use a master key, admin key, or search-only key.
3. Click **Connect**. MeiliScope tests the URL and credentials immediately.

A green indicator confirms a successful connection. If the connection fails, an error message describes the reason.

---

## Switching Connections

Click a saved connection to make it the active connection. The Index Explorer reloads automatically.

---

## Editing a Connection

Select a saved connection and click **Edit**. URL and name changes take effect immediately on reconnect. API key changes are stored in the OS keychain.

---

## Deleting a Connection

Select a saved connection and click **Delete**. The stored API key is removed from the OS keychain at the same time.

---

## API Key Requirements

| Feature | Minimum permission |
|---|---|
| Health check | None |
| List indexes | Index read |
| Browse documents | Document read |
| Search | Search |
| Create / Update / Delete documents | Document write |
| Task tracking | Task read |

A **search-only key** is sufficient for browsing and searching. Document editing requires a key with document write and task read permissions.

If index enumeration returns HTTP 403, MeiliScope sets the connection to **limited access** mode and shows a text field to enter the index UID directly.

---

## Credential Storage

API keys are stored in the OS keychain and are never written to the configuration file or log files.

| Platform | Backend |
|---|---|
| Windows | Windows Credential Manager |
| Linux | Secret Service (GNOME Keyring / KWallet) |
| Fallback | Session-only (in-process memory) |

In session-only mode, credentials are lost on exit. The connection panel displays a warning when this mode is active.

---

## Connection Isolation

Requests always use the connection that was active at the time the request was initiated. Switching connections during an in-flight operation does not redirect that operation. Document mutation targets are locked at the time of confirmation.
