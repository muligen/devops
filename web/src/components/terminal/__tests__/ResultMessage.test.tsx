import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import type { TerminalMessage } from '@/stores/terminal'
import ResultMessage from '../ResultMessage'

// Mock the utils
vi.mock('@/utils', () => ({
  formatRelativeTime: vi.fn(() => '刚刚'),
  formatDuration: vi.fn((seconds: number) => `${seconds}秒`),
}))

const createMockMessage = (overrides: Partial<TerminalMessage> = {}): TerminalMessage => ({
  id: 'test-1',
  type: 'result',
  content: 'test output',
  timestamp: Date.now(),
  ...overrides,
})

describe('ResultMessage 结果消息组件', () => {
  const mockOnCopy = vi.fn()
  const mockOnExpand = vi.fn()

  it('应该正确渲染结果内容', () => {
    const message = createMockMessage({ content: 'test output' })
    render(<ResultMessage message={message} />)

    expect(screen.getByText('test output')).toBeInTheDocument()
  })

  it('状态为 running 时应该显示运行状态指示器', () => {
    const message = createMockMessage({ status: 'running', content: '' })
    const { container } = render(<ResultMessage message={message} />)

    expect(container.querySelector('.loadingText')).toHaveTextContent('执行中...')
    expect(container.querySelector('.statusSpinner')).toBeInTheDocument()
    expect(container.querySelector('.statusIndicator.running')).toBeInTheDocument()
  })

  it('状态为 completed 时应该显示完成状态指示器', () => {
    const message = createMockMessage({ status: 'completed', exitCode: 0, duration: 5 })
    const { container } = render(<ResultMessage message={message} />)

    expect(container.querySelector('.statusIndicator.completed')).toBeInTheDocument()
  })

  it('提供 exitCode 时应该显示退出码', () => {
    const message = createMockMessage({ exitCode: 0 })
    render(<ResultMessage message={message} />)

    expect(screen.getByText('退出码: 0')).toBeInTheDocument()
  })

  it('提供 duration 时应该显示执行时长', () => {
    const message = createMockMessage({ duration: 5 })
    render(<ResultMessage message={message} />)

    expect(screen.getByText('耗时: 5秒')).toBeInTheDocument()
  })

  it('当内容较长且未展开时应显示展开按钮', () => {
    const longContent = 'x'.repeat(6000)
    const message = createMockMessage({ content: longContent })
    render(<ResultMessage message={message} onExpand={mockOnExpand} />)

    expect(screen.getByText(/展开完整输出/)).toBeInTheDocument()
  })

  it('短内容不应显示展开按钮', () => {
    const message = createMockMessage({ content: 'short content' })
    render(<ResultMessage message={message} />)

    expect(screen.queryByText(/展开完整输出/)).not.toBeInTheDocument()
  })

  it('当 isExpanded 为 true 时不应显示展开按钮', () => {
    const longContent = 'x'.repeat(6000)
    const message = createMockMessage({ content: longContent })
    render(<ResultMessage message={message} onExpand={mockOnExpand} isExpanded={true} />)

    expect(screen.queryByText(/展开完整输出/)).not.toBeInTheDocument()
  })

  it('点击展开按钮时应该调用 onExpand', () => {
    const longContent = 'x'.repeat(6000)
    const message = createMockMessage({ content: longContent })
    render(<ResultMessage message={message} onExpand={mockOnExpand} />)

    const expandButton = screen.getByText(/展开完整输出/)
    fireEvent.click(expandButton)

    expect(mockOnExpand).toHaveBeenCalledTimes(1)
  })

  it('提供 onCopy 回调时应该显示复制按钮', () => {
    const message = createMockMessage()
    render(<ResultMessage message={message} onCopy={mockOnCopy} />)

    expect(screen.getByText('复制')).toBeInTheDocument()
  })

  it('未提供 onCopy 回调时不应显示复制按钮', () => {
    const message = createMockMessage()
    render(<ResultMessage message={message} />)

    expect(screen.queryByText('复制')).not.toBeInTheDocument()
  })

  it('点击复制按钮时应该调用 onCopy', () => {
    const message = createMockMessage()
    render(<ResultMessage message={message} onCopy={mockOnCopy} />)

    const copyButton = screen.getByText('复制')
    fireEvent.click(copyButton)

    expect(mockOnCopy).toHaveBeenCalledTimes(1)
  })

  it('Agent 消息应该有正确的 CSS 类名', () => {
    const message = createMockMessage()
    const { container } = render(<ResultMessage message={message} />)

    const messageDiv = container.querySelector('.message.agent')
    expect(messageDiv).toBeInTheDocument()
  })

  it('消息类型为 error 时应该添加 error 类', () => {
    const message = createMockMessage({ type: 'error' })
    const { container } = render(<ResultMessage message={message} />)

    const messageDiv = container.querySelector('.message.agent.error')
    expect(messageDiv).toBeInTheDocument()
  })
})
