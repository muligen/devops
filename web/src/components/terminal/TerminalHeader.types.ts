interface TerminalHeaderProps {
  agent: {
    id: string
    name: string
    status: string
  }
  connected: boolean
  runningTasks: number
  onClear: () => void
  disabled?: boolean
}

export type { TerminalHeaderProps }
