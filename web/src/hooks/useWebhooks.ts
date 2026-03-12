import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { webhooksApi } from '../api/webhooks'
import type { CreateWebhookRequest, UpdateWebhookRequest } from '../api/types'

export function useWebhooks() {
  return useQuery({
    queryKey: ['webhooks'],
    queryFn: webhooksApi.list,
  })
}

export function useWebhook(id: string) {
  return useQuery({
    queryKey: ['webhooks', id],
    queryFn: () => webhooksApi.get(id),
    enabled: !!id,
  })
}

export function useCreateWebhook() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: CreateWebhookRequest) => webhooksApi.create(req),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['webhooks'] }),
  })
}

export function useUpdateWebhook() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, ...req }: { id: string } & UpdateWebhookRequest) => webhooksApi.update(id, req),
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: ['webhooks', vars.id] })
      qc.invalidateQueries({ queryKey: ['webhooks'] })
    },
  })
}

export function useDeleteWebhook() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => webhooksApi.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['webhooks'] }),
  })
}

export function useRotateSecret() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => webhooksApi.rotateSecret(id),
    onSuccess: (_, id) => qc.invalidateQueries({ queryKey: ['webhooks', id] }),
  })
}

export function useDeliveries(webhookId: string) {
  return useQuery({
    queryKey: ['webhooks', webhookId, 'deliveries'],
    queryFn: () => webhooksApi.deliveries(webhookId),
    enabled: !!webhookId,
    refetchInterval: 10000,
  })
}
