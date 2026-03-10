import { useEffect, useState, useMemo } from 'react'
import { Row, Col, Typography, message } from 'antd'
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
  const { stats, setStats, metrics } = useDashboardStore()

  // Connect WebSocket
  useWebSocket()

  // Merge real-time metrics with agents data
  const agentsWithMetrics = useMemo(() => {
    if (!metrics || Object.keys(metrics).length === 0) {
      return agents
    }

    return agents.map((agent) => {
      const agentMetrics = metrics[agent.id]
      if (agentMetrics) {
        return {
          ...agent,
          cpu_usage: agentMetrics.cpu_usage ?? agent.cpu_usage,
          memory_usage: agentMetrics.memory_usage ?? agent.memory_usage,
          disk_usage: agentMetrics.disk_usage ?? agent.disk_usage,
        }
      }
      return agent
    })
  }, [agents, metrics])

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
        message.error('加载仪表盘数据失败')
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

      <AgentStatusGrid agents={agentsWithMetrics} loading={loading} />
    </div>
  )
}
