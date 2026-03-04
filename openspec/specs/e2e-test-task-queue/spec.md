## ADDED Requirements

### Requirement: 批量任务创建
测试 SHALL 验证批量创建任务到多个 Agent。

#### Scenario: 批量创建成功
- **WHEN** 提交批量任务请求包含多个 agent_id
- **THEN** 系统为每个 Agent 创建独立任务

#### Scenario: 部分失败处理
- **WHEN** 部分目标 Agent 不存在或离线
- **THEN** 系统返回成功和失败的任务列表

### Requirement: 任务队列管理
测试 SHALL 验证任务队列状态和优先级。

#### Scenario: 任务排队
- **WHEN** 多个任务发往同一 Agent
- **THEN** 任务按优先级和时间顺序排队执行

#### Scenario: 任务取消
- **WHEN** 取消 pending 状态的任务
- **THEN** 任务状态变为 cancelled

#### Scenario: 取消执行中任务
- **WHEN** 取消 running 状态的任务
- **THEN** Agent 收到取消信号并终止执行
