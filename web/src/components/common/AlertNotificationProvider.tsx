import { useEffect } from 'react'
import { notification } from 'antd'
import { WarningOutlined, CheckCircleOutlined } from '@ant-design/icons'
import { useWebSocket } from '@/hooks'
import { useWebSocketStore } from '@/stores'
import { setAlertHandler, clearAlertHandler } from './alertHandler'
import type { AlertMessage } from '@/types'

interface AlertNotificationProviderProps {
  children: React.ReactNode
}

export default function AlertNotificationProvider({ children }: AlertNotificationProviderProps) {
  const isConnected = useWebSocketStore((state) => state.isConnected)

  // Initialize WebSocket connection at app level (not per-page)
  useWebSocket()

  useEffect(() => {
    setAlertHandler((alert: AlertMessage) => {
      const isResolved = alert.status === 'resolved'

      notification.open({
        message: isResolved ? '告警已恢复' : '新告警',
        description: alert.message,
        icon: isResolved ? (
          <CheckCircleOutlined style={{ color: '#52c41a' }} />
        ) : (
          <WarningOutlined style={{ color: '#faad14' }} />
        ),
        duration: 5,
        placement: 'topRight',
      })
    })

    return () => {
      clearAlertHandler()
    }
  }, [])

  // Connection status notification
  useEffect(() => {
    if (!isConnected) {
      notification.warning({
        key: 'ws-disconnected',
        message: '连接断开',
        description: 'WebSocket 连接已断开，正在尝试重连...',
        duration: 0,
        placement: 'topRight',
      })
    } else {
      notification.destroy('ws-disconnected')
    }
  }, [isConnected])

  return <>{children}</>
}
