import { useEffect, useRef, useCallback } from 'react'
import { useAuthStore, useWebSocketStore } from '@/stores'
import { useDashboardStore } from '@/stores/dashboard'
import { triggerAlertNotification } from '@/components/common/alertHandler'
import type { WebSocketMessage, DashboardStats, AlertMessage } from '@/types'

const WS_RECONNECT_DELAY_BASE = 1000
const WS_RECONNECT_DELAY_MAX = 30000
const WS_PING_INTERVAL = 30000

export function useWebSocket() {
  const wsRef = useRef<WebSocket | null>(null)
  const pingIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const token = useAuthStore((state) => state.token)
  const { setConnected, incrementReconnectAttempts, resetReconnectAttempts, setError } =
    useWebSocketStore()

  // Define all callback functions first
  const startPing = useCallback(() => {
    pingIntervalRef.current = setInterval(() => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(JSON.stringify({ type: 'ping' }))
      }
    }, WS_PING_INTERVAL)
  }, [])

  const stopPing = useCallback(() => {
    if (pingIntervalRef.current) {
      clearInterval(pingIntervalRef.current)
      pingIntervalRef.current = null
    }
  }, [])

  const clearReconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
      reconnectTimeoutRef.current = null
    }
  }, [])

  const handleMessage = useCallback((message: WebSocketMessage) => {
    const dashboardStore = useDashboardStore.getState()

    switch (message.type) {
      case 'connected':
        console.log('WebSocket connected:', message.data)
        break
      case 'agent_status':
        // Handle agent status change - could update agent list store
        break
      case 'metrics':
        // Handle metrics update
        if (message.data && typeof message.data === 'object') {
          dashboardStore.updateMetrics(message.data as Record<string, unknown>)
        }
        break
      case 'alert':
        // Handle alert event - trigger notification
        if (message.data && typeof message.data === 'object') {
          triggerAlertNotification(message.data as AlertMessage)
        }
        break
      case 'dashboard':
        // Handle dashboard stats update
        if (message.data && typeof message.data === 'object') {
          dashboardStore.setStats(message.data as DashboardStats)
        }
        break
      case 'pong':
        // Pong response
        break
      default:
        console.log('Unknown WebSocket message type:', message.type)
    }
  }, [])

  const scheduleReconnect = useCallback(() => {
    const { reconnectAttempts } = useWebSocketStore.getState()
    const delay = Math.min(
      WS_RECONNECT_DELAY_BASE * Math.pow(2, reconnectAttempts),
      WS_RECONNECT_DELAY_MAX
    )

    incrementReconnectAttempts()

    reconnectTimeoutRef.current = setTimeout(() => {
      connect()
    }, delay)
    // connect is defined below but referenced here for reconnection
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [incrementReconnectAttempts])

  const disconnect = useCallback(() => {
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
    stopPing()
    clearReconnect()
  }, [stopPing, clearReconnect])

  const connect = useCallback(() => {
    if (!token) return

    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsHost = window.location.host
    const wsUrl = `${wsProtocol}//${wsHost}/api/v1/ws/dashboard?token=${token}`

    try {
      const ws = new WebSocket(wsUrl)
      wsRef.current = ws

      ws.onopen = () => {
        setConnected(true)
        resetReconnectAttempts()
        startPing()
      }

      ws.onclose = () => {
        setConnected(false)
        stopPing()
        scheduleReconnect()
      }

      ws.onerror = (error) => {
        console.error('WebSocket error:', error)
        setError('WebSocket 连接错误')
      }

      ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data)
          handleMessage(message)
        } catch (err) {
          console.error('Failed to parse WebSocket message:', err)
        }
      }
    } catch (err) {
      console.error('Failed to create WebSocket:', err)
      setError('无法创建 WebSocket 连接')
    }
  }, [token, setConnected, resetReconnectAttempts, setError, startPing, stopPing, scheduleReconnect, handleMessage])

  useEffect(() => {
    if (token) {
      connect()
    }

    return () => {
      disconnect()
    }
  }, [token, connect, disconnect])

  return {
    isConnected: useWebSocketStore.getState().isConnected,
    connect,
    disconnect,
  }
}

export default useWebSocket
