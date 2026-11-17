# MeiliScope Security Model

## 1. Overview

MeiliScope is a local desktop application. It connects to a user-configured Meilisearch server over HTTP or HTTPS and provides read-only browsing, search, and controlled document editing. No data is sent to external services.

## 2. Trust Boundaries

| Component | Trust Level | Notes |
|---|---|---|
| Local Go process | Trusted | Runs as current OS user |
| Wails WebView (Chromium/WebKit) | Semi-trusted | Runs frontend JS; API calls go through Wails bridge only |
| Meilisearch server | External | Treated as a data source; responses are never executed |
| Local config files | Trusted | Same user; no privilege escalation |
| OS credential store | Trusted | Windows Credential Manager / Linux Secret Service |
| Local export files | User-controlled | Written at user-specified paths |

## 3. Credential Storage

### 3.1 Storage Backends

| Platform | Backend | Fallback |
|---|---|---|
| Windows | Windows Credential Manager (`wincred`) | In-memory session store |
| Linux | Secret Service (D-Bus, e.g. GNOME Keyring, KWallet) | In-memory session store |
| Other | In-memory session store | — |

### 3.2 Properties

- API Keys are **never** written to `config.json`.
- API Keys are **never** logged at any level.
- The `ConnectionSummary` type returned to the frontend omits the API key field entirely.
- Session-only mode (`credentialMode: "session"`) stores keys in process memory; they are lost on exit.
- The `IsCredentialStoreSecure` field warns the UI when session mode is active.

### 3.3 Known Limitations

- If the OS keyring is unavailable (locked screen, missing daemon), the application falls back to session-only mode; credentials must be re-entered on restart.
- The local config directory is protected only by OS file permissions (the current user). MeiliScope does not add additional encryption.
- Exported files may contain sensitive document data and are written at user-chosen paths without additional encryption.

## 4. HTTP Security

### 4.1 Allowed Schemes

Only `http` and `https` are accepted. URL schemes other than these are rejected at client construction time.

### 4.2 TLS

TLS certificate verification is **always enabled**. The application uses Go's standard `net/http` transport which enforces certificate chain validation and hostname matching by default. TLS cannot be disabled without recompiling.

### 4.3 Credentials in URLs

URLs containing `userinfo` (`user:pass@host`) are rejected at validation time.

### 4.4 Cross-host Redirect Protection

**Fixed in Phase 6.** Go's default HTTP client forwards `Authorization` headers across all redirects, including cross-host redirects.

MeiliScope uses a custom `CheckRedirect` policy:
- Same-host redirects: `Authorization` header preserved.
- Cross-host redirects: `Authorization` header stripped before following.
- More than 10 redirects: request aborted.

This matches the [Fetch specification](https://fetch.spec.whatwg.org/#http-redirect-fetch) and browser behaviour.

Regression test: `internal/meilisearch/client_test.go` → `TestRedirect_CrossHost`.

### 4.5 Response Size

Responses are limited to 10 MiB by default (`defaultMaxBodyBytes`). Responses exceeding this limit are rejected with `ErrKindResponseTooLarge`.

### 4.6 Request Timeouts

Requests time out after 10 seconds by default. All requests carry a `context.Context` for cancellation.

### 4.7 Dedicated Transport

Each `meilisearch.Client` instance uses a **cloned** `http.Transport` (not the shared `http.DefaultTransport`). This avoids global transport mutation by other libraries.

## 5. Frontend Security

### 5.1 Content Rendering

React escapes all string values before inserting them into the DOM. MeiliScope does not use `dangerouslySetInnerHTML` anywhere. Document values, facet labels, and error messages are displayed as text, not HTML.

### 5.2 Wails Bridge

Frontend JavaScript can only call explicitly registered Go service methods. It cannot make arbitrary HTTP requests, access the filesystem, or interact with the OS outside the declared service API.

### 5.3 Data Isolation

Each component reads data only from the currently active connection. The generation counter in `ConnectionService` invalidates in-flight responses when the user switches connections.

## 6. Document Mutation Safety

- Read operations (browsing, searching) **never** send write requests to Meilisearch.
- Create, Update (partial), Replace, and Delete are distinct operations with different HTTP verbs.
- Every mutation requires explicit user confirmation before submission.
- The mutation target (index UID and document ID) is locked at the moment of confirmation. Switching connections during the confirmation step does not redirect the write.
- Ambiguous task outcomes (e.g., network failure after submission) are presented as uncertain, not as success.

## 7. Export Security

- Exports are initiated by explicit user action (file dialog).
- The exported scope is always visible in the Export dialog before execution.
- Index UIDs used in suggested filenames are sanitised (`[^a-zA-Z0-9]` → `_`).
- CSV cells beginning with `=`, `+`, `-`, or `@` are prefixed with a tab character to prevent formula injection in spreadsheet applications.
- Exports use atomic write (temp file → rename) to avoid partial files at the destination path.
- Temporary files are cleaned up on failure.

## 8. Query History and Saved Queries

- Query History and Saved Queries are stored locally in the application data directory.
- They are **not** uploaded to any external service.
- API Keys are **not** stored in query records.
- History entries include the query JSON and metadata only.

## 9. Logging

- Logs are written to `{AppDataDir}/MeiliScope/logs/` with file mode `0600`.
- Log directory is created with mode `0700`.
- API Keys are **never** logged.
- Authorization headers are **never** logged.
- Error messages from the HTTP client are sanitised to exclude credential material.

## 10. Dependency Vulnerability Policy

Dependencies are reviewed at each phase. Known vulnerabilities are triaged based on actual use and exploitability. The Go standard library's `net/http` and `crypto/tls` are used for all network operations; no third-party HTTP library is introduced.

Go module checksums are enforced by `go.sum`. npm package integrity is enforced by `package-lock.json`.

## 11. What MeiliScope Does Not Protect Against

- A fully compromised OS user session (all local files are accessible to the current user).
- Meilisearch server-side data corruption or unauthorised access via the server's own API.
- Exfiltration of exported files after they are written to disk.
- Man-in-the-middle attacks against HTTP (non-TLS) connections.
- Weak or reused API Keys.
- Race conditions in Meilisearch's own task system (not controlled by this application).
