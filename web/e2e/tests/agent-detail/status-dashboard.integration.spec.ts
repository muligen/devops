import { test, expect } from '../../fixtures'
import { AgentDetailPage } from '../../pages/agent-detail.page'

test.describe('Agent Detail - Status Dashboard Integration Tests', () => {
  let agentPage: AgentDetailPage

  test.beforeEach(async ({ authenticatedPage }) => {
    agentPage = new AgentDetailPage(authenticatedPage)
  })

  async function getFirstAgent(apiClient: any) {
    const loginResponse = await apiClient.post('http://localhost:8080/api/v1/auth/login', {
      data: { username: 'admin', password: 'admin123' },
    })
    const loginData = await loginResponse.json()
    const token = loginData.data.access_token

    const response = await apiClient.get('http://localhost:8080/api/v1/agents', {
      headers: { Authorization: `Bearer ${token}` },
    })
    const data = await response.json()
    return data.data?.[0]
  }

  test('should display status dashboard card', async ({ authenticatedPage, apiClient }) => {
    const testAgent = await getFirstAgent(apiClient)
    if (testAgent) {
      await agentPage.goto(testAgent.id)
      await agentPage.waitForPageLoad()

      // Should show status card
      await expect(authenticatedPage.getByText('实时状态')).toBeVisible()
    }
  })

  test('should display disk usage gauge', async ({ authenticatedPage, apiClient }) => {
    const testAgent = await getFirstAgent(apiClient)
    if (testAgent) {
      await agentPage.goto(testAgent.id)
      await agentPage.waitForPageLoad()

      // Should show disk usage label
      await expect(authenticatedPage.getByText('磁盘使用')).toBeVisible()

      // Should show progress circle
      await expect(agentPage.diskProgress).toBeVisible()
    }
  })

  test('should display CPU and memory gauges', async ({ authenticatedPage, apiClient }) => {
    const testAgent = await getFirstAgent(apiClient)
    if (testAgent) {
      await agentPage.goto(testAgent.id)
      await agentPage.waitForPageLoad()

      // Should show gauge canvases (ECharts renders as canvas)
      await expect(agentPage.cpuGauge).toBeVisible()
      await expect(agentPage.memoryGauge).toBeVisible()
    }
  })

  test('should display history metrics section', async ({ authenticatedPage, apiClient }) => {
    const testAgent = await getFirstAgent(apiClient)
    if (testAgent) {
      await agentPage.goto(testAgent.id)
      await agentPage.waitForPageLoad()

      // Should show history metrics card
      await expect(authenticatedPage.getByText('历史指标')).toBeVisible()

      // Should show time range buttons
      await expect(authenticatedPage.getByRole('button', { name: '1 小时' })).toBeVisible()
      await expect(authenticatedPage.getByRole('button', { name: '24 小时' })).toBeVisible()
      await expect(authenticatedPage.getByRole('button', { name: '7 天' })).toBeVisible()
    }
  })

  test('should switch time ranges', async ({ authenticatedPage, apiClient }) => {
    const testAgent = await getFirstAgent(apiClient)
    if (testAgent) {
      await agentPage.goto(testAgent.id)
      await agentPage.waitForPageLoad()

      // Click 24 hour button
      await agentPage.selectTimeRange('24h')

      // Verify button is selected (has primary class)
      const button = authenticatedPage.getByRole('button', { name: '24 小时' })
      await expect(button).toHaveClass(/ant-btn-primary/)

      // Click 7 day button
      await agentPage.selectTimeRange('7d')
      const button7d = authenticatedPage.getByRole('button', { name: '7 天' })
      await expect(button7d).toHaveClass(/ant-btn-primary/)
    }
  })

  test('should display recent tasks section', async ({ authenticatedPage, apiClient }) => {
    const testAgent = await getFirstAgent(apiClient)
    if (testAgent) {
      await agentPage.goto(testAgent.id)
      await agentPage.waitForPageLoad()

      // Should show recent tasks card
      await expect(authenticatedPage.getByText('最近任务')).toBeVisible()
    }
  })
})
