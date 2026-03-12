/**
 * Test mocks and helpers
 */

// Mock dayjs for testing
vi.mock('dayjs', () => ({
  default: vi.fn(() => ({
    fromNow: vi.fn(() => '刚刚'),
    format: vi.fn(() => '2024-01-01 12:00:00'),
  })),
  extend: vi.fn(),
  locale: vi.fn(),
}))

// Mock format utilities
export const mockFormatRelativeTime = vi.fn((date: string | Date | null) => {
  if (!date) return '-'
  return '刚刚'
})

export const mockFormatDuration = vi.fn((seconds: number) => {
  if (seconds < 60) return `${seconds}秒`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}分${seconds % 60}秒`
  return `${Math.floor(seconds / 3600)}小时`
})

// Create a mock user event
export const createUserEvent = () => ({
  click: vi.fn(),
  type: vi.fn(),
  clear: vi.fn(),
})

// Get a mock message
export const createMockMessage = (overrides = {}) => ({
  id: 'test-message-1',
  type: 'result' as const,
  content: 'test output',
  timestamp: Date.now(),
  status: 'completed' as const,
  exitCode: 0,
  duration: 5,
  ...overrides,
})
