# Quick Start

This guide walks through the core MeiliScope workflow from first launch to your first search result.

---

## 1. Launch MeiliScope

**Windows:** Run `meiliscope.exe`.

**Linux:** Open a terminal and run `./meiliscope`.

The main window opens with the Index Explorer on the left and an empty content area on the right.

---

## 2. Add a Connection

1. Press **Ctrl+Shift+C**, or click the **Connections** button in the top bar.
2. Click **Add connection**.
3. Enter a display name, the server URL (e.g. `http://localhost:7700`), and an API key if your Meilisearch instance requires one.
4. Click **Connect**.

MeiliScope verifies the URL and credentials immediately. A green indicator confirms a successful connection.

> **Note:** A search-only key is sufficient for browsing and searching. Document editing and index listing require a key with the appropriate permissions. If index enumeration is blocked (HTTP 403), MeiliScope falls back to a manual index UID entry field.

---

## 3. Browse Indexes

After connecting, available indexes appear in the **Index Explorer** panel (Ctrl+1).

Click an index name to select it. The right panel shows:
- **Document count** and **field distribution** in the stats card.
- Primary key and index settings.

---

## 4. Browse Documents

Switch to the **Document Browser** tab (Ctrl+2) to page through raw documents.

- Use the **Previous / Next** buttons to paginate.
- Click any row to open the **JSON Inspector** for that document.
- The **Export** button saves the current page to JSON or CSV.

---

## 5. Inspect a Document

The **JSON Inspector** shows the full document as a collapsible tree. Expand nested objects and arrays by clicking the disclosure arrows.

---

## 6. Execute a Search Query

Switch to the **Query Console** tab (Ctrl+3).

The default query is `{}` — an empty search that returns all documents.

To search for a term:
```json
{
  "q": "coffee"
}
```

Click **Search** (or press Ctrl+Enter) to execute.

Results appear below the editor with the total hit count and timing.

---

## 7. Use the Search Builder

Click the **Builder** toggle in the Query Console toolbar to switch from the JSON editor to the form-based **Search Builder**.

Fill in fields such as **Query**, **Filter**, **Sort**, and **Limit**. The builder patches the underlying JSON without overwriting fields you have not set.

---

## 8. Save a Query

Click **★ Saved** in the Query Console toolbar to open the **Saved Queries** panel.

Enter a name and click **Save**. Saved queries are stored locally per connection and persist across sessions.

---

## 9. What's Read-Only and What Modifies Data

| Action | Modifies data |
|---|---|
| Browsing documents | No |
| Searching | No |
| Exporting | No |
| Creating a document | **Yes** |
| Editing a document | **Yes** |
| Deleting a document | **Yes** |

Document-editing operations require an explicit confirmation step and are described in the [Document Editing](../user-guide/document-editing.md) guide.

---

## Next Steps

- [User Guide](../user-guide/) — full documentation for all features
- [Keyboard Shortcuts](../user-guide/keyboard-shortcuts.md) — navigate without the mouse
- [Connections](../user-guide/connections.md) — managing multiple servers
