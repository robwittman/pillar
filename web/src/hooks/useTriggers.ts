import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { triggersApi } from '../api/triggers'
import type { CreateTriggerRequest, UpdateTriggerRequest } from '../api/types'

export function useTriggers(sourceId?: string) {
  return useQuery({
    queryKey: sourceId ? ['triggers', { source_id: sourceId }] : ['triggers'],
    queryFn: () => triggersApi.list(sourceId),
  })
}

export function useTrigger(id: string) {
  return useQuery({
    queryKey: ['triggers', id],
    queryFn: () => triggersApi.get(id),
    enabled: !!id,
  })
}

export function useCreateTrigger() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: CreateTriggerRequest) => triggersApi.create(req),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['triggers'] }),
  })
}

export function useUpdateTrigger() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, ...req }: { id: string } & UpdateTriggerRequest) => triggersApi.update(id, req),
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: ['triggers', vars.id] })
      qc.invalidateQueries({ queryKey: ['triggers'] })
    },
  })
}

export function useDeleteTrigger() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => triggersApi.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['triggers'] }),
  })
}
