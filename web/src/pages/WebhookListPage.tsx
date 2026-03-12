import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useWebhooks, useCreateWebhook } from '../hooks/useWebhooks'
import StatusBadge from '../components/shared/StatusBadge'
import LoadingSpinner from '../components/shared/LoadingSpinner'
import EmptyState from '../components/shared/EmptyState'
import ErrorAlert from '../components/shared/ErrorAlert'

export default function WebhookListPage() {
  const navigate = useNavigate()
  const { data: webhooks, isLoading, error } = useWebhooks()
  const createWebhook = useCreateWebhook()
  const [showCreate, setShowCreate] = useState(false)
  const [url, setUrl] = useState('')
  const [description, setDescription] = useState('')
  const [eventTypes, setEventTypes] = useState('')
  const [createError, setCreateError] = useState('')
  const [newSecret, setNewSecret] = useState('')

  const handleCreate = async () => {
    if (!url.trim()) return
    setCreateError('')
    try {
      const webhook = await createWebhook.mutateAsync({
        url: url.trim(),
        description: description.trim(),
        event_types: eventTypes.split(',').map(s => s.trim()).filter(Boolean),
      })
      if (webhook.secret) {
        setNewSecret(webhook.secret)
      }
      setUrl('')
      setDescription('')
      setEventTypes('')
      setShowCreate(false)
    } catch (e) {
      setCreateError(e instanceof Error ? e.message : 'Failed to create webhook')
    }
  }

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorAlert message="Failed to load webhooks" />

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Webhooks</h1>
        <button onClick={() => setShowCreate(true)} className="px-4 py-2 text-sm font-medium text-white bg-slate-800 rounded-md hover:bg-slate-700">
          Create Webhook
        </button>
      </div>

      {newSecret && (
        <div className="mb-6 bg-yellow-50 border border-yellow-200 rounded-lg p-4">
          <p className="text-sm font-semibold text-yellow-800 mb-1">Webhook Secret (shown once)</p>
          <code className="block bg-yellow-100 rounded px-3 py-2 text-sm font-mono break-all">{newSecret}</code>
          <button onClick={() => { navigator.clipboard.writeText(newSecret); }} className="mt-2 text-xs text-yellow-700 hover:text-yellow-900 font-medium">
            Copy to clipboard
          </button>
        </div>
      )}

      {showCreate && (
        <div className="mb-6 bg-white rounded-lg shadow p-4 space-y-3">
          <h3 className="text-sm font-semibold text-gray-900">New Webhook</h3>
          {createError && <ErrorAlert message={createError} />}
          <input type="text" value={url} onChange={e => setUrl(e.target.value)} placeholder="https://example.com/webhook" className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-500" />
          <input type="text" value={description} onChange={e => setDescription(e.target.value)} placeholder="Description (optional)" className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-500" />
          <input type="text" value={eventTypes} onChange={e => setEventTypes(e.target.value)} placeholder="Event types (comma-separated, e.g. agent.created, agent.started)" className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-500" />
          <div className="flex gap-2">
            <button onClick={handleCreate} disabled={createWebhook.isPending} className="px-4 py-2 text-sm font-medium text-white bg-slate-800 rounded-md hover:bg-slate-700 disabled:opacity-50">
              {createWebhook.isPending ? 'Creating...' : 'Create'}
            </button>
            <button onClick={() => setShowCreate(false)} className="px-4 py-2 text-sm font-medium text-gray-700 border border-gray-300 rounded-md hover:bg-gray-50">Cancel</button>
          </div>
        </div>
      )}

      {!webhooks || webhooks.length === 0 ? (
        <EmptyState title="No webhooks" description="Create a webhook to receive event notifications." />
      ) : (
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">URL</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Description</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Events</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {webhooks.map(wh => (
                <tr key={wh.id} onClick={() => navigate(`/webhooks/${wh.id}`)} className="hover:bg-gray-50 cursor-pointer">
                  <td className="px-6 py-4 text-sm font-mono text-gray-900 max-w-xs truncate">{wh.url}</td>
                  <td className="px-6 py-4 text-sm text-gray-500">{wh.description || '-'}</td>
                  <td className="px-6 py-4">
                    <div className="flex flex-wrap gap-1">
                      {(wh.event_types || []).map(et => (
                        <span key={et} className="inline-flex rounded bg-blue-50 px-1.5 py-0.5 text-xs text-blue-700">{et}</span>
                      ))}
                    </div>
                  </td>
                  <td className="px-6 py-4"><StatusBadge status={wh.status} /></td>
                  <td className="px-6 py-4 text-sm text-gray-500">{new Date(wh.created_at).toLocaleDateString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
