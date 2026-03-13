import { api } from './client'
import type { Task, CreateTaskRequest } from './types'

const base = '/api/v1/tasks'

export const tasksApi = {
  list: (agentId?: string) => api.get<Task[]>(agentId ? `${base}?agent_id=${agentId}` : base),
  get: (id: string) => api.get<Task>(`${base}/${id}`),
  create: (req: CreateTaskRequest) => api.post<Task>(base, req),
  listByAgent: (agentId: string) => api.get<Task[]>(`/api/v1/agents/${agentId}/tasks`),
}
