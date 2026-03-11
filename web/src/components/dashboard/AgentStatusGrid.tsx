import { Card, Tag, Typography, Empty } from 'antd'
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
      <Card
        title="Agent 状态"
        loading={loading}
        style={{
          background: 'rgba(26, 26, 36, 0.6)',
          border: '1px solid rgba(255, 255, 255, 0.06)',
          borderRadius: 12,
        }}
      >
        <Empty description="暂无 Agent" />
      </Card>
    )
  }

  return (
    <Card
      title="Agent 状态"
      loading={loading}
      className={styles.card}
    >
      <div className={styles.grid}>
        {agents.slice(0, 12).map((agent) => (
          <div
            key={agent.id}
            className={`${styles.agentCard} ${agent.status}`}
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
                <div className={`${styles.metric} cpu`}>
                  <Text type="secondary" className={styles.metricLabel}>CPU</Text>
                  <div className={styles.progressBar}>
                    <div
                      className={styles.progress}
                      style={{
                        width: `${agent.cpu_usage || 0}%`,
                        background: (agent.cpu_usage || 0) > 80 ? '#ff7875' :
                                   (agent.cpu_usage || 0) > 60 ? '#ffc53d' : '#73d13d',
                      }}
                    />
                  </div>
                  <Text className={styles.metricValue}>
                    {formatPercent(agent.cpu_usage, 0)}
                  </Text>
                </div>
                <div className={`${styles.metric} memory`}>
                  <Text type="secondary" className={styles.metricLabel}>内存</Text>
                  <div className={styles.progressBar}>
                    <div
                      className={styles.progress}
                      style={{
                        width: `${agent.memory_usage || 0}%`,
                        background: (agent.memory_usage || 0) > 80 ? '#ff7875' :
                                   (agent.memory_usage || 0) > 60 ? '#ffc53d' : '#73d13d',
                      }}
                    />
                  </div>
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
        ))}
      </div>
    </Card>
  )
}
