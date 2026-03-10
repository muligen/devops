import { useEffect, useState, useCallback } from 'react'
import { Card, Table, Tag, Typography, Empty, Button } from 'antd'
import { EyeOutlined } from '@ant-design/icons'
import { taskApi } from '@/api'
import { formatRelativeTime, getStatusColor, getStatusText } from '@/utils'
import type { Task } from '@/types'

const { Text } = Typography

interface RecentTasksProps {
  agentId: string
  limit?: number
  onViewTask?: (taskId: string) => void
}

export default function RecentTasks({ agentId, limit = 10, onViewTask }: RecentTasksProps) {
  const [loading, setLoading] = useState(false)
  const [tasks, setTasks] = useState<Task[]>([])

  const fetchTasks = useCallback(async () => {
    if (!agentId) return
    setLoading(true)
    try {
      const response = await taskApi.list({
        agent_id: agentId,
        page: 1,
        page_size: limit,
      })
      setTasks(response.data)
    } catch (error) {
      console.error('Failed to fetch tasks:', error)
    } finally {
      setLoading(false)
    }
  }, [agentId, limit])

  useEffect(() => {
    fetchTasks()
  }, [fetchTasks])

  const columns = [
    {
      title: '命令',
      key: 'command',
      ellipsis: true,
      render: (_: unknown, record: Task) => {
        const command = record.params?.command as string || '-'
        const isShell = record.type === 'exec_shell'
        return (
          <Text code style={{ fontSize: 12 }}>
            {isShell ? '$ ' : ''}
            {command}
          </Text>
        )
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 90,
      render: (status: Task['status']) => (
        <Tag color={getStatusColor(status)}>{getStatusText(status)}</Tag>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 120,
      render: (time: string) => formatRelativeTime(time),
    },
    {
      title: '',
      key: 'action',
      width: 60,
      render: (_: unknown, record: Task) => (
        <Button
          type="link"
          size="small"
          icon={<EyeOutlined />}
          onClick={() => onViewTask?.(record.id)}
        />
      ),
    },
  ]

  if (!tasks.length && !loading) {
    return (
      <Card title="最近任务" size="small">
        <Empty description="暂无任务记录" image={Empty.PRESENTED_IMAGE_SIMPLE} />
      </Card>
    )
  }

  return (
    <Card title="最近任务" size="small" loading={loading}>
      <Table
        columns={columns}
        dataSource={tasks}
        rowKey="id"
        pagination={false}
        size="small"
        scroll={{ x: 400 }}
      />
    </Card>
  )
}
