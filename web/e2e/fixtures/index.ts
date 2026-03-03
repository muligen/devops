import { test as base, Page, APIRequestContext } from '@playwright/test'

export interface TestFixtures {
  authenticatedPage: Page
  apiClient: APIRequestContext
}

const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:8080/api/v1'

export const test = base.extend<TestFixtures>({
  apiClient: async ({ request }, use) => {
    await use(request)
  },

  authenticatedPage: async ({ page, request }, use) => {
    // Login via API and get token
    const response = await request.post(`${API_BASE_URL}/auth/login`, {
      data: {
        username: 'admin',
        password: 'admin123',
      },
    })

    if (!response.ok()) {
      throw new Error(`Login failed: ${response.status()}`)
    }

    const data = await response.json()
    const accessToken = data.data.access_token

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
    }, accessToken)

    await use(page)
  },
})

// Helper function to get authorization header
export function getAuthHeaders(token: string): Record<string, string> {
  return {
    Authorization: `Bearer ${token}`,
    'Content-Type': 'application/json',
  }
}

export { expect } from '@playwright/test'
