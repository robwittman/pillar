import type { ReactNode } from 'react'
import { AuthContext, useAuthState } from '../hooks/useAuth'
import LoginPage from '../pages/LoginPage'

export default function AuthProvider({ children }: { children: ReactNode }) {
  const auth = useAuthState()

  if (auth.isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-900">
        <div className="text-white text-sm">Loading...</div>
      </div>
    )
  }

  return (
    <AuthContext.Provider value={auth}>
      {auth.isAuthenticated ? children : <LoginPage />}
    </AuthContext.Provider>
  )
}
