# MeiliScope — Phase 1 Implementation Instructions

## Connection Manager & Index Explorer

### 1. Project Context

MeiliScope is an open-source, standalone desktop application for exploring Meilisearch data.

**Mission:** Explore your Meilisearch data.

The application is implemented using Go, Wails v3, React, and TypeScript.

Phase 0 has been completed, and the application has been confirmed to run on both Windows and Linux.

The project proposal is maintained in the root-level `plan.md`.

Read the project proposal, current architecture documentation, Phase 0 reports, and existing source code before beginning implementation.

**Do not recreate the application foundation established in Phase 0.**

Reuse its application lifecycle, frontend/backend communication, configuration management, error-handling conventions, and build infrastructure.

### 2. Phase 1 Objectives

Implement the first functional version of MeiliScope that connects to Meilisearch and explores its indexes.

The required capabilities are:

1. Reusable Meilisearch HTTP API client.
2. Secure, cross-platform API Key storage.
3. Multiple connection profiles.
4. Connection testing and switching.
5. Index enumeration.
6. Index metadata and statistics.
7. Index Explorer user interface.
8. Reliable error and connection-state handling.

Phase 1 must work on both Windows and Linux.

Windows remains the primary distribution target, but Linux support must not be broken by Windows-specific implementations.

### 3. Architectural Requirements

Maintain the following architecture:

```text
React / TypeScript Frontend
          |
          | Wails v3 bindings
          v
Go Application Services
          |
          +-- Connection Manager
          |
          +-- Credential Store
          |
          +-- Index Service
          |
          v
Meilisearch HTTP Client
          |
          v
Meilisearch REST API
```

Responsibilities:

**Frontend**
- Presentation and user interaction.
- Connection selection.
- Index navigation.
- Loading, error, and empty states.

**Application Services**
- Connection lifecycle.
- Connection profile management.
- Index operations.
- Coordination between components.

**Meilisearch HTTP Client**
- HTTP requests.
- Authentication.
- Timeouts and cancellation.
- Response decoding.
- API error handling.

**Credential Store**
- Secure storage and retrieval of API Keys.
- OS-specific credential management.

The frontend must never communicate directly with Meilisearch.

Raw API Keys must not be returned through Wails bindings.

Do not introduce a local HTTP server solely for frontend/backend communication.

---

# Step 1.1 — Meilisearch HTTP Client

## Objective

Implement a reusable Go HTTP client for Meilisearch, independent of Wails and the frontend.

## Tasks

### 1.1.1 Inspect Existing Code

Check whether Phase 0 already introduced:

- HTTP utilities.
- Application error types.
- Logging conventions.
- Context cancellation patterns.
- Dependency injection.
- Service lifecycle management.

Reuse existing mechanisms when appropriate.

Avoid duplicate infrastructure.

### 1.1.2 Implement the HTTP Client

Implement a client using Go's standard `net/http` package.

Support:

- Configurable base URL.
- Optional API Key.
- Bearer authentication.
- HTTP GET and POST requests.
- JSON request serialization.
- JSON response decoding.
- Configurable timeout.
- Context cancellation.
- Structured errors.
- Safe response body handling.
- Maximum response size where appropriate.
- HTTP transport reuse.

Design the request layer so that additional HTTP methods can be supported in later phases.

However, implement only the operations needed for Phase 1.

Do not introduce an unnecessary generic API framework.

### 1.1.3 Handle Base URLs Safely

Support URLs such as:

```text
http://127.0.0.1:7700
http://localhost:7700
http://192.168.1.10:7700
https://search.example.com
https://search.example.com/meilisearch
```

Preserve valid deployment path prefixes.

Avoid incorrect path joining, duplicated slashes, or unexpected removal of URL prefixes.

Require HTTP or HTTPS schemes.

Reject malformed URLs and embedded userinfo credentials.

Do not silently downgrade HTTPS to HTTP.

Do not disable TLS certificate verification.

Prevent Meilisearch-controlled API paths from being interpreted as arbitrary external URLs.

### 1.1.4 Implement Structured Error Handling

Distinguish at least:

- Network unavailable.
- DNS failure.
- Connection refused.
- Timeout.
- Request canceled.
- TLS certificate failure.
- HTTP authentication failure.
- HTTP permission denied.
- Index not found.
- Invalid API request.
- Server error.
- Invalid JSON response.
- Unexpected response content.

Preserve useful Meilisearch error information when available:

- HTTP status.
- Error code.
- Error type.
- Error message.
- Documentation link, when provided.

Never expose the Authorization header, API Key, or other credentials through error messages.

### 1.1.5 Implement Safe Retry Behavior

Do not automatically retry every failed request.

For Phase 1:

- Do not retry authentication or authorization errors.
- Do not retry malformed requests.
- Do not retry canceled requests.
- Treat retryable connection and transient server failures conservatively.
- Do not retry indefinitely.
- Respect caller cancellation.

A simple implementation with no automatic retries is acceptable if errors are classified correctly.

### 1.1.6 Implement HTTP Client Tests

Use `httptest.Server` to verify:

- Correct request paths.
- Correct Bearer authentication.
- Requests without an API Key.
- Custom base paths.
- HTTP 200 responses.
- HTTP 400, 401, 403, 404, 429, and 500 responses.
- Malformed JSON.
- Timeouts.
- Context cancellation.
- Unexpected content types.
- Oversized responses.
- Sanitization of error messages.

Use synthetic credentials in tests.

## Acceptance Criteria

- The HTTP client works without Wails.
- Unit tests run without a live Meilisearch server.
- Cancellation and timeouts behave correctly.
- Errors are structured and useful.
- Credentials never appear in logs or returned errors.
- Existing application architecture remains intact.

---

# Step 1.2 — Cross-platform Credential Storage

## Objective

Implement secure API Key storage for Windows and Linux.

Credentials must be separated from ordinary application configuration.

## Tasks

### 1.2.1 Define a Credential Store Interface

Provide operations equivalent to:

```go
type CredentialStore interface {
    Save(ctx context.Context, id string, secret string) error
    Load(ctx context.Context, id string) (string, error)
    Delete(ctx context.Context, id string) error
}
```

Adapt the exact interface to existing project conventions.

Use a stable internal connection identifier rather than a user-visible connection name as the credential key.

Do not use API Keys themselves as credential identifiers.

### 1.2.2 Windows Credential Storage

Use a supported Windows OS credential mechanism, preferably Windows Credential Manager.

Requirements:

- Save API Keys securely.
- Retrieve API Keys for active connections.
- Delete stored credentials when requested.
- Handle missing or unavailable credentials.
- Avoid requiring administrator privileges.
- Never store credentials in ordinary configuration files.

Keep Windows-specific code in an isolated platform implementation.

### 1.2.3 Linux Credential Storage

Use the desktop Secret Service API through a maintained and suitable library or integration.

Support common desktop keyring implementations where possible.

Examples include GNOME Keyring and KWallet through their Secret Service compatibility interfaces.

Do not assume that every Linux environment has a running Secret Service provider.

Handle:

- Missing D-Bus session.
- Unavailable Secret Service.
- Locked keyring.
- User-denied access.
- Credential not found.
- Storage failure.

Do not automatically create an insecure credential file as a fallback.

### 1.2.4 Session-only Credential Mode

Provide a session-only alternative when secure persistent storage is unavailable.

Session-only behavior:

- Credentials remain in application process memory.
- Credentials are not written to persistent configuration.
- Credentials are not serialized through Wails responses.
- Credentials are discarded when the application exits.
- The user is informed that credentials will not persist.
- Reconnection may require re-entering the API Key.

Do not persist an API Key automatically when a secure backend is unavailable.

Avoid placing plaintext API Keys in immutable strings longer than necessary where practical, while acknowledging that complete memory erasure cannot be guaranteed in Go.

### 1.2.5 Credential Lifecycle

Support:

- Adding a credential.
- Replacing a credential.
- Removing a credential.
- Switching between connections.
- Recovering from a missing credential.
- Handling credential storage failures.

Ensure that connection profile deletion and credential deletion have clearly defined behavior.

If credential cleanup fails, report the failure rather than silently claiming complete deletion.

### 1.2.6 Security Testing

Test:

- Successful credential storage and retrieval.
- Credential replacement.
- Missing credentials.
- Storage failures.
- Session-only mode.
- Credential deletion.
- Connection profile changes.
- Absence of API Keys from logs and configuration.

Use mock credential stores for unit tests.

Perform platform-specific integration tests where a supported credential service is available.

## Acceptance Criteria

- Windows uses secure OS credential storage.
- Linux uses an appropriate secure credential service when available.
- Session-only operation is supported.
- Plaintext configuration storage is never used as a fallback.
- Credentials are not returned through frontend bindings.
- Both operating systems remain functional.

---

# Step 1.3 — Connection Manager

## Objective

Implement connection profile management and connection lifecycle handling.

## Tasks

### 1.3.1 Define the Connection Profile

A connection profile should contain:

```text
Connection ID
Display Name
Server URL
Credential Reference
Credential Persistence Mode
Optional Last Selected State
```

Do not place the API Key itself inside the persistent connection profile.

Use unique, stable IDs.

Support multiple profiles pointing to the same Meilisearch server.

### 1.3.2 Implement Connection CRUD

Support:

- Create connection.
- Edit connection.
- Delete connection.
- List saved connections.
- Select connection.
- Switch connection.
- Disconnect.
- Reconnect.

Integrate with the configuration service established in Phase 0.

Persist profile changes safely.

### 1.3.3 Implement Connection Testing

Test actual Meilisearch connectivity.

A successful `/health` response alone must not be treated as proof that the supplied API Key can access indexes.

Distinguish:

- Server reachable.
- Authentication accepted for a protected operation.
- Index enumeration authorized.
- Statistics access authorized.
- Connection partially usable.
- Connection unavailable.

Prefer a small set of read-only requests.

Do not create, update, or delete any Meilisearch data during connection testing.

Avoid requiring administrative permissions.

### 1.3.4 Implement Connection State

Define explicit connection states such as:

```text
Disconnected
Connecting
Connected
Limited Access
Authentication Required
Connection Error
```

Keep transport connectivity separate from individual API permissions.

A request returning HTTP 403 must not automatically imply that the server is offline.

Do not attempt to infer all API Key permissions from a single endpoint.

### 1.3.5 Implement Connection Manager UI

Provide a connection dialog or panel with:

- Display name.
- Server URL.
- API Key input.
- Show/hide credential control.
- Save credentials option.
- Session-only option.
- Test connection action.
- Save action.
- Cancel action.
- Delete action for existing profiles.

Display validation and connection errors clearly.

Do not show previously saved API Keys in plaintext.

For editing existing profiles, use an empty credential input to represent an unchanged stored credential.

Replacing a credential must be an explicit action.

### 1.3.6 Support Connection Switching

When switching connections:

- Cancel relevant in-flight requests.
- Clear previous connection-specific data immediately.
- Prevent stale responses from appearing under the new connection.
- Load the selected connection's credentials.
- Refresh the connection state.
- Notify dependent views.

Use request generations, connection IDs, or an equivalent mechanism to prevent stale asynchronous responses.

Do not leak data between connection profiles.

### 1.3.7 Handle Partial Persistence Failures

Credential and profile storage are separate operations and may fail independently.

Implement a deliberate transaction-like workflow where possible:

- Avoid saving broken credential references.
- Avoid deleting existing credentials before replacements are safely stored.
- Clean up newly created orphan credentials after failed profile creation.
- Report incomplete cleanup.
- Preserve an existing working connection when a profile update fails.

## Acceptance Criteria

- Multiple connections can be saved.
- Connections can be selected and switched.
- Connection profiles survive application restarts.
- Stored credentials are managed securely.
- Session-only credentials behave correctly.
- Connection testing distinguishes reachability from authorization.
- Stale responses cannot overwrite another connection's view.

---

# Step 1.4 — Index API & Statistics

## Objective

Implement read-only Index operations through the Meilisearch HTTP API.

## Tasks

### 1.4.1 Verify Supported Meilisearch API Versions

Consult the official Meilisearch API documentation.

Identify the API behavior relevant to:

- Index listing.
- Index details.
- Index statistics.
- Server information used for compatibility diagnostics.

Record the Meilisearch versions used for testing.

Do not assume every version exposes exactly the same fields.

Use optional fields where appropriate.

### 1.4.2 Implement Index Listing

Implement the documented equivalent of:

```http
GET /indexes?offset=0&limit=20
```

Support:

- Offset.
- Limit.
- Total count.
- Index results.
- Empty result sets.
- Cancellation.
- Authorization errors.

Do not download all indexes automatically when only one page is needed.

Preserve server-reported pagination information.

### 1.4.3 Implement Index Details

Implement:

```http
GET /indexes/{uid}
```

Expose relevant fields, including:

- Index UID.
- Primary Key.
- Creation timestamp.
- Update timestamp.

Handle missing and optional fields correctly.

Treat index UIDs as opaque identifiers.

Use appropriate URL path escaping.

### 1.4.4 Implement Index Statistics

Implement:

```http
GET /indexes/{uid}/stats
```

Support available statistics such as:

- Number of documents.
- Indexing state.
- Field distribution.
- Document storage size information, when available.
- Other relevant fields supported by the target API.

Do not assume all statistics are available in every supported Meilisearch version.

Avoid turning missing optional statistics into fatal parsing errors.

### 1.4.5 Handle Permission Differences

Index listing and index statistics may require permissions different from search access.

Handle permission failures independently.

Examples:

- Index listing succeeds but statistics fail.
- Index details are accessible but some fields are unavailable.
- A Search API Key cannot enumerate indexes.
- One API Key has access to only a subset of indexes.

Do not automatically request an Admin API Key.

Display accurate permission-related errors.

### 1.4.6 Design Typed Backend Models

Expose frontend-friendly, typed models.

Examples:

```text
IndexSummary
IndexDetails
IndexStatistics
PaginatedIndexResult
```

Prefer generated Wails TypeScript bindings.

Avoid exposing internal HTTP client objects or credential references through frontend models.

### 1.4.7 Implement Tests

Test:

- Index listing.
- Pagination.
- Empty indexes.
- Missing primary key.
- Null and optional fields.
- Field distribution.
- Large document counts.
- Index not found.
- Authentication and authorization errors.
- Malformed responses.
- Context cancellation.
- Different API response shapes where necessary.

## Acceptance Criteria

- Index operations are independently testable.
- Paginated index listing works.
- Index metadata is correctly decoded.
- Statistics are optional and safely handled.
- Permission errors do not invalidate otherwise usable connections.
- No write operations are implemented.

---

# Step 1.5 — Index Explorer UI

## Objective

Implement the first functional data exploration interface.

## Tasks

### 1.5.1 Index Explorer Layout

Use the application shell created during Phase 0.

Recommended structure:

```text
+--------------------------------------------------+
| MeiliScope   [Connection Selector]   [Refresh]   |
+-------------------+------------------------------+
|                   |                              |
| Index Explorer    | Index Details                |
|                   |                              |
|  documents        | Name: documents              |
|  users            | Primary Key: id              |
|  products         | Documents: 125,000           |
|                   | Updated: ...                 |
|                   |                              |
|                   | Field Distribution           |
|                   | ...                          |
+-------------------+------------------------------+
| Connected | Ready                                |
+--------------------------------------------------+
```

Adapt the layout to the actual Phase 0 application shell.

Do not replace the existing UI architecture unnecessarily.

### 1.5.2 Display Index List

Provide:

- Index UID.
- Selection state.
- Refresh action.
- Loading indicator.
- Empty state.
- Error state.
- Pagination.

Client-side filtering of the currently loaded page is acceptable, but it must not be misrepresented as searching all server indexes.

Avoid unbounded retrieval merely to implement filtering.

### 1.5.3 Display Index Details

When an index is selected, display:

- UID.
- Primary Key.
- Created At.
- Updated At.
- Document Count.
- Indexing status.
- Available statistics.

Handle optional or unavailable data gracefully.

### 1.5.4 Display Field Distribution

Show available field names and document counts.

If statistics are unavailable because of permissions, retain accessible index metadata and show a specific explanation.

Do not infer a complete document schema from field distribution.

### 1.5.5 Implement Connection Selector

Display saved connection profiles in the application toolbar.

Support:

- Current connection selection.
- Connection switching.
- Connection state indication.
- Opening Connection Manager.
- Refreshing the selected connection.

### 1.5.6 Prevent Stale UI State

When switching connections or indexes:

- Clear stale details.
- Cancel previous requests.
- Ignore late responses belonging to previous selections.
- Preserve consistent loading indicators.
- Avoid displaying misleading statistics.

### 1.5.7 Handle Large Index Lists

Use server-side pagination.

Avoid loading the complete index catalog into memory by default.

Do not introduce complex virtualization unless testing demonstrates a need.

### 1.5.8 Prepare Navigation for Phase 2

Expose the selected connection and index as application state that the future Document Browser can consume.

Do not implement Document Browser functionality yet.

## Acceptance Criteria

- Users can select a connection.
- Users can browse available indexes.
- Users can view index metadata.
- Users can view available statistics.
- Empty and restricted indexes are handled.
- Refresh and pagination work.
- Switching connections does not show stale data.
- The interface works on Windows and Linux.

---

# Step 1.6 — Integration & Compatibility Tests

## Objective

Validate Phase 1 as an integrated cross-platform feature.

## Tasks

### 1.6.1 Integration Test Coverage

Test the complete sequence:

1. Launch MeiliScope.
2. Add a Meilisearch connection.
3. Enter an API Key.
4. Test the connection.
5. Save the profile.
6. Connect to the instance.
7. Retrieve the index list.
8. Select an index.
9. Retrieve index details.
10. Retrieve statistics.
11. Switch to another connection.
12. Return to the original connection.
13. Restart the application.
14. Verify persistence and reconnection.

### 1.6.2 Authentication Scenarios

Validate at least:

- Unauthenticated local development instance.
- Valid read-only key with sufficient index permissions.
- Search-only API Key.
- Invalid API Key.
- Expired or revoked API Key, where testable.
- API Key restricted to specific indexes.
- Credential missing from secure storage.

Do not assume that a Search API Key can enumerate indexes.

Do not bypass permission checks.

### 1.6.3 Failure Scenarios

Validate:

- Server stopped.
- Connection refused.
- DNS failure.
- Network timeout.
- TLS certificate failure.
- Invalid server URL.
- Malformed server response.
- Missing index.
- HTTP 401.
- HTTP 403.
- HTTP 429.
- HTTP 500.
- Connection switching during an active request.
- Credential storage unavailable.

The UI must remain responsive after failures.

### 1.6.4 Windows Testing

Verify:

- Application startup.
- Windows Credential Manager integration.
- Saving and loading connection profiles.
- Index enumeration.
- Index statistics.
- Connection switching.
- Restart behavior.
- Production build.

Use the existing Windows build and CI infrastructure.

### 1.6.5 Linux Testing

Verify:

- Application startup.
- Secret Service credential integration.
- Session-only fallback.
- Index enumeration.
- Index statistics.
- Connection switching.
- Restart behavior.
- Production build.

Test the supported Linux environment established in Phase 0.

Do not claim compatibility with untested Linux distributions.

### 1.6.6 Meilisearch Compatibility

Use an explicitly documented Meilisearch version or version range.

Run automated unit tests using mock HTTP servers.

Where practical, run integration tests against a real Meilisearch instance.

Integration tests must not modify user data.

Use dedicated test instances or isolated test data when necessary.

Separate tests requiring a running server from the default unit test suite.

### 1.6.7 Security Review

Verify:

- No API Keys in ordinary config files.
- No API Keys in logs.
- No API Keys in frontend responses.
- No plaintext credential fallback.
- No TLS verification bypass.
- No unintended write operations.
- Safe cleanup of deleted credentials.
- Connection isolation.
- Safe handling of error messages.

### 1.6.8 Update CI

Extend existing GitHub Actions workflows to cover the new backend and frontend tests.

Preserve Windows CI support.

Add or extend Linux CI validation using the build and test configuration established in Phase 0.

Use appropriately configured Linux runners for tests requiring a graphical session or Secret Service.

If a secure credential backend cannot be exercised in CI, run mock-based tests and record the remaining platform integration requirement.

### 1.6.9 Update Documentation

Update:

- `README.md`
- `docs/en/getting-started/`
- `docs/ja/getting-started/`
- `docs/en/user-guide/`
- `docs/ja/user-guide/`
- `docs/architecture/`
- `plan.md`, if actual technical decisions or supported-platform commitments have changed.

Document:

- Creating connections.
- API Key requirements.
- Credential storage behavior.
- Linux Secret Service prerequisites.
- Session-only credentials.
- Index browsing.
- Permission limitations.
- Troubleshooting connection failures.

Keep documentation consistent with actual functionality.

### 1.6.10 Prepare the Phase Completion Report

Create:

`docs/development/phase-1-report.md`

Include:

- Implementation summary.
- Files created or modified.
- Architecture decisions.
- HTTP client test results.
- Credential storage validation.
- Windows test results.
- Linux test results.
- Meilisearch compatibility findings.
- Known limitations.
- Unresolved issues.
- Readiness for Phase 2.

## Acceptance Criteria

- All six steps are complete.
- Connection Manager works on Windows and Linux.
- Secure persistent credentials work where an OS credential backend is available.
- Session-only mode works when secure persistence is unavailable.
- Index Explorer works against a real Meilisearch instance.
- Connection failures do not crash or freeze the application.
- Authentication and permission errors are correctly distinguished.
- Automated tests pass.
- CI builds continue to work.
- Documentation reflects the actual implementation.

---

# 4. General Implementation Constraints

The following rules apply to every step.

## 4.1 Avoid Scope Creep

Do not implement:

- Document Browser.
- Document editing.
- Query Console.
- Search Builder.
- Index creation or deletion.
- Index settings modification.
- API Key creation or administration.
- Server configuration management.
- Task administration.
- Automatic updates.
- Additional background services.

Phase 1 is read-only with respect to Meilisearch data.

## 4.2 Preserve Cross-platform Compatibility

Do not introduce Windows-only dependencies into shared Go packages.

Use platform-specific implementation files or interfaces where necessary.

Respect build tags.

Avoid assuming that all Linux desktop environments provide the same secret storage service.

## 4.3 Keep Dependencies Minimal

Prefer the Go standard library for HTTP and JSON handling.

Use external libraries where they provide clear value, particularly for secure OS credential integration.

Document important dependency decisions.

## 4.4 Preserve Application Responsiveness

Network requests must be asynchronous from the frontend's perspective.

Use cancellation and request timeouts.

Never block UI rendering while waiting for Meilisearch.

## 4.5 Preserve Security Boundaries

Do not expose credentials through frontend models.

Use opaque connection identifiers.

Do not weaken certificate validation.

Avoid unnecessary privileged operations.

Do not use administrator or master credentials as a requirement for basic browsing.

## 4.6 Maintain Testability

The Meilisearch client, connection services, and credential-management logic must be independently testable without launching Wails.

Mock external dependencies at clearly defined boundaries.

Avoid tightly coupling business logic to UI components.

---

# 5. Required Reporting Format

After each step, provide:

### Completed Work
Describe the implemented functionality.

### Files Changed
List significant source and documentation changes.

### Tests
List executed tests and their results.

### Architecture Decisions
Describe relevant implementation choices.

### Remaining Issues
List unresolved problems, limitations, and deferred work.

### Next Step Readiness
State whether the next step can proceed.

Do not claim successful validation without executing the relevant tests.

If an integration test requires an unavailable environment, report it as not executed.

---

# 6. Phase 1 Definition of Done

Phase 1 is complete when a user can:

1. Launch MeiliScope on Windows or Linux.
2. Add a Meilisearch server connection.
3. Supply an API Key when required.
4. Securely store credentials or use session-only mode.
5. Test the connection.
6. Select a saved connection.
7. Browse accessible indexes.
8. View index metadata and statistics.
9. Switch between multiple connections.
10. Restart MeiliScope and recover saved profiles.
11. Understand connection and authorization failures.

All operations must be available through the existing Wails desktop application.

The implementation must not require a separate local MeiliScope server.

**Phase 1 must establish a reliable read-only foundation for Phase 2 — Document Browser.**
