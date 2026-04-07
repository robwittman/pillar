import { useState, useEffect, type ReactNode } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { orgsApi, type Organization } from '../api/organizations'
import { OrgContext } from '../hooks/useOrgContext'
import { useAuth } from '../hooks/useAuth'
import { setActiveOrgId } from '../api/client'

export default function OrgProvider({ children }: { children: ReactNode }) {
  const { isAuthenticated, authEnabled } = useAuth()
  const [activeOrg, setActiveOrgState] = useState<Organization | null>(null)

  const { data: orgs = [], isLoading } = useQuery({
    queryKey: ['organizations'],
    queryFn: orgsApi.list,
    enabled: isAuthenticated && authEnabled,
    retry: false,
  })

  // Auto-select org: restore from localStorage, or pick the first one.
  useEffect(() => {
    if (orgs.length === 0) return
    if (activeOrg && orgs.some(o => o.id === activeOrg.id)) return

    const savedId = localStorage.getItem('pillar_active_org')
    const saved = savedId ? orgs.find(o => o.id === savedId) : null
    const selected = saved || orgs[0]
    setActiveOrgState(selected)
    setActiveOrgId(selected.id)
  }, [orgs, activeOrg])

  const qc = useQueryClient()

  const setActiveOrg = (org: Organization) => {
    setActiveOrgState(org)
    setActiveOrgId(org.id)
    localStorage.setItem('pillar_active_org', org.id)
    // Refetch all resource queries since they're now scoped to a different org.
    qc.invalidateQueries()
  }

  // Don't block rendering if auth is disabled.
  if (!authEnabled) {
    return (
      <OrgContext.Provider value={{ orgs: [], activeOrg: null, setActiveOrg, isLoading: false }}>
        {children}
      </OrgContext.Provider>
    )
  }

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-gray-500 text-sm">Loading...</div>
      </div>
    )
  }

  return (
    <OrgContext.Provider value={{ orgs, activeOrg, setActiveOrg, isLoading }}>
      {children}
    </OrgContext.Provider>
  )
}
