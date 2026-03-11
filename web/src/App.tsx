import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { ConfigProvider, theme } from 'antd'
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

// Custom dark theme configuration
const customTheme = {
  algorithm: theme.darkAlgorithm,
  token: {
    // Primary color - Electric Blue
    colorPrimary: '#1677ff',
    colorSuccess: '#52c41a',
    colorWarning: '#faad14',
    colorError: '#ff4d4f',
    colorInfo: '#1677ff',

    // Background colors
    colorBgContainer: '#1a1a24',
    colorBgElevated: '#24242f',
    colorBgLayout: '#0a0a0f',
    colorBgSpotlight: '#2e2e3a',
    colorBgMask: 'rgba(0, 0, 0, 0.6)',

    // Text colors
    colorText: 'rgba(255, 255, 255, 0.88)',
    colorTextSecondary: 'rgba(255, 255, 255, 0.65)',
    colorTextTertiary: 'rgba(255, 255, 255, 0.45)',
    colorTextQuaternary: 'rgba(255, 255, 255, 0.25)',

    // Border
    colorBorder: 'rgba(255, 255, 255, 0.1)',
    colorBorderSecondary: 'rgba(255, 255, 255, 0.06)',

    // Fill colors
    colorFill: 'rgba(255, 255, 255, 0.08)',
    colorFillSecondary: 'rgba(255, 255, 255, 0.06)',
    colorFillTertiary: 'rgba(255, 255, 255, 0.04)',
    colorFillQuaternary: 'rgba(255, 255, 255, 0.02)',

    // Radius
    borderRadius: 8,
    borderRadiusLG: 12,
    borderRadiusSM: 4,

    // Font
    fontFamily: "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif",
    fontSize: 14,
    fontSizeHeading1: 38,
    fontSizeHeading2: 30,
    fontSizeHeading3: 24,
    fontSizeHeading4: 20,
    fontSizeHeading5: 16,

    // Spacing
    margin: 16,
    marginLG: 24,
    marginMD: 20,
    marginSM: 12,
    marginXL: 32,
    marginXS: 8,
    marginXXS: 4,

    // Control
    controlHeight: 36,
    controlHeightLG: 44,
    controlHeightSM: 28,

    // Shadows
    boxShadow: '0 4px 12px rgba(0, 0, 0, 0.4)',
    boxShadowSecondary: '0 8px 24px rgba(0, 0, 0, 0.5)',

    // Motion
    motionDurationFast: '0.15s',
    motionDurationMid: '0.25s',
    motionDurationSlow: '0.35s',
    motionEaseInOut: 'cubic-bezier(0.4, 0, 0.2, 1)',

    // Link
    colorLink: '#4096ff',
    colorLinkHover: '#69b1ff',
    colorLinkActive: '#1677ff',
  },
  components: {
    Layout: {
      headerBg: '#121218',
      headerColor: 'rgba(255, 255, 255, 0.88)',
      headerHeight: 64,
      headerPadding: '0 24px',
      siderBg: '#0a0a0f',
      bodyBg: '#0a0a0f',
      footerBg: '#0a0a0f',
      triggerBg: '#24242f',
      triggerColor: 'rgba(255, 255, 255, 0.65)',
    },
    Menu: {
      darkItemBg: '#0a0a0f',
      darkItemColor: 'rgba(255, 255, 255, 0.65)',
      darkItemHoverBg: 'rgba(255, 255, 255, 0.06)',
      darkItemSelectedBg: 'rgba(22, 119, 255, 0.15)',
      darkItemActiveBg: 'rgba(22, 119, 255, 0.15)',
      darkSubMenuItemBg: '#121218',
      itemBorderRadius: 8,
      itemMarginBlock: 4,
      itemMarginInline: 8,
      itemPaddingInline: 16,
      iconSize: 18,
    },
    Card: {
      colorBgContainer: '#1a1a24',
      colorBorderSecondary: 'rgba(255, 255, 255, 0.08)',
      paddingLG: 20,
      padding: 16,
      borderRadiusLG: 12,
    },
    Table: {
      headerBg: '#121218',
      headerColor: 'rgba(255, 255, 255, 0.88)',
      rowHoverBg: 'rgba(255, 255, 255, 0.04)',
      rowSelectedBg: 'rgba(22, 119, 255, 0.1)',
      rowSelectedHoverBg: 'rgba(22, 119, 255, 0.15)',
      borderColor: 'rgba(255, 255, 255, 0.06)',
      headerSortActiveBg: '#24242f',
      headerSortHoverBg: '#1a1a24',
      cellPaddingBlock: 12,
      cellPaddingInline: 16,
    },
    Input: {
      colorBgContainer: '#24242f',
      colorBorder: 'rgba(255, 255, 255, 0.12)',
      hoverBorderColor: '#4096ff',
      activeBorderColor: '#1677ff',
      colorBgContainerDisabled: '#1a1a24',
    },
    Select: {
      colorBgContainer: '#24242f',
      colorBorder: 'rgba(255, 255, 255, 0.12)',
      optionSelectedBg: 'rgba(22, 119, 255, 0.15)',
      optionActiveBg: 'rgba(255, 255, 255, 0.06)',
    },
    Button: {
      primaryShadow: 'none',
      defaultShadow: 'none',
      dangerShadow: 'none',
    },
    Tag: {
      defaultBg: 'rgba(255, 255, 255, 0.06)',
      defaultColor: 'rgba(255, 255, 255, 0.88)',
    },
    Progress: {
      remainingColor: 'rgba(255, 255, 255, 0.08)',
    },
    Drawer: {
      colorBgElevated: '#1a1a24',
    },
    Modal: {
      contentBg: '#1a1a24',
      headerBg: '#1a1a24',
      footerBg: '#1a1a24',
    },
    Dropdown: {
      colorBgElevated: '#24242f',
      controlItemBgHover: 'rgba(255, 255, 255, 0.06)',
      controlItemBgActive: 'rgba(22, 119, 255, 0.15)',
    },
    Badge: {
      textFontSize: 11,
      textFontWeight: 500,
    },
    Descriptions: {
      labelBg: '#121218',
    },
  },
}

function App() {
  return (
    <ConfigProvider locale={zhCN} theme={customTheme}>
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
