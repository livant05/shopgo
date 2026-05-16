import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface QuoteItem {
  product_id: string
  sku: string
  name: string
  unit_price: number
  qty: number
  image_url?: string
}

export interface QuoteHistoryItem {
  id: string
  quoteNumber: number
  total: number
  createdAt: string
  customerName: string
}

export const useQuoteStore = defineStore('quote', () => {
  const items   = ref<QuoteItem[]>([])
  const isOpen  = ref(false)
  const history = ref<QuoteHistoryItem[]>([])

  const count    = computed(() => items.value.reduce((s, i) => s + i.qty, 0))
  const subtotal = computed(() => items.value.reduce((s, i) => s + i.unit_price * i.qty, 0))
  const isEmpty  = computed(() => items.value.length === 0)

  function add(item: Omit<QuoteItem, 'qty'>) {
    const ex = items.value.find(i => i.product_id === item.product_id)
    if (ex) { ex.qty++ }
    else items.value.push({ ...item, qty: 1 })
    isOpen.value = true
    save()
  }

  function remove(id: string) {
    items.value = items.value.filter(i => i.product_id !== id)
    save()
  }

  function setQty(id: string, n: number) {
    const item = items.value.find(i => i.product_id === id)
    if (!item) return
    if (n <= 0) { remove(id); return }
    item.qty = n
    save()
  }

  function clear() { items.value = []; save() }

  function save() {
    localStorage.setItem('quote_items', JSON.stringify(items.value))
  }

  function saveToHistory(entry: QuoteHistoryItem) {
    if (history.value.find(h => h.id === entry.id)) return
    history.value.unshift(entry)
    localStorage.setItem('quote_history', JSON.stringify(history.value))
  }

  function restore() {
    try {
      const raw = localStorage.getItem('quote_items')
      if (raw) items.value = JSON.parse(raw)
    } catch {}
    try {
      const raw = localStorage.getItem('quote_history')
      if (raw) history.value = JSON.parse(raw)
    } catch {}
  }

  restore()

  return { items, isOpen, history, count, subtotal, isEmpty, add, remove, setQty, clear, saveToHistory }
})
