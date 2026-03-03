import { test, expect } from '../../fixtures'
import { mockOnlineAgent, mockOfflineAgent, createApiResponse, createPaginatedResponse } from '../../mocks/agents.mock'
import { mockPendingTask, createPaginatedResponse as createTaskPaginatedResponse } from '../../mocks/tasks.mock'
import { AgentDetailPage } from '../../pages/agent-detail.page'

test.describe('Agent Detail - Execute Task', () => {
  let agentPage: AgentDetailPage

  test.beforeEach(async ({ authenticatedPage }) => {
    agentPage = new AgentDetailPage(authenticatedPage)

    // Mock agent API for specific agent
    await authenticatedPage.route('**/api/v1/agents/agent-online-001', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(createApiResponse(mockOnlineAgent)),
      })
    })

    // Mock agents list for modal (any query with status=online)
    await authenticatedPage.route(/api\/v1\/agents.*status=online/, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(createPaginatedResponse([mockOnlineAgent])),
      })
    })

    // Mock any other agents list calls
    await authenticatedPage.route('**/api/v1/agents?*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(createPaginatedResponse([mockOnlineAgent])),
      })
    })

    // Mock metrics endpoint
    await authenticatedPage.route('**/api/v1/agents/*/metrics*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(createApiResponse([])),
      })
    })

    // Mock tasks list
    await authenticatedPage.route('**/api/v1/tasks**', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(createTaskPaginatedResponse([])),
      })
    })
  })

  test('should open execute task modal when button clicked', async ({ authenticatedPage }) => {
    await agentPage.goto(mockOnlineAgent.id)
    await agentPage.waitForPageLoad()

    await agentPage.openExecuteTaskModal()

    // Wait for modal to be visible using role
    await expect(authenticatedPage.getByRole('dialog')).toBeVisible({ timeout: 10000 })

    // Check modal title exists
    await expect(authenticatedPage.getByRole('dialog').getByText('执行任务')).toBeVisible()
  })

  test('should pre-select current agent in modal', async ({ authenticatedPage }) => {
    await agentPage.goto(mockOnlineAgent.id)
    await agentPage.waitForPageLoad()

    await agentPage.openExecuteTaskModal()
    await expect(authenticatedPage.getByRole('dialog')).toBeVisible({ timeout: 10000 })

    // Agent should be pre-selected
    const selectedAgent = authenticatedPage.locator('.ant-select-selection-item')
    await expect(selectedAgent).toBeVisible()
  })

  test('should close modal on cancel', async ({ authenticatedPage }) => {
    await agentPage.goto(mockOnlineAgent.id)
    await agentPage.waitForPageLoad()

    await agentPage.openExecuteTaskModal()
    await expect(authenticatedPage.getByRole('dialog')).toBeVisible({ timeout: 10000 })

    // Press Escape to close the modal
    await authenticatedPage.keyboard.press('Escape')

    // Wait for modal to close
    await expect(authenticatedPage.getByRole('dialog')).not.toBeVisible({ timeout: 10000 })
  })

  test('should submit shell command task successfully', async ({ authenticatedPage }) => {
    let taskCreated = false

    await authenticatedPage.route('**/api/v1/tasks', async (route) => {
      if (route.request().method() === 'POST') {
        taskCreated = true
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(createApiResponse(mockPendingTask)),
        })
      } else {
        await route.continue()
      }
    })

    await agentPage.goto(mockOnlineAgent.id)
    await agentPage.waitForPageLoad()

    await agentPage.openExecuteTaskModal()
    await expect(authenticatedPage.getByRole('dialog')).toBeVisible({ timeout: 10000 })

    // Fill form
    await agentPage.fillCommand('echo test')

    // Submit
    await agentPage.submitTask()

    // Wait for the task creation request to complete
    // The success message appears in a toast, and modal closes
    await expect(authenticatedPage.getByRole('dialog')).not.toBeVisible({ timeout: 15000 })
    expect(taskCreated).toBe(true)
  })

  test('should show validation error for empty command', async ({ authenticatedPage }) => {
    await agentPage.goto(mockOnlineAgent.id)
    await agentPage.waitForPageLoad()

    await agentPage.openExecuteTaskModal()
    await expect(authenticatedPage.getByRole('dialog')).toBeVisible({ timeout: 10000 })

    // Try to submit without filling command
    await agentPage.submitTask()

    // Should show validation error
    await expect(authenticatedPage.getByText('请输入要执行的命令')).toBeVisible({ timeout: 10000 })
  })

  test('should switch between shell and builtin command types', async ({ authenticatedPage }) => {
    await agentPage.goto(mockOnlineAgent.id)
    await agentPage.waitForPageLoad()

    await agentPage.openExecuteTaskModal()
    await expect(authenticatedPage.getByRole('dialog')).toBeVisible({ timeout: 10000 })

    // Default should be shell - check for the textarea placeholder
    await expect(authenticatedPage.getByPlaceholder('例如: ping -n 4 google.com')).toBeVisible()

    // Switch to builtin - click on the command type select
    await authenticatedPage.locator('#command_type').click()
    await authenticatedPage.getByText('内置').click()

    // Should show builtin select label - use exact match to avoid multiple elements
    await expect(authenticatedPage.getByText('内置命令', { exact: true })).toBeVisible({ timeout: 5000 })
  })

  test('should display timeout and priority inputs', async ({ authenticatedPage }) => {
    await agentPage.goto(mockOnlineAgent.id)
    await agentPage.waitForPageLoad()

    await agentPage.openExecuteTaskModal()
    await expect(authenticatedPage.getByRole('dialog')).toBeVisible({ timeout: 10000 })

    await expect(authenticatedPage.getByText('超时时间 (秒)')).toBeVisible()
    await expect(authenticatedPage.getByText('优先级')).toBeVisible()
  })

  test('should show selected agents count alert', async ({ authenticatedPage }) => {
    await agentPage.goto(mockOnlineAgent.id)
    await agentPage.waitForPageLoad()

    await agentPage.openExecuteTaskModal()
    await expect(authenticatedPage.getByRole('dialog')).toBeVisible({ timeout: 10000 })

    // Should show count alert
    await expect(authenticatedPage.getByText(/已选择.*Agent/)).toBeVisible()
  })
})

test.describe('Agent Detail - Execute Task (Offline Agent)', () => {
  let agentPage: AgentDetailPage

  test.beforeEach(async ({ authenticatedPage }) => {
    agentPage = new AgentDetailPage(authenticatedPage)

    // Mock offline agent
    await authenticatedPage.route('**/api/v1/agents/agent-offline-001', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(createApiResponse(mockOfflineAgent)),
      })
    })

    // Mock metrics
    await authenticatedPage.route('**/api/v1/agents/*/metrics*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(createApiResponse([])),
      })
    })

    // Mock tasks
    await authenticatedPage.route('**/api/v1/tasks**', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(createTaskPaginatedResponse([])),
      })
    })
  })

  test('should disable execute task button for offline agent', async ({ authenticatedPage }) => {
    await agentPage.goto(mockOfflineAgent.id)
    await agentPage.waitForPageLoad()

    const isDisabled = await agentPage.isExecuteTaskButtonDisabled()
    expect(isDisabled).toBe(true)
  })
})
