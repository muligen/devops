import type { Task } from '../../src/types'

export const mockPendingTask: Task = {
  id: 'task-pending-001',
  agent_id: 'agent-online-001',
  type: 'exec_shell',
  params: { command: 'echo test', command_type: 'shell' },
  status: 'pending',
  priority: 0,
  timeout: 300,
  created_by: 'admin',
  created_at: new Date().toISOString(),
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

export const mockRunningTask: Task = {
  id: 'task-running-001',
  agent_id: 'agent-online-001',
  type: 'exec_shell',
  params: { command: 'ping -n 4 google.com', command_type: 'shell' },
  status: 'running',
  priority: 0,
  timeout: 300,
  created_by: 'admin',
  created_at: new Date(Date.now() - 60000).toISOString(),
  started_at: new Date(Date.now() - 30000).toISOString(),
}

export const mockCompletedTask: Task = {
  id: 'task-completed-001',
  agent_id: 'agent-online-001',
  type: 'exec_shell',
  params: { command: 'whoami', command_type: 'shell' },
  status: 'completed',
  priority: 0,
  timeout: 300,
  result: { output: 'admin', exit_code: 0 },
  output: 'admin',
  exit_code: 0,
  duration: 150,
  created_by: 'admin',
  created_at: new Date(Date.now() - 3600000).toISOString(),
  started_at: new Date(Date.now() - 3500000).toISOString(),
  completed_at: new Date(Date.now() - 3400000).toISOString(),
}

export const mockFailedTask: Task = {
  id: 'task-failed-001',
  agent_id: 'agent-online-001',
  type: 'exec_shell',
  params: { command: 'invalid-command-xyz', command_type: 'shell' },
  status: 'failed',
  priority: 0,
  timeout: 300,
  result: { output: 'Command not found', exit_code: 1 },
  output: 'Command not found',
  exit_code: 1,
  duration: 50,
  created_by: 'admin',
  created_at: new Date(Date.now() - 7200000).toISOString(),
  started_at: new Date(Date.now() - 7150000).toISOString(),
  completed_at: new Date(Date.now() - 7145000).toISOString(),
}

export const mockTaskList: Task[] = [
  mockCompletedTask,
  mockRunningTask,
  mockPendingTask,
  mockFailedTask,
]

export function createMockTask(overrides: Partial<Task> & { command?: string; command_type?: string } = {}): Task {
  const { command = 'echo test', command_type = 'shell', ...rest } = overrides
  return {
    ...mockPendingTask,
    params: { command, command_type },
    ...rest,
  }
}

export function createTaskListForAgent(agentId: string, count: number = 5): Task[] {
  const statuses: Array<Task['status']> = ['completed', 'running', 'pending', 'failed']
  return Array.from({ length: count }, (_, i) =>
    createMockTask({
      id: `task-${agentId}-${i}`,
      agent_id: agentId,
      command: `command-${i}`,
      command_type: i % 2 === 0 ? 'shell' : 'builtin',
      status: statuses[i % 4],
      created_at: new Date(Date.now() - i * 3600000).toISOString(),
    })
  )
}
