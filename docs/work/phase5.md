# MeiliScope — Phase 5 Implementation Instructions

## Usability & Workflow Improvements

### 1. Project Context

MeiliScope is an open-source desktop application for exploring Meilisearch indexes, browsing documents, executing search queries, and performing controlled document modifications.

**Mission: Explore your Meilisearch data.**

The established technology stack includes:

- Go for backend services.
- Wails v3 for desktop integration.
- React and TypeScript for the frontend.
- Meilisearch HTTP API for data access.
- Windows and Linux as validated execution environments.
- MIT License.
- GitHub Actions for CI/CD.

The root-level `plan.md` is the authoritative project proposal.

Previous phases established:

**Phase 0 — Foundation**
- Desktop application infrastructure.
- Frontend/backend communication.
- Configuration, logging, and build infrastructure.

**Phase 1 — Connection & Index Explorer**
- Connection Manager.
- Secure credential storage.
- Meilisearch HTTP client.
- Index Explorer.

**Phase 2 — Document Browser**
- Document retrieval.
- Dynamic document tables.
- Server-side pagination.
- JSON Inspector.

**Phase 3 — Query Console**
- Search API execution.
- JSON Query Editor.
- Search results and diagnostics.
- Query History.

**Phase 4 — Document Editing**
- Document creation and modification.
- JSON Document Editor.
- Change confirmation.
- Task Tracking.

Phase 5 improves the usability and efficiency of these existing capabilities.

Before implementation:

1. Read `plan.md`.
2. Inspect the actual implementation of Phases 0–4.
3. Review architecture decisions and previous phase reports.
4. Identify reusable components and services.
5. Preserve established architecture and API boundaries.
6. Verify current third-party APIs before using them.
7. Avoid implementing functionality already present in the repository.

**Do not replace working components merely to introduce a different UI architecture.**

---

# 2. Phase 5 Objectives

Implement five major improvements:

1. A GUI-based Search Builder.
2. Named Saved Queries.
3. JSON and CSV export.
4. Consistent keyboard shortcuts.
5. UI/UX improvements and integration validation.

Users must be able to perform common exploration workflows without repeatedly editing JSON manually, while retaining unrestricted access to the existing Query Console.

The implementation should make MeiliScope more convenient without turning it into a comprehensive administration console.

## 2.1 Design Principles

**Query Console Remains Authoritative**

Search Builder must use the existing JSON query representation and execution service.

**No Duplicate Business Logic**

Reuse the existing search, document, connection, and persistence services.

**Predictable User Interaction**

Avoid unexpected requests, silent data changes, and ambiguous UI behavior.

**Explicit Export Scope**

Users must understand exactly which data will be exported.

**Local-first**

Saved queries and exports remain local unless the user explicitly chooses otherwise.

**Performance-conscious**

Avoid unnecessary network requests, full-index loading, and unbounded in-memory collections.

**Cross-platform**

Preserve Windows and Linux functionality.

---

# 3. Architecture

Extend the existing application architecture as follows:

```text
React / TypeScript
│
├── Search Builder
│   └── Shared Query State
│       └── Query Console
│
├── Saved Queries
│   └── Existing Query Execution Service
│
├── Export UI
│   ├── Document Browser
│   └── Query Results
│
├── Keyboard Shortcut Manager
│
└── Shared UI Components
    │
    ▼
Wails v3 Bindings
    │
    ▼
Go Application Services
│
├── Connection Manager
├── Document Service
├── Query Execution Service
├── Saved Query Repository
├── Export Service
└── Existing Meilisearch HTTP Client
    │
    ▼
Meilisearch REST API
```

Do not introduce another local HTTP server.

Do not communicate directly with Meilisearch from React.

Avoid adding a database unless the existing local persistence infrastructure is demonstrably insufficient.

---

# Step 5.1 — Search Builder

## Objective

Implement a GUI-based search parameter editor integrated with the existing Query Console.

Search Builder must generate and modify Meilisearch Search API JSON without creating a second query execution system.

## Tasks

### 5.1.1 Inspect Existing Query Infrastructure

Review:

- Query Console.
- JSON Query Editor.
- Query Execution Service.
- Query validation.
- Query state management.
- Search result rendering.
- Query History.
- Connection and index selection.

Identify the authoritative query state and reuse it.

Do not create independent query execution logic for Search Builder.

### 5.1.2 Implement Basic Search Parameters

Provide GUI controls for:

- Search keyword (`q`).
- Filter (`filter`).
- Sort (`sort`).
- Limit (`limit`).
- Offset (`offset`).
- Facets (`facets`).
- Attributes to retrieve (`attributesToRetrieve`).

Support the documented data types for each parameter.

For example:

```json
{
  "q": "rendering",
  "filter": "category = 'graphics'",
  "sort": ["created_at:desc"],
  "limit": 20,
  "offset": 0,
  "facets": ["category"],
  "attributesToRetrieve": [
    "id",
    "title",
    "category"
  ]
}
```

Verify the current Meilisearch API specification.

Do not assume every parameter is available in every supported server version.

### 5.1.3 Implement Filter Input

Initially, support a text-based filter expression editor.

Users should be able to enter expressions such as:

```text
category = 'graphics'
```

And:

```text
category = 'graphics' AND published = true
```

Preserve the full expression.

Do not implement a complete visual filter expression parser unless necessary.

Do not silently rewrite advanced filter expressions.

Meilisearch also supports structured filter expressions. Preserve existing array-based filter values when they cannot be represented safely by the simple text input.

### 5.1.4 Implement Sort Controls

Support multiple sort expressions.

Example:

```json
{
  "sort": [
    "created_at:desc",
    "title:asc"
  ]
}
```

Allow:

- Adding a sort expression.
- Removing an expression.
- Changing sort direction.
- Reordering expressions.

Preserve the specified order.

If supported index settings are available, they may be used to suggest sortable fields.

However:

- Do not require administrative permissions.
- Do not require index settings to be readable.
- Do not silently reject fields solely because metadata is unavailable.

Meilisearch remains responsible for semantic validation.

### 5.1.5 Implement Facet Controls

Support selecting or entering facet names.

Allow multiple facets.

If index metadata is available, it may be used for suggestions.

Do not require a complete field schema.

Do not implement index settings modification.

### 5.1.6 Preserve Unknown Query Parameters

This requirement is mandatory.

The existing Query Console accepts arbitrary Search API JSON.

Search Builder must not discard parameters it does not explicitly support.

For example:

```json
{
  "q": "rendering",
  "showRankingScore": true,
  "showRankingScoreDetails": true,
  "customFutureParameter": {
    "enabled": true
  }
}
```

If a user changes `q` through Search Builder, all unrelated properties must remain intact.

Do not reconstruct the entire query using only known GUI fields.

Prefer targeted updates to the parsed JSON object.

Preserve numeric precision and JSON types.

### 5.1.7 Handle Unsupported Representations

Some valid query structures may not map cleanly to simple GUI controls.

Examples:

- Array-based filter expressions.
- Complex nested parameters.
- Unknown future parameters.
- Conflicting pagination options.

When Search Builder cannot safely represent a value:

- Preserve the original JSON.
- Display an appropriate indication.
- Allow editing through Query Console.
- Do not silently replace the value.
- Do not silently reset the query.

If the current JSON is syntactically invalid, Search Builder must not overwrite it.

Offer a clear validation message and retain the editor contents.

### 5.1.8 Synchronize Builder and Query Console

Search Builder and Query Console must operate on the same logical query document.

Requirements:

- Builder changes update the JSON query.
- Supported JSON changes update Builder controls.
- Unknown fields are preserved.
- Switching views does not reset values.
- No automatic query execution occurs during editing.
- Explicit execution uses the current query snapshot.
- Pagination changes remain consistent with the JSON.

Avoid bidirectional synchronization loops.

A structured query editing service or reducer may be used to ensure predictable updates.

### 5.1.9 Handle Pagination Modes

Meilisearch supports offset-based and page-based search pagination.

If the query contains:

```json
{
  "page": 2,
  "hitsPerPage": 20
}
```

Do not silently convert it to `offset` and `limit`.

The Builder may primarily expose offset-based controls.

If page-based pagination is detected, either:

- Provide an appropriate page-based mode, or
- Preserve the existing values and clearly indicate that pagination is controlled in the JSON editor.

Do not generate contradictory pagination settings.

### 5.1.10 Integrate Execution

Search Builder must call the existing Query Execution Service through the same execution workflow as Query Console.

Reuse:

- Request validation.
- Cancellation.
- Loading states.
- Search results.
- Error diagnostics.
- Query History.

Do not create separate result storage.

### 5.1.11 Tests

Test:

- Basic query generation.
- Empty queries.
- Filters.
- Multiple sort expressions.
- Facets.
- Attributes to retrieve.
- Limit and offset.
- Unknown parameters.
- Array-based filters.
- Invalid JSON.
- Nested JSON values.
- Pagination mode preservation.
- Builder-to-JSON synchronization.
- JSON-to-Builder synchronization.
- No automatic execution.
- Connection switching.
- Query History integration.

## Acceptance Criteria

- Search Builder generates valid Search API JSON.
- Query Console and Builder remain synchronized.
- Unknown query fields are preserved.
- Complex values are not silently overwritten.
- Search execution uses the existing backend.
- Existing Query Console functionality remains unchanged.
- Automated tests pass.

---

# Step 5.2 — Saved Queries

## Objective

Implement a persistent library of named search queries.

Saved Queries must be separate from the automatic Query History implemented in Phase 3.

## Tasks

### 5.2.1 Distinguish Saved Queries from Query History

**Query History**

Automatically records previous executions according to existing history settings.

**Saved Queries**

Contains explicitly named queries that users intentionally preserve.

Do not implement Saved Queries as a renamed history list.

Saved Queries must not be removed by automatic history retention.

### 5.2.2 Define Saved Query Models

Suggested fields:

```text
Saved Query ID
Display Name
Description (optional)
Connection ID
Index UID
Query JSON
Created At
Updated At
```

Use stable identifiers.

Do not store:

- API Keys.
- Authentication headers.
- Full search results.
- Credential secrets.

### 5.2.3 Implement Local Persistence

Reuse the existing local configuration and persistence infrastructure.

Support:

- Create.
- Read.
- Update.
- Delete.
- List.
- Load.
- Duplicate.

Preserve saved queries across application restarts.

Do not introduce cloud synchronization.

Do not require a database for a small collection of query definitions.

### 5.2.4 Implement Save Workflow

Allow saving the current Query Console JSON with:

- Name.
- Optional description.
- Target connection.
- Target index.

Require a non-empty display name.

Allow duplicate names if queries have different identifiers, or enforce a documented uniqueness policy.

Do not silently overwrite an existing query.

### 5.2.5 Implement Saved Query UI

Provide a panel or dialog containing:

- Query name.
- Associated index.
- Associated connection.
- Optional description.
- Last modified timestamp.

Support:

- Save current query.
- Load saved query.
- Rename.
- Update saved definition.
- Duplicate.
- Delete.
- Execute.

Avoid excessive modal dialogs.

### 5.2.6 Handle Connection and Index Associations

A saved query may reference a connection that has been deleted or renamed.

Requirements:

- Associate queries with stable connection IDs.
- Preserve the original index UID.
- Handle missing connections.
- Handle inaccessible indexes.
- Never silently run a saved query against another server.
- Require explicit target selection if reassignment is necessary.

Changing a saved query's target must be an intentional action.

### 5.2.7 Implement Save vs. Save As

Distinguish:

**Save**

Update the currently selected saved query.

**Save As**

Create another saved query.

Clearly indicate when the editor contains unsaved changes.

Do not modify a saved query simply because its loaded JSON was edited.

### 5.2.8 Preserve Query JSON

Saved Queries must preserve arbitrary valid Search API JSON.

Do not serialize queries through a restricted Search Builder model.

Preserve:

- Unknown parameters.
- Nested objects.
- Arrays.
- Numeric precision.
- Unicode.

Do not alter query semantics during save and reload.

### 5.2.9 Implement Execution

Execute saved queries through the existing Query Execution Service.

Reuse:

- Connection resolution.
- Authentication.
- Search result presentation.
- Query diagnostics.
- Query History.

Do not create a separate API execution path.

### 5.2.10 Implement Persistence Safety

Use safe persistence practices.

Support:

- Atomic updates where appropriate.
- Corrupted-data detection.
- Stable schema versioning.
- Safe handling of missing fields.
- Migration when formats change.

Do not silently discard an entire saved query collection because one entry is malformed.

Preserve recoverable user data where practical.

### 5.2.11 Privacy and Security

Saved queries may contain sensitive business information or search terms.

Therefore:

- Keep saved queries local.
- Do not automatically upload saved queries.
- Do not log full query contents by default.
- Do not store credentials.
- Document the persistence location and privacy implications.

Saved Queries are not required to be encrypted unless the existing architecture explicitly provides this capability.

Do not claim encryption without implementing and verifying it.

### 5.2.12 Tests

Test:

- Create.
- Rename.
- Update.
- Duplicate.
- Delete.
- Save As.
- Persistence across restarts.
- Query JSON fidelity.
- Unknown parameters.
- Missing connection.
- Deleted index.
- Invalid saved data.
- Corrupted persistence.
- Schema version handling.
- Integration with Query Console.
- Integration with Query History.
- No credential leakage.

## Acceptance Criteria

- Users can explicitly save named queries.
- Queries survive application restarts.
- Saved Queries remain separate from Query History.
- Queries can be loaded, modified, duplicated, and rerun.
- Missing connections are handled safely.
- Arbitrary search JSON is preserved.
- Credentials are never persisted in saved query records.

---

# Step 5.3 — JSON / CSV Export

## Objective

Implement reliable export of document data and search results.

Support JSON and CSV formats.

Export must not modify Meilisearch data.

## Tasks

### 5.3.1 Define Export Sources

Support:

**Document Browser**

Export documents from the current index.

**Query Results**

Export the current search results.

Do not create an independent query execution engine for export.

Reuse existing services.

### 5.3.2 Define Explicit Export Scope

Export scope must be unambiguous.

At minimum, support:

- Current loaded page.
- Current search result page.

A selected-document export may also be supported where the existing selection model makes it straightforward.

Do not label a current-page export as a complete index export.

Do not implicitly fetch every document in an index.

### 5.3.3 Optional Full-result Export

Full-index or all-matching-result export is not a mandatory Phase 5 feature.

If implemented, it must use a separate bounded and cancelable workflow.

Requirements:

- Explicit user selection.
- Documented server pagination limitations.
- Incremental retrieval.
- Incremental file writing.
- Progress reporting.
- Cancellation.
- Export size safeguards.
- No unbounded accumulation in memory.
- Clear handling of incomplete exports.

Do not claim complete export if the selected API cannot guarantee access to the complete matching dataset.

Avoid using estimated search totals as a guarantee of complete export coverage.

### 5.3.4 Implement Export Service

Use a Go backend service for file writing.

Suggested abstraction:

```go
type ExportService interface {
    ExportJSON(
        ctx context.Context,
        request ExportRequest,
    ) (ExportResult, error)

    ExportCSV(
        ctx context.Context,
        request ExportRequest,
    ) (ExportResult, error)
}
```

Adapt to existing architecture.

A secure transfer mechanism between frontend result state and backend file writing must be defined.

Avoid returning enormous serialized datasets solely for export if the backend can safely retrieve or stream the required data.

For current-page export, reuse an existing immutable result snapshot where practical.

Do not silently execute a different query from the one currently displayed.

### 5.3.5 Implement JSON Export

Support UTF-8 JSON files.

For Document Browser, the preferred export format is an array of document objects.

Example:

```json
[
  {
    "id": 1,
    "title": "Example"
  }
]
```

For Query Results, distinguish:

- Export hits only.
- Export the complete response envelope.

The complete-response option should preserve metadata such as facets, pagination, and ranking information.

Clearly label each export format.

Preserve arbitrary JSON fields and numeric precision.

Do not silently remove nested structures or metadata.

### 5.3.6 Implement CSV Export

Support tabular export of document rows and search hits.

Use Go's standard `encoding/csv` package or an equivalent well-tested implementation.

Requirements:

- UTF-8 encoding.
- Consistent headers.
- Correct quoting.
- Correct escaping.
- Embedded commas.
- Embedded newlines.
- Double quotes.
- Unicode and Japanese text.
- Empty strings.
- Null values.
- Missing fields.
- Numeric values.
- Boolean values.

Avoid manually constructing CSV rows through string concatenation.

### 5.3.7 Handle Dynamic Fields

Documents may contain different field sets.

Generate CSV columns from the union of relevant top-level fields within the selected export dataset.

Ensure deterministic column order.

Keep the primary-key column first when available.

Do not assume the first document contains all possible fields.

For current-page exports, the discovered field set is limited to that page.

Do not present those columns as the complete index schema.

### 5.3.8 Handle Nested JSON Values

CSV cannot natively represent nested JSON types.

For object and array values, serialize the value as compact JSON inside the CSV cell.

Example:

```json
{
  "id": 1,
  "tags": ["go", "search"]
}
```

The `tags` CSV cell should contain a valid JSON representation of the array.

Do not flatten nested objects unpredictably.

Preserve the distinction between:

- Missing field.
- Explicit null.
- Empty string.
- Empty array.
- Empty object.

Document the chosen CSV representation.

### 5.3.9 Prevent Spreadsheet Formula Injection

CSV files may be opened with spreadsheet software.

Cells beginning with characters interpreted as formulas can present a security risk.

Implement an explicit, documented CSV safety policy.

Consider spreadsheet-sensitive prefixes such as:

```text
=
+
-
@
```

Also consider leading whitespace and control characters that may affect spreadsheet interpretation.

Do not silently corrupt ordinary data to achieve spreadsheet compatibility.

If values are transformed for spreadsheet safety:

- Document the transformation.
- Make the behavior predictable.
- Preserve raw data through JSON export.
- Test the resulting output with relevant spreadsheet applications where practical.

### 5.3.10 Implement File Save Dialog

Reuse Wails' supported native file dialog integration.

Support:

- JSON and CSV file extensions.
- Default filenames.
- Explicit destination selection.
- File overwrite confirmation.
- Cancellation.
- File-system error reporting.

Suggested filenames:

```text
documents_<index>_<timestamp>.json
documents_<index>_<timestamp>.csv
search_results_<index>_<timestamp>.json
search_results_<index>_<timestamp>.csv
```

Sanitize filenames.

Do not use unsanitized index names as file paths.

### 5.3.11 Implement Atomic File Writing

Where practical:

1. Write to a temporary file in the destination directory.
2. Flush and close the file.
3. Validate completion.
4. Replace or move the completed file to the requested destination.

Avoid leaving a partially written file with the appearance of a successful export.

If an export is canceled or fails, report the result accurately and clean up temporary files.

Do not delete unrelated existing files.

### 5.3.12 Export UI

Provide an Export action in:

- Document Browser.
- Query Results.

The export dialog should clearly display:

```text
Source
Connection
Index
Format
Export Scope
Expected Number of Rows, if known
Destination
```

For search results, distinguish the number of returned hits from estimated or exact total hits.

Require explicit confirmation for potentially expensive exports.

Do not automatically export data without user action.

### 5.3.13 Tests

Test:

- JSON document export.
- Full search response JSON.
- CSV export.
- Dynamic fields.
- Nested objects.
- Arrays.
- Unicode.
- Large integers.
- Quotes and commas.
- Embedded newlines.
- Null and missing fields.
- Formula-like values.
- Empty result sets.
- Invalid destination.
- File permission denied.
- User cancellation.
- Disk write failure.
- Existing file overwrite.
- Partial export cleanup.
- Export scope correctness.
- Connection switching during export.

## Acceptance Criteria

- Document Browser can export the current page.
- Query Results can export the current result set.
- JSON and CSV formats are supported.
- Export scope is explicit.
- Dynamic fields are handled correctly.
- JSON fidelity is preserved.
- CSV safety behavior is documented.
- File writes are safe and cancelable where practical.
- No full-index export occurs implicitly.
- Existing data is never modified by export.

---

# Step 5.4 — Keyboard Shortcuts

## Objective

Implement consistent keyboard shortcuts across the application.

## Tasks

### 5.4.1 Audit Existing Shortcuts

Inspect shortcuts from Phases 0–4.

Identify:

- Existing definitions.
- Conflicts.
- Duplicate handlers.
- Platform-specific behavior.
- Shortcuts already managed by Monaco Editor or other components.

Do not replace standard editor shortcuts unnecessarily.

### 5.4.2 Implement a Central Shortcut Registry

Create a lightweight centralized mechanism for application-level shortcuts.

Each shortcut should define:

```text
Action ID
Key Combination
Scope
Description
Enabled State
Handler
```

Do not introduce a heavy command framework unless justified.

Avoid multiple unrelated global keyboard listeners.

### 5.4.3 Define Default Shortcuts

Recommended shortcuts:

| Action | Windows / Linux |
|---|---|
| Execute Query | Ctrl+Enter |
| Refresh Current View | Ctrl+R |
| Open Connections | Ctrl+Shift+C |
| Open Index Explorer | Ctrl+1 |
| Open Document Browser | Ctrl+2 |
| Open Query Console | Ctrl+3 |
| Open Search Builder | Ctrl+4 |
| Save Query | Ctrl+S |
| Format JSON | Shift+Alt+F |
| Focus Search/Query Editor | Ctrl+L, if conflict-free |
| Copy | Ctrl+C |
| Escape Dialog | Esc |

These are proposed defaults, not mandatory assignments.

Inspect the actual Wails WebView behavior before finalizing shortcuts.

Some shortcuts may conflict with browser, system, or editor functions.

Choose conflict-free alternatives where needed.

### 5.4.4 Implement Scoped Shortcuts

Shortcuts must respect UI context.

Examples:

- Ctrl+S in Query Console may save a query.
- Ctrl+S in Document Editor may initiate the existing safe save workflow.
- Ctrl+Enter in Query Console executes a search.
- Ctrl+Enter in Document Editor must not accidentally submit a mutation.
- Escape may close a dialog without discarding unsaved edits silently.

Use clear scope priorities.

Do not trigger multiple actions from a single key combination.

### 5.4.5 Respect Focus and Editing State

Global shortcuts must not interfere with:

- Text inputs.
- JSON editors.
- Search fields.
- Dialogs.
- Confirmation workflows.
- Standard copy/paste.
- Undo/redo.
- Text selection.

Prevent accidental destructive actions.

Never bypass mutation confirmation.

### 5.4.6 Cross-platform Handling

Verify shortcut behavior on Windows and Linux.

Do not assume all window managers or desktop environments reserve the same keys.

Prefer application-level shortcuts over global operating-system hotkeys.

Do not introduce keyboard hooks requiring elevated permissions.

### 5.4.7 Discoverability

Display shortcuts in appropriate places:

- Tooltips.
- Menus.
- Action labels.
- Keyboard shortcuts help dialog.

A simple shortcut reference dialog is sufficient.

Customizable keybindings are optional and not required in Phase 5.

### 5.4.8 Tests

Test:

- Shortcut dispatch.
- Shortcut scope.
- Editor focus.
- Dialog behavior.
- Duplicate event prevention.
- Query execution.
- Refresh.
- Navigation.
- Save Query.
- Document Editor safety.
- Platform-specific modifier handling.

## Acceptance Criteria

- Common workflows are accessible through keyboard shortcuts.
- Shortcuts are centralized and documented.
- Context-specific commands behave correctly.
- Standard editor behavior remains intact.
- Destructive operations retain confirmation requirements.
- Windows and Linux shortcuts work in tested environments.

---

# Step 5.5 — UI/UX Polish, Integration & Validation

## Objective

Improve usability, consistency, responsiveness, and reliability across the existing application.

Complete Phase 5 without unnecessarily expanding the product scope.

## Tasks

### 5.5.1 Review Existing UI

Review:

- Connection Manager.
- Index Explorer.
- Document Browser.
- JSON Inspector.
- Query Console.
- Search Builder.
- Saved Queries.
- Document Editor.
- Task Tracker.
- Export workflows.

Identify inconsistent components and interaction patterns.

Prefer improving shared components over implementing isolated fixes.

### 5.5.2 Standardize Layout

Review:

- Main toolbar.
- Side navigation.
- Content panels.
- Split panes.
- Status bar.
- Dialogs.
- Resizable areas.
- Window resizing.
- Minimum window size.

Ensure that major workflows remain usable on ordinary desktop display sizes.

Do not require excessive horizontal or vertical scrolling for common operations.

Preserve user-adjusted panel sizes where appropriate.

### 5.5.3 Standardize Loading States

Use consistent loading indicators.

Distinguish:

- Initial loading.
- Refreshing existing data.
- Query execution.
- Document fetching.
- Task monitoring.
- File export.

Avoid blanking useful existing content unnecessarily during refresh.

However, never display data from a previous connection as if it belonged to the current connection.

### 5.5.4 Standardize Empty States

Provide useful empty states for:

- No connections configured.
- No connection selected.
- No indexes available.
- No index selected.
- No documents.
- Empty search results.
- No query history.
- No saved queries.
- No selected document.
- No tracked task.

Where appropriate, provide a clear next action.

Avoid displaying technical errors for ordinary empty states.

### 5.5.5 Improve Error Presentation

Review error handling across the application.

Distinguish:

- Validation errors.
- Network errors.
- Authentication errors.
- Authorization errors.
- Server errors.
- Task failures.
- File-system errors.
- Canceled operations.
- Unknown outcomes.

Provide:

- Concise user-facing messages.
- Optional technical details.
- Relevant recovery actions.

Do not expose credentials or sensitive headers.

Avoid showing duplicate notifications for the same failure.

### 5.5.6 Improve Confirmation Workflows

Review existing confirmation dialogs.

Ensure that:

- Destructive actions are clearly identified.
- Mutation targets are visible.
- Cancellation is safe.
- Keyboard shortcuts cannot bypass confirmation.
- Dialog focus behavior is predictable.
- Unsaved changes are protected.

Do not increase the number of confirmations for routine read-only operations unnecessarily.

### 5.5.7 Improve Connection and Index Switching

Verify consistent behavior when switching:

- Connections.
- Indexes.
- Views.
- Saved queries.
- Document selections.

Ensure that:

- Requests are canceled or invalidated appropriately.
- Stale responses are ignored.
- Cached data is scoped correctly.
- Editing operations retain their original targets.
- Active exports cannot silently change targets.
- Task references remain associated with their original connections.

### 5.5.8 Improve Performance

Review common performance concerns:

- Unnecessary React re-renders.
- Large JSON Inspector trees.
- Large table rendering.
- Repeated API requests.
- Excessive state synchronization.
- Query history loading.
- Saved query persistence.
- Export memory consumption.

Use profiling or measurements before introducing complex optimizations.

Maintain bounded caches.

Avoid retaining complete datasets unnecessarily.

### 5.5.9 Improve Accessibility

Address practical desktop accessibility concerns:

- Keyboard navigation.
- Visible focus indicators.
- Accessible form labels.
- Appropriate contrast.
- Predictable dialog focus.
- Accessible error messages.
- Non-color-only status indicators.

Reuse established accessibility patterns.

Do not introduce an unrelated accessibility framework unless necessary.

### 5.5.10 Validate Search Builder Integration

Test:

1. Open Query Console.
2. Enter a JSON query with advanced parameters.
3. Open Search Builder.
4. Change a supported parameter.
5. Return to Query Console.
6. Verify that unrelated parameters remain intact.
7. Execute the query.
8. Verify the result.
9. Open Query History.
10. Restore the exact submitted query.

No query parameter should be silently lost.

### 5.5.11 Validate Saved Queries

Test:

1. Create a query.
2. Save it with a name.
3. Modify the editor.
4. Reload the saved query.
5. Execute it.
6. Duplicate it.
7. Rename it.
8. Delete it.
9. Restart MeiliScope.
10. Verify the remaining saved queries.

Confirm that Query History retention does not delete Saved Queries.

### 5.5.12 Validate Export

Test:

1. Open Document Browser.
2. Load a document page.
3. Export the current page as JSON.
4. Export the same page as CSV.
5. Verify field preservation.
6. Open Query Console.
7. Execute a search.
8. Export search hits.
9. Export the complete JSON response.
10. Verify that reported export counts match the actual output scope.

Ensure exported data is not silently replaced with data from a newly selected index.

### 5.5.13 Validate Cross-platform Behavior

Run relevant functionality on Windows and Linux.

Verify:

- Application startup.
- Navigation.
- Search Builder.
- Saved Queries.
- Native file dialogs.
- Export.
- Keyboard shortcuts.
- Clipboard.
- JSON Editor.
- Document Editor.
- Task Tracker.
- Window resizing.
- Application shutdown.

Preserve existing build and CI compatibility.

### 5.5.14 Automated Tests

Extend backend tests for:

- Saved query persistence.
- JSON export.
- CSV export.
- File writing.
- Export cancellation.
- Error handling.
- Configuration migration.

Extend frontend tests for:

- Search Builder synchronization.
- Unknown parameter preservation.
- Saved query management.
- Export scope selection.
- Keyboard shortcuts.
- Dialog focus.
- Loading states.
- Error recovery.

Add integration tests for the main user workflows.

Prefer deterministic tests with fixtures and mock HTTP services.

Use real Meilisearch only where necessary for integration validation.

### 5.5.15 Security Review

Verify:

- API Keys are never included in saved queries.
- Query History remains local.
- Export does not trigger mutations.
- File paths are handled safely.
- CSV formula injection is addressed.
- No unsanitized HTML is rendered from Meilisearch data.
- No arbitrary API endpoint execution is introduced.
- Destructive actions retain confirmation.
- Connection-specific data remains isolated.
- Temporary export files are cleaned up safely.

### 5.5.16 Update Documentation

Update documentation using the existing repository conventions.

Recommended files:

```text
README.md

docs/en/user-guide/search-builder.md
docs/ja/user-guide/search-builder.md

docs/en/user-guide/saved-queries.md
docs/ja/user-guide/saved-queries.md

docs/en/user-guide/export.md
docs/ja/user-guide/export.md

docs/en/user-guide/keyboard-shortcuts.md
docs/ja/user-guide/keyboard-shortcuts.md

docs/architecture/overview.md
docs/development/phase-5-report.md
```

Document:

- Search Builder usage.
- Builder/JSON synchronization.
- Saved Queries.
- Query History differences.
- Export formats.
- Export scope limitations.
- CSV representation.
- CSV security considerations.
- Keyboard shortcuts.
- Error recovery.
- Cross-platform limitations.

Update `plan.md` only when actual implementation decisions materially change the proposal, roadmap, or supported-platform policy.

Do not rewrite unrelated sections.

### 5.5.17 Phase Completion Report

Create:

`docs/development/phase-5-report.md`

Include:

1. Completed implementation steps.
2. Files created or modified.
3. Search Builder design.
4. Query synchronization decisions.
5. Saved Query persistence.
6. Export implementation.
7. CSV representation and security.
8. Keyboard shortcut design.
9. UI/UX improvements.
10. Backend test results.
11. Frontend test results.
12. Windows validation.
13. Linux validation.
14. Performance observations.
15. Security review.
16. Known limitations.
17. Phase 6 readiness.

## Acceptance Criteria

- Search Builder works with Query Console.
- Unknown query parameters are preserved.
- Saved Queries can be managed independently of Query History.
- JSON and CSV export work.
- Export scope is always explicit.
- Keyboard shortcuts work without conflicting with editing operations.
- Loading, error, and empty states are consistent.
- Existing browsing, searching, and editing functionality remains intact.
- Windows and Linux builds remain functional.
- Automated tests pass.
- Documentation reflects the actual implementation.

---

# 4. General Implementation Constraints

## 4.1 Preserve Product Scope

MeiliScope remains a lightweight Meilisearch data browser.

Do not turn it into a general Meilisearch administration console.

## 4.2 Do Not Implement Server Administration

Do not add:

- API Key administration.
- Index creation or deletion.
- Server configuration management.
- Server monitoring.
- Backup and restoration.
- Cluster administration.
- Arbitrary HTTP endpoint execution.

These are outside the current Phase 5 scope.

## 4.3 Preserve JSON Query Transparency

Search Builder must not restrict Query Console functionality.

Unknown parameters must survive ordinary editing operations.

Do not introduce a hardcoded allowlist into the underlying Query Execution Service.

## 4.4 Avoid Duplicate Services

Reuse existing:

- HTTP Client.
- Query Execution Service.
- Document Service.
- Configuration Manager.
- Credential Store.
- JSON Editor.
- Document Table.
- JSON Inspector.

Refactor only when necessary for maintainability.

## 4.5 Avoid Unbounded Data Retrieval

Do not fetch an entire index automatically for export, filtering, or field discovery.

Keep data retrieval bounded.

Do not introduce background synchronization.

## 4.6 Preserve Existing Security Boundaries

Credentials remain backend-owned.

Search and browsing remain read-only.

Mutations remain explicit and confirmed.

Exports are local user-initiated operations.

## 4.7 Maintain Cross-platform Compatibility

Do not introduce Windows-only assumptions into shared code.

Preserve Windows and Linux builds.

Use supported Wails APIs for native dialogs and platform integration.

## 4.8 Keep Dependencies Minimal

Avoid introducing:

- A second JSON editor.
- Another table framework.
- A redundant query state manager.
- An unnecessary local database.
- An unnecessary local HTTP server.
- Heavy UI frameworks without clear benefits.

Record important dependency decisions.

---

# 5. Execution & Reporting Requirements

Execute the implementation steps in order:

1. Step 5.1 — Search Builder.
2. Step 5.2 — Saved Queries.
3. Step 5.3 — JSON / CSV Export.
4. Step 5.4 — Keyboard Shortcuts.
5. Step 5.5 — UI/UX Polish, Integration & Validation.

Each step must be independently reviewable.

After each step, report:

### Completed Work

Describe the implemented functionality.

### Files Changed

List significant source and documentation changes.

### Tests

Provide the actual commands executed and their results.

### Technical Decisions

Explain significant implementation choices.

### Known Issues

Identify unresolved issues and deferred work.

### Next Step Readiness

State whether the next step can proceed safely.

Do not claim tests passed unless they were executed.

Do not hide test failures or incomplete integration.

Do not implement unrelated functionality from later phases.

---

# 6. Phase 5 Definition of Done

Phase 5 is complete when a user can:

1. Launch MeiliScope on Windows or Linux.
2. Connect to Meilisearch.
3. Browse indexes and documents.
4. Inspect document JSON.
5. Execute arbitrary JSON search queries.
6. Construct common search queries through Search Builder.
7. Move between Search Builder and Query Console without losing unsupported parameters.
8. Save a query with a name.
9. Reload and rerun saved queries.
10. Manage Saved Queries separately from Query History.
11. Export current document pages to JSON or CSV.
12. Export search results to JSON or CSV.
13. Clearly identify the scope of exported data.
14. Use keyboard shortcuts for common operations.
15. Continue using the existing controlled document editing workflow.
16. Understand loading states and recoverable errors.
17. Switch connections without stale-data leakage.
18. Restart the application without losing saved queries or application preferences.
19. Use the application without installing development toolchains.
20. Complete normal workflows without major UI inconsistencies.

All major functionality must be covered by automated tests or documented integration validation.

**Successful completion of Phase 5 establishes the complete planned usability feature set and prepares MeiliScope for Phase 6 — Quality & Security.**
