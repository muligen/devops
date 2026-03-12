import { useEffect, useRef } from 'react'
import type { TerminalMessage } from '@/stores/terminal'
import CommandMessage from './CommandMessage'
import ResultMessage from './ResultMessage'
import ErrorMessage from './ErrorMessage'
import styles from './AgentTerminal.module.css'

interface MessageListProps {
  messages: TerminalMessage[]
  onCopyMessage?: (messageId: string) => void
  onExpandMessage?: (messageId: string) => void
  isExpanded?: Record<string, boolean>
}

export default function MessageList({ messages, onCopyMessage, onExpandMessage, isExpanded }: MessageListProps) {
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const listRef = useRef<HTMLDivElement>(null)
  const previousLengthRef = useRef(0)

  // Auto-scroll to bottom when messages are added
  useEffect(() => {
    if (messages.length > previousLengthRef.current) {
      messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
    }
    previousLengthRef.current = messages.length
  }, [messages.length])

  if (messages.length === 0) {
    return (
      <div className={styles.messageList}>
        <div className={styles.emptyState}>
          <div className={styles.emptyStateIcon}>💻</div>
          <div>输入命令开始对话</div>
        </div>
      </div>
    )
  }

  return (
    <div className={styles.messageList} ref={listRef}>
      {messages.map((message) => (
        <div key={message.id}>
          {message.type === 'command' && (
            <CommandMessage
              content={message.content}
              timestamp={message.timestamp}
              onCopy={() => onCopyMessage?.(message.id)}
            />
          )}
          {message.type === 'result' && (
            <ResultMessage
              message={message}
              onCopy={() => onCopyMessage?.(message.id)}
              onExpand={() => onExpandMessage?.(message.id)}
              isExpanded={isExpanded?.[message.id]}
            />
          )}
          {message.type === 'error' && (
            <ErrorMessage
              content={message.content}
              timestamp={message.timestamp}
            />
          )}
        </div>
      ))}
      <div ref={messagesEndRef} />
    </div>
  )
}
