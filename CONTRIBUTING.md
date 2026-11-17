# Contributing to MeiliScope

Thank you for your interest in contributing to MeiliScope.

---

## Ways to Contribute

- **Bug reports** — open a [GitHub Issue](https://github.com/taqu/meiliscope/issues)
- **Feature requests** — open a GitHub Issue describing the use case
- **Documentation improvements** — submit a pull request against `docs/`
- **Code contributions** — see the development setup below

---

## Before You Start

For non-trivial changes, please open an issue first to discuss the approach. This avoids duplicate work and ensures the change aligns with the project goals.

Phase 7 and later: MeiliScope is in release-preparation mode. Focus on bug fixes, documentation, and test coverage. Large feature additions may be deferred to a future phase.

---

## Development Setup

See [Building from Source](docs/en/development/building.md) for complete instructions.

**Quick setup:**
```bash
git clone https://github.com/taqu/meiliscope.git
cd meiliscope
cd frontend && npm install && cd ..
wails3 dev
```

---

## Pull Request Process

1. Fork the repository and create a branch from `main`.
2. Make your changes.
3. Run the full test suite:
   ```bash
   go test ./internal/...
   cd frontend && npm test
   ```
4. Run `gofmt -l .` and ensure there are no unformatted files.
5. Run `go vet ./...` and ensure there are no warnings.
6. Update documentation in `docs/` if your change affects user-facing behavior.
7. Open a pull request against `main` with a clear description.

All CI checks must pass before a pull request is merged.

---

## Code Style

**Go:**
- Format with `gofmt` (enforced in CI).
- Follow standard Go conventions.
- Keep functions focused and testable.
- Do not introduce new external dependencies without discussion.

**TypeScript / React:**
- Use TypeScript strict mode.
- Do not use `dangerouslySetInnerHTML`.
- Prefer functional components and hooks.

---

## Security Issues

Do not report security vulnerabilities in public GitHub Issues. See [SECURITY.md](SECURITY.md).

---

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md).
