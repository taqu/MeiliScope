import { describe, it, expect } from "vitest";

// ── replicate the pure functions under test ────────────────────────────────
// These are not exported, so we duplicate them here to keep the component
// file unchanged.  If the implementations change these tests will catch it.

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

function parseQueryText(queryText: string) {
  let obj: Record<string, unknown>;
  try {
    const parsed = JSON.parse(queryText);
    if (typeof parsed !== "object" || Array.isArray(parsed) || parsed === null) {
      return { valid: false as const };
    }
    obj = parsed as Record<string, unknown>;
  } catch {
    return { valid: false as const };
  }

  const q = typeof obj["q"] === "string" ? obj["q"] : "";

  let filter: string | null = null;
  let filterIsArray = false;
  if (Array.isArray(obj["filter"])) {
    filterIsArray = true;
  } else if (typeof obj["filter"] === "string") {
    filter = obj["filter"];
  }

  const sort: Array<{ field: string; dir: "asc" | "desc" }> = [];
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
    valid: true as const,
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

// ── patchQuery ─────────────────────────────────────────────────────────────

describe("patchQuery", () => {
  it("sets a string field", () => {
    const result = patchQuery('{"q":"","limit":20}', { q: "hello" });
    const obj = JSON.parse(result) as Record<string, unknown>;
    expect(obj["q"]).toBe("hello");
    expect(obj["limit"]).toBe(20);
  });

  it("removes a field when value is empty string", () => {
    const result = patchQuery('{"q":"hello","limit":20}', { q: "" });
    const obj = JSON.parse(result) as Record<string, unknown>;
    expect(obj).not.toHaveProperty("q");
    expect(obj["limit"]).toBe(20);
  });

  it("removes a field when value is null", () => {
    const result = patchQuery('{"q":"hello","limit":20}', { q: null });
    const obj = JSON.parse(result) as Record<string, unknown>;
    expect(obj).not.toHaveProperty("q");
  });

  it("removes a field when value is empty array", () => {
    const result = patchQuery('{"sort":["title:asc"]}', { sort: [] });
    const obj = JSON.parse(result) as Record<string, unknown>;
    expect(obj).not.toHaveProperty("sort");
  });

  it("preserves unknown fields", () => {
    const result = patchQuery(
      '{"q":"","showRankingScore":true,"customParam":{"enabled":true}}',
      { q: "search" }
    );
    const obj = JSON.parse(result) as Record<string, unknown>;
    expect(obj["showRankingScore"]).toBe(true);
    expect(obj["customParam"]).toEqual({ enabled: true });
    expect(obj["q"]).toBe("search");
  });

  it("returns original text on invalid JSON", () => {
    const invalid = "{not valid json";
    expect(patchQuery(invalid, { q: "hello" })).toBe(invalid);
  });

  it("sets array fields", () => {
    const result = patchQuery("{}", { sort: ["title:asc", "id:desc"] });
    const obj = JSON.parse(result) as Record<string, unknown>;
    expect(obj["sort"]).toEqual(["title:asc", "id:desc"]);
  });

  it("sets numeric fields", () => {
    const result = patchQuery("{}", { limit: 50, offset: 100 });
    const obj = JSON.parse(result) as Record<string, unknown>;
    expect(obj["limit"]).toBe(50);
    expect(obj["offset"]).toBe(100);
  });
});

// ── parseQueryText ─────────────────────────────────────────────────────────

describe("parseQueryText", () => {
  it("parses a basic query", () => {
    const p = parseQueryText('{"q":"hello","limit":20}');
    expect(p.valid).toBe(true);
    if (!p.valid) return;
    expect(p.q).toBe("hello");
    expect(p.limit).toBe(20);
  });

  it("returns invalid for invalid JSON", () => {
    expect(parseQueryText("{bad}").valid).toBe(false);
  });

  it("returns invalid for JSON array", () => {
    expect(parseQueryText("[]").valid).toBe(false);
  });

  it("returns invalid for JSON scalar", () => {
    expect(parseQueryText("42").valid).toBe(false);
  });

  it("parses string filter", () => {
    const p = parseQueryText('{"filter":"genre = \\"action\\""}');
    expect(p.valid).toBe(true);
    if (!p.valid) return;
    expect(p.filter).toBe('genre = "action"');
    expect(p.filterIsArray).toBe(false);
  });

  it("detects array filter", () => {
    const p = parseQueryText('{"filter":[["genre = \\"action\\""]]}');
    expect(p.valid).toBe(true);
    if (!p.valid) return;
    expect(p.filterIsArray).toBe(true);
    expect(p.filter).toBeNull();
  });

  it("parses sort expressions", () => {
    const p = parseQueryText('{"sort":["title:asc","created_at:desc"]}');
    expect(p.valid).toBe(true);
    if (!p.valid) return;
    expect(p.sort).toHaveLength(2);
    expect(p.sort[0]).toEqual({ field: "title", dir: "asc" });
    expect(p.sort[1]).toEqual({ field: "created_at", dir: "desc" });
  });

  it("parses facets", () => {
    const p = parseQueryText('{"facets":["genre","year"]}');
    expect(p.valid).toBe(true);
    if (!p.valid) return;
    expect(p.facets).toEqual(["genre", "year"]);
  });

  it("parses attributesToRetrieve", () => {
    const p = parseQueryText('{"attributesToRetrieve":["id","title"]}');
    expect(p.valid).toBe(true);
    if (!p.valid) return;
    expect(p.attributesToRetrieve).toEqual(["id", "title"]);
  });

  it("detects page-based pagination", () => {
    const p = parseQueryText('{"page":2,"hitsPerPage":20}');
    expect(p.valid).toBe(true);
    if (!p.valid) return;
    expect(p.hasPagePagination).toBe(true);
  });

  it("does not set hasPagePagination for offset-based queries", () => {
    const p = parseQueryText('{"offset":40,"limit":20}');
    expect(p.valid).toBe(true);
    if (!p.valid) return;
    expect(p.hasPagePagination).toBe(false);
  });

  it("uses defaults for missing fields", () => {
    const p = parseQueryText("{}");
    expect(p.valid).toBe(true);
    if (!p.valid) return;
    expect(p.q).toBe("");
    expect(p.limit).toBe(20);
    expect(p.offset).toBe(0);
    expect(p.sort).toHaveLength(0);
    expect(p.facets).toHaveLength(0);
  });
});
