## ADDED Requirements

### Requirement: Agent 注册
测试 SHALL 验证通过 API 注册新 Agent。

#### Scenario: 成功注册 Agent
- **WHEN** 管理员提交有效的 Agent 注册请求
- **THEN** 系统创建 Agent 记录并返回 agent_id 和 token

#### Scenario: 重复名称注册
- **WHEN** 提交已存在的 Agent 名称
- **THEN** 系统返回名称冲突错误

### Requirement: Agent 查询
测试 SHALL 验证 Agent 查询 API。

#### Scenario: 查询单个 Agent
- **WHEN** 请求 GET /api/v1/agents/{id}
- **THEN** 返回 Agent 详细信息

#### Scenario: 分页查询 Agent 列表
- **WHEN** 请求 GET /api/v1/agents?page=1&page_size=10
- **THEN** 返回分页的 Agent 列表

#### Scenario: 按状态筛选
- **WHEN** 请求 GET /api/v1/agents?status=online
- **THEN** 只返回在线状态的 Agent

### Requirement: Agent 删除
测试 SHALL 验证 Agent 删除 API。

#### Scenario: 删除离线 Agent
- **WHEN** 管理员删除离线状态的 Agent
- **THEN** Agent 记录被删除

#### Scenario: 禁止删除在线 Agent
- **WHEN** 尝试删除在线状态的 Agent
- **THEN** 系统返回错误，要求先断开连接
