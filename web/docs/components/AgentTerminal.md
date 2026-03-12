# AgentTerminal 组件

聊天式命令终端界面，用于在 Agent 详情页执行命令并查看实时输出。

## 特性

- ✨ 聊天式命令交互体验
- 📝 命令历史记录（每个 Agent 独立）
- ⚡ 实时输出流式显示
- ⌨️ 快捷命令支持
- 🎨 深色主题适配
- 📱 响应式布局（桌面端/移动端）
- 💾 localStorage 持久化

## 安装

组件位于 `web/src/components/terminal/` 目录，按需导入即可使用。

## 基本用法

```tsx
import AgentTerminal from '@/components/terminal/AgentTerminal'

function AgentDetails({ agent }) {
  return (
    <AgentTerminal agent={agent} />
  )
}
```

## Props

### `AgentTerminal`

| 属性 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `agent` | `Agent` | 是 | Agent 对象，包含 id、name、status 等信息 |

### `CommandMessage`

用户命令消息气泡。

| 属性 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `content` | `string` | 是 | 命令内容 |
| `timestamp` | `number` | 是 | 时间戳 |
| `onCopy` | `() => void` | 否 | 复制回调 |

### `ResultMessage`

机器响应消息气泡。

| 属性 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `message` | `TerminalMessage` | 是 | 消息对象 |
| `onCopy` | `() => void` | 否 | 复制回调 |
| `onExpand` | `() => void` | 否 | 展开回调 |
| `isExpanded` | `boolean` | 否 | 是否已展开 |

### `ErrorMessage`

错误消息气泡。

| 属性 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `content` | `string` | 是 | 错误内容 |
| `timestamp` | `number` | 是 | 时间戳 |

## 状态管理

使用 Zustand 存储管理终端状态：

```tsx
import useTerminalStore, { type TerminalMessage } from '@/stores/terminal'

const {
  messages,          // 所有消息（按 Agent ID 分组）
  commandHistory,    // 命令历史（按 Agent ID 分组）
  quickCommands,     // 快捷命令
  addMessage,        // 添加消息
  updateMessage,     // 更新消息
  clearMessages,     // 清空消息
  addCommandToHistory,  // 添加命令到历史
  navigateHistory,   // 浏览命令历史
  resetHistoryIndex, // 重置历史索引
} = useTerminalStore()
```

## 快捷命令

预置的快捷命令：

| 名称 | 命令 |
|------|------|
| 列出文件 | `ls -la` |
| 查看进程 | `top -b -n 1` |
| 查看磁盘 | `df -h` |
| 查看内存 | `free -h` |
| 查看网络 | `netstat -ano` |
| Ping 测试 | `ping -n 4 8.8.8.8` |

## 样式自定义

```css
/* AgentTerminal.module.css */

/* 调整消息间距 */
.message {
  margin: 12px 0;
}

/* 自定义气泡颜色 */
.message.user .messageBubble {
  background-color: #1677ff;
}

.message.agent .messageBubble {
  background-color: #1f1f1f;
}
```

## 键盘快捷键

| 快捷键 | 功能 |
|--------|------|
| `Enter` | 执行命令 |
| `Shift + Enter` | 换行（不执行） |
| `↑` | 浏览上一条历史命令 |
| `↓` | 浏览下一条历史命令 |

## 示例：带自动滚动的终端

```tsx
import { useEffect, useRef } from 'react'
import AgentTerminal from '@/components/terminal/AgentTerminal'

function AutoScrollTerminal({ agent }) {
  const terminalRef = useRef<HTMLDivElement>(null)
  const { messages } = useTerminalStore()

  useEffect(() => {
    const terminal = terminalRef.current
    if (terminal) {
      terminal.scrollTop = terminal.scrollHeight
    }
  }, [messages, agent.id])

  return (
    <div ref={terminalRef} style={{ height: '100%', overflow: 'auto' }}>
      <AgentTerminal agent={agent} />
    </div>
  )
}
```

## 注意事项

1. **任务创建**：通过 `taskApi.createTask` 创建任务，后端通过 WebSocket 推送实时输出
2. **内存管理**：消息列表无限增长，生产环境建议实现分页或自动清理
3. **安全**：所有命令都通过后端 API 执行，前端不直接执行命令
