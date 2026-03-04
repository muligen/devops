## ADDED Requirements

### Requirement: 监控仪表盘数据
测试 SHALL 验证监控仪表盘数据聚合 API。

#### Scenario: 获取仪表盘概览
- **WHEN** 请求 GET /api/v1/dashboard/overview
- **THEN** 返回 Agent 总数、在线数、任务统计

#### Scenario: 获取告警列表
- **WHEN** 请求 GET /api/v1/dashboard/alerts
- **THEN** 返回最近的告警列表

### Requirement: 实时数据推送
测试 SHALL 验证 WebSocket 实时数据推送。

#### Scenario: 订阅 Agent 事件
- **WHEN** 前端连接 WebSocket 并订阅 agent_events
- **THEN** 收到 Agent 上下线事件推送

#### Scenario: 订阅指标更新
- **WHEN** 订阅 metrics_update
- **THEN** 收到 Agent 指标实时更新

#### Scenario: 订阅任务事件
- **WHEN** 订阅 task_events
- **THEN** 收到任务状态变更事件

### Requirement: 历史数据查询
测试 SHALL 验证历史数据查询 API。

#### Scenario: 查询历史指标
- **WHEN** 请求 GET /api/v1/agents/{id}/metrics/history
- **THEN** 返回指定时间范围的历史指标数据

#### Scenario: 查询操作日志
- **WHEN** 请求 GET /api/v1/audit/logs
- **THEN** 返回操作审计日志
