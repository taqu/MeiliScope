export type ViewId = "connections" | "indexes" | "documents" | "query";

export interface NavItem {
  id: ViewId;
  label: string;
  title: string;
}

export const NAV_ITEMS: NavItem[] = [
  { id: "connections", label: "Conn", title: "Connections" },
  { id: "indexes", label: "Idx", title: "Index Explorer" },
  { id: "documents", label: "Docs", title: "Document Browser" },
  { id: "query", label: "Qry", title: "Query Console" },
];
