import apiClient from './client'
import type { AlertRule, AlertEvent, PaginatedResponse, ApiResponse, PaginationParams } from '@/types'

export interface AlertRuleListParams extends PaginationParams {
  enabled?: boolean
}

export interface AlertEventListParams extends PaginationParams {
  status?: string
  agent_id?: string
}

export interface CreateAlertRuleRequest {
  name: string
  description?: string
  metric_type: 'cpu' | 'memory' | 'disk' | 'custom'
  condition: string
  threshold: number
  duration: number
  severity: 'critical' | 'warning' | 'info'
  enabled?: boolean
}

export const alertApi = {
  // Alert Rules
  async listRules(params: AlertRuleListParams = {}): Promise<PaginatedResponse<AlertRule>> {
    const response = await apiClient.get<PaginatedResponse<AlertRule>>('/alerts/rules', { params })
    return response.data
  },

  async getRule(id: string): Promise<AlertRule> {
    const response = await apiClient.get<ApiResponse<AlertRule>>(`/alerts/rules/${id}`)
    return response.data.data
  },

  async createRule(data: CreateAlertRuleRequest): Promise<AlertRule> {
    const response = await apiClient.post<ApiResponse<AlertRule>>('/alerts/rules', data)
    return response.data.data
  },

  async updateRule(id: string, data: Partial<CreateAlertRuleRequest>): Promise<AlertRule> {
    const response = await apiClient.put<ApiResponse<AlertRule>>(`/alerts/rules/${id}`, data)
    return response.data.data
  },

  async deleteRule(id: string): Promise<void> {
    await apiClient.delete(`/alerts/rules/${id}`)
  },

  // Alert Events (History)
  async listEvents(params: AlertEventListParams = {}): Promise<PaginatedResponse<AlertEvent>> {
    const response = await apiClient.get<PaginatedResponse<AlertEvent>>('/alerts/history', { params })
    return response.data
  },

  async acknowledgeEvent(id: string): Promise<AlertEvent> {
    const response = await apiClient.put<ApiResponse<AlertEvent>>(`/alerts/history/${id}/acknowledge`)
    return response.data.data
  },
}

export default alertApi
