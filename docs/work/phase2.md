# MeiliScope — Phase 2 Implementation Instructions

## Document Browser & JSON Inspector

### 1. Project Context

MeiliScope is an open-source desktop application designed to explore and inspect Meilisearch data.

**Mission: Explore your Meilisearch data.**

The application uses:

- Go for backend services.
- Wails v3 for desktop application integration.
- React and TypeScript for the frontend.
- Meilisearch HTTP API for data access.
- Windows and Linux as currently validated development and execution platforms.
- MIT License.

Phase 0 established the application foundation, architecture, build infrastructure, and cross-platform desktop support.

Phase 1 implemented the Meilisearch HTTP client, secure credential storage, Connection Manager, Index API, and Index Explorer.

Phase 2 builds directly on those components.

The root-level `plan.md` is the authoritative project proposal.

Before making changes:

1. Read `plan.md`.
2. Review the existing architecture documentation.
3. Review the Phase 1 implementation and completion report, if available.
4. Inspect the current Go services and frontend structure.
5. Reuse existing connection, authentication, cancellation, and error-handling mechanisms.

Do not recreate or replace working Phase 0 or Phase 1 infrastructure without a documented technical reason.

---

# 2. Phase 2 Objectives

Implement a read-only Document Browser capable of exploring Meilisearch documents efficiently.

Users must be able to:

1. Select an index from Index Explorer.
2. Browse its documents in a table.
3. Navigate document pages.
4. Inspect individual documents.
5. View complete JSON structures.
6. Expand and collapse nested JSON.
7. Copy values and JSON content.
8. Select visible fields.
9. Refresh document data.
10. Handle indexes containing large numbers of documents.

The implementation must support Meilisearch's schema-flexible document model.

Documents within the same index may contain different fields, different JSON types, and deeply nested structures.

**Do not assume that all documents share an identical schema.**

## 2.1 Primary Design Principles

**Read-only by default**

Phase 2 must not modify documents.

**Server-side pagination**

Never load an entire index simply to display a table.

**Schema flexibility**

Preserve arbitrary document fields and JSON structures.

**Responsive UI**

Network requests and large JSON structures must not block the application interface.

**Incremental development**

Reuse Phase 1 services and establish reusable components for later phases.

**Cross-platform compatibility**

Preserve Windows and Linux functionality.

---

# 3. Architecture

Maintain the existing frontend/backend separation.

```text
React / TypeScript Frontend
          |
          +-- Index Explorer
          |
          +-- Document Browser
          |     |
          |     +-- Document Table
          |     +-- Pagination
          |     +-- Column Management
          |
          +-- JSON Inspector
          |
          v
Wails v3 Bindings
          |
          v
Go Application Services
          |
          +-- Connection Manager
          |
          +-- Index Service
          |
          +-- Document Service
          |
          v
Meilisearch HTTP Client
          |
          v
Meilisearch REST API
```

The frontend must not communicate directly with Meilisearch.

The Document Service must be independently testable without Wails.

The Document Browser and JSON Inspector must not contain duplicated HTTP request logic.

Preserve the existing connection security boundary.

Raw API Keys must never be included in frontend models, logs, or diagnostic information.

---

# Step 2.1 — Document API & Data Models

## Objective

Implement read-only Document retrieval services and flexible data models.

## Tasks

### 2.1.1 Verify the Meilisearch Document API

Consult the official Meilisearch API documentation.

Verify the current API specifications for:

- Listing documents.
- Retrieving an individual document.
- Selecting returned fields.
- Pagination.
- Fetching documents through the dedicated fetch endpoint.
- Document identifier encoding.
- Authentication and authorization.
- API compatibility across the versions supported by MeiliScope.

Relevant API operations include:

```http
GET /indexes/{index_uid}/documents
GET /indexes/{index_uid}/documents/{document_id}
POST /indexes/{index_uid}/documents/fetch
```

Use the first two operations for the basic browser.

The fetch endpoint may be used when it provides a concrete advantage, but do not introduce unnecessary functionality.

Verify supported query parameters and response structures against the actual target Meilisearch version.

Do not assume that pagination or filtering behavior is identical to the Search API.

### 2.1.2 Implement Document Service

Introduce a service responsible for document retrieval.

Provide operations equivalent to:

```go
type DocumentService interface {
    ListDocuments(
        ctx context.Context,
        connectionID string,
        indexUID string,
        options DocumentListOptions,
    ) (DocumentPage, error)

    GetDocument(
        ctx context.Context,
        connectionID string,
        indexUID string,
        documentID string,
    ) (Document, error)
}
```

Adapt these signatures to existing project conventions.

The service must:

- Reuse the Phase 1 HTTP client.
- Reuse connection profiles and credentials.
- Respect context cancellation.
- Preserve structured errors.
- Avoid unnecessary network requests.
- Avoid introducing GUI dependencies.

### 2.1.3 Implement Pagination Models

Support document listing using offset and limit.

A typical response contains:

```json
{
  "results": [
    {
      "id": 1,
      "title": "Example Document"
    }
  ],
  "offset": 0,
  "limit": 20,
  "total": 125000
}
```

Implement typed pagination metadata.

Suggested models:

```text
DocumentListOptions
DocumentPage
Document
DocumentIdentifier
DocumentField
```

Use the names and structure most appropriate for the existing codebase.

Validate pagination parameters.

Handle:

- Empty result sets.
- Zero total documents.
- Offset beyond the available results.
- Invalid limits.
- Server-reported pagination metadata.
- Large total counts.

Avoid hardcoded assumptions about the number of documents in an index.

### 2.1.4 Preserve Arbitrary JSON

Documents are dynamic JSON objects.

Do not model documents using a fixed structure such as:

```go
type Document struct {
    ID    int
    Title string
}
```

Use an appropriate dynamic representation.

Preserve:

- Strings.
- Numbers.
- Booleans.
- Null values.
- Arrays.
- Nested objects.
- Empty objects.
- Empty arrays.
- Unicode text.

Avoid unnecessary numeric conversion.

In particular, do not silently lose precision when decoding large JSON integer values.

Consider `json.RawMessage`, `json.Number`, or another suitable representation.

Document identifiers may be strings or numbers.

Preserve their exact values for later individual retrieval.

Do not derive document identifiers from table row positions.

### 2.1.5 Handle Field Selection

Support the documented Meilisearch mechanism for selecting returned fields.

The service must accept an optional list of field names.

Requirements:

- Preserve arbitrary field names.
- Avoid unsafe manual URL concatenation.
- Support nested field names where permitted by the API.
- Distinguish an unspecified field selection from an explicit selection.
- Preserve the primary-key field when required by the browser's internal behavior.

Do not assume that every field returned by Meilisearch is suitable for displaying as a table column.

### 2.1.6 Implement Individual Document Retrieval

Retrieve a selected document by its primary key.

Use the index's primary-key metadata when necessary.

Handle:

- String identifiers.
- Numeric identifiers.
- Special characters requiring URL encoding.
- Missing primary-key metadata.
- Document not found.
- Permission denied.
- Index deleted while browsing.

Do not silently fall back to a different document when a requested document is unavailable.

### 2.1.7 Implement Unit Tests

Use the existing mock HTTP testing infrastructure.

Test:

- Standard document listing.
- Empty index.
- Pagination.
- Invalid pagination input.
- Dynamic fields.
- Nested objects.
- Arrays.
- Null values.
- Large integer values.
- Unicode field names and values.
- Numeric and string document identifiers.
- Missing document.
- Invalid JSON responses.
- HTTP 401/403/404/500.
- Timeout.
- Context cancellation.

## Acceptance Criteria

- Document listing works through the existing HTTP client.
- Individual document retrieval works.
- Arbitrary JSON structures are preserved.
- Pagination metadata is accurate.
- Document IDs are not corrupted by JSON number conversion.
- Tests pass without launching Wails.
- No document write operations are introduced.

---

# Step 2.2 — Document Table UI

## Objective

Implement a flexible document table for browsing schema-less Meilisearch documents.

## Tasks

### 2.2.1 Evaluate Table Components

Review the table component candidates identified in the project proposal.

TanStack Table is a preferred candidate.

Evaluate the currently maintained version and its compatibility with the existing React/TypeScript stack.

Select a component or implementation strategy that supports:

- Dynamic columns.
- Row selection.
- Column resizing.
- Flexible cell rendering.
- Controlled pagination.
- Virtualization when necessary.
- Keyboard accessibility.

Avoid introducing a heavy data-grid dependency unless its benefits justify the added complexity.

Document any significant dependency decision.

### 2.2.2 Implement Document Table

The table must display documents returned by Document Service.

Provide:

- Column headers.
- Document rows.
- Row selection.
- Visible selection state.
- Resizable columns.
- Horizontal scrolling.
- Loading state.
- Empty state.
- Error state.
- Refresh action.

Do not require a fixed document schema.

### 2.2.3 Implement Dynamic Column Discovery

Derive available columns from document fields.

For the initially loaded page, determine the union of top-level fields across the returned documents.

Example:

```json
[
  {
    "id": 1,
    "title": "First",
    "category": "A"
  },
  {
    "id": 2,
    "title": "Second",
    "price": 100
  }
]
```

The table should support columns equivalent to:

```text
id | title | category | price
```

A missing field must not be treated as equivalent to an explicit JSON null.

Ensure columns remain stable when navigating between pages.

Consider maintaining a discovered-field union for the active browsing session.

Do not scan the entire index solely to discover columns.

Make it clear that discovered columns may not represent every field in the index.

### 2.2.4 Implement Cell Rendering

Support predictable rendering for:

| JSON type | Table representation |
|---|---|
| String | Text |
| Number | Number |
| Boolean | true / false |
| Null | null |
| Missing field | Empty or distinct placeholder |
| Object | Compact preview |
| Array | Compact preview |

Avoid rendering a complete nested object inside a table cell.

Provide truncation for large values.

Allow users to access the complete value through JSON Inspector.

Use type-aware formatting where practical.

Do not convert all values into strings in the underlying data model.

### 2.2.5 Implement Column Management

Support:

- Column visibility.
- Column resizing.
- Horizontal scrolling.
- Reasonable default widths.
- Persistent widths while browsing the same index.
- A reset-to-default action.

Column preferences may be stored locally using the existing settings infrastructure.

Do not create a new database for column settings.

Keep the initial implementation simple.

### 2.2.6 Implement Row Selection

Selecting a row must:

- Update the current document selection.
- Notify JSON Inspector.
- Preserve the correct document identity.
- Show a visible selection state.

When the selected row is no longer available, clear or update selection appropriately.

Do not accidentally select a different document after pagination or refresh.

### 2.2.7 Implement Copy Operations

Support copying:

- Individual cell values.
- The selected document's primary-key value.
- Selected document JSON through JSON Inspector.

Use the application's existing clipboard integration.

Avoid copying truncated display text when the user requests the complete underlying value.

### 2.2.8 Handle Large Cell Values

Documents may contain:

- Long text fields.
- Large arrays.
- Deep objects.
- Unicode content.
- Unexpectedly large strings.

Prevent oversized cells from causing excessive row height or layout instability.

Use compact previews and explicit inspection actions.

## Acceptance Criteria

- Documents appear in a functional table.
- Different documents may have different columns.
- Column resizing and visibility controls work.
- Nested JSON values do not break the layout.
- Row selection works reliably.
- Missing and null fields remain distinguishable.
- Large values are truncated safely for display.
- The table works on Windows and Linux.

---

# Step 2.3 — Pagination & Large Dataset Handling

## Objective

Ensure the Document Browser remains responsive when exploring large Meilisearch indexes.

## Tasks

### 2.3.1 Implement Server-side Pagination

Use Meilisearch's document pagination API.

Provide:

- Previous page.
- Next page.
- Current page indicator.
- Page size selection.
- Total document count.
- First-page navigation where useful.

Suggested initial page sizes:

```text
25
50
100
200
```

Use an appropriate default, such as 50.

Validate the selected page size.

Do not fetch all documents for pagination.

The UI must use server-reported totals rather than the number of currently loaded rows.

### 2.3.2 Define Pagination State

Maintain pagination state for the active connection and index.

At minimum:

```text
Connection ID
Index UID
Offset
Limit
Total
Loading State
Error State
```

A consistent state transition must occur when changing indexes.

Reset or restore pagination intentionally.

Do not accidentally retain an offset that belongs to another index.

### 2.3.3 Prevent Stale Responses

Document requests may complete out of order.

Example:

1. User opens Index A.
2. Request for Index A begins.
3. User immediately opens Index B.
4. Request for Index B completes.
5. Request for Index A completes later.

The response for Index A must not overwrite the visible data for Index B.

Use:

- Request cancellation.
- Request generation identifiers.
- Connection and index identity checks.

Apply the same protection to pagination and document detail requests.

### 2.3.4 Implement Virtualized Rendering

Use virtualization if needed to keep rendering efficient.

TanStack Virtual is a preferred candidate if TanStack Table is selected.

Virtualization should operate on the currently loaded page.

Do not implement virtual scrolling by fetching the entire index into memory.

Ensure:

- Smooth scrolling.
- Stable row identity.
- Reasonable overscan.
- Correct column alignment.
- Predictable row selection.
- No unnecessary re-rendering of every row.

If page sizes are sufficiently small and virtualization does not provide a measurable benefit, document that decision.

### 2.3.5 Implement Controlled Data Loading

Ensure that:

- Only the required page is loaded.
- Requests can be canceled.
- Repeated navigation does not trigger uncontrolled requests.
- Refresh does not duplicate data.
- Large indexes do not cause proportional client-memory usage.

If caching is introduced, keep it bounded.

Use connection ID, index UID, and request parameters as part of the cache identity.

Do not mix cached documents from different connections.

### 2.3.6 Handle Large Documents

A single document may contain a very large JSON payload.

Provide safe handling for:

- Large strings.
- Large nested arrays.
- Many object fields.
- Large response bodies.

Use bounded response handling where practical.

Avoid constructing multiple unnecessary copies of large JSON payloads.

Do not introduce arbitrary limits without documenting their behavior.

If a document exceeds a supported viewing limit, display an explicit error or limitation rather than silently truncating the underlying data.

### 2.3.7 Performance Tests

Create repeatable test datasets containing:

- 1,000 documents.
- 10,000 documents.
- 100,000 documents.
- Documents with hundreds of fields.
- Large nested JSON documents.

Performance testing should distinguish:

- HTTP response time.
- JSON decoding time.
- Frontend rendering time.
- Total user-visible loading time.

Do not treat network latency as frontend rendering cost.

Avoid inventing performance results.

Measure and document actual observations.

Use a disposable Meilisearch test instance where practical.

## Acceptance Criteria

- Document browsing uses server-side pagination.
- The application does not load an entire large index.
- Pagination remains responsive.
- Stale responses cannot overwrite newer selections.
- Large documents do not freeze the application under tested conditions.
- Memory usage does not grow proportionally to total index size during ordinary browsing.
- Performance findings are documented.

---

# Step 2.4 — JSON Inspector

## Objective

Implement an interactive, read-only JSON Inspector for selected documents.

## Tasks

### 2.4.1 Implement JSON Tree View

Display the selected document as a structured JSON tree.

Support:

- Objects.
- Arrays.
- Strings.
- Numbers.
- Booleans.
- Null.
- Empty objects.
- Empty arrays.

Each tree node must preserve its original JSON type.

Do not represent every value as a formatted string internally.

### 2.4.2 Implement Expand and Collapse

Support:

- Expanding individual nodes.
- Collapsing individual nodes.
- Expanding nested arrays.
- Collapsing nested objects.
- Default expansion depth.
- Expand-all and collapse-all where practical.

Avoid eagerly rendering every nested node of a very large document.

Prefer lazy rendering of collapsed branches.

Do not let a deeply nested document cause uncontrolled recursion or stack overflow.

### 2.4.3 Implement JSON Formatting

Display JSON in a readable format.

Provide:

- Indentation.
- Syntax highlighting or type-specific visual distinction.
- Escaped string handling.
- Unicode text.
- Clear null and boolean representations.

A dedicated JSON tree component may be used.

Evaluate its maintenance status, dependency size, accessibility, and compatibility before adoption.

Avoid installing Monaco Editor solely for JSON inspection unless it provides a clear benefit.

Monaco may be introduced later for Query Console and Document Editor.

### 2.4.4 Implement JSON Path Display

Show the path of a selected JSON node.

Example:

```text
$.metadata.author.name
```

For field names requiring escaping, support a correct alternative representation.

Consider JSON Pointer as the canonical machine-readable path format.

Example:

```text
/metadata/author/name
```

Implement path generation correctly for:

- Nested objects.
- Arrays.
- Field names containing dots.
- Field names containing slashes.
- Field names containing tildes.
- Unicode field names.

Do not generate ambiguous paths.

### 2.4.5 Implement Copy Operations

Support:

- Copy selected primitive value.
- Copy selected node as JSON.
- Copy selected JSON path.
- Copy entire document as formatted JSON.
- Copy entire document as compact JSON.

Copy operations must use the actual underlying value.

Do not copy abbreviated table previews.

Preserve JSON numeric precision whenever possible.

Do not silently alter JSON types during serialization.

### 2.4.6 Implement Inspector Integration

When the selected table row changes:

- Update the Inspector.
- Show the correct document.
- Clear the previous selection when appropriate.
- Avoid displaying stale data.
- Preserve loading and error states.

Support individual document retrieval when complete details are required.

A document loaded with selected fields must not be presented as the complete document.

If the full document is not yet available, retrieve it explicitly.

### 2.4.7 Implement Inspector Layout

Use the existing application shell.

The Inspector may appear:

- On the right side of Document Browser.
- In a resizable split pane.
- In a dedicated detail view where appropriate.

Prefer a layout that allows users to inspect a document without losing table context.

Support reasonable minimum widths and resizing behavior.

Avoid modal dialogs for ordinary inspection.

### 2.4.8 Read-only Enforcement

The Inspector must be read-only.

Do not implement:

- Inline value editing.
- JSON mutation.
- Document updates.
- Document deletion.
- Save actions.

Prepare component boundaries so that Phase 4 can introduce a separate editing workflow.

Do not mix read-only inspection with mutation logic.

### 2.4.9 JSON Inspector Tests

Test:

- Primitive values.
- Nested objects.
- Arrays.
- Empty collections.
- Null values.
- Unicode.
- Large integers.
- Special field names.
- Deep nesting.
- Long strings.
- JSON path generation.
- Copy behavior.
- Document switching.
- Missing or deleted documents.

## Acceptance Criteria

- Selected documents appear in JSON Inspector.
- Nested nodes can be expanded and collapsed.
- JSON types remain accurate.
- Values, paths, and complete JSON can be copied.
- Large JSON structures remain usable under tested conditions.
- The Inspector never modifies server data.
- Document selection remains synchronized with the table.

---

# Step 2.5 — Integration, Testing & Documentation

## Objective

Complete the Document Browser workflow and validate Phase 2 across Windows and Linux.

## Tasks

### 2.5.1 Integrate with Index Explorer

Implement the workflow:

1. Launch MeiliScope.
2. Select a connection.
3. Browse the index list.
4. Select an index.
5. Open Document Browser.
6. Fetch the first document page.
7. Select a document.
8. Inspect its JSON.
9. Change pages.
10. Refresh the document list.
11. Switch indexes.
12. Switch connections.

The selected connection and index must be managed consistently.

Avoid duplicating connection state across unrelated frontend components.

Reuse the application state conventions established in previous phases.

### 2.5.2 Implement Complete Loading States

Display appropriate feedback for:

- No connection selected.
- No index selected.
- Index loading.
- Documents loading.
- Empty index.
- Document loading.
- Refreshing documents.
- Request canceled.
- Connection unavailable.
- Access denied.
- Document not found.

A failed request must not leave the UI permanently displaying a loading indicator.

### 2.5.3 Implement Error Recovery

Support user recovery from:

- Expired or revoked credentials.
- Permission denied.
- Connection refused.
- Server timeout.
- Index deleted remotely.
- Document deleted remotely.
- Invalid response.
- Unexpected server error.

Reuse Phase 1's error classification.

Do not display raw internal error objects when a clear user-facing message is available.

Preserve technical details for debugging without exposing credentials.

### 2.5.4 Validate Dynamic Document Structures

Create test fixtures containing:

**Different field sets**

Documents containing different top-level fields.

**Nested objects**

Documents containing multiple nested object levels.

**Arrays**

Arrays of primitives, objects, and mixed values.

**Null and missing fields**

Explicit null values and absent fields.

**Large values**

Long text and large numeric values.

**Unicode**

Japanese text, emoji, and multilingual content.

**Special identifiers**

String and numeric primary keys requiring correct encoding.

**Empty collections**

Empty arrays and objects.

Verify that both the table and JSON Inspector display these cases correctly.

### 2.5.5 Validate Permission Scenarios

Test with:

- A key authorized to read documents.
- A key that can list indexes but cannot retrieve documents.
- A key restricted to specific indexes.
- An invalid key.
- An unauthenticated development server.

Do not infer document-read permissions solely from successful Index Explorer requests.

Document the required Meilisearch API Key actions.

### 2.5.6 Validate Cross-platform Operation

Run relevant tests on:

- Windows.
- Linux.

Verify:

- Document table rendering.
- Pagination.
- Clipboard operations.
- JSON Inspector.
- Resizable layout.
- Large document behavior.
- Application shutdown.
- Connection switching.

Preserve the build and runtime support achieved during Phase 0 and Phase 1.

### 2.5.7 Extend Automated Tests

Add backend unit tests for:

- Document Service.
- JSON decoding.
- Pagination validation.
- Document identifiers.
- HTTP error handling.

Add frontend tests for:

- Dynamic columns.
- Table cell formatting.
- Row selection.
- Pagination state.
- JSON path generation.
- JSON Inspector.
- Stale-response prevention.

Add integration tests for the complete Document Browser flow where practical.

Prefer deterministic tests using fixtures and mock HTTP services.

Real Meilisearch integration tests must use disposable test data.

### 2.5.8 Performance Validation

Use test datasets containing large numbers of documents.

Measure:

- Page loading.
- Table rendering.
- Memory usage.
- Document switching.
- JSON Inspector responsiveness.
- Connection switching during requests.

Record the actual test environment and results.

Investigate clear performance regressions before completing Phase 2.

### 2.5.9 Update Documentation

Update the existing project documentation.

At minimum:

```text
README.md
docs/en/user-guide/document-browser.md
docs/ja/user-guide/document-browser.md
docs/en/user-guide/json-inspector.md
docs/ja/user-guide/json-inspector.md
docs/architecture/overview.md
docs/development/phase-2-report.md
```

Use paths consistent with the repository's established documentation structure.

Document:

- Opening Document Browser.
- Navigating pages.
- Selecting documents.
- Customizing columns.
- Inspecting JSON.
- Copying JSON values.
- Read-only limitations.
- Handling large documents.
- Troubleshooting permission errors.

Update `plan.md` only if implementation decisions materially change the project proposal or roadmap.

Do not rewrite unrelated project sections.

### 2.5.10 Final Phase Report

Create:

`docs/development/phase-2-report.md`

Include:

1. Completed implementation steps.
2. Files created or modified.
3. Document API implementation details.
4. Table component selection and rationale.
5. JSON Inspector implementation.
6. Pagination and virtualization decisions.
7. Backend test results.
8. Frontend test results.
9. Windows validation.
10. Linux validation.
11. Performance observations.
12. Known limitations.
13. Unresolved issues.
14. Readiness for Phase 3.

## Acceptance Criteria

- Index Explorer and Document Browser are integrated.
- Documents can be viewed in a table.
- Pagination works.
- Dynamic document fields are supported.
- Individual documents can be inspected.
- JSON Inspector supports nested structures.
- Copy operations work.
- Connection switching does not expose stale data.
- Large datasets do not require full-index retrieval.
- Automated tests pass.
- Documentation reflects the implementation.
- Windows and Linux builds remain functional.

---

# 4. General Implementation Constraints

## 4.1 Preserve Read-only Behavior

Phase 2 must not perform any Meilisearch write operations.

Do not introduce:

- Document creation.
- Document updates.
- Document deletion.
- Index modification.
- Index creation or deletion.
- Task administration.

These belong to later phases or are outside the project scope.

## 4.2 Do Not Implement Query Console

Do not implement:

- Search query editing.
- Search execution.
- Ranking score visualization.
- Query history.
- Search Builder.

These belong to Phase 3 and Phase 5.

The Document Browser must use Document retrieval APIs, not Search API as a replacement for ordinary document listing.

## 4.3 Avoid Premature Export Implementation

JSON and CSV export are planned for a later phase.

Basic clipboard operations are in scope.

Do not implement a complete export subsystem during Phase 2.

## 4.4 Preserve Backend/Frontend Boundaries

Go owns:

- Meilisearch HTTP communication.
- Authentication.
- Document retrieval.
- API error classification.

React owns:

- Table presentation.
- Pagination controls.
- Column visibility.
- Selection state.
- JSON Inspector presentation.

Do not move HTTP authentication logic into the frontend.

## 4.5 Avoid Unbounded Memory Usage

Do not:

- Fetch all documents automatically.
- Retain every visited page indefinitely.
- Expand an entire large JSON tree eagerly.
- Render thousands of rows unnecessarily.
- Store duplicate full-document objects without a clear need.

Prefer bounded, on-demand loading.

## 4.6 Preserve JSON Fidelity

Do not silently lose:

- Numeric precision.
- JSON types.
- Unicode content.
- Nested structure.
- Field names.
- Distinctions between missing fields and null values.

When a value is intentionally shortened for visual display, preserve access to the complete underlying data.

## 4.7 Keep Dependencies Justified

Prefer existing project dependencies.

If introducing TanStack Table, TanStack Virtual, or a JSON tree library:

- Verify current maintained versions.
- Confirm React compatibility.
- Document the choice.
- Avoid unnecessary overlapping UI libraries.

## 4.8 Maintain Cross-platform Compatibility

Do not introduce Windows-only behavior into shared frontend or Go services.

Use existing platform abstractions.

Preserve Windows and Linux build compatibility.

---

# 5. Execution & Reporting Requirements

Execute the steps in the following order:

1. Step 2.1 — Document API & Data Models.
2. Step 2.2 — Document Table UI.
3. Step 2.3 — Pagination & Large Dataset Handling.
4. Step 2.4 — JSON Inspector.
5. Step 2.5 — Integration, Testing & Documentation.

After each step, report:

**Completed Work**

Describe the implementation.

**Files Changed**

List significant source files and documentation.

**Tests**

Provide commands executed and results.

**Technical Decisions**

Describe relevant implementation choices.

**Known Issues**

Identify unresolved issues and deferred functionality.

**Next Step Readiness**

State whether the next step can proceed.

Do not claim successful validation unless tests were actually executed.

Do not leave unrelated existing tests broken.

If tests fail because of an existing issue, document the issue and its impact.

---

# 6. Phase 2 Definition of Done

Phase 2 is complete when a user can:

1. Launch MeiliScope on Windows or Linux.
2. Connect to a Meilisearch instance.
3. Select an index.
4. Open Document Browser.
5. View the first page of documents.
6. Navigate to another page.
7. Adjust visible columns.
8. Select a document.
9. Inspect complete document JSON.
10. Expand and collapse nested JSON structures.
11. Copy individual values and complete JSON.
12. Refresh document data.
13. Switch indexes and connections safely.
14. Continue using the application after recoverable errors.

The application must not load an entire large index solely to support browsing.

Document data must remain read-only throughout Phase 2.

All important behavior must be covered by appropriate automated or documented integration tests.

**Phase 2 must deliver a reliable Document Browser and JSON Inspector, establishing the foundation for Phase 3 — Query Console.**
