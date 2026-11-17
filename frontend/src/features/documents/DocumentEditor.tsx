import { useState, useEffect, useRef, useCallback } from "react";
import { EditorView, basicSetup } from "codemirror";
import { EditorState } from "@codemirror/state";
import { json } from "@codemirror/lang-json";
import { keymap } from "@codemirror/view";
import { oneDark } from "@codemirror/theme-one-dark";
import { DocumentWriteService } from "../../services/documentwrite";
import { TaskService, type TaskResult } from "../../services/task";
import { DocumentService } from "../../services/document";

// ─── Types ────────────────────────────────────────────────────────────────────

export type EditorMode = "create" | "edit" | "delete";

export interface DocumentEditorProps {
  mode: EditorMode;
  indexUID: string;
  primaryKeyField: string | null;
  initialDocJson?: string;
  documentID?: string;
  onClose: () => void;
  onSuccess: (taskUid: number) => void;
  onRefreshNeeded: () => void;
}

// ─── Diff logic ───────────────────────────────────────────────────────────────

type DiffType = "added" | "removed" | "changed" | "unchanged";

interface DiffEntry {
  path: string;
  type: DiffType;
  oldValue?: unknown;
  newValue?: unknown;
}

function diffValues(
  original: unknown,
  edited: unknown,
  path: string,
  out: DiffEntry[],
): void {
  if (out.length >= 50) return;

  const bothObjects =
    original !== null &&
    edited !== null &&
    typeof original === "object" &&
    typeof edited === "object" &&
    !Array.isArray(original) &&
    !Array.isArray(edited);

  if (bothObjects) {
    const orig = original as Record<string, unknown>;
    const edit = edited as Record<string, unknown>;
    const allKeys = new Set([...Object.keys(orig), ...Object.keys(edit)]);
    for (const key of allKeys) {
      if (out.length >= 50) return;
      const childPath = path ? `${path}.${key}` : key;
      if (!(key in orig)) {
        out.push({ path: childPath, type: "added", newValue: edit[key] });
      } else if (!(key in edit)) {
        out.push({ path: childPath, type: "removed", oldValue: orig[key] });
      } else {
        diffValues(orig[key], edit[key], childPath, out);
      }
    }
    return;
  }

  const bothArrays = Array.isArray(original) && Array.isArray(edited);
  if (bothArrays) {
    const orig = original as unknown[];
    const edit = edited as unknown[];
    const len = Math.max(orig.length, edit.length);
    for (let i = 0; i < len; i++) {
      if (out.length >= 50) return;
      const childPath = `${path}[${i}]`;
      if (i >= orig.length) {
        out.push({ path: childPath, type: "added", newValue: edit[i] });
      } else if (i >= edit.length) {
        out.push({ path: childPath, type: "removed", oldValue: orig[i] });
      } else {
        diffValues(orig[i], edit[i], childPath, out);
      }
    }
    return;
  }

  if (original === edited) {
    out.push({ path, type: "unchanged", oldValue: original, newValue: edited });
  } else {
    out.push({ path, type: "changed", oldValue: original, newValue: edited });
  }
}

function computeDiff(original: string, edited: string): DiffEntry[] {
  const origObj = JSON.parse(original) as unknown;
  const editObj = JSON.parse(edited) as unknown;
  const out: DiffEntry[] = [];
  diffValues(origObj, editObj, "", out);
  return out;
}

function formatDiffValue(val: unknown): string {
  if (val === null) return "null";
  if (val === undefined) return "";
  if (typeof val === "string") return `"${val}"`;
  if (typeof val === "object") return JSON.stringify(val);
  return String(val);
}

// ─── DiffTable ────────────────────────────────────────────────────────────────

function DiffTable({ original, edited }: { original: string; edited: string }) {
  let entries: DiffEntry[];
  let truncated = false;
  try {
    entries = computeDiff(original, edited);
    truncated = entries.length >= 50;
    entries = entries.filter((e) => e.type !== "unchanged");
  } catch {
    return <p className="editor-warning">Could not compute diff: invalid JSON.</p>;
  }

  if (entries.length === 0) {
    return <p style={{ color: "var(--color-text-muted, #888)", margin: "0.5rem 0" }}>No changes detected.</p>;
  }

  return (
    <div>
      <table className="editor-diff-table">
        <thead>
          <tr>
            <th>Field</th>
            <th>Change</th>
          </tr>
        </thead>
        <tbody>
          {entries.map((entry, idx) => (
            <tr key={idx} className={`diff-row diff-row--${entry.type}`}>
              <td className="diff-path">{entry.path}</td>
              <td className="diff-change">
                {entry.type === "added" && (
                  <span className="diff-added">+ {formatDiffValue(entry.newValue)}</span>
                )}
                {entry.type === "removed" && (
                  <span className="diff-removed">- {formatDiffValue(entry.oldValue)}</span>
                )}
                {entry.type === "changed" && (
                  <>
                    <span className="diff-removed">{formatDiffValue(entry.oldValue)}</span>
                    <span className="diff-arrow"> → </span>
                    <span className="diff-added">{formatDiffValue(entry.newValue)}</span>
                  </>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      {truncated && (
        <p style={{ color: "var(--color-text-muted, #888)", fontSize: "0.85em", marginTop: "0.4rem" }}>
          Showing first 50 changed entries only.
        </p>
      )}
    </div>
  );
}

// ─── JsonEditor (CodeMirror 6) ────────────────────────────────────────────────

interface JsonEditorProps {
  value: string;
  onChange: (v: string) => void;
  readOnly?: boolean;
}

function JsonEditor({ value, onChange, readOnly }: JsonEditorProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const viewRef = useRef<EditorView | null>(null);
  const onChangeRef = useRef(onChange);
  const externalValueRef = useRef(value);

  onChangeRef.current = onChange;

  useEffect(() => {
    if (!containerRef.current) return;
    const extensions = [
      basicSetup,
      json(),
      oneDark,
      keymap.of([]),
      EditorView.updateListener.of((update) => {
        if (update.docChanged) {
          const text = update.state.doc.toString();
          externalValueRef.current = text;
          onChangeRef.current(text);
        }
      }),
      EditorView.theme({
        "&": { height: "100%", fontSize: "13px" },
        ".cm-scroller": {
          overflow: "auto",
          fontFamily: "ui-monospace, 'Cascadia Code', Consolas, monospace",
        },
      }),
    ];

    if (readOnly) {
      extensions.push(EditorState.readOnly.of(true));
    }

    const view = new EditorView({
      state: EditorState.create({
        doc: externalValueRef.current,
        extensions,
      }),
      parent: containerRef.current,
    });
    viewRef.current = view;
    return () => {
      view.destroy();
      viewRef.current = null;
    };
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    const view = viewRef.current;
    if (!view) return;
    const current = view.state.doc.toString();
    if (current === value) return;
    externalValueRef.current = value;
    view.dispatch({
      changes: { from: 0, to: current.length, insert: value },
    });
  }, [value]);

  return <div ref={containerRef} className="editor-cm-wrapper" style={{ height: "100%", overflow: "hidden" }} />;
}

// ─── StatusBadge ─────────────────────────────────────────────────────────────

function StatusBadge({ status }: { status: string }) {
  const normalized = status.toLowerCase();
  return (
    <span className={`status-badge status-badge--${normalized}`}>
      {status}
    </span>
  );
}

// ─── TaskPanel ────────────────────────────────────────────────────────────────

type MonitoringState =
  | { kind: "polling" }
  | { kind: "succeeded" }
  | { kind: "failed"; error: TaskResult["error"] }
  | { kind: "canceled" }
  | { kind: "timeout"; taskUid: number }
  | { kind: "forbidden"; taskUid: number }
  | { kind: "stopped" };

interface TaskPanelProps {
  taskRef: { taskUid: number; indexUid: string; status: string; type: string; enqueuedAt: string };
  onRefreshNeeded: () => void;
  onClose: () => void;
}

function TaskPanel({ taskRef, onRefreshNeeded, onClose }: TaskPanelProps) {
  const [task, setTask] = useState<TaskResult | null>(null);
  const [monState, setMonState] = useState<MonitoringState>({ kind: "polling" });
  const pollCountRef = useRef(0);
  const activeRef = useRef(true);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const MAX_POLLS = 20;
  const INITIAL_DELAY_MS = 1000;
  const MAX_DELAY_MS = 5000;

  const fetchTask = useCallback(async (): Promise<boolean> => {
    try {
      const result = await TaskService.getTask(taskRef.taskUid);
      if (!activeRef.current) return false;
      setTask(result);
      const terminal = ["succeeded", "failed", "canceled"].includes(result.status.toLowerCase());
      if (terminal) {
        const s = result.status.toLowerCase();
        if (s === "succeeded") {
          setMonState({ kind: "succeeded" });
          onRefreshNeeded();
        } else if (s === "failed") {
          setMonState({ kind: "failed", error: result.error });
        } else {
          setMonState({ kind: "canceled" });
        }
        return true;
      }
      return false;
    } catch (e: unknown) {
      if (!activeRef.current) return false;
      const msg = e instanceof Error ? e.message : String(e);
      if (msg.includes("403") || msg.toLowerCase().includes("forbidden") || msg.toLowerCase().includes("permission")) {
        setMonState({ kind: "forbidden", taskUid: taskRef.taskUid });
        return true;
      }
      return false;
    }
  }, [taskRef.taskUid, onRefreshNeeded]);

  const scheduleNextPoll = useCallback((delay: number) => {
    if (!activeRef.current) return;
    if (pollCountRef.current >= MAX_POLLS) {
      setMonState({ kind: "timeout", taskUid: taskRef.taskUid });
      return;
    }
    timeoutRef.current = setTimeout(() => {
      if (!activeRef.current) return;
      pollCountRef.current++;
      void fetchTask().then((done) => {
        if (!done && activeRef.current) {
          const nextDelay = Math.min(delay * 2, MAX_DELAY_MS);
          scheduleNextPoll(nextDelay);
        }
      });
    }, delay);
  }, [fetchTask, taskRef.taskUid]);

  useEffect(() => {
    activeRef.current = true;
    pollCountRef.current = 0;

    void fetchTask().then((done) => {
      if (!done && activeRef.current) {
        scheduleNextPoll(INITIAL_DELAY_MS);
      }
    });

    return () => {
      activeRef.current = false;
      if (timeoutRef.current !== null) {
        clearTimeout(timeoutRef.current);
        timeoutRef.current = null;
      }
    };
  }, [fetchTask, scheduleNextPoll]);

  const stopMonitoring = () => {
    activeRef.current = false;
    if (timeoutRef.current !== null) {
      clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
    }
    setMonState({ kind: "stopped" });
  };

  const refreshManually = () => {
    if (timeoutRef.current !== null) {
      clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
    }
    void fetchTask().then((done) => {
      if (!done && activeRef.current) {
        scheduleNextPoll(INITIAL_DELAY_MS);
      }
    });
  };

  const currentStatus = task?.status ?? taskRef.status;
  const isPolling = monState.kind === "polling";

  return (
    <div className="editor-task-panel">
      <div className="editor-task-panel__header">
        <span className="editor-task-panel__label">Task</span>
        <span className="editor-task-panel__uid">#{taskRef.taskUid}</span>
        <StatusBadge status={currentStatus} />
        {isPolling && (
          <span className="editor-task-panel__polling-indicator">Monitoring…</span>
        )}
      </div>

      <dl className="editor-task-panel__meta">
        <dt>Index</dt><dd>{taskRef.indexUid}</dd>
        <dt>Type</dt><dd>{taskRef.type}</dd>
        <dt>Enqueued</dt><dd>{taskRef.enqueuedAt}</dd>
        {task?.startedAt && <><dt>Started</dt><dd>{task.startedAt}</dd></>}
        {task?.finishedAt && <><dt>Finished</dt><dd>{task.finishedAt}</dd></>}
        {task?.duration && <><dt>Duration</dt><dd>{task.duration}</dd></>}
      </dl>

      {monState.kind === "succeeded" && (
        <div className="editor-task-panel__message editor-task-panel__message--success">
          Done — documents refreshed.
        </div>
      )}

      {monState.kind === "failed" && (
        <div className="editor-task-panel__message editor-task-panel__message--error">
          <strong>Task failed.</strong>
          {monState.error && (
            <dl className="editor-task-panel__error-dl">
              {monState.error.message && <><dt>Message</dt><dd>{monState.error.message}</dd></>}
              {monState.error.code && <><dt>Code</dt><dd>{monState.error.code}</dd></>}
              {monState.error.type && <><dt>Type</dt><dd>{monState.error.type}</dd></>}
              {monState.error.link && <><dt>Docs</dt><dd>{monState.error.link}</dd></>}
            </dl>
          )}
        </div>
      )}

      {monState.kind === "canceled" && (
        <div className="editor-task-panel__message editor-task-panel__message--warning">
          Task was canceled.
        </div>
      )}

      {monState.kind === "forbidden" && (
        <div className="editor-task-panel__message editor-task-panel__message--warning">
          Cannot retrieve task status (permission denied). Task UID: {monState.taskUid} was accepted.
        </div>
      )}

      {monState.kind === "timeout" && (
        <div className="editor-task-panel__message editor-task-panel__message--warning">
          Monitoring stopped after {MAX_POLLS} polls. Task UID: {monState.taskUid}. Use Refresh Status to check manually.
        </div>
      )}

      {monState.kind === "stopped" && (
        <div className="editor-task-panel__message editor-task-panel__message--warning">
          Monitoring stopped by user.
        </div>
      )}

      <div className="editor-task-panel__actions">
        {isPolling && (
          <button className="btn btn--ghost btn--sm" onClick={stopMonitoring}>
            Stop Monitoring
          </button>
        )}
        {!isPolling && monState.kind !== "succeeded" && (
          <button className="btn btn--ghost btn--sm" onClick={refreshManually}>
            Refresh Status
          </button>
        )}
        <button className="btn btn--primary btn--sm" onClick={onClose}>
          Close
        </button>
      </div>
    </div>
  );
}

// ─── DocumentEditor ───────────────────────────────────────────────────────────

type Step = "edit" | "confirm" | "task";

export function DocumentEditor({
  mode,
  indexUID,
  primaryKeyField,
  initialDocJson,
  documentID,
  onClose,
  onSuccess,
  onRefreshNeeded,
}: DocumentEditorProps) {
  const DEFAULT_EMPTY = "{}";
  const originalJson = initialDocJson ?? DEFAULT_EMPTY;

  const [editorValue, setEditorValue] = useState<string>(
    mode === "create" ? DEFAULT_EMPTY : originalJson,
  );
  const [step, setStep] = useState<Step>("edit");
  const [validationError, setValidationError] = useState<string | null>(null);
  const [pkWarning, setPkWarning] = useState<string | null>(null);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [taskRef, setTaskRef] = useState<{
    taskUid: number;
    indexUid: string;
    status: string;
    type: string;
    enqueuedAt: string;
  } | null>(null);
  const [concurrentWarning, setConcurrentWarning] = useState<string | null>(null);
  const [serverCurrentJson, setServerCurrentJson] = useState<string | null>(null);
  const [checkingConcurrent, setCheckingConcurrent] = useState(false);

  const hasUnsavedChanges = (): boolean => {
    if (mode === "delete") return false;
    const baseline = mode === "create" ? DEFAULT_EMPTY : originalJson;
    return editorValue.trim() !== baseline.trim();
  };

  const handleClose = () => {
    if (step === "task") {
      onClose();
      return;
    }
    if (hasUnsavedChanges()) {
      if (!window.confirm("You have unsaved changes. Discard and close?")) return;
    }
    onClose();
  };

  const validateJson = (text: string): { ok: boolean; parsed?: unknown } => {
    try {
      const parsed = JSON.parse(text);
      if (typeof parsed !== "object" || Array.isArray(parsed) || parsed === null) {
        setValidationError("Document must be a JSON object, not an array or scalar.");
        return { ok: false };
      }
      setValidationError(null);
      return { ok: true, parsed };
    } catch (e: unknown) {
      setValidationError(e instanceof SyntaxError ? e.message : "Invalid JSON.");
      return { ok: false };
    }
  };

  const handleProceedToConfirm = async () => {
    if (mode === "delete") {
      setStep("confirm");
      return;
    }

    const { ok, parsed } = validateJson(editorValue);
    if (!ok) return;

    const doc = parsed as Record<string, unknown>;

    // Primary key warning
    if (primaryKeyField && !(primaryKeyField in doc)) {
      setPkWarning(
        `Field "${primaryKeyField}" (primary key) is missing. The document may be rejected or given an auto-generated ID.`,
      );
    } else {
      setPkWarning(null);
    }

    // Concurrent modification check (edit mode only)
    if (mode === "edit" && documentID) {
      setCheckingConcurrent(true);
      setConcurrentWarning(null);
      setServerCurrentJson(null);
      try {
        const fetched = await DocumentService.getDocument(indexUID, documentID);
        const fetchedJson = fetched.rawJson ?? JSON.stringify(fetched.fields ?? {}, null, 2);
        const normalizedFetched = (() => {
          try { return JSON.stringify(JSON.parse(fetchedJson)); } catch { return fetchedJson; }
        })();
        const normalizedOriginal = (() => {
          try { return JSON.stringify(JSON.parse(originalJson)); } catch { return originalJson; }
        })();
        if (normalizedFetched !== normalizedOriginal) {
          setConcurrentWarning(
            "The document was modified since you opened it. Original (server), Your edits, and Current (server) versions are shown. Proceed carefully.",
          );
          setServerCurrentJson(fetchedJson);
        }
      } catch {
        // Non-blocking — ignore fetch errors for the concurrent check
      } finally {
        setCheckingConcurrent(false);
      }
    }

    setStep("confirm");
  };

  const getPkValueFromEditor = (): string | null => {
    if (!primaryKeyField) return null;
    try {
      const doc = JSON.parse(editorValue) as Record<string, unknown>;
      const val = doc[primaryKeyField];
      return val !== undefined ? String(val) : null;
    } catch {
      return null;
    }
  };

  const buildSubmitPayload = (): string => {
    if (!primaryKeyField || mode === "create") return editorValue;
    // Re-inject original PK value to prevent accidental PK change in edit mode
    try {
      const originalParsed = JSON.parse(originalJson) as Record<string, unknown>;
      const originalPkValue = originalParsed[primaryKeyField];
      const editedParsed = JSON.parse(editorValue) as Record<string, unknown>;
      editedParsed[primaryKeyField] = originalPkValue;
      return JSON.stringify(editedParsed, null, 2);
    } catch {
      return editorValue;
    }
  };

  const handleConfirm = async () => {
    setSubmitting(true);
    setSubmitError(null);
    try {
      let ref;
      if (mode === "create") {
        ref = await DocumentWriteService.addDocument(indexUID, editorValue, primaryKeyField ?? "");
      } else if (mode === "edit") {
        const payload = buildSubmitPayload();
        ref = await DocumentWriteService.updateDocument(indexUID, payload, primaryKeyField ?? "");
      } else {
        ref = await DocumentWriteService.deleteDocument(indexUID, documentID!);
      }
      setTaskRef({
        taskUid: ref.taskUid,
        indexUid: ref.indexUid,
        status: ref.status,
        type: ref.type,
        enqueuedAt: ref.enqueuedAt,
      });
      onSuccess(ref.taskUid);
      setStep("task");
    } catch (e: unknown) {
      setSubmitError(e instanceof Error ? e.message : String(e));
    } finally {
      setSubmitting(false);
    }
  };

  const modeLabel =
    mode === "create" ? "Add Document" : mode === "edit" ? "Edit Document" : "Delete Document";

  const operationLabel =
    mode === "create"
      ? "Add Document (POST)"
      : mode === "edit"
      ? "Update Document (PUT)"
      : "Delete Document";

  const payloadPreview = editorValue.slice(0, 200) + (editorValue.length > 200 ? "…" : "");

  return (
    <div className="editor-modal-overlay" role="dialog" aria-modal="true" aria-label={modeLabel}>
      <div className="editor-modal">
        {/* Header */}
        <div className="editor-modal-header">
          <span className="editor-modal-header__title">{modeLabel}</span>
          <span className="editor-modal-header__index">{indexUID}</span>
          <button
            className="btn btn--ghost btn--sm editor-modal-header__close"
            onClick={handleClose}
            aria-label="Close"
          >
            ✕
          </button>
        </div>

        {/* Body */}
        <div className="editor-modal-body">
          {/* ── EDIT STEP ───────────────────────────────────────────────── */}
          {step === "edit" && (
            <>
              {/* Delete mode: read-only view */}
              {mode === "delete" && (
                <>
                  <div className="editor-warning">
                    This action is permanent and cannot be undone.
                  </div>
                  {documentID && (
                    <div className="editor-delete-id">
                      <strong>Document ID:</strong>{" "}
                      <code>{documentID}</code>
                    </div>
                  )}
                  <div className="editor-cm-wrapper" style={{ height: "320px" }}>
                    <JsonEditor
                      value={originalJson}
                      onChange={() => {}}
                      readOnly
                    />
                  </div>
                </>
              )}

              {/* Create / Edit mode: editable */}
              {mode !== "delete" && (
                <>
                  {mode === "edit" && primaryKeyField && (
                    <p className="editor-pk-hint">
                      Primary key field: <code>{primaryKeyField}</code>. Its value cannot be changed.
                    </p>
                  )}
                  <div className="editor-cm-wrapper" style={{ height: "360px" }}>
                    <JsonEditor
                      value={editorValue}
                      onChange={(v) => { setEditorValue(v); setValidationError(null); }}
                    />
                  </div>
                  {validationError && (
                    <div className="editor-warning">{validationError}</div>
                  )}
                </>
              )}

              <div className="editor-step-actions">
                <button className="btn btn--ghost" onClick={handleClose} disabled={checkingConcurrent}>
                  Cancel
                </button>
                {mode === "delete" ? (
                  <button
                    className="btn btn--danger"
                    onClick={() => void handleProceedToConfirm()}
                  >
                    Review Delete
                  </button>
                ) : (
                  <button
                    className="btn btn--primary"
                    onClick={() => void handleProceedToConfirm()}
                    disabled={checkingConcurrent}
                  >
                    {checkingConcurrent ? "Checking…" : "Review Changes"}
                  </button>
                )}
              </div>
            </>
          )}

          {/* ── CONFIRM STEP ────────────────────────────────────────────── */}
          {step === "confirm" && (
            <div className="editor-confirm">
              <dl className="editor-confirm__meta">
                <dt>Index</dt><dd>{indexUID}</dd>
                <dt>Operation</dt><dd>{operationLabel}</dd>
                {(mode === "edit" || mode === "delete") && documentID && (
                  <><dt>Document ID</dt><dd><code>{documentID}</code></dd></>
                )}
                {mode === "edit" && !documentID && getPkValueFromEditor() && (
                  <><dt>Document ID</dt><dd><code>{getPkValueFromEditor()}</code></dd></>
                )}
              </dl>

              {/* PK warning */}
              {pkWarning && (
                <div className="editor-warning">{pkWarning}</div>
              )}

              {/* Concurrent modification warning */}
              {concurrentWarning && (
                <div className="editor-warning">{concurrentWarning}</div>
              )}

              {/* Delete: show document JSON preview */}
              {mode === "delete" && (
                <>
                  <div className="editor-warning">
                    This action is permanent and cannot be undone.
                  </div>
                  <div className="editor-confirm__section-label">Document to delete:</div>
                  <div className="editor-cm-wrapper" style={{ height: "200px" }}>
                    <JsonEditor value={originalJson} onChange={() => {}} readOnly />
                  </div>
                </>
              )}

              {/* Create: payload preview */}
              {mode === "create" && (
                <>
                  <div className="editor-confirm__section-label">Payload preview:</div>
                  <pre className="editor-confirm__preview">{payloadPreview}</pre>
                </>
              )}

              {/* Edit: diff */}
              {mode === "edit" && (
                <>
                  <div className="editor-confirm__section-label">Changes:</div>
                  <DiffTable original={originalJson} edited={editorValue} />

                  {/* Concurrent: show server current version too */}
                  {concurrentWarning && serverCurrentJson && (
                    <>
                      <div className="editor-confirm__section-label" style={{ marginTop: "1rem" }}>
                        Current version on server:
                      </div>
                      <div className="editor-cm-wrapper" style={{ height: "160px" }}>
                        <JsonEditor value={serverCurrentJson} onChange={() => {}} readOnly />
                      </div>
                    </>
                  )}
                </>
              )}

              {submitError && (
                <div className="editor-warning">{submitError}</div>
              )}

              <div className="editor-step-actions">
                <button className="btn btn--ghost" onClick={() => { setSubmitError(null); setStep("edit"); }} disabled={submitting}>
                  Back
                </button>
                <button className="btn btn--ghost" onClick={handleClose} disabled={submitting}>
                  Cancel
                </button>
                {mode === "delete" ? (
                  <button
                    className="btn btn--danger"
                    onClick={() => void handleConfirm()}
                    disabled={submitting}
                  >
                    {submitting ? "Deleting…" : "Confirm Delete"}
                  </button>
                ) : (
                  <button
                    className="btn btn--primary"
                    onClick={() => void handleConfirm()}
                    disabled={submitting}
                  >
                    {submitting ? "Submitting…" : "Confirm Submit"}
                  </button>
                )}
              </div>
            </div>
          )}

          {/* ── TASK STEP ───────────────────────────────────────────────── */}
          {step === "task" && taskRef && (
            <TaskPanel
              taskRef={taskRef}
              onRefreshNeeded={onRefreshNeeded}
              onClose={onClose}
            />
          )}
        </div>
      </div>
    </div>
  );
}
