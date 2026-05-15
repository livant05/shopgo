import { ref, onUnmounted } from 'vue'
import { useAuthStore } from '../stores/auth'

export interface AdminNotification {
  id: string
  message: string
  order_id?: string
  total?: number
  ts: number
}

const notifications = ref<AdminNotification[]>([])

export function useNotifications() {
  const auth = useAuthStore()
  let es: EventSource | null = null

  function connect() {
    if (!auth.token || es) return
    const url = `/api/v1/admin/notifications/stream`
    es = new EventSource(url)

    es.onopen = () => console.log('[SSE] connected')

    es.onmessage = (evt) => {
      try {
        const data = JSON.parse(evt.data)
        if (data.event === 'order.new') {
          const n: AdminNotification = {
            id: crypto.randomUUID(),
            message: `Nuevo pedido #${data.order_id?.slice(0, 8).toUpperCase()} — $${Number(data.total).toLocaleString('es-MX', { minimumFractionDigits: 2 })}`,
            order_id: data.order_id,
            total: data.total,
            ts: Date.now(),
          }
          notifications.value.unshift(n)
          // Auto-dismiss after 8 seconds
          setTimeout(() => dismiss(n.id), 8000)
        }
      } catch {}
    }

    es.onerror = () => {
      es?.close(); es = null
      // Reconnect after 5 seconds
      setTimeout(connect, 5000)
    }
  }

  function disconnect() {
    es?.close(); es = null
  }

  function dismiss(id: string) {
    notifications.value = notifications.value.filter(n => n.id !== id)
  }

  onUnmounted(disconnect)

  return { notifications, connect, disconnect, dismiss }
}
