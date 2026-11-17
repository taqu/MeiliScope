# MeiliScope Compatibility Matrix

## 1. Policy

MeiliScope targets the **current stable release** of Meilisearch. Older releases are supported on a best-effort basis where the API surface has not changed.

The supported version range is determined by actual validation, not by assumption.

## 2. Meilisearch API Operations Used

| Operation | Endpoint | HTTP Method | Required Permission |
|---|---|---|---|
| Health check | `GET /health` | GET | None |
| Server version | `GET /version` | GET | None (usually) |
| List indexes | `GET /indexes` | GET | Index read |
| Get index | `GET /indexes/{uid}` | GET | Index read |
| Get index stats | `GET /indexes/{uid}/stats` | GET | Index read |
| List documents | `GET /indexes/{uid}/documents` | GET | Document read |
| Get document | `GET /indexes/{uid}/documents/{id}` | GET | Document read |
| Search | `POST /indexes/{uid}/search` | POST | Search |
| Add/replace documents | `POST /indexes/{uid}/documents` | POST | Document write |
| Update documents (partial) | `PUT /indexes/{uid}/documents` | PUT | Document write |
| Delete document | `DELETE /indexes/{uid}/documents/{id}` | DELETE | Document write |
| Get task | `GET /tasks/{uid}` | GET | Task read |

## 3. Validation Matrix

### Meilisearch v1.x (current stable line)

| Feature | v1.9 | v1.10 | v1.11 | Notes |
|---|---|---|---|---|
| Health check | PASS | PASS | NOT RUN | Standard across all v1 |
| Index enumeration | PASS | PASS | NOT RUN | |
| Document browsing | PASS | PASS | NOT RUN | |
| Search (basic) | PASS | PASS | NOT RUN | |
| Search (filters) | PASS | PASS | NOT RUN | |
| Search (sort) | PASS | PASS | NOT RUN | |
| Search (facets) | PASS | PASS | NOT RUN | |
| Search (ranking scores) | PASS | PASS | NOT RUN | `showRankingScore` available since v1.3 |
| Document creation | PASS | PASS | NOT RUN | |
| Document partial update | PASS | PASS | NOT RUN | |
| Document deletion | PASS | PASS | NOT RUN | |
| Task tracking | PASS | PASS | NOT RUN | |
| Search-only credentials | PASS | PASS | NOT RUN | May not list indexes |
| Unauthenticated (no key) | PASS | PASS | NOT RUN | Only for open instances |

**Validation platform:** Windows 11 Pro 10.0.26200, Go 1.26.0  
**Test method:** Automated unit tests with `httptest` mock server + manual smoke test  
**Last updated:** 2026-11-17

### Notes on Older Versions

- **Meilisearch v0.x**: Not tested; breaking API changes prevent compatibility claims.
- **Meilisearch v1.0 – v1.2**: Likely compatible for core features; ranking score details not available.
- **Task response format**: Stable since v1.0; `taskUid` field present.
- **Pagination `total` field**: Returns `int64`; MeiliScope uses `int64` to avoid truncation.

## 4. Known Incompatibilities

| Version | Issue | Severity | Workaround |
|---|---|---|---|
| v0.x | Different API paths and response formats | Critical | Not supported |
| v1.0–1.2 | `showRankingScore` parameter not available | Low | Parameter ignored by server; no error |

## 5. Permission Behaviour

| Credential Type | Health | List Indexes | Search | Write | Task Read |
|---|---|---|---|---|---|
| Master key | ✓ | ✓ | ✓ | ✓ | ✓ |
| Admin key | ✓ | ✓ | ✓ | ✓ | ✓ |
| Search key | ✓ | ✗ (403) | ✓ | ✗ | ✗ |
| No key (open instance) | ✓ | ✓ | ✓ | ✓ | ✓ |
| Invalid key | ✓ | ✗ (401/403) | ✗ | ✗ | ✗ |

When index enumeration fails with 403, MeiliScope sets the connection state to `limited_access` and allows manual index UID entry.

## 6. Updating This Matrix

Add a row to the validation table when:
- A new Meilisearch release is validated.
- An incompatibility is discovered.
- An incompatibility is resolved.

Do not add rows for versions that have not been tested.
