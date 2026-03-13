import { useState, useEffect, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useAgent, useAgentStatus, useStartAgent, useStopAgent, useDeleteAgent } from '../hooks/useAgents'
import { useConfig, useCreateConfig, useUpdateConfig, useDeleteConfig } from '../hooks/useConfig'
import { useAttributes, useSetAttribute, useDeleteAttribute } from '../hooks/useAttributes'
import { useLogs, useLogStream } from '../hooks/useLogs'
import { useAgentTasks } from '../hooks/useTasks'
import type { TaskStatus } from '../api/types'
import StatusBadge from '../components/shared/StatusBadge'
import OnlineIndicator from '../components/shared/OnlineIndicator'
import ConfirmDialog from '../components/shared/ConfirmDialog'
import KeyValueDisplay from '../components/shared/KeyValueDisplay'
import LoadingSpinner from '../components/shared/LoadingSpinner'
import ErrorAlert from '../components/shared/ErrorAlert'
import EmptyState from '../components/shared/EmptyState'

type Tab = 'overview' | 'config' | 'attributes' | 'logs' | 'tasks'

export default function AgentDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [tab, setTab] = useState<Tab>('overview')
  const [confirmDelete, setConfirmDelete] = useState(false)

  const { data: agent, isLoading, error } = useAgent(id!)
  const { data: statusInfo } = useAgentStatus(id!)
  const startAgent = useStartAgent()
  const stopAgent = useStopAgent()
  const deleteAgent = useDeleteAgent()

  if (isLoading) return <LoadingSpinner />
  if (error || !agent) return <ErrorAlert message="Agent not found" />

  const handleDelete = async () => {
    await deleteAgent.mutateAsync(id!)
    navigate('/agents')
  }

  const tabs: { key: Tab; label: string }[] = [
    { key: 'overview', label: 'Overview' },
    { key: 'config', label: 'Config' },
    { key: 'attributes', label: 'Attributes' },
    { key: 'logs', label: 'Logs' },
    { key: 'tasks', label: 'Tasks' },
  ]

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-3">
          <h1 className="text-2xl font-bold text-gray-900">{agent.name}</h1>
          <StatusBadge status={agent.status} size="md" />
          {statusInfo && <OnlineIndicator online={statusInfo.online} />}
        </div>
        <div className="flex gap-2">
          {agent.status !== 'running' && (
            <button
              onClick={() => startAgent.mutateAsync(id!)}
              disabled={startAgent.isPending}
              className="px-4 py-2 text-sm font-medium text-white bg-green-600 rounded-md hover:bg-green-700 disabled:opacity-50"
            >
              Start
            </button>
          )}
          {agent.status === 'running' && (
            <button
              onClick={() => stopAgent.mutateAsync(id!)}
              disabled={stopAgent.isPending}
              className="px-4 py-2 text-sm font-medium text-white bg-yellow-600 rounded-md hover:bg-yellow-700 disabled:opacity-50"
            >
              Stop
            </button>
          )}
          <button
            onClick={() => setConfirmDelete(true)}
            className="px-4 py-2 text-sm font-medium text-red-700 border border-red-300 rounded-md hover:bg-red-50"
          >
            Delete
          </button>
        </div>
      </div>

      <div className="border-b border-gray-200 mb-6">
        <nav className="flex gap-6">
          {tabs.map(t => (
            <button
              key={t.key}
              onClick={() => setTab(t.key)}
              className={`pb-3 text-sm font-medium border-b-2 transition-colors ${
                tab === t.key
                  ? 'border-slate-800 text-slate-900'
                  : 'border-transparent text-gray-500 hover:text-gray-700'
              }`}
            >
              {t.label}
            </button>
          ))}
        </nav>
      </div>

      {tab === 'overview' && <OverviewTab agent={agent} />}
      {tab === 'config' && <ConfigTab agentId={id!} />}
      {tab === 'attributes' && <AttributesTab agentId={id!} />}
      {tab === 'logs' && <LogsTab agentId={id!} />}
      {tab === 'tasks' && <TasksTab agentId={id!} />}

      <ConfirmDialog
        open={confirmDelete}
        title="Delete Agent"
        message={`Are you sure you want to delete "${agent.name}"? This action cannot be undone.`}
        confirmLabel="Delete"
        onConfirm={handleDelete}
        onCancel={() => setConfirmDelete(false)}
      />
    </div>
  )
}

function OverviewTab({ agent }: { agent: { metadata: Record<string, string>; labels: Record<string, string>; created_at: string; updated_at: string; id: string } }) {
  return (
    <div className="space-y-6">
      <div className="bg-white rounded-lg shadow p-6">
        <h3 className="text-sm font-semibold text-gray-900 mb-4">Details</h3>
        <dl className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <dt className="text-gray-500">ID</dt>
            <dd className="mt-1 font-mono text-gray-900">{agent.id}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Created</dt>
            <dd className="mt-1 text-gray-900">{new Date(agent.created_at).toLocaleString()}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Updated</dt>
            <dd className="mt-1 text-gray-900">{new Date(agent.updated_at).toLocaleString()}</dd>
          </div>
        </dl>
      </div>
      <div className="bg-white rounded-lg shadow p-6">
        <KeyValueDisplay data={agent.metadata} label="Metadata" />
      </div>
      <div className="bg-white rounded-lg shadow p-6">
        <KeyValueDisplay data={agent.labels} label="Labels" />
      </div>
    </div>
  )
}

function ConfigTab({ agentId }: { agentId: string }) {
  const { data: config, isLoading, error } = useConfig(agentId)
  const createConfig = useCreateConfig()
  const updateConfig = useUpdateConfig()
  const deleteConfig = useDeleteConfig()
  const [confirmDeleteConfig, setConfirmDeleteConfig] = useState(false)

  const [form, setForm] = useState({
    model_provider: '',
    model_id: '',
    system_prompt: '',
    api_credential: '',
    max_iterations: 0,
    token_budget: 0,
    task_timeout_seconds: 0,
    allowed_tools: '',
    denied_tools: '',
  })
  const [initialized, setInitialized] = useState(false)
  const [saveMsg, setSaveMsg] = useState('')

  if (isLoading) return <LoadingSpinner />

  const hasConfig = !!config && !error

  if (hasConfig && !initialized) {
    setForm({
      model_provider: config.model_provider || '',
      model_id: config.model_id || '',
      system_prompt: config.system_prompt || '',
      api_credential: '',
      max_iterations: config.max_iterations || 0,
      token_budget: config.token_budget || 0,
      task_timeout_seconds: config.task_timeout_seconds || 0,
      allowed_tools: config.tool_permissions?.allowed_tools?.join(', ') || '',
      denied_tools: config.tool_permissions?.denied_tools?.join(', ') || '',
    })
    setInitialized(true)
  }

  const handleSave = async () => {
    setSaveMsg('')
    const req = {
      agentId,
      model_provider: form.model_provider,
      model_id: form.model_id,
      system_prompt: form.system_prompt,
      api_credential: form.api_credential || undefined,
      max_iterations: form.max_iterations || undefined,
      token_budget: form.token_budget || undefined,
      task_timeout_seconds: form.task_timeout_seconds || undefined,
      tool_permissions: {
        allowed_tools: form.allowed_tools ? form.allowed_tools.split(',').map(s => s.trim()).filter(Boolean) : undefined,
        denied_tools: form.denied_tools ? form.denied_tools.split(',').map(s => s.trim()).filter(Boolean) : undefined,
      },
    }
    try {
      if (hasConfig) {
        await updateConfig.mutateAsync(req)
      } else {
        await createConfig.mutateAsync(req)
        setInitialized(true)
      }
      setSaveMsg('Saved')
      setTimeout(() => setSaveMsg(''), 2000)
    } catch (e) {
      setSaveMsg(e instanceof Error ? e.message : 'Save failed')
    }
  }

  const handleDeleteConfig = async () => {
    await deleteConfig.mutateAsync(agentId)
    setConfirmDeleteConfig(false)
    setInitialized(false)
    setForm({ model_provider: '', model_id: '', system_prompt: '', api_credential: '', max_iterations: 0, token_budget: 0, task_timeout_seconds: 0, allowed_tools: '', denied_tools: '' })
  }

  if (!hasConfig && !initialized) {
    return (
      <EmptyState
        title="No configuration"
        description="This agent has no configuration yet."
        action={
          <button onClick={() => setInitialized(true)} className="text-sm text-slate-600 hover:text-slate-800 font-medium">
            Create Config
          </button>
        }
      />
    )
  }

  const inputCls = "w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-500"

  return (
    <div className="bg-white rounded-lg shadow p-6 space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Model Provider</label>
          <input type="text" value={form.model_provider} onChange={e => setForm(f => ({ ...f, model_provider: e.target.value }))} placeholder="claude" className={inputCls} />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Model ID</label>
          <input type="text" value={form.model_id} onChange={e => setForm(f => ({ ...f, model_id: e.target.value }))} placeholder="claude-sonnet-4-6" className={inputCls} />
        </div>
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">API Credential</label>
        <input type="password" value={form.api_credential} onChange={e => setForm(f => ({ ...f, api_credential: e.target.value }))} placeholder={hasConfig ? '(unchanged)' : 'sk-...'} className={inputCls} />
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">System Prompt</label>
        <textarea value={form.system_prompt} onChange={e => setForm(f => ({ ...f, system_prompt: e.target.value }))} rows={8} className={inputCls} />
      </div>
      <div className="grid grid-cols-3 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Max Iterations</label>
          <input type="number" value={form.max_iterations || ''} onChange={e => setForm(f => ({ ...f, max_iterations: parseInt(e.target.value) || 0 }))} className={inputCls} />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Token Budget</label>
          <input type="number" value={form.token_budget || ''} onChange={e => setForm(f => ({ ...f, token_budget: parseInt(e.target.value) || 0 }))} className={inputCls} />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Task Timeout (s)</label>
          <input type="number" value={form.task_timeout_seconds || ''} onChange={e => setForm(f => ({ ...f, task_timeout_seconds: parseInt(e.target.value) || 0 }))} className={inputCls} />
        </div>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Allowed Tools (comma-separated)</label>
          <input type="text" value={form.allowed_tools} onChange={e => setForm(f => ({ ...f, allowed_tools: e.target.value }))} placeholder="http_request, dns_lookup, ..." className={inputCls} />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Denied Tools (comma-separated)</label>
          <input type="text" value={form.denied_tools} onChange={e => setForm(f => ({ ...f, denied_tools: e.target.value }))} placeholder="read_file, ..." className={inputCls} />
        </div>
      </div>
      <div className="flex items-center gap-3 pt-2">
        <button onClick={handleSave} disabled={updateConfig.isPending || createConfig.isPending} className="px-4 py-2 text-sm font-medium text-white bg-slate-800 rounded-md hover:bg-slate-700 disabled:opacity-50">
          {hasConfig ? 'Save' : 'Create Config'}
        </button>
        {hasConfig && (
          <button onClick={() => setConfirmDeleteConfig(true)} className="px-4 py-2 text-sm font-medium text-red-700 border border-red-300 rounded-md hover:bg-red-50">
            Delete Config
          </button>
        )}
        {saveMsg && <span className="text-sm text-green-600">{saveMsg}</span>}
      </div>
      <ConfirmDialog
        open={confirmDeleteConfig}
        title="Delete Config"
        message="Delete the agent configuration? The agent will no longer be able to run."
        confirmLabel="Delete Config"
        onConfirm={handleDeleteConfig}
        onCancel={() => setConfirmDeleteConfig(false)}
      />
    </div>
  )
}

function AttributesTab({ agentId }: { agentId: string }) {
  const { data: attributes, isLoading } = useAttributes(agentId)
  const setAttribute = useSetAttribute()
  const deleteAttribute = useDeleteAttribute()
  const [showAdd, setShowAdd] = useState(false)
  const [ns, setNs] = useState('')
  const [val, setVal] = useState('{}')
  const [confirmDeleteNs, setConfirmDeleteNs] = useState<string | null>(null)

  if (isLoading) return <LoadingSpinner />

  const handleAdd = async () => {
    if (!ns.trim()) return
    try {
      const parsed = JSON.parse(val)
      await setAttribute.mutateAsync({ agentId, namespace: ns.trim(), value: parsed })
      setNs('')
      setVal('{}')
      setShowAdd(false)
    } catch { /* invalid JSON */ }
  }

  const handleDelete = async (namespace: string) => {
    await deleteAttribute.mutateAsync({ agentId, namespace })
    setConfirmDeleteNs(null)
  }

  const inputCls = "w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-500"

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <button onClick={() => setShowAdd(true)} className="px-4 py-2 text-sm font-medium text-white bg-slate-800 rounded-md hover:bg-slate-700">
          Add Attribute
        </button>
      </div>

      {showAdd && (
        <div className="bg-white rounded-lg shadow p-4 space-y-3">
          <input type="text" value={ns} onChange={e => setNs(e.target.value)} placeholder="Namespace" className={inputCls} />
          <textarea value={val} onChange={e => setVal(e.target.value)} rows={4} placeholder='{"key": "value"}' className={inputCls + " font-mono"} />
          <div className="flex gap-2">
            <button onClick={handleAdd} disabled={setAttribute.isPending} className="px-4 py-2 text-sm font-medium text-white bg-slate-800 rounded-md hover:bg-slate-700 disabled:opacity-50">Save</button>
            <button onClick={() => setShowAdd(false)} className="px-4 py-2 text-sm font-medium text-gray-700 border border-gray-300 rounded-md hover:bg-gray-50">Cancel</button>
          </div>
        </div>
      )}

      {!attributes || attributes.length === 0 ? (
        <EmptyState title="No attributes" description="External systems can write attributes to this agent." />
      ) : (
        <div className="space-y-3">
          {attributes.map(attr => (
            <div key={attr.namespace} className="bg-white rounded-lg shadow p-4">
              <div className="flex items-center justify-between mb-2">
                <h4 className="text-sm font-semibold text-gray-900">{attr.namespace}</h4>
                <button onClick={() => setConfirmDeleteNs(attr.namespace)} className="text-xs text-red-600 hover:text-red-800">Delete</button>
              </div>
              <pre className="text-xs bg-gray-50 rounded p-3 overflow-x-auto">{JSON.stringify(attr.value, null, 2)}</pre>
              <p className="text-xs text-gray-400 mt-2">Updated {new Date(attr.updated_at).toLocaleString()}</p>
            </div>
          ))}
        </div>
      )}

      <ConfirmDialog
        open={!!confirmDeleteNs}
        title="Delete Attribute"
        message={`Delete attribute namespace "${confirmDeleteNs}"?`}
        confirmLabel="Delete"
        onConfirm={() => confirmDeleteNs && handleDelete(confirmDeleteNs)}
        onCancel={() => setConfirmDeleteNs(null)}
      />
    </div>
  )
}

const LOG_TYPES = [
  'llm.text',
  'llm.tool_use',
  'llm.tool_result',
  'llm.request.start',
  'llm.request.end',
  'llm.loop.start',
  'llm.loop.end',
  'llm.error',
] as const

const TYPE_COLORS: Record<string, string> = {
  'llm.request.start': 'text-blue-400',
  'llm.request.end': 'text-blue-300',
  'llm.text': 'text-green-400',
  'llm.tool_use': 'text-yellow-400',
  'llm.tool_result': 'text-yellow-300',
  'llm.error': 'text-red-400',
  'llm.loop.start': 'text-purple-400',
  'llm.loop.end': 'text-purple-300',
}

function formatLogSummary(parsed: Record<string, unknown>): string {
  switch (parsed.type) {
    case 'llm.text':
      return (parsed.text as string) || ''
    case 'llm.tool_use':
      return `${parsed.tool_name} (${parsed.tool_id})`
    case 'llm.tool_result':
      return `${parsed.tool_name}: ${(parsed.result as string) || (parsed.is_error ? 'ERROR' : 'ok')}`
    case 'llm.request.start':
      return `iter=${parsed.iteration} model=${parsed.model} msgs=${parsed.message_count}`
    case 'llm.request.end':
      return `stop=${parsed.stop_reason} in=${parsed.input_tokens} out=${parsed.output_tokens}`
    case 'llm.loop.start':
      return `model=${parsed.model} max_iter=${parsed.iterations}`
    case 'llm.loop.end':
      return `iters=${parsed.iterations} stop=${parsed.stop_reason}`
    case 'llm.error':
      return (parsed.error as string) || ''
    default:
      return JSON.stringify(parsed)
  }
}

function LogsTab({ agentId }: { agentId: string }) {
  const { data: history, isLoading } = useLogs(agentId)
  const { entries: liveEntries, connected, start, stop, clear } = useLogStream(agentId)
  const [streaming, setStreaming] = useState(false)
  const [expandedRows, setExpandedRows] = useState<Set<number>>(new Set())
  const [typeFilter, setTypeFilter] = useState<Set<string>>(new Set())
  const logsEndRef = useRef<HTMLDivElement>(null)

  const allEntries = [...(history || []), ...liveEntries]

  useEffect(() => {
    logsEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [allEntries.length])

  const toggleStream = () => {
    if (streaming) {
      stop()
      setStreaming(false)
    } else {
      start()
      setStreaming(true)
    }
  }

  const toggleExpand = (idx: number) => {
    setExpandedRows(prev => {
      const next = new Set(prev)
      if (next.has(idx)) next.delete(idx)
      else next.add(idx)
      return next
    })
  }

  const toggleTypeFilter = (type: string) => {
    setTypeFilter(prev => {
      const next = new Set(prev)
      if (next.has(type)) next.delete(type)
      else next.add(type)
      return next
    })
  }

  const parsedEntries = allEntries.map((raw, i) => {
    try {
      return { idx: i, raw, parsed: JSON.parse(raw) as Record<string, unknown> }
    } catch {
      return { idx: i, raw, parsed: null }
    }
  })

  const filteredEntries = typeFilter.size === 0
    ? parsedEntries
    : parsedEntries.filter(e => e.parsed && typeFilter.has(e.parsed.type as string))

  if (isLoading) return <LoadingSpinner />

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-3">
        <button
          onClick={toggleStream}
          className={`px-4 py-2 text-sm font-medium rounded-md ${
            streaming
              ? 'text-red-700 border border-red-300 hover:bg-red-50'
              : 'text-white bg-slate-800 hover:bg-slate-700'
          }`}
        >
          {streaming ? 'Stop Stream' : 'Start Live Stream'}
        </button>
        {streaming && (
          <span className="inline-flex items-center gap-1.5 text-sm">
            <span className={`h-2 w-2 rounded-full ${connected ? 'bg-green-500 animate-pulse' : 'bg-gray-400'}`} />
            <span className="text-gray-500">{connected ? 'Connected' : 'Connecting...'}</span>
          </span>
        )}
        {liveEntries.length > 0 && (
          <button onClick={clear} className="text-sm text-gray-500 hover:text-gray-700">
            Clear live entries
          </button>
        )}
        <span className="text-xs text-gray-400 ml-auto">
          {filteredEntries.length}{typeFilter.size > 0 ? ` / ${allEntries.length}` : ''} entries
        </span>
      </div>

      {/* Type filter chips */}
      <div className="flex flex-wrap gap-1.5">
        {LOG_TYPES.map(type => {
          const active = typeFilter.has(type)
          const color = TYPE_COLORS[type] || 'text-gray-400'
          return (
            <button
              key={type}
              onClick={() => toggleTypeFilter(type)}
              className={`px-2.5 py-1 text-xs rounded-full font-mono transition-colors ${
                active
                  ? 'bg-slate-700 text-white'
                  : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
              }`}
            >
              <span className={active ? 'text-white' : color}>{type.replace('llm.', '')}</span>
            </button>
          )
        })}
        {typeFilter.size > 0 && (
          <button
            onClick={() => setTypeFilter(new Set())}
            className="px-2.5 py-1 text-xs rounded-full text-gray-500 hover:text-gray-700"
          >
            clear filters
          </button>
        )}
      </div>

      <div className="bg-gray-900 rounded-lg overflow-hidden">
        <div className="overflow-y-auto max-h-[600px] p-4 font-mono text-xs space-y-0.5">
          {filteredEntries.length === 0 ? (
            <p className="text-gray-500 text-center py-8">No log entries. Start the agent or enable live streaming.</p>
          ) : (
            filteredEntries.map(({ idx, raw, parsed }) => {
              if (!parsed) {
                return <div key={idx} className="text-gray-300 py-0.5 px-1">{raw}</div>
              }

              const typeColor = TYPE_COLORS[parsed.type as string] || 'text-gray-400'
              const expanded = expandedRows.has(idx)
              const summary = formatLogSummary(parsed)

              return (
                <div key={idx} className="hover:bg-gray-800 rounded">
                  <div
                    className="flex gap-2 leading-relaxed px-1 py-0.5 cursor-pointer"
                    onClick={() => toggleExpand(idx)}
                  >
                    <span className="text-gray-600 select-none">{expanded ? '▼' : '▶'}</span>
                    <span className="text-gray-600 shrink-0">
                      {parsed.timestamp ? new Date(parsed.timestamp as string).toLocaleTimeString() : ''}
                    </span>
                    <span className={`shrink-0 ${typeColor}`}>{parsed.type as string}</span>
                    {!expanded && (
                      <span className="text-gray-400 truncate">{summary}</span>
                    )}
                  </div>
                  {expanded && (
                    <div className="pl-8 pr-2 pb-2">
                      <pre className="text-gray-300 whitespace-pre-wrap break-words bg-gray-950 rounded p-2 max-h-96 overflow-y-auto">
                        {JSON.stringify(parsed, null, 2)}
                      </pre>
                    </div>
                  )}
                </div>
              )
            })
          )}
          <div ref={logsEndRef} />
        </div>
      </div>
    </div>
  )
}

const TASK_STATUS_COLORS: Record<TaskStatus, string> = {
  pending: 'bg-yellow-100 text-yellow-700',
  assigned: 'bg-blue-100 text-blue-700',
  running: 'bg-indigo-100 text-indigo-700',
  completed: 'bg-green-100 text-green-700',
  failed: 'bg-red-100 text-red-700',
}

function TasksTab({ agentId }: { agentId: string }) {
  const { data: tasks, isLoading } = useAgentTasks(agentId)
  const [expandedId, setExpandedId] = useState<string | null>(null)

  if (isLoading) return <LoadingSpinner />

  if (!tasks || tasks.length === 0) {
    return <EmptyState title="No tasks" description="Tasks appear here when triggered by events or created manually." />
  }

  return (
    <div className="space-y-2">
      <p className="text-xs text-gray-500 mb-3">{tasks.length} tasks (auto-refreshes)</p>
      {tasks.map(task => {
        const expanded = expandedId === task.id
        return (
          <div key={task.id} className="bg-white rounded-lg shadow">
            <div
              className="px-4 py-3 flex items-center gap-3 cursor-pointer hover:bg-gray-50"
              onClick={() => setExpandedId(expanded ? null : task.id)}
            >
              <span className="text-gray-400 text-xs">{expanded ? '▼' : '▶'}</span>
              <span className={`text-xs px-2 py-0.5 rounded-full font-medium shrink-0 ${TASK_STATUS_COLORS[task.status]}`}>
                {task.status}
              </span>
              <span className="text-sm text-gray-900 truncate flex-1">{task.prompt}</span>
              <span className="text-xs text-gray-400 shrink-0">{new Date(task.created_at).toLocaleString()}</span>
            </div>
            {expanded && (
              <div className="px-4 pb-4 border-t border-gray-100 pt-3 space-y-3">
                <div>
                  <p className="text-xs font-medium text-gray-500 mb-1">Prompt</p>
                  <pre className="text-sm bg-gray-50 rounded p-3 whitespace-pre-wrap break-words max-h-48 overflow-y-auto">{task.prompt}</pre>
                </div>
                {task.result && (
                  <div>
                    <p className="text-xs font-medium text-gray-500 mb-1">Result</p>
                    <pre className={`text-sm rounded p-3 whitespace-pre-wrap break-words max-h-64 overflow-y-auto ${
                      task.status === 'failed' ? 'bg-red-50 text-red-900' : 'bg-green-50 text-green-900'
                    }`}>{task.result}</pre>
                  </div>
                )}
                <div className="text-xs text-gray-400">
                  ID: {task.id}
                  {task.trigger_id && <> | Trigger: {task.trigger_id}</>}
                  {task.completed_at && <> | Completed: {new Date(task.completed_at).toLocaleString()}</>}
                </div>
              </div>
            )}
          </div>
        )
      })}
    </div>
  )
}
