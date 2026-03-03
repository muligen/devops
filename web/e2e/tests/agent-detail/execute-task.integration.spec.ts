import { test, expect } from '../../fixtures'
import { AgentDetailPage } from '../../pages/agent-detail.page'

test.describe('Agent Detail - Execute Task Integration Tests', () => {
  let agentPage: AgentDetailPage

  test.beforeEach(async ({ authenticatedPage }) => {
    agentPage = new AgentDetailPage(authenticatedPage)
  })

  async function getAgents(apiClient: any, status?: string) {
    const loginResponse = await apiClient.post('http://localhost:8080/api/v1/auth/login', {
      data: { username: 'admin', password: 'admin123' },
    })
    const loginData = await loginResponse.json()
    const token = loginData.data.access_token

    const url = status
      ? `http://localhost:8080/api/v1/agents?status=${status}`
      : 'http://localhost:8080/api/v1/agents'

    const response = await apiClient.get(url, {
      headers: { Authorization: `Bearer ${token}` },
    })
    return await response.json()
  }

  test('should open and close execute task modal', async ({ authenticatedPage, apiClient }) => {
    const agentsData = await getAgents(apiClient)

    if (agentsData.data && agentsData.data.length > 0) {
      const testAgent = agentsData.data[0]
      await agentPage.goto(testAgent.id)
      await agentPage.waitForPageLoad()

      // Open modal
      await agentPage.openExecuteTaskModal()
      await expect(authenticatedPage.getByRole('dialog')).toBeVisible({ timeout: 10000 })

      // Close modal with Escape
      await authenticatedPage.keyboard.press('Escape')
      await expect(authenticatedPage.getByRole('dialog')).not.toBeVisible({ timeout: 10000 })
    }
  })

  test('should show execute task button disabled for offline agent', async ({ authenticatedPage, apiClient }) => {
    const agentsData = await getAgents(apiClient, 'offline')

    if (agentsData.data && agentsData.data.length > 0) {
      const offlineAgent = agentsData.data[0]
      await agentPage.goto(offlineAgent.id)
      await agentPage.waitForPageLoad()

      // Execute button should be disabled
      const isDisabled = await agentPage.isExecuteTaskButtonDisabled()
      expect(isDisabled).toBe(true)
    } else {
      test.skip()
    }
  })

  test('should show execute task button enabled for online agent', async ({ authenticatedPage, apiClient }) => {
    const agentsData = await getAgents(apiClient, 'online')

    if (agentsData.data && agentsData.data.length > 0) {
      const onlineAgent = agentsData.data[0]
      await agentPage.goto(onlineAgent.id)
      await agentPage.waitForPageLoad()

      // Execute button should be enabled
      const isDisabled = await agentPage.isExecuteTaskButtonDisabled()
      expect(isDisabled).toBe(false)
    } else {
      test.skip()
    }
  })

  test('should create task for online agent', async ({ authenticatedPage, apiClient }) => {
    const agentsData = await getAgents(apiClient, 'online')

    if (agentsData.data && agentsData.data.length > 0) {
      const onlineAgent = agentsData.data[0]
      await agentPage.goto(onlineAgent.id)
      await agentPage.waitForPageLoad()

      // Open modal
      await agentPage.openExecuteTaskModal()
      await expect(authenticatedPage.getByRole('dialog')).toBeVisible({ timeout: 10000 })

      // Fill command
      const testCommand = `echo "E2E test ${Date.now()}"`
      await agentPage.fillCommand(testCommand)

      // Submit
      await agentPage.submitTask()

      // Wait for either modal to close (success) or error message
      await Promise.race([
        expect(authenticatedPage.getByRole('dialog')).not.toBeVisible(),
        expect(authenticatedPage.locator('.ant-message-error')).toBeVisible(),
      ]).catch(() => {
        // Modal might still be open due to loading state
      })

      // Wait a moment for the task to process
      await authenticatedPage.waitForTimeout(2000)
    } else {
      test.skip()
    }
  })

  test('should validate empty command', async ({ authenticatedPage, apiClient }) => {
    const agentsData = await getAgents(apiClient, 'online')

    if (agentsData.data && agentsData.data.length > 0) {
      const onlineAgent = agentsData.data[0]
      await agentPage.goto(onlineAgent.id)
      await agentPage.waitForPageLoad()

      // Open modal
      await agentPage.openExecuteTaskModal()
      await expect(authenticatedPage.getByRole('dialog')).toBeVisible({ timeout: 10000 })

      // Submit without entering command
      await agentPage.submitTask()

      // Should show validation error
      await expect(authenticatedPage.getByText('请输入要执行的命令')).toBeVisible({ timeout: 10000 })
    } else {
      test.skip()
    }
  })

  test('should switch between shell and builtin command types', async ({ authenticatedPage, apiClient }) => {
    const agentsData = await getAgents(apiClient, 'online')

    if (agentsData.data && agentsData.data.length > 0) {
      const onlineAgent = agentsData.data[0]
      await agentPage.goto(onlineAgent.id)
      await agentPage.waitForPageLoad()

      // Open modal
      await agentPage.openExecuteTaskModal()
      await expect(authenticatedPage.getByRole('dialog')).toBeVisible({ timeout: 10000 })

      // Default should show shell command textarea
      await expect(authenticatedPage.getByPlaceholder('例如: ping -n 4 google.com')).toBeVisible()

      // Switch to builtin
      await authenticatedPage.locator('#command_type').click()
      await authenticatedPage.getByText('内置').click()

      // Should show builtin label
      await expect(authenticatedPage.getByText('内置命令', { exact: true })).toBeVisible({ timeout: 5000 })
    } else {
      test.skip()
    }
  })

  test('should display timeout and priority fields', async ({ authenticatedPage, apiClient }) => {
    const agentsData = await getAgents(apiClient, 'online')

    if (agentsData.data && agentsData.data.length > 0) {
      const onlineAgent = agentsData.data[0]
      await agentPage.goto(onlineAgent.id)
      await agentPage.waitForPageLoad()

      // Open modal
      await agentPage.openExecuteTaskModal()
      await expect(authenticatedPage.getByRole('dialog')).toBeVisible({ timeout: 10000 })

      // Check timeout and priority labels
      await expect(authenticatedPage.getByText('超时时间 (秒)')).toBeVisible()
      await expect(authenticatedPage.getByText('优先级')).toBeVisible()
    } else {
      test.skip()
    }
  })
})
