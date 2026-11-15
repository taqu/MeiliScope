# ADR 0001 — Desktop GUI Framework

**Status:** Accepted  
**Date:** 2026-11-16  
**Deciders:** Phase 0 evaluation

---

## Context

MeiliScope is a Windows-first standalone desktop application for browsing Meilisearch indexes and documents. It requires:

- A native desktop window on Windows x64
- A React + TypeScript frontend
- Go as the backend language
- A single distributable executable with no separately installed runtime beyond what Windows provides
- No CGO/C compiler dependency for end-user builds or CI
- MIT-compatible licensing

The initial candidate was **Wails v3**, previously unvalidated. Phase 0 required an evidence-based decision before committing to any production application structure.

---

## Requirements

| # | Requirement | Priority |
|---|-------------|----------|
| R1 | Go backend | Mandatory |
| R2 | React + TypeScript frontend | Mandatory |
| R3 | Windows x64 single executable | Mandatory |
| R4 | No end-user Go/Node.js/Python/.NET install required | Mandatory |
| R5 | Frontend assets embedded in executable | Mandatory |
| R6 | Async Go ↔ frontend communication | Mandatory |
| R7 | No CGO (no C compiler in CI) | Mandatory |
| R8 | File dialogs, clipboard, keyboard shortcuts | Mandatory |
| R9 | Cancellable async operations | Mandatory |
| R10 | MIT-compatible licensing | Mandatory |

---

## Evaluated Alternatives

### Option A — Wails v3 (selected)

Wails v3 is a Go-native framework that embeds a WebView2 window and auto-generates TypeScript bindings from Go struct methods.

**Tested version:** v3.1.0  
**License:** MIT

Pros:
- No CGO required (uses `ebitengine/purego` for Win32 calls)
- Frontend assets embedded via Go `embed` package
- Auto-generated TypeScript bindings with `CancellablePromise`
- Official React + TypeScript template
- Single executable output (10.85 MB for PoC)
- WebView2 pre-installed on Windows 10 1803+ and all Windows 11
- `wails3 doctor` confirms system readiness
- Production build in ~25 seconds with standard Go toolchain

Cons:
- v3.1.0: post-alpha, actively developed — API may evolve in minor releases
- WebView2 required (OS runtime prerequisite; pre-installed on modern Windows)
- No production DevTools (dev-mode DevTools available)

### Option B — Fyne

Go-native toolkit using an OpenGL/software renderer — no WebView2.

Pros: Pure Go, no WebView2 dependency, works offline.  
Cons: Limited support for complex web-style UI (Monaco, virtual tables), no React/TypeScript, custom rendering engine diverges from browser standards.

**Not selected:** React + TypeScript is a mandatory requirement.

### Option C — Tauri v2 (Rust + WebView2)

Similar architecture to Wails but uses Rust for the backend.

Pros: Mature v2.0 release, strong Windows support.  
Cons: Requires Rust toolchain; MeiliScope requires Go backend; Rust/Go interop adds complexity.

**Not selected:** Go backend is a mandatory requirement.

### Option D — Electron

Node.js + Chromium desktop framework.

Pros: Mature ecosystem, excellent React support.  
Cons: Bundles Chromium (~150 MB), requires Node.js, mixed Go/Node runtime model is complex.

**Not selected:** No separately installed Node.js runtime is a mandatory requirement; distribution size is prohibitive.

---

## Technical Findings

From the Wails v3 feasibility study (`docs/architecture/wails-v3-evaluation.md`):

1. `wails3 doctor` confirmed: Windows 11 Pro 25H2, Go 1.27.1, CGO_ENABLED=0, WebView2 157.0.4341.46, npm 11.19.1 — all ready.
2. PoC built successfully: single `wails-v3.exe`, 10.85 MB, embedded assets, embedded version info.
3. Frontend-to-Go calls use auto-generated TypeScript with `CancellablePromise<T>` — full async + cancellation.
4. Go-to-frontend uses typed event constants generated alongside method bindings.
5. Build uses `-tags production -trimpath -ldflags="-w -s -H windowsgui"` — no console window, no debug paths.
6. No CGO required; standard `windows-latest` GitHub Actions runner is sufficient.
7. All mandatory requirements (R1–R10) are satisfied.

---

## Decision

**Adopt Wails v3.1.0 as the MeiliScope desktop framework.**

The mandatory requirement check passes on all ten items. The two identified conditions are acceptable and do not block development:

- WebView2 is an OS runtime prerequisite that is pre-installed on all target platforms (Windows 10 1803+, Windows 11). It must be documented clearly in installation instructions.
- API evolution risk is low for the core APIs used (window management, service binding, events, dialogs). The PoC-validated APIs are stable.

---

## Consequences

**Positive:**
- No C compiler in CI or on developer machines
- Single executable distribution
- Auto-generated TypeScript bindings eliminate manual type maintenance
- Standard React toolchain (Vite, TypeScript, npm) for frontend development
- Straightforward portable ZIP packaging

**Negative / Accepted Risks:**
- WebView2 must be documented as an OS prerequisite
- WebView2 bootstrapper requires internet access on machines without it
- Wails v3 minor-version updates must be reviewed for API changes

**Architecture constraints established by this decision:**
- Go owns all HTTP communication with Meilisearch
- The frontend communicates with Go exclusively through Wails-generated service bindings and events
- Application services must not depend on GUI components (testable in isolation)
- Credentials must not be passed through the Wails binding layer to the frontend

---

## Known Limitations

- Clean-machine WebView2 bootstrapper behaviour has not been tested (marked as release prerequisite)
- Production DevTools unavailable without custom build flag
- GOPATH == GOROOT warning on the build machine is a machine configuration issue, not a framework issue

---

## Reconsideration Criteria

Reconsider this decision if:

1. Wails v3 introduces a breaking change to service binding or window APIs that cannot be resolved without significant rework.
2. WebView2 is deprecated or removed from Windows as a platform component.
3. A mandatory feature requirement (e.g., OS-level accessibility integration) cannot be satisfied through the WebView2/Go boundary.
4. The project needs to target Windows versions before Windows 10 1803 as a supported platform.
