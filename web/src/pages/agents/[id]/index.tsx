import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Card, Row, Col, Typography, Tag, Descriptions, Button, Progress, Space, Empty, Spin } from 'antd'
import { ArrowLeftOutlined, ReloadOutlined, PlayCircleOutlined } from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'
import { agentApi } from '@/api'
import { formatRelativeTime, formatDate } from '@/utils'
import type { Agent } from '@/types'
import RecentTasks from '@/components/agents/RecentTasks'
import ExecuteTaskModal from '@/components/tasks/ExecuteTaskModal'

const { Title } = Typography

export default function AgentDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [loading, setLoading] = useState(true)
  const [agent, setAgent] = useState<Agent | null>(null)
  const [timeRange, setTimeRange] = useState('1h')
  const [executeModalOpen, setExecuteModalOpen] = useState(false)

  const fetchAgent = async () => {
    if (!id) return
    setLoading(true)
    try {
      const data = await agentApi.get(id)
      setAgent(data)
      // Fetch metrics history (currently unused, but available for future use)
      await agentApi.getMetrics(id, timeRange)
    } catch (error) {
      console.error('Failed to fetch agent:', error)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchAgent()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id, timeRange])

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 400 }}>
        <Spin size="large" />
      </div>
    )
  }

  if (!agent) {
    return <Empty description="Agent 不存在" />
  }

  const cpuGaugeOption = {
    series: [
      {
        type: 'gauge',
        startAngle: 200,
        endAngle: -20,
        min: 0,
        max: 100,
        splitNumber: 10,
        axisLine: {
          lineStyle: {
            width: 20,
            color: [
              [0.6, '#52c41a'],
              [0.8, '#faad14'],
              [1, '#ff4d4f'],
            ],
          },
        },
        pointer: {
          itemStyle: { color: 'auto' },
        },
        axisTick: { show: false },
        splitLine: { show: false },
        axisLabel: { show: false },
        title: { offsetCenter: [0, '70%'], fontSize: 14 },
        detail: { valueAnimation: true, fontSize: 24, offsetCenter: [0, '0%'] },
        data: [{ value: agent.cpu_usage || 0, name: 'CPU' }],
      },
    ],
  }

  const memoryGaugeOption = {
    series: [
      {
        type: 'gauge',
        startAngle: 200,
        endAngle: -20,
        min: 0,
        max: 100,
        splitNumber: 10,
        axisLine: {
          lineStyle: {
            width: 20,
            color: [
              [0.6, '#52c41a'],
              [0.8, '#faad14'],
              [1, '#ff4d4f'],
            ],
          },
        },
        pointer: {
          itemStyle: { color: 'auto' },
        },
        axisTick: { show: false },
        splitLine: { show: false },
        axisLabel: { show: false },
        title: { offsetCenter: [0, '70%'], fontSize: 14 },
        detail: { valueAnimation: true, fontSize: 24, offsetCenter: [0, '0%'] },
        data: [{ value: agent.memory_usage || 0, name: '内存' }],
      },
    ],
  }

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/agents')}>
          返回列表
        </Button>
      </div>

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={8}>
          <Card
            title={
              <Space>
                <span>{agent.name}</span>
                <Tag color={agent.status === 'online' ? 'success' : 'error'}>
                  {agent.status === 'online' ? '在线' : '离线'}
                </Tag>
              </Space>
            }
            extra={
              <Space>
                <Button icon={<ReloadOutlined />} onClick={fetchAgent}>
                  刷新
                </Button>
                <Button
                  type="primary"
                  icon={<PlayCircleOutlined />}
                  onClick={() => setExecuteModalOpen(true)}
                  disabled={agent.status !== 'online'}
                >
                  执行任务
                </Button>
              </Space>
            }
          >
            <Descriptions column={1} size="small">
              <Descriptions.Item label="主机名">{agent.hostname}</Descriptions.Item>
              <Descriptions.Item label="IP 地址">{agent.ip_address}</Descriptions.Item>
              <Descriptions.Item label="操作系统">{agent.os_info}</Descriptions.Item>
              <Descriptions.Item label="Agent 版本">{agent.version}</Descriptions.Item>
              <Descriptions.Item label="首次上线">{formatDate(agent.created_at)}</Descriptions.Item>
              <Descriptions.Item label="最后心跳">
                {formatRelativeTime(agent.last_seen_at)}
              </Descriptions.Item>
            </Descriptions>
          </Card>
        </Col>

        <Col xs={24} lg={16}>
          <Card title="实时状态">
            <Row gutter={24}>
              <Col span={8}>
                <ReactECharts option={cpuGaugeOption} style={{ height: 200 }} />
              </Col>
              <Col span={8}>
                <ReactECharts option={memoryGaugeOption} style={{ height: 200 }} />
              </Col>
              <Col span={8}>
                <div style={{ textAlign: 'center' }}>
                  <Title level={5}>磁盘使用</Title>
                  <Progress
                    type="dashboard"
                    percent={agent.disk_usage || 0}
                    strokeColor={
                      (agent.disk_usage || 0) > 80 ? '#ff4d4f' :
                      (agent.disk_usage || 0) > 60 ? '#faad14' : '#52c41a'
                    }
                    format={(percent) => `${percent}%`}
                  />
                </div>
              </Col>
            </Row>
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card title="历史指标">
            <Space style={{ marginBottom: 16 }}>
              <Button
                type={timeRange === '1h' ? 'primary' : 'default'}
                size="small"
                onClick={() => setTimeRange('1h')}
              >
                1 小时
              </Button>
              <Button
                type={timeRange === '24h' ? 'primary' : 'default'}
                size="small"
                onClick={() => setTimeRange('24h')}
              >
                24 小时
              </Button>
              <Button
                type={timeRange === '7d' ? 'primary' : 'default'}
                size="small"
                onClick={() => setTimeRange('7d')}
              >
                7 天
              </Button>
            </Space>
            <ReactECharts
              option={{
                tooltip: { trigger: 'axis' },
                legend: { data: ['CPU', '内存', '磁盘'], bottom: 0 },
                xAxis: { type: 'category', data: [] },
                yAxis: { type: 'value', max: 100 },
                series: [
                  { name: 'CPU', type: 'line', smooth: true, data: [] },
                  { name: '内存', type: 'line', smooth: true, data: [] },
                  { name: '磁盘', type: 'line', smooth: true, data: [] },
                ],
              }}
              style={{ height: 300 }}
            />
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <RecentTasks agentId={agent.id} limit={5} />
        </Col>
      </Row>

      <ExecuteTaskModal
        open={executeModalOpen}
        onClose={() => setExecuteModalOpen(false)}
        onSuccess={fetchAgent}
        initialAgentIds={[agent.id]}
      />
    </div>
  )
}
