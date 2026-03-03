import { APIRequestContext } from '@playwright/test'

export interface LoginResponse {
  access_token: string
  refresh_token: string
  user: {
    id: string
    username: string
    role: string
  }
}

export interface AuthState {
  token: string
  refreshToken: string | null
  user: {
    id: string
    username: string
    role: string
  }
  isAuthenticated: boolean
}

const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:8080/api/v1'
const DEFAULT_USERNAME = 'admin'
const DEFAULT_PASSWORD = 'admin123'

export async function login(
  request: APIRequestContext,
  username: string = DEFAULT_USERNAME,
  password: string = DEFAULT_PASSWORD
): Promise<LoginResponse> {
  const response = await request.post(`${API_BASE_URL}/auth/login`, {
    data: { username, password },
  })

  if (!response.ok()) {
    throw new Error(`Login failed: ${response.status()}`)
  }

  return response.json()
}

export function createAuthStorage(authResponse: LoginResponse): string {
  const authState = {
    state: {
      token: authResponse.access_token,
      refreshToken: authResponse.refresh_token,
      user: authResponse.user,
      isAuthenticated: true,
    },
    version: 0,
  }
  return JSON.stringify(authState)
}

export function getAuthHeaders(token: string): Record<string, string> {
  return {
    Authorization: `Bearer ${token}`,
  }
}
