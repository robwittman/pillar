import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useSource, useUpdateSource, useDeleteSource, useRotateSourceSecret } from '../hooks/useSources'
import { useTriggers, useCreateTrigger, useDeleteTrigger, useUpdateTrigger } from '../hooks/useTriggers'
import { useAgents } from '../hooks/useAgents'
import type { CreateTriggerRequest, FilterCondition } from '../api/types'
import LoadingSpinner from '../components/shared/LoadingSpinner'
import ErrorAlert from '../components/shared/ErrorAlert'
import EmptyState from '../components/shared/EmptyState'
import ConfirmDialog from '../components/shared/ConfirmDialog'

export default function SourceDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data: source, isLoading, error } = useSource(id!)
  const updateSource = useUpdateSource()
  const deleteSource = useDeleteSource()
  const rotateSecret = useRotateSourceSecret()
  const [confirmDelete, setConfirmDelete] = useState(false)
  const [editName, setEditName] = useState('')
  const [editing, setEditing] = useState(false)
  const [rotatedSecret, setRotatedSecret] = useState<string | null>(null)

  if (isLoading) return <LoadingSpinner />
  if (error || !source) return <ErrorAlert message="Source not found" />

  const handleUpdate = async () => {
    if (!editName.trim()) return
    await updateSource.mutateAsync({ id: id!, name: editName.trim() })
    setEditing(false)
  }

  const handleDelete = async () => {
    await deleteSource.mutateAsync(id!)
    navigate('/sources')
  }

  const handleRotate = async () => {
    const result = await rotateSecret.mutateAsync(id!)
    setRotatedSecret(result.secret)
  }

  const webhookUrl = `${window.location.origin}/api/v1/sources/${source.id}/events`

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          {editing ? (
            <div className="flex items-center gap-2">
              <input
                type="text"
                value={editName}
                onChange={e => setEditName(e.target.value)}
                className="rounded-md border border-gray-300 px-3 py-1 text-lg font-bold focus:outline-none focus:ring-2 focus:ring-slate-500"
                onKeyDown={e => e.key === 'Enter' && handleUpdate()}
              />
              <button onClick={handleUpdate} className="text-sm text-slate-600 hover:text-slate-800">Save</button>
              <button onClick={() => setEditing(false)} className="text-sm text-gray-500 hover:text-gray-700">Cancel</button>
            </div>
          ) : (
            <h1 className="text-2xl font-bold text-gray-900 cursor-pointer" onClick={() => { setEditName(source.name); setEditing(true) }}>
              {source.name}
            </h1>
          )}
          <p className="text-sm text-gray-500 font-mono mt-1">{source.id}</p>
        </div>
        <div className="flex gap-2">
          <button onClick={handleRotate} className="px-4 py-2 text-sm font-medium text-gray-700 border border-gray-300 rounded-md hover:bg-gray-50">
            Rotate Secret
          </button>
          <button onClick={() => setConfirmDelete(true)} className="px-4 py-2 text-sm font-medium text-red-700 border border-red-300 rounded-md hover:bg-red-50">
            Delete
          </button>
        </div>
      </div>

      {rotatedSecret && (
        <div className="mb-4 p-4 bg-green-50 border border-green-200 rounded-lg">
          <p className="text-sm font-medium text-green-800 mb-1">New secret (shown once):</p>
          <code className="text-sm font-mono bg-green-100 px-2 py-1 rounded break-all">{rotatedSecret}</code>
          <button onClick={() => setRotatedSecret(null)} className="ml-3 text-xs text-green-600 hover:text-green-800">Dismiss</button>
        </div>
      )}

      <div className="bg-white rounded-lg shadow p-6 mb-6">
        <h3 className="text-sm font-semibold text-gray-900 mb-3">Inbound Webhook URL</h3>
        <div className="flex items-center gap-2">
          <code className="flex-1 text-sm font-mono bg-gray-100 px-3 py-2 rounded break-all">{webhookUrl}</code>
          <button
            onClick={() => navigator.clipboard.writeText(webhookUrl)}
            className="px-3 py-2 text-xs text-slate-600 border border-gray-300 rounded hover:bg-gray-50"
          >
            Copy
          </button>
        </div>
        <p className="text-xs text-gray-500 mt-2">POST events here. Include <code className="bg-gray-100 px-1 rounded">X-Hub-Signature-256</code> header for HMAC verification.</p>
      </div>

      <TriggersSection sourceId={id!} />

      <ConfirmDialog
        open={confirmDelete}
        title="Delete Source"
        message={`Delete "${source.name}" and all its triggers? This cannot be undone.`}
        confirmLabel="Delete"
        onConfirm={handleDelete}
        onCancel={() => setConfirmDelete(false)}
      />
    </div>
  )
}

function TriggersSection({ sourceId }: { sourceId: string }) {
  const { data: triggers, isLoading } = useTriggers(sourceId)
  const { data: agents } = useAgents()
  const createTrigger = useCreateTrigger()
  const deleteTrigger = useDeleteTrigger()
  const updateTrigger = useUpdateTrigger()
  const [showCreate, setShowCreate] = useState(false)
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null)
  const [expandedId, setExpandedId] = useState<string | null>(null)

  // Create form state
  const [form, setForm] = useState({
    agent_id: '',
    name: '',
    task_template: '',
    conditions: [] as FilterCondition[],
  })

  const handleCreate = async () => {
    if (!form.agent_id || !form.name) return
    const req: CreateTriggerRequest = {
      source_id: sourceId,
      agent_id: form.agent_id,
      name: form.name,
      task_template: form.task_template,
      filter: { conditions: form.conditions },
    }
    await createTrigger.mutateAsync(req)
    setForm({ agent_id: '', name: '', task_template: '', conditions: [] })
    setShowCreate(false)
  }

  const handleDelete = async (id: string) => {
    await deleteTrigger.mutateAsync(id)
    setConfirmDeleteId(null)
  }

  const handleToggle = async (id: string, enabled: boolean) => {
    await updateTrigger.mutateAsync({ id, enabled: !enabled })
  }

  const addCondition = () => {
    setForm(f => ({
      ...f,
      conditions: [...f.conditions, { path: '', op: 'eq' as const, value: '' }],
    }))
  }

  const updateCondition = (idx: number, field: string, value: string) => {
    setForm(f => ({
      ...f,
      conditions: f.conditions.map((c, i) =>
        i === idx ? { ...c, [field]: value } : c
      ),
    }))
  }

  const removeCondition = (idx: number) => {
    setForm(f => ({ ...f, conditions: f.conditions.filter((_, i) => i !== idx) }))
  }

  if (isLoading) return <LoadingSpinner />

  const inputCls = "w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-500"

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold text-gray-900">Triggers</h2>
        <button onClick={() => setShowCreate(true)} className="px-4 py-2 text-sm font-medium text-white bg-slate-800 rounded-md hover:bg-slate-700">
          Add Trigger
        </button>
      </div>

      {showCreate && (
        <div className="mb-4 bg-white rounded-lg shadow p-4 space-y-3">
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
              <input type="text" value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} placeholder="PR Security Review" className={inputCls} />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Agent</label>
              <select value={form.agent_id} onChange={e => setForm(f => ({ ...f, agent_id: e.target.value }))} className={inputCls}>
                <option value="">Select agent...</option>
                {agents?.map(a => (
                  <option key={a.id} value={a.id}>{a.name}</option>
                ))}
              </select>
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Task Template</label>
            <textarea
              value={form.task_template}
              onChange={e => setForm(f => ({ ...f, task_template: e.target.value }))}
              rows={4}
              placeholder={'Review this PR for security issues: {{.pull_request.html_url}}\nRepo: {{.repository.full_name}}'}
              className={inputCls + " font-mono"}
            />
            <p className="text-xs text-gray-500 mt-1">Go template syntax. Event payload fields available as {'{{.field_name}}'}.</p>
          </div>

          <div>
            <div className="flex items-center justify-between mb-2">
              <label className="block text-sm font-medium text-gray-700">Filter Conditions</label>
              <button onClick={addCondition} className="text-xs text-slate-600 hover:text-slate-800">+ Add condition</button>
            </div>
            {form.conditions.length === 0 && (
              <p className="text-xs text-gray-500">No conditions = matches all events from this source.</p>
            )}
            {form.conditions.map((cond, idx) => (
              <div key={idx} className="flex gap-2 mb-2 items-center">
                <input type="text" value={cond.path} onChange={e => updateCondition(idx, 'path', e.target.value)} placeholder="action" className="flex-1 rounded-md border border-gray-300 px-2 py-1.5 text-sm font-mono" />
                <select value={cond.op} onChange={e => updateCondition(idx, 'op', e.target.value)} className="rounded-md border border-gray-300 px-2 py-1.5 text-sm">
                  <option value="eq">equals</option>
                  <option value="neq">not equals</option>
                  <option value="contains">contains</option>
                  <option value="exists">exists</option>
                </select>
                {cond.op !== 'exists' && (
                  <input type="text" value={cond.value || ''} onChange={e => updateCondition(idx, 'value', e.target.value)} placeholder="opened" className="flex-1 rounded-md border border-gray-300 px-2 py-1.5 text-sm font-mono" />
                )}
                <button onClick={() => removeCondition(idx)} className="text-xs text-red-500 hover:text-red-700">Remove</button>
              </div>
            ))}
          </div>

          <div className="flex gap-2">
            <button onClick={handleCreate} disabled={createTrigger.isPending} className="px-4 py-2 text-sm font-medium text-white bg-slate-800 rounded-md hover:bg-slate-700 disabled:opacity-50">Create Trigger</button>
            <button onClick={() => setShowCreate(false)} className="px-4 py-2 text-sm font-medium text-gray-700 border border-gray-300 rounded-md hover:bg-gray-50">Cancel</button>
          </div>
        </div>
      )}

      {!triggers || triggers.length === 0 ? (
        <EmptyState title="No triggers" description="Add a trigger to route events from this source to an agent." />
      ) : (
        <div className="space-y-3">
          {triggers.map(trigger => (
            <div key={trigger.id} className="bg-white rounded-lg shadow">
              <div className="px-4 py-3 flex items-center justify-between">
                <div className="flex items-center gap-3 cursor-pointer" onClick={() => setExpandedId(expandedId === trigger.id ? null : trigger.id)}>
                  <span className="text-gray-400 text-xs">{expandedId === trigger.id ? '▼' : '▶'}</span>
                  <div>
                    <span className="text-sm font-medium text-gray-900">{trigger.name}</span>
                    <span className="ml-2 text-xs text-gray-500">
                      → {agents?.find(a => a.id === trigger.agent_id)?.name || trigger.agent_id.substring(0, 8)}
                    </span>
                  </div>
                  <span className={`text-xs px-2 py-0.5 rounded-full ${trigger.enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'}`}>
                    {trigger.enabled ? 'enabled' : 'disabled'}
                  </span>
                  {trigger.filter.conditions.length > 0 && (
                    <span className="text-xs text-gray-400">{trigger.filter.conditions.length} condition{trigger.filter.conditions.length > 1 ? 's' : ''}</span>
                  )}
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={() => handleToggle(trigger.id, trigger.enabled)}
                    className="text-xs text-slate-600 hover:text-slate-800"
                  >
                    {trigger.enabled ? 'Disable' : 'Enable'}
                  </button>
                  <button
                    onClick={() => setConfirmDeleteId(trigger.id)}
                    className="text-xs text-red-600 hover:text-red-800"
                  >
                    Delete
                  </button>
                </div>
              </div>

              {expandedId === trigger.id && (
                <div className="px-4 pb-4 border-t border-gray-100 pt-3 space-y-3">
                  <div>
                    <p className="text-xs font-medium text-gray-500 mb-1">Task Template</p>
                    <pre className="text-xs bg-gray-50 p-3 rounded font-mono whitespace-pre-wrap">{trigger.task_template || '(raw payload)'}</pre>
                  </div>
                  {trigger.filter.conditions.length > 0 && (
                    <div>
                      <p className="text-xs font-medium text-gray-500 mb-1">Filter Conditions</p>
                      <div className="space-y-1">
                        {trigger.filter.conditions.map((c, i) => (
                          <div key={i} className="text-xs font-mono bg-gray-50 px-2 py-1 rounded">
                            <span className="text-blue-600">{c.path}</span>
                            <span className="text-gray-500"> {c.op} </span>
                            {c.op !== 'exists' && <span className="text-green-600">"{c.value}"</span>}
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                  <div className="text-xs text-gray-400">
                    ID: {trigger.id} | Created: {new Date(trigger.created_at).toLocaleString()}
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      <ConfirmDialog
        open={!!confirmDeleteId}
        title="Delete Trigger"
        message="Delete this trigger? Future events will no longer be routed."
        confirmLabel="Delete"
        onConfirm={() => confirmDeleteId && handleDelete(confirmDeleteId)}
        onCancel={() => setConfirmDeleteId(null)}
      />
    </div>
  )
}
