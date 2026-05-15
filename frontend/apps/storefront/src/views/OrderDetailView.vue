<template>
  <div class="page">
    <div class="container">

      <!-- Back -->
      <router-link to="/orders" class="back-link">← Mis pedidos</router-link>

      <div v-if="loading" class="loading">
        <div class="spinner"></div>
        <p>Cargando pedido…</p>
      </div>

      <div v-else-if="!order" class="not-found">
        <p class="nf-icon">🔍</p>
        <h2>Pedido no encontrado</h2>
        <router-link to="/orders" class="btn-primary">Ver mis pedidos</router-link>
      </div>

      <template v-else>
        <!-- Header -->
        <div class="order-header">
          <div>
            <h1 class="order-id">Pedido <span class="mono">#{{ order.id.slice(0,8).toUpperCase() }}</span></h1>
            <p class="order-date">Realizado el {{ fmtDate(order.created_at) }}</p>
          </div>
          <span class="status-badge large" :class="order.status">{{ LABELS[order.status] ?? order.status }}</span>
        </div>

        <!-- Timeline -->
        <div class="timeline-card">
          <h2 class="section-title">Estado del pedido</h2>
          <div class="timeline" v-if="order.status !== 'cancelled'">
            <div v-for="(step, idx) in steps" :key="step.key"
              class="step" :class="stepClass(step.key)">
              <div class="step-connector" v-if="idx > 0" :class="stepDone(steps[idx-1].key) ? 'done' : ''"></div>
              <div class="step-dot">
                <span v-if="stepDone(step.key) && step.key !== currentStatus">✓</span>
                <span v-else-if="step.key === currentStatus" class="dot-pulse"></span>
                <span v-else class="dot-empty"></span>
              </div>
              <div class="step-info">
                <p class="step-label">{{ step.label }}</p>
                <p class="step-desc">{{ step.desc }}</p>
              </div>
            </div>
          </div>

          <div v-else class="cancelled-box">
            <span class="cancelled-icon">✕</span>
            <div>
              <p class="cancelled-title">Pedido cancelado</p>
              <p class="cancelled-sub">Este pedido fue cancelado y no se procesará.</p>
            </div>
          </div>
        </div>

        <!-- Items -->
        <div class="detail-card">
          <h2 class="section-title">Artículos ({{ order.items.length }})</h2>
          <div class="items-list">
            <div v-for="item in order.items" :key="item.product_id" class="item-row">
              <div class="item-thumb">📦</div>
              <div class="item-info">
                <p class="item-name">{{ item.name }}</p>
                <p class="item-sku">SKU: {{ item.sku }}</p>
              </div>
              <div class="item-qty">×{{ item.quantity }}</div>
              <div class="item-price">{{ fmt(item.unit_price) }}</div>
              <div class="item-total">{{ fmt(item.line_total) }}</div>
            </div>
          </div>

          <div class="totals">
            <div class="tot-row"><span>Subtotal</span><span>{{ fmt(order.subtotal) }}</span></div>
            <div class="tot-row discount" v-if="(order.discount ?? 0) > 0">
              <span>Descuento</span><span>−{{ fmt(order.discount) }}</span>
            </div>
            <div class="tot-row" v-if="(order.tax ?? 0) > 0"><span>IVA</span><span>{{ fmt(order.tax) }}</span></div>
            <div class="tot-row grand"><span>Total</span><strong>{{ fmt(order.total) }}</strong></div>
          </div>
        </div>

        <!-- Actions -->
        <div class="actions">
          <router-link to="/catalog" class="btn-secondary">Seguir comprando</router-link>
          <router-link to="/orders" class="btn-outline">Mis pedidos</router-link>
        </div>
      </template>

    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '../api/client'
import type { Order } from '../types'

const route   = useRoute()
const order   = ref<Order | null>(null)
const loading = ref(true)

const STATUS_ORDER = ['pending', 'confirmed', 'processing', 'shipped', 'delivered']

const steps = [
  { key: 'pending',    label: 'Pedido recibido',  desc: 'Tu pedido ha sido registrado exitosamente.' },
  { key: 'confirmed',  label: 'Confirmado',        desc: 'El pago fue verificado y el pedido confirmado.' },
  { key: 'processing', label: 'En preparación',    desc: 'Tu pedido está siendo preparado.' },
  { key: 'shipped',    label: 'En camino',          desc: 'Tu pedido salió para entrega.' },
  { key: 'delivered',  label: 'Entregado',          desc: '¡Tu pedido fue entregado con éxito!' },
]

const LABELS: Record<string, string> = {
  pending:    'Pendiente',
  confirmed:  'Confirmado',
  processing: 'En proceso',
  shipped:    'Enviado',
  delivered:  'Entregado',
  cancelled:  'Cancelado',
}

const currentStatus = computed(() => order.value?.status ?? '')
const currentIdx    = computed(() => STATUS_ORDER.indexOf(currentStatus.value))

function stepDone(key: string) {
  return STATUS_ORDER.indexOf(key) <= currentIdx.value
}
function stepClass(key: string) {
  const idx = STATUS_ORDER.indexOf(key)
  if (idx < currentIdx.value)  return 'done'
  if (idx === currentIdx.value) return 'current'
  return 'pending-step'
}

const fmt     = (n: number) => (n ?? 0).toLocaleString('es-MX', { style: 'currency', currency: 'MXN' })
const fmtDate = (s: string) => new Date(s).toLocaleDateString('es-MX', { year:'numeric', month:'long', day:'numeric' })

onMounted(async () => {
  try {
    const { data } = await api.get(`/orders/${route.params.id}`)
    order.value = data
  } catch {}
  loading.value = false
})
</script>

<style scoped>
.page { min-height: 80vh; background: #f1f5f9; padding: 2rem 1rem; }
.container { max-width: 740px; margin: 0 auto; display: flex; flex-direction: column; gap: 1.5rem; }

.back-link { color: #3b82f6; text-decoration: none; font-size: .875rem; font-weight: 600; }
.back-link:hover { text-decoration: underline; }

.loading { display: flex; flex-direction: column; align-items: center; gap: 1rem; padding: 4rem; color: #64748b; }
.spinner { width: 36px; height: 36px; border: 3px solid #e2e8f0; border-top-color: #3b82f6; border-radius: 50%; animation: spin .7s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

.not-found { text-align: center; padding: 4rem; }
.nf-icon   { font-size: 3rem; margin-bottom: .5rem; }
.not-found h2 { font-size: 1.3rem; font-weight: 700; color: #1e293b; margin-bottom: 1.5rem; }

/* Header */
.order-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 1rem; flex-wrap: wrap; }
.order-id   { font-size: 1.5rem; font-weight: 800; color: #1e293b; margin: 0 0 .25rem; }
.mono       { font-family: monospace; letter-spacing: .05em; }
.order-date { font-size: .85rem; color: #64748b; margin: 0; }

.status-badge { padding: .3rem .9rem; border-radius: 999px; font-size: .8rem; font-weight: 700; }
.status-badge.large { padding: .4rem 1.2rem; font-size: .9rem; }
.status-badge.pending    { background: #fef9c3; color: #854d0e; }
.status-badge.confirmed  { background: #dbeafe; color: #1d4ed8; }
.status-badge.processing { background: #ede9fe; color: #6d28d9; }
.status-badge.shipped    { background: #e0f2fe; color: #0369a1; }
.status-badge.delivered  { background: #dcfce7; color: #15803d; }
.status-badge.cancelled  { background: #fee2e2; color: #b91c1c; }

/* Cards */
.timeline-card, .detail-card {
  background: #fff; border-radius: 16px; padding: 1.75rem;
  box-shadow: 0 1px 4px rgba(0,0,0,.07);
}
.section-title { font-size: 1rem; font-weight: 700; color: #1e293b; margin: 0 0 1.5rem; }

/* Timeline */
.timeline { display: flex; flex-direction: column; gap: 0; }
.step     { display: flex; align-items: flex-start; gap: 1rem; position: relative; padding-bottom: 1.5rem; }
.step:last-child { padding-bottom: 0; }

.step-connector { position: absolute; left: 15px; top: -1.5rem; width: 2px; height: 1.5rem; background: #e2e8f0; }
.step-connector.done { background: #3b82f6; }

.step-dot { width: 32px; height: 32px; border-radius: 50%; border: 2px solid #e2e8f0; background: #fff; display: flex; align-items: center; justify-content: center; font-size: .85rem; font-weight: 700; flex-shrink: 0; z-index: 1; }
.step.done    .step-dot { background: #3b82f6; border-color: #3b82f6; color: #fff; }
.step.current .step-dot { background: #fff; border-color: #3b82f6; color: #3b82f6; }
.dot-pulse { width: 10px; height: 10px; background: #3b82f6; border-radius: 50%; animation: pulse 1.4s ease-in-out infinite; }
@keyframes pulse { 0%,100% { transform: scale(1); opacity: 1; } 50% { transform: scale(1.5); opacity: .6; } }
.dot-empty { width: 10px; height: 10px; background: #e2e8f0; border-radius: 50%; }

.step-info { padding-top: .35rem; }
.step-label { font-size: .9rem; font-weight: 700; color: #1e293b; margin: 0 0 .2rem; }
.step.pending-step .step-label { color: #94a3b8; font-weight: 500; }
.step-desc  { font-size: .8rem; color: #64748b; margin: 0; }
.step.pending-step .step-desc { color: #cbd5e1; }

.cancelled-box { display: flex; align-items: center; gap: 1rem; background: #fff5f5; border: 1px solid #fecaca; border-radius: 10px; padding: 1.25rem; }
.cancelled-icon { width: 36px; height: 36px; background: #fee2e2; border-radius: 50%; display: flex; align-items: center; justify-content: center; color: #b91c1c; font-weight: 700; flex-shrink: 0; }
.cancelled-title { font-size: .95rem; font-weight: 700; color: #b91c1c; margin: 0 0 .2rem; }
.cancelled-sub   { font-size: .8rem; color: #ef4444; margin: 0; }

/* Items */
.items-list { display: flex; flex-direction: column; gap: .5rem; margin-bottom: 1.5rem; }
.item-row   { display: flex; align-items: center; gap: 1rem; padding: .75rem 0; border-bottom: 1px solid #f1f5f9; }
.item-row:last-child { border-bottom: none; }
.item-thumb { font-size: 1.5rem; flex-shrink: 0; }
.item-info  { flex: 1; min-width: 0; }
.item-name  { font-size: .875rem; font-weight: 600; color: #1e293b; margin: 0 0 .15rem; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.item-sku   { font-size: .75rem; color: #94a3b8; font-family: monospace; margin: 0; }
.item-qty   { font-size: .85rem; font-weight: 700; color: #475569; background: #f1f5f9; padding: .2rem .6rem; border-radius: 6px; white-space: nowrap; }
.item-price { font-size: .85rem; color: #64748b; white-space: nowrap; }
.item-total { font-size: .9rem; font-weight: 700; color: #1e293b; white-space: nowrap; }

.totals  { background: #f8fafc; border-radius: 10px; padding: 1rem 1.25rem; display: flex; flex-direction: column; gap: .4rem; }
.tot-row { display: flex; justify-content: space-between; font-size: .875rem; color: #475569; }
.tot-row.discount { color: #10b981; }
.tot-row.grand    { font-size: 1.05rem; color: #1e293b; border-top: 1px solid #e2e8f0; margin-top: .4rem; padding-top: .6rem; }

/* Actions */
.actions { display: flex; gap: 1rem; flex-wrap: wrap; }
.btn-primary   { background: #1d4ed8; color: #fff; padding: .75rem 1.75rem; border-radius: 10px; text-decoration: none; font-weight: 700; font-size: .9rem; }
.btn-primary:hover { background: #1e40af; }
.btn-secondary { background: #f1f5f9; color: #475569; padding: .75rem 1.75rem; border-radius: 10px; text-decoration: none; font-weight: 600; font-size: .9rem; }
.btn-secondary:hover { background: #e2e8f0; }
.btn-outline   { background: none; border: 1.5px solid #3b82f6; color: #3b82f6; padding: .7rem 1.6rem; border-radius: 10px; text-decoration: none; font-weight: 600; font-size: .9rem; }
.btn-outline:hover { background: #eff6ff; }
</style>
