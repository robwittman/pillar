import { Link } from 'react-router-dom'
import { useAgents } from '../hooks/useAgents'
import { useWebhooks } from '../hooks/useWebhooks'
import LoadingSpinner from '../components/shared/LoadingSpinner'

export default function DashboardPage() {
  const { data: agents, isLoading: loadingAgents } = useAgents()
  const { data: webhooks, isLoading: loadingWebhooks } = useWebhooks()

  if (loadingAgents || loadingWebhooks) return <LoadingSpinner />

  const byStatus = (agents || []).reduce<Record<string, number>>((acc, a) => {
    acc[a.status] = (acc[a.status] || 0) + 1
    return acc
  }, {})

  const cards = [
    { label: 'Total Agents', value: agents?.length || 0, color: 'bg-blue-500', to: '/agents' },
    { label: 'Running', value: byStatus['running'] || 0, color: 'bg-green-500', to: '/agents' },
    { label: 'Stopped', value: byStatus['stopped'] || 0, color: 'bg-gray-500', to: '/agents' },
    { label: 'Error', value: byStatus['error'] || 0, color: 'bg-red-500', to: '/agents' },
    { label: 'Pending', value: byStatus['pending'] || 0, color: 'bg-yellow-500', to: '/agents' },
    { label: 'Webhooks', value: webhooks?.length || 0, color: 'bg-purple-500', to: '/webhooks' },
  ]

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Dashboard</h1>
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
        {cards.map(c => (
          <Link key={c.label} to={c.to} className="block rounded-lg bg-white shadow p-4 hover:shadow-md transition-shadow">
            <div className={`h-1 w-8 rounded ${c.color} mb-3`} />
            <p className="text-2xl font-bold text-gray-900">{c.value}</p>
            <p className="text-sm text-gray-500">{c.label}</p>
          </Link>
        ))}
      </div>
    </div>
  )
}
