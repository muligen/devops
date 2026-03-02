import apiClient from './client'
import type { Task, PaginatedResponse, ApiResponse, PaginationParams } from '@/types'

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
  async list(params: TaskListParams = {}): Promise<PaginatedResponse<Task>> {
    const response = await apiClient.get<PaginatedResponse<Task>>('/tasks', { params })
    return response.data
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
