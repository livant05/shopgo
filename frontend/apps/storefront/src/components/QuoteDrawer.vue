<template>
  <Teleport to="body">
    <Transition name="drawer">
      <div v-if="quote.isOpen" class="overlay" @click.self="quote.isOpen = false">
        <div class="drawer">

          <div class="drawer-header">
            <h2>📄 Cotización</h2>
            <button class="btn-close" @click="quote.isOpen = false">✕</button>
          </div>

          <!-- Vacía -->
          <div v-if="quote.isEmpty" class="empty">
            <p>No tienes productos en tu cotización.</p>
            <button class="btn-sec" @click="quote.isOpen = false">Explorar catálogo</button>
          </div>

          <template v-else>

            <!-- Paso 1: items -->
            <div v-if="step === 1" class="step-items">
              <div class="item-list">
                <div v-for="item in quote.items" :key="item.product_id" class="item-row">
                  <div class="item-img">
                    <img v-if="item.image_url" :src="item.image_url" :alt="item.name" />
                    <span v-else>📦</span>
                  </div>
                  <div class="item-info">
                    <p class="item-name">{{ item.name }}</p>
                    <p class="item-sku">{{ item.sku }}</p>
                    <p class="item-price">{{ fmt(item.unit_price) }} c/u</p>
                  </div>
                  <div class="item-controls">
                    <div class="qty-ctrl">
                      <button @click="quote.setQty(item.product_id, item.qty - 1)">−</button>
                      <span>{{ item.qty }}</span>
                      <button @click="quote.setQty(item.product_id, item.qty + 1)">+</button>
                    </div>
                    <p class="item-sub">{{ fmt(item.unit_price * item.qty) }}</p>
                    <button class="btn-remove" @click="quote.remove(item.product_id)">🗑</button>
                  </div>
                </div>
              </div>

              <!-- ITBMS preview -->
              <div class="totals-box">
                <div class="total-row"><span>Subtotal</span><span>{{ fmt(quote.subtotal) }}</span></div>
                <div class="total-row tax"><span>ITBMS (7%)</span><span>{{ fmt(quote.subtotal * 0.07) }}</span></div>
                <div class="total-row grand"><span>Total estimado</span><span>{{ fmt(quote.subtotal * 1.07) }}</span></div>
                <p class="tax-note">* Precios en Balboas (B/. = USD). Impuesto calculado: 7% ITBMS.</p>
              </div>

              <div class="step-actions">
                <button class="btn-ghost" @click="quote.clear()">Limpiar</button>
                <button class="btn-primary" @click="step = 2">Continuar →</button>
              </div>
            </div>

            <!-- Paso 2: datos del cliente -->
            <div v-else class="step-form">
              <p class="form-subtitle">Completa tus datos para generar la cotización formal.</p>

              <div class="form-grid">
                <label class="form-label" style="grid-column:1/-1">
                  Nombre completo *
                  <input v-model="form.customer_name" class="form-input" placeholder="Juan Pérez" />
                </label>
                <label class="form-label">
                  Correo electrónico
                  <input v-model="form.customer_email" type="email" class="form-input" placeholder="juan@empresa.com" />
                </label>
                <label class="form-label">
                  Teléfono
                  <input v-model="form.customer_phone" class="form-input" placeholder="+507 6000-0000" />
                </label>
                <label class="form-label" style="grid-column:1/-1">
                  Notas u observaciones
                  <textarea v-model="form.note" class="form-input" rows="3" placeholder="Entrega en oficina, fecha requerida, etc." />
                </label>
              </div>

              <p v-if="errMsg" class="err-msg">{{ errMsg }}</p>

              <div class="step-actions">
                <button class="btn-ghost" @click="step = 1">← Volver</button>
                <button class="btn-primary" :disabled="submitting" @click="submit()">
                  {{ submitting ? 'Generando…' : '📄 Generar cotización' }}
                </button>
              </div>
            </div>

          </template>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useQuoteStore } from '../stores/quote'
import { api } from '../api/client'

const quote = useQuoteStore()
const router = useRouter()

const step = ref(1)
const submitting = ref(false)
const errMsg = ref('')

const form = ref({
  customer_name: '',
  customer_email: '',
  customer_phone: '',
  note: '',
})

function fmt(v: number) {
  return `B/. ${v.toFixed(2)}`
}

async function submit() {
  if (!form.value.customer_name.trim()) {
    errMsg.value = 'El nombre es obligatorio.'
    return
  }
  submitting.value = true
  errMsg.value = ''
  try {
    const payload = {
      items: quote.items.map(i => ({
        product_id: i.product_id,
        sku: i.sku,
        name: i.name,
        qty: i.qty,
        unit_price: i.unit_price,
      })),
      ...form.value,
    }
    const { data } = await api.post('/quotes', payload)
    quote.clear()
    quote.isOpen = false
    step.value = 1
    form.value = { customer_name: '', customer_email: '', customer_phone: '', note: '' }
    router.push(`/quote/${data.id}`)
  } catch (e: any) {
    errMsg.value = e?.response?.data?.message ?? 'Error al generar la cotización.'
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,.5);
  z-index: 1000; display: flex; justify-content: flex-end;
}
.drawer {
  width: 460px; max-width: 100vw; height: 100%; background: #fff;
  display: flex; flex-direction: column; box-shadow: -4px 0 24px rgba(0,0,0,.2);
  overflow-y: auto;
}
.drawer-header {
  display: flex; justify-content: space-between; align-items: center;
  padding: 1.25rem 1.5rem; border-bottom: 1px solid #e2e8f0; flex-shrink: 0;
}
.drawer-header h2 { margin: 0; font-size: 1.15rem; font-weight: 700; color: #0f172a; }
.btn-close {
  background: none; border: none; font-size: 1.2rem; cursor: pointer; color: #64748b;
}

.empty { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 1rem; color: #94a3b8; padding: 2rem; }

.step-items, .step-form { display: flex; flex-direction: column; flex: 1; padding: 1.25rem 1.5rem; gap: 1rem; }

/* Items */
.item-list { display: flex; flex-direction: column; gap: .75rem; }
.item-row { display: flex; gap: .75rem; align-items: flex-start; padding: .75rem; background: #f8fafc; border-radius: .75rem; }
.item-img {
  width: 56px; height: 56px; border-radius: 8px; overflow: hidden;
  background: #e2e8f0; display: flex; align-items: center; justify-content: center;
  font-size: 1.5rem; flex-shrink: 0;
}
.item-img img { width: 100%; height: 100%; object-fit: cover; }
.item-info { flex: 1; min-width: 0; }
.item-name { font-size: .875rem; font-weight: 600; color: #1e293b; margin: 0; }
.item-sku  { font-size: .75rem; color: #94a3b8; margin: .1rem 0; }
.item-price { font-size: .8rem; color: #3b82f6; margin: 0; }
.item-controls { display: flex; flex-direction: column; align-items: flex-end; gap: .35rem; flex-shrink: 0; }
.qty-ctrl { display: flex; align-items: center; border: 1px solid #e2e8f0; border-radius: 6px; overflow: hidden; }
.qty-ctrl button { width: 28px; height: 28px; border: none; background: #f1f5f9; cursor: pointer; font-size: 1rem; }
.qty-ctrl button:hover { background: #e2e8f0; }
.qty-ctrl span { width: 30px; text-align: center; font-size: .875rem; font-weight: 600; }
.item-sub { font-size: .875rem; font-weight: 700; color: #0f172a; }
.btn-remove { background: none; border: none; cursor: pointer; font-size: .9rem; color: #94a3b8; padding: .2rem; }
.btn-remove:hover { color: #ef4444; }

/* Totals */
.totals-box { background: #f8fafc; border-radius: .75rem; padding: 1rem; display: flex; flex-direction: column; gap: .4rem; }
.total-row { display: flex; justify-content: space-between; font-size: .875rem; color: #475569; }
.total-row.tax { color: #64748b; }
.total-row.grand { font-weight: 700; font-size: 1rem; color: #0f172a; border-top: 1px solid #e2e8f0; padding-top: .5rem; margin-top: .1rem; }
.tax-note { font-size: .72rem; color: #94a3b8; margin: .25rem 0 0; }

/* Form */
.form-subtitle { font-size: .875rem; color: #64748b; margin: 0; }
.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: .85rem; }
.form-label { display: flex; flex-direction: column; gap: .3rem; font-size: .8rem; font-weight: 600; color: #374151; }
.form-input { padding: .5rem .65rem; border: 1px solid #d1d5db; border-radius: .5rem; font-size: .875rem; outline: none; }
.form-input:focus { border-color: #3b82f6; }

.err-msg { color: #dc2626; font-size: .82rem; background: #fef2f2; padding: .5rem .75rem; border-radius: .5rem; }

/* Actions */
.step-actions { display: flex; justify-content: space-between; gap: .75rem; margin-top: auto; padding-top: .5rem; }
.btn-primary {
  flex: 1; background: #1d4ed8; color: #fff; border: none; border-radius: .6rem;
  padding: .65rem 1.25rem; font-weight: 700; font-size: .9rem; cursor: pointer;
}
.btn-primary:hover:not(:disabled) { background: #1e40af; }
.btn-primary:disabled { opacity: .55; cursor: not-allowed; }
.btn-ghost { background: none; border: 1px solid #d1d5db; border-radius: .6rem; padding: .65rem 1rem; font-size: .875rem; cursor: pointer; color: #475569; }
.btn-ghost:hover { background: #f1f5f9; }
.btn-sec { background: #3b82f6; color: #fff; border: none; border-radius: .6rem; padding: .6rem 1.25rem; font-weight: 600; cursor: pointer; }

/* Transition */
.drawer-enter-active, .drawer-leave-active { transition: transform .3s ease; }
.drawer-enter-from, .drawer-leave-to { transform: translateX(100%); }
</style>
