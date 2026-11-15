# MeiliScope

**Desktop browser for Meilisearch — Windows x64**

MeiliScope is an open-source desktop application for exploring Meilisearch indexes and documents, executing search queries, and performing limited document editing.

> **Phase 0** — Application foundation. Core features (Connection Manager, Index Explorer, Document Browser, Query Console) are implemented in subsequent phases.

---

## Download

Download the latest release from [GitHub Releases](https://github.com/taqu/meiliscope/releases).

Extract the ZIP and run `meiliscope.exe`. No installation required.

**Requirements:** Windows 10 version 1803 or later (or Windows 11). WebView2 is pre-installed on these platforms.

---

## Development

### Prerequisites

- Go 1.26.0 or later
- Node.js 20.0.0 or later
- Wails CLI v3.1.0: `go install github.com/wailsapp/wails/v3/cmd/wails3@v3.1.0`

No C compiler required.

### Setup

```powershell
git clone https://github.com/taqu/meiliscope.git
cd meiliscope
cd frontend && npm install && cd ..
```

### Commands

| Command | Description |
|---------|-------------|
| `wails3 dev` | Development mode with hot reload |
| `go test ./internal/...` | Run Go unit tests |
| `wails3 build` | Production build → `bin/meiliscope.exe` |
| `wails3 doctor` | Check system requirements |

See [docs/development/build.md](docs/development/build.md) for full documentation.

---

## License

[MIT](LICENSE)
