import { api } from './client'
import type { Webhook, WebhookDelivery, CreateWebhookRequest, UpdateWebhookRequest } from './types'

const base = '/api/v1/webhooks'

export const webhooksApi = {
  list: () => api.get<Webhook[]>(base),
  get: (id: string) => api.get<Webhook>(`${base}/${id}`),
  create: (req: CreateWebhookRequest) => api.post<Webhook>(base, req),
  update: (id: string, req: UpdateWebhookRequest) => api.put<Webhook>(`${base}/${id}`, req),
  delete: (id: string) => api.del(`${base}/${id}`),
  rotateSecret: (id: string) => api.post<Webhook>(`${base}/${id}/rotate-secret`),
  deliveries: (id: string) => api.get<WebhookDelivery[]>(`${base}/${id}/deliveries`),
}
