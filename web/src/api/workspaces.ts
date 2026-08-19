import { api } from './client';
import type { Workspace, Member, CreateWorkspaceRequest, UpdateWorkspaceRequest, ScoringConfig } from './types';

interface ListResponse {
  workspaces: Workspace[];
  total: number;
}

export async function listWorkspaces(userID: string, limit = 50, offset = 0): Promise<ListResponse> {
  return api.get(`/api/v1/workspaces?userID=${encodeURIComponent(userID)}&limit=${limit}&offset=${offset}`);
}

export async function createWorkspace(req: CreateWorkspaceRequest): Promise<Workspace> {
  const r = await api.post<{ workspace: Workspace }>('/api/v1/workspaces', { req });
  return r.workspace;
}

export async function getWorkspace(id: string): Promise<Workspace> {
  return api.get(`/api/v1/workspaces/${id}`);
}

export async function updateWorkspace(id: string, req: UpdateWorkspaceRequest): Promise<Workspace> {
  return api.patch(`/api/v1/workspaces/${id}`, { req });
}

export async function deleteWorkspace(id: string): Promise<void> {
  return api.del(`/api/v1/workspaces/${id}`);
}

export async function setScoringConfig(id: string, config: ScoringConfig): Promise<Workspace> {
  return api.put(`/api/v1/workspaces/${id}/scoring`, { config });
}

export async function listMembers(id: string): Promise<{ members: Member[] }> {
  return api.get(`/api/v1/workspaces/${id}/members`);
}

export async function inviteMember(id: string, email: string, role: string): Promise<Member> {
  const r = await api.post<{ member: Member }>(`/api/v1/workspaces/${id}/members`, { email, role });
  return r.member;
}

export async function removeMember(id: string, memberID: string): Promise<void> {
  return api.del(`/api/v1/workspaces/${id}/members/${memberID}`);
}
