# Document Editing

MeiliScope supports creating, updating, replacing, and deleting documents with an explicit confirmation step before every write.

---

## Opening the Document Editor

- From the **JSON Inspector**: click **Edit**.
- From the **Document Browser**: select a row and use the action menu.

---

## Edit Modes

| Mode | Meilisearch operation | Description |
|---|---|---|
| **Partial Update** | `PUT /indexes/{uid}/documents` | Sends only the modified fields; unspecified fields are unchanged |
| **Replace** | `POST /indexes/{uid}/documents` | Replaces the entire document |
| **Create** | `POST /indexes/{uid}/documents` | Submits a new document |
| **Delete** | `DELETE /indexes/{uid}/documents/{id}` | Removes the document by primary key |

---

## Partial Update vs. Replace

**Partial update** (`PUT`): only the fields you edit are sent to Meilisearch. Fields not present in your edit are left untouched on the server. Use this for routine field edits.

**Replace** (`POST`): the entire document object you submit becomes the new document. Fields you omit are removed. Use this only when you intend to completely overwrite the document.

---

## Diff Review

Before submitting, MeiliScope shows a diff between the original document and your edited version. Review all changes carefully.

---

## Confirmation

After reviewing the diff, click **Confirm** to submit the write. The mutation target (index UID and document ID) is locked at this moment. Switching connections or indexes after clicking Confirm does not redirect the write.

---

## Task Tracking

Document mutations are asynchronous in Meilisearch. After submission, MeiliScope opens the **Task Tracker** to show the task status. See [Task Tracking](task-tracking.md).

---

## Limitations

- An accepted task is not necessarily completed. Wait for the task to reach a terminal state (`succeeded` or `failed`) before relying on the result.
- Meilisearch applies indexing operations in order. A partial update and a subsequent replacement on the same document may execute in a different order than expected under high load.
- Mutation results are presented as uncertain if the network connection drops after submission.

---

## Safety Rules

- Browse operations (document listing, search) never send write requests.
- Every mutation requires explicit confirmation.
- MeiliScope does not perform automated bulk writes on your behalf.
- Do not run document-editing operations against production indexes unless you understand the data model and the difference between partial update and replace semantics.
