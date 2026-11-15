# MeiliScope — Project Proposal

**Version:** 0.2.0 (Phase 0 — updated after Wails v3 validation)  
**Last updated:** 2026-11-16

---

## 1. Project Mission

**"Explore your Meilisearch data."**

MeiliScope is an open-source, Windows-first desktop application for exploring Meilisearch indexes and documents, executing search queries, and performing limited document editing.

---

## 2. Problem Statement

Meilisearch is a fast, developer-friendly search engine. Its REST API and dashboard cover administration and basic exploration, but developers and data teams often need a richer local tool for:

- Browsing index contents without writing curl commands
- Iterating on search queries and filter expressions
- Inspecting and editing individual documents
- Switching between multiple Meilisearch instances

MeiliScope fills this gap as a lightweight, portable desktop application.

---

## 3. Target Users

- Developers integrating Meilisearch into applications
- Data teams managing search configurations
- QA engineers verifying search result quality
- System administrators managing multiple Meilisearch instances

---

## 4. Product Positioning

MeiliScope is a developer tool, not an end-user product. It assumes:

- The user has access to at least one Meilisearch instance
- The user understands basic Meilisearch concepts (indexes, documents, search)
- The user is comfortable managing connection settings

MeiliScope does not replace the Meilisearch dashboard but complements it with a richer local exploration experience.

---

## 5. Core Features

### Priority 1 — Connection and Index Management

- Save and manage multiple named Meilisearch connections
- Connect via URL and optional API key
- Browse available indexes
- Display index health and statistics

### Priority 2 — Document Browser

- List documents with pagination
- View full document JSON
- Basic document filtering

### Priority 3 — Search and Query Console

- Execute full-text search queries
- Configure search parameters (filters, attributes, pagination)
- Display search results with highlighting
- Save and recall search presets

### Priority 4 — Document Editing

- Edit individual documents as JSON
- Validate JSON before submission
- Save and revert changes

### Priority 5 — Usability and Polish

- Keyboard shortcuts
- Persistent window and layout state
- Compact and comfortable density modes
- Localisation support (English and Japanese)

---

## 6. MVP Definition

The MVP is complete at the end of Phase 3 (Query Console).

A usable MVP requires:

1. At least one saved Meilisearch connection that persists between launches
2. Index browser showing available indexes
3. Document browser with pagination
4. Search console with filter support
5. Working Windows x64 portable executable

Document editing (Phase 4) is a post-MVP feature.

---

## 7. Non-Goals

- Linux distribution (Phase 0 scope)
- macOS distribution (Phase 0 scope)
- Meilisearch server administration (create/delete indexes via UI)
- API key creation or management
- Automatic application updates
- Multi-user or collaborative features
- Cloud synchronisation of connection settings
- Meilisearch version management

---

## 8. License

MIT License. See `LICENSE` in the repository root.

---

## 9. Technical Approach

### 9.1 Technology Stack

| Component | Technology | Version | Notes |
|-----------|------------|---------|-------|
| Backend language | Go | 1.27.1 | Minimum: 1.26.0 (Wails requirement) |
| Desktop framework | Wails v3 | 3.1.0 | Validated in Phase 0 (see ADR 0001) |
| Frontend framework | React | 18.2.0 | TypeScript, Vite build |
| Frontend language | TypeScript | 5.2.2 | |
| Frontend build | Vite | 8.3.8 | |
| Frontend bindings | Wails generator | auto | Generated from Go service structs |
| Package manager | npm | 11.19.1 | lockfile committed |
| Configuration | JSON | — | `%AppData%\MeiliScope\config.json` |
| Credential storage | Windows Credential Manager | — | Via `danieljoos/wincred` |
| Logging | structured (Go `log/slog`) | — | Per-session log files |

### 9.2 Architecture

The application consists of two main parts running in the same process:

**Go backend (internal packages):**

- `internal/app` — application lifecycle and Wails integration
- `internal/services` — business operations exposed to the frontend via Wails bindings
- `internal/meilisearch` — HTTP client for Meilisearch API (Phase 1+)
- `internal/config` — JSON configuration persistence in `%AppData%`
- `internal/platform` — Windows-specific integrations (Credential Manager)

**React frontend (`frontend/src`):**

- Renders in a WebView2 window managed by Wails
- Communicates with Go exclusively through auto-generated Wails service bindings
- Owns all presentation state, routing, and user interaction

**Architecture boundaries:**

- Go owns all HTTP communication with Meilisearch
- The frontend owns presentation and user interaction
- Application services must not depend on GUI components
- The Meilisearch API client must be independently testable
- Credentials must not be exposed to the frontend or passed through bindings
- Settings and credentials have separate storage responsibilities

### 9.3 Communication

Frontend-to-Go method calls use Wails auto-generated TypeScript bindings returning `CancellablePromise<T>`. The generator runs during each build and produces type-safe TypeScript from Go struct methods.

Go-to-frontend notifications use the Wails event system with typed event name constants generated alongside method bindings.

---

## 10. Distribution and Installation

### 10.1 Target Format

Primary distribution: portable ZIP package.

```
MeiliScope_<version>_windows_amd64.zip
├── meiliscope.exe     (single self-contained executable)
├── LICENSE
└── README.md
```

The executable contains:
- Compiled Go runtime and application code
- All React/TypeScript frontend assets (embedded via Go `embed`)
- Windows version and icon resources

### 10.2 Runtime Prerequisites

**End users require:**

| Prerequisite | Notes |
|-------------|-------|
| Windows 10 (1803 or later) or Windows 11 | WebView2 is pre-installed on these versions via Windows Update |
| Microsoft Edge WebView2 Runtime | Pre-installed on Windows 10 1803+ and all Windows 11; Wails bootstrapper can silently install it on demand if absent |

**End users do not require:**

- Go installation
- Node.js installation
- Python installation
- .NET application runtime
- Docker
- Administrator privileges for normal use

### 10.3 Configuration Storage

User configuration is stored in `%AppData%\MeiliScope\` (per-user, roaming), not beside the executable. This allows the executable to be replaced without losing configuration.

### 10.4 NSIS Installer (optional, post-MVP)

An NSIS installer is supported by Wails v3 but is not required for the initial distribution. The portable ZIP format is the Phase 0–3 distribution target.

---

## 11. Source Repository

- Host: GitHub
- License: MIT
- Primary branch: `main`
- Issue tracking: GitHub Issues
- Pull request reviews required before merging to `main`

---

## 12. CI/CD and Package Management

### 12.1 Continuous Integration

GitHub Actions, triggered on:
- Pushes to `main`
- Pull requests targeting `main`

CI performs:
1. `go vet` and `gofmt` checks
2. Go unit tests with race detector
3. Frontend TypeScript type checks
4. Frontend ESLint
5. Frontend production build
6. Windows x64 desktop application build
7. Build artifact upload

**Runner:** Windows (`windows-latest`) for the application build. No CGO required; standard Go toolchain is sufficient.

### 12.2 Release Workflow

Triggered by version tags (`v*.*.*`).

Produces:
- `meiliscope.exe` (Windows x64)
- `MeiliScope_<version>_windows_amd64.zip`
- `MeiliScope_<version>_windows_amd64.zip.sha256`

Publishes to GitHub Releases.

### 12.3 Dependency Management

Dependabot configured for:
- Go Modules
- npm packages
- GitHub Actions

All dependencies are MIT-compatible or similarly permissive.

---

## 13. Development Environment

**Required for contributors:**

| Tool | Version |
|------|---------|
| Go | 1.26.0 or later |
| Node.js | 20.0.0 or later |
| npm | 10.0.0 or later |
| Wails CLI | v3.1.0 (via `go install`) |

No C compiler is required. No Docker is required.

**Common commands:**

```powershell
# Install Wails CLI
go install github.com/wailsapp/wails/v3/cmd/wails3@latest

# Install frontend dependencies
cd frontend && npm install

# Development mode (hot reload)
wails3 dev

# Run Go tests
go test ./...

# TypeScript type check
cd frontend && npm run build

# Production build
wails3 build
```

---

## 14. Non-functional Requirements

| Requirement | Target |
|------------|--------|
| Startup time | < 3 seconds on modern hardware with WebView2 cached |
| Memory at idle | < 200 MB resident |
| Search query display | < 500 ms for typical result sets |
| Configuration save | Atomic write, no data loss on crash |
| Minimum Windows version | Windows 10 1803 (Build 17134) |
| Executable size | < 30 MB (Phase 0 PoC: 10.85 MB) |
| Accessibility | Keyboard-navigable primary flows |

---

## 15. Security Requirements

- API keys are stored in Windows Credential Manager, not in configuration files or environment variables
- API keys are never transmitted to the frontend layer
- Configuration files contain no secrets
- Log files must not contain API keys or request bodies with sensitive data
- The application communicates with Meilisearch over HTTP or HTTPS as configured; TLS certificate validation is on by default

---

## 16. Development Roadmap

### Phase 0 — Foundation & Technology Validation

**Goal:** Verified application foundation. GUI framework decision. Architecture documents. CI pipeline. Windows distribution validated.

Deliverables:
- Wails v3 feasibility evaluation (`docs/architecture/wails-v3-evaluation.md`)
- Architecture Decision Record (`docs/architecture/adr/0001-desktop-framework.md`)
- Architecture overview (`docs/architecture/overview.md`)
- Project scaffolding (Go + React + TypeScript)
- Application shell with placeholder navigation
- Configuration persistence (`%AppData%`)
- Structured logging
- GitHub Actions CI and release workflows
- Phase 0 completion report

### Phase 1 — Connection & Index Explorer

**Goal:** Connect to a real Meilisearch instance and browse indexes.

Deliverables:
- Connection Manager (add/edit/delete named connections)
- Secure API key storage via Windows Credential Manager
- Meilisearch HTTP client
- Index list with health and statistics
- Connection test

### Phase 2 — Document Browser

**Goal:** Browse and view documents from a selected index.

Deliverables:
- Document list with pagination
- Document detail view (JSON)
- Basic attribute filtering

### Phase 3 — Query Console *(MVP milestone)*

**Goal:** Execute and iterate on Meilisearch search queries.

Deliverables:
- Full-text search with result display
- Filter and facet support
- Attribute configuration
- Saved search presets

### Phase 4 — Document Editing

**Goal:** Edit individual documents.

Deliverables:
- JSON editor with validation
- Save and revert

### Phase 5 — Usability

**Goal:** Polish for regular use.

Deliverables:
- Keyboard shortcut map
- Persistent layout state
- Compact/comfortable density
- English and Japanese localisation

### Phase 6 — Quality & Security

**Goal:** Hardened, well-tested application.

Deliverables:
- Comprehensive test coverage
- Security review
- Penetration test of Meilisearch communication layer
- Performance profiling

### Phase 7 — Release

**Goal:** Public v1.0.0 release.

Deliverables:
- Final packaging
- Code signing (optional)
- Release documentation
- GitHub Release with signed artifacts

---

## 17. Technical Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| Wails v3 API change in minor release | Low | Medium | Pin version; review changelog before updating |
| WebView2 not present on user machine | Very Low | High | Wails bootstrapper installs silently; document as prerequisite |
| WebView2 deprecated by Microsoft | Very Low | High | Would require framework change; no mitigation currently needed |
| Go module dependency vulnerability | Low | Medium | Dependabot automated updates |
| npm dependency vulnerability | Low | Medium | Dependabot automated updates |
| Windows Credential Manager API change | Very Low | High | Use stable `danieljoos/wincred` wrapper |

---

## 18. Assumptions

1. Users have Windows 10 1803 or later (WebView2 pre-installed).
2. Meilisearch instances are accessible over HTTP or HTTPS from the user's machine.
3. The Go standard library's HTTP client is sufficient for Meilisearch communication without a dedicated SDK.
4. TypeScript auto-generated bindings from Wails provide sufficient type safety without manual duplication.

---

## 19. Outstanding Decisions

| Decision | Status | Notes |
|----------|--------|-------|
| GUI framework | ✓ Resolved | Wails v3.1.0 (ADR 0001) |
| Credential storage implementation | Pending | Interface defined; Windows Credential Manager planned for Phase 1 |
| Meilisearch API client | Pending | Custom Go HTTP client vs. third-party SDK — Phase 1 decision |
| Code signing certificate | Pending | Required before public release (Phase 7) |
| Windows version minimum floor | Pending | Windows 10 1803 assumed; verify with WebView2 bootstrapper test |
| NSIS installer vs. portable ZIP | Pending | Portable ZIP for initial releases; installer post-MVP |

---

## 20. Document History

| Version | Date | Changes |
|---------|------|---------|
| 0.1.0 | (initial) | Initial proposal |
| 0.2.0 | 2026-11-16 | Phase 0: Wails v3 validated; technology stack finalised; distribution model documented; roadmap aligned |
