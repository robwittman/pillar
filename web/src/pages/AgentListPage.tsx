import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAgents, useCreateAgent } from '../hooks/useAgents'
import StatusBadge from '../components/shared/StatusBadge'
import LoadingSpinner from '../components/shared/LoadingSpinner'
import EmptyState from '../components/shared/EmptyState'
import ErrorAlert from '../components/shared/ErrorAlert'

export default function AgentListPage() {
  const navigate = useNavigate()
  const { data: agents, isLoading, error } = useAgents()
  const createAgent = useCreateAgent()
  const [showCreate, setShowCreate] = useState(false)
  const [name, setName] = useState('')
  const [createError, setCreateError] = useState('')

  const handleCreate = async () => {
    if (!name.trim()) return
    setCreateError('')
    try {
      const agent = await createAgent.mutateAsync({ name: name.trim() })
      setName('')
      setShowCreate(false)
      navigate(`/agents/${agent.id}`)
    } catch (e) {
      setCreateError(e instanceof Error ? e.message : 'Failed to create agent')
    }
  }

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorAlert message="Failed to load agents" />

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Agents</h1>
        <button
          onClick={() => setShowCreate(true)}
          className="px-4 py-2 text-sm font-medium text-white bg-slate-800 rounded-md hover:bg-slate-700"
        >
          Create Agent
        </button>
      </div>

      {showCreate && (
        <div className="mb-6 bg-white rounded-lg shadow p-4">
          <h3 className="text-sm font-semibold text-gray-900 mb-3">New Agent</h3>
          {createError && <ErrorAlert message={createError} />}
          <div className="flex gap-3 mt-2">
            <input
              type="text"
              value={name}
              onChange={e => setName(e.target.value)}
              placeholder="Agent name"
              className="flex-1 rounded-md border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-500"
              onKeyDown={e => e.key === 'Enter' && handleCreate()}
            />
            <button onClick={handleCreate} disabled={createAgent.isPending} className="px-4 py-2 text-sm font-medium text-white bg-slate-800 rounded-md hover:bg-slate-700 disabled:opacity-50">
              {createAgent.isPending ? 'Creating...' : 'Create'}
            </button>
            <button onClick={() => { setShowCreate(false); setName('') }} className="px-4 py-2 text-sm font-medium text-gray-700 border border-gray-300 rounded-md hover:bg-gray-50">
              Cancel
            </button>
          </div>
        </div>
      )}

      {!agents || agents.length === 0 ? (
        <EmptyState
          title="No agents"
          description="Create your first agent to get started."
          action={
            <button onClick={() => setShowCreate(true)} className="text-sm text-slate-600 hover:text-slate-800 font-medium">
              Create Agent
            </button>
          }
        />
      ) : (
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Labels</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {agents.map(agent => (
                <tr
                  key={agent.id}
                  onClick={() => navigate(`/agents/${agent.id}`)}
                  className="hover:bg-gray-50 cursor-pointer"
                >
                  <td className="px-6 py-4 text-sm font-medium text-gray-900">{agent.name}</td>
                  <td className="px-6 py-4"><StatusBadge status={agent.status} /></td>
                  <td className="px-6 py-4">
                    <div className="flex flex-wrap gap-1">
                      {Object.entries(agent.labels || {}).map(([k, v]) => (
                        <span key={k} className="inline-flex rounded bg-gray-100 px-1.5 py-0.5 text-xs text-gray-600">
                          {k}={v}
                        </span>
                      ))}
                    </div>
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-500">{new Date(agent.created_at).toLocaleDateString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
