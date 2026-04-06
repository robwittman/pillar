import { NavLink } from 'react-router-dom'
import { useAuth } from '../../hooks/useAuth'

const links = [
  { to: '/', label: 'Dashboard' },
  { to: '/agents', label: 'Agents' },
  { to: '/sources', label: 'Sources' },
  { to: '/tasks', label: 'Tasks' },
  { to: '/webhooks', label: 'Webhooks' },
  { to: '/settings', label: 'Settings' },
]

export default function Sidebar() {
  const { user, authEnabled, logout } = useAuth()

  return (
    <aside className="w-64 bg-slate-900 text-white flex flex-col">
      <div className="px-6 py-5 border-b border-slate-700">
        <h1 className="text-xl font-bold tracking-tight">Pillar</h1>
        <p className="text-xs text-slate-400 mt-0.5">Agent Management</p>
      </div>
      <nav className="flex-1 px-3 py-4 space-y-1">
        {links.map(({ to, label }) => (
          <NavLink
            key={to}
            to={to}
            end={to === '/'}
            className={({ isActive }) =>
              `block px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                isActive
                  ? 'bg-slate-700 text-white'
                  : 'text-slate-300 hover:bg-slate-800 hover:text-white'
              }`
            }
          >
            {label}
          </NavLink>
        ))}
      </nav>
      {authEnabled && user && (
        <div className="px-3 py-4 border-t border-slate-700">
          <div className="px-3 mb-2">
            <p className="text-sm text-white truncate">{user.display_name || user.email}</p>
            {user.email && <p className="text-xs text-slate-400 truncate">{user.email}</p>}
          </div>
          <button
            onClick={() => logout()}
            className="block w-full px-3 py-2 text-left rounded-md text-sm font-medium text-slate-300 hover:bg-slate-800 hover:text-white transition-colors"
          >
            Sign out
          </button>
        </div>
      )}
    </aside>
  )
}
