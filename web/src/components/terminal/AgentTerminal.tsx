import { useEffect, useRef, useState } from 'react'
import { useTerminalStore } from '@/stores/terminal'
import TerminalHeader from './TerminalHeader'
import MessageList from './MessageList'
import QuickCommands from './QuickCommands'
import CommandInput, { type CommandInputRef } from './CommandInput'
import styles from './AgentTerminal.module.css'

interface Agent {
  id: string
  name: string
  status: string
}

interface AgentTerminalProps {
  agent: Agent
}

export default function AgentTerminal({ agent }: AgentTerminalProps) {
  const {
    messages,
    quickCommands,
    addMessage,
    addCommandToHistory,
    navigateHistory,
    resetHistoryIndex,
    clearMessages,
    setCurrentAgentId,
  } = useTerminalStore()

  const inputRef = useRef<CommandInputRef | null>(null)
  const [expandedMessages, setExpandedMessages] = useState<Record<string, boolean>>({})

  // 设置当前 Agent ID
  useEffect(() => {
    setCurrentAgentId(agent.id)
  }, [agent.id, setCurrentAgentId])

  // 处理命令执行
  const handleExecuteCommand = async (command: string) => {
    const trimmedCommand = command.trim()
    if (!trimmedCommand) return

    // 添加命令消息
    addMessage(agent.id, {
      type: 'command',
      content: trimmedCommand,
    })

    // 添加到历史记录
    addCommandToHistory(agent.id, trimmedCommand)
    resetHistoryIndex(agent.id)

    try {
      // 创建任务
      const result = await fetch('/api/v1/tasks', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          agent_id: agent.id,
          command: trimmedCommand,
        }),
      }).then((res) => res.json())

      // 添加结果消息
      addMessage(agent.id, {
        type: 'result',
        content: '',
        taskId: result.id,
        status: 'running',
      })
    } catch {
      addMessage(agent.id, {
        type: 'error',
        content: '命令执行失败',
      })
    }
  }

  // 处理复制消息
  const handleCopyMessage = (messageId: string) => {
    const agentMessages = messages[agent.id] || []
    const message = agentMessages.find((m) => m.id === messageId)
    if (message) {
      navigator.clipboard.writeText(message.content)
    }
  }

  // 处理展开消息
  const handleExpandMessage = (messageId: string) => {
    setExpandedMessages((prev) => ({
      ...prev,
      [messageId]: !prev[messageId],
    }))
  }

  // 处理快捷命令直接执行
  const handleQuickCommandExecute = (command: string) => {
    handleExecuteCommand(command)
  }

  // 处理清空终端
  const handleClear = () => {
    clearMessages(agent.id)
  }

  const agentMessages = messages[agent.id] || []
  const agentHistoryIndex = useTerminalStore.getState().historyIndex[agent.id] ?? -1

  return (
    <div className={styles.terminal}>
      <TerminalHeader
        agent={agent}
        connected={agent.status === 'online'}
        runningTasks={agentMessages.filter((m) => m.status === 'running').length}
        onClear={handleClear}
      />

      <MessageList
        messages={agentMessages}
        onCopyMessage={handleCopyMessage}
        onExpandMessage={handleExpandMessage}
        isExpanded={expandedMessages}
      />

      <QuickCommands
        commands={quickCommands}
        onExecute={handleQuickCommandExecute}
      />

      <CommandInput
        ref={inputRef}
        agentId={agent.id}
        onExecute={handleExecuteCommand}
        disabled={agent.status !== 'online'}
        onNavigateHistory={(direction) => {
          const command = navigateHistory(agent.id, direction)
          const input = inputRef.current?.input
          if (command && input) {
            input.value = command
          }
          return command
        }}
        historyIndex={agentHistoryIndex}
      />
    </div>
  )
}

export type { Agent, AgentTerminalProps }
