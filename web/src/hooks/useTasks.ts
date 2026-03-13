import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { tasksApi } from '../api/tasks'
import type { CreateTaskRequest } from '../api/types'

export function useTasks(agentId?: string) {
  return useQuery({
    queryKey: agentId ? ['tasks', { agent_id: agentId }] : ['tasks'],
    queryFn: () => tasksApi.list(agentId),
    refetchInterval: 5000,
  })
}

export function useAgentTasks(agentId: string) {
  return useQuery({
    queryKey: ['agents', agentId, 'tasks'],
    queryFn: () => tasksApi.listByAgent(agentId),
    enabled: !!agentId,
    refetchInterval: 5000,
  })
}

export function useTask(id: string) {
  return useQuery({
    queryKey: ['tasks', id],
    queryFn: () => tasksApi.get(id),
    enabled: !!id,
    refetchInterval: 5000,
  })
}

export function useCreateTask() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: CreateTaskRequest) => tasksApi.create(req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tasks'] })
    },
  })
}
