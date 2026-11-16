# MeiliScope — Phase 3 Implementation Instructions

## Query Console & MVP Completion

### 1. Project Context

MeiliScope is an open-source desktop application for exploring Meilisearch indexes, browsing documents, and executing search queries.

**Mission: Explore your Meilisearch data.**

The established technology stack includes:

- Go for application backend services.
- Wails v3 for desktop integration.
- React and TypeScript for the frontend.
- Meilisearch HTTP API for backend communication.
- Windows and Linux as validated execution environments.
- MIT License.

The project proposal is maintained in `plan.md`.

Phase 0 established the desktop application foundation, architecture, and build infrastructure.

Phase 1 implemented the Connection Manager, credential storage, Meilisearch HTTP client, and Index Explorer.

Phase 2 implemented the Document Browser, dynamic document tables, pagination, and JSON Inspector.

Phase 3 must build on these existing components.

**Do not recreate working services or UI components from previous phases.**

Inspect the repository, project proposal, architecture documentation, and existing implementation before writing code.

Preserve existing conventions unless a concrete technical issue justifies changing them.

---

# 2. Phase 3 Objectives

Implement a fully functional Query Console that enables users to execute and inspect Meilisearch search requests.

Users must be able to:

1. Select a Meilisearch connection and index.
2. Write an arbitrary Search API JSON request.
3. Validate JSON syntax before execution.
4. Execute the request.
5. Cancel an in-progress request.
6. Inspect search results in Table and JSON views.
7. Examine search metadata and facet information.
8. Inspect ranking scores and ranking score details when available.
9. Review server-reported and client-measured execution times.
10. Diagnose API errors.
11. Review and rerun previous queries.
12. Switch connections without leaking stale results.

Phase 3 completes the initial MVP feature set.

It does not introduce document editing or administrative functionality.

## 2.1 Design Principles

**JSON First**

The Query Console must accept search parameters as arbitrary JSON objects rather than restricting requests to a fixed form schema.

**API Transparency**

Preserve Meilisearch request and response structures as faithfully as practical.

**Read-only Execution**

Only execute the documented Search API operation.

**Reusable Presentation**

Reuse the Document Table and JSON Inspector from Phase 2.

**Accurate Diagnostics**

Distinguish server execution time from client-observed round-trip time.

**Safe Persistence**

Query history must never contain API Keys or authentication headers.

**Cross-platform Compatibility**

Maintain functionality on Windows and Linux.

---

# 3. Architecture

Maintain the existing application architecture.

```text
React / TypeScript
│
├── Query Console
│   ├── JSON Query Editor
│   ├── Execute / Cancel
│   ├── Query History
│   └── Query Results
│       ├── Results Table
│       ├── JSON Inspector
│       ├── Search Metadata
│       ├── Facet Viewer
│       └── Ranking Diagnostics
│
▼
Wails v3 Bindings
│
▼
Go Application Services
│
├── Connection Manager
├── Query Execution Service
├── Search Response Models
└── Existing Meilisearch HTTP Client
    │
    ▼
Meilisearch Search API
```

Responsibilities:

**Frontend**

- Editing JSON.
- Validating syntax locally.
- Managing query UI state.
- Displaying results and diagnostics.
- Maintaining query history.
- Presenting loading, cancellation, and error states.

**Go Backend**

- Resolving connections and credentials.
- Executing Search API requests.
- Enforcing the supported API endpoint.
- Managing request cancellation.
- Measuring client-side request duration.
- Classifying API and transport errors.
- Returning structured and raw response data.

**Existing Components**

- Connection Manager provides the active connection.
- Index Explorer provides the active index.
- Document Table displays result hits.
- JSON Inspector displays selected hits and raw JSON.

Do not duplicate HTTP authentication or connection handling in the Query Console.

---

# Step 3.1 — Search API & Query Execution

## Objective

Implement a reusable Go service for executing arbitrary Meilisearch Search API JSON requests.

## Tasks

### 3.1.1 Verify the Search API

Consult the current official Meilisearch documentation.

Verify:

- Search endpoint.
- POST request format.
- Authentication requirements.
- Supported request body structure.
- Response structure.
- Pagination modes.
- Ranking metadata.
- Facet metadata.
- Error responses.
- Relevant version differences.

Primary endpoint:

```http
POST /indexes/{index_uid}/search
Content-Type: application/json
Authorization: Bearer <API_KEY>
```

Do not implement the GET search endpoint as the primary mechanism.

POST supports structured JSON request bodies and is the preferred interface for the Query Console.

Use the Meilisearch versions actually supported by the project.

Record any compatibility limitations.

### 3.1.2 Implement Query Execution Service

Create a backend service responsible for search execution.

Suggested interface:

```go
type QueryExecutionService interface {
    Execute(
        ctx context.Context,
        connectionID string,
        indexUID string,
        requestJSON json.RawMessage,
    ) (*SearchExecutionResult, error)
}
```

Adapt the interface to existing project conventions.

The service must:

- Resolve the connection through Connection Manager.
- Retrieve credentials securely.
- Reuse the existing HTTP client.
- Send the request to the selected index.
- Apply existing timeout and cancellation behavior.
- Return the response without silently dropping fields.
- Report API and transport errors consistently.

Do not accept arbitrary HTTP URLs from the frontend.

Do not allow the Query Console to invoke other endpoints.

### 3.1.3 Accept Arbitrary Search JSON

The service must accept a JSON object containing documented Meilisearch search parameters.

Example:

```json
{
  "q": "rendering",
  "filter": "category = 'graphics'",
  "sort": ["created_at:desc"],
  "limit": 20,
  "showRankingScore": true,
  "showRankingScoreDetails": true
}
```

Requirements:

- Reject invalid JSON.
- Reject non-object top-level JSON values.
- Preserve unknown search parameters.
- Do not remove fields simply because the UI does not recognize them.
- Do not silently rewrite parameter values.
- Do not convert numeric values through lossy floating-point representations.
- Leave semantic validation to Meilisearch where appropriate.

JSON syntax validation is the application's responsibility.

Validation of search-specific parameter semantics is primarily the server's responsibility.

Do not implement a duplicate Meilisearch query validation engine.

### 3.1.4 Preserve Search Responses

Meilisearch search responses may contain:

```text
hits
query
processingTimeMs
limit
offset
estimatedTotalHits
page
hitsPerPage
totalPages
totalHits
facetDistribution
facetStats
semanticHitCount
```

Depending on the server version and requested features, additional fields may be present.

The response model must support known metadata while preserving unknown fields.

A recommended structure is:

```text
SearchExecutionResult
├── Request Information
├── Raw Response JSON
├── Parsed Search Metadata
├── Search Hits
├── Client Duration
└── Diagnostics
```

Do not assume every response contains all listed fields.

Support missing and optional values.

### 3.1.5 Preserve JSON Fidelity

Search hits are dynamic JSON documents.

Reuse the JSON preservation strategy from Phase 2.

Avoid converting arbitrary JSON numbers to floating-point values when doing so would lose precision.

Preserve:

- Nested objects.
- Arrays.
- Strings.
- Booleans.
- Null values.
- Unicode.
- Large numeric values.
- Unknown response properties.

Raw response viewing must not depend on the typed metadata model recognizing every property.

### 3.1.6 Implement Cancellation

Support cancellation of active search requests.

Requirements:

- Starting a new query must not allow an older result to overwrite it.
- Switching connections must cancel or invalidate active requests.
- Switching indexes must cancel or invalidate active requests.
- Closing the application must not leave operations running indefinitely.
- Cancellation must be distinguishable from a server error.

Use existing Wails and Go cancellation mechanisms where applicable.

If the frontend binding does not support direct context cancellation, implement a narrow, safe operation identifier mechanism.

Do not create a new background HTTP server.

### 3.1.7 Protect Against Oversized Responses

Search requests may return large result sets.

Use the existing HTTP response size controls.

Do not silently truncate JSON responses.

If a response exceeds a configured safety limit, return a clearly classified error.

The UI should explain the limit and allow the user to adjust the query.

Use bounded memory where practical.

### 3.1.8 Implement Backend Tests

Test:

- Valid JSON search.
- Empty JSON object.
- Empty search query.
- Filter expressions.
- Sort arrays.
- Facet requests.
- Ranking score requests.
- Unknown request parameters.
- Nested JSON parameters.
- Invalid JSON.
- Non-object JSON.
- Large JSON numbers.
- Empty search results.
- HTTP 400.
- HTTP 401.
- HTTP 403.
- HTTP 404.
- HTTP 429.
- HTTP 500.
- Timeout.
- Cancellation.
- Malformed response JSON.
- Unexpected response fields.
- Excessive response size.

Use `httptest.Server`.

Do not require a live Meilisearch server for unit tests.

## Acceptance Criteria

- Arbitrary valid Search API JSON objects can be executed.
- Unknown request fields are preserved.
- Raw responses are available.
- Search metadata is parsed safely.
- Cancellation works.
- Error handling reuses Phase 1 infrastructure.
- Backend unit tests pass.
- No write endpoints are exposed.

---

# Step 3.2 — JSON Query Editor

## Objective

Implement an editor optimized for writing and executing Search API JSON requests.

## Tasks

### 3.2.1 Evaluate Monaco Editor

Monaco Editor is the preferred candidate identified in the project proposal.

Verify:

- Current maintained version.
- React compatibility.
- Wails WebView compatibility.
- Worker configuration.
- Bundling requirements.
- Production asset loading.
- Windows behavior.
- Linux behavior.
- Dependency and bundle size impact.

If Monaco is unsuitable, choose a maintained alternative and document the decision.

Do not introduce an additional editor framework without a concrete reason.

### 3.2.2 Implement Query Editor

Provide:

- JSON syntax highlighting.
- Line numbers.
- Automatic indentation.
- Bracket matching.
- Syntax error markers.
- JSON formatting.
- Undo and redo.
- Keyboard navigation.
- Resizable editor area.

The editor should be responsive with reasonably large queries.

Do not block the UI during normal editing.

### 3.2.3 Implement JSON Validation

Validate JSON syntax locally.

Requirements:

- Detect invalid JSON before execution.
- Display clear error locations when available.
- Reject non-object top-level JSON.
- Preserve the editor contents after validation errors.
- Allow empty search objects such as `{}`.
- Do not require a `q` property.

Avoid validating every Meilisearch parameter using a fixed schema.

Future Meilisearch versions may introduce additional parameters.

### 3.2.4 Implement Query Execution Controls

Provide:

- Execute.
- Cancel.
- Format JSON.
- Clear editor.
- Restore default query.

Use a default template such as:

```json
{
  "q": "",
  "limit": 20
}
```

Do not automatically execute queries merely because the editor content changes.

The user must explicitly request execution.

### 3.2.5 Keyboard Shortcuts

Implement:

- Ctrl+Enter: Execute query.
- Escape or a clearly documented alternative: Cancel active query.
- Standard editor undo and redo.
- Standard editor copy and paste.

Avoid overriding text-editor shortcuts unnecessarily.

Prevent duplicate execution from a single shortcut event.

### 3.2.6 Implement Query Editor State

Maintain:

```text
Selected Connection
Selected Index
Query Text
Validation State
Execution State
Last Result
Last Error
```

Query text must not be lost unexpectedly during ordinary navigation.

Define intentional behavior when the active index or connection changes.

A query may remain in the editor, but results associated with another connection or index must not remain displayed as current results.

### 3.2.7 Handle In-progress Operations

While a query is running:

- Show execution progress.
- Provide cancellation.
- Prevent accidental duplicate submission where appropriate.
- Allow editing the next query if practical.
- Preserve the exact submitted query separately from subsequent editor changes.

A response must be associated with the submitted query snapshot, not with whatever text currently appears in the editor.

### 3.2.8 Implement Frontend Tests

Test:

- Valid JSON.
- Invalid JSON.
- Non-object JSON.
- Empty query object.
- JSON formatting.
- Execute button.
- Keyboard execution.
- Cancellation.
- Editor state preservation.
- Connection switching.
- Index switching.
- Stale response prevention.

## Acceptance Criteria

- A functional JSON editor is available.
- JSON syntax errors are displayed before execution.
- Arbitrary Search API parameters can be entered.
- Query execution is explicit.
- Execution can be canceled.
- Editor state remains consistent.
- The editor works on Windows and Linux.

---

# Step 3.3 — Query Results Viewer

## Objective

Display search responses through reusable Table and JSON interfaces.

## Tasks

### 3.3.1 Reuse Phase 2 Components

Inspect the existing:

- Document Table.
- JSON Inspector.
- Dynamic column handling.
- Row selection.
- Pagination controls.
- Large JSON rendering.

Reuse or extract common components where appropriate.

Do not duplicate the complete Document Table implementation.

Preserve existing Document Browser behavior.

### 3.3.2 Implement Results Table

Display search hits in a table.

Support:

- Dynamic columns.
- Row selection.
- Column resizing.
- Column visibility.
- Long-value truncation.
- Nested-value previews.
- Horizontal scrolling.
- Empty results.
- Loading states.

Search hits may contain additional metadata fields such as:

```text
_rankingScore
_rankingScoreDetails
_formatted
_matchesPosition
```

Do not assume these fields are always present.

Treat them as response metadata when appropriate, without altering their raw representation.

### 3.3.3 Implement Raw JSON Response View

Provide a JSON view of the complete response.

It must include the response envelope, not only the `hits` array.

Users should be able to inspect:

- Search hits.
- Pagination fields.
- Facet information.
- Ranking metadata.
- Additional server-provided fields.

Reuse JSON Inspector where practical.

Support copying the complete JSON response.

Do not silently normalize away unknown fields.

### 3.3.4 Implement View Switching

Provide:

```text
Table | JSON
```

Switching views must not:

- Execute a new search.
- Lose the response.
- Modify the response.
- Reset the editor.
- Trigger unnecessary backend requests.

Keep a single authoritative result snapshot for both views.

### 3.3.5 Display Search Summary

Show relevant metadata such as:

- Number of returned hits.
- Estimated or exact total hits.
- Current pagination information.
- Server processing time.
- Query text returned by the server when available.

Do not confuse:

- `hits.length`
- `estimatedTotalHits`
- `totalHits`

These represent different quantities.

Handle missing values gracefully.

### 3.3.6 Support Search Pagination Modes

Meilisearch supports different search pagination conventions.

**Offset-based pagination**

Uses:

```json
{
  "offset": 0,
  "limit": 20
}
```

**Page-based pagination**

Uses:

```json
{
  "page": 1,
  "hitsPerPage": 20
}
```

The response metadata differs between these modes.

Requirements:

- Identify the active pagination mode.
- Display pagination information correctly.
- Do not fabricate an exact total from `estimatedTotalHits`.
- Do not automatically change the requested pagination mode.
- Preserve unrelated query parameters.
- Ensure navigation updates the submitted JSON consistently.

If the raw query uses an unsupported or ambiguous combination of pagination parameters, avoid silently rewriting it.

Prefer showing the server response and allowing manual editing.

Query pagination controls must remain synchronized with the JSON editor.

Do not create a second independent search query state.

### 3.3.7 Implement Facet Viewer

When the response contains facet metadata, display it in an appropriate panel.

Support:

- `facetDistribution`.
- `facetStats`, when available.
- Dynamic facet names.
- String and numeric facet values.
- Empty facet results.

Do not assume the facet result schema is fixed across all versions.

The complete raw facet data must remain available in JSON view.

Do not implement a full graphical Search Builder.

Interactive facet filtering belongs to a later usability phase unless already supported by the current architecture without additional complexity.

### 3.3.8 Handle Search Highlights Safely

Search responses may contain formatted or highlighted content.

Do not render server-provided highlight text as trusted HTML.

Avoid direct unsafe HTML injection.

Prefer escaped text or a safe, purpose-built highlight renderer.

Never execute scripts or interpret arbitrary HTML supplied by document data.

### 3.3.9 Implement Results Viewer Tests

Test:

- Empty hits.
- Dynamic result fields.
- Nested objects.
- Missing fields.
- Ranking metadata.
- Large numeric values.
- Raw response preservation.
- View switching.
- Offset pagination.
- Page-based pagination.
- Facet distribution.
- Facet statistics.
- Highlighted content safety.
- Large result payloads.

## Acceptance Criteria

- Search hits appear in Table view.
- The full response appears in JSON view.
- Both views share the same result snapshot.
- Phase 2 components are reused.
- Pagination metadata is displayed accurately.
- Facet information is visible.
- Unknown response fields remain inspectable.
- Untrusted highlight content cannot execute HTML or scripts.

---

# Step 3.4 — Query Diagnostics

## Objective

Provide diagnostic information that helps developers understand search behavior and API failures.

## Tasks

### 3.4.1 Implement Execution Timing

Measure client-observed execution duration using Go's monotonic clock facilities.

Record the duration around the relevant HTTP request lifecycle.

Display:

- Server-reported `processingTimeMs`.
- Client-observed request duration.

Explain that these values measure different things.

Do not calculate a definitive network latency by simply subtracting one from the other.

Client duration may include:

- Connection setup.
- Network transfer.
- Server waiting.
- Response transfer.
- Other HTTP client overhead.

Do not represent these measurements as more precise than the implementation supports.

### 3.4.2 Display Ranking Scores

Support `_rankingScore` when returned.

Requirements:

- Preserve the numeric value.
- Display it clearly.
- Avoid inventing a score when absent.
- Do not infer ranking quality from missing metadata.
- Allow inspection of the original value in JSON.

If appropriate, display ranking scores in a dedicated result column or detail panel.

### 3.4.3 Display Ranking Score Details

Support `_rankingScoreDetails` when available.

The response may contain detailed information for ranking rules.

Requirements:

- Display nested ranking details.
- Preserve the original JSON.
- Handle unknown ranking rule names.
- Handle optional fields.
- Avoid hardcoded assumptions about every ranking rule.
- Do not claim that ranking details are available when the server does not return them.

Reuse JSON Inspector for detailed structural inspection.

Do not implement a custom ranking visualization engine in Phase 3.

### 3.4.4 Implement API Error Diagnostics

Display relevant information from Meilisearch API errors:

- HTTP status.
- Error message.
- Error code.
- Error type.
- Documentation link, when provided.

Distinguish:

- Invalid JSON.
- Invalid search parameter.
- Invalid filter.
- Invalid sort.
- Index not found.
- Authentication failure.
- Permission denied.
- Network failure.
- Timeout.
- Cancellation.
- Server error.

Use the structured error handling established in Phase 1.

Do not expose credentials or internal authentication headers.

### 3.4.5 Preserve Failed Query Context

When execution fails, retain:

- The submitted query snapshot.
- The target connection identifier.
- The target index UID.
- The execution timestamp.
- The error details.
- Available timing data.

Do not retain sensitive authentication data.

Ensure that an error from a previous connection cannot be displayed as an error belonging to a newly selected connection.

### 3.4.6 Implement Diagnostics UI

Provide a compact diagnostics area containing:

```text
Status
Returned Hits
Total Hits / Estimated Hits
Server Processing Time
Client Request Duration
Ranking Information
API Error Details
```

Only show fields that are relevant or available.

Avoid excessive permanent UI clutter.

Detailed diagnostic information may be placed in an expandable panel.

### 3.4.7 Implement Diagnostics Tests

Test:

- Successful execution timings.
- Missing server timing.
- Ranking score values.
- Missing ranking score.
- Ranking score details.
- Unknown ranking rules.
- Invalid filter errors.
- Invalid sort errors.
- Authentication errors.
- Permission errors.
- Server errors.
- Cancellation.
- Credential sanitization.

## Acceptance Criteria

- Server and client timings are separately displayed.
- Ranking scores and details can be inspected.
- Errors contain useful diagnostic information.
- Unknown metadata does not break rendering.
- Sensitive information is not disclosed.
- Failed requests preserve enough context for debugging.

---

# Step 3.5 — Query History

## Objective

Implement a lightweight query history system that supports reviewing and rerunning previous searches.

## Tasks

### 3.5.1 Define History Models

A history entry should contain:

```text
History ID
Connection ID
Index UID
Submitted Query JSON
Execution Timestamp
Execution Status
Optional Duration
Optional Error Summary
```

Do not include:

- API Keys.
- Authentication headers.
- Credential references that reveal secrets.
- Full HTTP request headers.
- Raw search result documents by default.

Use immutable submitted-query snapshots.

### 3.5.2 Implement Local Persistence

Use the configuration and persistence infrastructure from Phase 0.

A separate database is not required for the initial history implementation.

Support:

- Loading history on startup.
- Adding history entries.
- Removing entries.
- Clearing history.
- Retention limits.
- Safe recovery from corrupted history.

Store history outside the executable directory.

Use the established application-data location.

### 3.5.3 Define Retention Policy

Use a bounded history.

A reasonable initial default is:

```text
Maximum entries: 100
```

Make the value configurable internally.

Avoid unbounded growth.

Limit the maximum permitted query size for history persistence if necessary.

Do not silently discard the user's active query when a history storage limit is exceeded.

### 3.5.4 Implement Query History UI

Provide a history panel showing:

- Timestamp.
- Connection name when available.
- Index UID.
- Query summary.
- Success or failure status.

Support:

- Select history entry.
- Restore query into editor.
- Rerun query.
- Delete entry.
- Clear history.

Use a confirmation for clearing all history.

Do not require confirmation for ordinary history selection.

### 3.5.5 Handle Connection Changes

History entries must retain their original connection and index association.

When restoring a query:

- Identify the original connection.
- Identify the original index.
- Explain when the connection no longer exists.
- Do not silently run the query against another connection.
- Require explicit user action if changing the current target.

A history entry must never trigger automatic credential disclosure.

### 3.5.6 Implement Rerun Behavior

Rerunning history must:

1. Restore the submitted query.
2. Resolve the target connection.
3. Resolve the target index.
4. Obtain credentials through Connection Manager.
5. Execute the normal Query Execution Service.
6. Update results and diagnostics.

Do not create a separate search execution implementation for history.

### 3.5.7 Address Privacy and Security

Query bodies may contain sensitive application data even when they contain no API Keys.

Therefore:

- History must be stored locally.
- History must not be uploaded automatically.
- Users must be able to delete history.
- Diagnostic logs must not automatically contain full query bodies.
- Do not persist authentication headers.
- Document that query history may contain sensitive search terms.

Provide an option to disable history persistence or use session-only history.

Do not promise encryption unless the implementation actually provides it.

### 3.5.8 Implement History Tests

Test:

- History creation.
- Persistence.
- Reload after restart.
- Retention limit.
- Rerun.
- Connection association.
- Deleted connection.
- Invalid stored JSON.
- Corrupted history file.
- History deletion.
- Clear-all confirmation.
- Disabled persistence.
- Absence of API Keys.

## Acceptance Criteria

- Successful and failed query executions can appear in history.
- History survives restart when persistence is enabled.
- Queries can be restored and rerun.
- History remains bounded.
- Original connection and index associations are preserved.
- Users can delete history.
- Credentials are not persisted in history.
- Session-only or disabled persistence is supported.

---

# Step 3.6 — Integration, Testing & MVP Validation

## Objective

Validate the complete Query Console and confirm that MeiliScope meets its MVP requirements.

## Tasks

### 3.6.1 Verify End-to-End Workflow

Test the following workflow:

1. Launch MeiliScope.
2. Select a saved connection.
3. Select an index.
4. Open Query Console.
5. Enter a JSON search query.
6. Validate the JSON.
7. Execute the query.
8. Inspect returned hits in Table view.
9. Switch to JSON view.
10. Inspect ranking metadata.
11. Inspect available facets.
12. Review execution timings.
13. Change pagination.
14. Execute another query.
15. Review query history.
16. Restore and rerun a previous query.
17. Switch connections.
18. Verify that stale results are not displayed.
19. Restart the application.
20. Verify saved history behavior.

Use real Meilisearch instances for integration validation where practical.

### 3.6.2 Test Search Parameters

Create test cases covering:

- Empty query.
- Keyword search.
- Filter expressions.
- Sort parameters.
- Facets.
- Limit and offset.
- Page and hitsPerPage.
- Attributes to retrieve.
- Highlighting parameters.
- Ranking scores.
- Ranking score details.
- Unknown parameters.
- Empty results.
- Large result sets.

Use current official Meilisearch documentation to determine expected behavior.

Do not assume every feature is available in every supported server version.

### 3.6.3 Test Error Scenarios

Validate:

- Invalid JSON.
- Non-object JSON.
- Invalid filter syntax.
- Invalid sort configuration.
- Nonexistent index.
- Invalid API Key.
- Search-only API Key.
- Permission denied.
- Server unavailable.
- Timeout.
- Request cancellation.
- Rapid repeated execution.
- Connection switching during execution.
- Index switching during execution.
- Malformed server response.
- Oversized response.
- History corruption.

The application must remain responsive.

### 3.6.4 Validate Search-only API Keys

A Search API Key may have search permission without index enumeration permission.

Ensure Query Console does not unnecessarily depend on successful Index Explorer listing.

Where the server authorizes a known index for search, users should be able to target that index even if index enumeration is unavailable.

Use a safe, explicit mechanism for selecting or entering a known index UID.

Do not bypass Meilisearch authorization.

Do not require an Admin API Key for ordinary search execution.

### 3.6.5 Verify Cross-platform Functionality

Test Windows and Linux.

Validate:

- Query editor rendering.
- Editor keyboard shortcuts.
- Monaco worker loading, if Monaco is used.
- Search execution.
- Cancellation.
- Results Table.
- JSON Inspector.
- Clipboard operations.
- Query history.
- Persistent configuration.
- Application shutdown.
- Release builds.

Reuse existing CI and build infrastructure.

Do not claim compatibility with untested operating-system versions or distributions.

### 3.6.6 Add Automated Tests

Backend unit tests:

- Query Execution Service.
- Request validation.
- Response decoding.
- Error classification.
- Cancellation.
- Timing.
- Response preservation.

Frontend tests:

- Query editor validation.
- Query execution state.
- Results rendering.
- Pagination state.
- Ranking diagnostics.
- Facet rendering.
- History management.
- Stale-response prevention.

Integration tests:

- Complete query execution workflow.
- Connection switching.
- Authorization scenarios.
- Real Meilisearch search responses.

Use deterministic fixtures and mock servers for the default test suite.

Do not make a live Meilisearch server mandatory for all unit tests.

### 3.6.7 Performance Validation

Create or reuse representative datasets.

Test:

- Repeated search execution.
- Large result payloads.
- Rapid query switching.
- Large JSON responses.
- Large result tables.
- Query history growth.
- Memory usage during repeated searches.

Measure:

- HTTP request duration.
- Response decoding overhead.
- Frontend rendering time.
- User-visible execution latency.

Document actual observations.

Do not invent benchmark numbers or performance improvements.

### 3.6.8 Security Review

Verify:

- Search requests use the selected connection credentials.
- API Keys are never embedded in query history.
- API Keys are never returned through frontend bindings.
- Full query bodies are not logged by default.
- Search-only keys work where authorized.
- Search requests cannot invoke arbitrary HTTP endpoints.
- Query cancellation does not leak credentials.
- Highlighted content is rendered safely.
- Results are isolated by connection and index.
- Saved history remains local.

### 3.6.9 Update Documentation

Update the project's existing documentation structure.

Recommended files:

```text
README.md

docs/en/user-guide/query-console.md
docs/ja/user-guide/query-console.md

docs/en/user-guide/query-history.md
docs/ja/user-guide/query-history.md

docs/architecture/overview.md
docs/development/phase-3-report.md
```

Document:

- Opening Query Console.
- Selecting a connection and index.
- Executing a search.
- JSON validation.
- Supported pagination behavior.
- Interpreting ranking scores.
- Inspecting facets.
- Understanding execution timings.
- Diagnosing API errors.
- Managing query history.
- Using search-only API Keys.
- Security and privacy considerations.

Include practical query examples.

Do not use examples that require administrative permissions.

### 3.6.10 Validate MVP Requirements

Review the MVP definition in `plan.md`.

Confirm that the application supports:

**Connection**

- Creating and selecting server profiles.
- Secure credentials.
- Connection status and errors.

**Explore**

- Index enumeration.
- Index metadata and statistics.
- Document browsing.
- Document inspection.
- Server-side pagination.

**Query**

- Arbitrary Search API JSON execution.
- Query validation.
- Table and JSON result views.
- Search metadata.
- Ranking diagnostics.
- Error diagnostics.
- Query history.

Document any remaining MVP gaps.

Do not declare MVP completion if required functionality remains unimplemented.

### 3.6.11 Prepare MVP Completion Report

Create:

`docs/development/phase-3-report.md`

Include:

1. Completed implementation steps.
2. Architecture changes.
3. Query editor technology decision.
4. Search API compatibility findings.
5. Query execution implementation.
6. Result viewer integration.
7. Ranking and diagnostics implementation.
8. History persistence behavior.
9. Backend test results.
10. Frontend test results.
11. Windows validation.
12. Linux validation.
13. Performance observations.
14. Security review.
15. Known limitations.
16. MVP acceptance status.
17. Phase 4 readiness.

## Acceptance Criteria

- Query Console is fully integrated.
- Arbitrary valid Search API JSON requests can be executed.
- Search results can be viewed as Table and JSON.
- Search metadata is displayed correctly.
- Ranking scores can be inspected when available.
- Facet data can be inspected.
- Server and client timings are distinguished.
- Query history and rerun work.
- API errors are understandable.
- Search-only permissions are handled appropriately.
- Windows and Linux builds remain functional.
- Automated tests pass.
- MVP requirements are satisfied or remaining gaps are explicitly reported.

---

# 4. General Implementation Constraints

## 4.1 Preserve the Existing Architecture

Do not replace working Phase 0, Phase 1, or Phase 2 components.

Refactor shared components only when necessary.

Prefer incremental, compatible changes.

## 4.2 Do Not Implement Search Builder

Search Builder is planned for Phase 5.

Do not introduce a separate GUI form for every Meilisearch search parameter.

Phase 3 is primarily a JSON Query Console.

## 4.3 Do Not Implement Document Editing

Do not implement:

- Document insertion.
- Document updates.
- Document deletion.
- Index modification.
- Index deletion.
- Administrative API operations.

Phase 3 remains read-only with respect to Meilisearch data.

## 4.4 Do Not Implement a Generic HTTP Console

Only the Search API endpoint is in scope.

Do not allow arbitrary endpoint selection.

Do not allow arbitrary HTTP methods.

A generic HTTP Request Console belongs to a future optional phase.

## 4.5 Preserve JSON Transparency

Avoid hardcoded query parameter allowlists.

Preserve unknown request and response fields.

Do not silently normalize server responses.

Do not lose JSON numeric precision.

## 4.6 Reuse UI Components

Reuse Phase 2's Document Table and JSON Inspector.

Do not create duplicate table or tree implementations unless a documented technical limitation makes reuse impractical.

## 4.7 Prevent Stale Results

A result must always be associated with the exact:

- Connection.
- Index.
- Submitted query.
- Execution identifier.

Old responses must not overwrite newer execution results.

## 4.8 Keep Query History Safe

Do not persist credentials.

Do not upload history.

Do not retain complete result documents in history by default.

Support clearing history.

## 4.9 Keep Dependencies Minimal

Use existing project libraries where practical.

Evaluate Monaco Editor before adoption.

Do not introduce unnecessary global state libraries, databases, or backend services.

## 4.10 Maintain Cross-platform Support

Preserve Windows and Linux builds.

Avoid platform-specific assumptions in shared application logic.

Use the existing Wails integration.

---

# 5. Execution & Reporting Requirements

Execute the steps in order:

1. Step 3.1 — Search API & Query Execution.
2. Step 3.2 — JSON Query Editor.
3. Step 3.3 — Query Results Viewer.
4. Step 3.4 — Query Diagnostics.
5. Step 3.5 — Query History.
6. Step 3.6 — Integration, Testing & MVP Validation.

Each step should be independently reviewable.

After each step, report:

### Completed Work

Describe implemented functionality.

### Files Changed

List important source and documentation changes.

### Tests

List executed commands and actual results.

### Technical Decisions

Describe implementation choices and their rationale.

### Known Issues

Identify unresolved problems and deferred work.

### Next Step Readiness

State whether the following step can proceed safely.

Do not claim tests passed unless executed.

Do not hide known failures.

Avoid implementing functionality from later phases without a clear technical necessity.

---

# 6. Phase 3 Definition of Done

Phase 3 is complete when a user can:

1. Launch MeiliScope on Windows or Linux.
2. Connect to Meilisearch.
3. Select an index or explicitly target an authorized index UID.
4. Open Query Console.
5. Enter an arbitrary valid Search API JSON object.
6. Validate and execute the query.
7. View results in Table format.
8. View the complete response in JSON format.
9. Inspect individual search hits.
10. Review pagination metadata.
11. Inspect available facet information.
12. Inspect ranking score metadata.
13. Compare server processing time and client-observed duration.
14. Diagnose API errors.
15. Cancel an active search.
16. Review query history.
17. Restore and rerun previous queries.
18. Switch connections without stale-data leakage.
19. Restart the application without losing enabled persistent history.
20. Continue normal operation after recoverable errors.

The implementation must remain read-only with respect to Meilisearch data.

The application must not require a separate local MeiliScope server.

All required functionality must be validated by automated tests or documented integration tests.

**Successful completion of Phase 3 establishes MeiliScope's initial MVP, targeting version 0.1.0.**

The official release process, additional usability features, document editing, and broader quality validation remain assigned to subsequent phases.
