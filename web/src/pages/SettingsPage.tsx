import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { authApi, type APIToken, type ServiceAccount } from '../api/auth'
import { orgsApi, type Membership, type Team } from '../api/organizations'
import { useOrgContext } from '../hooks/useOrgContext'
import { useAuth } from '../hooks/useAuth'
import ConfirmDialog from '../components/shared/ConfirmDialog'

export default function SettingsPage() {
  const [tab, setTab] = useState<'tokens' | 'service-accounts' | 'organization'>('tokens')
  const { authEnabled } = useAuth()

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Settings</h1>

      <div className="border-b border-gray-200 mb-6">
        <nav className="flex space-x-6">
          {(['tokens', 'service-accounts', 'organization'] as const).map(t => (
            <button
              key={t}
              onClick={() => setTab(t)}
              className={`pb-2 text-sm font-medium border-b-2 transition-colors ${
                tab === t
                  ? 'border-slate-800 text-slate-900'
                  : 'border-transparent text-gray-500 hover:text-gray-700'
              }`}
            >
              {t === 'tokens' ? 'API Tokens' : t === 'service-accounts' ? 'Service Accounts' : 'Organization'}
            </button>
          ))}
        </nav>
      </div>

      {tab === 'tokens' && <APITokensSection />}
      {tab === 'service-accounts' && <ServiceAccountsSection />}
      {tab === 'organization' && authEnabled && <OrganizationSection />}
    </div>
  )
}

// --- Organization Section ---

function OrganizationSection() {
  const { activeOrg } = useOrgContext()

  if (!activeOrg) {
    return <p className="text-sm text-gray-500">No organization selected.</p>
  }

  return (
    <div className="space-y-8">
      <OrgDetailsCard />
      <MembersSection orgId={activeOrg.id} />
      <TeamsSection orgId={activeOrg.id} />
    </div>
  )
}

function OrgDetailsCard() {
  const { activeOrg } = useOrgContext()
  if (!activeOrg) return null

  return (
    <section>
      <h2 className="text-lg font-semibold text-gray-900 mb-3">Details</h2>
      <div className="bg-white border border-gray-200 rounded-md p-4 space-y-2">
        <div className="flex justify-between text-sm">
          <span className="text-gray-500">Name</span>
          <span className="text-gray-900">{activeOrg.name}</span>
        </div>
        <div className="flex justify-between text-sm">
          <span className="text-gray-500">Slug</span>
          <span className="text-gray-900 font-mono">{activeOrg.slug}</span>
        </div>
        <div className="flex justify-between text-sm">
          <span className="text-gray-500">Type</span>
          <span className="text-gray-900">{activeOrg.personal ? 'Personal' : 'Organization'}</span>
        </div>
        <div className="flex justify-between text-sm">
          <span className="text-gray-500">ID</span>
          <span className="text-gray-900 font-mono text-xs">{activeOrg.id}</span>
        </div>
      </div>
    </section>
  )
}

function MembersSection({ orgId }: { orgId: string }) {
  const qc = useQueryClient()
  const [userId, setUserId] = useState('')
  const [role, setRole] = useState('member')
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null)

  const { data: members = [] } = useQuery({
    queryKey: ['organizations', orgId, 'members'],
    queryFn: () => orgsApi.listMembers(orgId),
  })

  const addMutation = useMutation({
    mutationFn: () => orgsApi.addMember(orgId, userId.trim(), role),
    onSuccess: () => {
      setUserId('')
      qc.invalidateQueries({ queryKey: ['organizations', orgId, 'members'] })
    },
  })

  const removeMutation = useMutation({
    mutationFn: (uid: string) => orgsApi.removeMember(orgId, uid),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['organizations', orgId, 'members'] }),
  })

  const updateRoleMutation = useMutation({
    mutationFn: ({ uid, newRole }: { uid: string; newRole: string }) =>
      orgsApi.updateMemberRole(orgId, uid, newRole),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['organizations', orgId, 'members'] }),
  })

  const handleAdd = (e: React.FormEvent) => {
    e.preventDefault()
    if (userId.trim()) addMutation.mutate()
  }

  return (
    <section>
      <h2 className="text-lg font-semibold text-gray-900 mb-3">Members</h2>

      <form onSubmit={handleAdd} className="flex gap-2 mb-4">
        <input
          type="text"
          value={userId}
          onChange={e => setUserId(e.target.value)}
          placeholder="User ID"
          className="flex-1 px-3 py-1.5 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-slate-500"
          required
        />
        <select
          value={role}
          onChange={e => setRole(e.target.value)}
          className="px-3 py-1.5 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-slate-500"
        >
          <option value="owner">Owner</option>
          <option value="admin">Admin</option>
          <option value="member">Member</option>
          <option value="viewer">Viewer</option>
        </select>
        <button
          type="submit"
          disabled={addMutation.isPending}
          className="px-4 py-1.5 bg-slate-800 text-white text-sm rounded-md hover:bg-slate-700 disabled:opacity-50"
        >
          Add
        </button>
      </form>

      <div className="bg-white border border-gray-200 rounded-md overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">User ID</th>
              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Role</th>
              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Joined</th>
              <th className="px-4 py-2" />
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {members.length === 0 && (
              <tr><td colSpan={4} className="px-4 py-3 text-sm text-gray-500 text-center">No members</td></tr>
            )}
            {members.map((m: Membership) => (
              <tr key={m.id}>
                <td className="px-4 py-2 text-sm text-gray-900 font-mono">{m.user_id}</td>
                <td className="px-4 py-2">
                  <select
                    value={m.role}
                    onChange={e => updateRoleMutation.mutate({ uid: m.user_id, newRole: e.target.value })}
                    className="text-sm border border-gray-200 rounded px-2 py-0.5"
                  >
                    <option value="owner">Owner</option>
                    <option value="admin">Admin</option>
                    <option value="member">Member</option>
                    <option value="viewer">Viewer</option>
                  </select>
                </td>
                <td className="px-4 py-2 text-sm text-gray-500">{new Date(m.created_at).toLocaleDateString()}</td>
                <td className="px-4 py-2 text-right">
                  <button onClick={() => setDeleteTarget(m.user_id)} className="text-sm text-red-600 hover:underline">Remove</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <ConfirmDialog
        open={deleteTarget !== null}
        title="Remove Member"
        message="This user will lose access to all resources in this organization."
        onConfirm={() => { if (deleteTarget) removeMutation.mutate(deleteTarget); setDeleteTarget(null) }}
        onCancel={() => setDeleteTarget(null)}
      />
    </section>
  )
}

function TeamsSection({ orgId }: { orgId: string }) {
  const qc = useQueryClient()
  const [name, setName] = useState('')
  const [deleteId, setDeleteId] = useState<string | null>(null)

  const { data: teams = [] } = useQuery({
    queryKey: ['organizations', orgId, 'teams'],
    queryFn: () => orgsApi.listTeams(orgId),
  })

  const createMutation = useMutation({
    mutationFn: (n: string) => orgsApi.createTeam(orgId, n),
    onSuccess: () => {
      setName('')
      qc.invalidateQueries({ queryKey: ['organizations', orgId, 'teams'] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (teamId: string) => orgsApi.deleteTeam(orgId, teamId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['organizations', orgId, 'teams'] }),
  })

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault()
    if (name.trim()) createMutation.mutate(name.trim())
  }

  return (
    <section>
      <h2 className="text-lg font-semibold text-gray-900 mb-3">Teams</h2>

      <form onSubmit={handleCreate} className="flex gap-2 mb-4">
        <input
          type="text"
          value={name}
          onChange={e => setName(e.target.value)}
          placeholder="Team name"
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
              <th className="px-4 py-2" />
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {teams.length === 0 && (
              <tr><td colSpan={3} className="px-4 py-3 text-sm text-gray-500 text-center">No teams</td></tr>
            )}
            {teams.map((t: Team) => (
              <tr key={t.id}>
                <td className="px-4 py-2 text-sm text-gray-900">{t.name}</td>
                <td className="px-4 py-2 text-sm text-gray-500">{new Date(t.created_at).toLocaleDateString()}</td>
                <td className="px-4 py-2 text-right">
                  <button onClick={() => setDeleteId(t.id)} className="text-sm text-red-600 hover:underline">Delete</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <ConfirmDialog
        open={deleteId !== null}
        title="Delete Team"
        message="This will permanently delete the team and remove all members."
        onConfirm={() => { if (deleteId) deleteMutation.mutate(deleteId); setDeleteId(null) }}
        onCancel={() => setDeleteId(null)}
      />
    </section>
  )
}

// --- API Tokens Section ---

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

// --- Service Accounts Section ---

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
