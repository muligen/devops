import { Card, Row, Col, Typography } from 'antd'
import {
  DesktopOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  AlertOutlined,
} from '@ant-design/icons'
import type { DashboardStats } from '@/types'
import styles from './StatCards.module.css'

const { Text, Title } = Typography

interface StatCardsProps {
  stats: DashboardStats | null
}

export default function StatCards({ stats }: StatCardsProps) {
  const statItems = [
    {
      title: '在线 Agent',
      value: stats?.online_agents || 0,
      total: stats?.total_agents || 0,
      icon: <DesktopOutlined style={{ fontSize: 22 }} />,
      iconClass: styles.iconSuccess,
      valueClass: styles.valueSuccess,
    },
    {
      title: '离线 Agent',
      value: stats?.offline_agents || 0,
      icon: <CloseCircleOutlined style={{ fontSize: 22 }} />,
      iconClass: styles.iconError,
      valueClass: styles.valueError,
    },
    {
      title: '今日任务',
      value: stats?.completed_tasks || 0,
      suffix: `完成 / ${stats?.failed_tasks || 0} 失败`,
      icon: <CheckCircleOutlined style={{ fontSize: 22 }} />,
      iconClass: styles.iconPrimary,
      valueClass: styles.valuePrimary,
    },
    {
      title: '待处理告警',
      value: stats?.pending_alerts || 0,
      icon: <AlertOutlined style={{ fontSize: 22 }} />,
      iconClass: styles.iconWarning,
      valueClass: styles.valueWarning,
    },
  ]

  return (
    <Row gutter={[16, 16]}>
      {statItems.map((item, index) => (
        <Col xs={24} sm={12} lg={6} key={index}>
          <Card className={styles.statCard} bordered={false}>
            <div className={styles.statContent}>
              <div className={`${styles.statIcon} ${item.iconClass}`}>
                {item.icon}
              </div>
              <div className={styles.statInfo}>
                <Text type="secondary" className={styles.statTitle}>
                  {item.title}
                </Text>
                <Title level={3} className={`${styles.statValue} ${item.valueClass}`}>
                  {item.value}
                  {item.total && <span className={styles.statTotal}>/ {item.total}</span>}
                </Title>
                {item.suffix && (
                  <Text type="secondary" className={styles.statSuffix}>
                    {item.suffix}
                  </Text>
                )}
              </div>
            </div>
          </Card>
        </Col>
      ))}
    </Row>
  )
}
