import { api } from './client'
import type { AgentConfig, CreateConfigRequest, UpdateConfigRequest } from './types'

export const configApi = {
  get: (agentId: string) => api.get<AgentConfig>(`/api/v1/agents/${agentId}/config`),
  create: (agentId: string, req: CreateConfigRequest) => api.post<AgentConfig>(`/api/v1/agents/${agentId}/config`, req),
  update: (agentId: string, req: UpdateConfigRequest) => api.put<AgentConfig>(`/api/v1/agents/${agentId}/config`, req),
  delete: (agentId: string) => api.del(`/api/v1/agents/${agentId}/config`),
}
