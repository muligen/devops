import { test, expect } from '../../fixtures'
import { mockOnlineAgent, createApiResponse } from '../../mocks/agents.mock'
import { createPaginatedResponse as createTaskPaginatedResponse, createTaskListForAgent } from '../../mocks/tasks.mock'
import { AgentDetailPage } from '../../pages/agent-detail.page'

test.describe('Agent Detail - Navigation', () => {
  let agentPage: AgentDetailPage

  test.beforeEach(async ({ authenticatedPage }) => {
    agentPage = new AgentDetailPage(authenticatedPage)

    // Mock agent API
    await authenticatedPage.route('**/api/v1/agents/*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(createApiResponse(mockOnlineAgent)),
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

    // Mock tasks endpoint
    await authenticatedPage.route('**/api/v1/tasks*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(createTaskPaginatedResponse(createTaskListForAgent(mockOnlineAgent.id), 5)),
      })
    })
  })

  test('should navigate back to agents list', async ({ authenticatedPage }) => {
    await agentPage.goto(mockOnlineAgent.id)
    await agentPage.waitForPageLoad()

    await agentPage.clickBack()

    // Should navigate to agents list
    await expect(authenticatedPage).toHaveURL('/agents')
  })

  test('should refresh agent data when refresh button clicked', async ({ authenticatedPage }) => {
    let requestCount = 0

    await authenticatedPage.route('**/api/v1/agents/*', async (route) => {
      requestCount++
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(createApiResponse(mockOnlineAgent)),
      })
    })

    await agentPage.goto(mockOnlineAgent.id)
    await agentPage.waitForPageLoad()

    const initialCount = requestCount

    await agentPage.clickRefresh()

    // Wait for the request to complete
    await authenticatedPage.waitForResponse((resp) => resp.url().includes('/agents/'))

    expect(requestCount).toBeGreaterThan(initialCount)
  })

  test('should display back button on page', async ({ authenticatedPage }) => {
    await agentPage.goto(mockOnlineAgent.id)
    await agentPage.waitForPageLoad()

    await expect(agentPage.backButton).toBeVisible()
  })

  test('should display refresh button on page', async ({ authenticatedPage }) => {
    await agentPage.goto(mockOnlineAgent.id)
    await agentPage.waitForPageLoad()

    await expect(agentPage.refreshButton).toBeVisible()
  })

  test('should display execute task button', async ({ authenticatedPage }) => {
    await agentPage.goto(mockOnlineAgent.id)
    await agentPage.waitForPageLoad()

    await expect(agentPage.executeTaskButton).toBeVisible()
  })
})

test.describe('Agent Detail - Time Range Navigation', () => {
  let agentPage: AgentDetailPage

  test.beforeEach(async ({ authenticatedPage }) => {
    agentPage = new AgentDetailPage(authenticatedPage)

    await authenticatedPage.route('**/api/v1/agents/*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(createApiResponse(mockOnlineAgent)),
      })
    })

    await authenticatedPage.route('**/api/v1/agents/*/metrics*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(createApiResponse([])),
      })
    })

    await authenticatedPage.route('**/api/v1/tasks*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(createTaskPaginatedResponse([], 0)),
      })
    })
  })

  test('should display time range buttons', async ({ authenticatedPage }) => {
    await agentPage.goto(mockOnlineAgent.id)
    await agentPage.waitForPageLoad()

    await expect(authenticatedPage.getByRole('button', { name: '1 小时' })).toBeVisible()
    await expect(authenticatedPage.getByRole('button', { name: '24 小时' })).toBeVisible()
    await expect(authenticatedPage.getByRole('button', { name: '7 天' })).toBeVisible()
  })

  test('should have 1 hour selected by default', async ({ authenticatedPage }) => {
    await agentPage.goto(mockOnlineAgent.id)
    await agentPage.waitForPageLoad()

    const button = authenticatedPage.getByRole('button', { name: '1 小时' })
    await expect(button).toHaveClass(/ant-btn-primary/)
  })

  test('should switch time range on button click', async ({ authenticatedPage }) => {
    await agentPage.goto(mockOnlineAgent.id)
    await agentPage.waitForPageLoad()

    // Click 24 hour button
    await agentPage.selectTimeRange('24h')

    const button = authenticatedPage.getByRole('button', { name: '24 小时' })
    await expect(button).toHaveClass(/ant-btn-primary/)
  })

  test('should make API call with new time range', async ({ authenticatedPage }) => {
    let metricsRange = ''

    await authenticatedPage.route('**/api/v1/agents/*/metrics*', async (route) => {
      const url = new URL(route.request().url())
      metricsRange = url.searchParams.get('range') || ''
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(createApiResponse([])),
      })
    })

    await agentPage.goto(mockOnlineAgent.id)
    await agentPage.waitForPageLoad()

    await agentPage.selectTimeRange('7d')

    // Wait for the API call
    await authenticatedPage.waitForResponse((resp) => resp.url().includes('range=7d'))

    expect(metricsRange).toBe('7d')
  })
})

test.describe('Agent Detail - URL Routing', () => {
  test('should load agent by ID from URL', async ({ authenticatedPage }) => {
    const agentId = 'test-agent-123'

    await authenticatedPage.route('**/api/v1/agents/*', async (route) => {
      const url = route.request().url()
      const id = url.split('/').pop()

      expect(id).toBe(agentId)

      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(
          createApiResponse({
            ...mockOnlineAgent,
            id: agentId,
          })
        ),
      })
    })

    await authenticatedPage.route('**/api/v1/agents/*/metrics*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(createApiResponse([])),
      })
    })

    await authenticatedPage.route('**/api/v1/tasks*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(createTaskPaginatedResponse([], 0)),
      })
    })

    await authenticatedPage.goto(`/agents/${agentId}`)

    // Verify URL is correct
    await expect(authenticatedPage).toHaveURL(new RegExp(`/agents/${agentId}`))
  })
})
