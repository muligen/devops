import { UserOutlined } from '@ant-design/icons'
import { formatRelativeTime } from '@/utils'

interface CommandMessageProps {
  content: string
  timestamp: number
  onCopy?: () => void
}

export default function CommandMessage({ content, timestamp, onCopy }: CommandMessageProps) {
  return (
    <div className="message user">
      <div className="messageAvatar">
        <UserOutlined style={{ color: '#fff', fontSize: 16 }} />
      </div>
      <div className="messageContent">
        <div className="messageBubble" title={content}>
          <code>{content || '<空命令>'}</code>
        </div>
        <div className="messageTimestamp">
          {formatRelativeTime(new Date(timestamp).toISOString())}
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
