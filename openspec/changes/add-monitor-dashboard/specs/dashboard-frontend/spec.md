## ADDED Requirements

### Requirement: Dashboard 总览页面
系统 SHALL 提供 Dashboard 总览页面，展示系统整体运行状态。

#### Scenario: 查看系统概览
- **WHEN** 用户访问 Dashboard 页面
- **THEN** 系统显示在线/离线 Agent 数量、运行中任务数量、告警数量

#### Scenario: 查看 Agent 状态卡片
- **WHEN** 用户查看 Dashboard 页面
- **THEN** 系统以卡片形式展示所有 Agent 的实时状态
- **AND** 每个卡片显示 CPU、内存、磁盘使用百分比
- **AND** 超过阈值的指标以警告色高亮

### Requirement: Agent 列表页面
系统 SHALL 提供 Agent 列表页面，支持筛选、排序和搜索。

#### Scenario: 筛选 Agent
- **WHEN** 用户选择状态筛选条件（在线/离线/全部）
- **THEN** 系统显示符合条件的 Agent 列表

#### Scenario: 排序 Agent
- **WHEN** 用户选择排序条件（CPU/内存/磁盘/名称）
- **THEN** 系统按指定字段排序显示 Agent 列表

#### Scenario: 搜索 Agent
- **WHEN** 用户输入搜索关键词
- **THEN** 系统显示名称或 IP 包含关键词的 Agent

### Requirement: Agent 详情页面
系统 SHALL 提供 Agent 详情页面，展示单个 Agent 的完整信息。

#### Scenario: 查看基本信息
- **WHEN** 用户点击 Agent 卡片或列表项
- **THEN** 系统跳转到 Agent 详情页
- **AND** 显示 Agent ID、主机名、IP、系统版本、Agent 版本

#### Scenario: 查看实时状态
- **WHEN** 用户查看 Agent 详情页
- **THEN** 系统通过 WebSocket 实时更新 CPU、内存、磁盘、运行时间

### Requirement: 任务管理页面
系统 SHALL 提供任务管理页面，展示任务执行历史。

#### Scenario: 查看任务列表
- **WHEN** 用户访问任务管理页面
- **THEN** 系统显示任务列表，包含任务 ID、类型、状态、创建时间

#### Scenario: 查看任务详情
- **WHEN** 用户点击任务列表项
- **THEN** 系统显示任务详情，包含输入参数、输出结果、执行时长

### Requirement: 告警管理页面
系统 SHALL 提供告警管理页面，管理告警规则和查看告警历史。

#### Scenario: 查看告警规则
- **WHEN** 用户访问告警管理页面
- **THEN** 系统显示所有告警规则列表

#### Scenario: 创建告警规则
- **WHEN** 用户点击新建规则并填写表单
- **THEN** 系统创建新的告警规则

#### Scenario: 查看告警历史
- **WHEN** 用户切换到告警历史标签
- **THEN** 系统显示告警事件列表，包含时间、Agent、告警内容、状态

### Requirement: 用户认证
前端 SHALL 与 Server JWT 认证体系集成。

#### Scenario: 用户登录
- **WHEN** 用户输入用户名密码并提交
- **THEN** 系统调用登录 API 获取 JWT Token
- **AND** 存储 Token 到 localStorage

#### Scenario: Token 过期处理
- **WHEN** API 返回 401 未授权
- **THEN** 系统清除 Token 并跳转到登录页

#### Scenario: 请求携带认证
- **WHEN** 前端发送 API 请求
- **THEN** 请求头携带 Authorization: Bearer <token>
