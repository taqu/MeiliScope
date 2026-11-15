# MeiliScope — Architecture Overview

**Version:** 0.1.0  
**Date:** 2026-11-16

---

## 1. High-Level Architecture

MeiliScope runs as a single Windows process. The process hosts:

1. A **Go application core** — business logic, configuration, Meilisearch HTTP client, logging
2. A **Wails v3 runtime** — WebView2 window management, service binding, event dispatch
3. A **React frontend** embedded as static assets — presentation, routing, user interaction

There is no separate server process, no background service, and no installer requirement.

```
┌─────────────────────────────────────────────────────────────────┐
│  meiliscope.exe (single process)                                │
│                                                                 │
│  ┌─────────────────────────────┐  ┌──────────────────────────┐ │
│  │  Go Backend                 │  │  WebView2 Window          │ │
│  │                             │  │                           │ │
│  │  ┌───────────────────────┐  │  │  ┌────────────────────┐  │ │
│  │  │  internal/app         │  │  │  │  React + TS        │  │ │
│  │  │  (lifecycle)          │◄─┼──┼──┤  (frontend/src)    │  │ │
│  │  └───────────────────────┘  │  │  │                    │  │ │
│  │                             │  │  │  Auto-generated    │  │ │
│  │  ┌───────────────────────┐  │  │  │  TS bindings       │  │ │
│  │  │  internal/services    │◄─┼──┼──┤  (CancellableProm) │  │ │
│  │  │  (AppService, etc.)   │  │  │  └────────────────────┘  │ │
│  │  └───────────────────────┘  │  │                           │ │
│  │                             │  │  Embedded via go:embed    │ │
│  │  ┌───────────────────────┐  │  │  (frontend/dist)          │ │
│  │  │  internal/meilisearch │  │  └──────────────────────────┘ │
│  │  │  (HTTP client, Ph.1+) │  │                               │
│  │  └───────────────────────┘  │        Wails v3 runtime       │
│  │                             │        (service bridge,       │
│  │  ┌───────────────────────┐  │         event dispatch)       │
│  │  │  internal/config      │  │                               │
│  │  │  (%AppData% JSON)     │  │                               │
│  │  └───────────────────────┘  │                               │
│  │                             │                               │
│  │  ┌───────────────────────┐  │                               │
│  │  │  internal/platform    │  │                               │
│  │  │  (Win Credential Mgr) │  │                               │
│  │  └───────────────────────┘  │                               │
│  └─────────────────────────────┘                               │
└─────────────────────────────────────────────────────────────────┘
         │
         │ HTTP / HTTPS
         ▼
  ┌──────────────┐
  │  Meilisearch │
  │  instance    │
  └──────────────┘
```

---

## 2. Module Boundaries

### 2.1 Application Layer (`internal/app`)

**Responsibility:** Application lifecycle and Wails integration.

- Constructs `application.Options` and registers services
- Owns the `app.Run()` call
- Coordinates startup and shutdown sequences
- Does not contain business logic

**Rule:** This package may import `internal/services`, `internal/config`, and `internal/platform`. It must not import `internal/meilisearch` directly.

### 2.2 Service Layer (`internal/services`)

**Responsibility:** Business operations exposed to the React frontend.

- Each service is a Go struct whose methods are exported as Wails bindings
- Methods return plain Go types; Wails serialises them to JSON automatically
- Error values are propagated as rejected promises in the frontend
- Services must not hold frontend state

**Rule:** Services may import `internal/config`, `internal/platform`, and `internal/meilisearch`. They must not import `internal/app`. They must be testable without a running Wails application.

**Current services (Phase 0):**
- `AppService` — application version, runtime info

**Planned services (Phase 1+):**
- `ConnectionService` — connection CRUD, test
- `IndexService` — list indexes, statistics
- `DocumentService` — browse, view documents
- `SearchService` — execute queries

### 2.3 Meilisearch Layer (`internal/meilisearch`)

**Responsibility:** HTTP communication with Meilisearch.

- HTTP client wrapping the Meilisearch REST API
- Request/response types
- Error mapping

**Rule:** This package must have no dependency on Wails, `internal/app`, or `internal/services`. It must be testable with a mock HTTP server.

**Status:** Interface defined; implementation deferred to Phase 1.

### 2.4 Configuration Layer (`internal/config`)

**Responsibility:** Non-secret application settings persistence.

- Reads and writes `%AppData%\MeiliScope\config.json`
- Handles schema versioning and corrupted config recovery
- Atomic file writes (write to temp, rename)
- Does not store API keys or other credentials

**Rule:** This package must have no dependency on Wails, `internal/app`, `internal/services`, or `internal/meilisearch`.

### 2.5 Platform Layer (`internal/platform`)

**Responsibility:** Windows-specific operating system integrations.

- Credential storage interface and Windows Credential Manager implementation
- Future: system notifications, OS-specific paths

**Rule:** The credential storage interface (`CredentialStore`) is defined here. The Windows implementation uses the Windows Credential Manager API. The interface must be testable with a mock implementation.

### 2.6 Frontend (`frontend/src`)

**Responsibility:** All UI rendering, user interaction, and presentation state.

- React components
- Application routing (simple state-based navigation in Phase 0)
- Calls Go services through auto-generated Wails TypeScript bindings
- Subscribes to Go events via the Wails runtime event API
- Manages loading states, error display, and empty states

**Rule:** The frontend must not construct or manage Meilisearch HTTP requests. All backend communication goes through Wails service bindings.

---

## 3. Communication Patterns

### 3.1 Frontend → Go (Method Calls)

Auto-generated TypeScript bindings call Go struct methods:

```typescript
// Generated by wails3 generate bindings
export function GetApplicationInfo(): $CancellablePromise<ApplicationInfo> {
    return $Call.ByID(/* hash */, ...);
}
```

All calls are asynchronous and return `CancellablePromise<T>`.

Errors from Go (returned as `error`, or panics caught by Wails) propagate as rejected promises.

### 3.2 Go → Frontend (Events)

Go emits typed events; TypeScript event name constants are generated alongside method bindings:

```go
// Go
app.Event.Emit("statusUpdate", StatusUpdate{Message: "connected"})
```

```typescript
// TypeScript (generated constant)
Events.On(StatusUpdateEvent, (data: StatusUpdate) => { ... })
```

### 3.3 Cancellation

`CancellablePromise` exposes a `.Cancel()` method. The Wails runtime cancels the in-flight Go goroutine via context cancellation when the frontend calls `.Cancel()`.

---

## 4. Data Storage

| Data type | Location | Format | Notes |
|-----------|----------|--------|-------|
| Application settings | `%AppData%\MeiliScope\config.json` | JSON | No secrets |
| Connection metadata | `%AppData%\MeiliScope\config.json` | JSON | URLs and names only |
| API keys | Windows Credential Manager | OS-encrypted | Never in config files |
| Log files | `%AppData%\MeiliScope\logs\` | Line-delimited text | Rotated per session |

---

## 5. Security Boundaries

```
Frontend (WebView2)                 Go Backend
┌─────────────────┐                ┌──────────────────────────┐
│                 │                │                          │
│  Displays data  │◄── bindings ──►│  Fetches from Meilisearch│
│  No API keys    │                │                          │
│  No raw HTTP    │                │  API keys in OS keychain │
│                 │                │  Never cross this line   │
└─────────────────┘                └──────────────────────────┘
```

- API keys are stored in and retrieved from the Windows Credential Manager by the Go backend only
- API keys are never serialised into binding return values
- The frontend never constructs Meilisearch HTTP requests
- Log files are sanitised before writing (no credentials, no request bodies with keys)

---

## 6. Build Architecture

```
Source                    Build steps                    Output
──────                    ───────────                    ──────
frontend/src/      →  tsc + vite build  →  frontend/dist/  ──┐
                                                               │ go:embed
Go source/         →  wails3 generate bindings               │
                   →  go build -tags production  ←────────────┘
                                                 →  bin/meiliscope.exe
```

The frontend is compiled to static assets in `frontend/dist/`. These are embedded into the Go binary at compile time. The resulting executable is self-contained.

---

## 7. Framework Version Lock

| Component | Pinned version | Location |
|-----------|---------------|----------|
| Wails v3 Go module | v3.1.0 | `go.mod` |
| `@wailsio/runtime` npm | 3.1.0 | `frontend/package.json` |
| React | ^18.2.0 | `frontend/package.json` |
| TypeScript | ^5.2.2 | `frontend/package.json` |
| Vite | ^8.2.4 | `frontend/package.json` |
| npm lockfile | committed | `frontend/package-lock.json` |

---

## 8. Dependency on Wails v3 Internals

MeiliScope depends on the Wails v3 public API only:

- `github.com/wailsapp/wails/v3/pkg/application` — window, app, options
- `github.com/wailsapp/wails/v3/pkg/events` — event names
- `@wailsio/runtime` — TypeScript runtime (Events, Call, WML)

Internal packages of Wails are not imported. This reduces migration surface if Wails changes its internal implementation.

---

## 9. Decisions Deferred to Later Phases

| Topic | Phase | Notes |
|-------|-------|-------|
| Meilisearch HTTP client design | 1 | Custom vs. SDK decision |
| Windows Credential Manager implementation | 1 | Interface ready; implementation pending |
| Application routing library | 1 | Simple state navigation sufficient for Phase 0 |
| Virtual table library for large datasets | 2 | Needed for Document Browser |
| JSON editor component | 4 | Needed for Document Editing |
