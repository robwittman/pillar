const colors: Record<string, string> = {
  running: 'bg-green-100 text-green-800',
  active: 'bg-green-100 text-green-800',
  pending: 'bg-yellow-100 text-yellow-800',
  stopped: 'bg-gray-100 text-gray-800',
  inactive: 'bg-gray-100 text-gray-800',
  error: 'bg-red-100 text-red-800',
  failed: 'bg-red-100 text-red-800',
  delivered: 'bg-green-100 text-green-800',
}

interface Props {
  status: string
  size?: 'sm' | 'md'
}

export default function StatusBadge({ status, size = 'sm' }: Props) {
  const cls = colors[status] || 'bg-gray-100 text-gray-800'
  const pad = size === 'sm' ? 'px-2 py-0.5 text-xs' : 'px-2.5 py-1 text-sm'
  return (
    <span className={`inline-flex items-center rounded-full font-medium ${cls} ${pad}`}>
      {status}
    </span>
  )
}
