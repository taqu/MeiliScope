# MeiliScope

**Desktop browser for Meilisearch — Windows & Linux**

MeiliScope is an open-source desktop application for exploring Meilisearch indexes and documents, executing search queries, and performing controlled document editing. No server administration privileges required — a search key is enough to start browsing.

**Mission: Explore your Meilisearch data.**

---

## Features

- **Connection Manager** — save and switch between multiple Meilisearch servers; credentials stored in the OS keychain (Windows Credential Manager / Linux Secret Service)
- **Index Explorer** — browse all indexes, view statistics and settings at a glance
- **Document Browser** — paginate through raw documents; dynamic column layout
- **JSON Inspector** — interactive collapsible tree for any document
- **Query Console** — full-featured JSON query editor with syntax highlighting and query history
- **Search Builder** — form-based query editor; set filters, sort, facets, and pagination without writing JSON
- **Saved Queries** — persist named queries per connection
- **Document Editing** — create, partial-update, replace, and delete documents with diff review and task tracking
- **JSON / CSV Export** — export the current page or search results; CSV formula injection protection built in
- **Keyboard Shortcuts** — navigate the entire application without touching the mouse

---

## Supported Platforms

| Platform | Architecture | Status |
|---|---|---|
| Windows 10 1803+ / Windows 11 | x64 | **Supported** |
| Linux (Ubuntu 22.04+ or equivalent) | x64 | **Supported** (requires GTK3 + WebKit2GTK) |

macOS and ARM are not supported in this release.

---

## Download

Download the latest release from [GitHub Releases](https://github.com/taqu/meiliscope/releases).

**Windows:** Extract `MeiliScope_vX.Y.Z_windows_amd64.zip` and run `meiliscope.exe`. No installer required.

**Linux:** Extract `MeiliScope_vX.Y.Z_linux_amd64.tar.gz`, ensure the binary is executable, and run `./meiliscope`. GTK3 and WebKit2GTK must be installed — see [Installation](docs/en/getting-started/installation.md).

Verify the SHA-256 checksum using `SHA256SUMS.txt` before running.

---

## Runtime Prerequisites

**Windows:**
- Windows 10 version 1803 (Build 17134) or later, or Windows 11
- Microsoft Edge WebView2 (pre-installed on supported Windows versions; auto-downloaded on first launch if absent)
- No Go, Node.js, or other development tools required

**Linux:**
- GTK 3: `libgtk-3-0`
- WebKit2GTK 4.0: `libwebkit2gtk-4.0-37`
- A running desktop session with a display server

Install on Debian/Ubuntu:
```bash
sudo apt-get install libgtk-3-0 libwebkit2gtk-4.0-37
```

---

## Quick Start

1. Launch MeiliScope.
2. Open **Connections** (Ctrl+Shift+C) and click **Add**.
3. Enter the Meilisearch URL (e.g. `http://localhost:7700`) and an API key if required.
4. Click **Connect** — the connection is verified and saved.
5. Select an index from the **Index Explorer** (Ctrl+1).
6. Browse documents in the **Document Browser** (Ctrl+2).
7. Click any document to open the **JSON Inspector**.
8. Open the **Query Console** (Ctrl+3) to execute search queries.

See [Quick Start Guide](docs/en/getting-started/quick-start.md) for a detailed walkthrough.

---

## Documentation

| Document | Description |
|---|---|
| [Installation](docs/en/getting-started/installation.md) | Windows and Linux installation guide |
| [Quick Start](docs/en/getting-started/quick-start.md) | First connection and basic usage |
| [User Guide](docs/en/user-guide/) | All features documented |
| [Keyboard Shortcuts](docs/en/user-guide/keyboard-shortcuts.md) | Full shortcut reference |
| [Building from Source](docs/en/development/building.md) | Developer build instructions |
| [Testing](docs/en/development/testing.md) | Running the test suite |
| [Security Model](docs/architecture/security.md) | Threat model and security properties |

**Japanese documentation:** [docs/ja/](docs/ja/)

---

## Development

### Prerequisites

| Tool | Version |
|---|---|
| Go | 1.26.0 or later |
| Node.js | 20.0.0 or later |
| Wails CLI | v3.1.0 |

Install Wails:
```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.1.0
```

**Linux only:** Install system dependencies:
```bash
sudo apt-get install libgtk-3-dev libwebkit2gtk-4.0-dev pkg-config libglib2.0-dev
```

### Setup

```bash
git clone https://github.com/taqu/meiliscope.git
cd meiliscope
cd frontend && npm install && cd ..
```

### Commands

| Command | Description |
|---|---|
| `wails3 dev` | Development mode with hot reload |
| `go test ./internal/...` | Run Go unit tests |
| `cd frontend && npm test` | Run frontend tests (Vitest) |
| `wails3 build` | Production build → `bin/meiliscope.exe` / `bin/meiliscope` |
| `wails3 doctor` | Check system requirements |

See [Building from Source](docs/en/development/building.md) for full details.

---

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

## Security

Report security vulnerabilities privately via [GitHub Security Advisories](https://github.com/taqu/meiliscope/security/advisories/new).

See [SECURITY.md](SECURITY.md) for the full policy.

---

## License

[MIT](LICENSE) — Copyright (c) 2026 taqu
