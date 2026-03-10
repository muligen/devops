import apiClient from './client'
import type { Agent, PaginatedApiResponse, ApiResponse, PaginationParams } from '@/types'

export interface AgentListParams extends PaginationParams {
  status?: 'online' | 'offline'
  search?: string
  sort?: string
  order?: 'asc' | 'desc'
}

export const agentApi = {
  async list(params: AgentListParams = {}): Promise<{ data: Agent[]; total: number }> {
    const response = await apiClient.get<PaginatedApiResponse<Agent>>('/agents', { params })
    return {
      data: response.data.data,
      total: response.data.pagination.total,
    }
  },

  async get(id: string): Promise<Agent> {
    const response = await apiClient.get<ApiResponse<Agent>>(`/agents/${id}`)
    return response.data.data
  },

  async delete(id: string): Promise<void> {
    await apiClient.delete(`/agents/${id}`)
  },

  async getMetrics(id: string, range: string = '1h'): Promise<unknown> {
    // Adjust limit based on time range for better chart display
    const limitMap: Record<string, number> = {
      '1h': 120,   // ~2 data points per minute
      '24h': 500,  // ~1 data point per 3 minutes
      '7d': 1000,  // ~1 data point per 10 minutes
    }
    const limit = limitMap[range] || 100

    const response = await apiClient.get(`/agents/${id}/metrics`, {
      params: { range, limit },
    })
    return response.data.data
  },
}

export default agentApi
