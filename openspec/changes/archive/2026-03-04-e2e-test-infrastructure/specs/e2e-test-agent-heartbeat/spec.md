## ADDED Requirements

### Requirement: Agent 心跳发送
测试 SHALL 验证 Agent 定期发送心跳消息。

#### Scenario: 定期心跳
- **WHEN** Agent 已连接并认证
- **THEN** Agent 每 10 秒发送一次心跳消息

#### Scenario: 心跳响应
- **WHEN** Server 收到心跳消息
- **THEN** Server 返回心跳确认并更新 last_seen 时间

### Requirement: Server 心跳超时检测
测试 SHALL 验证 Server 检测 Agent 心跳超时。

#### Scenario: 心跳超时
- **WHEN** Agent 超过 60 秒未发送心跳
- **THEN** Server 更新 Agent 状态为 offline

#### Scenario: 心跳恢复
- **WHEN** Agent 恢复发送心跳
- **THEN** Server 更新 Agent 状态为 online
