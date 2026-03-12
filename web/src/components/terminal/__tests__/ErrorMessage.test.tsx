import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import ErrorMessage from '../ErrorMessage'

// Mock the utils
vi.mock('@/utils', () => ({
  formatRelativeTime: vi.fn(() => '刚刚'),
}))

describe('ErrorMessage 错误消息组件', () => {
  const mockTimestamp = Date.now()

  it('应该正确渲染错误内容', () => {
    render(<ErrorMessage content="Command failed" timestamp={mockTimestamp} />)

    expect(screen.getByText('Command failed')).toBeInTheDocument()
  })

  it('应该显示时间戳', () => {
    render(<ErrorMessage content="Command failed" timestamp={mockTimestamp} />)

    expect(screen.getByText('刚刚')).toBeInTheDocument()
  })

  it('Agent 错误消息应该有正确的 CSS 类名', () => {
    const { container } = render(<ErrorMessage content="Command failed" timestamp={mockTimestamp} />)

    const messageDiv = container.querySelector('.message.agent.error')
    expect(messageDiv).toBeInTheDocument()
  })

  it('应该渲染错误图标', () => {
    const { container } = render(<ErrorMessage content="Command failed" timestamp={mockTimestamp} />)

    const avatar = container.querySelector('.messageAvatar')
    expect(avatar).toBeInTheDocument()
    expect(avatar?.innerHTML).toContain('svg')
  })
})
