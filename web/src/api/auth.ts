import { api } from './client'

export interface AuthProvider {
  name: string
  type: 'local' | 'oidc' | 'github'
  auth_url?: string
}

export interface Principal {
  id: string
  type: 'user' | 'service_account'
  display_name: string
  email?: string
  roles: string[]
}

export interface APIToken {
  id: string
  name: string
  owner_id: string
  owner_type: string
  scopes: string[]
  expires_at?: string
  last_used_at?: string
  created_at: string
}

export interface ServiceAccount {
  id: string
  name: string
  description: string
  disabled: boolean
  created_at: string
  updated_at: string
}

export interface CreateTokenResponse {
  token: string
  meta: APIToken
}

export interface CreateServiceAccountResponse {
  service_account: ServiceAccount
  client_id: string
  client_secret: string
}

export interface RotateSecretResponse {
  client_id: string
  client_secret: string
}

export interface ProvidersResponse {
  providers: AuthProvider[]
  allow_signup: boolean
}

export const authApi = {
  listProviders: () =>
    api.get<ProvidersResponse>('/auth/providers'),

  login: (email: string, password: string) =>
    api.post<{ status: string }>('/auth/login', { email, password }),

  register: (email: string, password: string, displayName: string) =>
    api.post<{ status: string }>('/auth/register', { email, password, display_name: displayName }),

  logout: () =>
    api.post<{ status: string }>('/auth/logout'),

  me: () =>
    api.get<Principal>('/auth/me'),

  // API tokens
  createToken: (name: string, expiresAt?: string) =>
    api.post<CreateTokenResponse>('/api/v1/auth/tokens', { name, expires_at: expiresAt }),

  listTokens: () =>
    api.get<APIToken[]>('/api/v1/auth/tokens'),

  revokeToken: (id: string) =>
    api.del(`/api/v1/auth/tokens/${id}`),

  // Service accounts
  createServiceAccount: (name: string, description: string) =>
    api.post<CreateServiceAccountResponse>('/api/v1/auth/service-accounts', { name, description }),

  listServiceAccounts: () =>
    api.get<ServiceAccount[]>('/api/v1/auth/service-accounts'),

  deleteServiceAccount: (id: string) =>
    api.del(`/api/v1/auth/service-accounts/${id}`),

  rotateSecret: (id: string) =>
    api.post<RotateSecretResponse>(`/api/v1/auth/service-accounts/${id}/rotate-secret`),
}
