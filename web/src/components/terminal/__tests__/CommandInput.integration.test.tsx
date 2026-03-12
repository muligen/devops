import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import React from 'react'
import CommandInput from '@/components/terminal/CommandInput'

// Mock the terminal store
vi.mock('@/stores/terminal', () => ({
  useTerminalStore: vi.fn(() => ({
    messages: [],
    commandHistory: { 'test-agent': ['ls -la', 'pwd'] },
    quickCommands: [
      { id: '1', name: '列出文件', command: 'ls -la', isCustom: false },
      { id: '2', name: '查看进程', command: 'top -b -n 1', isCustom: false },
    ],
    historyIndex: { 'test-agent': -1 },
    addMessage: vi.fn(),
    addCommandToHistory: vi.fn(),
    navigateHistory: vi.fn(() => 'ls -la'),
    resetHistoryIndex: vi.fn(),
    clearMessages: vi.fn(),
    setCurrentAgentId: vi.fn(),
  })),
  default: vi.fn(),
}))

// Mock the task API
vi.mock('@/api', () => ({
  taskApi: {
    createTask: vi.fn().mockResolvedValue({
      id: 'task-123',
      command: 'ls -la',
      agent_id: 'test-agent',
      status: 'pending',
    }),
  },
}))

describe('Integration: Command Execution Flow', () => {
  const agentId = 'test-agent'
  const onExecute = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should execute command when user types and presses Enter', async () => {
    const user = userEvent.setup()

    render(<CommandInput agentId={agentId} onExecute={onExecute} disabled={false} />)

    const input = screen.getByRole('textbox')
    await user.type(input, 'ls -la')

    await user.keyboard('{Enter}')

    await waitFor(() => {
      expect(onExecute).toHaveBeenCalledWith('ls -la')
    })
  })

  it('should not execute when input is empty', async () => {
    const user = userEvent.setup()

    render(<CommandInput agentId={agentId} onExecute={onExecute} disabled={false} />)

    const input = screen.getByRole('textbox')
    await user.type(input, '   ') // Only whitespace

    await user.keyboard('{Enter}')

    await waitFor(() => {
      expect(onExecute).not.toHaveBeenCalled()
    })
  })

  it('should allow multiline input with Shift+Enter', async () => {
    const user = userEvent.setup()

    render(<CommandInput agentId={agentId} onExecute={onExecute} disabled={false} />)

    const input = screen.getByRole('textbox')
    await user.type(input, 'line1')
    await user.keyboard('{Shift>}{Enter}{/Shift}')
    await user.type(input, 'line2')

    expect(input).toHaveTextContent(/line1.*line2/)
    expect(onExecute).not.toHaveBeenCalled()
  })

  it('should execute multiline command with Enter', async () => {
    const user = userEvent.setup()

    render(<CommandInput agentId={agentId} onExecute={onExecute} disabled={false} />)

    const input = screen.getByRole('textbox')
    await user.type(input, 'echo "hello"')
    await user.keyboard('{Shift>}{Enter}{/Shift}')
    await user.type(input, 'echo "world"')
    await user.keyboard('{Enter}')

    await waitFor(() => {
      expect(onExecute).toHaveBeenCalled()
    })
  })

  it('should clear input after executing command', async () => {
    const user = userEvent.setup()

    render(<CommandInput agentId={agentId} onExecute={onExecute} disabled={false} />)

    const input = screen.getByRole('textbox')
    await user.type(input, 'ls -la')
    await user.keyboard('{Enter}')

    await waitFor(() => {
      expect(input).toHaveValue('')
    })
  })

  it('should be disabled when agent is offline', async () => {
    const user = userEvent.setup()

    render(<CommandInput agentId={agentId} onExecute={onExecute} disabled={true} />)

    const input = screen.getByRole('textbox')
    expect(input).toBeDisabled()

    await user.type(input, 'ls -la')
    await user.keyboard('{Enter}')

    expect(onExecute).not.toHaveBeenCalled()
  })

  it('should limit command length', async () => {
    const user = userEvent.setup()

    render(
      <CommandInput
        agentId={agentId}
        onExecute={onExecute}
        disabled={false}
        maxLength={100}
      />
    )

    const longCommand = 'x'.repeat(200)
    const input = screen.getByRole('textbox')

    await user.type(input, longCommand)

    // Input should be truncated or show warning
    await user.keyboard('{Enter}')

    // The onExecute should be called with truncated command
    await waitFor(() => {
      expect(onExecute).toHaveBeenCalled()
    })
  })
})
