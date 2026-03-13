import { useEffect, useRef, useState, useCallback } from 'react'
import { useTerminalStore } from '@/stores/terminal'
import { useAuthStore } from '@/stores/auth'
import { apiClient } from '@/api/client'
import { taskApi } from '@/api'
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
    updateMessage,
    setCurrentAgentId,
  } = useTerminalStore()

  const inputRef = useRef<CommandInputRef | null>(null)
  const [expandedMessages, setExpandedMessages] = useState<Record<string, boolean>>({})
  const token = useAuthStore((state) => state.token)
  const [pendingTaskIds, setPendingTaskIds] = useState<Set<string>>(new Set())
  const pollingRef = useRef<NodeJS.Timeout>()

  // 设置当前 Agent ID
  useEffect(() => {
    setCurrentAgentId(agent.id)
  }, [agent.id, setCurrentAgentId])

  // Poll for task result updates
  const pollTaskResults = useCallback(async () => {
    const agentMessages = messages[agent.id] || []

    // Find messages with taskId and 'running' status
    const runningMessages = agentMessages.filter(m => m.status === 'running' && m.taskId)

    for (const message of runningMessages) {
      try {
        const task = await taskApi.get(message.taskId!)
        if (task.status !== 'running' && task.status !== 'pending') {
          // Map task status to message status
          const statusMap: Record<string, 'success' | 'failed' | 'timeout'> = {
            success: 'success',
            failed: 'failed',
            timeout: 'timeout',
          }
          const taskStatus = statusMap[task.status] || 'completed'

          // Update the message using message.id (not taskId)
          updateMessage(agent.id, message.id, {
            status: taskStatus,
            content: task.output,
            exitCode: task.exit_code,
            duration: task.duration,
          })
        }
      } catch (error) {
        console.error('Failed to get task status:', error)
      }
    }
  }, [agent.id, messages, updateMessage])

  // Set up polling
  useEffect(() => {
    // Poll every 1 second for task updates
    pollingRef.current = setInterval(() => {
      pollTaskResults()
    }, 1000)

    return () => {
      if (pollingRef.current) {
        clearInterval(pollingRef.current)
      }
    }
  }, [pollTaskResults])

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
      // 创建任务 - 使用正确的 API 格式
      const { data } = await apiClient.post<{ data: { id: string } }>('/tasks', {
        agent_id: agent.id,
        type: 'exec_shell',
        params: {
          command: trimmedCommand,
        },
      }, {
        timeout: 30000,
      })

      // 添加结果消息
      addMessage(agent.id, {
        type: 'result',
        content: '',
        taskId: data?.data?.id || data?.id,
        status: 'running',
      })
    } catch (error) {
      console.error('Command execution failed:', error)
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
