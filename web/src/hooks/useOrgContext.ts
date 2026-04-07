import { createContext, useContext } from 'react'
import type { Organization } from '../api/organizations'

export interface OrgContextValue {
  orgs: Organization[]
  activeOrg: Organization | null
  setActiveOrg: (org: Organization) => void
  isLoading: boolean
}

export const OrgContext = createContext<OrgContextValue | null>(null)

export function useOrgContext(): OrgContextValue {
  const ctx = useContext(OrgContext)
  if (!ctx) throw new Error('useOrgContext must be used within OrgProvider')
  return ctx
}
