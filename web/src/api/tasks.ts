import apiClient from './client'
import type { Task, PaginatedApiResponse, ApiResponse, PaginationParams } from '@/types'

export interface TaskListParams extends PaginationParams {
  status?: string
  agent_id?: string
}

export interface CreateTaskRequest {
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

  async cancel(id: string): Promise<void> {
    await apiClient.post(`/tasks/${id}/cancel`)
  },
}

export default taskApi
