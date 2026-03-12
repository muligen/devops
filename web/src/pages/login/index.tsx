import { useState } from 'react'
import { Form, Input, Button, message, Typography } from 'antd'
import { UserOutlined, LockOutlined, ApiOutlined } from '@ant-design/icons'
import { useNavigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/stores/auth'
import styles from './index.module.css'

const { Title, Text } = Typography

interface LoginForm {
  username: string
  password: string
}

export default function LoginPage() {
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const login = useAuthStore((state) => state.login)

  const from = (location.state as { from?: { pathname: string } })?.from?.pathname || '/'

  const handleSubmit = async (values: LoginForm) => {
    setLoading(true)
    try {
      await login(values.username, values.password)
      message.success('登录成功')
      navigate(from, { replace: true })
    } catch {
      message.error('用户名或密码错误')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className={styles.container}>
      {/* Background effects */}
      <div className={styles.backgroundGrid} />
      <div className={styles.backgroundGlow} />
      <div className={styles.backgroundGlow2} />

      {/* Login card */}
      <div className={styles.loginCard}>
        <div className={styles.cardHeader}>
          <div className={styles.logoWrapper}>
            <div className={styles.logoIcon}>
              <ApiOutlined />
            </div>
          </div>
          <Title level={2} className={styles.title}>
            AgentTeams
          </Title>
          <Text type="secondary" className={styles.subtitle}>
            企业级机器管理系统
          </Text>
        </div>

        <Form
          name="login"
          onFinish={handleSubmit}
          autoComplete="off"
          layout="vertical"
          size="large"
          className={styles.form}
        >
          <Form.Item
            name="username"
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input
              prefix={<UserOutlined style={{ color: 'rgba(255,255,255,0.45)' }} />}
              placeholder="用户名"
              autoComplete="username"
              className={styles.input}
            />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[{ required: true, message: '请输入密码' }]}
          >
            <Input.Password
              prefix={<LockOutlined style={{ color: 'rgba(255,255,255,0.45)' }} />}
              placeholder="密码"
              autoComplete="current-password"
              className={styles.input}
            />
          </Form.Item>

          <Form.Item style={{ marginBottom: 0, marginTop: 8 }}>
            <Button
              type="primary"
              htmlType="submit"
              loading={loading}
              block
              className={styles.submitButton}
            >
              登录
            </Button>
          </Form.Item>
        </Form>

        <div className={styles.footer}>
          <Text type="secondary" className={styles.hint}>
            默认账号: admin / admin123
          </Text>
        </div>
      </div>

      {/* Decorative elements */}
      <div className={styles.floatingOrb1} />
      <div className={styles.floatingOrb2} />
      <div className={styles.floatingOrb3} />
    </div>
  )
}
