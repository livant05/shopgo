<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">Órdenes</h1>
        <p class="page-sub">Gestión de pedidos</p>
      </div>
    </div>

    <!-- Filtros -->
    <div class="filters">
      <div class="status-tabs">
        <button v-for="s in statuses" :key="s.key"
          class="tab" :class="{ active: filter === s.key }"
          @click="filter = s.key; load()">
          {{ s.label }}
        </button>
      </div>
      <input v-model="search" class="search-input" placeholder="Buscar por ID..." @input="load()" />
    </div>

    <!-- Tabla -->
    <div class="table-wrap">
      <div v-if="loading" class="loading">Cargando órdenes…</div>
      <table v-else class="tbl">
        <thead>
          <tr>
            <th>ID</th><th>Cliente</th><th>Items</th>
            <th>Total</th><th>Estado</th><th>Fecha</th><th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="orders.length === 0"><td colspan="7" class="empty">Sin órdenes</td></tr>
          <tr v-for="o in orders" :key="o.id" class="tbl-row" @click="open(o)">
            <td class="mono">{{ o.id.slice(0,8) }}…</td>
            <td class="td-muted">{{ o.customer_id.slice(0,8) }}…</td>
            <td>{{ o.items?.length ?? 0 }} items</td>
            <td class="td-amount">${{ fmt(o.total) }}</td>
            <td><span class="badge" :class="o.status">{{ o.status }}</span></td>
            <td class="td-muted">{{ fmtDate(o.created_at) }}</td>
            <td><button class="btn-sm" @click.stop="open(o)">Ver</button></td>
          </tr>
        </tbody>
      </table>
      <!-- Paginación -->
      <div class="pagination" v-if="page.total_pages > 1">
        <button @click="page.page--; load()" :disabled="!page.has_prev">‹ Ant</button>
        <span>{{ page.page }} / {{ page.total_pages }}</span>
        <button @click="page.page++; load()" :disabled="!page.has_next">Sig ›</button>
      </div>
    </div>

    <!-- Modal detalle -->
    <div v-if="selected" class="modal-overlay" @click.self="selected=null">
      <div class="modal">
        <div class="modal-header">
          <h3>Orden <span class="mono">{{ selected.id.slice(0,8) }}…</span></h3>
          <button class="modal-close" @click="selected=null">✕</button>
        </div>
        <div class="modal-body">
          <div class="detail-row"><span>Estado:</span>
            <span class="badge" :class="selected.status">{{ selected.status }}</span>
          </div>
          <div class="detail-row"><span>Total:</span><strong>${{ fmt(selected.total) }}</strong></div>
          <div class="detail-row"><span>Subtotal:</span><span>${{ fmt(selected.subtotal) }}</span></div>
          <div class="detail-row"><span>IVA:</span><span>${{ fmt(selected.tax) }}</span></div>
          <div class="detail-row"><span>Descuento:</span><span>${{ fmt(selected.discount) }}</span></div>
          <div class="detail-row"><span>Fecha:</span><span>{{ fmtDate(selected.created_at) }}</span></div>

          <h4 class="items-title">Productos</h4>
          <div v-for="it in selected.items" :key="it.product_id" class="order-item">
            <span>{{ it.name }}</span>
            <span class="td-muted">x{{ it.quantity }}</span>
            <span>${{ fmt(it.line_total) }}</span>
          </div>

          <div class="status-change" v-if="canAdvance(selected.status)">
            <select v-model="newStatus" class="select-input">
              <option v-for="opt in nextStatuses(selected.status)" :key="opt" :value="opt">{{ opt }}</option>
            </select>
            <button class="btn-primary" :disabled="updating" @click="updateStatus()">
              {{ updating ? 'Actualizando…' : 'Cambiar estado' }}
            </button>
          </div>
          <div v-if="statusErr" class="err-msg">{{ statusErr }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../../api/client'
import type { Order } from '../../types'

const TRANSITIONS: Record<string, string[]> = {
  pending:    ['confirmed','cancelled'],
  confirmed:  ['processing','cancelled'],
  processing: ['shipped','cancelled'],
  shipped:    ['delivered'],
  delivered:  [],
  cancelled:  [],
  refunded:   [],
}

const statuses = [
  { key: '', label: 'Todos' },
  { key: 'pending',    label: 'Pendiente' },
  { key: 'confirmed',  label: 'Confirmado' },
  { key: 'processing', label: 'En proceso' },
  { key: 'shipped',    label: 'Enviado' },
  { key: 'delivered',  label: 'Entregado' },
  { key: 'cancelled',  label: 'Cancelado' },
]

const orders   = ref<Order[]>([])
const loading  = ref(true)
const filter   = ref('')
const search   = ref('')
const selected = ref<Order | null>(null)
const newStatus = ref('')
const updating  = ref(false)
const statusErr = ref('')
const page     = ref({ page: 1, total_pages: 1, has_next: false, has_prev: false })

async function load() {
  loading.value = true
  try {
    const p = new URLSearchParams({ page: String(page.value.page), page_size: '20' })
    if (filter.value) p.set('status', filter.value)
    const { data } = await api.get(`/ops/orders?${p}`)
    orders.value = data.data ?? []
    page.value   = data
  } catch (e) { console.error(e) }
  finally { loading.value = false }
}

function open(o: Order) {
  selected.value = o
  newStatus.value = TRANSITIONS[o.status]?.[0] ?? ''
  statusErr.value = ''
}

function canAdvance(s: string) { return (TRANSITIONS[s]?.length ?? 0) > 0 }
function nextStatuses(s: string) { return TRANSITIONS[s] ?? [] }

async function updateStatus() {
  if (!selected.value || !newStatus.value) return
  updating.value = true; statusErr.value = ''
  try {
    await api.patch(`/ops/orders/${selected.value.id}/status`, { status: newStatus.value })
    selected.value.status = newStatus.value as any
    await load()
    selected.value = null
  } catch (e: any) {
    statusErr.value = e.response?.data?.message ?? 'Error actualizando estado'
  } finally { updating.value = false }
}

const fmt     = (n: number) => (n ?? 0).toLocaleString('es-MX', { minimumFractionDigits: 2 })
const fmtDate = (d: string) => new Date(d).toLocaleString('es-MX', { dateStyle:'short', timeStyle:'short' })

onMounted(load)
</script>

<style scoped>
.page        { display:flex; flex-direction:column; gap:24px; }
.page-header { display:flex; align-items:flex-start; justify-content:space-between; }
.page-title  { font-size:22px; font-weight:700; color:#eaf0f7; margin:0; }
.page-sub    { font-size:13px; color:#5a6a87; margin:4px 0 0; }

.filters      { display:flex; align-items:center; gap:12px; flex-wrap:wrap; }
.status-tabs  { display:flex; gap:4px; background:#1c2333; border-radius:8px; padding:4px; }
.tab          { background:none; border:none; color:#5a6a87; padding:6px 12px; border-radius:6px; font-size:12px; cursor:pointer; white-space:nowrap; }
.tab.active   { background:#253047; color:#38bdf8; }
.search-input { background:#1c2333; border:1px solid #2d3a52; color:#d6dfe8; padding:8px 12px; border-radius:7px; font-size:13px; width:220px; }
.search-input:focus { outline:none; border-color:#38bdf8; }

.table-wrap { background:#1c2333; border:1px solid #2d3a52; border-radius:10px; overflow:hidden; }
.loading    { padding:40px; text-align:center; color:#5a6a87; }
.tbl        { width:100%; border-collapse:collapse; font-size:13px; }
.tbl thead th { background:#253047; color:#8494ac; padding:10px 14px; text-align:left; font-weight:600; font-size:11px; text-transform:uppercase; letter-spacing:.5px; }
.tbl-row      { border-top:1px solid #253047; cursor:pointer; transition:background .15s; }
.tbl-row:hover { background:rgba(56,189,248,.04); }
.tbl-row td   { padding:11px 14px; color:#d6dfe8; }
.empty        { padding:40px; text-align:center; color:#5a6a87; }
.mono         { font-family:monospace; font-size:12px; }
.td-muted     { color:#5a6a87; }
.td-amount    { font-weight:600; color:#4ade80; }
.btn-sm       { background:#253047; border:none; color:#38bdf8; padding:5px 10px; border-radius:5px; font-size:12px; cursor:pointer; }

.badge           { display:inline-block; padding:3px 9px; border-radius:12px; font-size:11px; font-weight:600; text-transform:uppercase; letter-spacing:.5px; }
.badge.pending    { background:rgba(251,191,36,.12); color:#fbbf24; }
.badge.confirmed  { background:rgba(56,189,248,.12); color:#38bdf8; }
.badge.processing { background:rgba(167,139,250,.12); color:#a78bfa; }
.badge.shipped    { background:rgba(94,234,212,.12); color:#5eead4; }
.badge.delivered  { background:rgba(74,222,128,.12); color:#4ade80; }
.badge.cancelled  { background:rgba(248,113,113,.12); color:#f87171; }
.badge.refunded   { background:rgba(148,163,184,.12); color:#94a3b8; }

.pagination { display:flex; align-items:center; justify-content:center; gap:16px; padding:14px; border-top:1px solid #253047; }
.pagination button { background:#253047; border:none; color:#8494ac; padding:6px 12px; border-radius:6px; cursor:pointer; }
.pagination button:disabled { opacity:.4; cursor:not-allowed; }
.pagination span { font-size:13px; color:#5a6a87; }

.modal-overlay { position:fixed; inset:0; background:rgba(0,0,0,.7); display:flex; align-items:center; justify-content:center; z-index:100; }
.modal         { background:#1c2333; border:1px solid #2d3a52; border-radius:12px; width:500px; max-height:80vh; overflow-y:auto; }
.modal-header  { display:flex; align-items:center; justify-content:space-between; padding:20px 24px 0; }
.modal-header h3 { font-size:16px; color:#eaf0f7; margin:0; }
.modal-close   { background:none; border:none; color:#5a6a87; font-size:18px; cursor:pointer; }
.modal-body    { padding:20px 24px 24px; display:flex; flex-direction:column; gap:12px; }
.detail-row    { display:flex; align-items:center; justify-content:space-between; font-size:13px; color:#8494ac; }
.detail-row strong { color:#eaf0f7; }
.items-title   { font-size:13px; color:#8494ac; text-transform:uppercase; letter-spacing:.5px; margin:4px 0 0; }
.order-item    { display:flex; align-items:center; justify-content:space-between; padding:8px 12px; background:#253047; border-radius:7px; font-size:13px; color:#d6dfe8; }
.status-change { display:flex; gap:8px; margin-top:8px; }
.select-input  { flex:1; background:#0f1623; border:1px solid #2d3a52; color:#d6dfe8; padding:8px 10px; border-radius:7px; font-size:13px; }
.btn-primary   { background:#38bdf8; color:#080c14; border:none; padding:8px 16px; border-radius:7px; font-size:13px; font-weight:700; cursor:pointer; }
.btn-primary:disabled { opacity:.5; cursor:not-allowed; }
.err-msg       { color:#f87171; font-size:12px; }
</style>
