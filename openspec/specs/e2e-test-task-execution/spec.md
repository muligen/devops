## ADDED Requirements

### Requirement: 任务创建
测试 SHALL 验证通过 API 创建任务并下发到 Agent。

#### Scenario: 创建即时任务
- **WHEN** 操作员提交任务请求
- **THEN** 任务创建并推送到在线 Agent

#### Scenario: Agent 离线时创建任务
- **WHEN** Agent 离线时创建任务
- **THEN** 任务进入 pending 状态等待 Agent 上线

### Requirement: 任务执行
测试 SHALL 验证任务执行和结果上报。

#### Scenario: Shell 命令执行
- **WHEN** Agent 收到 shell 类型任务
- **THEN** Agent 执行命令并返回结果

#### Scenario: 任务超时
- **WHEN** 任务执行超过设定超时时间
- **THEN** Agent 终止任务并返回超时状态

#### Scenario: 任务失败重试
- **WHEN** 任务执行失败且配置了重试
- **THEN** 系统按配置重试任务

### Requirement: 任务状态跟踪
测试 SHALL 验证任务状态更新。

#### Scenario: 状态更新
- **WHEN** Agent 上报任务结果
- **THEN** 系统更新任务状态为 success/failed/timeout

#### Scenario: 任务查询
- **WHEN** 请求 GET /api/v1/tasks/{id}
- **THEN** 返回任务详情包含状态、输出、耗时
