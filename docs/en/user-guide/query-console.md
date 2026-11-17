# Query Console

The Query Console (Ctrl+3) lets you execute Meilisearch search queries against the selected index.

---

## Editor Modes

The Query Console has two editor modes:

| Mode | Description |
|---|---|
| **JSON** | Full JSON editor with syntax highlighting, bracket matching, and formatting |
| **Builder** | Form-based editor for common search parameters |

Toggle between modes using the **JSON / Builder** button group in the toolbar. Switching modes preserves your query.

---

## JSON Editor

The JSON editor accepts any valid Meilisearch search request body.

Minimal query (returns all documents):
```json
{}
```

Typical search query:
```json
{
  "q": "coffee arabica",
  "filter": "price < 20",
  "sort": ["price:asc"],
  "limit": 20
}
```

**Toolbar actions:**
- **Format** (Shift+Alt+F) — pretty-print the JSON
- **Clear** — reset the editor to `{}`
- **Reset** — restore the last successfully executed query

---

## Search Builder

See [Search Builder](search-builder.md) for details.

---

## Executing a Query

Click **Search** or press **Ctrl+Enter** to run the query.

The results panel shows:
- **Total hits** — the estimated total number of matching documents
- **Processing time** — time reported by Meilisearch
- **Results list** — matching documents as rows; click any row to open the JSON Inspector

---

## Query History

Every executed query is saved to the query history automatically. Click the **History** tab in the results panel to browse past queries. Click any history entry to load it into the editor.

---

## Saved Queries

Click **★ Saved** (or press Ctrl+S) to open the Saved Queries panel. See [Saved Queries](saved-queries.md).

---

## Exporting Results

Click **Export** in the results tabs bar to save the current result set. See [Export](export.md).

---

## Limitations

- The Query Console executes read-only search requests. It does not support write operations.
- Very large result sets are truncated by Meilisearch's `limit` parameter. Set a higher `limit` in your query to retrieve more results.
- Ranking score display requires Meilisearch v1.3 or later.
