# Export

MeiliScope can export documents to JSON or CSV files.

---

## Starting an Export

Export is available in two places:

- **Document Browser** — click **Export** to export the current document page.
- **Query Console** — click **Export** in the results tabs bar to export the current search results.

---

## Export Formats

### JSON

Exports the documents as a JSON array. Each element is the full document object.

```json
[
  { "id": 1, "title": "First document" },
  { "id": 2, "title": "Second document" }
]
```

### CSV

Exports the documents as comma-separated values.

- The first row is a header containing the column names.
- One row per document.
- Nested objects and arrays are serialized as JSON strings within the cell.

**Formula injection protection:** cells whose first character is `=`, `+`, `-`, or `@` are prefixed with a tab character. This prevents spreadsheet applications (Excel, LibreOffice Calc) from treating cell values as formulas.

---

## Choosing a Save Location

Clicking **Export** opens the native OS file dialog to select the save location and filename.

---

## Export Scope

Export saves only the documents currently loaded — the current page (Document Browser) or the current result set (Query Console).

> **Important:** Exporting does not export the entire index. To export more documents, increase the `limit` in your query or use the Meilisearch CLI or API for bulk exports.

---

## Limitations

- Large exports may be slow depending on the number of documents and their size.
- Very deeply nested documents may produce verbose CSV output.
- Null values are exported as empty cells in CSV.
- The export uses the documents as returned by Meilisearch, which may apply attribute filtering if `attributesToRetrieve` is set.
