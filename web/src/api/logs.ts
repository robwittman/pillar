import { api } from './client'

export const logsApi = {
  get: (agentId: string, since?: number, limit?: number) => {
    const params = new URLSearchParams()
    if (since) params.set('since', String(since))
    if (limit) params.set('limit', String(limit))
    const qs = params.toString()
    return api.get<string[]>(`/api/v1/agents/${agentId}/logs${qs ? '?' + qs : ''}`)
  },

  stream: (agentId: string, onEntry: (entry: string) => void, onError?: (err: Event) => void): EventSource => {
    const es = new EventSource(`/api/v1/agents/${agentId}/logs/stream`)
    es.onmessage = (e) => onEntry(e.data)
    es.onerror = (e) => {
      if (onError) onError(e)
    }
    return es
  },
}
