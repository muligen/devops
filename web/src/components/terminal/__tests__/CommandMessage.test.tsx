import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import CommandMessage from '../CommandMessage'

// Mock the utils
vi.mock('@/utils', () => ({
  formatRelativeTime: vi.fn(() => '刚刚'),
}))

describe('CommandMessage 命令消息组件', () => {
  const mockTimestamp = Date.now()
  const mockOnCopy = vi.fn()

  it('应该正确渲染命令内容', () => {
    render(<CommandMessage content="ls -la" timestamp={mockTimestamp} />)

    expect(screen.getByText('ls -la')).toBeInTheDocument()
  })

  it('当内容为空时应该渲染空命令占位符', () => {
    render(<CommandMessage content="" timestamp={mockTimestamp} />)

    expect(screen.getByText('<空命令>')).toBeInTheDocument()
  })

  it('应该显示时间戳', () => {
    render(<CommandMessage content="ls -la" timestamp={mockTimestamp} />)

    expect(screen.getByText('刚刚')).toBeInTheDocument()
  })

  it('当提供 onCopy 回调时应该显示复制按钮', () => {
    render(<CommandMessage content="ls -la" timestamp={mockTimestamp} onCopy={mockOnCopy} />)

    expect(screen.getByText('复制')).toBeInTheDocument()
  })

  it('当未提供 onCopy 回调时不应显示复制按钮', () => {
    render(<CommandMessage content="ls -la" timestamp={mockTimestamp} />)

    expect(screen.queryByText('复制')).not.toBeInTheDocument()
  })

  it('点击复制按钮时应该调用 onCopy', () => {
    render(<CommandMessage content="ls -la" timestamp={mockTimestamp} onCopy={mockOnCopy} />)

    const copyButton = screen.getByText('复制')
    fireEvent.click(copyButton)

    expect(mockOnCopy).toHaveBeenCalledTimes(1)
  })

  it('用户消息应该有正确的 CSS 类名', () => {
    const { container } = render(<CommandMessage content="ls -la" timestamp={mockTimestamp} />)

    const messageDiv = container.querySelector('.message.user')
    expect(messageDiv).toBeInTheDocument()
  })

  it('命令内容应该有 title 属性', () => {
    const { container } = render(<CommandMessage content="ls -la" timestamp={mockTimestamp} />)

    const bubble = container.querySelector('.messageBubble')
    expect(bubble).toHaveAttribute('title', 'ls -la')
  })
})
