import apiClient from './client'
import type { User, ApiResponse } from '@/types'

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
  expires_in: number
  user: User
}

export interface RefreshTokenResponse {
  access_token: string
  expires_in: number
}

export const authApi = {
  async login(data: LoginRequest): Promise<LoginResponse> {
    const response = await apiClient.post<LoginResponse>('/auth/login', data)
    return response.data
  },

  async logout(): Promise<void> {
    await apiClient.post('/auth/logout')
  },

  async refreshToken(refreshToken: string): Promise<RefreshTokenResponse> {
    const response = await apiClient.post<RefreshTokenResponse>('/auth/refresh', {
      refresh_token: refreshToken,
    })
    return response.data
  },

  async getCurrentUser(): Promise<User> {
    const response = await apiClient.get<ApiResponse<User>>('/auth/me')
    return response.data.data
  },
}

export default authApi
