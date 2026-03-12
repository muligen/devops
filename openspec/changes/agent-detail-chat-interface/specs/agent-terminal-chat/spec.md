## ADDED Requirements

### Requirement: 终端聊天界面布局

系统 SHALL 在 Agent 详情页提供左右分栏的聊天式终端界面。

#### Scenario: 查看终端界面
- **WHEN** 用户访问 Agent 详情页的终端标签
- **THEN** 系统显示左右分栏布局
- **AND** 左侧区域显示命令执行历史记录（可滚动）
- **AND** 右侧区域固定显示命令输入框

#### Scenario: 终端响应式布局
- **WHEN** 用户在小屏幕设备（< 768px）访问
- **THEN** 系统切换为上下布局
- **AND** 输入框固定在底部，历史记录在上方

### Requirement: 命令消息展示

系统 SHALL 以气泡样式区分显示用户命令和机器响应。

#### Scenario: 用户命令消息
- **WHEN** 用户提交命令
- **THEN** 系统在右侧显示蓝色/紫色气泡
- **AND** 气泡包含命令文本和时间戳
- **AND** 气泡右上角显示用户头像/图标

#### Scenario: 机器响应消息
- **WHEN** Agent 返回执行结果
- **THEN** 系统在左侧显示灰色/绿色气泡
- **AND** 气泡包含 Agent 头像/图标
- **AND** 气泡下方显示执行时长和退出码

#### Scenario: 错误消息
- **WHEN** 命令执行失败或出错
- **THEN** 系统显示红色错误气泡

#### Scenario: 消息时间戳
- **WHEN** 消息气泡超过一定时间（如 5 分钟）
- **THEN** 系统显示消息时间戳

### Requirement: 实时输出流式显示

系统 SHALL 支持 Agent 命令输出的实时流式展示。

#### Scenario: 接收输出流
- **WHEN** 任务执行时产生输出
- **THEN** 系统通过 WebSocket 接收 `output_chunk` 消息
- **AND** 系统追加输出内容到对应的结果消息气泡

#### Scenario: 流式输出状态指示
- **WHEN** 任务正在执行中
- **THEN** 系统显示"执行中..."状态指示器
- **AND** 状态指示器显示闪烁或动画效果

#### Scenario: 输出流完成
- **WHEN** 任务执行完成
- **THEN** 系统显示最终退出码和执行时长
- **AND** 系统完成状态指示动画

### Requirement: 命令输入

系统 SHALL 提供便捷的命令输入功能。

#### Scenario: 单行命令输入
- **WHEN** 用户输入单行命令并按 Enter
- **THEN** 系统提交命令执行

#### Scenario: 多行命令输入
- **WHEN** 用户按 Shift + Enter
- **THEN** 系统在输入框中换行（不执行）

#### Scenario: 命令输入框焦点
- **WHEN** 用户点击终端区域
- **THEN** 命令输入框自动获取焦点

#### Scenario: 命令输入长度限制
- **WHEN** 用户输入超过最大长度限制（如 4096 字符）
- **THEN** 系统截断输入并提示

### Requirement: 命令历史记录

系统 SHALL 为每个 Agent 独立保存命令历史。

#### Scenario: 浏览历史命令
- **WHEN** 用户在空输入框按上箭头键（↑）
- **THEN** 系统显示上一条历史命令
- **AND** 按下箭头键（↓）显示下一条命令

#### Scenario: 执行历史命令
- **WHEN** 用户浏览到某条历史命令后按 Enter
- **THEN** 系统重用该命令并执行

#### Scenario: 命令历史持久化
- **WHEN** 用户刷新页面或关闭浏览器
- **THEN** 系统保留命令历史记录（存储在 localStorage）

#### Scenario: 命令历史隔离
- **WHEN** 用户切换到不同 Agent 的终端
- **THEN** 系统显示该 Agent 独立的命令历史
- **AND** 不同 Agent 的命令历史互不干扰

### Requirement: 快捷命令

系统 SHALL 提供常用命令的快捷按钮。

#### Scenario: 使用快捷命令
- **WHEN** 用户点击某个快捷命令按钮
- **THEN** 系统自动填充对应命令到输入框
- **AND** 自动执行命令

#### Scenario: 快捷命令填充
- **WHEN** 用户右键点击快捷命令按钮
- **THEN** 系统填充命令但不执行（允许用户编辑）

#### Scenario: 快捷命令列表
- **THEN** 系统预置常用命令列表：
  - `ls -la` - 列出详细信息
  - `top` - 查看进程
  - `df -h` - 查看磁盘
  - `ps aux` - 查看进程列表
  - `netstat -ano` - 查看网络连接

#### Scenario: 自定义快捷命令
- **THEN** 系统支持用户添加自定义快捷命令（存储在 localStorage）

### Requirement: 终端状态指示

系统 SHALL 在终端顶部显示当前 Agent 的连接和执行状态。

#### Scenario: Agent 在线状态
- **WHEN** Agent 处于在线状态
- **THEN** 系统在顶部显示绿色指示灯和"在线"文字

#### Scenario: Agent 离线状态
- **WHEN** Agent 处于离线状态
- **THEN** 系统在顶部显示灰色指示灯和"离线"文字
- **AND** 禁用命令输入功能

#### Scenario: 执行任务中状态
- **WHEN** 有任务正在执行
- **THEN** 系统显示执行中的任务数量
- **AND** 显示任务类型（Shell/内置命令）

#### Scenario: 清空终端
- **WHEN** 用户点击"清空"按钮
- **THEN** 系统删除所有本地消息显示
- **AND** 不影响命令历史记录

### Requirement: 终端输出控制

系统 SHALL 提供对输出内容的控制功能。

#### Scenario: 复制输出内容
- **WHEN** 用户选中输出内容并按 Ctrl+C（或点击复制按钮）
- **THEN** 系统复制选中内容到剪贴板

#### Scenario: 查看完整输出
- **WHEN** 命令输出超过显示阈值（如 1000 行）
- **THEN** 系统显示"展开更多"按钮
- **AND** 点击后显示完整输出

#### Scenario: 导出执行结果
- **WHEN** 用户点击"导出"按钮
- **THEN** 系统下载当前会话的命令和输出为文本文件

#### Scenario: 搜索历史消息
- **WHEN** 用户输入搜索关键词
- **THEN** 系统高亮显示匹配的历史消息
