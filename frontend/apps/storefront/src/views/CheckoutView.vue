<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api/client'
import { useCartStore } from '../stores/cart'
import type { Branch } from '../types'

const router = useRouter()
const cart   = useCartStore()

const branches  = ref<Branch[]>([])
const loading   = ref(false)
const error     = ref('')

const form = ref({
  branch_id: '',
  street:  '', city: '', state: '', zip: '', country: 'México',
  notes: '',
})

onMounted(async () => {
  try {
    const { data } = await api.get('/branches')
    branches.value = (data.data ?? []).filter((b: Branch) => b.is_active)
    if (branches.value.length) form.value.branch_id = branches.value[0].id
  } catch {}
})

const fmt = (n: number) => n.toLocaleString('es-MX', { style: 'currency', currency: 'MXN' })

const subtotal = computed(() => cart.subtotal)
const tax      = computed(() => +(subtotal.value * 0.16).toFixed(2))
const total    = computed(() => +(subtotal.value + tax.value - cart.discount).toFixed(2))

async function placeOrder() {
  if (!form.value.branch_id) { error.value = 'Selecciona una sucursal'; return }
  if (!form.value.street || !form.value.city) { error.value = 'Completa la dirección'; return }
  error.value = ''
  loading.value = true
  try {
    const payload = {
      branch_id: form.value.branch_id,
      items: cart.items.map(i => ({
        product_id: i.product_id, quantity: i.quantity, unit_price: i.unit_price,
      })),
      subtotal:  subtotal.value,
      tax:       tax.value,
      discount:  cart.discount,
      total:     total.value,
      currency:  'MXN',
      coupon_code: cart.coupon || undefined,
      shipping_address: {
        street: form.value.street, city: form.value.city,
        state:  form.value.state,  zip:  form.value.zip,
        country: form.value.country,
      },
      notes: form.value.notes || undefined,
      quote_id: cart.fromQuoteId || undefined,
    }
    const { data } = await api.post('/orders', payload)
    cart.clear()
    router.push({ name: 'Success', params: { id: data.id ?? 'ok' } })
  } catch (e: any) {
    error.value = e.response?.data?.message ?? 'Error al crear el pedido'
  }
  loading.value = false
}
</script>

<template>
  <div class="container">
    <h1 class="title">Finalizar compra</h1>

    <div class="layout">

      <!-- Form -->
      <form class="form" @submit.prevent="placeOrder">

        <section class="card">
          <h2 class="card-title">Sucursal de entrega</h2>
          <select v-model="form.branch_id" class="input">
            <option v-for="b in branches" :key="b.id" :value="b.id">{{ b.name }}</option>
            <option v-if="!branches.length" value="">Sin sucursales disponibles</option>
          </select>
        </section>

        <section class="card">
          <h2 class="card-title">Dirección de envío</h2>
          <div class="grid2">
            <div class="field">
              <label>Calle y número *</label>
              <input v-model="form.street" class="input" required placeholder="Av. Siempre Viva 742" />
            </div>
            <div class="field">
              <label>Ciudad *</label>
              <input v-model="form.city" class="input" required placeholder="Ciudad de México" />
            </div>
            <div class="field">
              <label>Estado</label>
              <input v-model="form.state" class="input" placeholder="CDMX" />
            </div>
            <div class="field">
              <label>Código postal</label>
              <input v-model="form.zip" class="input" placeholder="06600" />
            </div>
          </div>
        </section>

        <section class="card">
          <h2 class="card-title">Notas del pedido (opcional)</h2>
          <textarea v-model="form.notes" class="input" rows="3" placeholder="Instrucciones especiales…" />
        </section>

        <p v-if="error" class="error-msg">⚠️ {{ error }}</p>

        <button type="submit" class="btn-place" :disabled="loading || cart.isEmpty">
          {{ loading ? 'Procesando…' : `Confirmar pedido · ${fmt(total)}` }}
        </button>

      </form>

      <!-- Order summary -->
      <div class="summary">
        <h2>Tu pedido</h2>
        <div v-for="item in cart.items" :key="item.product_id" class="order-item">
          <div class="item-left">
            <div class="item-thumb">
              <img v-if="item.image_url" :src="item.image_url" :alt="item.name" />
              <span v-else>📦</span>
            </div>
            <div>
              <p class="item-name">{{ item.name }}</p>
              <p class="item-qty">× {{ item.quantity }}</p>
            </div>
          </div>
          <span class="item-price">{{ fmt(item.unit_price * item.quantity) }}</span>
        </div>
        <hr />
        <div class="sum-row"><span>Subtotal</span><span>{{ fmt(subtotal) }}</span></div>
        <div class="sum-row"><span>IVA (16%)</span><span>{{ fmt(tax) }}</span></div>
        <div v-if="cart.discount > 0" class="sum-row green">
          <span>Descuento</span><span>−{{ fmt(cart.discount) }}</span>
        </div>
        <hr />
        <div class="sum-row total"><span>Total</span><strong>{{ fmt(total) }}</strong></div>
      </div>

    </div>
  </div>
</template>

<style scoped>
.container { max-width: 1100px; margin: 0 auto; padding: 2rem 1.5rem; }
.title { font-size: 1.75rem; font-weight: 800; margin-bottom: 2rem; }
.layout { display: grid; grid-template-columns: 1fr 360px; gap: 2rem; align-items: start; }

.card { background: #fff; border-radius: 16px; padding: 1.5rem; box-shadow: 0 1px 4px rgba(0,0,0,.07); margin-bottom: 1.25rem; }
.card-title { font-size: 1.05rem; font-weight: 700; margin-bottom: 1rem; }
.grid2 { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
.field { display: flex; flex-direction: column; gap: .35rem; }
.field label { font-size: .8rem; font-weight: 600; color: #64748b; text-transform: uppercase; letter-spacing: .04em; }
.input {
  width: 100%; padding: .55rem .85rem; border: 1px solid #e2e8f0; border-radius: 8px;
  font-size: .9rem; outline: none; font-family: inherit; color: #1e293b;
  transition: border-color .15s;
}
.input:focus { border-color: #3b82f6; }
textarea.input { resize: vertical; }
.error-msg { color: #ef4444; font-size: .9rem; margin-bottom: .75rem; }
.btn-place {
  width: 100%; padding: 1rem; background: #1d4ed8; color: #fff;
  border: none; border-radius: 12px; font-size: 1rem; font-weight: 700;
  cursor: pointer; transition: background .15s;
}
.btn-place:hover:not(:disabled) { background: #1e40af; }
.btn-place:disabled { background: #e2e8f0; color: #94a3b8; cursor: not-allowed; }

.summary { background: #fff; border-radius: 16px; padding: 1.5rem; box-shadow: 0 2px 8px rgba(0,0,0,.08); position: sticky; top: 80px; }
.summary h2 { font-size: 1.05rem; font-weight: 700; margin-bottom: 1.25rem; }
.order-item { display: flex; justify-content: space-between; align-items: center; gap: .75rem; margin-bottom: .9rem; }
.item-left { display: flex; align-items: center; gap: .6rem; flex: 1; min-width: 0; }
.item-thumb { width: 44px; height: 44px; border-radius: 8px; background: #f8fafc; display: flex; align-items: center; justify-content: center; overflow: hidden; flex-shrink: 0; font-size: 1.2rem; }
.item-thumb img { width: 100%; height: 100%; object-fit: cover; }
.item-name { font-size: .85rem; font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.item-qty  { font-size: .75rem; color: #94a3b8; }
.item-price { font-weight: 600; font-size: .9rem; white-space: nowrap; }
hr { border: none; border-top: 1px solid #e2e8f0; margin: .75rem 0; }
.sum-row { display: flex; justify-content: space-between; font-size: .9rem; margin-bottom: .6rem; color: #475569; }
.sum-row.green { color: #10b981; }
.sum-row.total { font-size: 1.1rem; color: #1e293b; }

@media (max-width: 768px) {
  .layout { grid-template-columns: 1fr; }
  .grid2  { grid-template-columns: 1fr; }
}
</style>
