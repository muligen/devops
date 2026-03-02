import { create } from 'zustand'
import type { DashboardStats } from '@/types'

interface DashboardState {
  stats: DashboardStats | null
  metrics: Record<string, { cpu_usage: number; memory_usage: number; disk_usage: number }>
  loading: boolean
  error: string | null

  // Actions
  setStats: (stats: DashboardStats) => void
  updateMetrics: (metrics: Record<string, unknown>) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
}

export const useDashboardStore = create<DashboardState>((set) => ({
  stats: null,
  metrics: {},
  loading: false,
  error: null,

  setStats: (stats: DashboardStats) => {
    set({ stats, error: null })
  },

  updateMetrics: (metrics: Record<string, unknown>) => {
    set((state) => ({
      metrics: {
        ...state.metrics,
        ...Object.fromEntries(
          Object.entries(metrics).map(([agentId, data]) => [
            agentId,
            data as { cpu_usage: number; memory_usage: number; disk_usage: number },
          ])
        ),
      },
    }))
  },

  setLoading: (loading: boolean) => {
    set({ loading })
  },

  setError: (error: string | null) => {
    set({ error })
  },
}))

export default useDashboardStore
