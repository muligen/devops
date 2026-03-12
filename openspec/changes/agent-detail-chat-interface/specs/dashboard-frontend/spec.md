## ADDED Requirements

### Requirement: Agent 详情页终端标签页

系统 SHALL 在 Agent 详情页提供终端聊天界面的标签页选项。

#### Scenario: 打开终端标签页
- **WHEN** 用户访问 Agent 详情页并点击"终端"标签
- **THEN** 系统显示聊天式命令终端界面
- **AND** 界面包含命令历史区域、命令输入区域、快捷命令

#### Scenario: 终端标签页默认选中
- **THEN** 新版 Agent 详情页默认选中"终端"标签页
- **AND** "最近任务"列表作为备用标签页保留

### Requirement: 终端与详情页联动

系统 SHALL 实现终端界面与 Agent 详情页的信息联动。

#### Scenario: 同步 Agent 状态
- **WHEN** Agent 状态从在线变为离线（或反之）
- **THEN** 系统在终端顶部同步显示新的状态
- **AND** 相应启用或禁用命令输入功能

#### Scenario: 路由同步
- **WHEN** 用户切换到不同 Agent
- **THEN** 系统重置终端界面为新 Agent 的专属界面
- **AND** 加载新 Agent 的命令历史

#### Scenario: 实时指标更新
- **WHEN** Agent 发送心跳更新指标（CPU/内存/磁盘）
- **THEN** 系统在 Agent 详情页其他区域更新显示
- **AND** 终端不受影响（独立运行）
