import { useEffect, useState, useCallback } from 'react'
import { Table, Card, Tag, Select, Space, Button, Typography, Tabs, Modal, Form, Input, InputNumber, Switch, message } from 'antd'
import { ReloadOutlined, PlusOutlined, CheckOutlined } from '@ant-design/icons'
import { alertApi } from '@/api'
import { formatRelativeTime, getStatusColor, getStatusText } from '@/utils'
import type { AlertRule, AlertEvent, AlertEventListParams, CreateAlertRuleRequest } from '@/types'
import styles from './index.module.css'

const { Title, Text } = Typography

export default function AlertsPage() {
  const [loading, setLoading] = useState(false)
  const [rules, setRules] = useState<AlertRule[]>([])
  const [events, setEvents] = useState<AlertEvent[]>([])
  const [eventsTotal, setEventsTotal] = useState(0)
  const [eventsParams, setEventsParams] = useState<AlertEventListParams>({
    page: 1,
    page_size: 20,
    status: undefined,
  })
  const [createModalOpen, setCreateModalOpen] = useState(false)
  const [form] = Form.useForm()

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const [rulesRes, eventsRes] = await Promise.all([
        alertApi.listRules({ page: 1, page_size: 100 }),
        alertApi.listEvents(eventsParams),
      ])
      setRules(rulesRes.data)
      setEvents(eventsRes.data)
      setEventsTotal(eventsRes.total)
    } catch (error) {
      console.error('Failed to fetch alerts:', error)
    } finally {
      setLoading(false)
    }
  }, [eventsParams])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  const handleAcknowledge = async (eventId: string) => {
    try {
      await alertApi.acknowledgeEvent(eventId)
      message.success('告警已确认')
      fetchData()
    } catch {
      message.error('确认失败')
    }
  }

  const handleCreateRule = async (values: CreateAlertRuleRequest) => {
    try {
      await alertApi.createRule(values)
      message.success('规则创建成功')
      setCreateModalOpen(false)
      form.resetFields()
      fetchData()
    } catch {
      message.error('创建失败')
    }
  }

  const ruleColumns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '指标类型',
      dataIndex: 'metric_type',
      key: 'metric_type',
      render: (type: string) => <Tag>{type.toUpperCase()}</Tag>,
    },
    {
      title: '阈值',
      dataIndex: 'threshold',
      key: 'threshold',
      render: (value: number) => `${value}%`,
    },
    {
      title: '持续时间',
      dataIndex: 'duration',
      key: 'duration',
      render: (value: number) => `${value}秒`,
    },
    {
      title: '严重程度',
      dataIndex: 'severity',
      key: 'severity',
      render: (severity: string) => (
        <Tag color={severity === 'critical' ? 'error' : severity === 'warning' ? 'warning' : 'info'}>
          {severity === 'critical' ? '严重' : severity === 'warning' ? '警告' : '信息'}
        </Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      key: 'enabled',
      render: (enabled: boolean) => (
        <Tag color={enabled ? 'success' : 'default'}>{enabled ? '启用' : '禁用'}</Tag>
      ),
    },
  ]

  const eventColumns = [
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
      title: '规则',
      dataIndex: 'rule_name',
      key: 'rule_name',
    },
    {
      title: 'Agent',
      dataIndex: 'agent_name',
      key: 'agent_name',
    },
    {
      title: '指标值',
      dataIndex: 'metric_value',
      key: 'metric_value',
      render: (value: number, record: AlertEvent) => (
        <Text>
          {value.toFixed(1)}% (阈值: {record.threshold}%)
        </Text>
      ),
    },
    {
      title: '消息',
      dataIndex: 'message',
      key: 'message',
      ellipsis: true,
    },
    {
      title: '触发时间',
      dataIndex: 'triggered_at',
      key: 'triggered_at',
      width: 140,
      render: (time: string) => formatRelativeTime(time),
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      render: (_: unknown, record: AlertEvent) => (
        record.status === 'pending' && (
          <Button
            type="link"
            size="small"
            icon={<CheckOutlined />}
            onClick={() => handleAcknowledge(record.id)}
          >
            确认
          </Button>
        )
      ),
    },
  ]

  const items = [
    {
      key: 'events',
      label: `告警事件 (${eventsTotal})`,
      children: (
        <>
          <Space style={{ marginBottom: 16 }}>
            <Select
              placeholder="状态筛选"
              allowClear
              style={{ width: 120 }}
              onChange={(value) => setEventsParams({ ...eventsParams, page: 1, status: value })}
              options={[
                { value: 'pending', label: '待处理' },
                { value: 'acknowledged', label: '已确认' },
                { value: 'resolved', label: '已解决' },
              ]}
            />
          </Space>
          <Table
            className={styles.table}
            columns={eventColumns}
            dataSource={events}
            rowKey="id"
            loading={loading}
            pagination={{
              current: eventsParams.page,
              pageSize: eventsParams.page_size,
              total: eventsTotal,
              showSizeChanger: true,
              showTotal: (total) => `共 ${total} 条`,
            }}
            onChange={(pagination) => setEventsParams({
              ...eventsParams,
              page: pagination.current || 1,
              page_size: pagination.pageSize || 20,
            })}
          />
        </>
      ),
    },
    {
      key: 'rules',
      label: `告警规则 (${rules.length})`,
      children: (
        <>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            style={{ marginBottom: 16 }}
            onClick={() => setCreateModalOpen(true)}
          >
            创建规则
          </Button>
          <Table
            className={styles.table}
            columns={ruleColumns}
            dataSource={rules}
            rowKey="id"
            loading={loading}
            pagination={false}
          />
        </>
      ),
    },
  ]

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <Title level={4} style={{ margin: 0 }}>告警管理</Title>
        <Button
          icon={<ReloadOutlined />}
          onClick={fetchData}
          className={styles.refreshButton}
        >
          刷新
        </Button>
      </div>

      <Card className={styles.card}>
        <Tabs className={styles.tabs} items={items} />
      </Card>

      <Modal
        className={styles.modal}
        title="创建告警规则"
        open={createModalOpen}
        onOk={() => form.submit()}
        onCancel={() => setCreateModalOpen(false)}
        width={500}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleCreateRule}
          initialValues={{ metric_type: 'cpu', condition: '>', duration: 60, severity: 'warning', enabled: true }}
        >
          <Form.Item name="name" label="规则名称" rules={[{ required: true }]}>
            <Input placeholder="输入规则名称" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea placeholder="输入规则描述" />
          </Form.Item>
          <Space style={{ width: '100%' }}>
            <Form.Item name="metric_type" label="指标类型" style={{ width: 120 }}>
              <Select options={[
                { value: 'cpu', label: 'CPU' },
                { value: 'memory', label: '内存' },
                { value: 'disk', label: '磁盘' },
              ]} />
            </Form.Item>
            <Form.Item name="threshold" label="阈值 (%)" rules={[{ required: true }]}>
              <InputNumber min={0} max={100} style={{ width: 100 }} />
            </Form.Item>
            <Form.Item name="duration" label="持续时间 (秒)">
              <InputNumber min={0} style={{ width: 100 }} />
            </Form.Item>
          </Space>
          <Form.Item name="severity" label="严重程度">
            <Select options={[
              { value: 'critical', label: '严重' },
              { value: 'warning', label: '警告' },
              { value: 'info', label: '信息' },
            ]} />
          </Form.Item>
          <Form.Item name="enabled" label="启用" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
