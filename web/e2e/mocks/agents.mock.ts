import type { Agent } from '../../src/types'

export const mockOnlineAgent: Agent = {
  id: 'agent-online-001',
  name: 'Test Agent Online',
  hostname: 'TEST-WIN-001',
  ip_address: '192.168.1.100',
  os_info: 'Windows 10 Pro',
  version: '1.0.0',
  status: 'online',
  last_seen_at: new Date().toISOString(),
  created_at: '2024-01-01T00:00:00Z',
  cpu_usage: 45.5,
  memory_usage: 62.3,
  disk_usage: 55.0,
}

export const mockOfflineAgent: Agent = {
  id: 'agent-offline-001',
  name: 'Test Agent Offline',
  hostname: 'TEST-WIN-002',
  ip_address: '192.168.1.101',
  os_info: 'Windows Server 2019',
  version: '1.0.0',
  status: 'offline',
  last_seen_at: new Date(Date.now() - 3600000).toISOString(),
  created_at: '2024-01-02T00:00:00Z',
}

export const mockHighCpuAgent: Agent = {
  id: 'agent-high-cpu-001',
  name: 'Test Agent High CPU',
  hostname: 'TEST-WIN-003',
  ip_address: '192.168.1.102',
  os_info: 'Windows 11 Pro',
  version: '1.0.0',
  status: 'online',
  last_seen_at: new Date().toISOString(),
  created_at: '2024-01-03T00:00:00Z',
  cpu_usage: 95.8,
  memory_usage: 88.5,
  disk_usage: 92.0,
}

export const mockAgentList: Agent[] = [
  mockOnlineAgent,
  mockOfflineAgent,
  mockHighCpuAgent,
]

export function createMockAgent(overrides: Partial<Agent> = {}): Agent {
  return {
    ...mockOnlineAgent,
    ...overrides,
  }
}

export function createApiResponse<T>(data: T) {
  return {
    code: 0,
    message: 'success',
    data,
  }
}

export function createPaginatedResponse<T>(data: T[], total: number = data.length) {
  return {
    code: 0,
    message: 'success',
    data,
    pagination: {
      page: 1,
      page_size: 10,
      total,
      total_pages: Math.ceil(total / 10),
    },
  }
}
