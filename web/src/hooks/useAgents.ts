import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { agentsApi } from '../api/agents'
import type { CreateAgentRequest, UpdateAgentRequest } from '../api/types'

export function useAgents() {
  return useQuery({
    queryKey: ['agents'],
    queryFn: agentsApi.list,
    refetchInterval: 10000,
  })
}

export function useAgent(id: string) {
  return useQuery({
    queryKey: ['agents', id],
    queryFn: () => agentsApi.get(id),
    enabled: !!id,
  })
}

export function useAgentStatus(id: string) {
  return useQuery({
    queryKey: ['agents', id, 'status'],
    queryFn: () => agentsApi.status(id),
    refetchInterval: 5000,
    enabled: !!id,
  })
}

export function useCreateAgent() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: CreateAgentRequest) => agentsApi.create(req),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['agents'] }),
  })
}

export function useUpdateAgent() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, ...req }: { id: string } & UpdateAgentRequest) => agentsApi.update(id, req),
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: ['agents', vars.id] })
      qc.invalidateQueries({ queryKey: ['agents'] })
    },
  })
}

export function useDeleteAgent() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => agentsApi.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['agents'] }),
  })
}

export function useStartAgent() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => agentsApi.start(id),
    onSuccess: (_, id) => {
      qc.invalidateQueries({ queryKey: ['agents', id] })
      qc.invalidateQueries({ queryKey: ['agents'] })
    },
  })
}

export function useStopAgent() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => agentsApi.stop(id),
    onSuccess: (_, id) => {
      qc.invalidateQueries({ queryKey: ['agents', id] })
      qc.invalidateQueries({ queryKey: ['agents'] })
    },
  })
}
