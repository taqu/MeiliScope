# MeiliScope — Phase 4 Implementation Instructions

## Document Editing & Task Tracking

### 1. Project Context

MeiliScope is an open-source, standalone desktop application for exploring Meilisearch data.

**Mission: Explore your Meilisearch data.**

The established technology stack includes:

- Go for backend services.
- Wails v3 for desktop integration.
- React and TypeScript for the frontend.
- Meilisearch HTTP API for data access.
- Windows and Linux as validated execution platforms.
- MIT License.

The project proposal is maintained in the root-level `plan.md`.

Previous phases established:

**Phase 0 — Foundation**
- Desktop application infrastructure.
- Go/Frontend communication.
- Configuration and logging.
- Cross-platform builds and CI.

**Phase 1 — Connection & Index Explorer**
- Connection Manager.
- Secure credential storage.
- Meilisearch HTTP client.
- Index Explorer.

**Phase 2 — Document Browser**
- Document retrieval.
- Dynamic document tables.
- Pagination.
- JSON Inspector.

**Phase 3 — Query Console**
- Search API execution.
- JSON Query Editor.
- Search results and diagnostics.
- Query history.

Phase 4 introduces controlled document modification while preserving the application's read-only-first design.

Before implementation:

1. Read `plan.md`.
2. Review the existing architecture and Phase 0–3 reports.
3. Inspect the actual implementation.
4. Reuse established HTTP, credential, JSON, and frontend services.
5. Verify the supported Meilisearch API versions.
6. Identify existing editor, dialog, notification, and task-related infrastructure.

Do not rewrite working components without a documented reason.

---

# 2. Phase 4 Objectives

Implement safe document modification workflows.

Users must be able to:

1. Create a document using JSON.
2. Edit an existing document.
3. Delete an individual document.
4. Review changes before submission.
5. Confirm the target connection and index.
6. Submit write operations explicitly.
7. Track asynchronous Meilisearch tasks.
8. Inspect task errors.
9. Refresh affected documents after successful completion.
10. Recover from failures without losing unsaved edits.

## 2.1 Design Principles

**Read-only by Default**

Browsing and inspecting documents must never trigger modifications.

**Explicit User Intent**

A write operation must require an explicit user action.

**Precise Update Semantics**

Partial updates and complete replacements must not be confused.

**Asynchronous Completion**

An accepted request does not mean indexing has completed.

**No Silent Data Loss**

Preserve unsaved changes and handle concurrent modifications conservatively.

**Least Privilege**

Use the active connection's existing credentials. Never request administrative credentials automatically.

**Cross-platform Compatibility**

Maintain Windows and Linux functionality.

---

# 3. Architecture

Extend the existing architecture without introducing a separate application server.

```text
React / TypeScript
│
├── Document Browser
│   └── Edit / Delete actions
│
├── Document Editor
│   ├── JSON Editor
│   ├── Validation
│   ├── Diff Viewer
│   └── Confirmation
│
├── Task Status Panel
│
▼
Wails v3 Bindings
│
▼
Go Application Services
│
├── Connection Manager
├── Document Read Service
├── Document Write Service
├── Task Tracking Service
└── Existing Meilisearch HTTP Client
    │
    ▼
Meilisearch HTTP API
```

The frontend must not communicate directly with Meilisearch.

Go owns authentication, API execution, task retrieval, and transport errors.

React owns user interaction, editing state, diff presentation, and confirmation.

Keep document read operations available independently of write permissions.

Do not embed mutation behavior inside the existing read-only JSON Inspector.

---

# Step 4.1 — Document Write API

## Objective

Implement controlled document creation, update, replacement, and deletion through the existing Go HTTP client.

## Tasks

### 4.1.1 Verify Current API Specifications

Consult the official Meilisearch API documentation.

Verify the supported behavior of:

```http
POST   /indexes/{index_uid}/documents
PUT    /indexes/{index_uid}/documents
DELETE /indexes/{index_uid}/documents/{document_id}
```

Also verify:

```http
GET /tasks/{task_uid}
```

Record:

- Request and response formats.
- Required API Key actions.
- Primary-key requirements.
- Partial-update semantics.
- Complete-replacement semantics.
- Task response structure.
- Index creation side effects.
- Relevant API version differences.

Do not assume that PUT and POST have equivalent behavior.

### 4.1.2 Define Explicit Write Operations

Implement a service with separate operations equivalent to:

```go
type DocumentWriteService interface {
    AddDocument(
        ctx context.Context,
        target WriteTarget,
        document json.RawMessage,
    ) (TaskReference, error)

    UpdateDocument(
        ctx context.Context,
        target WriteTarget,
        document json.RawMessage,
    ) (TaskReference, error)

    ReplaceDocument(
        ctx context.Context,
        target WriteTarget,
        document json.RawMessage,
    ) (TaskReference, error)

    DeleteDocument(
        ctx context.Context,
        target WriteTarget,
        documentID string,
    ) (TaskReference, error)
}
```

Adapt exact signatures to the codebase.

The operations must have clearly documented semantics.

Do not expose a generic arbitrary HTTP mutation method through Wails.

For the initial GUI, use a conservative update workflow.

The full-replacement operation may remain a separate advanced operation if the distinction can be communicated safely.

### 4.1.3 Handle Document Creation

Support creation of one document at a time.

Requirements:

- Accept a JSON object.
- Reject invalid JSON.
- Reject arrays and non-object top-level values.
- Validate the presence of the required primary-key field when the index already defines one.
- Preserve arbitrary JSON fields.
- Preserve numeric precision.
- Preserve Unicode.
- Avoid automatic index creation.

Do not rely on implicit primary-key inference when it could unexpectedly create or alter an index.

If the target index does not exist or has no established primary key, explain the limitation instead of silently creating or configuring it.

Distinguish a document-addition workflow from a strictly enforced create-only operation.

Meilisearch document ingestion may update an existing document with the same primary key.

Therefore, do not claim that the API guarantees create-only behavior unless the supported server version actually provides such a guarantee.

### 4.1.4 Handle Document Updates

Support updating an existing document.

Before editing:

- Retrieve the complete original document.
- Identify its primary key.
- Preserve its original JSON snapshot.
- Record its connection and index identity.

An update must not rely on a partial table row or a formatted search hit as the original document.

Use the correct API semantics for the selected operation.

For partial updates:

- Preserve fields that are not included in the update payload.
- Handle explicit null values correctly.
- Do not interpret omission as deletion unless the API explicitly supports that behavior.

For complete replacement:

- Clearly explain that omitted fields may be removed.
- Require explicit confirmation of replacement semantics.
- Never substitute replacement for partial update silently.

The editor must not present unsupported field deletion semantics.

If field removal cannot be expressed safely with the chosen operation, show the limitation and offer only a supported, explicitly confirmed workflow.

### 4.1.5 Preserve Primary-key Identity

The existing document's primary-key value identifies the update target.

Requirements:

- Handle string and numeric identifiers.
- Preserve numeric precision.
- Encode document IDs correctly in URL paths.
- Reject missing required identifiers.
- Prevent accidental changes to the primary key during ordinary editing.

Treat primary-key modification as a separate unsupported operation during Phase 4.

Do not silently create another document when an attempted edit changes the identifier.

### 4.1.6 Implement Document Deletion

Support deleting one explicitly selected document.

Requirements:

- Use the selected document's actual identifier.
- Use the selected connection and index.
- Require explicit confirmation.
- Reject empty or invalid document identifiers.
- Return the resulting task reference.
- Do not delete an entire index.
- Do not introduce bulk or filter-based deletion.

Do not remove the document permanently from the local view merely because the deletion request was accepted.

Wait for confirmed task completion before presenting deletion as successful.

### 4.1.7 Enforce Operation Scope

A write request must be bound to an immutable target snapshot:

```text
Connection ID
Index UID
Document ID, where applicable
Operation Type
```

Once submission begins, a connection or index change must not redirect the operation to a different target.

Validate the target again immediately before dispatch.

The backend must not trust frontend-provided credentials.

Credentials must be resolved internally through Connection Manager.

### 4.1.8 Handle Permission Errors

Write operations may require permissions distinct from read operations.

Handle:

- Valid read-only credentials.
- Missing write permission.
- Restricted index permissions.
- Invalid credentials.
- Revoked credentials.
- Missing index.
- Server rejection.

Do not require an Admin API Key.

Do not claim that a document is editable merely because it can be read.

Avoid inferring write permission from unrelated API calls.

### 4.1.9 Avoid Unsafe Retries

Do not automatically retry document mutations after ambiguous network failures.

A timeout or connection reset may occur after the server accepted the operation.

Classify these outcomes as potentially submitted when appropriate.

Inform the user that the final state is unknown.

Do not report a failed mutation as definitely not applied unless that conclusion is supported by the response.

### 4.1.10 Implement Backend Tests

Use `httptest.Server`.

Test:

- Document creation request.
- Partial-update request.
- Replacement request.
- Individual deletion.
- Invalid JSON.
- Invalid primary key.
- Numeric and string identifiers.
- Nested JSON.
- Unicode.
- Large JSON numbers.
- Unsupported operation.
- Incorrect HTTP method.
- HTTP 401/403/404/409/429/500.
- Timeout.
- Context cancellation.
- Task response decoding.
- Missing task UID.
- Ambiguous submission failure.
- Credential sanitization.

## Acceptance Criteria

- Document write operations use the existing HTTP client.
- Add, update, replace, and delete semantics are explicit.
- Document identifiers remain accurate.
- No administrative API is required.
- No implicit index creation is performed.
- Mutations are not automatically retried after uncertain submission.
- All backend tests pass.
- Wails does not expose arbitrary write endpoints.

---

# Step 4.2 — JSON Document Editor

## Objective

Implement a safe JSON editing interface with change review and explicit confirmation.

## Tasks

### 4.2.1 Reuse the Existing Editor Infrastructure

Inspect the JSON editor implemented in Phase 3.

Reuse its:

- Syntax highlighting.
- Formatting.
- Validation.
- Keyboard handling.
- Error markers.
- Editor integration.

Do not introduce another large editor framework.

Keep Query Console and Document Editor state separate.

### 4.2.2 Define Editor Modes

Support:

**Create Mode**

Create a document from a new JSON object.

**Edit Mode**

Modify an existing document obtained through Document Service.

**Delete Confirmation**

Confirm deletion of an existing document.

Do not implement inline editing inside Document Table during this phase.

### 4.2.3 Implement Editing Workflow

For existing documents:

1. User selects a document.
2. User explicitly chooses Edit.
3. Retrieve the full document if necessary.
4. Create an immutable original snapshot.
5. Open the JSON editor.
6. User modifies the document.
7. Validate JSON syntax and primary-key identity.
8. Present a diff.
9. Confirm the operation.
10. Submit through Document Write Service.

Do not modify the original snapshot while editing.

### 4.2.4 Implement JSON Validation

Validate:

- Correct JSON syntax.
- Object top-level structure.
- Required primary-key field.
- Primary-key identity.
- Supported document size.
- Valid operation semantics.

Preserve arbitrary user-defined fields.

Do not enforce a fixed document schema.

Avoid losing precision when parsing or serializing large numbers.

Do not convert missing fields into null values automatically.

### 4.2.5 Implement Diff Viewer

Display differences between the original and edited JSON.

Support:

- Added fields.
- Changed values.
- Removed fields.
- Nested changes.
- Array changes.
- Null changes.

Use a readable side-by-side or unified presentation.

Prefer structural comparison where practical.

Do not treat object key reordering as a semantic change.

Preserve array order.

Make the diff accurately reflect the effective operation.

For partial updates, distinguish omitted properties from explicit values.

For complete replacement, show the possible removal of omitted fields.

The diff viewer must not claim that a field is deleted when the selected API operation would preserve it.

### 4.2.6 Implement Save Confirmation

Before submitting a mutation, display:

```text
Connection
Server URL
Index UID
Document ID
Operation Type
Change Summary
```

Use the validated submission snapshot.

Do not allow the confirmation dialog to change the operation target.

Require explicit approval.

Do not submit automatically when the user closes the editor.

For destructive or replacement operations, use more prominent confirmation wording.

### 4.2.7 Implement Delete Confirmation

Deletion must show:

- Connection.
- Index.
- Document ID.
- Irreversible-operation warning.
- Explicit Delete action.
- Cancel action.

Use a visually distinct destructive action.

Do not allow accidental deletion through ordinary row selection or keyboard navigation.

### 4.2.8 Handle Unsaved Changes

Detect whether the document has been modified.

If the user attempts to:

- Close the editor.
- Switch documents.
- Switch indexes.
- Switch connections.
- Navigate away.

Offer an appropriate choice to discard changes or remain in the editor.

Do not silently discard unsaved changes.

Keep editor state in memory unless the user explicitly chooses to submit the changes.

Do not persist arbitrary unsaved document contents automatically in configuration files.

### 4.2.9 Handle Concurrent Modifications

Meilisearch documents may be modified by other clients while the editor is open.

Before submission, re-fetch the current document when possible.

Compare it with the original snapshot.

If the server document changed:

- Display a conflict warning.
- Show the original, remote, and edited versions where practical.
- Do not silently overwrite the remote changes.
- Allow the user to cancel and reload.

A pre-submit comparison is not an atomic concurrency guarantee.

Do not claim that it prevents all races unless the server supports an appropriate conditional-write mechanism.

Document the remaining limitation.

### 4.2.10 Prevent Search Metadata Pollution

Phase 3 search hits may contain fields such as:

```text
_rankingScore
_rankingScoreDetails
_formatted
_matchesPosition
```

Do not use raw search hits as edit payloads.

Always obtain the actual source document through Document Service before editing.

Do not automatically include search-only metadata in document updates.

Preserve legitimate user-defined fields, even if they resemble metadata fields, unless their special meaning is confirmed.

### 4.2.11 Implement Editor Tests

Test:

- Create mode.
- Edit mode.
- JSON syntax errors.
- Non-object JSON.
- Missing primary key.
- Changed primary key.
- Nested JSON edits.
- Field additions.
- Field removals.
- Explicit null values.
- Array changes.
- Diff generation.
- Confirmation dialog.
- Cancel behavior.
- Unsaved changes.
- Concurrent modification warning.
- Read-only credential behavior.
- Search metadata isolation.

## Acceptance Criteria

- Users can create and edit JSON documents.
- The original document is preserved for comparison.
- Diff visualization reflects actual update semantics.
- Save requires explicit confirmation.
- Delete requires explicit confirmation.
- Unsaved changes are protected.
- Search-specific metadata is not accidentally persisted.
- The existing JSON Inspector remains read-only.

---

# Step 4.3 — Task Tracker

## Objective

Implement reliable tracking of asynchronous Meilisearch document mutations.

## Tasks

### 4.3.1 Verify the Tasks API

Consult the official Meilisearch Tasks API documentation.

Verify:

```http
GET /tasks/{task_uid}
```

Support documented task statuses:

```text
enqueued
processing
succeeded
failed
canceled
```

Verify the response format and required permissions.

Do not infer Task success solely from the HTTP response returned by a document mutation.

### 4.3.2 Implement Task Models

Define models equivalent to:

```text
TaskReference
TaskStatus
TrackedTask
TaskError
```

Track:

- Task UID.
- Connection ID.
- Index UID.
- Operation type.
- Target document ID, if applicable.
- Submission timestamp.
- Current status.
- Start and finish timestamps, when available.
- Task details.
- Task error.
- Last successful status check.

Preserve unknown optional task fields where appropriate.

### 4.3.3 Implement Task Retrieval

Implement a reusable backend service that retrieves a task by UID.

Use:

- Existing HTTP client.
- Existing credential resolution.
- Existing error classification.
- Context cancellation.

Do not introduce a generic server administration interface.

Only task retrieval needed for document mutation tracking is in scope.

### 4.3.4 Implement Task Polling

Implement bounded polling.

Recommended behavior:

- Begin polling after receiving a valid Task UID.
- Use a reasonable initial interval.
- Avoid unnecessary high-frequency requests.
- Support bounded backoff.
- Stop polling on terminal states.
- Stop or pause polling after repeated transport failures.
- Support explicit user cancellation of monitoring.

Monitoring cancellation must not be confused with cancellation of the server-side task.

Do not automatically call the Meilisearch task cancellation endpoint.

### 4.3.5 Handle Task State Transitions

Support:

```text
enqueued → processing → succeeded
                     → failed
                     → canceled
```

Do not require every intermediate state to be observed.

A task may transition to a terminal state between polling requests.

Treat succeeded, failed, and canceled as terminal states.

Display unknown or future task statuses safely rather than crashing.

### 4.3.6 Handle Task Errors

On failure, display available information:

- Task UID.
- Status.
- Error message.
- Error code.
- Error type.
- Documentation link, when provided.
- Relevant task details.

Do not expose credentials.

Distinguish task execution failure from failure to retrieve task status.

### 4.3.7 Handle Unavailable Task Permissions

A credential may allow document mutation without allowing task retrieval.

If task retrieval returns HTTP 403:

- Preserve the accepted task reference.
- Explain that task status cannot be verified.
- Do not claim success or failure.
- Do not automatically request administrator credentials.
- Allow the user to retry status retrieval when permissions permit.

Do not treat a successful mutation submission as proof of completed indexing.

### 4.3.8 Handle Timeouts and Unknown Outcomes

Polling may expire before a task reaches a terminal state.

Display an explicit monitoring timeout.

Preserve the Task UID.

Provide a manual Refresh Status action.

Avoid indefinite background polling.

If submission failed ambiguously and no Task UID was received, do not automatically retry the mutation.

Explain the uncertainty.

### 4.3.9 Implement Task Status UI

Provide a task panel displaying:

```text
Task UID
Operation
Connection
Index
Document ID
Status
Elapsed Time
Error Details
```

Use the existing application status area where practical.

Provide:

- Refresh status.
- Stop monitoring.
- View details.

Keep the UI lightweight.

Do not implement a complete task administration console.

### 4.3.10 Handle Connection Switching

Each tracked task belongs to a specific connection.

Do not retrieve it using credentials from another connection.

When the user switches connections:

- Keep task identities unambiguous.
- Stop or pause monitoring as appropriate.
- Prevent stale task results from appearing under the wrong connection.
- Preserve useful status information where practical.

Do not redirect existing tasks to the new connection.

### 4.3.11 Implement Task Tests

Test:

- Enqueued task.
- Processing task.
- Successful task.
- Failed task.
- Canceled task.
- Missing intermediate states.
- Unknown task status.
- HTTP 403 for task retrieval.
- Task not found.
- Network timeout.
- Polling cancellation.
- Monitoring timeout.
- Connection switching.
- Malformed responses.
- Missing task UID.
- Duplicate polling prevention.

## Acceptance Criteria

- Task references are captured.
- Asynchronous status changes are displayed.
- Polling terminates appropriately.
- Failures include useful diagnostics.
- Monitoring can be stopped safely.
- Permission limitations are handled.
- Accepted tasks are never incorrectly reported as completed.
- Task tracking works independently of the frontend.

---

# Step 4.4 — Integration, Safety & Testing

## Objective

Integrate document editing into the existing application and validate the complete mutation lifecycle.

## Tasks

### 4.4.1 Integrate with Document Browser

Add explicit actions:

- Add Document.
- Edit Document.
- Delete Document.

Do not make cells editable by default.

Do not modify the read-only JSON Inspector.

Use a dedicated editing interface.

### 4.4.2 Implement the Complete Edit Workflow

Validate:

1. Connect to Meilisearch.
2. Select an index.
3. Select a document.
4. Open the editor.
5. Retrieve the complete document.
6. Modify JSON.
7. Validate changes.
8. Review the diff.
9. Confirm the operation.
10. Submit the mutation.
11. Receive Task UID.
12. Monitor task status.
13. Observe successful completion.
14. Re-fetch the document.
15. Display the updated document.

Do not treat successful HTTP submission as completed indexing.

### 4.4.3 Implement Create Workflow

Validate:

1. Select the target index.
2. Open Add Document.
3. Enter a JSON object.
4. Validate the primary key.
5. Confirm the target and payload.
6. Submit the request.
7. Track the task.
8. Verify completion.
9. Retrieve the resulting document.
10. Refresh the affected Document Browser state.

If the document already exists, do not misrepresent overwrite-capable API behavior as create-only behavior.

### 4.4.4 Implement Delete Workflow

Validate:

1. Select a document.
2. Choose Delete.
3. Display the exact target.
4. Confirm deletion.
5. Submit the request.
6. Track the task.
7. Verify completion.
8. Refresh the Document Browser.
9. Clear or update selection.

Do not remove the item permanently from the UI before completion is confirmed.

### 4.4.5 Refresh Related Views

After successful mutations:

- Invalidate relevant Document Browser caches.
- Refresh affected document details.
- Refresh index statistics where appropriate.
- Clear stale selections.
- Preserve safe pagination behavior.

Do not automatically rerun every Query Console request.

Search results may be stale after indexing.

Provide clear refresh behavior without excessive background requests.

### 4.4.6 Validate Permission Scenarios

Test:

- Read-only API Key.
- Document write permission.
- Document write permission without task-read permission.
- Index-restricted credentials.
- Invalid API Key.
- Revoked API Key.
- Missing credentials.
- Unavailable server.
- Task failure after accepted submission.

Do not assume all write permissions are equivalent.

### 4.4.7 Validate Failure Recovery

Test:

- Invalid JSON.
- Primary-key errors.
- Document deleted before editing.
- Document modified by another client.
- Connection loss before submission.
- Connection loss after submission.
- Missing Task UID.
- Failed indexing task.
- Monitoring timeout.
- Application navigation during editing.
- Connection switching during a mutation.
- Application restart while a task remains active.

Preserve unsaved edits where appropriate.

For accepted tasks, retain enough context for manual recovery when practical.

A lightweight bounded record of pending task references may be used, provided it does not persist credentials or document payloads.

### 4.4.8 Cross-platform Validation

Validate on Windows and Linux:

- JSON Editor.
- Diff Viewer.
- Confirmations.
- Clipboard behavior.
- Document creation.
- Document updates.
- Document deletion.
- Task tracking.
- Error recovery.
- Application shutdown.

Preserve existing CI workflows.

### 4.4.9 Automated Integration Tests

Use a disposable Meilisearch instance for write-operation integration tests.

Never run destructive tests against an arbitrary configured user connection.

Use unique test indexes and documents.

Ensure cleanup is scoped strictly to test-created resources.

Test:

- Add document and confirm indexing.
- Update document and confirm changes.
- Replace document and verify removed fields.
- Delete document and confirm absence.
- Failed mutation.
- Task status transitions.
- Authorization failures.
- Concurrent modification detection.
- Connection isolation.

Do not claim that tests passed unless executed.

### 4.4.10 Security Review

Verify:

- No API Keys appear in logs.
- No credentials are included in Wails responses.
- No mutation is triggered by browsing.
- No mutation is submitted without confirmation.
- No administrative API is exposed.
- No arbitrary mutation endpoint is exposed.
- TLS validation remains enabled.
- No unsafe automatic mutation retries.
- No uncontrolled polling.
- No accidental document payload persistence.
- No stale data leaks between connections.

### 4.4.11 Documentation

Update documentation consistent with the existing repository structure.

Recommended files:

```text
README.md

docs/en/user-guide/document-editing.md
docs/ja/user-guide/document-editing.md

docs/en/user-guide/task-tracking.md
docs/ja/user-guide/task-tracking.md

docs/architecture/overview.md
docs/development/phase-4-report.md
```

Document:

- Creating documents.
- Editing documents.
- Partial update versus replacement.
- Deleting documents.
- Reviewing changes.
- Required API Key permissions.
- Task status interpretation.
- Handling failed tasks.
- Concurrent modification limitations.
- Recovery from uncertain submissions.

Update `plan.md` only when implementation decisions materially change the proposal or roadmap.

### 4.4.12 Phase Completion Report

Create:

`docs/development/phase-4-report.md`

Include:

1. Completed implementation steps.
2. Files created or modified.
3. API operation semantics.
4. JSON Editor design.
5. Diff Viewer design.
6. Confirmation workflow.
7. Task tracking implementation.
8. Permission handling.
9. Failure recovery behavior.
10. Backend test results.
11. Frontend test results.
12. Windows validation.
13. Linux validation.
14. Security review.
15. Known limitations.
16. Phase 5 readiness.

## Acceptance Criteria

- Document creation works.
- Existing documents can be edited.
- Individual documents can be deleted.
- All mutations require explicit user intent.
- Changes can be reviewed before submission.
- Task UID and status are visible.
- Failed tasks display useful errors.
- Successfully completed mutations trigger appropriate refresh.
- Unsaved edits are protected.
- Read-only users retain normal browsing functionality.
- Windows and Linux remain functional.
- Automated and integration tests pass.

---

# 4. General Implementation Constraints

## 4.1 Preserve Read-only Defaults

Existing browsing and searching features must remain read-only.

No automatic write operations.

No editing triggered by navigation or selection.

## 4.2 Do Not Implement Server Administration

Do not implement:

- Index creation or deletion.
- Index settings changes.
- API Key administration.
- Server configuration.
- Task deletion.
- Task cancellation.
- Bulk document deletion.
- Filter-based deletion.

Task tracking is allowed only as necessary for document mutation workflows.

## 4.3 Avoid Unsafe Write Semantics

Never silently substitute:

- Replacement for partial update.
- Partial update for replacement.
- Document insertion for a failed edit.
- A guessed document ID for an explicit identifier.

Do not claim optimistic concurrency guarantees that Meilisearch does not provide.

## 4.4 Preserve JSON Fidelity

Do not lose:

- Large numeric precision.
- Nested JSON structures.
- Array ordering.
- Unicode.
- Null values.
- Primary-key identity.
- Arbitrary user-defined fields.

Do not mistake omitted properties for explicit null values.

## 4.5 Preserve Connection Isolation

Every write operation and task must belong to an explicit connection and index.

Connection switching must not redirect in-progress operations.

## 4.6 Prevent Unintended Duplicate Writes

Do not automatically retry ambiguous mutation submissions.

Disable repeated submission while the same operation is in flight.

Re-enable editing controls only after a clearly defined outcome.

## 4.7 Keep the Implementation Focused

Phase 4 is an auxiliary editing capability.

Do not turn MeiliScope into a general database administration platform.

Prefer clear, conservative workflows over advanced editing features.

## 4.8 Maintain Cross-platform Compatibility

Preserve Windows and Linux behavior.

Reuse existing Wails and platform abstractions.

Avoid unnecessary dependencies.

---

# 5. Execution & Reporting Requirements

Execute the steps in order:

1. Step 4.1 — Document Write API.
2. Step 4.2 — JSON Document Editor.
3. Step 4.3 — Task Tracker.
4. Step 4.4 — Integration, Safety & Testing.

Each step must be independently reviewable.

After each step, report:

### Completed Work

Describe the implemented features.

### Files Changed

List important source files and documentation.

### Tests

Provide the actual executed commands and results.

### Technical Decisions

Explain significant implementation decisions.

### Known Issues

List unresolved problems and deferred functionality.

### Next Step Readiness

State whether the following step can proceed.

Do not hide test failures or unverified conditions.

Do not implement functionality from later phases unnecessarily.

---

# 6. Phase 4 Definition of Done

Phase 4 is complete when a user can:

1. Launch MeiliScope on Windows or Linux.
2. Connect to a Meilisearch instance.
3. Select an existing index.
4. Create a document using JSON.
5. Select and edit an existing document.
6. Review a meaningful change diff.
7. Explicitly confirm the selected write operation.
8. Submit the mutation.
9. Obtain and inspect the task reference.
10. Track asynchronous completion.
11. Understand a failed or canceled task.
12. Re-fetch the updated document after successful completion.
13. Delete an individual document after confirmation.
14. Continue browsing with read-only credentials.
15. Recover from ordinary editing and communication errors.
16. Switch connections without mutation-target confusion.
17. Continue using Query Console and Document Browser without regressions.

The application must never equate task acceptance with successful completion.

Document modification must remain an explicit, secondary feature.

**Successful completion of Phase 4 establishes safe, controlled document editing and asynchronous task tracking, preparing MeiliScope for Phase 5 — Usability & Release Preparation.**
# MeiliScope — Phase 4 Implementation Instructions

## Document Editing & Task Tracking

### 1. Project Context

MeiliScope is an open-source, standalone desktop application for exploring Meilisearch data.

**Mission: Explore your Meilisearch data.**

The established technology stack includes:

- Go for backend services.
- Wails v3 for desktop integration.
- React and TypeScript for the frontend.
- Meilisearch HTTP API for data access.
- Windows and Linux as validated execution platforms.
- MIT License.

The project proposal is maintained in the root-level `plan.md`.

Previous phases established:

**Phase 0 — Foundation**
- Desktop application infrastructure.
- Go/Frontend communication.
- Configuration and logging.
- Cross-platform builds and CI.

**Phase 1 — Connection & Index Explorer**
- Connection Manager.
- Secure credential storage.
- Meilisearch HTTP client.
- Index Explorer.

**Phase 2 — Document Browser**
- Document retrieval.
- Dynamic document tables.
- Pagination.
- JSON Inspector.

**Phase 3 — Query Console**
- Search API execution.
- JSON Query Editor.
- Search results and diagnostics.
- Query history.

Phase 4 introduces controlled document modification while preserving the application's read-only-first design.

Before implementation:

1. Read `plan.md`.
2. Review the existing architecture and Phase 0–3 reports.
3. Inspect the actual implementation.
4. Reuse established HTTP, credential, JSON, and frontend services.
5. Verify the supported Meilisearch API versions.
6. Identify existing editor, dialog, notification, and task-related infrastructure.

Do not rewrite working components without a documented reason.

---

# 2. Phase 4 Objectives

Implement safe document modification workflows.

Users must be able to:

1. Create a document using JSON.
2. Edit an existing document.
3. Delete an individual document.
4. Review changes before submission.
5. Confirm the target connection and index.
6. Submit write operations explicitly.
7. Track asynchronous Meilisearch tasks.
8. Inspect task errors.
9. Refresh affected documents after successful completion.
10. Recover from failures without losing unsaved edits.

## 2.1 Design Principles

**Read-only by Default**

Browsing and inspecting documents must never trigger modifications.

**Explicit User Intent**

A write operation must require an explicit user action.

**Precise Update Semantics**

Partial updates and complete replacements must not be confused.

**Asynchronous Completion**

An accepted request does not mean indexing has completed.

**No Silent Data Loss**

Preserve unsaved changes and handle concurrent modifications conservatively.

**Least Privilege**

Use the active connection's existing credentials. Never request administrative credentials automatically.

**Cross-platform Compatibility**

Maintain Windows and Linux functionality.

---

# 3. Architecture

Extend the existing architecture without introducing a separate application server.

```text
React / TypeScript
│
├── Document Browser
│   └── Edit / Delete actions
│
├── Document Editor
│   ├── JSON Editor
│   ├── Validation
│   ├── Diff Viewer
│   └── Confirmation
│
├── Task Status Panel
│
▼
Wails v3 Bindings
│
▼
Go Application Services
│
├── Connection Manager
├── Document Read Service
├── Document Write Service
├── Task Tracking Service
└── Existing Meilisearch HTTP Client
    │
    ▼
Meilisearch HTTP API
```

The frontend must not communicate directly with Meilisearch.

Go owns authentication, API execution, task retrieval, and transport errors.

React owns user interaction, editing state, diff presentation, and confirmation.

Keep document read operations available independently of write permissions.

Do not embed mutation behavior inside the existing read-only JSON Inspector.

---

# Step 4.1 — Document Write API

## Objective

Implement controlled document creation, update, replacement, and deletion through the existing Go HTTP client.

## Tasks

### 4.1.1 Verify Current API Specifications

Consult the official Meilisearch API documentation.

Verify the supported behavior of:

```http
POST   /indexes/{index_uid}/documents
PUT    /indexes/{index_uid}/documents
DELETE /indexes/{index_uid}/documents/{document_id}
```

Also verify:

```http
GET /tasks/{task_uid}
```

Record:

- Request and response formats.
- Required API Key actions.
- Primary-key requirements.
- Partial-update semantics.
- Complete-replacement semantics.
- Task response structure.
- Index creation side effects.
- Relevant API version differences.

Do not assume that PUT and POST have equivalent behavior.

### 4.1.2 Define Explicit Write Operations

Implement a service with separate operations equivalent to:

```go
type DocumentWriteService interface {
    AddDocument(
        ctx context.Context,
        target WriteTarget,
        document json.RawMessage,
    ) (TaskReference, error)

    UpdateDocument(
        ctx context.Context,
        target WriteTarget,
        document json.RawMessage,
    ) (TaskReference, error)

    ReplaceDocument(
        ctx context.Context,
        target WriteTarget,
        document json.RawMessage,
    ) (TaskReference, error)

    DeleteDocument(
        ctx context.Context,
        target WriteTarget,
        documentID string,
    ) (TaskReference, error)
}
```

Adapt exact signatures to the codebase.

The operations must have clearly documented semantics.

Do not expose a generic arbitrary HTTP mutation method through Wails.

For the initial GUI, use a conservative update workflow.

The full-replacement operation may remain a separate advanced operation if the distinction can be communicated safely.

### 4.1.3 Handle Document Creation

Support creation of one document at a time.

Requirements:

- Accept a JSON object.
- Reject invalid JSON.
- Reject arrays and non-object top-level values.
- Validate the presence of the required primary-key field when the index already defines one.
- Preserve arbitrary JSON fields.
- Preserve numeric precision.
- Preserve Unicode.
- Avoid automatic index creation.

Do not rely on implicit primary-key inference when it could unexpectedly create or alter an index.

If the target index does not exist or has no established primary key, explain the limitation instead of silently creating or configuring it.

Distinguish a document-addition workflow from a strictly enforced create-only operation.

Meilisearch document ingestion may update an existing document with the same primary key.

Therefore, do not claim that the API guarantees create-only behavior unless the supported server version actually provides such a guarantee.

### 4.1.4 Handle Document Updates

Support updating an existing document.

Before editing:

- Retrieve the complete original document.
- Identify its primary key.
- Preserve its original JSON snapshot.
- Record its connection and index identity.

An update must not rely on a partial table row or a formatted search hit as the original document.

Use the correct API semantics for the selected operation.

For partial updates:

- Preserve fields that are not included in the update payload.
- Handle explicit null values correctly.
- Do not interpret omission as deletion unless the API explicitly supports that behavior.

For complete replacement:

- Clearly explain that omitted fields may be removed.
- Require explicit confirmation of replacement semantics.
- Never substitute replacement for partial update silently.

The editor must not present unsupported field deletion semantics.

If field removal cannot be expressed safely with the chosen operation, show the limitation and offer only a supported, explicitly confirmed workflow.

### 4.1.5 Preserve Primary-key Identity

The existing document's primary-key value identifies the update target.

Requirements:

- Handle string and numeric identifiers.
- Preserve numeric precision.
- Encode document IDs correctly in URL paths.
- Reject missing required identifiers.
- Prevent accidental changes to the primary key during ordinary editing.

Treat primary-key modification as a separate unsupported operation during Phase 4.

Do not silently create another document when an attempted edit changes the identifier.

### 4.1.6 Implement Document Deletion

Support deleting one explicitly selected document.

Requirements:

- Use the selected document's actual identifier.
- Use the selected connection and index.
- Require explicit confirmation.
- Reject empty or invalid document identifiers.
- Return the resulting task reference.
- Do not delete an entire index.
- Do not introduce bulk or filter-based deletion.

Do not remove the document permanently from the local view merely because the deletion request was accepted.

Wait for confirmed task completion before presenting deletion as successful.

### 4.1.7 Enforce Operation Scope

A write request must be bound to an immutable target snapshot:

```text
Connection ID
Index UID
Document ID, where applicable
Operation Type
```

Once submission begins, a connection or index change must not redirect the operation to a different target.

Validate the target again immediately before dispatch.

The backend must not trust frontend-provided credentials.

Credentials must be resolved internally through Connection Manager.

### 4.1.8 Handle Permission Errors

Write operations may require permissions distinct from read operations.

Handle:

- Valid read-only credentials.
- Missing write permission.
- Restricted index permissions.
- Invalid credentials.
- Revoked credentials.
- Missing index.
- Server rejection.

Do not require an Admin API Key.

Do not claim that a document is editable merely because it can be read.

Avoid inferring write permission from unrelated API calls.

### 4.1.9 Avoid Unsafe Retries

Do not automatically retry document mutations after ambiguous network failures.

A timeout or connection reset may occur after the server accepted the operation.

Classify these outcomes as potentially submitted when appropriate.

Inform the user that the final state is unknown.

Do not report a failed mutation as definitely not applied unless that conclusion is supported by the response.

### 4.1.10 Implement Backend Tests

Use `httptest.Server`.

Test:

- Document creation request.
- Partial-update request.
- Replacement request.
- Individual deletion.
- Invalid JSON.
- Invalid primary key.
- Numeric and string identifiers.
- Nested JSON.
- Unicode.
- Large JSON numbers.
- Unsupported operation.
- Incorrect HTTP method.
- HTTP 401/403/404/409/429/500.
- Timeout.
- Context cancellation.
- Task response decoding.
- Missing task UID.
- Ambiguous submission failure.
- Credential sanitization.

## Acceptance Criteria

- Document write operations use the existing HTTP client.
- Add, update, replace, and delete semantics are explicit.
- Document identifiers remain accurate.
- No administrative API is required.
- No implicit index creation is performed.
- Mutations are not automatically retried after uncertain submission.
- All backend tests pass.
- Wails does not expose arbitrary write endpoints.

---

# Step 4.2 — JSON Document Editor

## Objective

Implement a safe JSON editing interface with change review and explicit confirmation.

## Tasks

### 4.2.1 Reuse the Existing Editor Infrastructure

Inspect the JSON editor implemented in Phase 3.

Reuse its:

- Syntax highlighting.
- Formatting.
- Validation.
- Keyboard handling.
- Error markers.
- Editor integration.

Do not introduce another large editor framework.

Keep Query Console and Document Editor state separate.

### 4.2.2 Define Editor Modes

Support:

**Create Mode**

Create a document from a new JSON object.

**Edit Mode**

Modify an existing document obtained through Document Service.

**Delete Confirmation**

Confirm deletion of an existing document.

Do not implement inline editing inside Document Table during this phase.

### 4.2.3 Implement Editing Workflow

For existing documents:

1. User selects a document.
2. User explicitly chooses Edit.
3. Retrieve the full document if necessary.
4. Create an immutable original snapshot.
5. Open the JSON editor.
6. User modifies the document.
7. Validate JSON syntax and primary-key identity.
8. Present a diff.
9. Confirm the operation.
10. Submit through Document Write Service.

Do not modify the original snapshot while editing.

### 4.2.4 Implement JSON Validation

Validate:

- Correct JSON syntax.
- Object top-level structure.
- Required primary-key field.
- Primary-key identity.
- Supported document size.
- Valid operation semantics.

Preserve arbitrary user-defined fields.

Do not enforce a fixed document schema.

Avoid losing precision when parsing or serializing large numbers.

Do not convert missing fields into null values automatically.

### 4.2.5 Implement Diff Viewer

Display differences between the original and edited JSON.

Support:

- Added fields.
- Changed values.
- Removed fields.
- Nested changes.
- Array changes.
- Null changes.

Use a readable side-by-side or unified presentation.

Prefer structural comparison where practical.

Do not treat object key reordering as a semantic change.

Preserve array order.

Make the diff accurately reflect the effective operation.

For partial updates, distinguish omitted properties from explicit values.

For complete replacement, show the possible removal of omitted fields.

The diff viewer must not claim that a field is deleted when the selected API operation would preserve it.

### 4.2.6 Implement Save Confirmation

Before submitting a mutation, display:

```text
Connection
Server URL
Index UID
Document ID
Operation Type
Change Summary
```

Use the validated submission snapshot.

Do not allow the confirmation dialog to change the operation target.

Require explicit approval.

Do not submit automatically when the user closes the editor.

For destructive or replacement operations, use more prominent confirmation wording.

### 4.2.7 Implement Delete Confirmation

Deletion must show:

- Connection.
- Index.
- Document ID.
- Irreversible-operation warning.
- Explicit Delete action.
- Cancel action.

Use a visually distinct destructive action.

Do not allow accidental deletion through ordinary row selection or keyboard navigation.

### 4.2.8 Handle Unsaved Changes

Detect whether the document has been modified.

If the user attempts to:

- Close the editor.
- Switch documents.
- Switch indexes.
- Switch connections.
- Navigate away.

Offer an appropriate choice to discard changes or remain in the editor.

Do not silently discard unsaved changes.

Keep editor state in memory unless the user explicitly chooses to submit the changes.

Do not persist arbitrary unsaved document contents automatically in configuration files.

### 4.2.9 Handle Concurrent Modifications

Meilisearch documents may be modified by other clients while the editor is open.

Before submission, re-fetch the current document when possible.

Compare it with the original snapshot.

If the server document changed:

- Display a conflict warning.
- Show the original, remote, and edited versions where practical.
- Do not silently overwrite the remote changes.
- Allow the user to cancel and reload.

A pre-submit comparison is not an atomic concurrency guarantee.

Do not claim that it prevents all races unless the server supports an appropriate conditional-write mechanism.

Document the remaining limitation.

### 4.2.10 Prevent Search Metadata Pollution

Phase 3 search hits may contain fields such as:

```text
_rankingScore
_rankingScoreDetails
_formatted
_matchesPosition
```

Do not use raw search hits as edit payloads.

Always obtain the actual source document through Document Service before editing.

Do not automatically include search-only metadata in document updates.

Preserve legitimate user-defined fields, even if they resemble metadata fields, unless their special meaning is confirmed.

### 4.2.11 Implement Editor Tests

Test:

- Create mode.
- Edit mode.
- JSON syntax errors.
- Non-object JSON.
- Missing primary key.
- Changed primary key.
- Nested JSON edits.
- Field additions.
- Field removals.
- Explicit null values.
- Array changes.
- Diff generation.
- Confirmation dialog.
- Cancel behavior.
- Unsaved changes.
- Concurrent modification warning.
- Read-only credential behavior.
- Search metadata isolation.

## Acceptance Criteria

- Users can create and edit JSON documents.
- The original document is preserved for comparison.
- Diff visualization reflects actual update semantics.
- Save requires explicit confirmation.
- Delete requires explicit confirmation.
- Unsaved changes are protected.
- Search-specific metadata is not accidentally persisted.
- The existing JSON Inspector remains read-only.

---

# Step 4.3 — Task Tracker

## Objective

Implement reliable tracking of asynchronous Meilisearch document mutations.

## Tasks

### 4.3.1 Verify the Tasks API

Consult the official Meilisearch Tasks API documentation.

Verify:

```http
GET /tasks/{task_uid}
```

Support documented task statuses:

```text
enqueued
processing
succeeded
failed
canceled
```

Verify the response format and required permissions.

Do not infer Task success solely from the HTTP response returned by a document mutation.

### 4.3.2 Implement Task Models

Define models equivalent to:

```text
TaskReference
TaskStatus
TrackedTask
TaskError
```

Track:

- Task UID.
- Connection ID.
- Index UID.
- Operation type.
- Target document ID, if applicable.
- Submission timestamp.
- Current status.
- Start and finish timestamps, when available.
- Task details.
- Task error.
- Last successful status check.

Preserve unknown optional task fields where appropriate.

### 4.3.3 Implement Task Retrieval

Implement a reusable backend service that retrieves a task by UID.

Use:

- Existing HTTP client.
- Existing credential resolution.
- Existing error classification.
- Context cancellation.

Do not introduce a generic server administration interface.

Only task retrieval needed for document mutation tracking is in scope.

### 4.3.4 Implement Task Polling

Implement bounded polling.

Recommended behavior:

- Begin polling after receiving a valid Task UID.
- Use a reasonable initial interval.
- Avoid unnecessary high-frequency requests.
- Support bounded backoff.
- Stop polling on terminal states.
- Stop or pause polling after repeated transport failures.
- Support explicit user cancellation of monitoring.

Monitoring cancellation must not be confused with cancellation of the server-side task.

Do not automatically call the Meilisearch task cancellation endpoint.

### 4.3.5 Handle Task State Transitions

Support:

```text
enqueued → processing → succeeded
                     → failed
                     → canceled
```

Do not require every intermediate state to be observed.

A task may transition to a terminal state between polling requests.

Treat succeeded, failed, and canceled as terminal states.

Display unknown or future task statuses safely rather than crashing.

### 4.3.6 Handle Task Errors

On failure, display available information:

- Task UID.
- Status.
- Error message.
- Error code.
- Error type.
- Documentation link, when provided.
- Relevant task details.

Do not expose credentials.

Distinguish task execution failure from failure to retrieve task status.

### 4.3.7 Handle Unavailable Task Permissions

A credential may allow document mutation without allowing task retrieval.

If task retrieval returns HTTP 403:

- Preserve the accepted task reference.
- Explain that task status cannot be verified.
- Do not claim success or failure.
- Do not automatically request administrator credentials.
- Allow the user to retry status retrieval when permissions permit.

Do not treat a successful mutation submission as proof of completed indexing.

### 4.3.8 Handle Timeouts and Unknown Outcomes

Polling may expire before a task reaches a terminal state.

Display an explicit monitoring timeout.

Preserve the Task UID.

Provide a manual Refresh Status action.

Avoid indefinite background polling.

If submission failed ambiguously and no Task UID was received, do not automatically retry the mutation.

Explain the uncertainty.

### 4.3.9 Implement Task Status UI

Provide a task panel displaying:

```text
Task UID
Operation
Connection
Index
Document ID
Status
Elapsed Time
Error Details
```

Use the existing application status area where practical.

Provide:

- Refresh status.
- Stop monitoring.
- View details.

Keep the UI lightweight.

Do not implement a complete task administration console.

### 4.3.10 Handle Connection Switching

Each tracked task belongs to a specific connection.

Do not retrieve it using credentials from another connection.

When the user switches connections:

- Keep task identities unambiguous.
- Stop or pause monitoring as appropriate.
- Prevent stale task results from appearing under the wrong connection.
- Preserve useful status information where practical.

Do not redirect existing tasks to the new connection.

### 4.3.11 Implement Task Tests

Test:

- Enqueued task.
- Processing task.
- Successful task.
- Failed task.
- Canceled task.
- Missing intermediate states.
- Unknown task status.
- HTTP 403 for task retrieval.
- Task not found.
- Network timeout.
- Polling cancellation.
- Monitoring timeout.
- Connection switching.
- Malformed responses.
- Missing task UID.
- Duplicate polling prevention.

## Acceptance Criteria

- Task references are captured.
- Asynchronous status changes are displayed.
- Polling terminates appropriately.
- Failures include useful diagnostics.
- Monitoring can be stopped safely.
- Permission limitations are handled.
- Accepted tasks are never incorrectly reported as completed.
- Task tracking works independently of the frontend.

---

# Step 4.4 — Integration, Safety & Testing

## Objective

Integrate document editing into the existing application and validate the complete mutation lifecycle.

## Tasks

### 4.4.1 Integrate with Document Browser

Add explicit actions:

- Add Document.
- Edit Document.
- Delete Document.

Do not make cells editable by default.

Do not modify the read-only JSON Inspector.

Use a dedicated editing interface.

### 4.4.2 Implement the Complete Edit Workflow

Validate:

1. Connect to Meilisearch.
2. Select an index.
3. Select a document.
4. Open the editor.
5. Retrieve the complete document.
6. Modify JSON.
7. Validate changes.
8. Review the diff.
9. Confirm the operation.
10. Submit the mutation.
11. Receive Task UID.
12. Monitor task status.
13. Observe successful completion.
14. Re-fetch the document.
15. Display the updated document.

Do not treat successful HTTP submission as completed indexing.

### 4.4.3 Implement Create Workflow

Validate:

1. Select the target index.
2. Open Add Document.
3. Enter a JSON object.
4. Validate the primary key.
5. Confirm the target and payload.
6. Submit the request.
7. Track the task.
8. Verify completion.
9. Retrieve the resulting document.
10. Refresh the affected Document Browser state.

If the document already exists, do not misrepresent overwrite-capable API behavior as create-only behavior.

### 4.4.4 Implement Delete Workflow

Validate:

1. Select a document.
2. Choose Delete.
3. Display the exact target.
4. Confirm deletion.
5. Submit the request.
6. Track the task.
7. Verify completion.
8. Refresh the Document Browser.
9. Clear or update selection.

Do not remove the item permanently from the UI before completion is confirmed.

### 4.4.5 Refresh Related Views

After successful mutations:

- Invalidate relevant Document Browser caches.
- Refresh affected document details.
- Refresh index statistics where appropriate.
- Clear stale selections.
- Preserve safe pagination behavior.

Do not automatically rerun every Query Console request.

Search results may be stale after indexing.

Provide clear refresh behavior without excessive background requests.

### 4.4.6 Validate Permission Scenarios

Test:

- Read-only API Key.
- Document write permission.
- Document write permission without task-read permission.
- Index-restricted credentials.
- Invalid API Key.
- Revoked API Key.
- Missing credentials.
- Unavailable server.
- Task failure after accepted submission.

Do not assume all write permissions are equivalent.

### 4.4.7 Validate Failure Recovery

Test:

- Invalid JSON.
- Primary-key errors.
- Document deleted before editing.
- Document modified by another client.
- Connection loss before submission.
- Connection loss after submission.
- Missing Task UID.
- Failed indexing task.
- Monitoring timeout.
- Application navigation during editing.
- Connection switching during a mutation.
- Application restart while a task remains active.

Preserve unsaved edits where appropriate.

For accepted tasks, retain enough context for manual recovery when practical.

A lightweight bounded record of pending task references may be used, provided it does not persist credentials or document payloads.

### 4.4.8 Cross-platform Validation

Validate on Windows and Linux:

- JSON Editor.
- Diff Viewer.
- Confirmations.
- Clipboard behavior.
- Document creation.
- Document updates.
- Document deletion.
- Task tracking.
- Error recovery.
- Application shutdown.

Preserve existing CI workflows.

### 4.4.9 Automated Integration Tests

Use a disposable Meilisearch instance for write-operation integration tests.

Never run destructive tests against an arbitrary configured user connection.

Use unique test indexes and documents.

Ensure cleanup is scoped strictly to test-created resources.

Test:

- Add document and confirm indexing.
- Update document and confirm changes.
- Replace document and verify removed fields.
- Delete document and confirm absence.
- Failed mutation.
- Task status transitions.
- Authorization failures.
- Concurrent modification detection.
- Connection isolation.

Do not claim that tests passed unless executed.

### 4.4.10 Security Review

Verify:

- No API Keys appear in logs.
- No credentials are included in Wails responses.
- No mutation is triggered by browsing.
- No mutation is submitted without confirmation.
- No administrative API is exposed.
- No arbitrary mutation endpoint is exposed.
- TLS validation remains enabled.
- No unsafe automatic mutation retries.
- No uncontrolled polling.
- No accidental document payload persistence.
- No stale data leaks between connections.

### 4.4.11 Documentation

Update documentation consistent with the existing repository structure.

Recommended files:

```text
README.md

docs/en/user-guide/document-editing.md
docs/ja/user-guide/document-editing.md

docs/en/user-guide/task-tracking.md
docs/ja/user-guide/task-tracking.md

docs/architecture/overview.md
docs/development/phase-4-report.md
```

Document:

- Creating documents.
- Editing documents.
- Partial update versus replacement.
- Deleting documents.
- Reviewing changes.
- Required API Key permissions.
- Task status interpretation.
- Handling failed tasks.
- Concurrent modification limitations.
- Recovery from uncertain submissions.

Update `plan.md` only when implementation decisions materially change the proposal or roadmap.

### 4.4.12 Phase Completion Report

Create:

`docs/development/phase-4-report.md`

Include:

1. Completed implementation steps.
2. Files created or modified.
3. API operation semantics.
4. JSON Editor design.
5. Diff Viewer design.
6. Confirmation workflow.
7. Task tracking implementation.
8. Permission handling.
9. Failure recovery behavior.
10. Backend test results.
11. Frontend test results.
12. Windows validation.
13. Linux validation.
14. Security review.
15. Known limitations.
16. Phase 5 readiness.

## Acceptance Criteria

- Document creation works.
- Existing documents can be edited.
- Individual documents can be deleted.
- All mutations require explicit user intent.
- Changes can be reviewed before submission.
- Task UID and status are visible.
- Failed tasks display useful errors.
- Successfully completed mutations trigger appropriate refresh.
- Unsaved edits are protected.
- Read-only users retain normal browsing functionality.
- Windows and Linux remain functional.
- Automated and integration tests pass.

---

# 4. General Implementation Constraints

## 4.1 Preserve Read-only Defaults

Existing browsing and searching features must remain read-only.

No automatic write operations.

No editing triggered by navigation or selection.

## 4.2 Do Not Implement Server Administration

Do not implement:

- Index creation or deletion.
- Index settings changes.
- API Key administration.
- Server configuration.
- Task deletion.
- Task cancellation.
- Bulk document deletion.
- Filter-based deletion.

Task tracking is allowed only as necessary for document mutation workflows.

## 4.3 Avoid Unsafe Write Semantics

Never silently substitute:

- Replacement for partial update.
- Partial update for replacement.
- Document insertion for a failed edit.
- A guessed document ID for an explicit identifier.

Do not claim optimistic concurrency guarantees that Meilisearch does not provide.

## 4.4 Preserve JSON Fidelity

Do not lose:

- Large numeric precision.
- Nested JSON structures.
- Array ordering.
- Unicode.
- Null values.
- Primary-key identity.
- Arbitrary user-defined fields.

Do not mistake omitted properties for explicit null values.

## 4.5 Preserve Connection Isolation

Every write operation and task must belong to an explicit connection and index.

Connection switching must not redirect in-progress operations.

## 4.6 Prevent Unintended Duplicate Writes

Do not automatically retry ambiguous mutation submissions.

Disable repeated submission while the same operation is in flight.

Re-enable editing controls only after a clearly defined outcome.

## 4.7 Keep the Implementation Focused

Phase 4 is an auxiliary editing capability.

Do not turn MeiliScope into a general database administration platform.

Prefer clear, conservative workflows over advanced editing features.

## 4.8 Maintain Cross-platform Compatibility

Preserve Windows and Linux behavior.

Reuse existing Wails and platform abstractions.

Avoid unnecessary dependencies.

---

# 5. Execution & Reporting Requirements

Execute the steps in order:

1. Step 4.1 — Document Write API.
2. Step 4.2 — JSON Document Editor.
3. Step 4.3 — Task Tracker.
4. Step 4.4 — Integration, Safety & Testing.

Each step must be independently reviewable.

After each step, report:

### Completed Work

Describe the implemented features.

### Files Changed

List important source files and documentation.

### Tests

Provide the actual executed commands and results.

### Technical Decisions

Explain significant implementation decisions.

### Known Issues

List unresolved problems and deferred functionality.

### Next Step Readiness

State whether the following step can proceed.

Do not hide test failures or unverified conditions.

Do not implement functionality from later phases unnecessarily.

---

# 6. Phase 4 Definition of Done

Phase 4 is complete when a user can:

1. Launch MeiliScope on Windows or Linux.
2. Connect to a Meilisearch instance.
3. Select an existing index.
4. Create a document using JSON.
5. Select and edit an existing document.
6. Review a meaningful change diff.
7. Explicitly confirm the selected write operation.
8. Submit the mutation.
9. Obtain and inspect the task reference.
10. Track asynchronous completion.
11. Understand a failed or canceled task.
12. Re-fetch the updated document after successful completion.
13. Delete an individual document after confirmation.
14. Continue browsing with read-only credentials.
15. Recover from ordinary editing and communication errors.
16. Switch connections without mutation-target confusion.
17. Continue using Query Console and Document Browser without regressions.

The application must never equate task acceptance with successful completion.

Document modification must remain an explicit, secondary feature.

**Successful completion of Phase 4 establishes safe, controlled document editing and asynchronous task tracking, preparing MeiliScope for Phase 5 — Usability & Release Preparation.**
