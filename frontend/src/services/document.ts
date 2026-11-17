import * as DocSvc from "../../bindings/github.com/taqu/meiliscope/internal/services/documentservice.js";

export type {
  DocumentListOptions,
  DocumentPage,
  DocumentRecord,
} from "../../bindings/github.com/taqu/meiliscope/internal/services/models.js";

export const DocumentService = {
  listDocuments: (indexUID: string, opts: { offset: number; limit: number; fields?: string[] }) =>
    DocSvc.ListDocuments(indexUID, opts),
  getDocument: (indexUID: string, docID: string) =>
    DocSvc.GetDocument(indexUID, docID),
};
