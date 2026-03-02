import { useEffect, useState, useCallback } from 'react'
import { Table, Card, Tag, Input, Select, Space, Button, Typography, Progress, Tooltip } from 'antd'
import type { TablePaginationConfig, FilterValue, SorterResult } from 'antd/es/table/interface'
import { ReloadOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { agentApi } from '@/api'
import { formatRelativeTime, formatPercent } from '@/utils'
import type { Agent, AgentListParams } from '@/types'

const { Title } = Typography
const { Search } = Input

export default function AgentsPage() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [agents, setAgents] = useState<Agent[]>([])
  const [total, setTotal] = useState(0)
  const [params, setParams] = useState<AgentListParams>({
    page: 1,
    page_size: 20,
    status: undefined,
    search: '',
    sort: 'cpu_usage',
    order: 'desc',
  })

  const fetchAgents = useCallback(async () => {
    setLoading(true)
    try {
      const response = await agentApi.list(params)
      setAgents(response.data)
      setTotal(response.total)
    } catch (error) {
      console.error('Failed to fetch agents:', error)
    } finally {
      setLoading(false)
    }
  }, [params])

  useEffect(() => {
    fetchAgents()
  }, [fetchAgents])

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (name: string, record: Agent) => (
        <a onClick={() => navigate(`/agents/${record.id}`)}>{name}</a>
      ),
    },
    {
      title: '主机名',
      dataIndex: 'hostname',
      key: 'hostname',
      ellipsis: true,
    },
    {
      title: 'IP 地址',
      dataIndex: 'ip_address',
      key: 'ip_address',
      width: 140,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Tag color={status === 'online' ? 'success' : 'error'}>
          {status === 'online' ? '在线' : '离线'}
        </Tag>
      ),
    },
    {
      title: 'CPU',
      dataIndex: 'cpu_usage',
      key: 'cpu_usage',
      width: 120,
      sorter: true,
      render: (value: number) => (
        <Tooltip title={formatPercent(value)}>
          <Progress
            percent={value || 0}
            size="small"
            showInfo={false}
            strokeColor={
              (value || 0) > 80 ? '#ff4d4f' : (value || 0) > 60 ? '#faad14' : '#52c41a'
            }
          />
        </Tooltip>
      ),
    },
    {
      title: '内存',
      dataIndex: 'memory_usage',
      key: 'memory_usage',
      width: 120,
      sorter: true,
      render: (value: number) => (
        <Tooltip title={formatPercent(value)}>
          <Progress
            percent={value || 0}
            size="small"
            showInfo={false}
            strokeColor={
              (value || 0) > 80 ? '#ff4d4f' : (value || 0) > 60 ? '#faad14' : '#52c41a'
            }
          />
        </Tooltip>
      ),
    },
    {
      title: '磁盘',
      dataIndex: 'disk_usage',
      key: 'disk_usage',
      width: 120,
      sorter: true,
      render: (value: number) => (
        <Tooltip title={formatPercent(value)}>
          <Progress
            percent={value || 0}
            size="small"
            showInfo={false}
            strokeColor={
              (value || 0) > 80 ? '#ff4d4f' : (value || 0) > 60 ? '#faad14' : '#52c41a'
            }
          />
        </Tooltip>
      ),
    },
    {
      title: 'Agent 版本',
      dataIndex: 'agent_version',
      key: 'agent_version',
      width: 120,
    },
    {
      title: '最后心跳',
      dataIndex: 'last_heartbeat_at',
      key: 'last_heartbeat_at',
      width: 120,
      render: (time: string) => formatRelativeTime(time),
    },
  ]

  const handleTableChange = (
    pagination: TablePaginationConfig,
    _filters: Record<string, FilterValue | null>,
    sorter: SorterResult<Agent> | SorterResult<Agent>[]
  ) => {
    const sortInfo = Array.isArray(sorter) ? sorter[0] : sorter
    setParams({
      ...params,
      page: pagination.current || 1,
      page_size: pagination.pageSize || 20,
      sort: sortInfo.field as string | undefined,
      order: sortInfo.order === 'ascend' ? 'asc' : 'desc',
    })
  }

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Title level={4} style={{ margin: 0 }}>Agent 管理</Title>
        <Button icon={<ReloadOutlined />} onClick={fetchAgents}>
          刷新
        </Button>
      </div>

      <Card>
        <Space style={{ marginBottom: 16 }}>
          <Search
            placeholder="搜索 Agent 名称/主机名"
            allowClear
            style={{ width: 250 }}
            onSearch={(value) => setParams({ ...params, page: 1, search: value })}
          />
          <Select
            placeholder="状态筛选"
            allowClear
            style={{ width: 120 }}
            onChange={(value) => setParams({ ...params, page: 1, status: value })}
            options={[
              { value: 'online', label: '在线' },
              { value: 'offline', label: '离线' },
            ]}
          />
        </Space>

        <Table
          columns={columns}
          dataSource={agents}
          rowKey="id"
          loading={loading}
          pagination={{
            current: params.page,
            pageSize: params.page_size,
            total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条`,
          }}
          onChange={handleTableChange}
          scroll={{ x: 1200 }}
        />
      </Card>
    </div>
  )
}
