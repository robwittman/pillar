interface Props {
  data: Record<string, string> | null | undefined
  label?: string
}

export default function KeyValueDisplay({ data, label }: Props) {
  const entries = data ? Object.entries(data) : []
  return (
    <div>
      {label && <h4 className="text-sm font-medium text-gray-500 mb-1">{label}</h4>}
      {entries.length === 0 ? (
        <p className="text-sm text-gray-400">None</p>
      ) : (
        <div className="flex flex-wrap gap-1.5">
          {entries.map(([k, v]) => (
            <span key={k} className="inline-flex items-center rounded bg-gray-100 px-2 py-0.5 text-xs text-gray-700">
              <span className="font-medium">{k}</span>
              {v && <span className="ml-1 text-gray-500">= {v}</span>}
            </span>
          ))}
        </div>
      )}
    </div>
  )
}
