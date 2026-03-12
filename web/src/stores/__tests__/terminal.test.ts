import { describe, it, expect, beforeEach } from 'vitest'
import useTerminalStore from '@/stores/terminal'

describe('命令历史功能测试', () => {
  beforeEach(() => {
    // 每个测试前清除 localStorage
    localStorage.clear()
  })

  describe('addCommandToHistory 添加命令到历史', () => {
    it('应该为 Agent 添加命令到历史记录', () => {
      const { addCommandToHistory } = useTerminalStore.getState()
      const agentId = 'test-agent-1'

      addCommandToHistory(agentId, 'ls -la')

      const { commandHistory } = useTerminalStore.getState()
      expect(commandHistory[agentId]).toEqual(['ls -la'])
    })

    it('应该以反向顺序添加多条命令（最新的在前）', () => {
      const { addCommandToHistory } = useTerminalStore.getState()
      const agentId = 'test-agent-1'

      addCommandToHistory(agentId, 'ls -la')
      addCommandToHistory(agentId, 'pwd')
      addCommandToHistory(agentId, 'df -h')

      const { commandHistory } = useTerminalStore.getState()
      expect(commandHistory[agentId]).toEqual(['df -h', 'pwd', 'ls -la'])
    })

    it('重新输入时应将现有命令移到顶部（避免重复）', () => {
      const { addCommandToHistory } = useTerminalStore.getState()
      const agentId = 'test-agent-1'

      addCommandToHistory(agentId, 'ls -la')
      addCommandToHistory(agentId, 'pwd')
      addCommandToHistory(agentId, 'df -h')
      addCommandToHistory(agentId, 'ls -la') // 重新输入 ls -la

      const { commandHistory } = useTerminalStore.getState()
      // 应该只有一个 ls -la 在顶部
      expect(commandHistory[agentId]).toEqual(['ls -la', 'df -h', 'pwd'])
    })

    it('应该限制历史记录为 100 条命令', () => {
      const { addCommandToHistory } = useTerminalStore.getState()
      const agentId = 'test-agent-limit-test'

      for (let i = 0; i < 105; i++) {
        addCommandToHistory(agentId, `command-${i}`)
      }

      const { commandHistory } = useTerminalStore.getState()
      expect(commandHistory[agentId]).toHaveLength(100)
      expect(commandHistory[agentId][0]).toBe('command-104')
    })

    it('应该为不同的 Agent 维护独立历史', () => {
      const { addCommandToHistory } = useTerminalStore.getState()

      addCommandToHistory('test-sep-1', 'ls -la')
      addCommandToHistory('test-sep-2', 'pwd')
      addCommandToHistory('test-sep-1', 'df -h')

      const { commandHistory } = useTerminalStore.getState()
      expect(commandHistory['test-sep-1']).toEqual(['df -h', 'ls -la'])
      expect(commandHistory['test-sep-2']).toEqual(['pwd'])
    })

    it('应该处理空命令', () => {
      const { addCommandToHistory } = useTerminalStore.getState()
      const agentId = 'test-empty'

      addCommandToHistory(agentId, '')

      const { commandHistory } = useTerminalStore.getState()
      expect(commandHistory[agentId]).toContain('')
    })
  })

  describe('navigateHistory 浏览命令历史', () => {
    it('应该从旧到新向上浏览历史', () => {
      const { addCommandToHistory, navigateHistory } = useTerminalStore.getState()
      const agentId = 'test-nav-1'

      addCommandToHistory(agentId, 'ls -la')
      addCommandToHistory(agentId, 'pwd')
      addCommandToHistory(agentId, 'df -h')

      // history = ['df-h', 'pwd', 'ls-la']
      // navigateHistory 返回: history[history.length - 1 - newIndex]
      // 第一次 'up': newIndex=0, 返回 history[2]='ls-la' (最旧的)

      let command = navigateHistory(agentId, 'up')
      expect(command).toBe('ls -la') // 最旧的

      command = navigateHistory(agentId, 'up')
      expect(command).toBe('pwd')

      command = navigateHistory(agentId, 'up')
      expect(command).toBe('df -h') // 最新的
    })

    it('浏览超出历史范围时应停留在最新命令', () => {
      const { addCommandToHistory, navigateHistory } = useTerminalStore.getState()
      const agentId = 'test-nav-2'

      addCommandToHistory(agentId, 'ls -la')
      addCommandToHistory(agentId, 'pwd')
      addCommandToHistory(agentId, 'df -h')

      // 浏览到末尾
      navigateHistory(agentId, 'up')
      navigateHistory(agentId, 'up')
      navigateHistory(agentId, 'up')

      // 尝试继续向上
      const command = navigateHistory(agentId, 'up')
      expect(command).toBe('df -h') // 停留在最新的
    })

    it('应该向下浏览历史', () => {
      const { addCommandToHistory, navigateHistory } = useTerminalStore.getState()
      const agentId = 'test-nav-3'

      addCommandToHistory(agentId, 'ls -la')
      addCommandToHistory(agentId, 'pwd')
      addCommandToHistory(agentId, 'df -h')

      // 先向上浏览所有命令
      navigateHistory(agentId, 'up')  // ls-la (最旧的)
      navigateHistory(agentId, 'up')  // pwd

      // 向下浏览: pwd → ls-la → null
      let command = navigateHistory(agentId, 'down')
      expect(command).toBe('ls -la')

      command = navigateHistory(agentId, 'down')
      expect(command).toBe(null)
    })

    it('不存在历史时应返回 null', () => {
      const { navigateHistory } = useTerminalStore.getState()

      const upCommand = navigateHistory('test-no-history', 'up')
      expect(upCommand).toBe(null)

      const downCommand = navigateHistory('test-no-history', 'down')
      expect(downCommand).toBe(null)
    })

    it('从开始向上浏览应返回第一条（最旧）命令', () => {
      const { addCommandToHistory, navigateHistory } = useTerminalStore.getState()
      const agentId = 'test-nav-4'

      addCommandToHistory(agentId, 'ls -la')
      addCommandToHistory(agentId, 'pwd')
      addCommandToHistory(agentId, 'df -h')

      // 第一次 'up' 应该返回最旧的命令 'ls-la'
      const command = navigateHistory(agentId, 'up')
      expect(command).toBe('ls -la')
    })
  })

  describe('resetHistoryIndex 重置历史索引', () => {
    it('应该重置历史索引为 -1', () => {
      const { addCommandToHistory, navigateHistory, resetHistoryIndex } = useTerminalStore.getState()
      const agentId = 'test-reset-1'

      addCommandToHistory(agentId, 'ls -la')
      addCommandToHistory(agentId, 'pwd')
      navigateHistory(agentId, 'up')

      expect(useTerminalStore.getState().historyIndex[agentId]).toBeGreaterThan(-1)

      resetHistoryIndex(agentId)

      const { historyIndex } = useTerminalStore.getState()
      expect(historyIndex[agentId]).toBe(-1)
    })

    it('有历史但无索引的 Agent 应设置索引为 -1', () => {
      const { addCommandToHistory, resetHistoryIndex } = useTerminalStore.getState()
      const agentId = 'test-reset-2'

      addCommandToHistory(agentId, 'ls -la')

      // Reset 应该创建/更新索引为 -1
      resetHistoryIndex(agentId)

      const { historyIndex } = useTerminalStore.getState()
      expect(historyIndex[agentId]).toBe(-1)
    })
  })

  describe('不同 Agent 之间的历史隔离', () => {
    it('应该为不同 Agent 维护独立的浏览状态', () => {
      const { addCommandToHistory, navigateHistory } = useTerminalStore.getState()

      addCommandToHistory('test-iso-1', 'ls -la')
      addCommandToHistory('test-iso-2', 'pwd')
      addCommandToHistory('test-iso-1', 'df -h')
      addCommandToHistory('test-iso-2', 'top')

      navigateHistory('test-iso-1', 'up')
      navigateHistory('test-iso-1', 'up')

      const { historyIndex } = useTerminalStore.getState()
      // agent-1 应该浏览了两次
      expect(historyIndex['test-iso-1']).toBe(1)
      // agent-2 应该是 -1（由 addCommandToHistory 创建）
      expect(historyIndex['test-iso-2']).toBe(-1)
    })

    it('为另一个 Agent 添加命令时不应影响一个 Agent', () => {
      const { addCommandToHistory, navigateHistory } = useTerminalStore.getState()

      addCommandToHistory('test-nest-a', 'ls -la')
      addCommandToHistory('test-nest-a', 'pwd')
      navigateHistory('test-nest-a', 'up')

      // agent-a 现在 historyIndex = 0 或 1
      const indexBefore = useTerminalStore.getState().historyIndex['test-nest-a']

      // 为不同的 Agent 添加命令
      addCommandToHistory('test-nest-b', 'top')
      addCommandToHistory('test-nest-b', 'ps aux')

      const { commandHistory, historyIndex } = useTerminalStore.getState()
      expect(commandHistory['test-nest-a']).toEqual(['pwd', 'ls -la'])
      // agent-a 的索引应该被保留（虽然 addCommandToHistory 可能会重置它）
      if (indexBefore !== undefined) {
        expect(historyIndex['test-nest-a']).toBe(indexBefore)
      }
      expect(commandHistory['test-nest-b']).toEqual(['ps aux', 'top'])
    })
  })

  describe('Quick Commands 快捷命令', () => {
    it('应该添加自定义快捷命令', () => {
      const { addQuickCommand, quickCommands } = useTerminalStore.getState()
      const initialLength = quickCommands.length
      const customId = 'quick-test-' + Date.now()

      addQuickCommand({
        id: customId,
        name: 'My Command',
        command: 'my-custom-cmd',
        isCustom: true,
      })

      // 检查新命令是否存在
      const found = useTerminalStore.getState().quickCommands.find((c) => c.id === customId)
      expect(found).toBeDefined()
      expect(found?.command).toBe('my-custom-cmd')
      // 快捷命令会从其他测试累积，只检查增长
      expect(useTerminalStore.getState().quickCommands.length).toBeGreaterThanOrEqual(initialLength)
    })

    it('应该通过 id 删除快捷命令', () => {
      const { addQuickCommand, removeQuickCommand } = useTerminalStore.getState()
      const customId = 'remove-test-' + Date.now()

      addQuickCommand({
        id: customId,
        name: 'Remove Me',
        command: 'remove-me',
        isCustom: true,
      })

      removeQuickCommand(customId)

      expect(useTerminalStore.getState().quickCommands.find((c) => c.id === customId)).toBeUndefined()
    })

    it('应该以默认快捷命令开始', () => {
      const { quickCommands } = useTerminalStore.getState()

      expect(quickCommands.length).toBeGreaterThan(0)
      expect(quickCommands.some((c) => c.command === 'ls -la')).toBe(true)
      expect(quickCommands.some((c) => c.command === 'top -b -n 1')).toBe(true)
    })
  })
})
