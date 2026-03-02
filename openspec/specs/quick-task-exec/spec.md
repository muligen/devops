## Requirements

### Requirement: 快速执行面板
系统 SHALL 提供快速执行任务的对话框/面板。

#### Scenario: 打开执行面板
- **WHEN** 用户点击"执行任务"按钮
- **THEN** 系统显示任务执行对话框

#### Scenario: 关闭执行面板
- **WHEN** 用户点击关闭或取消按钮
- **THEN** 系统关闭对话框，不执行任务

### Requirement: 目标选择
系统 SHALL 支持选择一个或多个 Agent 作为执行目标。

#### Scenario: 单选 Agent
- **WHEN** 用户从 Agent 列表选择一个 Agent
- **THEN** 任务只发送到该 Agent

#### Scenario: 多选 Agent
- **WHEN** 用户勾选多个 Agent
- **THEN** 任务批量发送到所有选中的 Agent

#### Scenario: 全选在线 Agent
- **WHEN** 用户点击"全选在线"按钮
- **THEN** 系统选中所有状态为在线的 Agent

#### Scenario: 排除离线 Agent
- **WHEN** 选中的 Agent 包含离线状态
- **THEN** 系统显示警告，提示部分 Agent 离线

### Requirement: 命令类型选择
系统 SHALL 支持多种命令类型。

#### Scenario: Shell 命令
- **WHEN** 用户选择 Shell 命令类型
- **THEN** 系统显示命令输入框
- **AND** 用户可输入任意 shell 命令

#### Scenario: 内置命令
- **WHEN** 用户选择内置命令类型
- **THEN** 系统显示内置命令下拉列表（clean_disk, init_machine 等）

### Requirement: 执行结果展示
系统 SHALL 实时展示任务执行结果。

#### Scenario: 任务提交成功
- **WHEN** 任务提交成功
- **THEN** 系统显示任务 ID 和执行状态

#### Scenario: 实时输出流
- **WHEN** 任务正在执行
- **THEN** 系统实时显示命令输出（SSE 或轮询）

#### Scenario: 执行完成
- **WHEN** 任务执行完成
- **THEN** 系统显示退出码和执行时长
- **AND** 显示完整输出

### Requirement: 批量执行汇总
系统 SHALL 汇总批量执行结果。

#### Scenario: 批量执行进度
- **WHEN** 批量任务执行中
- **THEN** 系统显示进度（已完成 X/总数 Y）

#### Scenario: 批量执行结果汇总
- **WHEN** 所有批量任务完成
- **THEN** 系统显示成功数和失败数
- **AND** 支持查看每个 Agent 的执行详情
