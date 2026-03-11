import { Card, Table, Typography, Empty, Button } from 'antd'
import { WarningOutlined, CheckCircleOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { formatRelativeTime } from '@/utils'
import type { AlertEvent } from '@/types'
import styles from './RecentAlerts.module.css'

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
      width: 90,
      render: (status: string) => {
        const config: Record<string, { icon: React.ReactNode }> = {
          pending: { icon: <WarningOutlined style={{ fontSize: 12 }} /> },
          acknowledged: { icon: <CheckCircleOutlined style={{ fontSize: 12 }} /> },
          resolved: { icon: <CheckCircleOutlined style={{ fontSize: 12 }} /> },
        }
        const { icon } = config[status] || { icon: null }
        const label = status === 'pending' ? '待处理' : status === 'acknowledged' ? '已确认' : '已解决'
        return (
          <span className={`${styles.statusBadge} ${status}`}>
            {icon}
            {label}
          </span>
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
      width: 110,
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
      width: 80,
      render: (time: string) => formatRelativeTime(time),
    },
  ]

  return (
    <Card
      className={styles.card}
      title="最近告警事件"
      loading={loading}
      extra={
        <Button type="link" onClick={() => navigate('/alerts')}>
          查看全部
        </Button>
      }
    >
      {alerts.length === 0 ? (
        <Empty
          className={styles.empty}
          description="暂无告警事件"
        />
      ) : (
        <Table
          className={styles.table}
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
