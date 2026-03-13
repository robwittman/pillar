import { useParams, Link } from 'react-router-dom'
import { useTask } from '../hooks/useTasks'
import { useAgent } from '../hooks/useAgents'
import type { TaskStatus } from '../api/types'
import LoadingSpinner from '../components/shared/LoadingSpinner'
import ErrorAlert from '../components/shared/ErrorAlert'

const STATUS_COLORS: Record<TaskStatus, string> = {
  pending: 'bg-yellow-100 text-yellow-700',
  assigned: 'bg-blue-100 text-blue-700',
  running: 'bg-indigo-100 text-indigo-700',
  completed: 'bg-green-100 text-green-700',
  failed: 'bg-red-100 text-red-700',
}

export default function TaskDetailPage() {
  const { id } = useParams<{ id: string }>()
  const { data: task, isLoading, error } = useTask(id!)
  const { data: agent } = useAgent(task?.agent_id || '')

  if (isLoading) return <LoadingSpinner />
  if (error || !task) return <ErrorAlert message="Task not found" />

  const isTerminal = task.status === 'completed' || task.status === 'failed'

  return (
    <div>
      <div className="flex items-center gap-3 mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Task</h1>
        <span className={`text-sm px-2.5 py-0.5 rounded-full font-medium ${STATUS_COLORS[task.status]}`}>
          {task.status}
        </span>
      </div>

      <div className="space-y-6">
        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-sm font-semibold text-gray-900 mb-4">Details</h3>
          <dl className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <dt className="text-gray-500">ID</dt>
              <dd className="mt-1 font-mono text-gray-900">{task.id}</dd>
            </div>
            <div>
              <dt className="text-gray-500">Agent</dt>
              <dd className="mt-1">
                <Link to={`/agents/${task.agent_id}`} className="text-slate-600 hover:text-slate-800 font-medium">
                  {agent?.name || task.agent_id}
                </Link>
              </dd>
            </div>
            {task.trigger_id && (
              <div>
                <dt className="text-gray-500">Trigger</dt>
                <dd className="mt-1 font-mono text-gray-900">{task.trigger_id}</dd>
              </div>
            )}
            <div>
              <dt className="text-gray-500">Created</dt>
              <dd className="mt-1 text-gray-900">{new Date(task.created_at).toLocaleString()}</dd>
            </div>
            {task.completed_at && (
              <div>
                <dt className="text-gray-500">Completed</dt>
                <dd className="mt-1 text-gray-900">{new Date(task.completed_at).toLocaleString()}</dd>
              </div>
            )}
          </dl>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-sm font-semibold text-gray-900 mb-3">Prompt</h3>
          <pre className="text-sm bg-gray-50 rounded p-4 whitespace-pre-wrap break-words max-h-64 overflow-y-auto">{task.prompt}</pre>
        </div>

        {isTerminal && task.result && (
          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-sm font-semibold text-gray-900 mb-3">Result</h3>
            <pre className={`text-sm rounded p-4 whitespace-pre-wrap break-words max-h-96 overflow-y-auto ${
              task.status === 'failed' ? 'bg-red-50 text-red-900' : 'bg-green-50 text-green-900'
            }`}>{task.result}</pre>
          </div>
        )}

        {task.context != null && (
          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-sm font-semibold text-gray-900 mb-3">Event Context</h3>
            <pre className="text-xs bg-gray-50 rounded p-4 whitespace-pre-wrap break-words max-h-64 overflow-y-auto font-mono">
              {JSON.stringify(task.context, null, 2)}
            </pre>
          </div>
        )}

        {!isTerminal && (
          <div className="text-center text-sm text-gray-500 py-4">
            <span className="inline-flex items-center gap-2">
              <span className="h-2 w-2 bg-blue-500 rounded-full animate-pulse" />
              Task is {task.status}... this page auto-refreshes.
            </span>
          </div>
        )}
      </div>
    </div>
  )
}
