import { useMemo } from "react";

// ─── Types ────────────────────────────────────────────────────────────────────

export interface SearchBuilderProps {
  queryText: string;
  onQueryChange: (newQuery: string) => void;
  indexUID: string;
  onIndexChange: (uid: string) => void;
  indexOptions: Array<{ uid: string }>;
  onExecute: () => void;
  executing: boolean;
}

interface SortRow {
  field: string;
  dir: "asc" | "desc";
}

interface ParsedFields {
  valid: true;
  q: string;
  filter: string | null;
  filterIsArray: boolean;
  sort: SortRow[];
  limit: number;
  offset: number;
  facets: string[];
  attributesToRetrieve: string[];
  hasPagePagination: boolean;
}

interface ParsedInvalid {
  valid: false;
}

type Parsed = ParsedFields | ParsedInvalid;

// ─── Helpers ──────────────────────────────────────────────────────────────────

function patchQuery(queryText: string, patches: Record<string, unknown>): string {
  try {
    const obj = JSON.parse(queryText) as Record<string, unknown>;
    for (const [key, value] of Object.entries(patches)) {
      if (
        value === undefined ||
        value === null ||
        value === "" ||
        (Array.isArray(value) && value.length === 0)
      ) {
        delete obj[key];
      } else {
        obj[key] = value;
      }
    }
    return JSON.stringify(obj, null, 2);
  } catch {
    return queryText;
  }
}

function parseQueryText(queryText: string): Parsed {
  let obj: Record<string, unknown>;
  try {
    const parsed = JSON.parse(queryText);
    if (typeof parsed !== "object" || Array.isArray(parsed) || parsed === null) {
      return { valid: false };
    }
    obj = parsed as Record<string, unknown>;
  } catch {
    return { valid: false };
  }

  const q = typeof obj["q"] === "string" ? obj["q"] : "";

  let filter: string | null = null;
  let filterIsArray = false;
  if (Array.isArray(obj["filter"])) {
    filterIsArray = true;
  } else if (typeof obj["filter"] === "string") {
    filter = obj["filter"];
  }

  const sort: SortRow[] = [];
  if (Array.isArray(obj["sort"])) {
    for (const item of obj["sort"]) {
      if (typeof item === "string") {
        const lastColon = item.lastIndexOf(":");
        if (lastColon !== -1) {
          const field = item.slice(0, lastColon);
          const dir = item.slice(lastColon + 1);
          if (field && (dir === "asc" || dir === "desc")) {
            sort.push({ field, dir });
          } else {
            sort.push({ field: item, dir: "asc" });
          }
        } else {
          sort.push({ field: item, dir: "asc" });
        }
      }
    }
  }

  const limit = typeof obj["limit"] === "number" ? obj["limit"] : 20;
  const offset = typeof obj["offset"] === "number" ? obj["offset"] : 0;

  const facets: string[] = Array.isArray(obj["facets"])
    ? (obj["facets"] as unknown[]).filter((x): x is string => typeof x === "string")
    : [];

  const attributesToRetrieve: string[] = Array.isArray(obj["attributesToRetrieve"])
    ? (obj["attributesToRetrieve"] as unknown[]).filter((x): x is string => typeof x === "string")
    : [];

  const hasPagePagination = "page" in obj || "hitsPerPage" in obj;

  return {
    valid: true,
    q,
    filter,
    filterIsArray,
    sort,
    limit,
    offset,
    facets,
    attributesToRetrieve,
    hasPagePagination,
  };
}

function sortRowsToStrings(rows: SortRow[]): string[] {
  return rows.filter((r) => r.field.trim() !== "").map((r) => `${r.field.trim()}:${r.dir}`);
}

function commaJoin(arr: string[]): string {
  return arr.join(", ");
}

function commaSplit(value: string): string[] {
  return value
    .split(",")
    .map((s) => s.trim())
    .filter((s) => s.length > 0);
}

// ─── Component ────────────────────────────────────────────────────────────────

export function SearchBuilder({
  queryText,
  onQueryChange,
  indexUID,
  onIndexChange,
  indexOptions,
  onExecute,
  executing,
}: SearchBuilderProps) {
  const parsed: Parsed = useMemo(() => parseQueryText(queryText), [queryText]);

  if (!parsed.valid) {
    return (
      <div className="sb-container">
        <div className="sb-warning">
          Query JSON is invalid — edit it in the JSON editor
        </div>
        {/* Index selector — still usable even when JSON is invalid */}
        <div className="sb-field-row">
          <label>Index</label>
          <input
            className="form-input"
            list="sb-index-list"
            value={indexUID}
            onChange={(e) => onIndexChange(e.target.value)}
            placeholder="Index UID…"
            disabled={executing}
          />
          <datalist id="sb-index-list">
            {indexOptions.map((i) => (
              <option key={i.uid} value={i.uid} />
            ))}
          </datalist>
        </div>
        <div className="sb-field-row">
          <button
            className="btn btn--primary btn--sm"
            onClick={onExecute}
            disabled={executing || !indexUID.trim()}
          >
            {executing ? "Running…" : "▶ Execute"}
          </button>
        </div>
      </div>
    );
  }

  const {
    q,
    filter,
    filterIsArray,
    sort,
    limit,
    offset,
    facets,
    attributesToRetrieve,
    hasPagePagination,
  } = parsed;

  // ── Handlers ────────────────────────────────────────────────────────────────

  function handleQChange(value: string) {
    onQueryChange(patchQuery(queryText, { q: value }));
  }

  function handleFilterChange(value: string) {
    onQueryChange(patchQuery(queryText, { filter: value }));
  }

  function handleLimitChange(value: number) {
    onQueryChange(patchQuery(queryText, { limit: value }));
  }

  function handleOffsetChange(value: number) {
    onQueryChange(patchQuery(queryText, { offset: value }));
  }

  function handleFacetsChange(value: string) {
    onQueryChange(patchQuery(queryText, { facets: commaSplit(value) }));
  }

  function handleAttributesToRetrieveChange(value: string) {
    onQueryChange(patchQuery(queryText, { attributesToRetrieve: commaSplit(value) }));
  }

  function handleSortFieldChange(index: number, field: string) {
    const updated = sort.map((row, i) => (i === index ? { ...row, field } : row));
    onQueryChange(patchQuery(queryText, { sort: sortRowsToStrings(updated) }));
  }

  function handleSortDirToggle(index: number) {
    const updated = sort.map((row, i) =>
      i === index ? { ...row, dir: row.dir === "asc" ? "desc" : "asc" } as SortRow : row,
    );
    onQueryChange(patchQuery(queryText, { sort: sortRowsToStrings(updated) }));
  }

  function handleSortAdd() {
    const updated = [...sort, { field: "", dir: "asc" as const }];
    onQueryChange(patchQuery(queryText, { sort: sortRowsToStrings(updated) }));
  }

  function handleSortRemove(index: number) {
    const updated = sort.filter((_, i) => i !== index);
    onQueryChange(patchQuery(queryText, { sort: sortRowsToStrings(updated) }));
  }

  // ── Render ──────────────────────────────────────────────────────────────────

  return (
    <div className="sb-container">
      {/* Index */}
      <div className="sb-field-row">
        <label>Index</label>
        <input
          className="form-input"
          list="sb-index-list"
          value={indexUID}
          onChange={(e) => onIndexChange(e.target.value)}
          placeholder="Index UID…"
          disabled={executing}
        />
        <datalist id="sb-index-list">
          {indexOptions.map((i) => (
            <option key={i.uid} value={i.uid} />
          ))}
        </datalist>
      </div>

      {/* Execute */}
      <div className="sb-field-row">
        <button
          className="btn btn--primary btn--sm"
          onClick={onExecute}
          disabled={executing || !indexUID.trim()}
        >
          {executing ? "Running…" : "▶ Execute"}
        </button>
      </div>

      {/* Search query */}
      <div className="sb-field-row">
        <label htmlFor="sb-q">Query (q)</label>
        <input
          id="sb-q"
          className="form-input"
          type="text"
          value={q}
          onChange={(e) => handleQChange(e.target.value)}
          placeholder='Search query string…'
          disabled={executing}
        />
      </div>

      {/* Filter */}
      <div className="sb-field-row">
        <label htmlFor="sb-filter">Filter</label>
        {filterIsArray ? (
          <div className="sb-warning">
            Array filter — use JSON editor
          </div>
        ) : (
          <textarea
            id="sb-filter"
            className="form-input"
            rows={3}
            value={filter ?? ""}
            onChange={(e) => handleFilterChange(e.target.value)}
            placeholder='e.g. genre = "action" AND year > 2000'
            disabled={executing}
          />
        )}
      </div>

      {/* Sort */}
      <div className="sb-field-row">
        <label>Sort</label>
        <div className="sb-sort-list">
          {sort.map((row, i) => (
            <div key={i} className="sb-sort-row">
              <input
                className="form-input"
                type="text"
                value={row.field}
                onChange={(e) => handleSortFieldChange(i, e.target.value)}
                placeholder="field name"
                disabled={executing}
              />
              <button
                className="btn btn--ghost btn--sm"
                type="button"
                onClick={() => handleSortDirToggle(i)}
                disabled={executing}
                title="Toggle direction"
              >
                {row.dir === "asc" ? "ASC" : "DESC"}
              </button>
              <button
                className="btn btn--ghost btn--sm"
                type="button"
                onClick={() => handleSortRemove(i)}
                disabled={executing}
                title="Remove sort"
              >
                ×
              </button>
            </div>
          ))}
          <button
            className="btn btn--ghost btn--sm"
            type="button"
            onClick={handleSortAdd}
            disabled={executing}
          >
            + Add sort
          </button>
        </div>
      </div>

      {/* Limit / Offset */}
      {hasPagePagination ? (
        <div className="sb-warning">
          Page-based pagination detected — offset/limit controls show extracted values only
        </div>
      ) : null}

      <div className="sb-field-row">
        <label htmlFor="sb-limit">Limit</label>
        <input
          id="sb-limit"
          className="form-input"
          type="number"
          min={0}
          max={1000}
          value={limit}
          onChange={(e) => handleLimitChange(Number(e.target.value))}
          disabled={executing || hasPagePagination}
        />
      </div>

      <div className="sb-field-row">
        <label htmlFor="sb-offset">Offset</label>
        <input
          id="sb-offset"
          className="form-input"
          type="number"
          min={0}
          value={offset}
          onChange={(e) => handleOffsetChange(Number(e.target.value))}
          disabled={executing || hasPagePagination}
        />
      </div>

      {/* Facets */}
      <div className="sb-field-row">
        <label htmlFor="sb-facets">Facets</label>
        <input
          id="sb-facets"
          className="form-input"
          type="text"
          value={commaJoin(facets)}
          onChange={(e) => handleFacetsChange(e.target.value)}
          placeholder="genre, year, author"
          disabled={executing}
        />
      </div>

      {/* Attributes to retrieve */}
      <div className="sb-field-row">
        <label htmlFor="sb-attrs">Attributes to Retrieve</label>
        <input
          id="sb-attrs"
          className="form-input"
          type="text"
          value={commaJoin(attributesToRetrieve)}
          onChange={(e) => handleAttributesToRetrieveChange(e.target.value)}
          placeholder="id, title, description"
          disabled={executing}
        />
      </div>
    </div>
  );
}
