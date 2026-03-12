import type { TerminalHeaderProps } from './TerminalHeader.types'

export default function TerminalHeader({
  agent,
  connected,
  runningTasks,
  onClear,
  disabled = false,
}: TerminalHeaderProps) {
  const isOnline = agent.status === 'online' && connected

  return (
    <div className="header">
      <div className="headerLeft">
        <div className={`statusDot ${isOnline ? 'online' : 'offline'}`} />
        <div>
          <span className="agentName">{agent.name}</span>
          <span className="statusText">
            {' '}|{' '}
            {isOnline ? '在线' : '离线'}
          </span>
        </div>
      </div>
      <div className="headerRight">
        {runningTasks > 0 && (
          <div className="runningTasks">
            <div className="statusSpinner" />
            <span>执行中: {runningTasks}</span>
          </div>
        )}
        <button
          className="clearButton"
          onClick={onClear}
          disabled={disabled}
        >
          清空
        </button>
      </div>
    </div>
  )
}
