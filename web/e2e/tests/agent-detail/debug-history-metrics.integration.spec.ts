import { test, expect } from '../../fixtures'
import { AgentDetailPage } from '../../pages/agent-detail.page'

test.describe('Agent Detail - Debug History', () => {
  test('debug: check page structure and data loading', async ({ authenticatedPage, apiClient }) => {
    // Login to get token
    const loginResponse = await apiClient.post('http://localhost:8080/api/v1/auth/login', {
      data: { username: 'admin', password: 'admin123' },
    })
    const loginData = await loginResponse.json()
    const token = loginData.data.access_token

    // Get agents list
    const agentsResponse = await apiClient.get('http://localhost:8080/api/v1/agents', {
      headers: { Authorization: `Bearer ${token}` },
    })
    const agentsData = await agentsResponse.json()
    const agent = agentsData.data?.[0]

    if (!agent) {
      test.skip()
      return
    }

    // Navigate to agent detail page
    await authenticatedPage.goto(`/agents/${agent.id}`)

    // Wait for page to load
    await authenticatedPage.waitForSelector('.ant-card', { timeout: 15000 })

    // Take screenshot for debugging
    await authenticatedPage.screenshot({ path: 'test-results/debug-page.png', fullPage: true })

    // Check if metrics API is called
    const metricsResponse = await apiClient.get(`http://localhost:8080/api/v1/agents/${agent.id}/metrics?range=1h`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    const metricsData = await metricsResponse.json()

    console.log('Agent data:', JSON.stringify(agent, null, 2))
    console.log('Metrics count:', metricsData.data?.length)

    // Check if ECharts canvas exists
    const canvasCount = await authenticatedPage.locator('canvas').count()
    console.log('Canvas elements found:', canvasCount)

    // Check page HTML for history metrics card
    const historyCard = authenticatedPage.locator('.ant-card').filter({ hasText: '历史指标' })
    const historyCardContent = await historyCard.locator('.ant-card-body').innerHTML()
    console.log('History card content length:', historyCardContent.length)

    // Check if ReactECharts container exists
    const echartsDiv = await authenticatedPage.locator('.echarts-for-react').count()
    console.log('ECharts containers found:', echartsDiv)

    // Verify chart legend appears
    const legend = await authenticatedPage.locator('.ant-card').filter({ hasText: '历史指标' }).locator('text=CPU').count()
    console.log('CPU legend count:', legend)

    // Assertions
    expect(canvasCount).toBeGreaterThan(0)
    expect(metricsData.data?.length).toBeGreaterThan(0)
  })
})