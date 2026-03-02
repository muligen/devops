## ADDED Requirements

### Requirement: 前端 WebSocket 连接
系统 SHALL 提供前端 WebSocket 端点用于实时数据推送。

#### Scenario: 建立连接
- **WHEN** 前端发起 WebSocket 连接到 /api/v1/ws/dashboard
- **THEN** 系统验证 JWT Token
- **AND** 建立连接并返回连接成功消息

#### Scenario: 认证失败
- **WHEN** WebSocket 连接携带无效 Token
- **THEN** 系统返回认证失败消息并关闭连接

#### Scenario: 自动重连
- **WHEN** WebSocket 连接断开
- **THEN** 前端以指数退避策略自动重连
- **AND** 显示连接状态指示器

### Requirement: Agent 状态推送
系统 SHALL 实时推送 Agent 状态变更。

#### Scenario: Agent 上线推送
- **WHEN** Agent 连接成功
- **THEN** 系统推送 agent.online 事件到订阅的前端

#### Scenario: Agent 离线推送
- **WHEN** Agent 断开连接
- **THEN** 系统推送 agent.offline 事件到订阅的前端

### Requirement: 指标推送
系统 SHALL 定期推送 Agent 指标数据。

#### Scenario: 指标批量推送
- **WHEN** Agent 上报指标
- **THEN** 系统每 5 秒批量推送指标更新到前端
- **AND** 推送数据包含 Agent ID 和最新指标值

### Requirement: 告警事件推送
系统 SHALL 实时推送告警事件。

#### Scenario: 告警触发推送
- **WHEN** 告警规则触发
- **THEN** 系统推送 alert.triggered 事件到前端

#### Scenario: 告警恢复推送
- **WHEN** 告警条件恢复正常
- **THEN** 系统推送 alert.resolved 事件到前端

### Requirement: 消息格式
WebSocket 消息 SHALL 使用统一的 JSON 格式。

#### Scenario: 消息结构
- **WHEN** 系统推送任何事件
- **THEN** 消息格式为 {"type": "<event-type>", "data": {...}, "timestamp": <unix>}
