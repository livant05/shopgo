import { ref, onUnmounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { api } from '../api/client'

export type NotificationType = 'order' | 'quote'

export interface AdminNotification {
  id: string
  type: NotificationType
  message: string
  resource_id?: string
  total?: number
  ts: number
}

const notifications  = ref<AdminNotification[]>([])
const pendingQuotes  = ref(0)

export function useNotifications() {
  const auth = useAuthStore()
  let es: EventSource | null = null

  async function fetchPendingQuotes() {
    try {
      const { data } = await api.get('/admin/quotes', { params: { status: 'pending', page_size: 1 } })
      pendingQuotes.value = data.total ?? 0
    } catch {}
  }

  function connect() {
    if (!auth.token || es) return
    fetchPendingQuotes()

    const url = `/api/v1/admin/notifications/stream`
    es = new EventSource(url)

    es.onopen = () => console.log('[SSE] connected')

    es.onmessage = (evt) => {
      try {
        const data = JSON.parse(evt.data)

        if (data.event === 'order.new') {
          const n: AdminNotification = {
            id: crypto.randomUUID(),
            type: 'order',
            message: `Nuevo pedido #${data.order_id?.slice(0, 8).toUpperCase()} — $${Number(data.total).toLocaleString('es-MX', { minimumFractionDigits: 2 })}`,
            resource_id: data.order_id,
            total: data.total,
            ts: Date.now(),
          }
          notifications.value.unshift(n)
          setTimeout(() => dismiss(n.id), 8000)
        }

        if (data.event === 'quote.new') {
          const num = String(data.quote_number ?? 0).padStart(5, '0')
          const n: AdminNotification = {
            id: crypto.randomUUID(),
            type: 'quote',
            message: `N.° ${num} — ${data.customer_name ?? 'Cliente'} · B/. ${Number(data.total).toFixed(2)}`,
            resource_id: data.quote_id,
            total: data.total,
            ts: Date.now(),
          }
          notifications.value.unshift(n)
          pendingQuotes.value++
          setTimeout(() => dismiss(n.id), 8000)
        }

        if (data.event === 'quote.status_changed' &&
            (data.new_status === 'accepted' || data.new_status === 'rejected')) {
          pendingQuotes.value = Math.max(0, pendingQuotes.value - 1)
        }
      } catch {}
    }

    es.onerror = () => {
      es?.close(); es = null
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

  return { notifications, pendingQuotes, connect, disconnect, dismiss }
}
