import type { AlertMessage } from '@/types'

// Global alert notification handler
let alertHandler: ((alert: AlertMessage) => void) | null = null

export function setAlertHandler(handler: (alert: AlertMessage) => void) {
  alertHandler = handler
}

export function clearAlertHandler() {
  alertHandler = null
}

export function triggerAlertNotification(alert: AlertMessage) {
  if (alertHandler) {
    alertHandler(alert)
  }
}
