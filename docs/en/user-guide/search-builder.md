# Search Builder

The Search Builder is a form-based editor in the Query Console that lets you configure common Meilisearch search parameters without writing JSON manually.

---

## Opening the Search Builder

In the **Query Console** toolbar, click **Builder**. The JSON editor is replaced with the form.

Clicking **JSON** switches back to the JSON editor. All changes made in the builder are preserved in the underlying JSON.

---

## Supported Fields

| Field | Description |
|---|---|
| **Query** | Search string (`q`) |
| **Filter** | Filter expression string |
| **Sort** | Comma-separated sort expressions (e.g. `price:asc, date:desc`) |
| **Facets** | Comma-separated facet attribute names |
| **Attributes to retrieve** | Comma-separated attribute names to include in results |
| **Limit** | Maximum number of results |
| **Offset** | Number of results to skip |
| **Page** | Page number (alternative to offset-based pagination) |
| **Hits per page** | Results per page (used with page) |
| **Show ranking score** | Include `_rankingScore` in results |
| **Show matches position** | Include `_matchesPosition` in results |

---

## Partial Editing

The Search Builder only modifies fields you explicitly set. Fields in the JSON that have no corresponding builder control are preserved unchanged. This allows you to mix the builder with custom JSON fields.

Clearing a field in the builder removes it from the JSON entirely (which uses the Meilisearch server default for that parameter).

---

## Validation

If the JSON in the editor is not valid when you open the builder, a warning is displayed. The builder can still be used; it will generate a valid JSON object from scratch.

---

## Limitations

- Complex nested filter expressions may require the JSON editor.
- Not all Meilisearch search parameters are exposed in the builder. Use the JSON editor for parameters not shown.
