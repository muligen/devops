import { Card, Table, Tag, Typography, Empty, Button } from 'antd'
import { WarningOutlined, CheckCircleOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { formatRelativeTime } from '@/utils'
import type { AlertEvent } from '@/types'

const { Text } = Typography

interface RecentAlertsProps {
  alerts: AlertEvent[]
  loading?: boolean
}

export default function RecentAlerts({ alerts, loading }: RecentAlertsProps) {
  const navigate = useNavigate()

  const columns = [
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: string) => {
        const config: Record<string, { color: string; icon: React.ReactNode }> = {
          pending: { color: 'warning', icon: <WarningOutlined /> },
          acknowledged: { color: 'processing', icon: <CheckCircleOutlined /> },
          resolved: { color: 'success', icon: <CheckCircleOutlined /> },
        }
        const { color, icon } = config[status] || { color: 'default', icon: null }
        return (
          <Tag color={color} icon={icon}>
            {status === 'pending' ? '待处理' : status === 'acknowledged' ? '已确认' : '已解决'}
          </Tag>
        )
      },
    },
    {
      title: '规则',
      dataIndex: 'rule_name',
      key: 'rule_name',
      ellipsis: true,
    },
    {
      title: 'Agent',
      dataIndex: 'agent_name',
      key: 'agent_name',
      ellipsis: true,
    },
    {
      title: '指标值',
      dataIndex: 'metric_value',
      key: 'metric_value',
      width: 100,
      render: (value: number, record: AlertEvent) => (
        <Text>
          {value.toFixed(1)}%{' '}
          <Text type="secondary" style={{ fontSize: 12 }}>
            (阈值: {record.threshold}%)
          </Text>
        </Text>
      ),
    },
    {
      title: '触发时间',
      dataIndex: 'triggered_at',
      key: 'triggered_at',
      width: 120,
      render: (time: string) => formatRelativeTime(time),
    },
  ]

  return (
    <Card
      title="最近告警事件"
      loading={loading}
      extra={
        <Button type="link" onClick={() => navigate('/alerts')}>
          查看全部
        </Button>
      }
    >
      {alerts.length === 0 ? (
        <Empty description="暂无告警事件" />
      ) : (
        <Table
          columns={columns}
          dataSource={alerts}
          rowKey="id"
          pagination={false}
          size="small"
          scroll={{ x: 600 }}
        />
      )}
    </Card>
  )
}
