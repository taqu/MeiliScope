# MeiliScope — Phase 6 Implementation Instructions

## Quality, Security & Stability

### 1. Project Context

MeiliScope is an open-source desktop application for exploring and inspecting Meilisearch data.

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
- Cross-platform build system.
- Configuration management.
- Application logging.
- CI/CD foundation.

**Phase 1 — Connection & Index Explorer**
- Connection Manager.
- Credential storage.
- HTTP client.
- Index browsing and statistics.

**Phase 2 — Document Browser**
- Document retrieval.
- Dynamic tables.
- Server-side pagination.
- JSON Inspector.

**Phase 3 — Query Console**
- Search API execution.
- JSON Query Editor.
- Search results and diagnostics.
- Query History.

**Phase 4 — Document Editing**
- Document creation, modification, and deletion.
- Diff visualization.
- Confirmation workflows.
- Asynchronous Task Tracking.

**Phase 5 — Usability**
- Search Builder.
- Saved Queries.
- JSON/CSV Export.
- Keyboard shortcuts.
- UI/UX improvements.

Phase 6 focuses on making the existing application reliable, secure, and maintainable.

**Do not introduce significant new user-facing features during this phase.**

The goal is to identify defects, fix them, prevent regressions, and establish objective release-readiness evidence.

---

# 2. Phase 6 Objectives

Phase 6 has three major objectives:

1. Validate Meilisearch API compatibility and existing feature behavior.
2. Strengthen security, failure handling, and recovery.
3. Validate performance, resource usage, and long-running stability.

The implementation must also establish a reliable quality gate for Phase 7.

## 2.1 General Principles

**Evidence-Based Validation**

Do not report success without executing the relevant tests.

**Fix, Test, Verify**

Every confirmed defect must have a reproducible test where practical.

**Preserve Existing Behavior**

Avoid broad architectural rewrites or unnecessary dependency changes.

**Safe Testing**

Never execute destructive tests against a user's production Meilisearch instance.

**Reproducibility**

Document test environments, commands, dataset sizes, versions, and results.

**Cross-platform**

Validate both Windows and Linux.

**Release-Oriented**

Prioritize defects that affect data integrity, security, application stability, and core workflows.

---

# 3. Quality Gates and Defect Classification

Before executing the individual steps, establish a consistent classification system.

## 3.1 Defect Severity

**Critical**

Examples:

- Credential disclosure.
- Unintended document deletion or replacement.
- Writes sent to the wrong connection or index.
- Data corruption.
- Known exploitable security vulnerabilities.
- Application behavior that can silently damage user data.

**High**

Examples:

- Application crashes during normal workflows.
- Core browsing or search functionality unavailable.
- Persistent settings lost unexpectedly.
- Task failure incorrectly reported as success.
- Repeated operations causing severe resource exhaustion.
- Major Windows or Linux regression.

**Medium**

Examples:

- Recoverable workflow errors.
- Incorrect non-destructive metadata display.
- Inconsistent navigation state.
- Broken optional shortcuts.
- Localized performance regressions.

**Low**

Examples:

- Minor visual inconsistencies.
- Non-blocking cosmetic defects.
- Minor documentation discrepancies.

Record severity, reproduction steps, affected versions, root cause, fix, and regression test for each confirmed defect.

## 3.2 Release-blocking Policy

The following must block Phase 6 completion:

- Any unresolved Critical defect.
- Any unresolved High defect affecting the supported release configuration.
- Failed required CI checks.
- Known credential leakage.
- Known unintended document mutation.
- Known silent data corruption.
- Unverified behavior that is essential to a claimed release guarantee.

Medium and Low issues may be deferred when their impact and rationale are documented.

Do not silently downgrade defects to bypass release gates.

## 3.3 Validation Status

Use the following statuses consistently:

- PASS — Executed successfully.
- FAIL — Executed and failed.
- BLOCKED — Could not proceed because of a specific blocker.
- NOT RUN — Not executed.
- NOT APPLICABLE — Not relevant, with justification.

Do not equate BLOCKED or NOT RUN with PASS.

---

# Step 6.1 — API Compatibility & Regression Testing

## Objective

Verify that MeiliScope works correctly with the Meilisearch API versions it intends to support.

Establish comprehensive regression coverage for functionality implemented in Phases 1–5.

## Tasks

### 6.1.1 Inspect the Existing Compatibility Policy

Read:

- `plan.md`
- Architecture documentation.
- Existing API client implementation.
- Existing version compatibility notes.
- Previous phase reports.
- Existing integration tests.

Identify the Meilisearch versions actually tested during development.

Do not assume that support for one server version guarantees compatibility with every newer or older version.

### 6.1.2 Define the Compatibility Matrix

Create:

`docs/development/compatibility-matrix.md`

Record:

- MeiliScope revision or version.
- Meilisearch version.
- Operating system.
- API capability.
- Test result.
- Known limitations.
- Verification date.

Use the project's documented target Meilisearch version as the baseline.

Where feasible, test additional release lines to establish a defensible compatibility range.

Do not invent compatibility claims for untested versions.

The supported version range must be based on actual validation.

### 6.1.3 Audit API Usage

Review all Meilisearch API operations used by the application.

Cover:

**Connection & Index**
- Connection testing.
- Index enumeration.
- Index details.
- Index statistics.

**Document Browser**
- Document listing.
- Document pagination.
- Document detail retrieval.
- Field selection.

**Query Console**
- Search execution.
- Search pagination.
- Facets.
- Ranking metadata.
- Search diagnostics.

**Document Editing**
- Document creation.
- Partial update.
- Replacement.
- Individual deletion.
- Task retrieval.

Verify against the official Meilisearch documentation for the supported versions.

Check:

- Endpoint paths.
- HTTP methods.
- Request formats.
- Response formats.
- Permission requirements.
- Error codes.
- Optional fields.
- Version-specific features.

Do not rely solely on previously written implementation assumptions.

### 6.1.4 Validate JSON Response Compatibility

Test the application's handling of:

- Missing optional fields.
- Additional unknown fields.
- Null fields.
- Empty arrays.
- Empty objects.
- Large numeric values.
- Unicode values.
- Unexpected but valid field ordering.
- Different pagination response structures.
- Unknown task metadata.

Avoid fragile decoding that rejects harmless additional response fields.

Preserve raw JSON where required by the established architecture.

Do not silently lose numeric precision.

### 6.1.5 Validate API Key Permissions

Create tests using credentials representing different permission sets.

Cover:

- Unauthenticated local instances.
- Index enumeration permission.
- Document read permission.
- Search-only permission.
- Document write permission.
- Task read permission.
- Index-restricted credentials.
- Invalid credentials.
- Revoked credentials.

Verify the actual permission behavior of the supported Meilisearch versions.

Important:

A Search API Key may allow searching a known index without allowing index enumeration.

Document write permission does not necessarily imply task-read permission.

Do not require administrative credentials for ordinary read-only workflows.

### 6.1.6 Validate Connection Management

Test:

- Multiple saved profiles.
- Duplicate server URLs.
- Different credentials for the same server.
- Invalid server URL.
- HTTP and HTTPS.
- URLs containing supported path prefixes.
- Connection switching.
- Connection deletion.
- Credential replacement.
- Application restart.

Ensure that connection-specific data cannot leak between profiles.

### 6.1.7 Validate Document Browser

Test:

- Empty index.
- Single document.
- Large index.
- Pagination.
- Dynamic columns.
- Missing fields.
- Null fields.
- Nested JSON.
- Large numbers.
- Large strings.
- Unicode.
- Document deletion while browsing.
- Index deletion by another client.
- Rapid connection switching.

Ensure that large indexes are not loaded into memory in their entirety.

### 6.1.8 Validate Query Console

Test:

- Empty query.
- Keyword query.
- Filters.
- Sort.
- Facets.
- Offset pagination.
- Page-based pagination.
- Ranking scores.
- Ranking score details.
- Search-only credentials.
- Unknown request parameters.
- Unknown response fields.
- Invalid search parameters.
- Request cancellation.

Verify that Query Console continues to accept arbitrary valid Search API JSON objects.

Do not introduce a restrictive parameter allowlist.

### 6.1.9 Validate Search Builder

Verify that Search Builder does not damage query JSON.

Test:

- Builder-to-JSON synchronization.
- JSON-to-Builder synchronization.
- Unknown parameters.
- Complex filters.
- Sort ordering.
- Facet selection.
- Pagination mode preservation.
- Invalid JSON handling.

Compare query semantics before and after Builder edits.

Unrelated query parameters must remain unchanged.

### 6.1.10 Validate Document Editing

Use a dedicated disposable Meilisearch test instance.

Test:

- Document creation.
- Existing document updates.
- Complete replacement.
- Explicit deletion.
- Primary-key validation.
- Diff accuracy.
- Confirmation.
- Task success.
- Task failure.
- Task permission denied.
- Concurrent modification warning.
- Connection switching during writes.

Never target an arbitrary user-configured server.

Never delete a non-test index during automated validation.

### 6.1.11 Validate Local Persistence

Test:

- Connection profiles.
- UI preferences.
- Query History.
- Saved Queries.
- Column preferences.
- Credential references.
- Pending task references, if persisted.

Verify that application restarts do not corrupt or unexpectedly discard user data.

Test configuration migrations where supported.

### 6.1.12 Establish Regression Test Suites

Organize automated validation into clearly separated categories:

**Unit Tests**

Fast, deterministic tests without external services.

**Component Tests**

Frontend rendering and interaction tests.

**Integration Tests**

Tests involving a disposable Meilisearch instance.

**Desktop Smoke Tests**

Tests involving the actual Wails desktop application.

**Compatibility Tests**

Tests against selected Meilisearch versions.

Do not require a running Meilisearch instance for ordinary unit tests.

Use explicit setup and cleanup for integration tests.

### 6.1.13 Update CI

Extend the existing GitHub Actions workflow.

Ensure that the required validation jobs cover:

- Go formatting and static analysis.
- Go unit tests.
- Frontend lint and type checking.
- Frontend tests.
- Windows build.
- Linux build.
- Applicable integration tests.

Use an explicit Meilisearch test-version matrix where practical.

Avoid significantly increasing the default PR test duration without justification.

Expensive compatibility tests may run in a separate scheduled or manually triggered workflow.

Use pinned Meilisearch versions and record the resolved version in test output.

### 6.1.14 Fix Compatibility Defects

For each confirmed issue:

1. Reproduce it.
2. Identify the affected API version.
3. Determine the root cause.
4. Implement the smallest safe correction.
5. Add a regression test.
6. Run related existing tests.
7. Update compatibility documentation.

Avoid implementing speculative compatibility shims for untested server versions.

## Deliverables

- Compatibility matrix.
- Updated API integration tests.
- Regression test coverage.
- CI improvements.
- Fixed compatibility defects.
- Documentation of unsupported features or versions.

## Acceptance Criteria

- A supported Meilisearch version range is explicitly documented.
- Core API workflows pass for the declared supported configurations.
- Authentication and permission scenarios are tested.
- Unknown JSON fields do not cause unnecessary failures.
- Existing browsing, searching, and editing behavior remains functional.
- Confirmed compatibility defects have regression tests.
- No unverified server version is claimed as supported.

---

# Step 6.2 — Security & Failure Recovery

## Objective

Review the application for credential leakage, unsafe mutations, insecure persistence, and failures that could compromise user data or application reliability.

Fix identified defects and add regression tests.

## Tasks

### 6.2.1 Establish a Threat Model

Create:

`docs/architecture/security.md`

Consider the following assets:

- Meilisearch API Keys.
- Connection profiles.
- Saved queries.
- Query History.
- Document data.
- Exported files.
- Local configuration.
- Pending mutation tasks.

Consider relevant threats:

- Credential exposure.
- Unauthorized API operations.
- Malicious document content.
- Unsafe file writes.
- Accidental mutations.
- Connection identity confusion.
- Stale asynchronous responses.
- Local storage corruption.
- Dependency vulnerabilities.

Clearly document trust boundaries.

The frontend, local files, and remote Meilisearch responses must not be assumed to contain trusted data.

Avoid claiming protection against operating-system compromise or a fully compromised user session.

### 6.2.2 Audit Credential Storage

Review Windows credential storage and Linux Secret Service integration.

Verify:

- API Keys are stored separately from ordinary settings.
- No plaintext credential fallback exists.
- Session-only mode works.
- Credential deletion behaves correctly.
- Missing credentials are handled.
- Credential replacement does not destroy a working profile unexpectedly.
- Linux keyring failures are reported.
- Credentials do not appear in frontend responses.

Use mock credential stores for automated failure tests.

Run platform-specific integration tests where available.

### 6.2.3 Audit Logging and Diagnostics

Inspect all application logs.

Ensure that logs do not contain:

- API Keys.
- Authorization headers.
- Credential values.
- Full document payloads by default.
- Full search queries by default.
- Sensitive raw HTTP bodies without explicit diagnostic authorization.

Review error messages and stack traces.

Implement consistent secret redaction.

Do not rely only on UI masking to protect credentials.

Test error formatting using synthetic secret values.

### 6.2.4 Audit HTTP Security

Verify:

- TLS certificate verification remains enabled.
- HTTPS is never silently downgraded.
- Authentication headers are sent only to the intended server.
- URL path joining is correct.
- Credentials are not embedded in URLs.
- Redirect behavior cannot forward credentials to unintended origins.
- Request timeouts exist.
- Response sizes are bounded where necessary.
- Cancellation works.

Review HTTP redirects, including cross-origin redirects, explicitly.

Reject unsupported URL schemes.

Do not introduce a general-purpose arbitrary HTTP proxy.

### 6.2.5 Audit Frontend Injection Risks

Review all components displaying content from Meilisearch.

Include:

- Document Table.
- JSON Inspector.
- Query Results.
- Facet Viewer.
- Ranking Diagnostics.
- Error Details.
- Export UI.

Test content containing:

- HTML markup.
- Script-like strings.
- HTML event attributes.
- Malformed Unicode.
- Control characters.
- Extremely long strings.

Do not interpret document values as trusted HTML.

Avoid unsafe rendering APIs unless content is properly sanitized using a justified approach.

Preserve ordinary text rendering.

### 6.2.6 Audit Document Mutations

Verify that:

- Browsing never triggers writes.
- Search never triggers writes.
- Create, Update, Replace, and Delete have distinct semantics.
- Every mutation requires explicit user intent.
- Mutation targets are immutable once confirmed.
- Primary-key identity is preserved.
- Destructive actions are clearly labeled.
- Task acceptance is not presented as task success.
- Ambiguous submission failures are not automatically retried.

Test attempts to switch connections during confirmation and submission.

Verify that no write can be redirected accidentally.

### 6.2.7 Audit Concurrent Modification Handling

Test situations where another client modifies a document while it is being edited.

Verify:

- The editor preserves the original snapshot.
- Remote changes are detected where practical.
- The user receives a conflict warning.
- Overwrites are not silent.
- The limitations of non-atomic conflict detection are documented.

Do not claim transaction-level protection if the server does not provide it.

### 6.2.8 Audit Task Recovery

Test:

- Task success.
- Task failure.
- Task cancellation.
- Missing Task UID.
- Task status permission denied.
- Server unavailable during polling.
- Polling timeout.
- Connection switching.
- Application shutdown during tracking.
- Restart with pending task references.

Avoid uncontrolled background polling.

Preserve enough task identity information for recovery where supported.

Do not confuse stopping local monitoring with canceling a server-side task.

### 6.2.9 Audit Local Configuration Recovery

Test:

- Missing configuration files.
- Invalid JSON.
- Empty files.
- Truncated writes.
- Unsupported schema versions.
- Invalid enum values.
- Invalid saved connection references.
- Corrupted Query History.
- Corrupted Saved Queries.
- Unwritable settings directories.

Preserve recoverable user data where possible.

Use atomic writes where practical.

Never silently reset all application state because one optional setting is malformed.

Provide a useful recovery message.

### 6.2.10 Audit Export Security

Review JSON and CSV export.

Verify:

- Explicit export scope.
- Safe file paths.
- Filename sanitization.
- Native save dialog behavior.
- Overwrite confirmation.
- Temporary file cleanup.
- File permission errors.
- Disk-full behavior.
- Export cancellation.
- JSON fidelity.
- CSV quoting.
- Spreadsheet formula injection policy.

Test filenames containing invalid or special characters.

Do not allow user-controlled index names to become unrestricted filesystem paths.

Do not silently truncate exported data.

### 6.2.11 Validate Network Failure Recovery

Test:

- Server stopped during browsing.
- Server stopped during searching.
- Server stopped during mutation submission.
- Server restarted.
- DNS resolution failure.
- Connection refused.
- TLS failure.
- HTTP 401.
- HTTP 403.
- HTTP 429.
- HTTP 500.
- Slow responses.
- Connection reset.
- Request cancellation.

Ensure the UI remains responsive.

Users must be able to reconnect or retry safe operations.

Do not automatically retry mutations when the submission outcome is unknown.

### 6.2.12 Audit Dependency Security

Review:

- Go modules.
- npm dependencies.
- Wails framework version.
- GitHub Actions dependencies.
- Third-party editor and table libraries.
- Credential storage libraries.

Use established vulnerability scanning tools appropriate to the project, such as Go vulnerability analysis and npm dependency auditing.

Triage findings based on actual use and exploitability.

Do not automatically upgrade every dependency without compatibility testing.

For necessary changes:

1. Identify the issue.
2. Determine whether MeiliScope is affected.
3. Apply an appropriate fix.
4. Execute regression tests.
5. Record remaining risk.

### 6.2.13 Implement Security Regression Tests

Add tests for:

- Credential redaction.
- Credential storage failures.
- HTTP redirect handling.
- TLS validation.
- Untrusted HTML rendering.
- Malicious document field values.
- Mutation target isolation.
- Duplicate mutation prevention.
- Task uncertainty.
- Configuration corruption.
- Path sanitization.
- CSV injection behavior.
- Oversized responses.

Use controlled fixtures.

Do not require production credentials.

### 6.2.14 Document Known Security Limitations

Record:

- Local data storage behavior.
- Query History privacy.
- Saved Queries privacy.
- Credential storage prerequisites.
- Exported file sensitivity.
- Concurrent modification limitations.
- Task status permission limitations.
- Supported trust assumptions.

Avoid claiming that the application guarantees complete protection of user data.

## Deliverables

- `docs/architecture/security.md`
- Security review report.
- Fixed security defects.
- Security regression tests.
- Improved error recovery.
- Updated security documentation.
- Dependency vulnerability assessment.

## Acceptance Criteria

- No known Critical security defect remains.
- Credentials are protected according to the documented threat model.
- Untrusted document content is rendered safely.
- Mutation operations cannot silently change targets.
- Ambiguous task outcomes are handled accurately.
- Configuration corruption is recoverable where practical.
- Export operations follow documented safety rules.
- Network failures do not cause application-wide crashes under tested conditions.
- Security findings are documented and reproducible.

---

# Step 6.3 — Performance, Stability & Release Readiness

## Objective

Validate MeiliScope's performance and stability under realistic and adverse workloads.

Establish objective quality gates and determine whether the application is ready for Phase 7.

## Tasks

### 6.3.1 Define the Performance Test Environment

Document:

- Operating system and version.
- CPU.
- Memory.
- Storage type.
- MeiliScope revision.
- Go version.
- Wails version.
- WebView runtime.
- Meilisearch version.
- Meilisearch deployment environment.
- Network conditions.
- Dataset characteristics.

Record whether tests use local or remote Meilisearch.

Avoid comparing measurements from unrelated environments without qualification.

### 6.3.2 Create Reproducible Test Data

Prepare representative test datasets.

Suggested scales:

| Dataset | Document Count | Purpose |
|---|---:|---|
| Small | 1,000 | Basic behavior |
| Medium | 10,000 | Normal browsing |
| Large | 100,000 | Pagination and responsiveness |
| Stress | 1,000,000 | Optional large-scale evaluation |

Also include specialized documents containing:

- Hundreds of fields.
- Deeply nested objects.
- Large arrays.
- Long text fields.
- Unicode and Japanese text.
- Large numeric values.
- Mixed and missing fields.

Use a disposable Meilisearch instance.

Document dataset generation and setup procedures.

Avoid storing enormous generated fixtures directly in Git.

### 6.3.3 Define Performance Metrics

Measure:

**Startup**
- Application startup time.
- Time until the main UI becomes responsive.

**Document Browser**
- Document page retrieval time.
- JSON decoding time.
- Table rendering time.
- Page navigation latency.

**JSON Inspector**
- Time to open a document.
- Expand/collapse responsiveness.
- Large JSON rendering performance.

**Query Console**
- Client-observed query duration.
- Server-reported processing time.
- Result decoding overhead.
- Result rendering time.

**Resource Usage**
- Application process memory.
- Related WebView process memory where measurable.
- CPU usage.
- Memory growth over repeated operations.
- File descriptor or handle growth.

Distinguish cold-start and warm-start measurements.

Do not treat all WebView resource usage as part of the Go process.

### 6.3.4 Establish Baseline Measurements

Run tests on the existing implementation before making performance changes.

Record:

- Test command or scenario.
- Dataset size.
- Iteration count.
- Median or comparable representative latency.
- High-percentile latency where useful.
- Peak memory usage.
- Observed failures.

Use repeatable procedures.

Do not invent performance numbers.

If the environment is unsuitable for reliable timing measurements, record that limitation.

### 6.3.5 Validate Large Document Browsing

Test:

- Large index enumeration.
- Document pagination.
- Rapid page switching.
- Column resizing.
- Column visibility changes.
- Repeated row selection.
- Dynamic field discovery.
- Large cell values.

Verify that ordinary browsing does not load an entire index.

Check for:

- UI freezes.
- Excessive re-renders.
- Unbounded caches.
- Unexpected memory growth.
- Stale response display.

### 6.3.6 Validate JSON Inspector Performance

Test documents containing:

- Deep nesting.
- Large arrays.
- Thousands of fields.
- Large strings.
- Repeated nested structures.

Verify:

- Lazy or bounded node rendering where practical.
- Predictable expand/collapse behavior.
- Correct JSON types.
- No stack overflow.
- No application crash.
- Controlled memory usage.

Do not unnecessarily duplicate complete JSON payloads.

### 6.3.7 Validate Query Console Performance

Test:

- Repeated identical queries.
- Rapid query changes.
- Large result sets.
- Facet-heavy responses.
- Ranking score details.
- Switching Table/JSON views.
- Query cancellation.
- Connection switching during execution.

Verify that:

- Old requests do not overwrite newer results.
- Search metadata remains accurate.
- Repeated execution does not leak resources.
- Query History remains bounded.

### 6.3.8 Validate Export Performance

Test:

- Small JSON exports.
- Medium CSV exports.
- Large current-page exports.
- Nested JSON serialization.
- Unicode output.
- Disk write failures.
- Export cancellation.
- Repeated exports.

If optional full-result export exists, test bounded memory usage and incremental writing.

Do not add full-result export solely for performance testing.

Measure output size and memory behavior.

### 6.3.9 Validate Task Tracking Stability

Test:

- Repeated document mutations.
- Multiple tracked tasks.
- Slow task completion.
- Task failure.
- Polling timeout.
- Server disconnection.
- Application shutdown while monitoring.

Verify that:

- Polling is bounded.
- Duplicate pollers are not created accidentally.
- Completed tasks stop polling.
- Task data remains associated with the correct connection.
- The application does not accumulate abandoned background operations.

### 6.3.10 Conduct Long-running Stability Tests

Perform sustained operation testing.

Suggested scenarios:

**Idle Stability**
- Leave the application running without interaction.
- Observe memory and CPU behavior.

**Repeated Browsing**
- Alternate between indexes.
- Navigate document pages.
- Open and close JSON Inspector details.

**Repeated Search**
- Execute searches continuously at a controlled rate.
- Switch result views.
- Refresh and cancel requests.

**Connection Churn**
- Connect and disconnect repeatedly.
- Switch saved profiles.
- Restart the test server.

**Mixed Workload**
- Browse documents.
- Execute searches.
- Inspect results.
- Save queries.
- Export small datasets.

Use repeatable automation where practical.

A representative initial soak test may run for one hour, with longer tests when resources permit.

Do not claim long-running stability based only on a brief smoke test.

Record the actual duration.

### 6.3.11 Investigate Memory Growth

Observe memory usage across repeated workload cycles.

Distinguish:

- Expected runtime allocation.
- Go garbage collector behavior.
- WebView process allocation.
- Cached application data.
- Genuine unbounded growth.

Use Go profiling and frontend development tools where practical.

Do not label normal garbage collector behavior as a memory leak without evidence.

If suspicious growth occurs:

1. Reproduce it.
2. Isolate the responsible component.
3. Capture diagnostic evidence.
4. Implement a focused correction.
5. Repeat the workload.
6. Compare results under equivalent conditions.

### 6.3.12 Investigate UI Responsiveness

Review:

- Synchronous frontend work.
- Excessive React rendering.
- Expensive JSON formatting.
- Unbounded object traversal.
- Repeated state updates.
- Blocking backend calls.
- Excessive bridge payloads.

Prefer targeted optimization based on measurements.

Avoid broad rewrites.

Keep user-facing operations cancelable where appropriate.

### 6.3.13 Validate Windows

Test the production build on a representative Windows environment.

Verify:

- Startup.
- UI rendering.
- WebView2 behavior.
- Credential storage.
- Network operations.
- File dialogs.
- Clipboard.
- JSON Editor.
- Document Browser.
- Query Console.
- Document Editor.
- Export.
- Application shutdown.
- Restart behavior.

Record the Windows version and test environment.

### 6.3.14 Validate Linux

Test the production build on a representative Linux environment.

Verify:

- Startup.
- GTK/WebKit runtime behavior.
- Credential storage.
- Session-only credentials.
- File dialogs.
- Clipboard.
- Keyboard shortcuts.
- JSON Editor.
- Document Browser.
- Query Console.
- Document Editor.
- Export.
- Application shutdown.

Record:

- Distribution and version.
- Desktop environment.
- Display server where relevant.
- GTK/WebKit versions.
- Credential service availability.

Do not claim support for untested distributions.

### 6.3.15 Review Application Resource Cleanup

Verify cleanup of:

- HTTP connections.
- Canceled requests.
- Timers and polling loops.
- File handles.
- Temporary export files.
- Event listeners.
- Frontend subscriptions.
- Background operations.
- Application lifecycle resources.

Ensure shutdown does not hang indefinitely.

Do not introduce excessive forced termination logic that could corrupt persistent data.

### 6.3.16 Establish Performance Acceptance Criteria

Define reasonable, measurable thresholds using the baseline and expected product workloads.

Avoid arbitrary universal performance guarantees.

At minimum, establish release gates for:

- No reproducible application crash during required workflows.
- No reproducible unbounded memory growth.
- No UI deadlock during required workflows.
- No full-index retrieval during ordinary paginated browsing.
- No duplicate uncontrolled background polling.
- No unbounded Query History or Saved Query cache growth.
- No major unexplained regression against baseline measurements.

Additional numeric thresholds may be introduced when supported by measurements.

Document thresholds and the tested hardware.

### 6.3.17 Improve Automated Quality Gates

Review CI performance and reliability.

Ensure required workflows include:

- Backend tests.
- Frontend tests.
- Static analysis.
- Security checks.
- Windows build.
- Linux build.
- Core integration tests where practical.

Use separate workflows for expensive stress tests.

Recommended distinction:

**Pull Request CI**
- Fast validation.
- Unit and component tests.
- Static analysis.
- Production builds.

**Scheduled or Manual Validation**
- Compatibility matrix.
- Security regression suite.
- Performance benchmarks.
- Long-running stability tests.

Do not make an excessively expensive stress workload mandatory for every small pull request.

### 6.3.18 Generate a Release-readiness Matrix

Create:

`docs/development/release-readiness.md`

Include:

| Area | Required Evidence | Status |
|---|---|---|
| Windows Build | Production build and smoke test | Pending |
| Linux Build | Production build and smoke test | Pending |
| API Compatibility | Verified compatibility matrix | Pending |
| Core Features | End-to-end regression tests | Pending |
| Credential Security | Security tests and audit | Pending |
| Mutation Safety | Confirmed write workflows | Pending |
| Failure Recovery | Recovery test results | Pending |
| Performance | Recorded benchmark results | Pending |
| Stability | Recorded sustained testing | Pending |
| CI | Required checks passing | Pending |
| Documentation | Updated technical and user guides | Pending |

Replace the placeholder statuses with actual verification results.

Do not mark items as complete without evidence.

### 6.3.19 Update Documentation

Update relevant documentation.

Recommended files:

```text id="j4qo7i"
docs/development/compatibility-matrix.md
docs/development/performance.md
docs/development/release-readiness.md
docs/development/phase-6-report.md

docs/architecture/security.md
docs/architecture/overview.md

docs/en/getting-started/
docs/ja/getting-started/

README.md
```

Document:

- Supported Meilisearch versions.
- Tested Windows and Linux environments.
- Runtime prerequisites.
- Known security limitations.
- Performance findings.
- Known defects.
- Release blockers.
- Reproduction procedures.

Update `plan.md` when verified findings materially change the technical requirements, supported platforms, compatibility policy, or release assumptions.

Do not rewrite unrelated product goals.

### 6.3.20 Final Phase Report

Create:

`docs/development/phase-6-report.md`

Include:

1. Phase objectives.
2. Completed implementation and validation work.
3. API compatibility matrix.
4. Regression testing results.
5. Security findings.
6. Failure recovery findings.
7. Performance baselines.
8. Stress testing results.
9. Long-running stability results.
10. Windows validation.
11. Linux validation.
12. CI quality gates.
13. Defects discovered.
14. Defects fixed.
15. Remaining issues.
16. Release-blocking issues.
17. Known limitations.
18. Phase 7 readiness decision.

## Acceptance Criteria

- Required regression tests pass.
- API compatibility is documented accurately.
- Security review is complete.
- No unresolved Critical or High release-blocking defects remain.
- Core failure recovery scenarios are validated.
- Performance baselines are recorded.
- Required stability tests have been executed.
- No significant unresolved resource leak or deadlock is known.
- Windows and Linux production builds are validated.
- CI quality gates pass.
- Remaining limitations are explicitly documented.
- Phase 7 readiness is supported by evidence.

---

# 4. General Implementation Constraints

## 4.1 Do Not Add Major Features

Do not implement:

- Additional Meilisearch administration tools.
- Index creation or deletion UI.
- API Key administration UI.
- Generic HTTP Request Console.
- Full-index backup or synchronization.
- Automatic application updates.
- Cloud synchronization.
- Multi-user collaboration.
- New supported platforms without a separate decision.

Phase 6 is for quality improvement.

## 4.2 Avoid Unnecessary Refactoring

Do not redesign the architecture simply because improvements are possible.

Focus on demonstrable correctness, security, performance, or maintenance issues.

Preserve existing working behavior.

## 4.3 Do Not Weaken Security

Do not:

- Disable TLS verification.
- Store API Keys in plaintext.
- Expose credentials to the frontend.
- Remove destructive-operation confirmation.
- Automatically retry uncertain mutations.
- Suppress important API errors.
- Hide task failures.

## 4.4 Protect Test Environments

All destructive integration tests must run against disposable test resources.

Never select a default user connection for write tests.

Require explicit test-instance configuration and use unique test resource identifiers.

Fail safely if the configured target cannot be verified as a test environment.

Do not use the application's saved production credentials in automated tests.

## 4.5 Preserve Cross-platform Compatibility

Changes must not break Windows or Linux builds.

Keep platform-dependent implementations isolated.

Do not add a dependency requiring a new runtime installation without documenting and justifying the change.

## 4.6 Preserve JSON Fidelity

Do not introduce numeric precision loss.

Do not silently discard unknown fields.

Do not change JSON mutation semantics during optimization.

Do not truncate exports without informing the user.

## 4.7 Maintain Test Independence

Unit tests must not require a live Meilisearch server.

Integration tests may require a disposable server.

Desktop tests may require a graphical session.

Make these requirements explicit.

## 4.8 Avoid Unsupported Quality Claims

Do not claim:

- Zero memory leaks.
- Universal Linux compatibility.
- Compatibility with every Meilisearch version.
- Complete security.
- Guaranteed conflict-free editing.
- Performance characteristics not measured.

Report only verified results.

---

# 5. Execution & Reporting Requirements

Execute Phase 6 in this order:

1. Step 6.1 — API Compatibility & Regression Testing.
2. Step 6.2 — Security & Failure Recovery.
3. Step 6.3 — Performance, Stability & Release Readiness.

Each step must produce independently reviewable changes and evidence.

After each step, provide:

### Completed Work

Describe the validation and corrective work performed.

### Defects Found

List confirmed issues with severity and reproduction information.

### Fixes Implemented

Describe the corrections and corresponding regression tests.

### Files Changed

List important source, test, workflow, and documentation changes.

### Tests Executed

Provide actual commands and results.

### Environment

Describe the relevant operating systems, toolchain versions, and server versions.

### Remaining Issues

List open defects, blockers, and deferred work.

### Next Step Readiness

State whether the next step can proceed.

Do not claim success for work that was not executed.

---

# 6. Phase 6 Definition of Done

Phase 6 is complete when:

1. The supported Meilisearch versions are documented.
2. Core workflows have regression coverage.
3. Windows and Linux builds pass.
4. Core desktop workflows are validated on both platforms.
5. Credential handling has been reviewed.
6. Mutation safety has been verified.
7. Network failure recovery has been tested.
8. Local configuration recovery has been tested.
9. Query and Document data remain isolated between connections.
10. Large document browsing remains responsive under documented test conditions.
11. JSON Inspector handles representative large structures.
12. Repeated searches do not cause uncontrolled resource growth.
13. Task polling terminates appropriately.
14. Export operations behave safely.
15. Performance baselines are recorded.
16. Sustained operation testing is documented.
17. No unresolved Critical or High release-blocking defects remain.
18. CI quality gates pass.
19. Documentation matches verified behavior.
20. The release-readiness report explicitly approves progression to Phase 7.

If required evidence is unavailable, report the exact blocker instead of marking the corresponding criterion complete.

**Successful completion of Phase 6 establishes an evidence-based quality baseline for MeiliScope and prepares the project for Phase 7 — Release.**
