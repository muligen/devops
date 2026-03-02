import { useEffect, useState } from 'react'
import { Row, Col, Typography } from 'antd'
import { useWebSocket } from '@/hooks'
import { useDashboardStore } from '@/stores'
import { dashboardApi, agentApi, alertApi } from '@/api'
import StatCards from '@/components/dashboard/StatCards'
import AgentStatusGrid from '@/components/dashboard/AgentStatusGrid'
import TaskTrendChart from '@/components/dashboard/TaskTrendChart'
import RecentAlerts from '@/components/dashboard/RecentAlerts'
import type { Agent, AlertEvent } from '@/types'

const { Title } = Typography

export default function DashboardPage() {
  const [agents, setAgents] = useState<Agent[]>([])
  const [alerts, setAlerts] = useState<AlertEvent[]>([])
  const [loading, setLoading] = useState(true)
  const { stats, setStats } = useDashboardStore()

  // Connect WebSocket
  useWebSocket()

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true)
      try {
        const [statsRes, agentsRes, alertsRes] = await Promise.all([
          dashboardApi.getStats(),
          agentApi.list({ page: 1, page_size: 12, sort: 'cpu_usage', order: 'desc' }),
          alertApi.listEvents({ page: 1, page_size: 5 }),
        ])

        setStats(statsRes)
        setAgents(agentsRes.data)
        setAlerts(alertsRes.data)
      } catch (error) {
        console.error('Failed to fetch dashboard data:', error)
      } finally {
        setLoading(false)
      }
    }

    fetchData()
  }, [setStats])

  return (
    <div>
      <Title level={4} style={{ marginBottom: 24 }}>
        仪表盘
      </Title>

      <StatCards stats={stats} />

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={16}>
          <TaskTrendChart data={stats?.task_trend} loading={loading} />
        </Col>
        <Col xs={24} lg={8}>
          <RecentAlerts alerts={alerts} loading={loading} />
        </Col>
      </Row>

      <AgentStatusGrid agents={agents} loading={loading} />
    </div>
  )
}
