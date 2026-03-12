import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { configApi } from '../api/config'
import type { CreateConfigRequest, UpdateConfigRequest } from '../api/types'

export function useConfig(agentId: string) {
  return useQuery({
    queryKey: ['agents', agentId, 'config'],
    queryFn: () => configApi.get(agentId),
    enabled: !!agentId,
    retry: (count, error) => {
      if (error && 'status' in error && (error as { status: number }).status === 404) return false
      return count < 2
    },
  })
}

export function useCreateConfig() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ agentId, ...req }: { agentId: string } & CreateConfigRequest) =>
      configApi.create(agentId, req),
    onSuccess: (_, vars) => qc.invalidateQueries({ queryKey: ['agents', vars.agentId, 'config'] }),
  })
}

export function useUpdateConfig() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ agentId, ...req }: { agentId: string } & UpdateConfigRequest) =>
      configApi.update(agentId, req),
    onSuccess: (_, vars) => qc.invalidateQueries({ queryKey: ['agents', vars.agentId, 'config'] }),
  })
}

export function useDeleteConfig() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (agentId: string) => configApi.delete(agentId),
    onSuccess: (_, agentId) => qc.invalidateQueries({ queryKey: ['agents', agentId, 'config'] }),
  })
}
