import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { attributesApi } from '../api/attributes'

export function useAttributes(agentId: string) {
  return useQuery({
    queryKey: ['agents', agentId, 'attributes'],
    queryFn: () => attributesApi.list(agentId),
    enabled: !!agentId,
  })
}

export function useSetAttribute() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ agentId, namespace, value }: { agentId: string; namespace: string; value: unknown }) =>
      attributesApi.set(agentId, namespace, value),
    onSuccess: (_, vars) => qc.invalidateQueries({ queryKey: ['agents', vars.agentId, 'attributes'] }),
  })
}

export function useDeleteAttribute() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ agentId, namespace }: { agentId: string; namespace: string }) =>
      attributesApi.delete(agentId, namespace),
    onSuccess: (_, vars) => qc.invalidateQueries({ queryKey: ['agents', vars.agentId, 'attributes'] }),
  })
}
