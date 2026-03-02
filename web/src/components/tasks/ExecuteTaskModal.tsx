import { useState, useEffect } from 'react'
import { Modal, Form, Select, Input, InputNumber, Space, Tag, message, Spin, Alert } from 'antd'
import { PlayCircleOutlined } from '@ant-design/icons'
import { agentApi, taskApi } from '@/api'
import type { Agent, CreateTaskRequest } from '@/types'

interface ExecuteTaskModalProps {
  open: boolean
  onClose: () => void
  onSuccess: () => void
  initialAgentIds?: string[]
}

const builtinCommands = [
  { value: 'system_info', label: '系统信息' },
  { value: 'network_info', label: '网络信息' },
  { value: 'disk_usage', label: '磁盘使用情况' },
  { value: 'process_list', label: '进程列表' },
  { value: 'service_list', label: '服务列表' },
  { value: 'restart_service', label: '重启服务' },
  { value: 'clear_cache', label: '清理缓存' },
]

export default function ExecuteTaskModal({ open, onClose, onSuccess, initialAgentIds }: ExecuteTaskModalProps) {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [agents, setAgents] = useState<Agent[]>([])
  const [agentsLoading, setAgentsLoading] = useState(false)
  const [commandType, setCommandType] = useState<'shell' | 'builtin'>('shell')
  const [selectedAgents, setSelectedAgents] = useState<string[]>(initialAgentIds || [])

  const fetchAgents = async () => {
    setAgentsLoading(true)
    try {
      const response = await agentApi.list({ page: 1, page_size: 100, status: 'online' })
      setAgents(response.data)
    } catch (error) {
      console.error('Failed to fetch agents:', error)
    } finally {
      setAgentsLoading(false)
    }
  }

  useEffect(() => {
    if (open) {
      fetchAgents()
      if (initialAgentIds && initialAgentIds.length > 0) {
        setSelectedAgents(initialAgentIds)
        form.setFieldValue('agent_ids', initialAgentIds)
      }
    }
  }, [open, initialAgentIds, form])

  const handleSubmit = async (values: CreateTaskRequest) => {
    if (selectedAgents.length === 0) {
      message.error('请选择至少一个 Agent')
      return
    }

    setLoading(true)
    try {
      // Create tasks for all selected agents
      const promises = selectedAgents.map((agentId) =>
        taskApi.create({
          agent_ids: [agentId],
          command_type: values.command_type,
          command: values.command,
          timeout: values.timeout || 300,
          priority: values.priority || 0,
        })
      )

      await Promise.all(promises)
      message.success(`已创建 ${selectedAgents.length} 个任务`)
      form.resetFields()
      setSelectedAgents([])
      onSuccess()
      onClose()
    } catch {
      message.error('创建任务失败')
    } finally {
      setLoading(false)
    }
  }

  const handleClose = () => {
    form.resetFields()
    setSelectedAgents([])
    setCommandType('shell')
    onClose()
  }

  return (
    <Modal
      title={
        <Space>
          <PlayCircleOutlined />
          <span>执行任务</span>
        </Space>
      }
      open={open}
      onCancel={handleClose}
      onOk={() => form.submit()}
      okText="执行"
      cancelText="取消"
      confirmLoading={loading}
      width={600}
      destroyOnClose
    >
      {agentsLoading ? (
        <div style={{ textAlign: 'center', padding: 40 }}>
          <Spin />
        </div>
      ) : (
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            command_type: 'shell',
            timeout: 300,
            priority: 0,
          }}
        >
          <Form.Item
            name="agent_ids"
            label="选择 Agent"
            rules={[{ required: true, message: '请选择至少一个 Agent' }]}
          >
            <Select
              mode="multiple"
              placeholder="选择要执行任务的 Agent"
              loading={agentsLoading}
              value={selectedAgents}
              onChange={(values) => setSelectedAgents(values as string[])}
              optionFilterProp="label"
              options={agents.map((agent) => ({
                value: agent.id,
                label: `${agent.name} (${agent.hostname})`,
              }))}
              maxTagCount={5}
              maxTagPlaceholder={(omitted) => `+${omitted.length} 更多`}
            />
          </Form.Item>

          {selectedAgents.length > 0 && (
            <Alert
              type="info"
              showIcon
              message={`已选择 ${selectedAgents.length} 个 Agent`}
              style={{ marginBottom: 16 }}
            />
          )}

          <Form.Item
            name="command_type"
            label="命令类型"
            rules={[{ required: true }]}
          >
            <Select onChange={(value) => setCommandType(value as 'shell' | 'builtin')}>
              <Select.Option value="shell">
                <Tag color="blue">Shell</Tag> 执行 Shell 命令
              </Select.Option>
              <Select.Option value="builtin">
                <Tag color="green">内置</Tag> 执行内置命令
              </Select.Option>
            </Select>
          </Form.Item>

          {commandType === 'shell' ? (
            <Form.Item
              name="command"
              label="Shell 命令"
              rules={[{ required: true, message: '请输入要执行的命令' }]}
            >
              <Input.TextArea
                placeholder="例如: ping -n 4 google.com"
                rows={3}
                autoSize={{ minRows: 2, maxRows: 6 }}
              />
            </Form.Item>
          ) : (
            <Form.Item
              name="command"
              label="内置命令"
              rules={[{ required: true, message: '请选择内置命令' }]}
            >
              <Select placeholder="选择要执行的内置命令" options={builtinCommands} />
            </Form.Item>
          )}

          <Space style={{ width: '100%' }} size="large">
            <Form.Item
              name="timeout"
              label="超时时间 (秒)"
              tooltip="任务执行的最大等待时间"
            >
              <InputNumber min={10} max={3600} style={{ width: 120 }} />
            </Form.Item>

            <Form.Item
              name="priority"
              label="优先级"
              tooltip="数值越高优先级越高"
            >
              <InputNumber min={0} max={100} style={{ width: 120 }} />
            </Form.Item>
          </Space>
        </Form>
      )}
    </Modal>
  )
}
