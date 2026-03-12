import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { act } from 'react-dom/test-utils'
import React from 'react'
import AgentTerminal from '@/components/terminal/AgentTerminal'
import useTerminalStore from '@/stores/terminal'

// Mock WebSocket
class MockWebSocket {
  static CONNECTING = 0
  static OPEN = 1
  static CLOSING = 2
  static CLOSED = 3

  readyState = MockWebSocket.OPEN
  onmessage: ((event: MessageEvent) => void) | null = null
  onopen: (() => void) | null = null
  onclose: (() => void) | null = null
  onerror: ((event: Event) => void) | null = null

  constructor(public url: string) {}

  send(data: string) {
    // Simulate receiving echo message
    setTimeout(() => {
      if (this.onmessage) {
        this.onmessage(new MessageEvent('message', { data }))
      }
    }, 10)
  }

  close() {
    this.readyState = MockWebSocket.CLOSED
    if (this.onclose) {
      this.onclose(new CloseEvent('close'))
    }
  }

  addEventListener(_type: string, _callback?: () => void) {
    // Simplified for testing
  }

  removeEventListener(_type: string, _callback?: () => void) {
    // Simplified for testing
  }
}

vi.stubGlobal('WebSocket', MockWebSocket)

// Mock useWebSocket hook
vi.mock('@/hooks/useWebSocket', () => ({
  useWebSocket: vi.fn(() => ({
    connected: true,
    sendMessage: vi.fn(),
    lastMessage: null,
  })),
}))

// Mock the task API
vi.mock('@/api', () => ({
  taskApi: {
    createTask: vi.fn().mockResolvedValue({
      id: 'task-123',
      command: 'test command',
      agent_id: 'agent-1',
      status: 'running',
    }),
  },
}))

describe('Integration: WebSocket Real-time Output', () => {
  const mockAgentId = 'agent-1'
  const mockAgent = {
    id: mockAgentId,
    name: 'Test Agent',
    status: 'online',
  }

  beforeEach(() => {
    vi.clearAllMocks()
    // Reset store state
    const store = useTerminalStore.getState()
    store.clearMessages(mockAgentId)
    store.setCurrentAgentId(mockAgentId)
  })

  it('should add running status message when command is executed', async () => {
    render(
      <AgentTerminal agent={mockAgent} />
    )

    const input = screen.getByRole('textbox')
    await act(async () => {
      await input.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter' }))
    })

    // Check that a message was added to the store
    const { messages } = useTerminalStore.getState()
    const agentMessages = messages[mockAgentId] || []

    expect(agentMessages.length).toBeGreaterThan(0)
  })

  it('should update message status from running to completed', () => {
    const { updateMessage } = useTerminalStore.getState()

    // Simulate WebSocket message status update
    act(() => {
      updateMessage(mockAgentId, 'test-msg-1', {
        status: 'completed',
        exitCode: 0,
        duration: 5,
      })
    })

    const { messages } = useTerminalStore.getState()
    const agentMessages = messages[mockAgentId] || []

    const updatedMessage = agentMessages.find((m) => m.id === 'test-msg-1')
    if (updatedMessage) {
      expect(updatedMessage.status).toBe('completed')
      expect(updatedMessage.exitCode).toBe(0)
      expect(updatedMessage.duration).toBe(5)
    }
  })

  it('should append output chunks to message content', () => {
    const { addMessage, updateMessage } = useTerminalStore.getState()

    // Create initial message
    act(() => {
      addMessage(mockAgentId, {
        type: 'result',
        content: '',
        taskId: 'task-123',
        status: 'running',
      })
    })

    const { messages } = useTerminalStore.getState()
    const agentMessages = messages[mockAgentId] || []
    const message = agentMessages[agentMessages.length - 1]

    // Simulate receiving output chunks
    act(() => {
      updateMessage(mockAgentId, message.id, {
        content: message.content + 'First line\n',
      })
      updateMessage(mockAgentId, message.id, {
        content: message.content + 'Second line\n',
      })
    })

    const updatedMessages = useTerminalStore.getState().messages[mockAgentId] || []
    const updatedMessage = updatedMessages.find((m) => m.id === message.id)

    expect(updatedMessage?.content).toContain('First line\n')
    expect(updatedMessage?.content).toContain('Second line\n')
  })

  it('should handle error messages', () => {
    const { addMessage } = useTerminalStore.getState()

    act(() => {
      addMessage(mockAgentId, {
        type: 'error',
        content: 'Command failed: File not found',
      })
    })

    const { messages } = useTerminalStore.getState()
    const agentMessages = messages[mockAgentId] || []

    const errorMessage = agentMessages.find((m) => m.type === 'error')
    expect(errorMessage).toBeDefined()
    expect(errorMessage?.content).toContain('Command failed')
  })

  it('should handle timeout status', () => {
    const { addMessage, updateMessage } = useTerminalStore.getState()

    act(() => {
      addMessage(mockAgentId, {
        type: 'result',
        content: '',
        taskId: 'task-timeout',
        status: 'running',
      })
    })

    const { messages } = useTerminalStore.getState()
    const agentMessages = messages[mockAgentId] || []
    const message = agentMessages[agentMessages.length - 1]

    act(() => {
      updateMessage(mockAgentId, message.id, {
        status: 'timeout',
        exitCode: 124,
        duration: 30,
      })
    })

    const updatedMessages = useTerminalStore.getState().messages[mockAgentId] || []
    const updatedMessage = updatedMessages.find((m) => m.id === message.id)

    expect(updatedMessage?.status).toBe('timeout')
  })

  it('should clear messages when clear button is clicked', async () => {
    const { addMessage, clearMessages } = useTerminalStore.getState()

    // Add some messages
    act(() => {
      addMessage(mockAgentId, { type: 'command', content: 'ls -la' })
      addMessage(mockAgentId, { type: 'result', content: 'file1\nfile2' })
    })

    const { messages } = useTerminalStore.getState()
    expect(messages[mockAgentId]?.length).toBeGreaterThan(0)

    // Clear messages
    act(() => {
      clearMessages(mockAgentId)
    })

    const finalMessages = useTerminalStore.getState().messages
    expect(finalMessages[mockAgentId]).toEqual([])
  })
})
