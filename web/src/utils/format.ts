import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/zh-cn'

dayjs.extend(relativeTime)
dayjs.locale('zh-cn')

// Format date to relative time
export function formatRelativeTime(date: string | Date | null): string {
  if (!date) return '-'
  return dayjs(date).fromNow()
}

// Format date to standard format
export function formatDate(date: string | Date | null, format = 'YYYY-MM-DD HH:mm:ss'): string {
  if (!date) return '-'
  return dayjs(date).format(format)
}

// Format bytes to human readable
export function formatBytes(bytes: number, decimals = 2): string {
  if (bytes === 0) return '0 B'

  const k = 1024
  const dm = decimals < 0 ? 0 : decimals
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']

  const i = Math.floor(Math.log(bytes) / Math.log(k))

  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`
}

// Format percentage
export function formatPercent(value: number | undefined | null, decimals = 1): string {
  if (value === undefined || value === null) return '-'
  return `${value.toFixed(decimals)}%`
}

// Format duration in seconds to human readable
export function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}秒`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}分${seconds % 60}秒`
  if (seconds < 86400) {
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    return `${hours}小时${minutes}分`
  }
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  return `${days}天${hours}小时`
}

// Get status color
export function getStatusColor(status: string): string {
  const colors: Record<string, string> = {
    online: '#52c41a',
    offline: '#ff4d4f',
    pending: '#faad14',
    running: '#1890ff',
    completed: '#52c41a',
    failed: '#ff4d4f',
    cancelled: '#8c8c8c',
    acknowledged: '#1890ff',
    resolved: '#52c41a',
    critical: '#ff4d4f',
    warning: '#faad14',
    info: '#1890ff',
  }
  return colors[status] || '#8c8c8c'
}

// Get status text in Chinese
export function getStatusText(status: string): string {
  const texts: Record<string, string> = {
    online: '在线',
    offline: '离线',
    pending: '待处理',
    running: '运行中',
    completed: '已完成',
    failed: '失败',
    cancelled: '已取消',
    acknowledged: '已确认',
    resolved: '已解决',
    critical: '严重',
    warning: '警告',
    info: '信息',
  }
  return texts[status] || status
}

// Truncate text
export function truncateText(text: string, maxLength: number): string {
  if (text.length <= maxLength) return text
  return `${text.slice(0, maxLength)}...`
}
