import { createContext, useContext } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { authApi, type Principal, type AuthProvider } from '../api/auth'

export interface AuthContextValue {
  user: Principal | null
  providers: AuthProvider[]
  isLoading: boolean
  isAuthenticated: boolean
  authEnabled: boolean
  allowSignup: boolean
  login: (email: string, password: string) => Promise<void>
  register: (email: string, password: string, displayName: string) => Promise<void>
  logout: () => Promise<void>
}

export const AuthContext = createContext<AuthContextValue | null>(null)

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}

export function useAuthState() {
  const qc = useQueryClient()

  // Check if auth is enabled by trying to fetch providers.
  // If the endpoint 404s, auth is disabled.
  const providersQuery = useQuery({
    queryKey: ['auth', 'providers'],
    queryFn: async () => {
      try {
        return await authApi.listProviders()
      } catch {
        return null // auth not enabled
      }
    },
    retry: false,
    staleTime: Infinity,
  })

  const authEnabled = providersQuery.data != null && providersQuery.data.providers.length > 0

  // Only fetch /auth/me if auth is enabled.
  const meQuery = useQuery({
    queryKey: ['auth', 'me'],
    queryFn: async () => {
      try {
        return await authApi.me()
      } catch {
        return null
      }
    },
    enabled: authEnabled,
    retry: false,
    staleTime: 30000,
  })

  const loginMutation = useMutation({
    mutationFn: ({ email, password }: { email: string; password: string }) =>
      authApi.login(email, password),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['auth', 'me'] })
    },
  })

  const registerMutation = useMutation({
    mutationFn: ({ email, password, displayName }: { email: string; password: string; displayName: string }) =>
      authApi.register(email, password, displayName),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['auth', 'me'] })
    },
  })

  const logoutMutation = useMutation({
    mutationFn: () => authApi.logout(),
    onSuccess: () => {
      qc.setQueryData(['auth', 'me'], null)
    },
  })

  const isLoading = providersQuery.isLoading || (authEnabled && meQuery.isLoading)

  return {
    user: meQuery.data ?? null,
    providers: providersQuery.data?.providers ?? [],
    isLoading,
    isAuthenticated: !authEnabled || meQuery.data != null,
    authEnabled,
    allowSignup: providersQuery.data?.allow_signup ?? false,
    login: async (email: string, password: string) => {
      await loginMutation.mutateAsync({ email, password })
    },
    register: async (email: string, password: string, displayName: string) => {
      await registerMutation.mutateAsync({ email, password, displayName })
    },
    logout: async () => {
      await logoutMutation.mutateAsync()
    },
  }
}
