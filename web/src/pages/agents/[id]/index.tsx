import { useEffect, useState, useMemo } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Card, Row, Col, Typography, Descriptions, Button, Progress, Space, Tabs } from 'antd'
import { ArrowLeftOutlined, ReloadOutlined, PlayCircleOutlined } from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'
import { agentApi } from '@/api'
import { formatDate } from '@/utils'
import type { Agent } from '@/types'
import { AgentTerminal } from '@/components/terminal'
import ExecuteTaskModal from '@/components/tasks/ExecuteTaskModal'
import styles from './index.module.css'

const { Title } = Typography

interface MetricData {
  collected_at: string
  cpu_usage: number
  memory_percent: number
  disk_percent: number
}

export default function AgentDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [loading, setLoading] = useState(true)
  const [agent, setAgent] = useState<Agent | null>(null)
  const [metrics, setMetrics] = useState<MetricData[]>([])
  const [timeRange, setTimeRange] = useState('1h')
  const [executeModalOpen, setExecuteModalOpen] = useState(false)
  const [activeTab, setActiveTab] = useState('terminal')

  const fetchAgent = async () => {
    if (!id) return
    setLoading(true)
    try {
      const data = await agentApi.get(id)
      setAgent(data)
    } catch (error) {
      console.error('Failed to fetch agent:', error)
    } finally {
      setLoading(false)
    }
  }

  const fetchMetrics = async () => {
    if (!id) return
    try {
      const data = await agentApi.getMetrics(id, timeRange)
      setMetrics(data as MetricData[])
    } catch (error) {
      console.error('Failed to fetch metrics:', error)
    }
  }

  useEffect(() => {
    fetchAgent()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id])

  useEffect(() => {
    fetchMetrics()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id, timeRange])

  const historyChartOption = useMemo(() => {
    const chartData = metrics.slice().reverse()
    const xAxisData = chartData.map((m) => {
      const date = new Date(m.collected_at)
      if (timeRange === '1h') {
        return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
      } else {
        return date.toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' }) + ' ' +
               date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
      }
    })
    const cpuData = chartData.map((m) => m.cpu_usage)
        const memoryData = chartData.map((m) => m.memory_percent)
        const diskData = chartData.map((m) => m.disk_percent)

    return {
      tooltip: {
        trigger: 'axis',
        formatter: (params: Array<{ seriesName: string; value: number; axisValue: string }>) => {
          let result = params[0]?.axisValue + '<br/>'
          params.forEach(item => {
            result += `${item.seriesName}: ${item.value?.toFixed(1)}%<br/>`
          })
          return result
        },
      },
      legend: { data: ['CPU', '内存', '磁盘'], bottom: 0 },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '15%',
        top: '10%',
        containLabel: true,
      },
      xAxis: {
        type: 'category',
        data: xAxisData,
        axisLabel: {
          rotate: timeRange === '1h' ? 0 : 45,
          interval: timeRange === '1h' ? 'auto' : Math.floor(xAxisData.length / 8),
        }
      },
      yAxis: { type: 'value', max: 100 },
      series: [
        { name: 'CPU', type: 'line', smooth: true, data: cpuData },
        { name: '内存', type: 'line', smooth: true, data: memoryData },
        { name: '磁盘', type: 'line', smooth: true, data: diskData },
      ],
    }
  }, [metrics, timeRange])

  if (loading) {
    return (
      <div className={styles.loading}>
        <span>加载中...</span>
      </div>
    )
  }

  if (!agent) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 400 }}>
        <span>Agent 不存在</span>
      </div>
    )
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
        pointer: { itemStyle: { color: 'auto' } },
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
        pointer: { itemStyle: { color: 'auto' } },
        axisTick: { show: false },
        splitLine: { show: false },
        axisLabel: { show: false },
        title: { offsetCenter: [0, '70%'], fontSize: 14 },
        detail: { valueAnimation: true, fontSize: 24, offsetCenter: [0, '0%'] },
        data: [{ value: agent.memory_usage || 0, name: '内存' }],
      },
    ],
  }

  const tabItems = [
    {
      key: 'terminal',
      label: (
        <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
          <span>终端</span>
          {agent.status === 'online' && (
            <span style={{
              fontSize: 11,
              color: '#73d13d',
            }}>
              ● 在线
            </span>
          )}
        </span>
      ),
      children: (
        <div style={{ height: '600px' }}>
          <AgentTerminal
            agent={agent}
          />
        </div>
      ),
    },
    {
      key: 'tasks',
      label: '最近任务',
      children: (
        <div style={{ padding: 16 }}>
          <span>最近任务列表将在此显示...</span>
        </div>
      ),
    },
  ]

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/agents')}>
          返回列表
        </Button>
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
            批量执行
          </Button>
        </Space>
      </div>

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={8}>
          <Card title={agent.name}>
            <Descriptions column={1} size="small">
              <Descriptions.Item label="主机名">{agent.hostname || '-'}</Descriptions.Item>
              <Descriptions.Item label="IP 地址">{agent.ip_address || '-'}</Descriptions.Item>
              <Descriptions.Item label="操作系统">{agent.os_info || '-'}</Descriptions.Item>
              <Descriptions.Item label="Agent 版本">{agent.version || '-'}</Descriptions.Item>
              <Descriptions.Item label="首次上线">{agent.created_at ? formatDate(agent.created_at) : '-'}</Descriptions.Item>
              <Descriptions.Item label="最后心跳">{agent.last_seen_at ? formatDate(agent.last_seen_at) : '-'}</Descriptions.Item>
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
              option={historyChartOption}
              style={{ height: 300 }}
              opts={{ renderer: 'svg' }}
            />
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card title={<span className={styles.terminalTabTitle}>交互终端</span>}>
            <Tabs
              activeKey={activeTab}
              onChange={setActiveTab}
              items={tabItems}
            />
          </Card>
        </Col>
      </Row>

      <ExecuteTaskModal
        open={executeModalOpen}
        onClose={() => setExecuteModalOpen(false)}
        onSuccess={fetchAgent}
        agentId={id}
      />
    </div>
  )
}
