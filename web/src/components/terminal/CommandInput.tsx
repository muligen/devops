import { useState, useRef, forwardRef, useImperativeHandle, type KeyboardEvent } from 'react'
import type { InputRef } from 'antd'
import { Input, Button } from 'antd'
import { SendOutlined } from '@ant-design/icons'

export interface CommandInputRef {
  input?: {
    value?: string
  }
}

interface CommandInputProps {
  agentId?: string
  onExecute: (command: string) => void
  disabled?: boolean
  onNavigateHistory?: (direction: 'up' | 'down') => string | null
  historyIndex?: number
  placeholder?: string
}

const CommandInput = forwardRef<CommandInputRef, CommandInputProps>(({
  onExecute,
  disabled = false,
  onNavigateHistory,
  placeholder = '输入命令，按 Enter 执行，Shift + Enter 换行',
}, ref) => {
  const [inputValue, setInputValue] = useState('')
  const inputRef = useRef<InputRef | null>(null)

  // Expose the internal Input's 'input' property via our custom ref interface
  useImperativeHandle(
    ref,
    () => ({
      input: { value: inputValue },
    }),
    [inputValue]
  )

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      if (inputValue.trim()) {
        onExecute(inputValue)
        setInputValue('')
      }
    } else if (e.key === 'ArrowUp' && onNavigateHistory) {
      e.preventDefault()
      const command = onNavigateHistory('up')
      if (command !== null && inputRef.current?.input) {
        inputRef.current.input.value = command
        setInputValue(command)
      }
    } else if (e.key === 'ArrowDown' && onNavigateHistory) {
      e.preventDefault()
      const command = onNavigateHistory('down')
      if (command !== null && inputRef.current?.input) {
        inputRef.current.input.value = command
        setInputValue(command)
      } else if (inputRef.current?.input) {
        inputRef.current.input.value = ''
        setInputValue('')
      }
    }
  }

  const handleSend = () => {
    if (inputValue.trim()) {
      onExecute(inputValue)
      setInputValue('')
    }
  }

  return (
    <div className="inputArea">
      <div className="inputWrapper">
        <Input
          ref={inputRef}
          className="commandInput"
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={placeholder}
          disabled={disabled}
        />
        <Button
          type="primary"
          icon={<SendOutlined />}
          onClick={handleSend}
          disabled={disabled || !inputValue.trim()}
          className="sendButton"
        >
          执行
        </Button>
      </div>
      <div className="hint">
        {placeholder}
      </div>
    </div>
  )
})

CommandInput.displayName = 'CommandInput'

export default CommandInput
