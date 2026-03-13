import { api } from './client'
import type { Source } from './types'

const base = '/api/v1/sources'

export const sourcesApi = {
  list: () => api.get<Source[]>(base),
  get: (id: string) => api.get<Source>(`${base}/${id}`),
  create: (name: string) => api.post<Source & { secret: string }>(base, { name }),
  update: (id: string, name: string) => api.put<Source>(`${base}/${id}`, { name }),
  delete: (id: string) => api.del(`${base}/${id}`),
  rotateSecret: (id: string) => api.post<Source & { secret: string }>(`${base}/${id}/rotate-secret`),
}
