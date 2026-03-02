import apiClient from './client'
import type { DashboardStats, ApiResponse } from '@/types'

export const dashboardApi = {
  async getStats(): Promise<DashboardStats> {
    const response = await apiClient.get<ApiResponse<DashboardStats>>('/dashboard/stats')
    return response.data.data
  },
}

export default dashboardApi
