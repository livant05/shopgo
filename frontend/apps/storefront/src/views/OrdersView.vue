<template>
  <div class="container">
    <div class="page-header">
      <h1 class="title">Mis pedidos</h1>
      <router-link to="/catalog" class="btn-shop">Seguir comprando →</router-link>
    </div>

    <div v-if="loading" class="skeleton-list">
      <div v-for="n in 4" :key="n" class="skeleton-row" />
    </div>

    <div v-else-if="orders.length === 0" class="empty">
      <p class="empty-icon">📦</p>
      <h2>Sin pedidos aún</h2>
      <p>Cuando realices una compra aparecerá aquí</p>
      <router-link to="/catalog" class="btn-cta">Ver catálogo</router-link>
    </div>

    <div v-else class="orders-list">
      <div v-for="o in orders" :key="o.id" class="order-card" @click="toggle(o.id)">
        <div class="order-head">
          <div class="order-meta">
            <span class="order-id"># {{ o.id.slice(0, 8).toUpperCase() }}</span>
            <span class="order-date">{{ fmtDate(o.created_at) }}</span>
          </div>
          <div class="order-right">
            <span class="status-badge" :class="o.status">{{ LABELS[o.status] ?? o.status }}</span>
            <span class="order-total">{{ fmt(o.total) }}</span>
            <span class="chevron" :class="{ open: expanded === o.id }">›</span>
          </div>
        </div>

        <!-- Expanded items -->
        <div v-if="expanded === o.id" class="order-body" @click.stop>
          <div v-for="item in o.items" :key="item.product_id" class="order-item">
            <div class="item-left">
              <span class="item-qty">×{{ item.quantity }}</span>
              <span class="item-name">{{ item.name }}</span>
              <span class="item-sku">{{ item.sku }}</span>
            </div>
            <span class="item-total">{{ fmt(item.line_total) }}</span>
          </div>
          <div class="order-sums">
            <div class="sum-row"><span>Subtotal</span><span>{{ fmt(o.subtotal) }}</span></div>
            <div class="sum-row" v-if="o.discount > 0"><span>Descuento</span><span class="green">−{{ fmt(o.discount) }}</span></div>
            <div class="sum-row" v-if="o.tax > 0"><span>IVA</span><span>{{ fmt(o.tax) }}</span></div>
            <div class="sum-row total"><span>Total</span><strong>{{ fmt(o.total) }}</strong></div>
          </div>
        </div>
      </div>

      <!-- Pagination -->
      <div v-if="totalPages > 1" class="pagination">
        <button :disabled="page === 1" @click="page--; load()">← Anterior</button>
        <span>{{ page }} / {{ totalPages }}</span>
        <button :disabled="page >= totalPages" @click="page++; load()">Siguiente →</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '../api/client'
import type { Order } from '../types'

const orders    = ref<Order[]>([])
const loading   = ref(true)
const total     = ref(0)
const page      = ref(1)
const pageSize  = 10
const expanded  = ref<string | null>(null)

const totalPages = computed(() => Math.ceil(total.value / pageSize))

const LABELS: Record<string, string> = {
  pending:    'Pendiente',
  confirmed:  'Confirmado',
  processing: 'En proceso',
  shipped:    'Enviado',
  delivered:  'Entregado',
  cancelled:  'Cancelado',
}

async function load() {
  loading.value = true
  try {
    const { data } = await api.get('/orders', { params: { page: page.value, page_size: pageSize } })
    orders.value = data.data ?? []
    total.value  = data.total ?? 0
  } catch {}
  loading.value = false
}

function toggle(id: string) {
  expanded.value = expanded.value === id ? null : id
}

const fmt     = (n: number) => (n ?? 0).toLocaleString('es-MX', { style: 'currency', currency: 'MXN' })
const fmtDate = (s: string) => new Date(s).toLocaleDateString('es-MX', { year: 'numeric', month: 'short', day: 'numeric' })

onMounted(load)
</script>

<style scoped>
.container  { max-width: 900px; margin: 0 auto; padding: 2rem 1.5rem; }
.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 2rem; flex-wrap: wrap; gap: 1rem; }
.title      { font-size: 1.75rem; font-weight: 800; color: #1e293b; }
.btn-shop   { background: #eff6ff; color: #1d4ed8; padding: .5rem 1.25rem; border-radius: 8px; text-decoration: none; font-weight: 600; font-size: .875rem; }
.btn-shop:hover { background: #dbeafe; }

/* skeleton */
.skeleton-list { display: flex; flex-direction: column; gap: .75rem; }
.skeleton-row  { height: 72px; border-radius: 12px; background: linear-gradient(90deg,#e2e8f0 25%,#f1f5f9 50%,#e2e8f0 75%); background-size:200% 100%; animation:shimmer 1.4s infinite; }
@keyframes shimmer { to { background-position: -200% 0; } }

/* empty */
.empty { text-align: center; padding: 5rem 2rem; }
.empty-icon { font-size: 3.5rem; margin-bottom: .75rem; }
.empty h2   { font-size: 1.3rem; font-weight: 700; color: #1e293b; margin-bottom: .4rem; }
.empty p    { color: #94a3b8; margin-bottom: 1.5rem; }
.btn-cta    { display: inline-block; background: #1d4ed8; color: #fff; padding: .7rem 1.75rem; border-radius: 10px; text-decoration: none; font-weight: 700; }
.btn-cta:hover { background: #1e40af; }

/* orders */
.orders-list { display: flex; flex-direction: column; gap: .75rem; }
.order-card  { background: #fff; border-radius: 14px; box-shadow: 0 1px 4px rgba(0,0,0,.07); overflow: hidden; cursor: pointer; transition: box-shadow .15s; }
.order-card:hover { box-shadow: 0 4px 14px rgba(0,0,0,.12); }
.order-head  { display: flex; align-items: center; justify-content: space-between; padding: 1.1rem 1.4rem; gap: 1rem; }
.order-meta  { display: flex; flex-direction: column; gap: .25rem; }
.order-id    { font-family: monospace; font-weight: 700; font-size: .95rem; color: #1e293b; letter-spacing: .06em; }
.order-date  { font-size: .78rem; color: #94a3b8; }
.order-right { display: flex; align-items: center; gap: .9rem; }
.order-total { font-weight: 700; font-size: 1rem; color: #1e293b; }
.chevron     { color: #94a3b8; font-size: 1.2rem; transition: transform .2s; display: inline-block; }
.chevron.open { transform: rotate(90deg); }

/* status badges */
.status-badge { padding: .25rem .75rem; border-radius: 999px; font-size: .75rem; font-weight: 700; }
.status-badge.pending    { background: #fef9c3; color: #854d0e; }
.status-badge.confirmed  { background: #dbeafe; color: #1d4ed8; }
.status-badge.processing { background: #ede9fe; color: #6d28d9; }
.status-badge.shipped    { background: #e0f2fe; color: #0369a1; }
.status-badge.delivered  { background: #dcfce7; color: #15803d; }
.status-badge.cancelled  { background: #fee2e2; color: #b91c1c; }

/* expanded body */
.order-body  { border-top: 1px solid #f1f5f9; padding: 1rem 1.4rem 1.4rem; }
.order-item  { display: flex; align-items: center; justify-content: space-between; padding: .55rem 0; border-bottom: 1px solid #f1f5f9; gap: .75rem; }
.order-item:last-child { border-bottom: none; }
.item-left   { display: flex; align-items: center; gap: .75rem; flex: 1; min-width: 0; }
.item-qty    { background: #f1f5f9; color: #475569; border-radius: 6px; padding: .15rem .5rem; font-size: .8rem; font-weight: 700; flex-shrink: 0; }
.item-name   { font-weight: 600; font-size: .875rem; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.item-sku    { font-size: .75rem; color: #94a3b8; font-family: monospace; flex-shrink: 0; }
.item-total  { font-weight: 700; font-size: .9rem; white-space: nowrap; }

.order-sums  { margin-top: 1rem; background: #f8fafc; border-radius: 10px; padding: 1rem; }
.sum-row     { display: flex; justify-content: space-between; font-size: .875rem; color: #475569; padding: .3rem 0; }
.sum-row.total { font-size: 1rem; color: #1e293b; border-top: 1px solid #e2e8f0; margin-top: .3rem; padding-top: .6rem; }
.green       { color: #10b981; }

/* pagination */
.pagination  { display: flex; align-items: center; justify-content: center; gap: 1.5rem; margin-top: 2rem; }
.pagination button { padding: .5rem 1.25rem; border: 1px solid #e2e8f0; border-radius: 8px; background: #fff; cursor: pointer; font-weight: 600; color: #3b82f6; }
.pagination button:disabled { color: #cbd5e1; cursor: not-allowed; }
.pagination span { font-size: .875rem; color: #64748b; }
</style>
