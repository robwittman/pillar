import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTasks, useCreateTask } from '../hooks/useTasks'
import { useAgents } from '../hooks/useAgents'
import type { TaskStatus } from '../api/types'
import LoadingSpinner from '../components/shared/LoadingSpinner'
import ErrorAlert from '../components/shared/ErrorAlert'
import EmptyState from '../components/shared/EmptyState'

const STATUS_COLORS: Record<TaskStatus, string> = {
  pending: 'bg-yellow-100 text-yellow-700',
  assigned: 'bg-blue-100 text-blue-700',
  running: 'bg-indigo-100 text-indigo-700',
  completed: 'bg-green-100 text-green-700',
  failed: 'bg-red-100 text-red-700',
}

export default function TaskListPage() {
  const [statusFilter, setStatusFilter] = useState<string>('')
  const [agentFilter, setAgentFilter] = useState<string>('')
  const { data: tasks, isLoading, error } = useTasks()
  const { data: agents } = useAgents()
  const createTask = useCreateTask()
  const navigate = useNavigate()
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState({ agent_id: '', prompt: '' })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorAlert message="Failed to load tasks" />

  const filtered = (tasks || []).filter(t => {
    if (statusFilter && t.status !== statusFilter) return false
    if (agentFilter && t.agent_id !== agentFilter) return false
    return true
  })

  const handleCreate = async () => {
    if (!form.agent_id || !form.prompt.trim()) return
    await createTask.mutateAsync({ agent_id: form.agent_id, prompt: form.prompt.trim() })
    setForm({ agent_id: '', prompt: '' })
    setShowCreate(false)
  }

  const inputCls = "rounded-md border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-500"

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Tasks</h1>
        <button onClick={() => setShowCreate(true)} className="px-4 py-2 text-sm font-medium text-white bg-slate-800 rounded-md hover:bg-slate-700">
          Create Task
        </button>
      </div>

      {showCreate && (
        <div className="mb-4 bg-white rounded-lg shadow p-4 space-y-3">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Agent</label>
            <select value={form.agent_id} onChange={e => setForm(f => ({ ...f, agent_id: e.target.value }))} className={inputCls + " w-full"}>
              <option value="">Select agent...</option>
              {agents?.map(a => (
                <option key={a.id} value={a.id}>{a.name}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Prompt</label>
            <textarea
              value={form.prompt}
              onChange={e => setForm(f => ({ ...f, prompt: e.target.value }))}
              rows={4}
              placeholder="What should the agent do?"
              className={inputCls + " w-full"}
            />
          </div>
          <div className="flex gap-2">
            <button onClick={handleCreate} disabled={createTask.isPending} className="px-4 py-2 text-sm font-medium text-white bg-slate-800 rounded-md hover:bg-slate-700 disabled:opacity-50">Create Task</button>
            <button onClick={() => setShowCreate(false)} className="px-4 py-2 text-sm font-medium text-gray-700 border border-gray-300 rounded-md hover:bg-gray-50">Cancel</button>
          </div>
        </div>
      )}

      {/* Filters */}
      <div className="flex gap-3 mb-4">
        <select value={statusFilter} onChange={e => setStatusFilter(e.target.value)} className={inputCls}>
          <option value="">All statuses</option>
          <option value="pending">Pending</option>
          <option value="assigned">Assigned</option>
          <option value="running">Running</option>
          <option value="completed">Completed</option>
          <option value="failed">Failed</option>
        </select>
        <select value={agentFilter} onChange={e => setAgentFilter(e.target.value)} className={inputCls}>
          <option value="">All agents</option>
          {agents?.map(a => (
            <option key={a.id} value={a.id}>{a.name}</option>
          ))}
        </select>
        <span className="text-sm text-gray-500 self-center ml-auto">{filtered.length} tasks</span>
      </div>

      {filtered.length === 0 ? (
        <EmptyState title="No tasks" description={tasks && tasks.length > 0 ? "No tasks match your filters." : "Tasks are created when triggers fire or manually via the API."} />
      ) : (
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Agent</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Prompt</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {filtered.map(task => (
                <tr key={task.id} className="hover:bg-gray-50 cursor-pointer" onClick={() => navigate(`/tasks/${task.id}`)}>
                  <td className="px-6 py-4">
                    <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${STATUS_COLORS[task.status]}`}>
                      {task.status}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-900">
                    {agents?.find(a => a.id === task.agent_id)?.name || task.agent_id.substring(0, 8)}
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-500 max-w-md truncate">{task.prompt}</td>
                  <td className="px-6 py-4 text-sm text-gray-500">{new Date(task.created_at).toLocaleString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
