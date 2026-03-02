import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { User } from '@/types'
import authApi from '@/api/auth'

interface AuthState {
  token: string | null
  refreshToken: string | null
  user: User | null
  isAuthenticated: boolean
  loading: boolean
  error: string | null

  // Actions
  login: (username: string, password: string) => Promise<void>
  logout: () => void
  setUser: (user: User) => void
  refreshAuth: () => Promise<void>
  clearError: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      refreshToken: null,
      user: null,
      isAuthenticated: false,
      loading: false,
      error: null,

      login: async (username: string, password: string) => {
        set({ loading: true, error: null })
        try {
          const response = await authApi.login({ username, password })
          set({
            token: response.access_token,
            refreshToken: response.refresh_token,
            user: response.user,
            isAuthenticated: true,
            loading: false,
          })
        } catch (err) {
          const errorMessage = err instanceof Error ? err.message : '登录失败'
          set({
            error: errorMessage,
            loading: false,
            isAuthenticated: false,
          })
          throw err
        }
      },

      logout: () => {
        // Call logout API if token exists
        const { token } = get()
        if (token) {
          authApi.logout().catch(() => {
            // Ignore logout API errors
          })
        }
        set({
          token: null,
          refreshToken: null,
          user: null,
          isAuthenticated: false,
          error: null,
        })
      },

      setUser: (user: User) => {
        set({ user })
      },

      refreshAuth: async () => {
        const { refreshToken } = get()
        if (!refreshToken) {
          throw new Error('No refresh token')
        }
        try {
          const response = await authApi.refreshToken(refreshToken)
          set({ token: response.access_token })
        } catch (err) {
          // Refresh failed, logout user
          get().logout()
          throw err
        }
      },

      clearError: () => {
        set({ error: null })
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        token: state.token,
        refreshToken: state.refreshToken,
        user: state.user,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
)

export default useAuthStore
