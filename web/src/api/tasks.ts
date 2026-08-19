import { api } from './client';
import type { Task, TaskWithRank, CreateTaskRequest, UpdateTaskRequest } from './types';

export interface ListTasksResponse {
  tasks: Task[];
  total: number;
}

export interface RankedResponse {
  result: { tasks: TaskWithRank[] };
}

export interface TaskFilters {
  workspaceID: string;
  status?: string;
  search?: string;
  sortBy?: string;
  sortDesc?: boolean;
  limit?: number;
  offset?: number;
}

export async function listTasks(f: TaskFilters): Promise<ListTasksResponse> {
  const p = new URLSearchParams();
  p.set('workspaceID', f.workspaceID);
  if (f.status) p.set('status', f.status);
  if (f.search) p.set('search', f.search);
  if (f.sortBy) p.set('sortBy', f.sortBy);
  if (f.sortDesc) p.set('sortDesc', 'true');
  if (f.limit) p.set('limit', String(f.limit));
  if (f.offset) p.set('offset', String(f.offset));
  return api.get(`/api/v1/tasks?${p}`);
}

export async function createTask(req: CreateTaskRequest): Promise<Task> {
  const r = await api.post<{ task: Task }>('/api/v1/tasks', { req });
  return r.task;
}

export async function getTask(id: string): Promise<Task> {
  return api.get(`/api/v1/tasks/${id}`);
}

export async function updateTask(id: string, req: UpdateTaskRequest): Promise<Task> {
  return api.patch(`/api/v1/tasks/${id}`, { req });
}

export async function deleteTask(id: string): Promise<void> {
  return api.del(`/api/v1/tasks/${id}`);
}

export async function setScores(id: string, scores: Record<string, number>): Promise<Task> {
  return api.put(`/api/v1/tasks/${id}/scores`, { scores });
}

export async function vote(id: string, weight: number): Promise<Task> {
  return api.post(`/api/v1/tasks/${id}/vote`, { weight });
}

export async function removeVote(id: string): Promise<Task> {
  return api.del(`/api/v1/tasks/${id}/vote`, {});
}

export async function estimate(id: string, value: number, unit: string): Promise<Task> {
  return api.post(`/api/v1/tasks/${id}/estimate`, { value, unit });
}

export async function getRanked(workspaceID: string, limit = 50, offset = 0): Promise<RankedResponse> {
  return api.get(`/api/v1/workspaces/${workspaceID}/tasks/ranked?workspaceID=${encodeURIComponent(workspaceID)}&limit=${limit}&offset=${offset}`);
}
