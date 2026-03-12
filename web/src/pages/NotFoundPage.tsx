import { Link } from 'react-router-dom'

export default function NotFoundPage() {
  return (
    <div className="text-center py-20">
      <h1 className="text-4xl font-bold text-gray-900">404</h1>
      <p className="mt-2 text-gray-500">Page not found.</p>
      <Link to="/" className="mt-4 inline-block text-sm font-medium text-slate-600 hover:text-slate-800">
        Back to Dashboard
      </Link>
    </div>
  )
}
