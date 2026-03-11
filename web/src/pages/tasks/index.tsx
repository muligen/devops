import { useEffect, useState, useCallback } from 'react'
import { Table, Card, Tag, Select, Space, Button, Typography, Drawer, Descriptions, Modal, message } from 'antd'
import { ReloadOutlined, EyeOutlined, StopOutlined } from '@ant-design/icons'
import { taskApi } from '@/api'
import { formatRelativeTime, formatDate, formatDuration, getStatusColor, getStatusText } from '@/utils'
import type { Task, TaskListParams } from '@/types'
import styles from './index.module.css'

const { Title, Text } = Typography
const { confirm } = Modal

export default function TasksPage() {
  const [loading, setLoading] = useState(false)
  const [tasks, setTasks] = useState<Task[]>([])
  const [total, setTotal] = useState(0)
  const [selectedTask, setSelectedTask] = useState<Task | null>(null)
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [params, setParams] = useState<TaskListParams>({
    page: 1,
    page_size: 20,
    status: undefined,
  })

  const fetchTasks = useCallback(async () => {
    setLoading(true)
    try {
      const response = await taskApi.list(params)
      setTasks(response.data)
      setTotal(response.total)
    } catch (error) {
      console.error('Failed to fetch tasks:', error)
    } finally {
      setLoading(false)
    }
  }, [params])

  useEffect(() => {
    fetchTasks()
  }, [fetchTasks])

  const handleCancelTask = (task: Task) => {
    confirm({
      title: '确认取消任务',
      content: `确定要取消任务 "${task.id}" 吗？`,
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await taskApi.cancel(task.id)
          message.success('任务已取消')
          fetchTasks()
        } catch {
          message.error('取消任务失败')
        }
      },
    })
  }

  const columns = [
    {
      title: '任务 ID',
      dataIndex: 'id',
      key: 'id',
      width: 100,
      ellipsis: true,
      render: (id: string) => (
        <Text copyable style={{ fontSize: 12 }}>{id.slice(0, 8)}</Text>
      ),
    },
    {
      title: 'Agent',
      dataIndex: 'agent_id',
      key: 'agent_id',
      width: 100,
      ellipsis: true,
      render: (id: string) => <Text style={{ fontSize: 12 }}>{id.slice(0, 8)}</Text>,
    },
    {
      title: '类型',
      key: 'type',
      width: 100,
      render: (_: unknown, record: Task) => (
        <Tag color={record.type === 'exec_shell' ? 'blue' : 'green'}>
          {record.type === 'exec_shell' ? 'Shell' : record.type}
        </Tag>
      ),
    },
    {
      title: '命令',
      key: 'command',
      ellipsis: true,
      render: (_: unknown, record: Task) => (
        <Text code style={{ fontSize: 12 }}>
          {record.params?.command as string || '-'}
        </Text>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>{getStatusText(status)}</Tag>
      ),
    },
    {
      title: '退出码',
      dataIndex: 'exit_code',
      key: 'exit_code',
      width: 80,
      render: (code: number | null) => code !== null ? code : '-',
    },
    {
      title: '超时',
      dataIndex: 'timeout',
      key: 'timeout',
      width: 100,
      render: (timeout: number) => formatDuration(timeout),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 140,
      render: (time: string) => formatRelativeTime(time),
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_: unknown, record: Task) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => {
              setSelectedTask(record)
              setDrawerOpen(true)
            }}
          >
            详情
          </Button>
          {(record.status === 'pending' || record.status === 'running') && (
            <Button
              type="link"
              size="small"
              danger
              icon={<StopOutlined />}
              onClick={() => handleCancelTask(record)}
            >
              取消
            </Button>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <Title level={4} style={{ margin: 0 }}>任务管理</Title>
        <Button
          icon={<ReloadOutlined />}
          onClick={fetchTasks}
          className={styles.refreshButton}
        >
          刷新
        </Button>
      </div>

      <Card className={styles.card}>
        <Space style={{ marginBottom: 16 }}>
          <Select
            placeholder="状态筛选"
            allowClear
            style={{ width: 120 }}
            onChange={(value) => setParams({ ...params, page: 1, status: value })}
            options={[
              { value: 'pending', label: '待处理' },
              { value: 'running', label: '运行中' },
              { value: 'completed', label: '已完成' },
              { value: 'failed', label: '失败' },
              { value: 'cancelled', label: '已取消' },
            ]}
          />
        </Space>

        <Table
          className={styles.table}
          columns={columns}
          dataSource={tasks}
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
          onChange={(pagination) => setParams({
            ...params,
            page: pagination.current || 1,
            page_size: pagination.pageSize || 20,
          })}
          scroll={{ x: 1200 }}
        />
      </Card>

      <Drawer
        className={styles.drawer}
        title="任务详情"
        width={600}
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
      >
        {selectedTask && (
          <>
            <Descriptions column={2} bordered size="small">
              <Descriptions.Item label="任务 ID" span={2}>
                <Text copyable>{selectedTask.id}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="Agent ID">{selectedTask.agent_id}</Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={getStatusColor(selectedTask.status)}>
                  {getStatusText(selectedTask.status)}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="任务类型">{selectedTask.type}</Descriptions.Item>
              <Descriptions.Item label="退出码">{selectedTask.exit_code ?? '-'}</Descriptions.Item>
              <Descriptions.Item label="命令" span={2}>
                <Text code>{selectedTask.params?.command as string || '-'}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="超时">{formatDuration(selectedTask.timeout)}</Descriptions.Item>
              <Descriptions.Item label="优先级">{selectedTask.priority}</Descriptions.Item>
              <Descriptions.Item label="创建时间">
                {formatDate(selectedTask.created_at)}
              </Descriptions.Item>
              <Descriptions.Item label="开始时间">
                {selectedTask.started_at ? formatDate(selectedTask.started_at) : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="完成时间">
                {selectedTask.completed_at ? formatDate(selectedTask.completed_at) : '-'}
              </Descriptions.Item>
            </Descriptions>

            <Title level={5} style={{ marginTop: 24 }}>输出</Title>
            <pre className={styles.output}>
              {selectedTask.output || '无输出'}
            </pre>
          </>
        )}
      </Drawer>
    </div>
  )
}
