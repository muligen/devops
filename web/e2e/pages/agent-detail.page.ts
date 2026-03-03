import { Page, Locator } from '@playwright/test'

export class AgentDetailPage {
  readonly page: Page

  // Navigation
  readonly backButton: Locator
  readonly refreshButton: Locator
  readonly executeTaskButton: Locator

  // Info Card
  readonly infoCard: Locator
  readonly agentName: Locator
  readonly statusTag: Locator
  readonly hostname: Locator
  readonly ipAddress: Locator
  readonly osInfo: Locator
  readonly version: Locator
  readonly createdAt: Locator
  readonly lastSeen: Locator

  // Status Dashboard
  readonly statusCard: Locator
  readonly cpuGauge: Locator
  readonly memoryGauge: Locator
  readonly diskProgress: Locator

  // Recent Tasks
  readonly recentTasksCard: Locator
  readonly tasksTable: Locator
  readonly noTasksMessage: Locator

  // Execute Task Modal
  readonly executeTaskModal: Locator
  readonly agentSelect: Locator
  readonly commandTypeSelect: Locator
  readonly commandInput: Locator
  readonly timeoutInput: Locator
  readonly priorityInput: Locator
  readonly submitButton: Locator
  readonly cancelButton: Locator

  // Time Range Buttons
  readonly timeRangeButtons: Locator

  constructor(page: Page) {
    this.page = page

    // Navigation
    this.backButton = page.getByRole('button', { name: '返回列表' })
    this.refreshButton = page.getByRole('button', { name: '刷新' })
    this.executeTaskButton = page.getByRole('button', { name: '执行任务' })

    // Info Card
    this.infoCard = page.locator('.ant-card').first()
    this.agentName = page.locator('.ant-card-head-title span').first()
    this.statusTag = page.locator('.ant-tag').first()
    this.hostname = page.locator('.ant-descriptions-item').filter({ hasText: '主机名' }).locator('.ant-descriptions-item-content')
    this.ipAddress = page.locator('.ant-descriptions-item').filter({ hasText: 'IP 地址' }).locator('.ant-descriptions-item-content')
    this.osInfo = page.locator('.ant-descriptions-item').filter({ hasText: '操作系统' }).locator('.ant-descriptions-item-content')
    this.version = page.locator('.ant-descriptions-item').filter({ hasText: 'Agent 版本' }).locator('.ant-descriptions-item-content')
    this.createdAt = page.locator('.ant-descriptions-item').filter({ hasText: '首次上线' }).locator('.ant-descriptions-item-content')
    this.lastSeen = page.locator('.ant-descriptions-item').filter({ hasText: '最后心跳' }).locator('.ant-descriptions-item-content')

    // Status Dashboard
    this.statusCard = page.locator('.ant-card').filter({ hasText: '实时状态' })
    this.cpuGauge = page.locator('canvas').first()
    this.memoryGauge = page.locator('canvas').nth(1)
    this.diskProgress = page.locator('.ant-progress-circle').first()

    // Recent Tasks
    this.recentTasksCard = page.locator('.ant-card').filter({ hasText: '最近任务' })
    this.tasksTable = page.locator('.ant-table')
    this.noTasksMessage = page.locator('.ant-empty-description')

    // Execute Task Modal
    this.executeTaskModal = page.locator('.ant-modal').filter({ hasText: '执行任务' })
    this.agentSelect = page.locator('#agent_ids')
    this.commandTypeSelect = page.locator('#command_type')
    this.commandInput = page.locator('#command')
    this.timeoutInput = page.locator('#timeout')
    this.priorityInput = page.locator('#priority')
    this.submitButton = page.getByRole('button', { name: '执 行' })
    this.cancelButton = page.getByRole('button', { name: '取 消' })

    // Time Range
    this.timeRangeButtons = page.locator('.ant-card').filter({ hasText: '历史指标' }).locator('.ant-btn')
  }

  async goto(agentId: string) {
    await this.page.goto(`/agents/${agentId}`)
  }

  async clickBack() {
    await this.backButton.click()
  }

  async clickRefresh() {
    await this.refreshButton.click()
  }

  async openExecuteTaskModal() {
    await this.executeTaskButton.click()
  }

  async selectAgents(agentNames: string[]) {
    for (const name of agentNames) {
      await this.agentSelect.click()
      await this.page.getByTitle(name).click()
    }
  }

  async selectCommandType(type: 'shell' | 'builtin') {
    await this.commandTypeSelect.click()
    await this.page.getByText(type === 'shell' ? 'Shell' : '内置').click()
  }

  async fillCommand(command: string) {
    await this.commandInput.fill(command)
  }

  async fillTimeout(timeout: number) {
    await this.timeoutInput.fill(timeout.toString())
  }

  async fillPriority(priority: number) {
    await this.priorityInput.fill(priority.toString())
  }

  async submitTask() {
    // Use force: true to bypass any overlay issues
    await this.submitButton.click({ force: true })
  }

  async cancelTask() {
    // Use force: true to bypass any overlay issues
    await this.cancelButton.click({ force: true })
  }

  async selectTimeRange(range: '1h' | '24h' | '7d') {
    const labels: Record<string, string> = {
      '1h': '1 小时',
      '24h': '24 小时',
      '7d': '7 天',
    }
    await this.page.getByRole('button', { name: labels[range] }).click()
  }

  async isExecuteTaskButtonDisabled(): Promise<boolean> {
    return await this.executeTaskButton.isDisabled()
  }

  async getStatusText(): Promise<string> {
    return await this.statusTag.innerText()
  }

  async isOnline(): Promise<boolean> {
    const text = await this.getStatusText()
    return text === '在线'
  }

  async waitForPageLoad() {
    await this.page.waitForSelector('.ant-card', { state: 'visible' })
  }

  async waitForModal() {
    await this.executeTaskModal.waitFor({ state: 'visible' })
  }

  async waitForModalClose() {
    await this.executeTaskModal.waitFor({ state: 'hidden' })
  }
}
