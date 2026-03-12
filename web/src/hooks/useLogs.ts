import { useQuery } from '@tanstack/react-query'
import { useEffect, useRef, useState, useCallback } from 'react'
import { logsApi } from '../api/logs'

export function useLogs(agentId: string, limit = 200) {
  return useQuery({
    queryKey: ['agents', agentId, 'logs'],
    queryFn: () => logsApi.get(agentId, undefined, limit),
    enabled: !!agentId,
  })
}

export function useLogStream(agentId: string) {
  const [entries, setEntries] = useState<string[]>([])
  const [connected, setConnected] = useState(false)
  const esRef = useRef<EventSource | null>(null)

  const start = useCallback(() => {
    if (esRef.current) return
    const es = logsApi.stream(
      agentId,
      (entry) => setEntries(prev => [...prev.slice(-999), entry]),
      () => setConnected(false),
    )
    es.onopen = () => setConnected(true)
    esRef.current = es
  }, [agentId])

  const stop = useCallback(() => {
    esRef.current?.close()
    esRef.current = null
    setConnected(false)
  }, [])

  const clear = useCallback(() => setEntries([]), [])

  useEffect(() => {
    return () => {
      esRef.current?.close()
      esRef.current = null
    }
  }, [agentId])

  return { entries, connected, start, stop, clear }
}
