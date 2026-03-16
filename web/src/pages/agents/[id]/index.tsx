import { useEffect, useState, useMemo, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Button, Typography, Space, Tabs } from 'antd'
import {
  ArrowLeftOutlined,
  ReloadOutlined,
  PlayCircleOutlined,
  DashboardOutlined,
  LineChartOutlined,
  DatabaseOutlined,
  ClockCircleOutlined,
  SettingOutlined,
} from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'
import { agentApi } from '@/api'
import { formatDate } from '@/utils'
import type { Agent } from '@/types'
import { AgentTerminal } from '@/components/terminal'
import ExecuteTaskModal from '@/components/tasks/ExecuteTaskModal'
import styles from './index.module.css'

const { Text } = Typography

type TimeRange = '1h' | '24h' | '7d'
type TabKey = 'terminal' | 'tasks'

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
  const [timeRange, setTimeRange] = useState<TimeRange>('1h')
  const [executeModalOpen, setExecuteModalOpen] = useState(false)
  const [activeTab, setActiveTab] = useState<TabKey>('terminal')

  const isOnline = agent?.status === 'online'

  const fetchAgent = useCallback(async () => {
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
  }, [id])

  const fetchMetrics = useCallback(async () => {
    if (!id) return
    try {
      const data = await agentApi.getMetrics(id, timeRange)
      setMetrics(data as MetricData[])
    } catch (error) {
      console.error('Failed to fetch metrics:', error)
    }
  }, [id, timeRange])

  useEffect(() => { fetchAgent() }, [id, fetchAgent])
  useEffect(() => { fetchMetrics() }, [id, timeRange, fetchMetrics])

  const historyChartOption = useMemo(() => {
    const chartData = metrics.slice().reverse()
    const xAxisData = chartData.map((m) => {
      const date = new Date(m.collected_at)
      if (timeRange === '1h') {
        return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
      }
      return `${date.toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' })} ${date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })}`
    })
    const cpuData = chartData.map((m) => m.cpu_usage)
    const memoryData = chartData.map((m) => m.memory_percent)
    const diskData = chartData.map((m) => m.disk_percent)

    return {
      backgroundColor: 'transparent',
      grid: { top: 15, right: 10, bottom: 25, left: 40, containLabel: false },
      tooltip: {
        trigger: 'axis',
        backgroundColor: 'rgba(10, 10, 15, 0.95)',
        borderColor: 'rgba(0, 243, 255, 0.3)',
        borderWidth: 1,
        textStyle: { color: '#ffffff', fontSize: 12 },
        formatter: (params: Array<{ seriesName: string; value: number; axisValue: string }>) => {
          let result = `<span style="color:rgba(255,255,255,0.6)">${params[0]?.axisValue}</span><br/><br/>`
          params.forEach(item => {
            result += `<span style="display:inline-block;width:8px;height:8px;border-radius:50%;margin-right:6px;background:${item.color}"></span>`
            result += `<span style="color:white;">${item.seriesName}: ${item.value?.toFixed(1)}%</span><br/>`
          })
          return result
        },
      },
      legend: {
        show: true,
        bottom: 0,
        textStyle: { color: 'rgba(255, 255, 255, 0.5)', fontSize: 11 },
        itemGap: 20,
      },
      xAxis: {
        type: 'category',
        data: xAxisData,
        axisLine: { lineStyle: { color: 'rgba(255, 255, 255, 0.08)' } },
        axisTick: { show: false },
        axisLabel: {
          color: 'rgba(255, 255, 255, 0.4)',
          fontSize: 10,
          rotate: timeRange === '1h' ? 0 : 45,
          interval: timeRange === '1h' ? 'auto' : Math.floor(xAxisData.length / 8),
        },
      },
      yAxis: {
        type: 'value',
        max: 100,
        splitLine: {
          lineStyle: {
            color: 'rgba(255, 255, 255, 0.04)',
            type: [5, 5],
          },
        },
        axisLine: { show: false },
        axisTick: { show: false },
        axisLabel: {
          color: 'rgba(255, 255, 255, 0.4)',
          fontSize: 10,
          formatter: '{value}%',
        },
      },
      series: [
        {
          name: 'CPU',
          type: 'line',
          smooth: false,
          symbol: 'none',
          lineStyle: { width: 1.5, color: '#00f3ff' },
          areaStyle: {
            color: {
              type: 'linear',
              x: 0, y: 0, x2: 0, y2: 1,
              colorStops: [
                { offset: 0, color: 'rgba(0, 243, 255, 0.3)' },
                { offset: 1, color: 'rgba(0, 243, 255, 0)' },
              ],
            },
          },
          data: cpuData,
        },
        {
          name: '内存',
          type: 'line',
          smooth: false,
          symbol: 'none',
          lineStyle: { width: 1.5, color: '#00ff8f' },
          areaStyle: {
            color: {
              type: 'linear',
              x: 0, y: 0, x2: 0, y2: 1,
              colorStops: [
                { offset: 0, color: 'rgba(0, 255, 143, 0.3)' },
                { offset: 1, color: 'rgba(0, 255, 143, 0)' },
              ],
            },
          },
          data: memoryData,
        },
        {
          name: '磁盘',
          type: 'line',
          smooth: false,
          symbol: 'none',
          lineStyle: { width: 1.5, color: '#ff6b35' },
          areaStyle: {
            color: {
              type: 'linear',
              x: 0, y: 0, x2: 0, y2: 1,
              colorStops: [
                { offset: 0, color: 'rgba(255, 107, 53, 0.3)' },
                { offset: 1, color: 'rgba(255, 107, 53, 0)' },
              ],
            },
          },
          data: diskData,
        },
      ],
    }
  }, [metrics, timeRange])

  const createGaugeOption = (value: number, name: string, color: string) => ({
    series: [{
      type: 'gauge',
      startAngle: 180,
      endAngle: 0,
      min: 0,
      max: 100,
      radius: '80%',
      splitNumber: 5,
      axisLine: { lineStyle: { width: 8, color: [[1, 'rgba(255, 255, 255, 0.08)']] } },
      pointer: { itemStyle: { color }, length: '55%', width: 4 },
      axisTick: { show: true, splitNumber: 2, length: 4, lineStyle: { color: 'rgba(255, 255, 255, 0.15)' } },
      splitLine: { show: true, length: 8, lineStyle: { color: 'rgba(255, 255, 255, 0.08)', width: 2 } },
      axisLabel: { show: true, distance: 12, color: 'rgba(255, 255, 255, 0.35)', fontSize: 11 },
      title: { offsetCenter: [0, '55%'], fontSize: 13, color: 'rgba(255, 255, 255, 0.6)' },
      detail: {
        valueAnimation: true,
        fontSize: 32,
        offsetCenter: [0, '5%'],
        color,
        fontWeight: 500,
        formatter: (value: number) => `${value.toFixed(1)}%`,
      },
      data: [{ value: value || 0, name }],
    }],
  })

  const tabItems = [
    {
      key: 'terminal',
      label: (
        <span className={styles.tabLabel}>
          <span>终端控制台</span>
          {isOnline && <span className={`${styles.statusIndicator} ${styles.online}`}></span>}
        </span>
      ),
      children: <div className={styles.terminalContainer}><AgentTerminal agent={agent} /></div>,
    },
    {
      key: 'tasks',
      label: <span className={styles.tabLabel}>任务执行记录</span>,
      children: <div className={styles.tasksContainer}><Text className={styles.tasksEmptyText}>暂无任务执行记录</Text></div>,
    },
  ]

  // Render Loading State
  if (loading) {
    return (
      <div className={styles.loading}>
        <div className={styles.loadingContent}>
          <div className={styles.loader}></div>
          <div className={styles.loadingText}>系统初始化中...</div>
        </div>
      </div>
    )
  }

  // Render Not Found State
  if (!agent) {
    return (
      <div className={styles.notFound}>
        <Text className={styles.notFoundText}>系统未找到该 Agent 节点</Text>
      </div>
    )
  }

  // Render Main Content
  return (
    <div className={`${styles.container} ${styles.detailPage}`}>
      {/* Scanline Overlay */}
      <div className={styles.scanlines} />

      {/* Header */}
      <header className={styles.header}>
        <div className={styles.headerLeft}>
          <Button
            className={styles.navButton}
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/agents')}
          >
            返回节点列表
          </Button>
          <div className={styles.headerDivider} />
          <div>
            <h1 className={styles.agentName}>{agent.name}</h1>
            <div className={styles.agentMeta}>
              <span className={styles.hostname}>{agent.hostname || 'Unknown Host'}</span>
              <span className={styles.metaSeparator}>·</span>
              <span className={styles.ip}>{agent.ip_address || '-'}</span>
            </div>
          </div>
        </div>
        <Space size={12}>
          <Button className={styles.headerButton} icon={<ReloadOutlined />} onClick={fetchAgent}>
            刷新数据
          </Button>
          <Button
            className={`${styles.headerButton} ${styles.primaryButton}`}
            type="primary"
            icon={<PlayCircleOutlined />}
            onClick={() => setExecuteModalOpen(true)}
            disabled={!isOnline}
          >
            批量执行任务
          </Button>
        </Space>
      </header>

      {/* Main Grid */}
      <div className={styles.grid}>
        {/* Info Card */}
        <section className={`${styles.card} ${styles.infoCard}`}>
          <h2 className={styles.cardTitle}>
            <DashboardOutlined />
            <span>系统信息</span>
            <span className={styles.cardTitleLine} />
          </h2>
          <div className={styles.infoGrid}>
            <div className={styles.infoRow}>
              <span className={styles.infoLabel}>Node ID</span>
              <span className={styles.infoValue}>{id}</span>
            </div>
            <div className={styles.infoRow}>
              <span className={styles.infoLabel}>状态</span>
              <span className={`${styles.status} ${styles[isOnline ? 'online' : 'offline']}`}>
                <span className={styles.statusDot} />
                {isOnline ? 'ONLINE' : 'OFFLINE'}
              </span>
            </div>
            <div className={styles.infoRow}>
              <span className={styles.infoLabel}>操作系统</span>
              <span className={styles.infoValue}>{agent.os_info || '-'}</span>
            </div>
            <div className={styles.infoRow}>
              <span className={styles.infoLabel}>Agent 版本</span>
              <span className={styles.infoValue}>{agent.version || '-'}</span>
            </div>
            <div className={styles.infoRow}>
              <span className={styles.infoLabel}><ClockCircleOutlined /></span>
              <span className={styles.infoValue}>{agent.last_seen_at ? formatDate(agent.last_seen_at) : '-'}</span>
            </div>
            <div className={styles.infoRow}>
              <span className={styles.infoLabel}>首次上线</span>
              <span className={styles.infoValue}>{agent.created_at ? formatDate(agent.created_at) : '-'}</span>
            </div>
          </div>
        </section>

        {/* Metrics Card */}
        <section className={`${styles.card} ${styles.metricsCard}`}>
          <h2 className={styles.cardTitle}>
            <LineChartOutlined />
            <span>实时负载</span>
            <span className={styles.cardTitleLine} />
          </h2>
          <div className={styles.gaugesGrid}>
            <div className={styles.gaugeWrapper}>
              <ReactECharts
                option={createGaugeOption(agent.cpu_usage || 0, 'CPU', '#00f3ff')}
                style={{ width: '100%', height: '100%' }}
                opts={{ renderer: 'svg' }}
                lazyUpdate
              />
            </div>
            <div className={styles.gaugeWrapper}>
              <ReactECharts
                option={createGaugeOption(agent.memory_usage || 0, '内存', '#00ff8f')}
                style={{ width: '100%', height: '100%' }}
                opts={{ renderer: 'svg' }}
                lazyUpdate
              />
            </div>
            <div className={styles.gaugeWrapper}>
              <ReactECharts
                option={createGaugeOption(agent.disk_usage || 0, '磁盘', '#ff6b35')}
                style={{ width: '100%', height: '100%' }}
                opts={{ renderer: 'svg' }}
                lazyUpdate
              />
            </div>
          </div>
        </section>

        {/* Chart Card */}
        <section className={`${styles.card} ${styles.chartCard}`}>
          <div className={styles.chartHeader}>
            <h2 className={styles.cardTitle}>
              <DatabaseOutlined />
              <span>指标趋势</span>
              <span className={styles.cardTitleLine} />
            </h2>
            <div className={styles.timeButtons}>
              {(['1h', '24h', '7d'] as TimeRange[]).map((range) => (
                <button
                  key={range}
                  className={`${styles.timeButton} ${timeRange === range ? styles.active : ''}`}
                  onClick={() => setTimeRange(range)}
                >
                  {range.toUpperCase()}
                </button>
              ))}
            </div>
          </div>
          <div className={styles.chartWrapper}>
            <ReactECharts
              option={historyChartOption}
              style={{ width: '100%', height: '50%' }}
              opts={{ renderer: 'svg' }}
              lazyUpdate
              notMerge
            />
          </div>
        </section>

        {/* Terminal Card */}
        <section className={`${styles.card} ${styles.terminalCard}`}>
          <h2 className={styles.cardTitle}>
            <SettingOutlined />
            <span>终端控制台</span>
            <span className={styles.cardTitleLine} />
          </h2>
          <Tabs
            activeKey={activeTab}
            onChange={setActiveTab}
            items={tabItems}
            className={styles.tabs}
          />
        </section>
      </div>

      <ExecuteTaskModal
        open={executeModalOpen}
        onClose={() => setExecuteModalOpen(false)}
        onSuccess={fetchAgent}
        agentId={id}
      />
    </div>
  )
}
