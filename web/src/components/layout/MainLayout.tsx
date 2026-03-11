import { Outlet } from 'react-router-dom'
import { Layout, Menu, Dropdown, Avatar, Space, Badge, Typography } from 'antd'
import {
  DashboardOutlined,
  DesktopOutlined,
  OrderedListOutlined,
  AlertOutlined,
  UserOutlined,
  LogoutOutlined,
  ApiOutlined,
} from '@ant-design/icons'
import { useNavigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/stores/auth'
import { useWebSocketStore } from '@/stores/websocket'
import styles from './MainLayout.module.css'

const { Header, Sider, Content } = Layout
const { Text } = Typography

const menuItems = [
  {
    key: '/',
    icon: <DashboardOutlined />,
    label: '仪表盘',
  },
  {
    key: '/agents',
    icon: <DesktopOutlined />,
    label: 'Agent 管理',
  },
  {
    key: '/tasks',
    icon: <OrderedListOutlined />,
    label: '任务管理',
  },
  {
    key: '/alerts',
    icon: <AlertOutlined />,
    label: '告警管理',
  },
]

export default function MainLayout() {
  const navigate = useNavigate()
  const location = useLocation()
  const user = useAuthStore((state) => state.user)
  const logout = useAuthStore((state) => state.logout)
  const isConnected = useWebSocketStore((state) => state.isConnected)

  const handleMenuClick = (key: string) => {
    navigate(key)
  }

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人信息',
    },
    {
      type: 'divider' as const,
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      danger: true,
    },
  ]

  return (
    <Layout className={styles.layout}>
      <Sider width={220} className={styles.sider}>
        <div className={styles.logo}>
          <div className={styles.logoIcon}>
            <ApiOutlined style={{ color: '#fff', fontSize: 18 }} />
          </div>
          <Text className={styles.logoText}>AgentTeams</Text>
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={({ key }) => handleMenuClick(key)}
        />
      </Sider>
      <Layout>
        <Header className={styles.header}>
          <div className={styles.headerRight}>
            <div className={styles.connectionStatus}>
              <Badge
                status={isConnected ? 'success' : 'error'}
              />
              <Text type="secondary">
                {isConnected ? '已连接' : '未连接'}
              </Text>
            </div>
            <Dropdown
              menu={{
                items: userMenuItems,
                onClick: ({ key }) => {
                  if (key === 'logout') {
                    handleLogout()
                  }
                },
              }}
              trigger={['click']}
            >
              <Space className={styles.userDropdown}>
                <Avatar
                  icon={<UserOutlined />}
                  style={{
                    background: 'linear-gradient(135deg, #1677ff 0%, #722ed1 100%)',
                  }}
                />
                <Text>{user?.username || '用户'}</Text>
              </Space>
            </Dropdown>
          </div>
        </Header>
        <Content className={styles.content}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
