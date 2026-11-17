# MeiliScope — Phase 0 Implementation Instructions

## Foundation & Technology Validation

### 1. Project Context

MeiliScope is an open-source, Windows-first desktop application for exploring Meilisearch indexes and documents, executing search queries, and performing limited document editing.

The project's mission is:

**"Explore your Meilisearch data."**

The project proposal is maintained in `plan.md` at the repository root.

Read the entire project proposal before making architectural decisions or implementing code.

The primary requirements are:

- Go as the primary backend language.
- Windows x64 as the initial supported platform.
- A standalone desktop application.
- No separately installed Go, Python, Node.js, or .NET runtime required by end users.
- No separate application server or background service.
- Meilisearch communication handled by the Go backend.
- React and TypeScript as the preferred frontend stack.
- MIT License.
- GitHub as the source repository.
- GitHub Actions for CI/CD.
- GitHub Releases for distribution.
- ZIP-based distribution, preferably containing a single application executable.

Wails v3 is currently the preferred GUI framework, but its adoption must be validated before being finalized.

**Phase 0 must not assume that Wails v3 is suitable merely because it can build and launch a sample application.**

The framework must be evaluated against MeiliScope's actual development and distribution requirements.

---

## 2. Phase 0 Objectives

Phase 0 has five primary objectives:

1. Evaluate Wails v3 against the project's technical and distribution requirements.
2. Make and document an evidence-based GUI framework decision.
3. Update the project proposal and architecture documents to reflect the decision.
4. Establish a maintainable application foundation.
5. Validate a reproducible Windows x64 build and distribution workflow.

Phase 0 must produce a working desktop application shell, not a complete Meilisearch client.

Do not implement Connection Manager, Index Explorer, Document Browser, Query Console, or Document Editor business functionality during this phase.

These features belong to subsequent phases.

## 3. Execution Rules

Execute Steps 0.1 through 0.7 in order.

Each step must have independently verifiable deliverables and acceptance criteria.

Use the following rules throughout Phase 0:

- Inspect existing files and preserve valid existing implementations.
- Use official documentation and verified library APIs.
- Prefer stable, versioned dependencies.
- Avoid unnecessary architectural abstractions.
- Do not introduce additional runtime services.
- Do not expose implementation details or credentials unnecessarily.
- Add automated tests for testable functionality.
- Keep the application buildable after each implementation step.
- Update documentation when implementation changes invalidate it.
- Do not silently change the product scope defined in `plan.md`.
- Do not claim that a test passed unless it was actually executed.
- Do not claim that a clean-machine validation passed if no clean environment was available.
- Record assumptions, limitations, and unverified conditions.

If a step is blocked, report the blocker and preserve all verified results.

### Important Decision Gate

Step 0.1 is a mandatory architecture decision gate.

Do not proceed with the production application structure until the GUI framework decision has been documented.

If Wails v3 fails a mandatory requirement, investigate whether the failure can be resolved without compromising the product's goals.

If it cannot, recommend an alternative with a justified technical comparison.

Do not replace Wails v3 automatically without documenting the decision.

---

# Step 0.1 — Wails v3 Feasibility Study

## Objective

Determine whether Wails v3 is suitable for MeiliScope as a Windows-first, standalone desktop application.

The evaluation must use a real, minimal proof-of-concept application.

## Tasks

### 0.1.1 Verify Current Wails v3 Status

Consult the official Wails documentation and GitHub repository.

Verify:

- Current stable release and version.
- Go toolchain requirements.
- Windows SDK and compiler requirements.
- Node.js requirements for frontend development.
- Windows WebView2 requirements.
- Supported Windows versions and architectures.
- React and TypeScript integration.
- Go-to-JavaScript communication mechanisms.
- Production build procedure.
- Packaging and distribution options.
- Known issues relevant to Windows deployment.

Do not use outdated Wails v2 examples or APIs.

Record exact versions, documentation references, and verification dates.

Recommended primary references:

- https://v3.wails.io/
- https://v3.wails.io/getting-started/installation/
- https://github.com/wailsapp/wails
- https://github.com/wailsapp/wails/releases

### 0.1.2 Create a Minimal Proof of Concept

Create a temporary Wails v3 application with:

- Go backend.
- React frontend.
- TypeScript.
- A native desktop window.
- A basic frontend view.
- A backend method callable from the frontend.
- Structured error propagation.
- A simple asynchronous operation.

The proof of concept should demonstrate:

1. Application startup.
2. Frontend rendering.
3. Frontend-to-Go communication.
4. Go-to-frontend response handling.
5. Error handling.
6. Application shutdown.
7. Production executable generation.

Prefer keeping this implementation under an isolated `experiments/wails-v3/` directory until the technology decision is finalized.

### 0.1.3 Validate Windows Distribution

Produce a release-mode Windows x64 executable.

Investigate:

- Whether the frontend assets can be embedded in the executable.
- Whether additional runtime files are required.
- Whether WebView2 is installed or provisioned automatically.
- Behavior when WebView2 is unavailable.
- Whether users need administrator privileges.
- Whether the application requires an installer.
- Whether a portable ZIP distribution is feasible.
- Whether the executable can launch without network access.
- Whether application startup requires any additional local services.
- Whether any development-time dependencies remain necessary at runtime.

Distinguish between:

- Build-time dependencies.
- Operating-system dependencies.
- Deployment prerequisites.
- Runtime dependencies shipped with the application.

A single EXE is preferred, but must not be achieved at the expense of unreliable deployment.

WebView2 must be explicitly documented as an operating-system runtime prerequisite if it is required.

Test on a separate clean Windows environment when available.

If such an environment is unavailable, mark the result as unverified.

### 0.1.4 Evaluate Application Requirements

Investigate whether Wails v3 supports the following requirements without major workarounds:

- Responsive, resizable application windows.
- Large data tables.
- JSON editors.
- Keyboard shortcuts.
- Clipboard operations.
- File open/save dialogs.
- Native notifications or equivalent UI feedback.
- Asynchronous backend operations.
- Context cancellation or equivalent cancellation support.
- Reliable application shutdown.
- Windows credential storage integration through Go.
- Frontend assets embedded in the production application.

Actual proof-of-concept testing is required for core requirements.

Secondary capabilities may be evaluated through documentation and small targeted tests.

### 0.1.5 Evaluate Maintenance Risks

Assess:

- Framework API stability.
- Dependency maintenance.
- Build reproducibility.
- CI compatibility.
- Windows-specific limitations.
- Frontend/backend coupling.
- Debugging experience.
- Likely migration costs.

Avoid speculative risk scoring without supporting evidence.

### 0.1.6 Make the Framework Decision

Classify the outcome as one of:

**ACCEPT**

Wails v3 satisfies the mandatory requirements and has no unresolved critical deployment blockers.

**ACCEPT WITH CONDITIONS**

Wails v3 is suitable, but specific limitations or prerequisites must be documented and addressed before release.

**REJECT**

Wails v3 fails one or more mandatory requirements without a practical mitigation.

If rejected, provide a focused comparison with viable alternatives. Do not implement a replacement framework as part of this feasibility study without an explicit, documented architectural decision.

## Deliverables

Create:

`docs/architecture/wails-v3-evaluation.md`

Include:

- Tested versions.
- Test environment.
- Commands used.
- Evaluation matrix.
- Proof-of-concept results.
- Windows runtime dependencies.
- Build and packaging findings.
- Performance and stability observations.
- Known limitations.
- Evidence references.
- Final recommendation.

Preserve the proof-of-concept source code when useful for reproducibility.

## Acceptance Criteria

- A minimal GUI application has been built and executed, or the exact blocker has been documented.
- Frontend/backend communication has been tested.
- Windows production build requirements have been verified.
- Packaging dependencies have been identified.
- The result is explicitly classified as ACCEPT, ACCEPT WITH CONDITIONS, or REJECT.
- No unverified assumption is presented as a successful test.

---

# Step 0.2 — Architecture Decision & Proposal Update

## Objective

Update MeiliScope's project proposal and technical architecture to reflect the actual results of Step 0.1.

The project proposal must remain the authoritative product-level document.

## Tasks

### 0.2.1 Record the Architecture Decision

Create:

`docs/architecture/adr/0001-desktop-framework.md`

Include:

- Context.
- Requirements.
- Evaluated alternatives.
- Technical findings.
- Decision.
- Consequences.
- Known limitations.
- Reconsideration criteria.

Use Wails v3 if the feasibility study supports its adoption.

Do not retain the statement that Wails v3 is merely an unverified candidate after a validated adoption decision.

### 0.2.2 Update the Project Proposal

Modify the existing root-level `plan.md`.

Do not replace the entire proposal with an unrelated document.

Preserve:

- Project mission.
- Product positioning.
- Core feature priorities.
- Windows-first strategy.
- MIT License.
- GitHub distribution strategy.
- MVP definition.
- Non-goals.

Review and update the relevant sections, including:

- Section 9: Technical Approach.
- Section 9.1: Technology Stack.
- Section 9.2: Architecture.
- Section 10: Distribution and Installation.
- Section 12: CI/CD and Package Management.
- Section 14: Non-functional Requirements.
- Section 16: Development Roadmap.
- Section 17: Technical Risks.
- Section 19: Outstanding Decisions.

If Wails v3 is adopted, document the selected version and verified requirements.

If adoption is conditional, explicitly identify the unresolved release prerequisites.

If Wails v3 is rejected, record the replacement decision and its implications before changing the technology stack.

### 0.2.3 Align the Roadmap

Update Section 16 of `plan.md` to reflect the implementation sequence:

- Phase 0 — Foundation & Technology Validation.
- Phase 1 — Connection & Index Explorer.
- Phase 2 — Document Browser.
- Phase 3 — Query Console.
- Phase 4 — Document Editing.
- Phase 5 — Usability.
- Phase 6 — Quality & Security.
- Phase 7 — Release.

Maintain the Phase 3 MVP milestone.

Ensure that the roadmap does not contradict the feature priorities or release scope in the rest of the proposal.

### 0.2.4 Document Technical Boundaries

Establish architectural rules:

- Go owns HTTP communication with Meilisearch.
- The frontend owns presentation and user interaction.
- Application services must not depend unnecessarily on GUI components.
- The Meilisearch API client must be independently testable.
- Credentials must not be exposed to the frontend.
- Settings and credentials must have separate storage responsibilities.
- Application startup must not require a separately managed local server.

## Deliverables

- `docs/architecture/adr/0001-desktop-framework.md`
- Updated `plan.md`
- `docs/architecture/overview.md`

## Acceptance Criteria

- The project proposal reflects the validated technical decision.
- There are no contradictory framework requirements.
- Build and runtime dependencies are accurately described.
- The architecture boundaries are documented.
- Remaining unknowns are explicitly identified.
- The updated roadmap matches the approved product scope.

---

# Step 0.3 — Project Scaffolding

## Objective

Establish the production application structure using the GUI framework selected in Step 0.2.

The following instructions assume Wails v3 was accepted. If another framework was selected, adapt only framework-specific implementation details while preserving the architectural requirements.

## Tasks

### 0.3.1 Initialize the Application

Create the production application with:

- Go backend.
- React.
- TypeScript.
- A modern frontend build system supported by the selected framework.
- Reproducible dependency installation.
- Separate development and production build configurations.

Pin dependency versions appropriately.

Do not use a moving development branch when a suitable stable release exists.

### 0.3.2 Establish Repository Structure

Use a simple, maintainable structure such as:

```text
MeiliScope/
├── cmd/
│   └── meiliscope/
├── internal/
│   ├── app/
│   ├── services/
│   ├── meilisearch/
│   ├── config/
│   └── platform/
├── frontend/
│   ├── src/
│   │   ├── app/
│   │   ├── components/
│   │   ├── features/
│   │   ├── services/
│   │   └── types/
│   └── package.json
├── build/
├── docs/
│   ├── architecture/
│   ├── development/
│   ├── en/
│   └── ja/
├── .github/
│   └── workflows/
├── go.mod
├── plan.md
├── LICENSE
└── README.md
```

Adapt the structure when required by the framework.

Do not create empty abstraction layers solely to match the suggested tree.

### 0.3.3 Define Module Boundaries

Establish clear responsibilities:

**Application Layer**

Application lifecycle and coordination.

**Service Layer**

Business operations exposed to the frontend.

**Meilisearch Layer**

Future HTTP API communication.

**Configuration Layer**

Non-secret application settings.

**Platform Layer**

Operating-system-specific integrations.

**Frontend**

UI components, routing, presentation state, and user interaction.

Do not implement future Meilisearch features.

### 0.3.4 Establish Development Commands

Provide documented commands for:

- Installing dependencies.
- Launching development mode.
- Running Go tests.
- Running frontend checks.
- Building the production executable.
- Cleaning generated artifacts.

Use platform-appropriate scripts when necessary.

Keep Windows development straightforward.

## Deliverables

- Production project structure.
- Build configuration.
- Dependency lockfiles.
- `.gitignore`.
- Development scripts.
- `docs/development/build.md`.

## Acceptance Criteria

- The application starts in development mode.
- The production build succeeds.
- Dependency installation is reproducible.
- The repository structure is documented.
- No future-phase business functionality has been added.

---

# Step 0.4 — Application Shell & Backend Bridge

## Objective

Create the basic desktop user interface and communication boundary between React and Go.

## Tasks

### 0.4.1 Implement the Application Shell

Create a desktop window with:

- Application title: MeiliScope.
- A main toolbar.
- A left navigation area.
- A central content area.
- An optional inspector area.
- A status bar.

Provide placeholder views for:

- Connections.
- Index Explorer.
- Document Browser.
- Query Console.

These placeholders must not communicate with a real Meilisearch instance.

### 0.4.2 Implement Basic Navigation

Support navigation between the placeholder views.

Maintain navigation state consistently.

Avoid introducing a complex routing framework unless it provides a clear benefit.

### 0.4.3 Implement Frontend/Backend Communication

Create a small backend service exposing operations such as:

- GetApplicationInfo.
- GetApplicationVersion.
- GetRuntimeInformation.

Use the framework's supported binding mechanism.

Prefer generated TypeScript types when available.

Avoid manual duplication of complex backend data types.

### 0.4.4 Establish Asynchronous Operation Patterns

Demonstrate:

- Successful backend calls.
- Backend errors.
- Asynchronous responses.
- Loading state.
- Cancellation behavior where supported.
- UI recovery after errors.

A failed backend call must not freeze or terminate the application.

### 0.4.5 Define UI Conventions

Establish basic conventions for:

- Loading indicators.
- Error messages.
- Empty states.
- Application status.
- Consistent layout spacing.
- Keyboard accessibility.

Do not over-engineer the UI component library during this phase.

## Deliverables

- Functional application shell.
- Navigation placeholders.
- Go backend service.
- TypeScript bindings.
- Basic UI error handling.
- Tests where practical.

## Acceptance Criteria

- The application window launches.
- Navigation between placeholders works.
- Frontend-to-Go calls succeed.
- Backend errors are visible and recoverable.
- No external Meilisearch server is required.

---

# Step 0.5 — Settings & Application Infrastructure

## Objective

Establish reliable local application settings and foundational infrastructure.

## Tasks

### 0.5.1 Implement Configuration Storage

Use the Windows-appropriate per-user application configuration directory.

Do not store user configuration beside the executable by default.

Support:

- Initial configuration creation.
- Configuration loading.
- Configuration saving.
- Schema versioning.
- Default values.
- Corrupted configuration detection.
- Safe recovery behavior.

Prefer atomic file replacement for updates.

Do not silently discard corrupted settings without preserving diagnostic information.

### 0.5.2 Define Initial Settings

Start with a minimal schema:

- Configuration version.
- Window dimensions.
- Window position where appropriate.
- Selected application view.
- Optional UI preferences.

Do not store API Keys in this configuration.

Do not introduce a separate database for simple application settings.

### 0.5.3 Define Credential Storage Interface

Prepare a narrow credential-storage abstraction for Phase 1.

The Windows implementation must eventually use an appropriate OS-secured credential mechanism.

Do not implement insecure plaintext credential persistence as a temporary shortcut.

Credential management may remain unimplemented in Phase 0, provided that the interface and planned implementation are documented.

### 0.5.4 Implement Application Logging

Provide structured diagnostic logging with:

- Log levels.
- Configurable diagnostic verbosity.
- Clear error categories.
- Safe log storage.
- No accidental secret disclosure.

Do not log sensitive request headers or credentials.

Avoid excessive logging in normal operation.

### 0.5.5 Handle Application Lifecycle

Ensure that:

- Settings can be saved during normal shutdown.
- Window closing is handled consistently.
- Application startup handles invalid settings.
- Background operations do not prevent shutdown indefinitely.
- Unexpected errors do not silently destroy persistent settings.

## Deliverables

- Configuration implementation.
- Credential interface.
- Logging infrastructure.
- Lifecycle integration.
- Unit tests.

## Acceptance Criteria

- Settings persist between launches.
- Corrupted settings are handled safely.
- Settings are independent of the executable directory.
- No credentials are stored in plaintext.
- Configuration tests pass.

---

# Step 0.6 — Testing & GitHub Actions CI

## Objective

Create a reliable automated quality and build pipeline.

## Tasks

### 0.6.1 Go Validation

Configure:

- `gofmt` checks.
- `go vet`.
- Unit tests.
- Dependency verification.

Enable the race detector for supported test configurations where practical.

### 0.6.2 Frontend Validation

Configure:

- TypeScript type checking.
- ESLint or an equivalent lint system.
- Frontend unit tests where useful.
- Production frontend build.

Use reproducible package installation with a committed lockfile.

### 0.6.3 Integration Testing

Test the application service boundary without requiring a live Meilisearch server.

Prefer mock implementations for future external dependencies.

Do not add a Docker requirement to the default Phase 0 test pipeline.

### 0.6.4 GitHub Actions CI

Create a CI workflow triggered by:

- Pull requests.
- Pushes to the primary development branch.

The workflow must perform:

1. Repository checkout.
2. Required toolchain setup.
3. Dependency installation.
4. Go formatting and static checks.
5. Go unit tests.
6. Frontend lint and type checks.
7. Frontend tests.
8. Frontend production build.
9. Windows x64 desktop build.
10. Build artifact upload.

Use a Windows runner for the Windows application build.

Pin or appropriately constrain tool and action versions to support reproducibility and security.

Use minimal GitHub Actions permissions.

### 0.6.5 Prepare Release Workflow

Create a release workflow supporting version tags.

Generate:

- Windows x64 executable.
- ZIP release package.
- SHA-256 checksum file.

Support GitHub Releases publication when the workflow is triggered by an approved release tag.

Do not create a production version tag or publish an actual public release merely to test the workflow.

Test packaging separately where possible.

### 0.6.6 Dependency Management

Configure Dependabot for:

- Go Modules.
- npm packages.
- GitHub Actions.

Review dependency licensing compatibility with the MIT-licensed project.

## Deliverables

- `.github/workflows/ci.yml`
- `.github/workflows/release.yml`
- `.github/dependabot.yml`
- Test configuration.
- Build scripts.
- CI documentation.

## Acceptance Criteria

- Local validation commands pass.
- Windows x64 CI build succeeds.
- CI artifacts are generated.
- Release packaging can be executed reproducibly.
- No developer-specific absolute paths are required.
- No credentials are embedded in artifacts.

If remote GitHub Actions execution is unavailable, distinguish local workflow validation from actual hosted CI execution.

---

# Step 0.7 — Windows Distribution Validation

## Objective

Validate that the MeiliScope desktop application can be distributed and launched with minimal installation burden.

## Tasks

### 0.7.1 Produce a Release Candidate Package

Create a ZIP package containing:

```text
MeiliScope_v0.1.0-dev_windows_amd64.zip
├── meiliscope.exe
├── LICENSE
└── README.md
```

This layout is the preferred target.

If additional runtime files are unavoidable, document and include them as necessary.

Do not omit required runtime files merely to preserve the preferred layout.

### 0.7.2 Validate Portable Execution

Extract the ZIP into a directory outside the source repository.

Test:

- Executable startup.
- Window rendering.
- Navigation.
- Backend communication.
- Configuration persistence.
- Logging.
- Application shutdown.
- Restart after configuration has been saved.

Verify that no separate application server is required.

### 0.7.3 Validate Runtime Requirements

Where a clean Windows environment is available, verify:

- No Go installation is required.
- No Node.js installation is required.
- No Python installation is required.
- No .NET application runtime installation is required, unless a verified platform dependency makes this unavoidable.
- No Docker installation is required.
- No administrator privileges are needed for normal execution.
- Required WebView2 behavior is understood and documented.

Do not describe a machine as clean if it already contains development tools whose presence may influence the result.

Record the operating system version and relevant installed dependencies.

### 0.7.4 Validate Release Integrity

Verify:

- ZIP contents.
- Executable version information.
- SHA-256 checksums.
- Included license.
- Required documentation.
- No unexpected development artifacts.
- No hardcoded machine-specific paths.
- No test credentials or other secrets.

### 0.7.5 Update Documentation

Create or update:

- `README.md`
- `docs/en/getting-started/installation.md`
- `docs/ja/getting-started/installation.md`
- `docs/development/build.md`

Document:

- Supported Windows versions.
- Installation and launch procedure.
- Runtime prerequisites.
- Build requirements for contributors.
- Known limitations.
- Troubleshooting guidance.

Clearly distinguish requirements for end users from requirements for developers.

### 0.7.6 Finalize Phase 0 Findings

Create:

`docs/development/phase-0-report.md`

Summarize:

- Completed steps.
- Framework decision.
- Verified build environment.
- Runtime dependency findings.
- Tests executed.
- Unverified conditions.
- Known issues.
- Remaining risks.
- Readiness for Phase 1.

Update `plan.md` if the final validation reveals material information that changes the architecture or distribution assumptions.

## Deliverables

- Windows x64 ZIP package.
- SHA-256 checksum.
- Installation documentation.
- Phase 0 completion report.
- Updated project proposal when necessary.

## Acceptance Criteria

- The Windows application builds successfully.
- The release package can be extracted.
- The extracted application launches in a verified environment.
- User configuration survives application replacement.
- Required runtime dependencies are documented.
- No developer toolchain is needed at runtime.
- The application does not depend on a separate MeiliScope service.
- All incomplete or unverified validation items are documented.

---

# 4. Phase 0 Completion Criteria

Phase 0 is complete only when:

1. The GUI framework decision has been validated and documented.
2. The root-level project proposal reflects the decision.
3. The application architecture is documented.
4. A production application shell exists.
5. Frontend/backend communication works.
6. Local settings persist safely.
7. Automated tests and build validation are available.
8. A Windows x64 release package can be generated.
9. Runtime prerequisites and packaging limitations are documented.
10. No unresolved critical blocker prevents Phase 1 implementation.

Do not treat missing verification as successful completion.

If clean-machine deployment validation remains unavailable, explicitly report the limitation and record it as a release prerequisite.

## 5. Required Final Report

At the end of Phase 0, provide a structured report containing:

### Implementation Summary

Describe the completed work for each step.

### Architecture Decision

State whether Wails v3 was accepted, conditionally accepted, or rejected, and why.

### Project Proposal Changes

List the modified sections of `plan.md` and explain significant changes.

### Files Created or Modified

Provide the major files and their responsibilities.

### Validation Results

Report the actual commands executed and their results.

Distinguish among:

- Passed.
- Failed.
- Blocked.
- Not executed.

### Distribution Findings

Explain the exact Windows runtime requirements and validated packaging behavior.

### Known Issues

List unresolved technical issues and their impact.

### Phase 1 Readiness

State whether Phase 1 can begin safely.

If not, identify the specific blockers.

## 6. Scope Restrictions

Do not implement:

- Meilisearch Connection Manager.
- Meilisearch API client business functionality.
- Index Explorer business functionality.
- Document Browser.
- JSON Inspector.
- Query Console.
- Search Builder.
- Document editing.
- API Key management UI.
- Automatic application updates.
- Server administration features.
- Linux or macOS distribution.

Phase 0 is strictly for the application foundation, framework validation, architecture decision, and distribution infrastructure.

Keep the implementation small, modular, testable, and ready for incremental development.

**The final result should be a verified foundation for MeiliScope, not a prematurely implemented application.**
