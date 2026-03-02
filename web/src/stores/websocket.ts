import { create } from 'zustand'

interface WebSocketState {
  isConnected: boolean
  reconnectAttempts: number
  error: string | null

  // Actions
  setConnected: (connected: boolean) => void
  incrementReconnectAttempts: () => void
  resetReconnectAttempts: () => void
  setError: (error: string | null) => void
}

export const useWebSocketStore = create<WebSocketState>((set) => ({
  isConnected: false,
  reconnectAttempts: 0,
  error: null,

  setConnected: (connected: boolean) => {
    set({
      isConnected: connected,
      error: connected ? null : 'WebSocket 连接断开',
    })
  },

  incrementReconnectAttempts: () => {
    set((state) => ({ reconnectAttempts: state.reconnectAttempts + 1 }))
  },

  resetReconnectAttempts: () => {
    set({ reconnectAttempts: 0 })
  },

  setError: (error: string | null) => {
    set({ error })
  },
}))

export default useWebSocketStore
