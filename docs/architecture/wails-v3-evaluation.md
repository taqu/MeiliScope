# Wails v3 Feasibility Evaluation

**Evaluation date:** 2026-11-16  
**Evaluator:** Phase 0 — Step 0.1  
**Decision:** ACCEPT WITH CONDITIONS

---

## 1. Test Environment

| Item | Value |
|------|-------|
| OS | Windows 11 Pro 25H2 (Build 26200) |
| Architecture | amd64 |
| Go version | 1.27.1 windows/amd64 |
| Node.js version | 24.22.1 |
| npm version | 11.19.1 |
| Wails CLI version | v3.1.0 |
| WebView2 version | 157.0.4341.46 |
| CPU | Intel Core i5-14500 |
| GPU | NVIDIA GeForce RTX 4070 Ti SUPER |

---

## 2. Tested Versions

| Component | Version |
|-----------|---------|
| `github.com/wailsapp/wails/v3` | v3.1.0 |
| `@wailsio/runtime` | 3.1.0 |
| React | 18.2.0 |
| TypeScript | 5.2.2 |
| Vite | 8.3.8 |

---

## 3. Commands Used

```
# Install Wails v3 CLI
go install github.com/wailsapp/wails/v3/cmd/wails3@latest

# System requirements check
wails3 doctor

# Create React + TypeScript PoC
wails3 init -n wails-v3 -t react -d wails-v3 -mod github.com/taqu/meiliscope-poc \
  -productname "MeiliScope PoC" -productversion "0.0.1" -s

# Install frontend dependencies
cd frontend && npm install

# Production build
wails3 build
```

---

## 4. Proof-of-Concept Results

The PoC was created under `experiments/wails-v3/` using the official `react` template (React + TypeScript + Vite).

### 4.1 Application Startup — PASSED

The application initialises via `application.New(application.Options{...})`, then opens a window with `app.Window.NewWithOptions(...)`, and blocks on `app.Run()`.

### 4.2 Frontend Rendering — PASSED

React components render correctly inside the WebView2 window. Vite compiles TypeScript to ES modules; the bundle is embedded in the Go binary via `//go:embed all:frontend/dist`.

### 4.3 Frontend-to-Go Communication — PASSED

Auto-generated TypeScript bindings call Go methods by ID:

```typescript
// Generated binding (frontend/bindings/…/greetservice.ts)
export function Greet(name: string): $CancellablePromise<string> {
    return $Call.ByID(1411160069, name);
}
```

Calling `GreetService.Greet("world")` from React invokes the Go `GreetService.Greet` method and returns a `CancellablePromise<string>`.

### 4.4 Go-to-Frontend Response Handling — PASSED

Go emits typed events with `app.Event.Emit("time", now)`. The frontend subscribes with `Events.On("time", handler)`. The event name constant is exported from the generated `bindings/events/index.ts`.

### 4.5 Error Handling — PASSED (via standard Promise rejection)

`CancellablePromise` implements the standard Promise interface. Errors from Go propagate as rejected promises: `.catch(console.error)` or `try/await` patterns work without special handling.

### 4.6 Application Shutdown — PASSED

`app.Run()` blocks until the last window is closed. Shutdown hooks can be registered. The goroutine that emits events exits naturally when the process terminates.

### 4.7 Production Executable Generation — PASSED

```
wails3 build
```

Output:
- `bin/wails-v3.exe` — single self-contained Windows x64 executable
- Size: **10.85 MB** (includes Go runtime, Wails framework, React frontend)
- Frontend assets embedded via Go `embed` package
- Windows version info embedded (product name, version, company)
- Build flags: `-tags production -trimpath -ldflags="-w -s -H windowsgui"`
  - `-H windowsgui` suppresses the console window
  - `-trimpath` removes source paths from debug info

---

## 5. Windows Runtime Dependencies

### 5.1 WebView2 Runtime

WebView2 is a **required operating-system runtime dependency**.

| Scenario | Behaviour |
|----------|-----------|
| Windows 11 (all versions) | Pre-installed, auto-updated via Windows Update |
| Windows 10 version 1803+ | Pre-installed, auto-updated via Windows Update |
| Windows 10 before 1803 | Manual installation required |
| `Go WebView2Loader: true` | Wails can bootstrap WebView2 silently if absent |

**Decision:** Document WebView2 as an OS prerequisite for pre-1803 Windows 10. For Windows 10 1803+ and all Windows 11, WebView2 is pre-installed. The `Go WebView2Loader` flag (`true` on this machine) means Wails can silently download and install WebView2 when it is absent, using the official Microsoft bootstrapper — no additional runtime file needs to ship with the application.

### 5.2 Other Runtime Files

| Dependency | Required at runtime? |
|------------|---------------------|
| Go installation | No — compiled into exe |
| Node.js | No — frontend compiled into exe |
| .NET runtime | No |
| Docker | No |
| Visual C++ Redistributable | No — CGO_ENABLED=0, no C runtime |
| Additional DLLs | No |

The production executable is fully self-contained except for WebView2.

---

## 6. Build and Packaging Findings

### 6.1 Build System

Wails v3 uses `Taskfile.yml` (similar to `make`). The `wails3 build` command:

1. Runs `go mod tidy`
2. Generates TypeScript bindings from Go source
3. Installs npm dependencies
4. Runs `tsc && vite build --mode production`
5. Generates a `.syso` file embedding Windows version info and icon
6. Runs `go build -tags production ...`
7. Removes the `.syso` file

### 6.2 CGO Status

`CGO_ENABLED=0` throughout. Wails v3 uses `github.com/ebitengine/purego` to call Win32 APIs directly from Go without a C compiler. This is a significant improvement over Wails v2 and simplifies CI/CD configuration.

### 6.3 Frontend Asset Embedding

The frontend build output (`frontend/dist`) is embedded using Go's `embed` package:

```go
//go:embed all:frontend/dist
var assets embed.FS
```

No separate asset server or external files are needed.

### 6.4 Binding Generation

TypeScript bindings are regenerated during each build. The generator:
- Introspects Go service structs
- Assigns stable integer IDs to methods (by hashing the qualified name)
- Generates `.ts` files with correct TypeScript types
- Exports event name constants as typed strings

No manual TypeScript type duplication is needed.

### 6.5 Distribution

A single `.exe` file is the complete distribution unit. ZIP packaging is straightforward:

```
MeiliScope_<version>_windows_amd64.zip
├── meiliscope.exe
├── LICENSE
└── README.md
```

NSIS installer support is available but not required for portable distribution.

---

## 7. Evaluation Matrix

| Requirement | Support | Notes |
|-------------|---------|-------|
| Go backend | ✓ Full | Native Go, no CGO required |
| React + TypeScript | ✓ Full | Official template, Vite build |
| Windows x64 | ✓ Full | Native target |
| Single executable | ✓ Full | Embedded assets, 10.85 MB |
| No Go runtime at user machine | ✓ Full | Compiled binary |
| No Node.js at user machine | ✓ Full | Frontend embedded |
| No .NET at user machine | ✓ Full | No dependency |
| No separate server | ✓ Full | In-process WebView2 |
| Resizable windows | ✓ Full | Configurable via WindowOptions |
| Keyboard shortcuts | ✓ Full | Standard browser keyboard events |
| Clipboard operations | ✓ Full | Browser Clipboard API + Go `wails3/pkg/clipboard` |
| File open/save dialogs | ✓ Full | `app.Dialog.OpenFile()`, `SaveFile()` |
| Asynchronous backend ops | ✓ Full | CancellablePromise |
| Context/cancellation | ✓ Full | CancellablePromise.Cancel() |
| Frontend assets embedded | ✓ Full | Go embed package |
| Windows credential storage | ✓ Via Go | Use `github.com/danieljoos/wincred` (Windows Credential Manager) |
| Large data tables | ✓ Via React | No framework limit; virtual scrolling library needed for >10k rows |
| JSON editors | ✓ Via React | `@monaco-editor/react` or `react-json-view` integrate normally |
| Native notifications | ✓ Partial | `app.Notification` API or browser Notification API |
| Reliable shutdown | ✓ Full | `app.Run()` blocks; `OnBeforeClose` hook available |
| Auto-generated TS types | ✓ Full | Regenerated on each build |
| WebView2 prerequisite | ⚠ Condition | Pre-installed on Win10 1803+/Win11; bootstrappable if absent |
| Framework API stability | ⚠ Condition | v3.1.0 — post-alpha, actively developed |

---

## 8. Performance and Stability Observations

- Build time (clean, cold): ~25 seconds (npm install cached after first run)
- Incremental build (Go change only): ~5 seconds
- Executable startup time: fast (< 1 second on test machine with WebView2 cached)
- Memory footprint at idle: typical for a WebView2-based app (~80–120 MB resident)
- No crashes or hangs observed during the PoC build and testing

---

## 9. Known Limitations

1. **WebView2 bootstrapper:** Silent bootstrap requires internet access on machines without WebView2. Offline installation requires distributing the WebView2 offline installer separately.
2. **API evolution:** Wails v3 is at v3.1.0. The API has stabilised significantly since the alpha but breaking changes between minor versions remain possible until a formal stable (v3.x.0 LTS) designation.
3. **GOPATH/GOROOT warning:** The build machine has GOPATH == GOROOT. This is a machine configuration issue, not a Wails issue. It does not affect builds.
4. **WML (Wails Markup Language):** The `data-wml-openURL` attribute relies on `WML.Reload()` being called after render. This is a minor footgun that must be remembered for external link handling.
5. **Inspector / DevTools:** Chrome DevTools can be enabled in development mode (`-tags dev` or `application.Options{Debug: ...}`). Not available in production builds by default.

---

## 10. Maintenance Risk Assessment

| Risk Area | Level | Evidence |
|-----------|-------|----------|
| API stability | Low–Medium | v3.1.0; active development; core APIs (`application`, `Window`, `Service`) are stable |
| Dependency footprint | Low | 6 direct Go dependencies, all well-maintained |
| Build reproducibility | Low | Taskfile deterministic; pinned npm deps; no CGO |
| CI compatibility | Low | No CGO means standard `windows-latest` GitHub Actions runner works |
| Frontend/backend coupling | Low | Auto-generated bindings decouple manual type duplication |
| Debugging experience | Medium | No production DevTools; dev-mode DevTools work normally |
| Migration cost (if rejected later) | Medium | Go backend code is framework-agnostic; frontend bindings are thin wrappers |

---

## 11. Evidence References

- Wails v3.1.0 installed: `wails3 version` → `v3.1.0`
- `wails3 doctor` output: system ready, CGO_ENABLED=0, WebView2 157.0.4341.46
- `go.mod`: `require github.com/wailsapp/wails/v3 v3.1.0`, `go 1.26.0`
- Production build output: `bin/wails-v3.exe`, 10.85 MB
- Build succeeded with `-tags production -trimpath -ldflags="-w -s -H windowsgui"`
- Generated bindings confirm `CancellablePromise` and typed event constants
- `(Get-Item '...wails-v3.exe').VersionInfo` confirmed embedded version 0.0.1
- PoC source preserved at `experiments/wails-v3/`

---

## 12. Final Recommendation

**ACCEPT WITH CONDITIONS**

Wails v3.1.0 satisfies all mandatory requirements for MeiliScope:

- Single self-contained Windows x64 executable with embedded frontend assets
- No CGO, no C compiler, no .NET, no Node.js required at runtime
- React + TypeScript frontend with auto-generated type-safe bindings
- Asynchronous Go-to-frontend and frontend-to-Go communication with cancellation
- File dialogs, clipboard, and keyboard shortcuts supported natively
- WebView2 pre-installed on all supported Windows versions (10 1803+, 11)
- Portable ZIP distribution is straightforward

**Conditions before v1.0.0 release:**

1. **Document WebView2 as OS prerequisite** in installation instructions.
2. **Test WebView2 bootstrapper** on a machine without a pre-installed WebView2 runtime.
3. **Monitor Wails v3 changelog** for breaking API changes with each minor update.
4. **Validate on a clean Windows environment** (no developer tools) before the first public release.
