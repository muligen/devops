import { useMemo } from 'react'
import { Modal, Table, Tag, Progress, Typography, Empty, Spin } from 'antd'
import { CheckCircleOutlined, CloseCircleOutlined, LoadingOutlined } from '@ant-design/icons'
import type { Task } from '@/types'
import { getStatusColor, getStatusText } from '@/utils'

const { Text } = Typography

interface TaskResult {
  agentId: string
  agentName: string
  taskId: string
  status: Task['status']
  output?: string
  exitCode?: number | null
}

interface ExecutionResultModalProps {
  open: boolean
  onClose: () => void
  tasks: TaskResult[]
  loading?: boolean
}

export default function ExecutionResultModal({
  open,
  onClose,
  tasks,
  loading,
}: ExecutionResultModalProps) {
  const summary = useMemo(() => {
    const total = tasks.length
    const completed = tasks.filter((t) => t.status === 'completed').length
    const failed = tasks.filter((t) => t.status === 'failed').length
    const running = tasks.filter((t) => t.status === 'running').length
    const pending = tasks.filter((t) => t.status === 'pending').length

    return { total, completed, failed, running, pending }
  }, [tasks])

  const getProgressPercent = () => {
    if (summary.total === 0) return 0
    return Math.round(((summary.completed + summary.failed) / summary.total) * 100)
  }

  const getProgressStatus = () => {
    if (summary.failed > 0) return 'exception'
    if (summary.completed === summary.total) return 'success'
    return 'active'
  }

  const columns = [
    {
      title: 'Agent',
      dataIndex: 'agentName',
      key: 'agentName',
      width: 150,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: Task['status']) => (
        <Tag color={getStatusColor(status)} icon={getStatusIcon(status)}>
          {getStatusText(status)}
        </Tag>
      ),
    },
    {
      title: '退出码',
      dataIndex: 'exitCode',
      key: 'exitCode',
      width: 80,
      render: (code: number | null | undefined) =>
        code !== null && code !== undefined ? code : '-',
    },
    {
      title: '输出',
      dataIndex: 'output',
      key: 'output',
      ellipsis: true,
      render: (output: string | undefined) =>
        output ? (
          <Text style={{ maxWidth: 200 }} ellipsis title={output}>
            {output}
          </Text>
        ) : (
          '-'
        ),
    },
  ]

  const getStatusIcon = (status: Task['status']) => {
    switch (status) {
      case 'completed':
        return <CheckCircleOutlined />
      case 'failed':
        return <CloseCircleOutlined />
      case 'running':
        return <LoadingOutlined spin />
      default:
        return null
    }
  }

  return (
    <Modal
      title="执行结果"
      open={open}
      onCancel={onClose}
      footer={null}
      width={700}
      destroyOnClose
    >
      {loading ? (
        <div style={{ textAlign: 'center', padding: 40 }}>
          <Spin size="large" />
          <Text type="secondary" style={{ display: 'block', marginTop: 16 }}>
            正在执行任务...
          </Text>
        </div>
      ) : tasks.length === 0 ? (
        <Empty description="暂无执行结果" />
      ) : (
        <>
          <div style={{ marginBottom: 24 }}>
            <Progress
              percent={getProgressPercent()}
              status={getProgressStatus()}
              strokeColor={{
                '0%': '#1890ff',
                '100%': summary.failed > 0 ? '#ff4d4f' : '#52c41a',
              }}
            />
            <div style={{ display: 'flex', justifyContent: 'center', gap: 24, marginTop: 8 }}>
              <Text>
                总计: <strong>{summary.total}</strong>
              </Text>
              <Text type="success">
                成功: <strong>{summary.completed}</strong>
              </Text>
              <Text type="danger">
                失败: <strong>{summary.failed}</strong>
              </Text>
              {summary.running > 0 && (
                <Text type="warning">
                  运行中: <strong>{summary.running}</strong>
                </Text>
              )}
            </div>
          </div>

          <Table
            columns={columns}
            dataSource={tasks}
            rowKey="taskId"
            pagination={false}
            size="small"
            scroll={{ y: 300 }}
          />
        </>
      )}
    </Modal>
  )
}
