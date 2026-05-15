import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '../api/client'

export interface CartItem {
  product_id: string; name: string; sku: string
  unit_price: number; quantity: number; image_url?: string; stock: number
}

export const useCartStore = defineStore('cart', () => {
  const items   = ref<CartItem[]>([])
  const isOpen  = ref(false)
  const coupon  = ref('')
  const discount = ref(0)

  const count    = computed(() => items.value.reduce((s, i) => s + i.quantity, 0))
  const subtotal = computed(() => items.value.reduce((s, i) => s + i.unit_price * i.quantity, 0))
  const total    = computed(() => subtotal.value - discount.value)
  const isEmpty  = computed(() => items.value.length === 0)

  function add(item: Omit<CartItem,'quantity'>) {
    const ex = items.value.find(i => i.product_id === item.product_id)
    if (ex) { if (ex.quantity < ex.stock) ex.quantity++ }
    else items.value.push({ ...item, quantity: 1 })
    isOpen.value = true
    save()
  }

  function remove(id: string) { items.value = items.value.filter(i => i.product_id !== id); save() }
  function qty(id: string, n: number) {
    const i = items.value.find(x => x.product_id === id)
    if (!i) return
    if (n <= 0) return remove(id)
    i.quantity = Math.min(n, i.stock); save()
  }
  function clear() { items.value = []; coupon.value = ''; discount.value = 0; save() }

  async function applyCoupon(code: string) {
    try {
      const r = await api.post('/coupons/validate', { code, subtotal: subtotal.value })
      coupon.value = code; discount.value = r.data.discount
      return { ok: true }
    } catch { return { ok: false, message: 'Cupón inválido' } }
  }

  function save() { localStorage.setItem('cart', JSON.stringify(items.value)) }
  function restore() {
    try { const d = localStorage.getItem('cart'); if (d) items.value = JSON.parse(d) } catch {}
  }
  restore()

  return { items, isOpen, coupon, discount, count, subtotal, total, isEmpty, add, remove, qty, clear, applyCoupon }
})
