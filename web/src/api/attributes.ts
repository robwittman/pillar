import { api } from './client'
import type { AgentAttribute } from './types'

export const attributesApi = {
  list: (agentId: string) => api.get<AgentAttribute[]>(`/api/v1/agents/${agentId}/attributes`),
  get: (agentId: string, namespace: string) => api.get<AgentAttribute>(`/api/v1/agents/${agentId}/attributes/${namespace}`),
  set: (agentId: string, namespace: string, value: unknown) => api.put<AgentAttribute>(`/api/v1/agents/${agentId}/attributes/${namespace}`, { value }),
  delete: (agentId: string, namespace: string) => api.del(`/api/v1/agents/${agentId}/attributes/${namespace}`),
}
