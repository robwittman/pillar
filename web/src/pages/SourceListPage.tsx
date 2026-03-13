import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useSources, useCreateSource, useDeleteSource } from '../hooks/useSources'
import LoadingSpinner from '../components/shared/LoadingSpinner'
import ErrorAlert from '../components/shared/ErrorAlert'
import EmptyState from '../components/shared/EmptyState'
import ConfirmDialog from '../components/shared/ConfirmDialog'

export default function SourceListPage() {
  const { data: sources, isLoading, error } = useSources()
  const createSource = useCreateSource()
  const deleteSource = useDeleteSource()
  const navigate = useNavigate()
  const [showCreate, setShowCreate] = useState(false)
  const [name, setName] = useState('')
  const [newSecret, setNewSecret] = useState<string | null>(null)
  const [confirmDelete, setConfirmDelete] = useState<string | null>(null)

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorAlert message="Failed to load sources" />

  const handleCreate = async () => {
    if (!name.trim()) return
    const result = await createSource.mutateAsync(name.trim())
    setNewSecret(result.secret)
    setName('')
    setShowCreate(false)
  }

  const handleDelete = async (id: string) => {
    await deleteSource.mutateAsync(id)
    setConfirmDelete(null)
  }

  const inputCls = "w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-500"

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Sources</h1>
        <button
          onClick={() => setShowCreate(true)}
          className="px-4 py-2 text-sm font-medium text-white bg-slate-800 rounded-md hover:bg-slate-700"
        >
          Create Source
        </button>
      </div>

      {newSecret && (
        <div className="mb-4 p-4 bg-green-50 border border-green-200 rounded-lg">
          <p className="text-sm font-medium text-green-800 mb-1">Source created! Save this secret (shown once):</p>
          <code className="text-sm font-mono bg-green-100 px-2 py-1 rounded break-all">{newSecret}</code>
          <button onClick={() => setNewSecret(null)} className="ml-3 text-xs text-green-600 hover:text-green-800">Dismiss</button>
        </div>
      )}

      {showCreate && (
        <div className="mb-4 bg-white rounded-lg shadow p-4 flex gap-3 items-end">
          <div className="flex-1">
            <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
            <input type="text" value={name} onChange={e => setName(e.target.value)} placeholder="e.g. Gitea" className={inputCls} onKeyDown={e => e.key === 'Enter' && handleCreate()} />
          </div>
          <button onClick={handleCreate} disabled={createSource.isPending} className="px-4 py-2 text-sm font-medium text-white bg-slate-800 rounded-md hover:bg-slate-700 disabled:opacity-50">Create</button>
          <button onClick={() => setShowCreate(false)} className="px-4 py-2 text-sm font-medium text-gray-700 border border-gray-300 rounded-md hover:bg-gray-50">Cancel</button>
        </div>
      )}

      {!sources || sources.length === 0 ? (
        <EmptyState title="No sources" description="Create an inbound webhook source to start receiving events." />
      ) : (
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">ID</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Created</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {sources.map(source => (
                <tr key={source.id} className="hover:bg-gray-50 cursor-pointer" onClick={() => navigate(`/sources/${source.id}`)}>
                  <td className="px-6 py-4 text-sm font-medium text-gray-900">{source.name}</td>
                  <td className="px-6 py-4 text-sm font-mono text-gray-500">{source.id.substring(0, 8)}...</td>
                  <td className="px-6 py-4 text-sm text-gray-500">{new Date(source.created_at).toLocaleDateString()}</td>
                  <td className="px-6 py-4 text-right">
                    <button
                      onClick={e => { e.stopPropagation(); setConfirmDelete(source.id) }}
                      className="text-xs text-red-600 hover:text-red-800"
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <ConfirmDialog
        open={!!confirmDelete}
        title="Delete Source"
        message="Delete this source and all its triggers? This cannot be undone."
        confirmLabel="Delete"
        onConfirm={() => confirmDelete && handleDelete(confirmDelete)}
        onCancel={() => setConfirmDelete(null)}
      />
    </div>
  )
}
