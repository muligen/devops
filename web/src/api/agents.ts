import apiClient from './client'
import type { Agent, PaginatedResponse, ApiResponse, PaginationParams } from '@/types'

export interface AgentListParams extends PaginationParams {
  status?: 'online' | 'offline'
  search?: string
  sort?: string
  order?: 'asc' | 'desc'
}

export const agentApi = {
  async list(params: AgentListParams = {}): Promise<PaginatedResponse<Agent>> {
    const response = await apiClient.get<PaginatedResponse<Agent>>('/agents', { params })
    return response.data
  },

  async get(id: string): Promise<Agent> {
    const response = await apiClient.get<ApiResponse<Agent>>(`/agents/${id}`)
    return response.data.data
  },

  async delete(id: string): Promise<void> {
    await apiClient.delete(`/agents/${id}`)
  },

  async getMetrics(id: string, range: string = '1h'): Promise<unknown> {
    const response = await apiClient.get(`/agents/${id}/metrics`, {
      params: { range },
    })
    return response.data
  },
}

export default agentApi
