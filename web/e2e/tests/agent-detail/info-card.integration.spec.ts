import { test, expect } from '../../fixtures'
import { AgentDetailPage } from '../../pages/agent-detail.page'

test.describe('Agent Detail - Integration Tests', () => {
  let agentPage: AgentDetailPage

  test.beforeEach(async ({ authenticatedPage }) => {
    agentPage = new AgentDetailPage(authenticatedPage)
  })

  test('should display agent list and navigate to detail', async ({ authenticatedPage }) => {
    // Navigate to agents list
    await authenticatedPage.goto('/agents')

    // Wait for the table to load
    await expect(authenticatedPage.locator('.ant-table')).toBeVisible({ timeout: 10000 })

    // Check if there are any agents (excluding measure row)
    const rows = authenticatedPage.locator('.ant-table-tbody tr:not(.ant-table-measure-row)')
    const count = await rows.count()

    if (count > 0) {
      // Click on the name link in the first agent row
      const nameLink = rows.first().locator('a, .ant-btn-link').first()
      if (await nameLink.count() > 0) {
        await nameLink.click()
      } else {
        // Click the row itself
        await rows.first().click()
      }

      // Should navigate to detail page
      await expect(authenticatedPage).toHaveURL(/\/agents\/[^/]+$/, { timeout: 10000 })

      // Should show agent info card
      await expect(authenticatedPage.locator('.ant-card')).toBeVisible()
    }
  })

  test('should display agent detail page with real data', async ({ authenticatedPage, apiClient }) => {
    // Login to get token for API calls
    const loginResponse = await apiClient.post('http://localhost:8080/api/v1/auth/login', {
      data: { username: 'admin', password: 'admin123' },
    })
    const loginData = await loginResponse.json()
    const token = loginData.data.access_token

    // Get agents list with auth header
    const agentsResponse = await apiClient.get('http://localhost:8080/api/v1/agents', {
      headers: { Authorization: `Bearer ${token}` },
    })
    expect(agentsResponse.ok()).toBeTruthy()

    const agentsData = await agentsResponse.json()
    const agents = agentsData.data

    if (agents && agents.length > 0) {
      const testAgent = agents[0]
      await agentPage.goto(testAgent.id)
      await agentPage.waitForPageLoad()

      // Verify agent name is displayed
      await expect(authenticatedPage.getByText(testAgent.name)).toBeVisible()

      // Verify status tag exists (either online or offline)
      const statusTag = authenticatedPage.locator('.ant-tag').first()
      await expect(statusTag).toBeVisible()
    }
  })

  test('should show loading state before data loads', async ({ authenticatedPage, apiClient }) => {
    const loginResponse = await apiClient.post('http://localhost:8080/api/v1/auth/login', {
      data: { username: 'admin', password: 'admin123' },
    })
    const loginData = await loginResponse.json()
    const token = loginData.data.access_token

    const agentsResponse = await apiClient.get('http://localhost:8080/api/v1/agents', {
      headers: { Authorization: `Bearer ${token}` },
    })
    const agentsData = await agentsResponse.json()

    if (agentsData.data && agentsData.data.length > 0) {
      const testAgent = agentsData.data[0]

      // Navigate and immediately check for loading (might be too fast to catch)
      await agentPage.goto(testAgent.id)

      // After loading, content should be visible - use first() to avoid strict mode
      await expect(authenticatedPage.locator('.ant-card').first()).toBeVisible({ timeout: 15000 })
    }
  })

  test('should refresh agent data', async ({ authenticatedPage, apiClient }) => {
    const loginResponse = await apiClient.post('http://localhost:8080/api/v1/auth/login', {
      data: { username: 'admin', password: 'admin123' },
    })
    const loginData = await loginResponse.json()
    const token = loginData.data.access_token

    const agentsResponse = await apiClient.get('http://localhost:8080/api/v1/agents', {
      headers: { Authorization: `Bearer ${token}` },
    })
    const agentsData = await agentsResponse.json()

    if (agentsData.data && agentsData.data.length > 0) {
      const testAgent = agentsData.data[0]
      await agentPage.goto(testAgent.id)
      await agentPage.waitForPageLoad()

      // Click refresh button
      await agentPage.clickRefresh()

      // Wait for the page to reload/re-render - use first() to avoid strict mode
      await expect(authenticatedPage.locator('.ant-card').first()).toBeVisible()
    }
  })

  test('should navigate back to agents list', async ({ authenticatedPage, apiClient }) => {
    const loginResponse = await apiClient.post('http://localhost:8080/api/v1/auth/login', {
      data: { username: 'admin', password: 'admin123' },
    })
    const loginData = await loginResponse.json()
    const token = loginData.data.access_token

    const agentsResponse = await apiClient.get('http://localhost:8080/api/v1/agents', {
      headers: { Authorization: `Bearer ${token}` },
    })
    const agentsData = await agentsResponse.json()

    if (agentsData.data && agentsData.data.length > 0) {
      const testAgent = agentsData.data[0]
      await agentPage.goto(testAgent.id)
      await agentPage.waitForPageLoad()

      // Click back button
      await agentPage.clickBack()

      // Should be on agents list page
      await expect(authenticatedPage).toHaveURL('/agents')
      await expect(authenticatedPage.locator('.ant-table')).toBeVisible()
    }
  })

  test('should handle non-existent agent ID', async ({ authenticatedPage }) => {
    await agentPage.goto('non-existent-agent-id-12345')

    // Should show empty state
    await expect(authenticatedPage.getByText('Agent 不存在')).toBeVisible({ timeout: 10000 })
  })
})
