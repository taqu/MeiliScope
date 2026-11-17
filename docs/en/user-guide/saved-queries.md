# Saved Queries

Saved Queries let you persist named search queries for quick reuse.

---

## Opening the Saved Queries Panel

In the **Query Console**, click **★ Saved** in the toolbar, or press **Ctrl+S**.

---

## Saving a Query

1. Enter a name in the **Save current query** field.
2. Click **Save**.

The current query JSON and selected index UID are saved together.

---

## Loading a Saved Query

Click any saved query entry. The query JSON is loaded into the editor and the index UID is restored (if the index exists in the current connection).

---

## Updating a Saved Query

Select a saved query and click **Update** to overwrite it with the current editor content.

---

## Deleting a Saved Query

Select a saved query and click **Delete**.

---

## Storage

Saved queries are stored locally on disk in the application data directory. They are not uploaded to any external service.

API keys are never stored in query records.

Saved queries are associated with the connection name. If you rename or delete a connection, the queries remain on disk but are no longer automatically associated.

---

## Query History

Every executed query is automatically recorded in the query history (separate from saved queries). The history appears in the **History** tab in the Query Console results area.
