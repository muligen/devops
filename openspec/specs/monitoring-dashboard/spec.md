## Requirements

### Requirement: Dashboard 统计 API
系统 SHALL 提供增强的 Dashboard 统计 API。

#### Scenario: 基础统计
- **WHEN** 前端请求 GET /api/v1/dashboard/stats
- **THEN** 系统返回 Agent 总数、在线数、离线数
- **AND** 返回任务总数、待执行数、运行中数、成功数、失败数
- **AND** 返回未处理告警数

#### Scenario: 趋势数据
- **WHEN** 前端请求 GET /api/v1/dashboard/stats 并携带 include_trend=true 参数
- **THEN** 系统额外返回最近 24 小时的任务执行趋势数据
- **AND** 返回最近 24 小时的告警触发趋势数据

### Requirement: Dashboard 总览实时推送
系统 SHALL 通过 WebSocket 推送 Dashboard 总览更新。

#### Scenario: 定期推送统计更新
- **WHEN** 前端订阅 Dashboard WebSocket
- **THEN** 系统每 10 秒推送最新的统计数据快照

### Requirement: 按资源排序 Agent 列表
系统 SHALL 支持按资源使用情况排序 Agent 列表。

#### Scenario: 按 CPU 排序
- **WHEN** 前端请求 GET /api/v1/agents?sort=cpu_usage&order=desc
- **THEN** 系统返回按 CPU 使用率降序排列的 Agent 列表

#### Scenario: 按内存排序
- **WHEN** 前端请求 GET /api/v1/agents?sort=memory_percent&order=desc
- **THEN** 系统返回按内存使用率降序排列的 Agent 列表

#### Scenario: 按磁盘排序
- **WHEN** 前端请求 GET /api/v1/agents?sort=disk_percent&order=desc
- **THEN** 系统返回按磁盘使用率降序排列的 Agent 列表

### Requirement: 前端 WebSocket 端点
系统 SHALL 提供独立的前端 WebSocket 端点。

#### Scenario: 前端订阅连接
- **WHEN** 前端连接 WS /api/v1/ws/dashboard 并携带有效 JWT
- **THEN** 系统建立推送连接
- **AND** 连接独立于 Agent WebSocket 连接

### Requirement: Server provides Agent metrics history

The Server SHALL provide Agent metrics history API.

#### Scenario: Query metrics history
- **WHEN** user requests metrics for Agent
- **THEN** Server returns metrics for specified time range
- **AND** supports resolution: raw (1min), hourly, daily
- **AND** default time range is last 24 hours

#### Scenario: Metrics aggregation
- **WHEN** user requests aggregated metrics
- **THEN** Server returns aggregated values:
  - average CPU usage
  - peak CPU usage
  - average memory usage
  - peak memory usage
  - average disk usage

### Requirement: Server provides task statistics

The Server SHALL provide task statistics API.

#### Scenario: Task statistics by Agent
- **WHEN** user requests task stats by Agent
- **THEN** Server returns:
  - tasks per Agent (24h)
  - success rate per Agent
  - average execution time per Agent

#### Scenario: Task statistics by type
- **WHEN** user requests task stats by type
- **THEN** Server returns:
  - tasks per command type
  - success rate per type
  - average execution time per type

### Requirement: Server supports alerting rules

The Server SHALL support alerting rules configuration.

#### Scenario: Create alert rule
- **WHEN** admin creates alert rule
- **THEN** Server stores rule with conditions:
  - metric type (cpu, memory, disk, agent_offline)
  - threshold value
  - duration
  - notification target

#### Scenario: Alert evaluation
- **WHEN** metric crosses threshold for specified duration
- **THEN** Server triggers alert
- **AND** Server sends notification
- **AND** Server logs alert event

#### Scenario: Alert types
- **WHEN** configuring alert rules
- **THEN** supported alert types are:
  - CPU usage exceeds threshold
  - Memory usage exceeds threshold
  - Disk usage exceeds threshold
  - Agent offline for extended period

### Requirement: Server provides alert history

The Server SHALL provide alert history API.

#### Scenario: Query alerts
- **WHEN** user requests alert history
- **THEN** Server returns alerts filtered by:
  - Agent ID
  - Alert type
  - Time range
  - Status (active, resolved)

#### Scenario: Alert acknowledgment
- **WHEN** user acknowledges alert
- **THEN** Server marks alert as acknowledged
- **AND** Server records acknowledged_by user

### Requirement: Server supports notification channels

The Server SHALL support multiple notification channels.

#### Scenario: Webhook notification
- **WHEN** alert triggers with webhook configured
- **THEN** Server sends POST request to webhook URL
- **AND** request contains alert details (JSON)

#### Scenario: Email notification
- **WHEN** alert triggers with email configured
- **THEN** Server sends email to configured addresses
- **AND** email contains alert details

### Requirement: Server provides system health API

The Server SHALL provide system health API.

#### Scenario: Health check
- **WHEN** health check endpoint is called
- **THEN** Server returns status of:
  - database connection
  - Redis connection
  - RabbitMQ connection
  - object storage connection
- **AND** returns 200 if all healthy
- **AND** returns 503 if any unhealthy

#### Scenario: Metrics endpoint
- **WHEN** Prometheus scrapes metrics endpoint
- **THEN** Server returns metrics in Prometheus format
- **AND** includes:
  - connected Agents gauge
  - task execution histogram
  - request latency histogram
