## MODIFIED Requirements

### Requirement: 快速执行面板

系统 SHALL 提供快速执行任务的对话框/面板，同时在 Agent 详情页终端提供直接交互方式。

#### Scenario: 打开执行面板
- **WHEN** 用户点击列表页的"批量执行任务"按钮
- **THEN** 系统显示任务执行对话框

#### Scenario: Agent 详情页直接执行
- **WHEN** 用户在 Agent 详情页终端输入命令
- **THEN** 系统直接执行命令，无需打开对话框

#### Scenario: 关闭执行面板
- **WHEN** 用户点击关闭或取消按钮
- **THEN** 系统关闭对话框，不执行任务

### Requirement: 执行结果展示

系统 SHALL 实时展示任务执行结果，同时 Agent 详情页终端提供流式输出展示。

#### Scenario: 批量执行结果汇总
- **WHEN** 批量任务执行中
- **THEN** 对话框显示进度（已完成 X/总数 Y）

#### Scenario: 批量执行结果汇总
- **WHEN** 所有批量任务完成
- **THEN** 对话框显示成功数和失败数
- **AND** 支持查看每个 Agent 的执行详情
- **AND** 点击详情跳转到对应 Agent 的终端界面

#### Scenario: 终端实时输出流
- **WHEN** 任务正在执行
- **THEN** 终端界面实时显示命令输出（WebSocket 推送）

#### Scenario: 执行完成
- **WHEN** 任务执行完成
- **THEN** 系统显示退出码和执行时长
- **AND** 终端界面显示完整输出气泡
