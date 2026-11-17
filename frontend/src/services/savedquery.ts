import * as SQSvc from "../../bindings/github.com/taqu/meiliscope/internal/services/savedqueryservice.js";

export interface SavedQuery {
  id: string;
  name: string;
  description?: string;
  connectionId: string;
  indexUid: string;
  queryJson: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateSavedQueryRequest {
  name: string;
  description?: string;
  connectionId: string;
  indexUid: string;
  queryJson: string;
}

export interface UpdateSavedQueryRequest {
  name: string;
  description?: string | null;
  connectionId: string;
  indexUid: string;
  queryJson: string;
}

export const SavedQueryService = {
  create: (req: CreateSavedQueryRequest): Promise<SavedQuery> =>
    SQSvc.CreateSavedQuery(req) as Promise<SavedQuery>,
  list: (): Promise<SavedQuery[]> =>
    SQSvc.ListSavedQueries() as Promise<SavedQuery[]>,
  get: (id: string): Promise<SavedQuery> =>
    SQSvc.GetSavedQuery(id) as Promise<SavedQuery>,
  update: (id: string, req: UpdateSavedQueryRequest): Promise<SavedQuery> =>
    SQSvc.UpdateSavedQuery(id, req) as Promise<SavedQuery>,
  delete: (id: string): Promise<void> =>
    SQSvc.DeleteSavedQuery(id) as Promise<void>,
  duplicate: (id: string): Promise<SavedQuery> =>
    SQSvc.DuplicateSavedQuery(id) as Promise<SavedQuery>,
};
