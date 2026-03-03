import apiClient from './client'
import type { Task, PaginatedApiResponse, ApiResponse, PaginationParams } from '@/types'

export interface TaskListParams extends PaginationParams {
  status?: string
  agent_id?: string
}

export interface CreateTaskRequest {
  agent_id: string
  type: 'exec_shell' | 'init_machine' | 'clean_disk'
  params?: Record<string, unknown>
  timeout?: number
  priority?: number
}

// Legacy interface for backward compatibility
export interface LegacyCreateTaskRequest {
  agent_ids: string[]
  command_type: 'shell' | 'builtin'
  command: string
  timeout?: number
  priority?: number
}

export const taskApi = {
  async list(params: TaskListParams = {}): Promise<{ data: Task[]; total: number }> {
    const response = await apiClient.get<PaginatedApiResponse<Task>>('/tasks', { params })
    return {
      data: response.data.data,
      total: response.data.pagination.total,
    }
  },

  async get(id: string): Promise<Task> {
    const response = await apiClient.get<ApiResponse<Task>>(`/tasks/${id}`)
    return response.data.data
  },

  async create(data: CreateTaskRequest): Promise<Task> {
    const response = await apiClient.post<ApiResponse<Task>>('/tasks', data)
    return response.data.data
  },

  // Legacy method for backward compatibility - converts to new format
  async createLegacy(data: LegacyCreateTaskRequest): Promise<Task[]> {
    const tasks: Task[] = []
    for (const agentId of data.agent_ids) {
      const task = await this.create({
        agent_id: agentId,
        type: 'exec_shell',
        params: {
          command: data.command,
          command_type: data.command_type,
        },
        timeout: data.timeout,
        priority: data.priority,
      })
      tasks.push(task)
    }
    return tasks
  },

  async cancel(id: string): Promise<void> {
    await apiClient.post(`/tasks/${id}/cancel`)
  },
}

export default taskApi
