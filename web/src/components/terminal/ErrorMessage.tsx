import { ExclamationCircleOutlined } from '@ant-design/icons'
import { formatRelativeTime } from '@/utils'

interface ErrorMessageProps {
  content: string
  timestamp: number
}

export default function ErrorMessage({ content, timestamp }: ErrorMessageProps) {
  return (
    <div className="message agent error">
      <div className="messageAvatar">
        <ExclamationCircleOutlined style={{ color: '#ff7875', fontSize: 16 }} />
      </div>
      <div className="messageContent">
        <div className="messageBubble">
          {content}
        </div>
        <div className="messageTimestamp">
          {formatRelativeTime(new Date(timestamp).toISOString())}
        </div>
      </div>
    </div>
  )
}
