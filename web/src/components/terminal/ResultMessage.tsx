import { DesktopOutlined, CheckCircleOutlined, CloseCircleOutlined, WarningOutlined } from '@ant-design/icons'
import { formatRelativeTime, formatDuration } from '@/utils'
import type { TerminalMessage } from '@/stores/terminal'

interface ResultMessageProps {
  message: TerminalMessage
  onCopy?: () => void
  onExpand?: () => void
  isExpanded?: boolean
}

export default function ResultMessage({ message, onCopy, onExpand, isExpanded }: ResultMessageProps) {
  const isRunning = message.status === 'running'
  const isCompleted = message.status === 'completed'
  const isFailed = message.status === 'failed'
  const isTimeout = message.status === 'timeout'
  const isError = message.type === 'error'

  return (
    <div className={`message agent ${isError ? 'error' : ''}`}>
      <div className="messageAvatar">
        <DesktopOutlined style={{ color: '#4096ff', fontSize: 16 }} />
      </div>
      <div className="messageContent">
        <div className="messageBubble">
          <div className="outputContent">
            {isRunning && <span className="loadingText">执行中...</span>}
            {message.content}
          </div>

          {message.content && message.content.length > 5000 && !isExpanded && (
            <button
              className="expandButton"
              onClick={onExpand}
            >
              展开完整输出 ({(message.content.length / 1024).toFixed(1)} KB)
            </button>
          )}

          {(isRunning || isCompleted || isFailed || isTimeout || message.exitCode !== undefined || message.duration !== undefined) && (
            <div className={`statusIndicator ${message.status || ''}`}>
              {isRunning && <div className="statusSpinner" />}
              {isCompleted && <CheckCircleOutlined />}
              {isFailed && <CloseCircleOutlined />}
              {isTimeout && <WarningOutlined />}

              {message.status === 'running' && <span>执行中...</span>}
              {message.exitCode !== undefined && (
                <span>退出码: {message.exitCode}</span>
              )}
              {message.duration !== undefined && (
                <span>耗时: {formatDuration(message.duration)}</span>
              )}
            </div>
          )}
        </div>

        <div className="messageTimestamp">
          {formatRelativeTime(new Date(message.timestamp).toISOString())}
          {onCopy && (
            <button
              className="copyButton"
              onClick={onCopy}
            >
              复制
            </button>
          )}
        </div>
      </div>
    </div>
  )
}
