// DTO-типы — зеркало pkg/types (json-теги camelCase).

export interface User {
  id: string;
  email: string;
  name: string;
  createdAt: string;
}

export interface Criterion {
  id: string;
  name: string;
  weight: number;
}

export interface ScoringConfig {
  type: string; // RICE | ICE | WSJF | CUSTOM
  criteria?: Criterion[];
  formula?: string;
}

export interface Workspace {
  id: string;
  name: string;
  description?: string;
  ownerId: string;
  scoring: ScoringConfig;
  createdAt: string;
  updatedAt: string;
}

export interface Member {
  id: string;
  workspaceId: string;
  userId: string;
  email: string;
  role: string; // owner | admin | member
  createdAt: string;
}

export interface Vote {
  userId: string;
  weight: number;
  createdAt: string;
}

export interface Estimation {
  userId: string;
  value: number;
  unit: string; // points | hours
  createdAt: string;
}

export interface Task {
  id: string;
  workspaceId: string;
  externalId?: string;
  externalType?: string;
  externalUrl?: string;
  title: string;
  description?: string;
  scores?: Record<string, number>;
  finalScore: number;
  votes?: Vote[];
  estimations?: Estimation[];
  dependencies?: string[];
  status: string; // backlog | in_progress | done | cancelled
  priority?: string; // low | medium | high | critical
  labels?: string[];
  assigneeId?: string;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
}

export interface TaskWithRank extends Task {
  rank: number;
  percentile: number;
}

export interface RegisterRequest {
  name: string;
  email: string;
  password: string;
}

export interface CreateWorkspaceRequest {
  name: string;
  description?: string;
  scoring?: ScoringConfig;
}

export interface UpdateWorkspaceRequest {
  name?: string;
  description?: string;
}

export interface CreateTaskRequest {
  workspaceId: string;
  title: string;
  description?: string;
  scores?: Record<string, number>;
  dependencies?: string[];
  status?: string;
  priority?: string;
  labels?: string[];
  assigneeId?: string;
}

export interface UpdateTaskRequest {
  title?: string;
  description?: string;
  status?: string;
  priority?: string;
  labels?: string[];
  assigneeId?: string;
}

export interface ListTasksRequest {
  workspaceId: string;
  status?: string;
  assigneeId?: string;
  labels?: string[];
  hasScore?: boolean;
  hasVotes?: boolean;
  minScore?: number;
  maxScore?: number;
  search?: string;
  sortBy?: string; // finalScore | createdAt | votes
  sortDesc?: boolean;
  limit?: number;
  offset?: number;
}

// Пресеты критериев скоринга (совпадают с бэкенд-конвенцией).
export const SCORING_PRESETS: Record<string, Criterion[]> = {
  RICE: [
    { id: 'reach', name: 'Reach', weight: 1 },
    { id: 'impact', name: 'Impact', weight: 1 },
    { id: 'confidence', name: 'Confidence', weight: 1 },
    { id: 'effort', name: 'Effort', weight: 1 },
  ],
  ICE: [
    { id: 'impact', name: 'Impact', weight: 1 },
    { id: 'confidence', name: 'Confidence', weight: 1 },
    { id: 'ease', name: 'Ease', weight: 1 },
  ],
  WSJF: [
    { id: 'businessValue', name: 'Business Value', weight: 1 },
    { id: 'timeCriticality', name: 'Time Criticality', weight: 1 },
    { id: 'riskReduction', name: 'Risk Reduction', weight: 1 },
    { id: 'jobSize', name: 'Job Size', weight: 1 },
  ],
};

export const TASK_STATUSES = ['backlog', 'in_progress', 'done', 'cancelled'] as const;
export const TASK_PRIORITIES = ['low', 'medium', 'high', 'critical'] as const;
