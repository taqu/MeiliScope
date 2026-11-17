import * as DocWriteSvc from "../../bindings/github.com/taqu/meiliscope/internal/services/documentwriteservice.js";

export interface TaskReference {
  taskUid: number;
  indexUid: string;
  status: string;
  type: string;
  enqueuedAt: string;
}

export const DocumentWriteService = {
  addDocument: (indexUID: string, documentJSON: string, primaryKeyField: string): Promise<TaskReference> =>
    DocWriteSvc.AddDocument(indexUID, documentJSON, primaryKeyField) as Promise<TaskReference>,
  updateDocument: (indexUID: string, documentJSON: string, primaryKeyField: string): Promise<TaskReference> =>
    DocWriteSvc.UpdateDocument(indexUID, documentJSON, primaryKeyField) as Promise<TaskReference>,
  replaceDocument: (indexUID: string, documentJSON: string, primaryKeyField: string): Promise<TaskReference> =>
    DocWriteSvc.ReplaceDocument(indexUID, documentJSON, primaryKeyField) as Promise<TaskReference>,
  deleteDocument: (indexUID: string, documentID: string): Promise<TaskReference> =>
    DocWriteSvc.DeleteDocument(indexUID, documentID) as Promise<TaskReference>,
};
