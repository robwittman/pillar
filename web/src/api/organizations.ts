import { api } from './client'

export interface Organization {
  id: string
  name: string
  slug: string
  personal: boolean
  owner_id: string
  created_at: string
  updated_at: string
}

export interface Membership {
  id: string
  org_id: string
  user_id: string
  role: 'owner' | 'admin' | 'member' | 'viewer'
  created_at: string
  updated_at: string
}

export interface Team {
  id: string
  org_id: string
  name: string
  created_at: string
  updated_at: string
}

export interface TeamMembership {
  id: string
  team_id: string
  user_id: string
  created_at: string
}

export const orgsApi = {
  list: () => api.get<Organization[]>('/api/v1/organizations'),
  get: (id: string) => api.get<Organization>(`/api/v1/organizations/${id}`),
  create: (name: string, slug: string) =>
    api.post<Organization>('/api/v1/organizations', { name, slug }),
  update: (id: string, data: { name?: string; slug?: string }) =>
    api.put<Organization>(`/api/v1/organizations/${id}`, data),
  delete: (id: string) => api.del(`/api/v1/organizations/${id}`),

  // Members
  listMembers: (orgId: string) =>
    api.get<Membership[]>(`/api/v1/organizations/${orgId}/members`),
  addMember: (orgId: string, userId: string, role: string) =>
    api.post<Membership>(`/api/v1/organizations/${orgId}/members`, { user_id: userId, role }),
  updateMemberRole: (orgId: string, userId: string, role: string) =>
    api.put<void>(`/api/v1/organizations/${orgId}/members/${userId}`, { role }),
  removeMember: (orgId: string, userId: string) =>
    api.del(`/api/v1/organizations/${orgId}/members/${userId}`),

  // Teams
  listTeams: (orgId: string) =>
    api.get<Team[]>(`/api/v1/organizations/${orgId}/teams`),
  createTeam: (orgId: string, name: string) =>
    api.post<Team>(`/api/v1/organizations/${orgId}/teams`, { name }),
  deleteTeam: (orgId: string, teamId: string) =>
    api.del(`/api/v1/organizations/${orgId}/teams/${teamId}`),
  listTeamMembers: (orgId: string, teamId: string) =>
    api.get<TeamMembership[]>(`/api/v1/organizations/${orgId}/teams/${teamId}/members`),
  addTeamMember: (orgId: string, teamId: string, userId: string) =>
    api.post<void>(`/api/v1/organizations/${orgId}/teams/${teamId}/members`, { user_id: userId }),
  removeTeamMember: (orgId: string, teamId: string, userId: string) =>
    api.del(`/api/v1/organizations/${orgId}/teams/${teamId}/members/${userId}`),
}
