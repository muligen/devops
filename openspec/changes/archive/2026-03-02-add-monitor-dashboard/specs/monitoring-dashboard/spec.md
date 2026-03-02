## MODIFIED Requirements

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

## ADDED Requirements

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
