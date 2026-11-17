# MeiliScope — Phase 7 Implementation Instructions

## Release Engineering, Distribution & Documentation

### 1. Project Context

MeiliScope is an open-source, standalone desktop application for exploring Meilisearch indexes, browsing documents, executing search queries, and performing controlled document modifications.

**Mission: Explore your Meilisearch data.**

The established technology stack includes:

- Go for backend services.
- Wails v3 for desktop integration.
- React and TypeScript for the frontend.
- Meilisearch HTTP API for data access.
- Windows and Linux as validated execution platforms.
- MIT License.
- GitHub Actions for CI/CD.
- GitHub Releases for distribution.

The root-level `plan.md` is the authoritative project proposal.

The previous phases established:

**Phase 0 — Foundation**
- Desktop application architecture.
- Windows and Linux build support.
- Configuration and logging infrastructure.
- Initial CI/CD infrastructure.

**Phase 1 — Connection & Index Explorer**
- Connection management.
- Secure credential storage.
- Meilisearch HTTP client.
- Index browsing.

**Phase 2 — Document Browser**
- Document listing.
- Pagination.
- Dynamic tables.
- JSON Inspector.

**Phase 3 — Query Console**
- Search execution.
- JSON Query Editor.
- Search diagnostics.
- Query History.

**Phase 4 — Document Editing**
- Controlled document creation, modification, and deletion.
- Diff review.
- Task tracking.

**Phase 5 — Usability**
- Search Builder.
- Saved Queries.
- JSON/CSV export.
- Keyboard shortcuts.
- UI/UX improvements.

**Phase 6 — Quality & Security**
- API compatibility testing.
- Security review.
- Failure recovery.
- Performance and stability validation.
- Release-readiness assessment.

Phase 7 is the final release preparation and distribution phase.

Do not implement significant new user-facing functionality during Phase 7.

The purpose of this phase is to produce a reliable, reproducible, documented, and verifiable release of MeiliScope.

---

# 2. Phase 7 Objectives

Phase 7 has three major implementation steps:

1. Step 7.1 — Documentation & OSS Readiness.
2. Step 7.2 — Release Automation & Packaging.
3. Step 7.3 — Distribution Validation & Release Approval.

The expected result is a release-ready version of MeiliScope that can be distributed through GitHub Releases.

The first stable release is intended to use version `v1.0.0`, provided the release requirements are satisfied.

If the repository's actual development history or release policy requires an earlier prerelease version, document the recommendation instead of changing version policy silently.

## 2.1 Release Principles

**Reproducibility**

Release artifacts must be produced through documented, repeatable procedures.

**Minimal Installation Burden**

End users must not need to install Go, Node.js, Python, or other development toolchains.

**Verified Artifacts**

A successful compilation alone is insufficient to declare a release package valid.

**Cross-platform Accuracy**

Document and distribute only platforms and configurations that have been verified.

**Secure Release Process**

Use minimal workflow permissions and protect release publication.

**Traceability**

Each release artifact must correspond to a specific source commit and version tag.

**Documentation Consistency**

English and Japanese documentation must accurately describe the released behavior.

**Explicit Publication Approval**

Do not create public releases, push tags, or publish release assets without the required authorization.

---

# 3. Release Scope

## 3.1 Windows

Windows x64 remains the primary distribution target.

Preferred release format:

```text
MeiliScope_v1.0.0_windows_amd64.zip
```

The package should preferably contain a standalone executable and required documentation.

Example:

```text
MeiliScope_v1.0.0_windows_amd64.zip
├── meiliscope.exe
├── LICENSE
└── README.md
```

Additional runtime files may be included if required by the verified Wails build.

Do not omit necessary files merely to maintain a single-executable layout.

Do not introduce a mandatory installer unless a verified technical requirement makes it necessary.

## 3.2 Linux

Linux x64 is included in release preparation because previous phases established Linux execution support.

The exact supported distributions and runtime prerequisites must come from Phase 6 verification.

Preferred initial format:

```text
MeiliScope_v1.0.0_linux_amd64.tar.gz
```

Example:

```text
MeiliScope_v1.0.0_linux_amd64.tar.gz
├── meiliscope
├── LICENSE
└── README.md
```

Preserve executable permissions.

Document any GTK, WebKitGTK, or other runtime dependencies.

Do not claim that the archive works on every Linux distribution.

Evaluate AppImage as an optional additional distribution format only if it can be built and validated reliably.

DEB and RPM packages are not mandatory for the initial release.

Do not introduce several packaging formats merely because Wails supports them.

## 3.3 Unsupported Platforms

Do not publish macOS, Windows ARM64, or Linux ARM64 packages unless those targets have been explicitly validated and approved.

Do not describe a successful cross-compilation as equivalent to runtime validation.

---

# Step 7.1 — Documentation & OSS Readiness

## Objective

Prepare the repository and documentation for public use, installation, development, and contribution.

Ensure that all user-facing documentation matches the actual implementation.

## Tasks

### 7.1.1 Audit Existing Documentation

Inspect:

- `README.md`
- `plan.md`
- `LICENSE`
- `CHANGELOG.md`
- `CONTRIBUTING.md`
- Existing `docs/en/`
- Existing `docs/ja/`
- Architecture documents.
- Phase completion reports.
- Release-readiness reports.
- Existing GitHub workflows.

Identify:

- Missing documentation.
- Outdated commands.
- Incorrect paths.
- Broken links.
- Unsupported claims.
- Missing runtime prerequisites.
- Duplicated content.
- English/Japanese inconsistencies.

Preserve valid documentation.

Do not rewrite unrelated documents unnecessarily.

### 7.1.2 Finalize the Root README

The root README must serve as the primary entry point for new users.

Include:

- Project name.
- Mission and short description.
- Main capabilities.
- Supported platforms.
- Download and installation instructions.
- Runtime prerequisites.
- Quick Start.
- Documentation links.
- Screenshots where useful.
- Development build instructions.
- License.
- Contribution information.

Keep the root README concise.

Move detailed explanations into `docs/`.

Do not include screenshots of features that are not implemented.

### 7.1.3 Implement English and Japanese Documentation

Maintain equivalent documentation structures.

Recommended structure:

```text
docs/
├── en/
│   ├── getting-started/
│   │   ├── installation.md
│   │   └── quick-start.md
│   ├── user-guide/
│   │   ├── connections.md
│   │   ├── index-explorer.md
│   │   ├── document-browser.md
│   │   ├── json-inspector.md
│   │   ├── query-console.md
│   │   ├── search-builder.md
│   │   ├── saved-queries.md
│   │   ├── document-editing.md
│   │   ├── task-tracking.md
│   │   └── export.md
│   ├── development/
│   │   ├── building.md
│   │   └── testing.md
│   └── troubleshooting.md
│
└── ja/
    ├── getting-started/
    ├── user-guide/
    ├── development/
    └── troubleshooting.md
```

Adapt paths to the existing repository conventions.

Do not generate duplicate documentation trees.

Use English as the primary technical reference where appropriate, while maintaining practical Japanese user documentation.

Do not use machine-translated technical terminology without reviewing it for consistency.

### 7.1.4 Finalize Installation Instructions

For Windows, document:

1. Download the release ZIP.
2. Extract the archive.
3. Launch `meiliscope.exe`.
4. Configure a Meilisearch connection.
5. Begin browsing indexes.

Explain:

- Supported Windows versions.
- WebView2 requirements.
- Any unsigned-executable warnings.
- Expected configuration location.
- User-level permissions.
- Known installation restrictions.

For Linux, document:

1. Download the appropriate archive.
2. Extract it.
3. Ensure the executable has execution permission.
4. Install documented runtime dependencies if necessary.
5. Launch MeiliScope.

Explain:

- Validated distributions.
- Required GTK/WebKitGTK libraries.
- Secret Service or keyring behavior.
- Session-only credential mode.
- Display environment requirements.
- Known runtime limitations.

Never claim that Linux execution is fully self-contained unless verified.

### 7.1.5 Provide a Practical Quick Start

The Quick Start must demonstrate:

1. Starting MeiliScope.
2. Creating a connection.
3. Entering an API Key when required.
4. Selecting an index.
5. Browsing documents.
6. Inspecting JSON.
7. Executing a search query.
8. Reviewing search results.

Use simple, realistic examples.

Clearly distinguish read-only operations from operations that modify data.

### 7.1.6 Finalize the User Guide

Document all major features implemented through Phase 5.

Include:

- Connection management.
- API Key requirements.
- Index Explorer.
- Document Browser.
- JSON Inspector.
- Query Console.
- Search Builder.
- Query History.
- Saved Queries.
- Document editing.
- Task tracking.
- JSON/CSV export.
- Keyboard shortcuts.

Explain important limitations.

Examples:

- Search-only keys may not permit index enumeration.
- Document writes are asynchronous.
- An accepted task is not necessarily completed.
- Partial update and replacement have different semantics.
- CSV export has documented representation and safety rules.
- Exporting the current page does not mean exporting the entire index.

### 7.1.7 Finalize Developer Documentation

Document:

- Required Go version.
- Required Node.js version.
- Required Wails version.
- Windows build prerequisites.
- Linux build prerequisites.
- Dependency installation.
- Development mode.
- Production build.
- Unit tests.
- Integration tests.
- Frontend validation.
- Desktop smoke tests.
- Release packaging.

Provide executable commands verified against the repository.

Do not copy generic Wails build instructions if the actual project uses customized Taskfiles.

### 7.1.8 Finalize OSS Repository Files

Verify or create:

```text
LICENSE
CONTRIBUTING.md
CODE_OF_CONDUCT.md
SECURITY.md
CHANGELOG.md
.github/ISSUE_TEMPLATE/
.github/PULL_REQUEST_TEMPLATE.md
```

Use proportional content.

Do not add unnecessary governance documents.

`SECURITY.md` should describe an appropriate vulnerability reporting process.

Do not publish personal contact information or security-reporting endpoints without authorization.

If GitHub private vulnerability reporting is used, document the actual configured mechanism.

### 7.1.9 Review License Compliance

Verify that the repository contains the complete MIT License text.

Review licenses for:

- Go dependencies.
- npm dependencies.
- Wails components.
- Bundled assets.
- Icons.
- Fonts.
- Other redistributed resources.

Preserve third-party license notices when required.

Do not assume that MeiliScope's MIT License eliminates obligations associated with bundled dependencies.

Do not include files or assets whose redistribution rights are unclear.

### 7.1.10 Finalize CHANGELOG

Use a consistent format such as Keep a Changelog.

Include an appropriate initial release entry.

Summarize:

- Connection Manager.
- Index Explorer.
- Document Browser.
- Query Console.
- Document editing.
- Search Builder.
- Saved Queries.
- Export.
- Platform support.

Do not claim features that are incomplete.

Avoid misleading release dates.

### 7.1.11 Review GitHub Wiki Policy

The project proposal treats GitHub Wiki as optional.

Keep authoritative versioned documentation in the repository.

Do not create a Wiki merely because documentation has grown.

Consider a Wiki only when there is a clear need for additional community-maintained reference material.

Do not automatically migrate version-sensitive documentation away from `docs/`.

### 7.1.12 Validate Documentation

Check:

- Internal links.
- External links where practical.
- File paths.
- Build commands.
- Installation commands.
- Screenshots.
- Version references.
- Platform requirements.
- English/Japanese navigation consistency.

Automate Markdown and link validation where practical.

Do not create a documentation workflow that depends on unavailable private services.

## Deliverables

- Finalized README.
- English and Japanese user documentation.
- Developer build and testing documentation.
- OSS policy files.
- Updated CHANGELOG.
- License compliance findings.
- Documentation validation results.

## Acceptance Criteria

- Users can install and operate the application using the documentation.
- Contributors can build and test the application using documented commands.
- English and Japanese user guides cover the released functionality.
- Runtime prerequisites are accurate.
- No known broken critical documentation links remain.
- License and redistribution requirements are addressed.
- No unsupported product claims remain.

---

# Step 7.2 — Release Automation & Packaging

## Objective

Implement a secure, reproducible release pipeline using GitHub Actions.

Generate versioned Windows and Linux packages, checksums, and release metadata.

## Tasks

### 7.2.1 Audit Existing Build Infrastructure

Inspect:

- Existing GitHub Actions workflows.
- Wails Taskfiles.
- Go build configuration.
- Frontend build configuration.
- Version metadata.
- Existing packaging scripts.
- Release-related scripts.
- Phase 6 CI quality gates.

Reuse verified build procedures.

Do not replace working automation solely for stylistic reasons.

### 7.2.2 Define Versioning Policy

Use Semantic Versioning.

Examples:

```text
v0.1.0
v0.9.0-rc.1
v1.0.0
v1.0.1
```

Define a single authoritative release version.

Ensure consistency between:

- Git tag.
- Application version.
- Windows executable metadata.
- Linux application metadata.
- Archive filenames.
- Release title.
- CHANGELOG.

Do not maintain several unrelated hardcoded version strings.

Use the release tag as the authoritative version where practical.

Embed:

- Version.
- Commit hash.
- Build date, if appropriate.

Avoid including machine-specific build paths.

### 7.2.3 Define Release Artifact Names

Use predictable filenames.

Required artifact names:

```text
MeiliScope_v1.0.0_windows_amd64.zip
MeiliScope_v1.0.0_linux_amd64.tar.gz
SHA256SUMS.txt
```

Optional additional artifacts may include:

```text
MeiliScope_v1.0.0_linux_amd64.AppImage
```

Generate names from the actual release version.

Never hardcode `v1.0.0` into reusable packaging logic.

Do not change archive naming conventions between workflow runs without a documented reason.

### 7.2.4 Implement Windows Packaging

Use the production build process validated in previous phases.

Requirements:

- Windows x64 executable.
- Production frontend assets embedded or packaged correctly.
- Correct application name.
- Correct version metadata.
- Correct icon where available.
- Required license and documentation.
- No development server dependency.
- No hardcoded development paths.

Prefer a ZIP-based portable application.

Do not make NSIS or MSIX mandatory for the initial release.

If additional runtime components are required, document and include them appropriately.

Do not bundle WebView2 runtime components without verifying their redistribution terms and deployment requirements.

### 7.2.5 Implement Linux Packaging

Use the verified Linux production build.

Requirements:

- Linux x64 executable.
- Correct executable permissions.
- Correct version metadata.
- Required documentation.
- Required license files.
- No development-only assets.
- Documented shared-library requirements.

Create a `.tar.gz` package.

Preserve executable permissions inside the archive.

If AppImage is supported and validated, it may be added as an optional artifact.

Do not claim universal Linux portability.

Do not introduce mandatory DEB or RPM support without a separate decision.

### 7.2.6 Implement Checksums

Generate SHA-256 checksums for all release archives.

Use the actual files produced by the workflow.

Example:

```text
SHA256SUMS.txt
```

Requirements:

- Include the exact artifact filenames.
- Include a checksum for each distribution artifact.
- Exclude checksum files themselves unless deliberately documented.
- Ensure checksums are computed after packaging.
- Verify checksums before publication.
- Use deterministic file ordering.

Do not calculate checksums for nonexistent or incorrectly named artifacts.

### 7.2.7 Implement Release Workflow

Create or update:

`.github/workflows/release.yml`

Use a release pipeline with logical stages:

```text
Version Validation
        |
        v
Required Quality Checks
        |
        v
Windows Build --- Linux Build
        |               |
        v               v
Package Windows   Package Linux
        |               |
        +-------+-------+
                |
                v
      Artifact Validation
                |
                v
       SHA-256 Generation
                |
                v
       Draft GitHub Release
                |
                v
        Manual Approval
                |
                v
       Public Release
```

The implementation may adapt this structure to existing workflows.

Do not create circular or duplicate build dependencies.

### 7.2.8 Configure Workflow Triggers

Support appropriate triggers:

- Version tag.
- Manual workflow dispatch for testing or controlled release preparation.

Example tag convention:

```text
v*
```

Validate tag syntax before building.

Do not publish stable releases from arbitrary development branches.

A tag-triggered workflow must still enforce the project's required release checks.

A successful build must not automatically imply authorization to publish.

### 7.2.9 Protect Release Publication

Separate artifact preparation from public publication.

The workflow should be capable of preparing a draft release without making it public immediately.

Use explicit authorization before:

- Creating a public stable release.
- Publishing a draft.
- Replacing public release assets.
- Moving a version tag.
- Deleting release artifacts.

Where appropriate, use GitHub environment protection rules.

If environment protection is unavailable, use an equivalent manual publication procedure.

Do not simulate approval by assuming a successful workflow is sufficient.

### 7.2.10 Apply Least-Privilege Workflow Permissions

Use minimal permissions for build jobs.

Grant `contents: write` only to jobs that actually need to create or update GitHub Releases.

Do not expose release tokens to pull-request builds.

Do not print tokens or secrets.

Use official GitHub-supported authentication mechanisms.

Avoid storing long-lived personal access tokens when the GitHub Actions token is sufficient.

### 7.2.11 Pin Toolchains and Actions

Use appropriate version pinning for:

- Go.
- Node.js.
- Wails CLI.
- npm dependencies.
- GitHub Actions.

Prefer immutable action commit references where practical.

Ensure dependency lockfiles are committed.

Avoid installing unpinned tooling from arbitrary URLs.

Do not silently adopt a new Wails release during packaging.

### 7.2.12 Ensure Build Reproducibility

Use:

- Go Modules.
- Frontend lockfile.
- Explicit toolchain versions.
- Verified Wails build configuration.
- Deterministic package layouts.

Record the source commit and toolchain information.

Aim for reproducible builds.

Do not claim byte-for-byte reproducibility unless it has been tested.

### 7.2.13 Validate Package Contents

Before uploading artifacts:

- List archive contents.
- Verify required files.
- Verify version metadata.
- Verify executable presence.
- Verify Linux execute permissions.
- Verify license inclusion.
- Verify documentation inclusion.
- Check for duplicate or unexpected files.

Reject packages containing:

- `.env` secrets.
- API Keys.
- Local settings.
- Developer credential stores.
- Test databases.
- Temporary files.
- Unnecessary source build caches.

### 7.2.14 Generate Release Notes

Generate or assemble release notes from:

- CHANGELOG.
- Version metadata.
- Verified features.
- Supported platforms.
- Known limitations.
- Installation prerequisites.

Provide links to documentation.

Do not publish claims about untested platforms.

The release notes should explain that MeiliScope is not a server administration console.

### 7.2.15 Handle Build Failure

If any required platform build fails:

- Mark the release workflow as failed.
- Preserve useful logs.
- Do not publish an incomplete stable release.
- Do not replace previously valid artifacts.
- Report the failing stage.

Do not silently downgrade required failures to warnings.

### 7.2.16 Handle Partial Publication Failure

A release upload may fail after some artifacts are uploaded.

Ensure the process is recoverable.

Prefer draft release creation and validation before public publication.

Do not silently publish a release missing required platform artifacts.

Do not overwrite public artifacts with different bytes under the same version without an explicit release correction policy.

### 7.2.17 Implement Local Packaging Verification

Provide a documented method to produce and inspect release packages locally.

This should not require uploading to GitHub.

Support:

- Building.
- Packaging.
- Listing archive contents.
- Verifying checksums.
- Inspecting version metadata.

Use existing build scripts where appropriate.

### 7.2.18 Document Release Procedure

Create:

`docs/development/releasing.md`

Document:

1. Verify Phase 6 approval.
2. Select release version.
3. Update CHANGELOG.
4. Run required tests.
5. Prepare release commit.
6. Create release tag after authorization.
7. Run release workflow.
8. Verify generated artifacts.
9. Validate checksums.
10. Validate draft release.
11. Obtain publication approval.
12. Publish the release.
13. Verify public download links.
14. Record release results.

Include rollback or withdrawal guidance for a defective release.

## Deliverables

- Release workflow.
- Windows packaging.
- Linux packaging.
- Checksum generation.
- Version metadata integration.
- Release note generation.
- Local packaging verification.
- Release procedure documentation.

## Acceptance Criteria

- Versioned Windows packages can be generated.
- Versioned Linux packages can be generated.
- Checksums are generated from actual artifacts.
- Required artifact contents are validated.
- Toolchain versions are controlled.
- Release jobs use minimal permissions.
- Public publication requires appropriate authorization.
- Release preparation can be tested without publishing publicly.
- Failed required checks prevent release publication.
- Packaging is documented and repeatable.

---

# Step 7.3 — Distribution Validation & Release Approval

## Objective

Validate the actual release artifacts and establish a final release approval decision.

Do not treat successful CI compilation as proof that the distributed application works correctly.

## Tasks

### 7.3.1 Prepare Release Candidates

Generate the Windows and Linux packages using the release workflow.

Use the intended release configuration.

Do not substitute development builds for release artifacts.

Record:

- Version.
- Source commit.
- Build workflow run.
- Toolchain versions.
- Artifact filenames.
- Artifact checksums.

Validate the exact files intended for distribution.

### 7.3.2 Verify Windows Clean Installation

Test on a clean or appropriately isolated Windows environment.

Requirements:

- Download or copy the release ZIP.
- Extract it outside the source repository.
- Launch the executable.
- Verify the main window.
- Verify frontend assets.
- Confirm no development server is required.
- Confirm no Go installation is required.
- Confirm no Node.js installation is required.
- Confirm no Python installation is required.
- Verify the documented WebView2 behavior.
- Verify settings persistence.
- Verify application shutdown and restart.

Record the Windows version and installed prerequisites.

If the environment is not clean, document the limitations of the test.

### 7.3.3 Verify Linux Distribution

Test the exact Linux archive on at least one validated target environment.

Requirements:

- Extract the archive.
- Verify execute permission.
- Launch the executable.
- Verify the main window.
- Verify frontend assets.
- Verify documented GTK/WebKitGTK dependencies.
- Verify credential storage behavior.
- Verify session-only mode if relevant.
- Verify settings persistence.
- Verify application shutdown and restart.

Record:

- Distribution and version.
- Desktop environment.
- Display server where relevant.
- GTK/WebKitGTK versions.
- Secret Service availability.
- Additional runtime dependencies.

Do not claim universal Linux compatibility.

### 7.3.4 Validate Basic Functional Workflow

Use the release builds, not development mode.

Test:

1. Launch application.
2. Create or select a connection.
3. Connect to a test Meilisearch instance.
4. Browse indexes.
5. View index statistics.
6. Browse documents.
7. Inspect JSON.
8. Execute a search query.
9. Review search results.
10. Use Search Builder.
11. Save and reload a query.
12. Export JSON and CSV.
13. Close and restart the application.

Verify saved settings after restart.

### 7.3.5 Validate Document Editing

Use a disposable test index.

Test:

- Create document.
- Edit document.
- Review diff.
- Confirm mutation.
- Track task.
- Verify task success.
- Re-fetch document.
- Delete document.
- Verify deletion.

Do not run mutation tests against production data.

Do not use arbitrary user-defined connections for automated destructive tests.

### 7.3.6 Validate Credentials

Test:

- Persistent credential storage where supported.
- Session-only credentials.
- Missing credentials.
- Credential replacement.
- Application restart.
- Credential deletion.

Verify that credentials are absent from release packages and diagnostic logs.

### 7.3.7 Validate Failure Recovery

Test representative failures:

- Server unavailable.
- Invalid API Key.
- Permission denied.
- Request timeout.
- Connection switching during a request.
- Invalid search query.
- Task failure.
- Export failure.
- Corrupted optional configuration.

The application must remain usable after recoverable failures.

### 7.3.8 Validate Release Artifact Integrity

For every artifact:

- Verify the checksum.
- Inspect archive contents.
- Check executable architecture.
- Check version metadata.
- Check required runtime files.
- Check license and documentation.
- Check for sensitive files.
- Check file permissions.
- Confirm archive extraction.

Do not approve artifacts with mismatched checksums.

### 7.3.9 Validate Documentation Against Packages

Follow installation instructions exactly as written.

Do not use undocumented build tools or setup commands.

Verify that:

- The package filenames match the documentation.
- Runtime prerequisites are correct.
- Quick Start instructions work.
- Troubleshooting guidance is useful.
- The documented features exist.
- Links to downloads and documentation are correct.

Fix discrepancies before release approval.

### 7.3.10 Verify Dependency and License Notices

Inspect the final distributed packages.

Confirm:

- MIT License is included.
- Required third-party notices are included.
- Bundled assets have appropriate redistribution rights.
- No unnecessary developer files are distributed.

Document unresolved licensing concerns as release blockers when applicable.

### 7.3.11 Review Known Issues

Review Phase 6 findings and all newly discovered release-candidate defects.

Classify outstanding issues.

Block release approval for:

- Critical security defects.
- Unintended document mutations.
- Credential leakage.
- Silent data corruption.
- Broken required platform packages.
- Core application failures.
- Failed required CI checks.
- Missing mandatory license obligations.

Document acceptable non-blocking limitations.

Do not hide known problems from release notes.

### 7.3.12 Review Artifact Authenticity

Evaluate whether GitHub artifact attestations can be included using the available workflow permissions and project configuration.

If implemented:

- Generate attestations for release artifacts.
- Associate them with the correct source commit.
- Document verification steps.

Artifact attestations are recommended but not mandatory for the initial release.

Do not claim code signing or cryptographic provenance guarantees beyond what the pipeline actually provides.

### 7.3.13 Review Unsigned Application Behavior

If executables are unsigned:

- State that explicitly.
- Validate ordinary launch behavior.
- Document any relevant operating-system warnings.
- Do not claim that unsigned binaries are trusted by Windows SmartScreen or other reputation systems.

Do not add code signing keys or external signing services without approval.

### 7.3.14 Prepare Release Approval Report

Create:

`docs/development/phase-7-report.md`

Include:

1. Release version.
2. Source commit.
3. Build environment.
4. Workflow results.
5. Windows artifact details.
6. Linux artifact details.
7. SHA-256 checksums.
8. Runtime prerequisites.
9. Windows test results.
10. Linux test results.
11. Functional smoke test results.
12. Security validation results.
13. License compliance findings.
14. Known issues.
15. Unverified conditions.
16. Release blockers.
17. Publication recommendation.

Use explicit statuses:

- PASS.
- FAIL.
- BLOCKED.
- NOT RUN.
- NOT APPLICABLE.

Do not report an unexecuted test as passing.

### 7.3.15 Establish Release Decision

The final recommendation must be one of:

**APPROVED FOR PUBLICATION**

All mandatory checks pass, and no release-blocking issues remain.

**APPROVED WITH DOCUMENTED LIMITATIONS**

Mandatory checks pass, and only accepted non-blocking limitations remain.

**NOT APPROVED**

One or more mandatory release conditions remain unmet.

If publication is not approved, identify the blockers and required corrective actions.

An approval recommendation is not authorization to publish.

### 7.3.16 Prepare Publication Without Performing Unauthorized Actions

When the release is approved:

- Prepare the release metadata.
- Prepare the final release notes.
- Verify the intended tag.
- Verify required artifacts.
- Verify checksums.
- Verify publication permissions.
- Prepare the GitHub release workflow.

Do not create or push a release tag without authorization.

Do not publish a draft release without authorization.

Do not modify or replace an existing public release without authorization.

After explicit approval, the configured GitHub workflow may perform the authorized publication.

### 7.3.17 Verify Published Release

If publication is authorized and actually performed:

- Verify the GitHub Release exists.
- Verify its version and tag.
- Verify the source commit.
- Verify all expected artifacts.
- Download and verify checksums.
- Verify release notes.
- Verify documentation links.
- Record the release URL.
- Record publication results.

If publication is not performed, report the state as release-ready rather than published.

Do not invent a GitHub Release URL.

## Deliverables

- Validated release-candidate packages.
- Artifact verification results.
- Windows smoke-test report.
- Linux smoke-test report.
- License compliance report.
- Release approval recommendation.
- Final release notes.
- Phase 7 completion report.

## Acceptance Criteria

- Windows release artifacts pass required validation.
- Linux release artifacts pass required validation for declared targets.
- Checksums match.
- Runtime requirements are documented.
- Core application features work in production builds.
- No release-blocking security or data-integrity issue remains.
- License requirements are satisfied.
- Documentation matches actual release packages.
- Release approval status is explicit.
- Public publication occurs only with authorization.

---

# 4. General Implementation Constraints

## 4.1 No New Major Features

Do not implement:

- New Meilisearch management capabilities.
- Generic HTTP Request Console.
- Server administration.
- API Key administration.
- Automatic updates.
- Cloud synchronization.
- Additional supported platforms without validation.
- New document indexing features.
- Unrequested UI redesigns.

Phase 7 is strictly for release preparation.

## 4.2 Do Not Break Existing Functionality

Preserve:

- Connection Manager.
- Index Explorer.
- Document Browser.
- Query Console.
- Search Builder.
- Saved Queries.
- Document editing.
- Task Tracking.
- JSON/CSV Export.

Any required correction must be accompanied by appropriate regression validation.

## 4.3 Preserve Security

Do not:

- Embed credentials in release packages.
- Disable TLS verification.
- Store API Keys in plaintext.
- Remove mutation confirmation.
- Expose arbitrary administrative APIs.
- Publish secrets through GitHub Actions.
- Use unnecessary privileged workflows.

## 4.4 Preserve Cross-platform Boundaries

Keep Windows-specific and Linux-specific packaging separate.

Do not assume platform runtime requirements are interchangeable.

Do not claim support for untested operating systems.

## 4.5 No Silent Publication

Do not:

- Push tags automatically without authorization.
- Publish public releases automatically without the required approval.
- Replace existing public release assets silently.
- Delete release history without approval.
- Publish incomplete artifacts as a stable release.

## 4.6 Keep Release Automation Maintainable

Prefer official GitHub Actions and existing Wails build procedures.

Avoid unnecessary external release services.

Do not introduce complicated packaging infrastructure unless required by verified deployment constraints.

## 4.7 Avoid Unsupported Claims

Do not claim:

- Universal Linux compatibility.
- Complete installation independence from OS libraries.
- Code signing when absent.
- Reproducible byte-identical builds without validation.
- Security certification.
- Compatibility with untested Meilisearch versions.

Describe actual verification results.

---

# 5. Execution & Reporting Requirements

Execute Phase 7 in the following order:

1. Step 7.1 — Documentation & OSS Readiness.
2. Step 7.2 — Release Automation & Packaging.
3. Step 7.3 — Distribution Validation & Release Approval.

Each step must produce independently reviewable deliverables.

After each step, report:

### Completed Work

Describe implementation and validation performed.

### Files Changed

List important source files, workflows, scripts, and documentation changes.

### Tests Executed

List actual commands and results.

### Artifacts Generated

List actual artifact filenames and paths.

### Technical Decisions

Explain important release engineering decisions.

### Known Issues

List unresolved problems and limitations.

### Next Step Readiness

State whether the next step may proceed.

Do not claim that GitHub Actions succeeded if it was not executed.

Do not claim that a clean-machine test passed when it was not performed.

Do not claim a public release exists unless publication was verified.

---

# 6. Phase 7 Definition of Done

Phase 7 is complete when:

1. README is complete and accurate.
2. English and Japanese user documentation is available.
3. Developer build and test instructions are verified.
4. MIT License and third-party obligations are addressed.
5. CHANGELOG is updated.
6. Release versioning is consistent.
7. Windows x64 production packages can be generated.
8. Linux x64 production packages can be generated for declared supported environments.
9. Release checksums are generated and verified.
10. GitHub Actions release automation is configured.
11. Required quality gates are enforced.
12. Public publication requires appropriate authorization.
13. Windows release artifacts are tested.
14. Linux release artifacts are tested.
15. Runtime prerequisites are documented.
16. Core workflows pass production-build smoke tests.
17. Credential security remains intact.
18. Document mutation safety remains intact.
19. No Critical or High release-blocking defects remain.
20. Final release notes are prepared.
21. Phase 7 completion report is written.
22. Release approval status is explicitly recorded.

The intended final result is a release-ready MeiliScope version suitable for public distribution through GitHub Releases.

If publication is not authorized, stop at the verified release-ready state and report the exact remaining publication action.

If required validation is incomplete, report the release as blocked rather than claiming completion.

**Successful completion of Phase 7 establishes MeiliScope as a documented, validated, and distributable open-source desktop application.**
