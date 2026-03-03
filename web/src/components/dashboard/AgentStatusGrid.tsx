import { Card, Row, Col, Progress, Tag, Typography, Empty } from 'antd'
import { WindowsOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { formatRelativeTime, formatPercent } from '@/utils'
import type { Agent } from '@/types'
import styles from './AgentStatusGrid.module.css'

const { Text } = Typography

interface AgentStatusGridProps {
  agents: Agent[]
  loading?: boolean
}

export default function AgentStatusGrid({ agents, loading }: AgentStatusGridProps) {
  const navigate = useNavigate()

  if (!agents.length) {
    return (
      <Card title="Agent 状态" loading={loading}>
        <Empty description="暂无 Agent" />
      </Card>
    )
  }

  return (
    <Card title="Agent 状态" loading={loading} className={styles.card}>
      <Row gutter={[12, 12]}>
        {agents.slice(0, 12).map((agent) => (
          <Col xs={12} sm={8} md={6} lg={4} key={agent.id}>
            <div
              className={styles.agentCard}
              onClick={() => navigate(`/agents/${agent.id}`)}
            >
              <div className={styles.agentHeader}>
                <div className={styles.agentIcon}>
                  <WindowsOutlined />
                </div>
                <Tag
                  color={agent.status === 'online' ? 'success' : 'error'}
                  className={styles.statusTag}
                >
                  {agent.status === 'online' ? '在线' : '离线'}
                </Tag>
              </div>
              <Text className={styles.agentName} ellipsis title={agent.name}>
                {agent.name}
              </Text>
              <Text type="secondary" className={styles.agentHostname} ellipsis>
                {agent.hostname}
              </Text>
              {agent.status === 'online' && (
                <div className={styles.metrics}>
                  <div className={styles.metric}>
                    <Text type="secondary" className={styles.metricLabel}>CPU</Text>
                    <Progress
                      percent={agent.cpu_usage || 0}
                      size="small"
                      showInfo={false}
                      strokeColor={
                        (agent.cpu_usage || 0) > 80 ? '#ff4d4f' :
                        (agent.cpu_usage || 0) > 60 ? '#faad14' : '#52c41a'
                      }
                    />
                    <Text className={styles.metricValue}>
                      {formatPercent(agent.cpu_usage, 0)}
                    </Text>
                  </div>
                  <div className={styles.metric}>
                    <Text type="secondary" className={styles.metricLabel}>内存</Text>
                    <Progress
                      percent={agent.memory_usage || 0}
                      size="small"
                      showInfo={false}
                      strokeColor={
                        (agent.memory_usage || 0) > 80 ? '#ff4d4f' :
                        (agent.memory_usage || 0) > 60 ? '#faad14' : '#52c41a'
                      }
                    />
                    <Text className={styles.metricValue}>
                      {formatPercent(agent.memory_usage, 0)}
                    </Text>
                  </div>
                </div>
              )}
              {agent.status === 'offline' && (
                <Text type="secondary" className={styles.lastSeen}>
                  离线: {formatRelativeTime(agent.last_seen_at)}
                </Text>
              )}
            </div>
          </Col>
        ))}
      </Row>
    </Card>
  )
}
