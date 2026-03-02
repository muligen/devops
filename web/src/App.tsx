import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import ProtectedRoute from '@/components/common/ProtectedRoute'
import AlertNotificationProvider from '@/components/common/AlertNotificationProvider'
import MainLayout from '@/components/layout/MainLayout'
import LoginPage from '@/pages/login'
import DashboardPage from '@/pages/dashboard'
import AgentsPage from '@/pages/agents'
import AgentDetailPage from '@/pages/agents/[id]'
import TasksPage from '@/pages/tasks'
import AlertsPage from '@/pages/alerts'

function App() {
  return (
    <ConfigProvider
      locale={zhCN}
      theme={{
        token: {
          colorPrimary: '#1890ff',
          borderRadius: 6,
        },
      }}
    >
      <AlertNotificationProvider>
        <BrowserRouter>
          <Routes>
            {/* Public routes */}
            <Route path="/login" element={<LoginPage />} />

            {/* Protected routes */}
            <Route
              path="/"
              element={
                <ProtectedRoute>
                  <MainLayout />
                </ProtectedRoute>
              }
            >
              <Route index element={<DashboardPage />} />
              <Route path="agents" element={<AgentsPage />} />
              <Route path="agents/:id" element={<AgentDetailPage />} />
              <Route path="tasks" element={<TasksPage />} />
              <Route path="alerts" element={<AlertsPage />} />
            </Route>

            {/* Fallback */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </BrowserRouter>
      </AlertNotificationProvider>
    </ConfigProvider>
  )
}

export default App
