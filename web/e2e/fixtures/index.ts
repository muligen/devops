import { test as base, Page } from '@playwright/test'

export interface TestFixtures {
  authenticatedPage: Page
}

export const test = base.extend<TestFixtures>({
  authenticatedPage: async ({ page }, use) => {
    // Login via API and store auth state
    const response = await page.request.post('http://localhost:8080/api/v1/auth/login', {
      data: {
        username: 'admin',
        password: 'admin123',
      },
    })

    if (!response.ok()) {
      throw new Error(`Login failed: ${response.status()}`)
    }

    const data = await response.json()

    // Set auth state in localStorage
    await page.goto('/')
    await page.evaluate((token) => {
      const authState = {
        state: {
          token: token,
          refreshToken: null,
          user: {
            id: 'test-user',
            username: 'admin',
            role: 'admin',
          },
          isAuthenticated: true,
        },
        version: 0,
      }
      localStorage.setItem('auth-storage', JSON.stringify(authState))
    }, data.access_token)

    await use(page)
  },
})

export { expect } from '@playwright/test'
