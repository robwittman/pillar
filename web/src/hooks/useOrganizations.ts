import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { orgsApi } from '../api/organizations'

export function useOrganizations() {
  return useQuery({
    queryKey: ['organizations'],
    queryFn: orgsApi.list,
  })
}

export function useOrganization(id: string) {
  return useQuery({
    queryKey: ['organizations', id],
    queryFn: () => orgsApi.get(id),
    enabled: !!id,
  })
}

export function useCreateOrganization() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ name, slug }: { name: string; slug: string }) => orgsApi.create(name, slug),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['organizations'] }),
  })
}

export function useOrgMembers(orgId: string) {
  return useQuery({
    queryKey: ['organizations', orgId, 'members'],
    queryFn: () => orgsApi.listMembers(orgId),
    enabled: !!orgId,
  })
}

export function useAddMember(orgId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ userId, role }: { userId: string; role: string }) =>
      orgsApi.addMember(orgId, userId, role),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['organizations', orgId, 'members'] }),
  })
}

export function useRemoveMember(orgId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (userId: string) => orgsApi.removeMember(orgId, userId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['organizations', orgId, 'members'] }),
  })
}

export function useUpdateMemberRole(orgId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ userId, role }: { userId: string; role: string }) =>
      orgsApi.updateMemberRole(orgId, userId, role),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['organizations', orgId, 'members'] }),
  })
}

export function useOrgTeams(orgId: string) {
  return useQuery({
    queryKey: ['organizations', orgId, 'teams'],
    queryFn: () => orgsApi.listTeams(orgId),
    enabled: !!orgId,
  })
}

export function useCreateTeam(orgId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (name: string) => orgsApi.createTeam(orgId, name),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['organizations', orgId, 'teams'] }),
  })
}

export function useDeleteTeam(orgId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (teamId: string) => orgsApi.deleteTeam(orgId, teamId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['organizations', orgId, 'teams'] }),
  })
}
