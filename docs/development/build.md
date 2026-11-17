# Build Guide

This document describes how to build MeiliScope from source on Windows.

---

## Prerequisites

### For contributors (building from source)

| Tool | Minimum version | Install |
|------|----------------|---------|
| Go | 1.26.0 | https://go.dev/dl/ |
| Node.js | 20.0.0 | https://nodejs.org/ |
| npm | 10.0.0 | Included with Node.js |
| Wails CLI | v3.1.0 | `go install github.com/wailsapp/wails/v3/cmd/wails3@v3.1.0` |

No C compiler is required. No Docker is required.

> **Note on GOPATH/GOROOT:** If Go is installed such that GOPATH == GOROOT, the Go toolchain prints a warning on each invocation. This does not affect builds. Refer to [Go install troubleshooting](https://go.dev/wiki/InstallTroubleshooting) for the fix.

### For end users (running the application)

- Windows 10 version 1803 (Build 17134) or later, or Windows 11
- Microsoft Edge WebView2 Runtime (pre-installed on Windows 10 1803+ and all Windows 11 via Windows Update)

---

## Clone and set up

```powershell
git clone https://github.com/taqu/meiliscope.git
cd meiliscope

# Install Wails CLI
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.1.0

# Install frontend dependencies
cd frontend
npm install
cd ..
```

---

## Common commands

### Development mode (hot reload)

```powershell
wails3 dev
```

This starts the Vite dev server and rebuilds the Go binary on file changes. The application window opens automatically.

### Run Go tests

```powershell
go test ./internal/...
```

> **Race detector:** The Go race detector requires `CGO_ENABLED=1`, which is not used in this project (Wails v3 uses `CGO_ENABLED=0` on Windows). Tests run without `-race`.

### TypeScript type check and frontend build

```powershell
cd frontend
npm run build
```

### Production build (Windows x64 executable)

```powershell
wails3 build
```

Output: `bin/meiliscope.exe`

The build process:
1. Regenerates TypeScript bindings from Go service structs
2. Installs npm dependencies (if changed)
3. Compiles TypeScript and bundles with Vite
4. Embeds frontend assets into the Go binary
5. Compiles `go build -tags production -trimpath -ldflags="-w -s -H windowsgui"`

### Check system requirements

```powershell
wails3 doctor
```

### Clean build artifacts

```powershell
Remove-Item -Recurse -Force bin, frontend/dist, frontend/bindings
```

---

## Project structure

```
MeiliScope/
├── main.go                    # Application entry point
├── go.mod                     # Go module definition
├── Taskfile.yml               # Wails task runner
├── build/                     # Build configuration and icons
│   ├── config.yml             # Wails build config
│   └── windows/               # Windows-specific build files
├── internal/
│   ├── app/                   # Application lifecycle, logging
│   ├── config/                # JSON config persistence (%AppData%)
│   ├── meilisearch/           # (Phase 1) Meilisearch HTTP client
│   ├── platform/              # Windows OS integrations
│   └── services/              # Wails-bound backend services
├── frontend/
│   ├── src/
│   │   ├── App.tsx            # Application shell
│   │   ├── styles.css         # Global styles
│   │   ├── components/        # Shared UI components
│   │   ├── features/          # View-specific modules
│   │   ├── services/          # TypeScript wrappers for Go bindings
│   │   └── types/             # Shared TypeScript types
│   ├── bindings/              # Auto-generated from Go (gitignored)
│   ├── package.json
│   └── package-lock.json      # Committed lockfile
├── docs/
│   ├── architecture/          # Architecture decisions and overview
│   ├── development/           # Developer documentation
│   ├── en/                    # English end-user documentation
│   └── ja/                    # Japanese end-user documentation
├── .github/
│   ├── workflows/
│   │   ├── ci.yml             # CI pipeline
│   │   └── release.yml        # Release pipeline
│   └── dependabot.yml
└── experiments/
    └── wails-v3/              # Phase 0 PoC (preserved for reference)
```

---

## Configuration file location

User settings are stored in `%AppData%\MeiliScope\config.json`.  
Logs are written to `%AppData%\MeiliScope\logs\`.

These directories are created on first launch and are independent of the executable location.

---

## CI/CD

### Continuous integration

GitHub Actions triggers on pushes to `main` and pull requests.

The `go-checks` job runs:
- `gofmt` formatting check
- `go vet` static analysis
- `go test ./internal/...`

The `build` job runs the full `wails3 build`, which includes:
- TypeScript binding generation
- TypeScript type check (`tsc`)
- Vite production frontend build
- Windows x64 Go application build

### Releases

Tag a commit with `v<major>.<minor>.<patch>` (e.g., `v0.2.0`) to trigger the release workflow. The workflow:
1. Builds the production executable
2. Creates `MeiliScope_<version>_windows_amd64.zip` containing `meiliscope.exe`, `LICENSE`, `README.md`
3. Generates a SHA-256 checksum file
4. Publishes to GitHub Releases

> **Do not create version tags to test the release workflow.** Test packaging locally with the commands above.

---

## Known limitations

- The Go race detector (`-race`) is not available because `CGO_ENABLED=0` is required by Wails v3 on Windows.
- `wails3 dev` launches a window that must be closed manually; there is no headless test mode in Phase 0.
- The `frontend/bindings/` directory is regenerated during each build and is gitignored. It is not committed to the repository.
