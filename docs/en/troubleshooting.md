# Troubleshooting

---

## Connection Problems

**"Connection failed" or timeout**
- Verify the Meilisearch server is running and accessible from this machine.
- Check that the URL scheme, host, and port are correct.
- Ensure no firewall rule is blocking the connection.
- Check the log file for the full error:
  - Windows: `%AppData%\MeiliScope\logs\`
  - Linux: `~/.local/share/MeiliScope/logs/`

**"Unauthorized" or "Forbidden"**
- Confirm the API key is correct.
- If using a search-only key, some features (index listing, document editing) will be unavailable. This is expected; MeiliScope reports this as `limited_access` mode.

**Index list is empty / "Enter index UID manually"**
- The API key does not have permission to list indexes. Enter the index UID directly in the text field that appears.

---

## Application Startup Problems

**Windows: blank white window**
WebView2 initialisation issue. Update WebView2 via Windows Update.

**Windows: application does not launch**
Check `%AppData%\MeiliScope\logs\` for startup errors. Verify Windows version is 10 1803 or later.

**Linux: "cannot open shared object file"**
Install missing runtime libraries. See [Installation — Linux](getting-started/installation.md#linux).

**Linux: "cannot connect to display"**
MeiliScope requires a graphical desktop session. It cannot run headless.

---

## Search and Query Problems

**No results for a query that should return results**
- Confirm the query JSON is valid.
- Check that filters reference filterable attributes in the Meilisearch index settings.
- Confirm sort expressions reference sortable attributes.
- Remove custom parameters one at a time to isolate the cause.

**"Invalid filter expression" or Meilisearch error in results**
Meilisearch returned an error for the query. The error message from Meilisearch is displayed directly in the results area.

---

## Export Problems

**Export dialog does not appear**
- Ensure no other modal dialog is open.
- Try again after the results have fully loaded.

**CSV file contains unexpected values**
- Nested JSON objects and arrays are serialized as JSON strings within CSV cells — this is expected behavior.
- Cells prefixed with a tab are values that began with a formula-injection character (`=`, `+`, `-`, `@`). This tab is intentional safety behavior.

---

## Document Editing Problems

**Task stays in `enqueued` or `processing` for a long time**
Meilisearch is busy processing. The task tracker will update when the state changes. Check the Meilisearch server logs if it appears stuck.

**Task failed**
The error from Meilisearch is shown in the task tracker. Common causes: invalid document structure, primary key conflict, missing required fields.

---

## Credential Problems

**Credentials not saved across restarts (Linux)**
The Secret Service daemon (GNOME Keyring / KWallet) is not available. MeiliScope falls back to session-only mode. Start the keyring daemon or use a desktop session that provides one.

---

## Log Files

| Platform | Path |
|---|---|
| Windows | `%AppData%\MeiliScope\logs\` |
| Linux | `~/.local/share/MeiliScope/logs/` |

Logs are rotated automatically. The current log is `app.log`. Log files do not contain API keys or Authorization headers.
