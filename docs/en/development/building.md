# Building from Source

---

## Prerequisites

| Tool | Version | Install |
|---|---|---|
| Go | 1.26.0 or later | [go.dev/dl](https://go.dev/dl/) |
| Node.js | 20.0.0 or later | [nodejs.org](https://nodejs.org/) |
| Wails CLI | v3.1.0 | `go install github.com/wailsapp/wails/v3/cmd/wails3@v3.1.0` |

Verify your environment:
```bash
wails3 doctor
```

**Linux only** — install system dependencies:
```bash
sudo apt-get install \
  libgtk-3-dev \
  libwebkit2gtk-4.0-dev \
  pkg-config \
  libglib2.0-dev
```

---

## Clone and Set Up

```bash
git clone https://github.com/taqu/meiliscope.git
cd meiliscope
cd frontend && npm install && cd ..
```

---

## Development Mode

Runs the application with hot reload. Frontend changes are reflected immediately without restarting.

```bash
wails3 dev
```

---

## Production Build

```bash
wails3 build
```

Output:
- **Windows:** `bin/meiliscope.exe`
- **Linux:** `bin/meiliscope`

The build process runs:
1. TypeScript binding generation
2. `npm ci` — installs exact dependency versions from lockfile
3. `tsc` — TypeScript type checking
4. `vite build` — production frontend bundle
5. `go build -tags production` — Go application with embedded frontend

---

## CGO

The Go race detector (`-race`) requires CGO. MeiliScope's standard build and test suite do not require CGO. The Windows build uses `CGO_ENABLED=0`.

The Linux build also uses `CGO_ENABLED=0` for the Go logic, but Wails links against the GTK/WebKit system libraries through its own build mechanism.

---

## Go Modules

Dependencies are managed with Go Modules. The `go.sum` lockfile is committed. Run `go mod tidy` before committing if you add or remove dependencies.

---

## npm Dependencies

Frontend dependencies are managed with npm. The `package-lock.json` lockfile is committed. Use `npm ci` (not `npm install`) in CI and release builds.

---

## Taskfile

A `Taskfile.yml` is provided for common development tasks. Install [Task](https://taskfile.dev) to use it.

---

## Notes

- Do not use `wails3 dev` to produce release packages. Always use `wails3 build`.
- The release workflow uses `wails3 build` with the exact toolchain versions specified in the CI configuration.
- The build embeds the frontend assets into the binary. No separate frontend deployment is needed.
