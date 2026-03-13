import { api } from './client'
import type { Trigger, CreateTriggerRequest, UpdateTriggerRequest } from './types'

const base = '/api/v1/triggers'

export const triggersApi = {
  list: (sourceId?: string) => api.get<Trigger[]>(sourceId ? `${base}?source_id=${sourceId}` : base),
  get: (id: string) => api.get<Trigger>(`${base}/${id}`),
  create: (req: CreateTriggerRequest) => api.post<Trigger>(base, req),
  update: (id: string, req: UpdateTriggerRequest) => api.put<Trigger>(`${base}/${id}`, req),
  delete: (id: string) => api.del(`${base}/${id}`),
}
