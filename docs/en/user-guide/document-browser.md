# Document Browser

The Document Browser (Ctrl+2) lists the raw documents of the selected index with pagination.

---

## Browsing Documents

Documents are retrieved from the Meilisearch `/indexes/{uid}/documents` endpoint in pages of 20 by default.

Use the **Previous** and **Next** buttons to paginate. The page indicator shows the current position and total document count.

Columns are generated dynamically from the fields present in the returned documents.

---

## Inspecting a Document

Click any row to open the **JSON Inspector** for that document.

The JSON Inspector shows the full document as a collapsible tree. Expand nested objects and arrays by clicking the disclosure arrows.

---

## Exporting

Click **Export** to save the current page to a file. See the [Export](export.md) guide for format options and limitations.

> **Note:** Export saves the current page only, not the entire index. The page size is limited by the Meilisearch `/documents` endpoint default limit.

---

## Limitations

- Document browsing uses offset-based pagination. Very large offsets may be slower on some Meilisearch configurations.
- The column layout is generated from the first page of results. Columns may differ if documents have inconsistent field sets.
- Binary or very large field values are truncated in the table view but are shown in full in the JSON Inspector.
