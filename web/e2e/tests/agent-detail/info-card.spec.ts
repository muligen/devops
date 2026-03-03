import { test, expect } from '../../fixtures'
import { mockOnlineAgent, mockOfflineAgent, createApiResponse } from '../../mocks/agents.mock'
import { createTaskListForAgent, createPaginatedResponse as createTaskPaginatedResponse } from '../../mocks/tasks.mock'
import { AgentDetailPage } from '../../pages/agent-detail.page'

test.describe('Agent Detail - Info Card', () => {
  let agentPage: AgentDetailPage

  test.beforeEach(async ({ authenticatedPage }) => {
    agentPage = new AgentDetailPage(authenticatedPage)

    // Mock metrics endpoint
    await authenticatedPage.route('**/api/v1/agents/*/metrics*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(createApiResponse([])),
      })
    })

    // Mock tasks endpoint
    await authenticatedPage.route('**/api/v1/tasks*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(createTaskPaginatedResponse(createTaskListForAgent(mockOnlineAgent.id), 5)),
      })
    })
  })

  test('should display online agent information correctly', async ({ authenticatedPage }) => {
    // Mock agent API
    await authenticatedPage.route('**/api/v1/agents/*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(createApiResponse(mockOnlineAgent)),
      })
    })

    await agentPage.goto(mockOnlineAgent.id)
    await agentPage.waitForPageLoad()

    // Check agent name is displayed
    await expect(authenticatedPage.getByText(mockOnlineAgent.name)).toBeVisible()

    // Check status tag shows online
    const statusText = await agentPage.getStatusText()
    expect(statusText).toBe('在线')

    // Check hostname is displayed
    await expect(authenticatedPage.getByText(mockOnlineAgent.hostname)).toBeVisible()

    // Check IP address is displayed
    await expect(authenticatedPage.getByText(mockOnlineAgent.ip_address)).toBeVisible()
  })

  test('should display offline agent status correctly', async ({ authenticatedPage }) => {
    // Mock offline agent
    await authenticatedPage.route('**/api/v1/agents/*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(createApiResponse(mockOfflineAgent)),
      })
    })

    await agentPage.goto(mockOfflineAgent.id)
    await agentPage.waitForPageLoad()

    // Check status tag shows offline
    const statusText = await agentPage.getStatusText()
    expect(statusText).toBe('离线')
  })

  test('should show empty state for non-existent agent', async ({ authenticatedPage }) => {
    // Mock 404 response
    await authenticatedPage.route('**/api/v1/agents/*', async (route) => {
      await route.fulfill({
        status: 404,
        contentType: 'application/json',
        body: JSON.stringify({ code: 404, message: 'Agent not found' }),
      })
    })

    await agentPage.goto('non-existent-id')

    // Should show empty state
    await expect(authenticatedPage.getByText('Agent 不存在')).toBeVisible()
  })

  test('should enable execute task button for online agent', async ({ authenticatedPage }) => {
    // Mock online agent
    await authenticatedPage.route('**/api/v1/agents/*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(createApiResponse(mockOnlineAgent)),
      })
    })

    await agentPage.goto(mockOnlineAgent.id)
    await agentPage.waitForPageLoad()

    const isDisabled = await agentPage.isExecuteTaskButtonDisabled()
    expect(isDisabled).toBe(false)
  })

  test('should disable execute task button for offline agent', async ({ authenticatedPage }) => {
    // Mock offline agent
    await authenticatedPage.route('**/api/v1/agents/*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(createApiResponse(mockOfflineAgent)),
      })
    })

    await agentPage.goto(mockOfflineAgent.id)
    await agentPage.waitForPageLoad()

    const isDisabled = await agentPage.isExecuteTaskButtonDisabled()
    expect(isDisabled).toBe(true)
  })

  test('should display loading state initially', async ({ authenticatedPage }) => {
    // Delay the response
    await authenticatedPage.route('**/api/v1/agents/*', async (route) => {
      await new Promise((resolve) => setTimeout(resolve, 500))
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(createApiResponse(mockOnlineAgent)),
      })
    })

    await agentPage.goto(mockOnlineAgent.id)

    // Should show loading spinner
    await expect(authenticatedPage.locator('.ant-spin')).toBeVisible()
  })
})
