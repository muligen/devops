// Common API response types

export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

export interface PaginatedApiResponse<T> {
  code: number
  message: string
  data: T[]
  pagination: {
    page: number
    page_size: number
    total: number
    total_pages: number
  }
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  page_size: number
}

export interface PaginationParams {
  page?: number
  page_size?: number
}

// List params for filtering

export interface AgentListParams extends PaginationParams {
  status?: 'online' | 'offline'
  search?: string
  sort?: string
  order?: 'asc' | 'desc'
}

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

// Common entity types

export interface User {
  id: string
  username: string
  role: string
  created_at: string
}

export interface Agent {
  id: string
  name: string
  hostname: string
  ip_address: string
  os_info: string
  version: string
  status: 'online' | 'offline' | 'unknown'
  last_seen_at: string | null
  created_at: string
  metadata?: Record<string, unknown>
  cpu_usage?: number
  memory_usage?: number
  disk_usage?: number
}

export interface Task {
  id: string
  agent_id: string
  type: string
  params: Record<string, unknown>
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'
  priority: number
  timeout: number
  result?: Record<string, unknown>
  output?: string
  exit_code?: number | null
  duration?: number | null
  created_by: string
  created_at: string
  started_at?: string | null
  completed_at?: string | null
}

export interface AlertRule {
  id: string
  name: string
  description: string
  metric_type: 'cpu' | 'memory' | 'disk' | 'custom'
  condition: string
  threshold: number
  duration: number
  severity: 'critical' | 'warning' | 'info'
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface AlertEvent {
  id: string
  rule_id: string
  rule_name?: string
  agent_id: string
  agent_name?: string
  metric_value: number
  threshold: number
  status: 'pending' | 'acknowledged' | 'resolved'
  message: string
  triggered_at: string
  resolved_at: string | null
  acknowledged_by: string | null
  acknowledged_at: string | null
}

// Dashboard types

export interface DashboardStats {
  total_agents: number
  online_agents: number
  offline_agents: number
  total_tasks: number
  pending_tasks: number
  running_tasks: number
  completed_tasks: number
  failed_tasks: number
  alerts_triggered: number
  pending_alerts: number
  task_trend?: TaskTrendItem[]
}

export interface TaskTrendItem {
  time: string
  completed: number
  failed: number
}

// WebSocket message types

export interface WebSocketMessage<T = unknown> {
  type: string
  data: T
  timestamp?: number
}

export interface AgentStatusMessage {
  agent_id: string
  status: 'online' | 'offline'
  timestamp: number
}

export interface MetricsMessage {
  [agent_id: string]: {
    cpu_usage: number
    memory_usage: number
    disk_usage: number
    timestamp: number
  }
}

export interface AlertMessage {
  event_id: string
  rule_id: string
  agent_id: string
  metric_value: number
  threshold: number
  status: 'triggered' | 'resolved'
  message: string
  timestamp: number
}
