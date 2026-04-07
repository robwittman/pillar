import { useState } from 'react'
import { NavLink } from 'react-router-dom'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useAuth } from '../../hooks/useAuth'
import { useOrgContext } from '../../hooks/useOrgContext'
import { orgsApi } from '../../api/organizations'

const links = [
  { to: '/', label: 'Dashboard' },
  { to: '/agents', label: 'Agents' },
  { to: '/sources', label: 'Sources' },
  { to: '/tasks', label: 'Tasks' },
  { to: '/webhooks', label: 'Webhooks' },
  { to: '/settings', label: 'Settings' },
]

export default function Sidebar() {
  const { user, authEnabled, logout } = useAuth()
  const { orgs, activeOrg, setActiveOrg } = useOrgContext()
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [newName, setNewName] = useState('')
  const [newSlug, setNewSlug] = useState('')
  const qc = useQueryClient()

  const createMutation = useMutation({
    mutationFn: () => orgsApi.create(newName.trim(), newSlug.trim()),
    onSuccess: (org) => {
      setNewName('')
      setNewSlug('')
      setShowCreateForm(false)
      qc.invalidateQueries({ queryKey: ['organizations'] })
      setActiveOrg(org)
    },
  })

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault()
    if (newName.trim() && newSlug.trim()) createMutation.mutate()
  }

  return (
    <aside className="w-64 bg-slate-900 text-white flex flex-col">
      <div className="px-6 py-5 border-b border-slate-700">
        <h1 className="text-xl font-bold tracking-tight">Pillar</h1>
        <p className="text-xs text-slate-400 mt-0.5">Agent Management</p>
      </div>

      {authEnabled && orgs.length > 0 && (
        <div className="px-3 py-3 border-b border-slate-700">
          <select
            value={activeOrg?.id || ''}
            onChange={e => {
              const org = orgs.find(o => o.id === e.target.value)
              if (org) setActiveOrg(org)
            }}
            className="w-full px-2 py-1.5 bg-slate-800 text-white text-sm rounded-md border border-slate-600 focus:outline-none focus:ring-1 focus:ring-slate-500"
          >
            {orgs.map(org => (
              <option key={org.id} value={org.id}>
                {org.personal ? 'Personal' : org.name}
              </option>
            ))}
          </select>

          {!showCreateForm ? (
            <button
              onClick={() => setShowCreateForm(true)}
              className="mt-2 w-full px-2 py-1 text-xs text-slate-400 hover:text-white transition-colors text-left"
            >
              + New organization
            </button>
          ) : (
            <form onSubmit={handleCreate} className="mt-2 space-y-2">
              <input
                type="text"
                value={newName}
                onChange={e => {
                  setNewName(e.target.value)
                  if (!newSlug || newSlug === slugify(newName)) {
                    setNewSlug(slugify(e.target.value))
                  }
                }}
                placeholder="Name"
                className="w-full px-2 py-1 bg-slate-800 text-white text-xs rounded border border-slate-600 focus:outline-none focus:ring-1 focus:ring-slate-500"
                required
                autoFocus
              />
              <input
                type="text"
                value={newSlug}
                onChange={e => setNewSlug(e.target.value)}
                placeholder="slug"
                className="w-full px-2 py-1 bg-slate-800 text-white text-xs rounded border border-slate-600 focus:outline-none focus:ring-1 focus:ring-slate-500 font-mono"
                required
              />
              <div className="flex gap-2">
                <button
                  type="submit"
                  disabled={createMutation.isPending}
                  className="flex-1 px-2 py-1 bg-slate-700 text-white text-xs rounded hover:bg-slate-600 disabled:opacity-50"
                >
                  Create
                </button>
                <button
                  type="button"
                  onClick={() => { setShowCreateForm(false); setNewName(''); setNewSlug('') }}
                  className="px-2 py-1 text-xs text-slate-400 hover:text-white"
                >
                  Cancel
                </button>
              </div>
              {createMutation.isError && (
                <p className="text-xs text-red-400">Failed to create organization</p>
              )}
            </form>
          )}
        </div>
      )}

      <nav className="flex-1 px-3 py-4 space-y-1">
        {links.map(({ to, label }) => (
          <NavLink
            key={to}
            to={to}
            end={to === '/'}
            className={({ isActive }) =>
              `block px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                isActive
                  ? 'bg-slate-700 text-white'
                  : 'text-slate-300 hover:bg-slate-800 hover:text-white'
              }`
            }
          >
            {label}
          </NavLink>
        ))}
      </nav>
      {authEnabled && user && (
        <div className="px-3 py-4 border-t border-slate-700">
          <div className="px-3 mb-2">
            <p className="text-sm text-white truncate">{user.display_name || user.email}</p>
            {user.email && <p className="text-xs text-slate-400 truncate">{user.email}</p>}
          </div>
          <button
            onClick={() => logout()}
            className="block w-full px-3 py-2 text-left rounded-md text-sm font-medium text-slate-300 hover:bg-slate-800 hover:text-white transition-colors"
          >
            Sign out
          </button>
        </div>
      )}
    </aside>
  )
}

function slugify(text: string): string {
  return text.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/(^-|-$)/g, '')
}
