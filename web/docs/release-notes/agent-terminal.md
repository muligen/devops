# Release Notes: Agent 终端聊天界面

发布日期: 2026-03-12

## 新增功能

### Agent 终端聊天界面
- **聊天式命令交互**: 将传统的任务列表替换为直观的聊天式终端界面
- **命令历史记录**: 每个独立的命令历史，支持方向键浏览
- **实时输出流式显示**: 通过 WebSocket 实时显示命令执行进度和结果
- **快捷命令按钮**: 提供常用命令快速执行能力
- **独立消息存储**: 消息和历史记录按 Agent ID 分隔存储
- **状态指示器**: 显示 Agent 连接状态、任务执行中/完成/失败状态

### 新增组件
- `AgentTerminal` - 终端主组件
- `CommandMessage` - 用户命令消息气泡
- `ResultMessage` - 执行结果消息气泡
- `ErrorMessage` - 错误消息气泡
- `TerminalHeader` - 终端顶部状态栏
- `MessageList` - 消息列表容器
- `CommandInput` - 命令输入区域
- `QuickCommands` - 快捷命令栏
- `CommandButton` - 快捷命令按钮

### 状态管理
- 新增 `useTerminalStore` (Zustand) 用于管理终端状态
- 支持消息、命令历史、快捷命令的持久化

### 布局改进
- 桌面端：左右分栏布局
- 移动端：上下布局，输入框固定底部
- 响应式设计适配各种屏幕尺寸

## 改进

### 用户体验
- 更直观的命令执行体验
- 减少页面跳转，提升操作效率
- 保持原有任务管理页面用于批量操作

### 界面优化
- 深色主题适配
- 支持输出内容复制
- 长输出「展开更多」功能
- 导出执行结果为文本文件

## 技术变更

### 前端依赖
- 新增 `vitest` 用于单元测试
- 新增 `@testing-library/react` 用于组件测试
- 新增 `@testing-library/user-event` 用于交互测试

### 测试覆盖
- 单元测试：消息组件渲染、命令历史功能
- E2E 测试：完整用户交互流程
- 响应式布局测试：移动端、桌面端适配

## 兼容性

- 浏览器支持：Chrome 90+, Firefox 88+, Safari 14+, Edge 90+
- 移动端支持：iOS Safari 14+, Chrome Mobile 90+

## 已知问题

1. 消息列表无限增长，未来版本将实现分页或自动清理
2. WebSocket 断连时需要手动刷新页面

## 后续计划

- [ ] 实现消息列表分页和自动清理
- [ ] 支持命令自动补全（基于历史）
- [ ] 支持多 Agent 并发命令执行
- [ ] 增强的 WebSocket 断连重连机制

## 升级指南

### 对用户的影响
- Agent 详情页默认显示「终端」标签页
- 「最近任务」列表作为备用标签页保留

### 对开发者的影响
- 新增终端组件和相关 store
- 修改 Agent 详情页布局以支持标签页切换
- 添加测试覆盖 (`web/src/components/terminal/__tests__/`)

## 相关文档

- [组件文档](./components/AgentTerminal.md)
- [用户指南](./user-guide/terminal.md)
