import * as TaskSvc from "../../bindings/github.com/taqu/meiliscope/internal/services/taskservice.js";

export interface TaskError {
  message: string;
  code: string;
  type: string;
  link: string;
}

export interface TaskResult {
  uid: number;
  indexUid: string;
  status: string;
  type: string;
  error?: TaskError | null;
  duration?: string;
  enqueuedAt: string;
  startedAt?: string;
  finishedAt?: string;
  rawJson: string;
}

export const TaskService = {
  getTask: (taskUID: number): Promise<TaskResult> =>
    TaskSvc.GetTask(taskUID) as Promise<TaskResult>,
};
