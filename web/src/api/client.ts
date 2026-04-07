class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

// Active org ID is set by OrgProvider and included on API requests.
let _activeOrgId: string | null = null

export function setActiveOrgId(orgId: string | null) {
  _activeOrgId = orgId
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }

  // Include org context on resource API calls.
  if (_activeOrgId && path.startsWith('/api/')) {
    headers['X-Org-ID'] = _activeOrgId
  }

  const opts: RequestInit = {
    method,
    headers,
    credentials: 'same-origin',
  }
  if (body !== undefined) {
    opts.body = JSON.stringify(body)
  }

  const res = await fetch(path, opts)

  if (!res.ok) {
    // On 401 from API routes, reload to trigger auth check
    if (res.status === 401 && path.startsWith('/api/')) {
      window.location.reload()
      return undefined as T
    }

    let msg = `HTTP ${res.status}`
    try {
      const err = await res.json()
      if (err.error) msg = err.error
    } catch { /* ignore */ }
    throw new ApiError(res.status, msg)
  }

  if (res.status === 204) return undefined as T
  return res.json()
}

export const api = {
  get: <T>(path: string) => request<T>('GET', path),
  post: <T>(path: string, body?: unknown) => request<T>('POST', path, body),
  put: <T>(path: string, body?: unknown) => request<T>('PUT', path, body),
  del: (path: string) => request<void>('DELETE', path),
}
