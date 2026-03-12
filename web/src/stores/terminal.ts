import { create } from 'zustand'

export interface TerminalMessage {
  id: string
  type: 'command' | 'result' | 'error'
  content: string
  timestamp: number
  taskId?: string
  status?: 'running' | 'completed' | 'failed' | 'timeout'
  exitCode?: number
  duration?: number
}

export interface QuickCommand {
  id: string
  name: string
  command: string
  isCustom: boolean
}

interface TerminalState {
  // Per-agent storage
  messages: Record<string, TerminalMessage[]>
  commandHistory: Record<string, string[]>
  quickCommands: QuickCommand[]
  historyIndex: Record<string, number>

  // Current agent being managed
  currentAgentId: string | null

  // Actions
  setCurrentAgentId: (agentId: string | null) => void
  addMessage: (agentId: string, message: Omit<TerminalMessage, 'id' | 'timestamp'>) => void
  updateMessage: (agentId: string, messageId: string, updates: Partial<TerminalMessage>) => void
  clearMessages: (agentId: string) => void

  addCommandToHistory: (agentId: string, command: string) => void
  navigateHistory: (agentId: string, direction: 'up' | 'down') => string | null
  resetHistoryIndex: (agentId: string) => void

  // Quick commands
  addQuickCommand: (command: QuickCommand) => void
  removeQuickCommand: (id: string) => void
}

export const useTerminalStore = create<TerminalState>((set, get) => ({
  messages: {},
  commandHistory: {},
  quickCommands: [
    { id: '1', name: '列出文件', command: 'ls -la', isCustom: false },
    { id: '2', name: '查看进程', command: 'top -b -n 1', isCustom: false },
    { id: '3', name: '查看磁盘', command: 'df -h', isCustom: false },
    { id: '4', name: '查看内存', command: 'free -h', isCustom: false },
    { id: '5', name: '查看网络', command: 'netstat -ano', isCustom: false },
    { id: '6', name: 'Ping 测试', command: 'ping -n 4 8.8.8.8', isCustom: false },
  ],
  historyIndex: {},
  currentAgentId: null,

  setCurrentAgentId: (agentId: string | null) => {
    set({ currentAgentId: agentId })
  },

  addMessage: (agentId: string, message: Omit<TerminalMessage, 'id' | 'timestamp'>) => {
    set((state) => ({
      messages: {
        ...state.messages,
        [agentId]: [
          ...(state.messages[agentId] || []),
          {
            ...message,
            id: `${Date.now()}-${Math.random().toString(36).slice(2, 11)}`,
            timestamp: Date.now(),
          },
        ],
      },
    }))
  },

  updateMessage: (agentId: string, messageId: string, updates: Partial<TerminalMessage>) => {
    set((state) => ({
      messages: {
        ...state.messages,
        [agentId]: (state.messages[agentId] || []).map((msg) =>
          msg.id === messageId ? { ...msg, ...updates } : msg
        ),
      },
    }))
  },

  clearMessages: (agentId: string) => {
    set((state) => ({
      messages: {
        ...state.messages,
        [agentId]: [],
      },
    }))
  },

  addCommandToHistory: (agentId: string, command: string) => {
    set((state) => {
      const history = state.commandHistory[agentId] || []
      const currentIndex = state.historyIndex[agentId] ?? -1
      // Avoid duplicates
      const newHistory = history.includes(command)
        ? [command, ...history.filter((c) => c !== command)]
        : [command, ...history].slice(0, 100) // Limit to 100

      return {
        commandHistory: {
          ...state.commandHistory,
          [agentId]: newHistory,
        },
        historyIndex: {
          ...state.historyIndex,
          [agentId]: currentIndex,
        },
      }
    })
  },

  navigateHistory: (agentId: string, direction: 'up' | 'down') => {
    const history = get().commandHistory[agentId] || []
    const currentIndex = get().historyIndex[agentId] ?? -1

    let newIndex: number
    if (direction === 'up') {
      newIndex = Math.min(currentIndex + 1, history.length - 1)
    } else {
      newIndex = Math.max(currentIndex - 1, -1)
    }

    const command = newIndex >= 0 ? history[history.length - 1 - newIndex] : null

    set((state) => ({
      historyIndex: {
        ...state.historyIndex,
        [agentId]: newIndex,
      },
    }))

    return command
  },

  resetHistoryIndex: (agentId: string) => {
    set((state) => ({
      historyIndex: {
        ...state.historyIndex,
        [agentId]: -1,
      },
    }))
  },

  addQuickCommand: (quickCommand: QuickCommand) => {
    set((state) => ({
      quickCommands: [...state.quickCommands, quickCommand],
    }))
  },

  removeQuickCommand: (id: string) => {
    set((state) => ({
      quickCommands: state.quickCommands.filter((cmd) => cmd.id !== id),
    }))
  },
}))

export default useTerminalStore
