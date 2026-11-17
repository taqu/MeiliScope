# Testing

---

## Go Unit Tests

Run all Go unit tests:

```bash
go test ./internal/...
```

The test suite covers:

| Package | Tests |
|---|---|
| `internal/config` | Configuration serialization and defaults |
| `internal/meilisearch` | HTTP client, URL validation, redirect policy, response handling |
| `internal/platform` | Credential storage backends |
| `internal/services` | All service layer logic (connection, index, document, query, export, etc.) |

**Current baseline:** 232 tests (all PASS as of Phase 6).

The race detector is not used in the standard suite because it requires CGO which is disabled in the Windows build. Run with `-race` on Linux if needed for a specific investigation.

---

## Frontend Tests

Run Vitest frontend tests:

```bash
cd frontend
npm test
```

Tests are located in `src/**/*.test.ts` and `src/**/*.test.tsx`.

The test suite covers:

| Suite | Tests |
|---|---|
| `SearchBuilder.test.ts` | `patchQuery` and `parseQueryText` logic |

**Current baseline:** 20 tests (all PASS as of Phase 6).

---

## TypeScript Type Check

TypeScript type checking runs as part of `wails3 build`. To run it standalone:

```bash
cd frontend
npx tsc --noEmit
```

---

## CI

The GitHub Actions CI workflow (`ci.yml`) runs on every push and pull request to `main`:

| Job | Runner | Description |
|---|---|---|
| `go-checks` | ubuntu-latest | gofmt, go vet, go test |
| `frontend-tests` | ubuntu-latest | npm ci, npm test |
| `build-windows` | windows-latest | wails3 build, upload artifact |
| `build-linux` | ubuntu-latest | apt-get GTK deps, wails3 build, upload artifact |

All jobs must pass before a pull request can be merged.

---

## Running the Full Check Suite Locally

```bash
# Go checks
gofmt -l .                    # should produce no output
go vet ./...
go test ./internal/...

# Frontend
cd frontend
npm ci
npm test
npx tsc --noEmit

# Full production build
cd ..
wails3 build
```
