import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { sourcesApi } from '../api/sources'

export function useSources() {
  return useQuery({
    queryKey: ['sources'],
    queryFn: sourcesApi.list,
  })
}

export function useSource(id: string) {
  return useQuery({
    queryKey: ['sources', id],
    queryFn: () => sourcesApi.get(id),
    enabled: !!id,
  })
}

export function useCreateSource() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (name: string) => sourcesApi.create(name),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['sources'] }),
  })
}

export function useUpdateSource() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, name }: { id: string; name: string }) => sourcesApi.update(id, name),
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: ['sources', vars.id] })
      qc.invalidateQueries({ queryKey: ['sources'] })
    },
  })
}

export function useDeleteSource() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => sourcesApi.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['sources'] }),
  })
}

export function useRotateSourceSecret() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => sourcesApi.rotateSecret(id),
    onSuccess: (_, id) => qc.invalidateQueries({ queryKey: ['sources', id] }),
  })
}
