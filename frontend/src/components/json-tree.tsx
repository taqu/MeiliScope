import { useState } from "react";

// ─── JSON parser (precision-preserving) ───────────────────────────────────────

export type JNode =
  | { kind: "null" }
  | { kind: "bool"; value: boolean }
  | { kind: "number"; raw: string }
  | { kind: "string"; value: string }
  | { kind: "array"; items: JNode[] }
  | { kind: "object"; entries: [string, JNode][] };

interface PS { t: string; p: number; }

function ws(s: PS) {
  while (s.p < s.t.length && " \t\r\n".includes(s.t[s.p])) s.p++;
}

function parseStr(s: PS): string {
  s.p++;
  let r = "";
  while (s.p < s.t.length && s.t[s.p] !== '"') {
    if (s.t[s.p] === "\\") {
      s.p++;
      const e = s.t[s.p++];
      if (e === "u") {
        r += String.fromCharCode(parseInt(s.t.slice(s.p, s.p + 4), 16));
        s.p += 4;
      } else {
        r += e === "n" ? "\n" : e === "r" ? "\r" : e === "t" ? "\t" :
             e === "b" ? "\b" : e === "f" ? "\f" : e;
      }
    } else {
      r += s.t[s.p++];
    }
  }
  s.p++;
  return r;
}

function parseVal(s: PS): JNode {
  ws(s);
  const c = s.t[s.p];
  if (c === "n") { s.p += 4; return { kind: "null" }; }
  if (c === "t") { s.p += 4; return { kind: "bool", value: true }; }
  if (c === "f") { s.p += 5; return { kind: "bool", value: false }; }
  if (c === '"') return { kind: "string", value: parseStr(s) };
  if (c === "[") {
    s.p++;
    const items: JNode[] = [];
    ws(s);
    if (s.t[s.p] === "]") { s.p++; return { kind: "array", items }; }
    while (s.p < s.t.length) {
      items.push(parseVal(s));
      ws(s);
      if (s.t[s.p] === "]") { s.p++; return { kind: "array", items }; }
      s.p++;
    }
    return { kind: "array", items };
  }
  if (c === "{") {
    s.p++;
    const entries: [string, JNode][] = [];
    ws(s);
    if (s.t[s.p] === "}") { s.p++; return { kind: "object", entries }; }
    while (s.p < s.t.length) {
      ws(s);
      const key = parseStr(s);
      ws(s);
      s.p++;
      const val = parseVal(s);
      entries.push([key, val]);
      ws(s);
      if (s.t[s.p] === "}") { s.p++; return { kind: "object", entries }; }
      s.p++;
    }
    return { kind: "object", entries };
  }
  const st = s.p;
  if (s.t[s.p] === "-") s.p++;
  while (s.p < s.t.length && s.t[s.p] >= "0" && s.t[s.p] <= "9") s.p++;
  if (s.p < s.t.length && s.t[s.p] === ".") {
    s.p++;
    while (s.p < s.t.length && s.t[s.p] >= "0" && s.t[s.p] <= "9") s.p++;
  }
  if (s.p < s.t.length && (s.t[s.p] === "e" || s.t[s.p] === "E")) {
    s.p++;
    if (s.t[s.p] === "+" || s.t[s.p] === "-") s.p++;
    while (s.p < s.t.length && s.t[s.p] >= "0" && s.t[s.p] <= "9") s.p++;
  }
  return { kind: "number", raw: s.t.slice(st, s.p) };
}

export function parseJsonDoc(text: string): JNode | null {
  try {
    const s: PS = { t: text, p: 0 };
    ws(s);
    return parseVal(s);
  } catch {
    return null;
  }
}

export function escapePointerToken(key: string): string {
  return key.replace(/~/g, "~0").replace(/\//g, "~1");
}

export function nodeToJson(node: JNode): string {
  if (node.kind === "null") return "null";
  if (node.kind === "bool") return node.value ? "true" : "false";
  if (node.kind === "number") return node.raw;
  if (node.kind === "string") return JSON.stringify(node.value);
  if (node.kind === "array") return "[" + node.items.map(nodeToJson).join(",") + "]";
  return "{" + node.entries.map(([k, v]) => JSON.stringify(k) + ":" + nodeToJson(v)).join(",") + "}";
}

function nodePreview(node: JNode): string {
  if (node.kind === "null") return "null";
  if (node.kind === "bool") return node.value ? "true" : "false";
  if (node.kind === "number") return node.raw;
  if (node.kind === "string") {
    const s = node.value.length > 40 ? node.value.slice(0, 40) + "…" : node.value;
    return `"${s}"`;
  }
  if (node.kind === "array") return `[ ${node.items.length} items ]`;
  return `{ ${node.entries.length} keys }`;
}

async function copyText(text: string) {
  try { await navigator.clipboard.writeText(text); } catch { /* ignore */ }
}

// ─── JSON node view ───────────────────────────────────────────────────────────

interface JsonNodeProps {
  node: JNode;
  depth: number;
  path: string;
  label?: string;
  defaultExpanded?: boolean;
}

export function JsonNodeView({ node, depth, path, label, defaultExpanded = false }: JsonNodeProps) {
  const [expanded, setExpanded] = useState(defaultExpanded || depth < 2);
  const [copied, setCopied] = useState(false);

  const handleCopy = async (text: string, e: React.MouseEvent) => {
    e.stopPropagation();
    await copyText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 1200);
  };

  const isContainer = node.kind === "array" || node.kind === "object";
  const isEmpty =
    (node.kind === "array" && node.items.length === 0) ||
    (node.kind === "object" && node.entries.length === 0);

  const keyEl = label !== undefined ? (
    <span className="json-key" title={path}>{label}: </span>
  ) : null;

  if (!isContainer || isEmpty) {
    return (
      <div className="json-row" style={{ paddingLeft: depth * 14 + 4 }}>
        {keyEl}
        <span className={`json-${node.kind}`}>{nodePreview(node)}</span>
        <button className="json-copy-btn" onClick={(e) => void handleCopy(nodeToJson(node), e)}>
          {copied ? "✓" : "⎘"}
        </button>
      </div>
    );
  }

  if (!expanded) {
    return (
      <div className="json-row" style={{ paddingLeft: depth * 14 + 4 }}>
        {keyEl}
        <button className="json-toggle" onClick={() => setExpanded(true)}>▶</button>
        <span className="json-preview">{nodePreview(node)}</span>
        <button className="json-copy-btn" onClick={(e) => void handleCopy(nodeToJson(node), e)}>
          {copied ? "✓" : "⎘"}
        </button>
      </div>
    );
  }

  if (node.kind === "array") {
    return (
      <div className="json-container">
        <div className="json-row" style={{ paddingLeft: depth * 14 + 4 }}>
          {keyEl}
          <button className="json-toggle" onClick={() => setExpanded(false)}>▼</button>
          <span className="json-bracket">[</span>
          <button className="json-copy-btn" onClick={(e) => void handleCopy(nodeToJson(node), e)}>
            {copied ? "✓" : "⎘"}
          </button>
        </div>
        <div>
          {node.items.map((item, i) => (
            <JsonNodeView key={i} node={item} depth={depth + 1}
              path={`${path}[${i}]`} label={String(i)} />
          ))}
        </div>
        <div className="json-row" style={{ paddingLeft: depth * 14 + 4 }}>
          <span className="json-bracket">]</span>
        </div>
      </div>
    );
  }

  return (
    <div className="json-container">
      <div className="json-row" style={{ paddingLeft: depth * 14 + 4 }}>
        {keyEl}
        <button className="json-toggle" onClick={() => setExpanded(false)}>▼</button>
        <span className="json-bracket">{"{"}</span>
        <button className="json-copy-btn" onClick={(e) => void handleCopy(nodeToJson(node), e)}>
          {copied ? "✓" : "⎘"}
        </button>
      </div>
      <div>
        {node.entries.map(([k, v]) => (
          <JsonNodeView key={k} node={v} depth={depth + 1}
            path={`${path}/${escapePointerToken(k)}`} label={k} />
        ))}
      </div>
      <div className="json-row" style={{ paddingLeft: depth * 14 + 4 }}>
        <span className="json-bracket">{"}"}</span>
      </div>
    </div>
  );
}

// ─── Inspector panel ──────────────────────────────────────────────────────────

interface InspectorPanelProps {
  rawJson: string | null;
  title?: string;
  emptyMessage?: string;
}

export function InspectorPanel({ rawJson, title = "JSON Inspector", emptyMessage = "Select a row to inspect its JSON." }: InspectorPanelProps) {
  const [copiedAll, setCopiedAll] = useState(false);

  const handleCopyAll = async () => {
    if (!rawJson) return;
    try {
      await copyText(JSON.stringify(JSON.parse(rawJson), null, 2));
    } catch {
      await copyText(rawJson);
    }
    setCopiedAll(true);
    setTimeout(() => setCopiedAll(false), 1200);
  };

  if (!rawJson) {
    return (
      <div className="inspector-panel inspector-panel--empty">
        <p className="inspector-empty-msg">{emptyMessage}</p>
      </div>
    );
  }

  const root = parseJsonDoc(rawJson);

  return (
    <div className="inspector-panel">
      <div className="inspector-header">
        <span className="inspector-title">{title}</span>
        <button className="btn btn--ghost btn--sm" onClick={() => void handleCopyAll()}>
          {copiedAll ? "Copied ✓" : "Copy JSON"}
        </button>
      </div>
      <div className="inspector-tree">
        {root ? (
          <JsonNodeView node={root} depth={0} path="" defaultExpanded={true} />
        ) : (
          <p className="inspector-parse-error">Failed to parse JSON.</p>
        )}
      </div>
    </div>
  );
}
