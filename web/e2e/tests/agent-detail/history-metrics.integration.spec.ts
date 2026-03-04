import { test, expect } from '../../fixtures'
import { AgentDetailPage } from '../../pages/agent-detail.page'

test.describe('Agent Detail - History Metrics', () => {
  let agentPage: AgentDetailPage

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
    return { agent: data.data?.[0], token }
  }

  test.beforeEach(async ({ authenticatedPage }) => {
    agentPage = new AgentDetailPage(authenticatedPage)
  })

  test('should fetch and display metrics history data', async ({ authenticatedPage, apiClient }) => {
    const { agent, token } = await getFirstAgent(apiClient)

    if (agent) {
      // 验证 metrics API 返回数据
      const metricsResponse = await apiClient.get(`http://localhost:8080/api/v1/agents/${agent.id}/metrics?range=1h`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      const metricsData = await metricsResponse.json()

      // 验证 API 返回了指标数据
      expect(metricsData.code).toBe(0)
      expect(Array.isArray(metricsData.data)).toBe(true)
      expect(metricsData.data.length).toBeGreaterThan(0)

      // 打印第一条数据用于调试
      console.log('First metric:', JSON.stringify(metricsData.data[0]))
    }
  })

  test('should render history chart with actual data points', async ({ authenticatedPage, apiClient }) => {
    const { agent } = await getFirstAgent(apiClient)

    if (agent) {
      await agentPage.goto(agent.id)
      await agentPage.waitForPageLoad()

      // 等待历史指标卡片加载
      await expect(authenticatedPage.getByText('历史指标')).toBeVisible()

      // 检查 ECharts 图表是否渲染
      const chartCanvas = authenticatedPage.locator('.ant-card').filter({ hasText: '历史指标' }).locator('canvas')
      await expect(chartCanvas).toBeVisible({ timeout: 10000 })

      // 检查时间范围按钮
      await expect(authenticatedPage.getByRole('button', { name: '1 小时' })).toBeVisible()
    }
  })

  test('should update chart when switching time range', async ({ authenticatedPage, apiClient }) => {
    const { agent } = await getFirstAgent(apiClient)

    if (agent) {
      await agentPage.goto(agent.id)
      await agentPage.waitForPageLoad()

      // 等待历史指标卡片
      await expect(authenticatedPage.getByText('历史指标')).toBeVisible()

      // 点击 24 小时按钮
      await agentPage.selectTimeRange('24h')

      // 验证按钮被选中
      const button24h = authenticatedPage.getByRole('button', { name: '24 小时' })
      await expect(button24h).toHaveClass(/ant-btn-primary/)

      // 点击 7 天按钮
      await agentPage.selectTimeRange('7d')
      const button7d = authenticatedPage.getByRole('button', { name: '7 天' })
      await expect(button7d).toHaveClass(/ant-btn-primary/)
    }
  })

  test('should show chart legend with CPU, Memory, Disk labels', async ({ authenticatedPage, apiClient }) => {
    const { agent } = await getFirstAgent(apiClient)

    if (agent) {
      await agentPage.goto(agent.id)
      await agentPage.waitForPageLoad()

      // 检查图例显示
      await expect(authenticatedPage.getByText('CPU')).toBeVisible()
      await expect(authenticatedPage.getByText('内存')).toBeVisible()
      await expect(authenticatedPage.getByText('磁盘')).toBeVisible()
    }
  })
})