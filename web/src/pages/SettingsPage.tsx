import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { authApi, type APIToken, type ServiceAccount } from '../api/auth'
import ConfirmDialog from '../components/shared/ConfirmDialog'

export default function SettingsPage() {
  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Settings</h1>
      <div className="space-y-8">
        <APITokensSection />
        <ServiceAccountsSection />
      </div>
    </div>
  )
}

function APITokensSection() {
  const qc = useQueryClient()
  const [name, setName] = useState('')
  const [newToken, setNewToken] = useState('')
  const [deleteId, setDeleteId] = useState<string | null>(null)

  const { data: tokens = [] } = useQuery({
    queryKey: ['auth', 'tokens'],
    queryFn: authApi.listTokens,
  })

  const createMutation = useMutation({
    mutationFn: (name: string) => authApi.createToken(name),
    onSuccess: (data) => {
      setNewToken(data.token)
      setName('')
      qc.invalidateQueries({ queryKey: ['auth', 'tokens'] })
    },
  })

  const revokeMutation = useMutation({
    mutationFn: (id: string) => authApi.revokeToken(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['auth', 'tokens'] }),
  })

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault()
    if (name.trim()) createMutation.mutate(name.trim())
  }

  return (
    <section>
      <h2 className="text-lg font-semibold text-gray-900 mb-3">API Tokens</h2>

      {newToken && (
        <div className="mb-4 p-3 bg-green-50 border border-green-200 rounded-md">
          <p className="text-sm font-medium text-green-800 mb-1">Token created. Copy it now — it won't be shown again.</p>
          <code className="block text-xs bg-green-100 p-2 rounded font-mono break-all select-all">{newToken}</code>
          <button onClick={() => setNewToken('')} className="mt-2 text-xs text-green-700 hover:underline">Dismiss</button>
        </div>
      )}

      <form onSubmit={handleCreate} className="flex gap-2 mb-4">
        <input
          type="text"
          value={name}
          onChange={e => setName(e.target.value)}
          placeholder="Token name"
          className="flex-1 px-3 py-1.5 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-slate-500"
          required
        />
        <button
          type="submit"
          disabled={createMutation.isPending}
          className="px-4 py-1.5 bg-slate-800 text-white text-sm rounded-md hover:bg-slate-700 disabled:opacity-50"
        >
          Create
        </button>
      </form>

      <div className="bg-white border border-gray-200 rounded-md overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Created</th>
              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Last Used</th>
              <th className="px-4 py-2" />
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {tokens.length === 0 && (
              <tr><td colSpan={4} className="px-4 py-3 text-sm text-gray-500 text-center">No tokens</td></tr>
            )}
            {tokens.map((t: APIToken) => (
              <tr key={t.id}>
                <td className="px-4 py-2 text-sm text-gray-900">{t.name}</td>
                <td className="px-4 py-2 text-sm text-gray-500">{new Date(t.created_at).toLocaleDateString()}</td>
                <td className="px-4 py-2 text-sm text-gray-500">{t.last_used_at ? new Date(t.last_used_at).toLocaleDateString() : 'Never'}</td>
                <td className="px-4 py-2 text-right">
                  <button onClick={() => setDeleteId(t.id)} className="text-sm text-red-600 hover:underline">Revoke</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <ConfirmDialog
        open={deleteId !== null}
        title="Revoke Token"
        message="This token will immediately stop working. This cannot be undone."
        onConfirm={() => { if (deleteId) revokeMutation.mutate(deleteId); setDeleteId(null) }}
        onCancel={() => setDeleteId(null)}
      />
    </section>
  )
}

function ServiceAccountsSection() {
  const qc = useQueryClient()
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [newCredentials, setNewCredentials] = useState<{ clientId: string; clientSecret: string } | null>(null)
  const [deleteId, setDeleteId] = useState<string | null>(null)
  const [rotatedSecret, setRotatedSecret] = useState<{ clientId: string; clientSecret: string } | null>(null)

  const { data: accounts = [] } = useQuery({
    queryKey: ['auth', 'service-accounts'],
    queryFn: authApi.listServiceAccounts,
  })

  const createMutation = useMutation({
    mutationFn: () => authApi.createServiceAccount(name.trim(), description.trim()),
    onSuccess: (data) => {
      setNewCredentials({ clientId: data.client_id, clientSecret: data.client_secret })
      setName('')
      setDescription('')
      qc.invalidateQueries({ queryKey: ['auth', 'service-accounts'] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => authApi.deleteServiceAccount(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['auth', 'service-accounts'] }),
  })

  const rotateMutation = useMutation({
    mutationFn: (id: string) => authApi.rotateSecret(id),
    onSuccess: (data) => {
      setRotatedSecret({ clientId: data.client_id, clientSecret: data.client_secret })
    },
  })

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault()
    if (name.trim()) createMutation.mutate()
  }

  return (
    <section>
      <h2 className="text-lg font-semibold text-gray-900 mb-3">Service Accounts</h2>

      {(newCredentials || rotatedSecret) && (
        <div className="mb-4 p-3 bg-green-50 border border-green-200 rounded-md">
          <p className="text-sm font-medium text-green-800 mb-1">
            {newCredentials ? 'Service account created.' : 'Secret rotated.'} Save these credentials — the secret won't be shown again.
          </p>
          <div className="text-xs bg-green-100 p-2 rounded font-mono space-y-1">
            <div><span className="text-green-700">Client ID:</span> <span className="select-all">{(newCredentials || rotatedSecret)!.clientId}</span></div>
            <div><span className="text-green-700">Secret:</span> <span className="select-all break-all">{(newCredentials || rotatedSecret)!.clientSecret}</span></div>
          </div>
          <button onClick={() => { setNewCredentials(null); setRotatedSecret(null) }} className="mt-2 text-xs text-green-700 hover:underline">Dismiss</button>
        </div>
      )}

      <form onSubmit={handleCreate} className="flex gap-2 mb-4">
        <input
          type="text"
          value={name}
          onChange={e => setName(e.target.value)}
          placeholder="Account name"
          className="flex-1 px-3 py-1.5 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-slate-500"
          required
        />
        <input
          type="text"
          value={description}
          onChange={e => setDescription(e.target.value)}
          placeholder="Description (optional)"
          className="flex-1 px-3 py-1.5 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-slate-500"
        />
        <button
          type="submit"
          disabled={createMutation.isPending}
          className="px-4 py-1.5 bg-slate-800 text-white text-sm rounded-md hover:bg-slate-700 disabled:opacity-50"
        >
          Create
        </button>
      </form>

      <div className="bg-white border border-gray-200 rounded-md overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Description</th>
              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Created</th>
              <th className="px-4 py-2" />
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {accounts.length === 0 && (
              <tr><td colSpan={4} className="px-4 py-3 text-sm text-gray-500 text-center">No service accounts</td></tr>
            )}
            {accounts.map((sa: ServiceAccount) => (
              <tr key={sa.id}>
                <td className="px-4 py-2 text-sm text-gray-900">{sa.name}</td>
                <td className="px-4 py-2 text-sm text-gray-500">{sa.description || '-'}</td>
                <td className="px-4 py-2 text-sm text-gray-500">{new Date(sa.created_at).toLocaleDateString()}</td>
                <td className="px-4 py-2 text-right space-x-3">
                  <button onClick={() => rotateMutation.mutate(sa.id)} className="text-sm text-slate-600 hover:underline">Rotate Secret</button>
                  <button onClick={() => setDeleteId(sa.id)} className="text-sm text-red-600 hover:underline">Delete</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <ConfirmDialog
        open={deleteId !== null}
        title="Delete Service Account"
        message="This will permanently delete the service account and invalidate its credentials."
        onConfirm={() => { if (deleteId) deleteMutation.mutate(deleteId); setDeleteId(null) }}
        onCancel={() => setDeleteId(null)}
      />
    </section>
  )
}
