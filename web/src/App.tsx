import { Routes, Route } from 'react-router-dom'
import AppShell from './components/layout/AppShell'
import DashboardPage from './pages/DashboardPage'
import AgentListPage from './pages/AgentListPage'
import AgentDetailPage from './pages/AgentDetailPage'
import WebhookListPage from './pages/WebhookListPage'
import WebhookDetailPage from './pages/WebhookDetailPage'
import SourceListPage from './pages/SourceListPage'
import SourceDetailPage from './pages/SourceDetailPage'
import TaskListPage from './pages/TaskListPage'
import TaskDetailPage from './pages/TaskDetailPage'
import NotFoundPage from './pages/NotFoundPage'

function App() {
  return (
    <AppShell>
      <Routes>
        <Route path="/" element={<DashboardPage />} />
        <Route path="/agents" element={<AgentListPage />} />
        <Route path="/agents/:id/*" element={<AgentDetailPage />} />
        <Route path="/webhooks" element={<WebhookListPage />} />
        <Route path="/webhooks/:id" element={<WebhookDetailPage />} />
        <Route path="/sources" element={<SourceListPage />} />
        <Route path="/sources/:id" element={<SourceDetailPage />} />
        <Route path="/tasks" element={<TaskListPage />} />
        <Route path="/tasks/:id" element={<TaskDetailPage />} />
        <Route path="*" element={<NotFoundPage />} />
      </Routes>
    </AppShell>
  )
}

export default App
