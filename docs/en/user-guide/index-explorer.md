# Index Explorer

The Index Explorer (Ctrl+1) lists all indexes available on the active connection and shows statistics for the selected index.

---

## Browsing Indexes

The left panel lists indexes returned by the Meilisearch `/indexes` endpoint. Click an index name to select it.

If index enumeration is blocked by a 403 error (search-only key), the list is replaced by a text field. Enter the index UID directly and press Enter.

---

## Index Statistics

After selecting an index, the right panel shows:

- **Document count** — total number of documents in the index.
- **Index size** — raw storage size reported by Meilisearch.
- **Primary key** — the field used as the document identifier.
- **Field distribution** — number of documents that contain each field, sorted by frequency.
- **Indexing status** — whether the index is currently processing a task.

Statistics are fetched from the `/indexes/{uid}/stats` endpoint.

---

## Refreshing

Click the **Refresh** button (or reselect the index) to reload statistics.

---

## Navigating to Documents

After selecting an index, switch to the **Document Browser** tab (Ctrl+2) or the **Query Console** tab (Ctrl+3). The selected index is used automatically.
