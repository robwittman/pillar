import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useWebhook, useUpdateWebhook, useDeleteWebhook, useRotateSecret, useDeliveries } from '../hooks/useWebhooks'
import StatusBadge from '../components/shared/StatusBadge'
import ConfirmDialog from '../components/shared/ConfirmDialog'
import LoadingSpinner from '../components/shared/LoadingSpinner'
import ErrorAlert from '../components/shared/ErrorAlert'

export default function WebhookDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data: webhook, isLoading, error } = useWebhook(id!)
  const { data: deliveries } = useDeliveries(id!)
  const updateWebhook = useUpdateWebhook()
  const deleteWebhook = useDeleteWebhook()
  const rotateSecret = useRotateSecret()

  const [confirmDelete, setConfirmDelete] = useState(false)
  const [confirmRotate, setConfirmRotate] = useState(false)
  const [newSecret, setNewSecret] = useState('')
  const [saveMsg, setSaveMsg] = useState('')

  const [url, setUrl] = useState('')
  const [description, setDescription] = useState('')
  const [eventTypes, setEventTypes] = useState('')
  const [status, setStatus] = useState('')
  const [initialized, setInitialized] = useState(false)

  if (isLoading) return <LoadingSpinner />
  if (error || !webhook) return <ErrorAlert message="Webhook not found" />

  if (!initialized) {
    setUrl(webhook.url)
    setDescription(webhook.description)
    setEventTypes((webhook.event_types || []).join(', '))
    setStatus(webhook.status)
    setInitialized(true)
  }

  const handleSave = async () => {
    setSaveMsg('')
    try {
      await updateWebhook.mutateAsync({
        id: id!,
        url,
        description,
        event_types: eventTypes.split(',').map(s => s.trim()).filter(Boolean),
        status: status as 'active' | 'inactive',
      })
      setSaveMsg('Saved')
      setTimeout(() => setSaveMsg(''), 2000)
    } catch (e) {
      setSaveMsg(e instanceof Error ? e.message : 'Save failed')
    }
  }

  const handleRotate = async () => {
    const result = await rotateSecret.mutateAsync(id!)
    if (result.secret) setNewSecret(result.secret)
    setConfirmRotate(false)
  }

  const handleDelete = async () => {
    await deleteWebhook.mutateAsync(id!)
    navigate('/webhooks')
  }

  const inputCls = "w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-500"

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 font-mono">{webhook.url}</h1>
          <p className="text-sm text-gray-500 mt-1">{webhook.description}</p>
        </div>
        <StatusBadge status={webhook.status} size="md" />
      </div>

      {newSecret && (
        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
          <p className="text-sm font-semibold text-yellow-800 mb-1">New Secret (shown once)</p>
          <code className="block bg-yellow-100 rounded px-3 py-2 text-sm font-mono break-all">{newSecret}</code>
          <button onClick={() => navigator.clipboard.writeText(newSecret)} className="mt-2 text-xs text-yellow-700 hover:text-yellow-900 font-medium">
            Copy to clipboard
          </button>
        </div>
      )}

      <div className="bg-white rounded-lg shadow p-6 space-y-4">
        <h3 className="text-sm font-semibold text-gray-900">Edit Webhook</h3>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">URL</label>
          <input type="text" value={url} onChange={e => setUrl(e.target.value)} className={inputCls} />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
          <input type="text" value={description} onChange={e => setDescription(e.target.value)} className={inputCls} />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Event Types (comma-separated)</label>
          <input type="text" value={eventTypes} onChange={e => setEventTypes(e.target.value)} className={inputCls} />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Status</label>
          <select value={status} onChange={e => setStatus(e.target.value)} className={inputCls}>
            <option value="active">Active</option>
            <option value="inactive">Inactive</option>
          </select>
        </div>
        <div className="flex items-center gap-3 pt-2">
          <button onClick={handleSave} disabled={updateWebhook.isPending} className="px-4 py-2 text-sm font-medium text-white bg-slate-800 rounded-md hover:bg-slate-700 disabled:opacity-50">
            Save
          </button>
          <button onClick={() => setConfirmRotate(true)} className="px-4 py-2 text-sm font-medium text-yellow-700 border border-yellow-300 rounded-md hover:bg-yellow-50">
            Rotate Secret
          </button>
          <button onClick={() => setConfirmDelete(true)} className="px-4 py-2 text-sm font-medium text-red-700 border border-red-300 rounded-md hover:bg-red-50">
            Delete
          </button>
          {saveMsg && <span className="text-sm text-green-600">{saveMsg}</span>}
        </div>
      </div>

      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-sm font-semibold text-gray-900">Delivery Log</h3>
        </div>
        {!deliveries || deliveries.length === 0 ? (
          <p className="px-6 py-8 text-sm text-gray-500 text-center">No deliveries yet.</p>
        ) : (
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Event</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Response</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Attempts</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {deliveries.map(d => (
                <tr key={d.id}>
                  <td className="px-6 py-3 text-sm text-gray-900">{d.event_type}</td>
                  <td className="px-6 py-3"><StatusBadge status={d.status} /></td>
                  <td className="px-6 py-3 text-sm text-gray-500">{d.response_code || '-'}</td>
                  <td className="px-6 py-3 text-sm text-gray-500">{d.attempts}</td>
                  <td className="px-6 py-3 text-sm text-gray-500">{new Date(d.created_at).toLocaleString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      <ConfirmDialog open={confirmDelete} title="Delete Webhook" message="Delete this webhook? All deliveries will also be deleted." confirmLabel="Delete" onConfirm={handleDelete} onCancel={() => setConfirmDelete(false)} />
      <ConfirmDialog open={confirmRotate} title="Rotate Secret" message="Generate a new signing secret? The old secret will stop working immediately." confirmLabel="Rotate" onConfirm={handleRotate} onCancel={() => setConfirmRotate(false)} />
    </div>
  )
}
