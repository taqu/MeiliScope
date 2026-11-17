# Changelog

All notable changes to MeiliScope are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

---

## [1.0.0] — 2026-11-17

### Added

**Connection Manager**
- Save and switch between multiple Meilisearch server connections
- API keys stored in the OS keychain (Windows Credential Manager / Linux Secret Service)
- Fallback to session-only credential mode when the keychain is unavailable
- Connection health check and status indicator
- Limited-access mode for search-only credentials (manual index UID entry)

**Index Explorer**
- Browse all indexes for the active connection
- View document count, index size, primary key, field distribution, and indexing status

**Document Browser**
- Paginate through documents in any index
- Dynamic column layout based on returned fields
- JSON Inspector for any selected document (collapsible tree)
- Export current page to JSON or CSV

**Query Console**
- Full-featured JSON query editor with syntax highlighting and bracket matching
- Execute Meilisearch search queries against any index
- Display total hits, processing time, and ranked results
- Query History (automatic, persistent)

**Search Builder**
- Form-based query editor for query, filter, sort, facets, pagination, and other common parameters
- Partial editing: preserves unknown JSON fields

**Saved Queries**
- Persist named queries per connection
- Load, update, and delete saved queries
- Local storage only; never uploaded to external services

**Document Editing**
- Create new documents
- Partial update (PUT): modify specific fields without overwriting others
- Replace (POST): overwrite the full document
- Delete document by primary key
- Diff review before every write
- Explicit confirmation step; mutation target locked at confirmation
- Task tracking after every write

**JSON / CSV Export**
- Export current document page or search results to JSON or CSV
- Native OS file dialog for save location
- CSV formula injection protection (tab prefix for cells starting with `=`, `+`, `-`, `@`)

**Keyboard Shortcuts**
- Ctrl+1 / 2 / 3: navigate between views
- Ctrl+Shift+C: open Connections
- Ctrl+Enter: execute search
- Ctrl+S: toggle Saved Queries panel
- Shift+Alt+F: format JSON

**Platform Support**
- Windows x64 (Windows 10 version 1803 or later, Windows 11)
- Linux x64 (Ubuntu 22.04+ with GTK3 and WebKit2GTK 4.0)

**Security**
- Cross-host redirect protection: Authorization header stripped on cross-host redirects
- Dedicated per-client HTTP transport (no shared DefaultTransport)
- TLS certificate verification always enabled
- Response size limit (10 MiB)
- API keys never logged

---

[Unreleased]: https://github.com/taqu/meiliscope/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/taqu/meiliscope/releases/tag/v1.0.0
