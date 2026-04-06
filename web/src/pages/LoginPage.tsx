import { useState } from 'react'
import { useAuth } from '../hooks/useAuth'

export default function LoginPage() {
  const { providers, allowSignup, login, register } = useAuth()
  const [mode, setMode] = useState<'login' | 'register'>('login')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const hasLocal = providers.some(p => p.type === 'local')
  const oauthProviders = providers.filter(p => p.type !== 'local')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      if (mode === 'register') {
        await register(email, password, displayName)
      } else {
        await login(email, password)
      }
    } catch (err: any) {
      setError(err?.message || (mode === 'register' ? 'Registration failed' : 'Invalid email or password'))
    } finally {
      setLoading(false)
    }
  }

  const switchMode = () => {
    setMode(m => m === 'login' ? 'register' : 'login')
    setError('')
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-slate-900">
      <div className="w-full max-w-sm">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-white tracking-tight">Pillar</h1>
          <p className="text-slate-400 text-sm mt-1">Agent Management</p>
        </div>

        <div className="bg-white rounded-lg shadow-xl p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">
            {mode === 'register' ? 'Create account' : 'Sign in'}
          </h2>

          {error && (
            <div className="mb-4 p-3 bg-red-50 border border-red-200 text-red-700 text-sm rounded-md">
              {error}
            </div>
          )}

          {hasLocal && (
            <form onSubmit={handleSubmit} className="space-y-4">
              {mode === 'register' && (
                <div>
                  <label htmlFor="displayName" className="block text-sm font-medium text-gray-700 mb-1">
                    Name
                  </label>
                  <input
                    id="displayName"
                    type="text"
                    value={displayName}
                    onChange={e => setDisplayName(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-slate-500 focus:border-transparent"
                    placeholder="Your name"
                  />
                </div>
              )}
              <div>
                <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1">
                  Email
                </label>
                <input
                  id="email"
                  type="email"
                  value={email}
                  onChange={e => setEmail(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-slate-500 focus:border-transparent"
                  placeholder="admin@pillar.local"
                  required
                />
              </div>
              <div>
                <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1">
                  Password
                </label>
                <input
                  id="password"
                  type="password"
                  value={password}
                  onChange={e => setPassword(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-slate-500 focus:border-transparent"
                  minLength={mode === 'register' ? 8 : undefined}
                  required
                />
              </div>
              <button
                type="submit"
                disabled={loading}
                className="w-full py-2 px-4 bg-slate-800 text-white text-sm font-medium rounded-md hover:bg-slate-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-slate-500 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {loading
                  ? (mode === 'register' ? 'Creating account...' : 'Signing in...')
                  : (mode === 'register' ? 'Create account' : 'Sign in')
                }
              </button>

              {allowSignup && (
                <p className="text-center text-sm text-gray-500">
                  {mode === 'login' ? (
                    <>Don't have an account? <button type="button" onClick={switchMode} className="text-slate-700 font-medium hover:underline">Register</button></>
                  ) : (
                    <>Already have an account? <button type="button" onClick={switchMode} className="text-slate-700 font-medium hover:underline">Sign in</button></>
                  )}
                </p>
              )}
            </form>
          )}

          {hasLocal && oauthProviders.length > 0 && (
            <div className="my-4 flex items-center">
              <div className="flex-1 border-t border-gray-200" />
              <span className="px-3 text-xs text-gray-500">or</span>
              <div className="flex-1 border-t border-gray-200" />
            </div>
          )}

          {oauthProviders.length > 0 && (
            <div className="space-y-2">
              {oauthProviders.map(p => (
                <a
                  key={p.name}
                  href={`/auth/oauth/${p.name}`}
                  className="block w-full py-2 px-4 border border-gray-300 text-sm font-medium text-gray-700 rounded-md text-center hover:bg-gray-50 transition-colors"
                >
                  Continue with {p.name.charAt(0).toUpperCase() + p.name.slice(1)}
                </a>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
