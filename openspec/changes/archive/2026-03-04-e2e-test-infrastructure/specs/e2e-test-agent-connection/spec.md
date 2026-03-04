## ADDED Requirements

### Requirement: Agent WebSocket 连接
测试 SHALL 验证 Agent 通过 WebSocket 连接到 Server 的完整流程。

#### Scenario: 成功建立连接
- **WHEN** Agent 使用有效配置连接 Server WebSocket 端点
- **THEN** 连接成功建立，状态变为 connecting

#### Scenario: 连接失败重试
- **WHEN** Server 不可用
- **THEN** Agent 按指数退避策略重试连接

### Requirement: Agent 认证流程
测试 SHALL 验证 Agent HMAC challenge-response 认证流程。

#### Scenario: 成功认证
- **WHEN** Agent 发送 auth 请求并正确响应 challenge
- **THEN** Server 返回认证成功，Agent 状态变为 online

#### Scenario: 认证失败
- **WHEN** Agent 响应错误的 challenge
- **THEN** Server 拒绝连接并返回认证失败原因

#### Scenario: 无效 Agent ID
- **WHEN** Agent 发送不存在的 agent_id
- **THEN** Server 返回 agent not found 错误

### Requirement: 连接断开处理
测试 SHALL 验证连接断开后的状态同步。

#### Scenario: Agent 主动断开
- **WHEN** Agent 关闭 WebSocket 连接
- **THEN** Server 更新 Agent 状态为 offline

#### Scenario: 网络异常断开
- **WHEN** 网络异常导致连接断开
- **THEN** Server 在超时后更新 Agent 状态为 offline
