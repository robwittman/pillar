import { api } from './client'
import type { Agent, AgentStatusInfo, CreateAgentRequest, UpdateAgentRequest } from './types'

const base = '/api/v1/agents'

export const agentsApi = {
  list: () => api.get<Agent[]>(base),
  get: (id: string) => api.get<Agent>(`${base}/${id}`),
  create: (req: CreateAgentRequest) => api.post<Agent>(base, req),
  update: (id: string, req: UpdateAgentRequest) => api.put<Agent>(`${base}/${id}`, req),
  delete: (id: string) => api.del(`${base}/${id}`),
  start: (id: string) => api.post<void>(`${base}/${id}/start`),
  stop: (id: string) => api.post<void>(`${base}/${id}/stop`),
  status: (id: string) => api.get<AgentStatusInfo>(`${base}/${id}/status`),
}
