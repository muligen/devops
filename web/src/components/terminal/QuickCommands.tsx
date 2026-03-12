import { useState } from 'react'
import { Input, Modal, Form, message } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import type { QuickCommand } from '@/stores/terminal'

interface QuickCommandsProps {
  commands: QuickCommand[]
  onExecute: (command: string) => void
  onAdd?: (command: QuickCommand) => void
  onRemove?: (id: string) => void
  disabled?: boolean
}

export default function QuickCommands({
  commands,
  onExecute,
  onAdd,
  onRemove,
  disabled = false,
}: QuickCommandsProps) {
  const [isAddModalOpen, setIsAddModalOpen] = useState(false)
  const [form] = Form.useForm()

  const handleRightClick = (e: React.MouseEvent, command: string) => {
    e.preventDefault()
    navigator.clipboard.writeText(command)
    message.success('命令已复制到剪贴板，可粘贴到命令行使用')
  }

  const handleAdd = async () => {
    try {
      const values = await form.validateFields()
      onAdd?.({
        id: `custom-${Date.now()}`,
        name: values.name,
        command: values.command,
        isCustom: true,
      })
      form.resetFields()
      setIsAddModalOpen(false)
      message.success('快捷命令添加成功')
    } catch {
      // Validation failed
    }
  }

  const builtinCommands = commands.filter((cmd) => !cmd.isCustom)
  const customCommands = commands.filter((cmd) => cmd.isCustom)

  return (
    <div className="quickCommands">
      {builtinCommands.length > 0 && (
        <>
          {builtinCommands.map((cmd) => (
            <button
              key={cmd.id}
              className="quickCommandButton"
              onClick={() => onExecute(cmd.command)}
              onContextMenu={(e) => handleRightClick(e, cmd.command)}
              disabled={disabled}
              title={`执行命令: ${cmd.command}`}
            >
              {cmd.name}
            </button>
          ))}
        </>
      )}

      {customCommands.length > 0 && (
        <>
          <div className="quickCommandLabel">自定义命令</div>
          {customCommands.map((cmd) => (
            <button
              key={cmd.id}
              className="quickCommandButton"
              onClick={() => onExecute(cmd.command)}
              onContextMenu={(e) => handleRightClick(e, cmd.command)}
              disabled={disabled}
              title={`${cmd.name}: ${cmd.command}`}
              onDoubleClick={() => onRemove?.(cmd.id)}
              style={{
                borderColor: 'rgba(22, 119, 255, 0.3)',
              }}
            >
              {cmd.name}
            </button>
          ))}
        </>
      )}

      {onAdd && (
        <button
          className="quickCommandButton"
          onClick={() => setIsAddModalOpen(true)}
          disabled={disabled}
        >
          <PlusOutlined /> 添加
        </button>
      )}

      {onAdd && (
        <Modal
          title="添加自定义快捷命令"
          open={isAddModalOpen}
          onOk={handleAdd}
          onCancel={() => {
            setIsAddModalOpen(false)
            form.resetFields()
          }}
          okText="添加"
          cancelText="取消"
        >
          <Form
            form={form}
            layout="vertical"
            initialValues={{
              name: '',
              command: '',
            }}
          >
            <Form.Item
              name="name"
              label="命令名称"
              rules={[{ required: true, message: '请输入命令名称' }]}
            >
              <Input placeholder="例如: 检查日志" />
            </Form.Item>
            <Form.Item
              name="command"
              label="Shell 命令"
              rules={[{ required: true, message: '请输入 Shell 命令' }]}
            >
              <Input.TextArea
                placeholder="例如: tail -f /var/log/app.log"
                rows={3}
                autoSize={{ minRows: 2, maxRows: 6 }}
              />
            </Form.Item>
          </Form>
        </Modal>
      )}
    </div>
  )
}
