import type { ReactNode } from 'react'

interface Props {
  title: string
  description: string
  action?: ReactNode
}

export default function EmptyState({ title, description, action }: Props) {
  return (
    <div className="text-center py-12">
      <h3 className="text-sm font-semibold text-gray-900">{title}</h3>
      <p className="mt-1 text-sm text-gray-500">{description}</p>
      {action && <div className="mt-4">{action}</div>}
    </div>
  )
}
