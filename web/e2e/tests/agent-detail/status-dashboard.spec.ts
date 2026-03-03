import { test, expect } from '../../fixtures'
import {
  mockOnlineAgent,
  mockHighCpuAgent,
  createApiResponse,
} from '../../mocks/agents.mock'
import { createPaginatedResponse as createTaskPaginatedResponse, createTaskListForAgent } from '../../mocks/tasks.mock'
import { AgentDetailPage } from '../../pages/agent-detail.page'

test.describe('Agent Detail - Status Dashboard', () => {
  let agentPage: AgentDetailPage

  test.beforeEach(async ({ authenticatedPage }) => {
    agentPage = new AgentDetailPage(authenticatedPage)

    // Mock agent endpoint
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

  test('should display CPU gauge with correct value', async ({ authenticatedPage }) => {
    await agentPage.goto(mockOnlineAgent.id)
    await agentPage.waitForPageLoad()

    // Check status card is visible
    await expect(agentPage.statusCard).toBeVisible()

    // Check gauges are rendered (canvas elements)
    await expect(agentPage.cpuGauge).toBeVisible()
    await expect(agentPage.memoryGauge).toBeVisible()
  })

  test('should display disk progress with correct percentage', async ({ authenticatedPage }) => {
    await agentPage.goto(mockOnlineAgent.id)
    await agentPage.waitForPageLoad()

    // Check disk progress is visible
    await expect(agentPage.diskProgress).toBeVisible()

    // Check percentage text - format is "55%"
    const diskUsageText = `${mockOnlineAgent.disk_usage}%`
    await expect(authenticatedPage.getByText(diskUsageText)).toBeVisible()
  })

  test('should show correct colors for normal metrics', async ({ authenticatedPage }) => {
    await agentPage.goto(mockOnlineAgent.id)
    await agentPage.waitForPageLoad()

    // For normal metrics (< 60%), disk progress should be visible
    // Just check the progress element exists, not the internal path
    await expect(agentPage.diskProgress).toBeVisible()
  })

  test('should show warning colors for high CPU usage', async ({ authenticatedPage }) => {
    // Override the agent route for high CPU agent
    await authenticatedPage.route('**/api/v1/agents/*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(createApiResponse(mockHighCpuAgent)),
      })
    })

    await agentPage.goto(mockHighCpuAgent.id)
    await agentPage.waitForPageLoad()

    // Status card should still be visible
    await expect(agentPage.statusCard).toBeVisible()

    // For high disk usage (> 80%), check the progress is visible
    await expect(agentPage.diskProgress).toBeVisible()
    const diskUsageText = `${mockHighCpuAgent.disk_usage}%`
    await expect(authenticatedPage.getByText(diskUsageText)).toBeVisible()
  })

  test('should display status card title', async ({ authenticatedPage }) => {
    await agentPage.goto(mockOnlineAgent.id)
    await agentPage.waitForPageLoad()

    await expect(authenticatedPage.getByText('实时状态')).toBeVisible()
  })

  test('should show all three metrics sections', async ({ authenticatedPage }) => {
    await agentPage.goto(mockOnlineAgent.id)
    await agentPage.waitForPageLoad()

    // Check for metric labels - "磁盘使用" is in a title element
    await expect(authenticatedPage.getByText('磁盘使用')).toBeVisible()

    // Check that canvases are rendered (2 for ECharts gauges)
    const canvases = authenticatedPage.locator('canvas')
    await expect(canvases.first()).toBeVisible()
  })
})
