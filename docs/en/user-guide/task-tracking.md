# Task Tracking

Meilisearch processes write operations asynchronously. After a document mutation is submitted, MeiliScope displays the task status so you can confirm the operation completed.

---

## How Tasks Work

When you create, update, or delete a document, Meilisearch responds with a **task UID** immediately. The actual indexing happens in the background. MeiliScope polls the `/tasks/{uid}` endpoint to track progress.

---

## Task States

| State | Meaning |
|---|---|
| `enqueued` | Task is waiting to be processed |
| `processing` | Meilisearch is executing the operation |
| `succeeded` | Operation completed successfully |
| `failed` | Operation failed; the error message is shown |
| `canceled` | Task was explicitly canceled (not done by MeiliScope) |

---

## Viewing Task Results

After confirming a mutation, the Task Tracker panel opens automatically. It polls until the task reaches a terminal state (`succeeded`, `failed`, or `canceled`).

A `succeeded` status means Meilisearch accepted and applied the change. Click **Re-fetch** in the JSON Inspector to confirm the document reflects the edit.

---

## Ambiguous Outcomes

If the network connection drops after MeiliScope submits a mutation but before it receives a task UID, the outcome is unknown. MeiliScope reports this explicitly rather than claiming success or failure.

In this case, check the Meilisearch task queue directly to determine whether the operation was applied.

---

## Limitations

- MeiliScope does not support canceling an in-progress task.
- The Task Tracker shows only tasks initiated in the current session.
- Task history is not persisted across application restarts.
