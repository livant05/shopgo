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
          class="tab" :class="{ active: filter === s.key && refundFilter === '' }"
          @click="filter = s.key; refundFilter = ''; load()">
          {{ s.label }}
        </button>
        <button class="tab tab-refund" :class="{ active: refundFilter === 'requested' }"
          @click="filter = ''; refundFilter = 'requested'; load()">
          ↩ Reembolsos pendientes
          <span v-if="pendingRefunds > 0" class="refund-dot">{{ pendingRefunds }}</span>
        </button>
      </div>
      <input v-model="search" class="search-input" placeholder="Buscar por ID…" @input="load()" />
    </div>

    <!-- Tabla -->
    <div class="table-wrap">
      <div v-if="loading" class="loading">Cargando órdenes…</div>
      <table v-else class="tbl">
        <thead>
          <tr>
            <th>ID</th><th>Cliente</th><th>Items</th>
            <th>Total</th><th>Estado</th><th>Devolución</th><th>Fecha</th><th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="orders.length === 0"><td colspan="8" class="empty">Sin órdenes</td></tr>
          <tr v-for="o in orders" :key="o.id" class="tbl-row"
            :class="{ 'row-refund': o.refund_status === 'requested' }"
            @click="open(o)">
            <td class="mono">{{ o.id.slice(0,8) }}…</td>
            <td class="td-muted">{{ o.customer_id.slice(0,8) }}…</td>
            <td>{{ o.items?.length ?? 0 }} items</td>
            <td class="td-amount">${{ fmt(o.total) }}</td>
            <td><span class="badge" :class="o.status">{{ STATUS_LABEL[o.status] ?? o.status }}</span></td>
            <td>
              <span v-if="o.refund_status !== 'none'" class="badge" :class="'refund-'+o.refund_status">
                {{ REFUND_LABEL[o.refund_status] ?? o.refund_status }}
              </span>
              <span v-else class="td-muted">—</span>
            </td>
            <td class="td-muted">{{ fmtDate(o.created_at) }}</td>
            <td><button class="btn-sm" @click.stop="open(o)">Ver</button></td>
          </tr>
        </tbody>
      </table>
      <div class="pagination" v-if="page.total_pages > 1">
        <button @click="page.page--; load()" :disabled="!page.has_prev">‹ Ant</button>
        <span>{{ page.page }} / {{ page.total_pages }}</span>
        <button @click="page.page++; load()" :disabled="!page.has_next">Sig ›</button>
      </div>
    </div>

    <!-- Modal detalle -->
    <Teleport to="body">
      <div v-if="selected" class="modal-overlay" @click.self="selected=null">
        <div class="modal">
          <div class="modal-header">
            <h3>Orden <span class="mono">{{ selected.id.slice(0,8) }}…</span></h3>
            <button class="modal-close" @click="selected=null">✕</button>
          </div>
          <div class="modal-body">
            <!-- Status row -->
            <div class="detail-row">
              <span>Estado:</span>
              <span class="badge" :class="selected.status">{{ STATUS_LABEL[selected.status] ?? selected.status }}</span>
            </div>
            <div class="detail-row"><span>Total:</span><strong>${{ fmt(selected.total) }}</strong></div>
            <div class="detail-row"><span>Subtotal:</span><span>${{ fmt(selected.subtotal) }}</span></div>
            <div class="detail-row"><span>IVA:</span><span>${{ fmt(selected.tax) }}</span></div>
            <div class="detail-row" v-if="selected.discount > 0"><span>Descuento:</span><span class="green">${{ fmt(selected.discount) }}</span></div>
            <div class="detail-row"><span>Fecha:</span><span>{{ fmtDate(selected.created_at) }}</span></div>

            <h4 class="items-title">Productos</h4>
            <div v-for="it in selected.items" :key="it.product_id" class="order-item">
              <span>{{ it.name }}</span>
              <span class="td-muted">×{{ it.quantity }}</span>
              <span>${{ fmt(it.line_total) }}</span>
            </div>

            <!-- Cambiar estado -->
            <div class="status-change" v-if="canAdvance(selected.status)">
              <select v-model="newStatus" class="select-input">
                <option v-for="opt in nextStatuses(selected.status)" :key="opt" :value="opt">
                  {{ STATUS_LABEL[opt] ?? opt }}
                </option>
              </select>
              <button class="btn-primary" :disabled="updating" @click="updateStatus()">
                {{ updating ? 'Actualizando…' : 'Cambiar estado' }}
              </button>
            </div>
            <div v-if="statusErr" class="err-msg">{{ statusErr }}</div>

            <!-- Sección devolución -->
            <template v-if="selected.refund_status !== 'none'">
              <div class="refund-section">
                <div class="refund-header">
                  <span class="refund-title">Solicitud de devolución</span>
                  <span class="badge" :class="'refund-'+selected.refund_status">
                    {{ REFUND_LABEL[selected.refund_status] }}
                  </span>
                </div>

                <div v-if="selected.refund_reason" class="refund-reason">
                  "{{ selected.refund_reason }}"
                </div>

                <div v-if="selected.refunded_at" class="refund-date">
                  Procesado: {{ fmtDate(selected.refunded_at) }}
                </div>

                <div v-if="selected.refund_status === 'requested'" class="refund-actions">
                  <p class="refund-info">
                    Al aprobar se restaurará el inventario de los {{ selected.items.length }} producto(s).
                  </p>
                  <div class="refund-buttons">
                    <button class="btn-danger" :disabled="refundLoading" @click="processRefund('reject')">
                      {{ refundLoading === 'reject' ? 'Rechazando…' : '✕ Rechazar' }}
                    </button>
                    <button class="btn-success" :disabled="refundLoading" @click="processRefund('approve')">
                      {{ refundLoading === 'approve' ? 'Aprobando…' : '✓ Aprobar y reembolsar' }}
                    </button>
                  </div>
                  <div v-if="refundErr" class="err-msg">{{ refundErr }}</div>
                </div>
              </div>
            </template>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../../api/client'
import type { Order } from '../../types'

const STATUS_LABEL: Record<string, string> = {
  pending:    'Pendiente',
  confirmed:  'Confirmado',
  processing: 'En proceso',
  shipped:    'Enviado',
  delivered:  'Entregado',
  cancelled:  'Cancelado',
  refunded:   'Reembolsado',
}

const REFUND_LABEL: Record<string, string> = {
  none:      '—',
  requested: 'Solicitado',
  approved:  'Aprobado',
  rejected:  'Rechazado',
}

const TRANSITIONS: Record<string, string[]> = {
  pending:    ['confirmed', 'cancelled'],
  confirmed:  ['processing', 'cancelled'],
  processing: ['shipped'],
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
  { key: 'refunded',   label: 'Reembolsado' },
]

const orders        = ref<Order[]>([])
const loading       = ref(true)
const filter        = ref('')
const refundFilter  = ref('')
const search        = ref('')
const selected      = ref<Order | null>(null)
const newStatus     = ref('')
const updating      = ref(false)
const statusErr     = ref('')
const refundLoading = ref<string | false>(false)
const refundErr     = ref('')
const pendingRefunds = ref(0)
const page = ref({ page: 1, total_pages: 1, has_next: false, has_prev: false })

async function load() {
  loading.value = true
  try {
    const p = new URLSearchParams({ page: String(page.value.page), page_size: '20' })
    if (filter.value)       p.set('status',        filter.value)
    if (refundFilter.value) p.set('refund_status',  refundFilter.value)
    const { data } = await api.get(`/admin/orders?${p}`)
    orders.value = data.data ?? []
    page.value   = data
  } catch (e) { console.error(e) }
  finally { loading.value = false }
}

async function loadPendingRefunds() {
  try {
    const { data } = await api.get('/admin/orders?refund_status=requested&page_size=1')
    pendingRefunds.value = data.total ?? 0
  } catch {}
}

function open(o: Order) {
  selected.value  = { ...o }
  newStatus.value = TRANSITIONS[o.status]?.[0] ?? ''
  statusErr.value = ''
  refundErr.value = ''
  refundLoading.value = false
}

function canAdvance(s: string) { return (TRANSITIONS[s]?.length ?? 0) > 0 }
function nextStatuses(s: string) { return TRANSITIONS[s] ?? [] }

async function updateStatus() {
  if (!selected.value || !newStatus.value) return
  updating.value = true; statusErr.value = ''
  try {
    await api.patch(`/admin/orders/${selected.value.id}/status`, { status: newStatus.value })
    selected.value.status = newStatus.value as any
    await load()
    selected.value = null
  } catch (e: any) {
    statusErr.value = e.response?.data?.message ?? 'Error actualizando estado'
  } finally { updating.value = false }
}

async function processRefund(action: 'approve' | 'reject') {
  if (!selected.value) return
  refundLoading.value = action; refundErr.value = ''
  try {
    await api.put(`/admin/orders/${selected.value.id}/refund`, { action })
    if (action === 'approve') {
      selected.value.refund_status = 'approved'
      selected.value.status = 'refunded' as any
    } else {
      selected.value.refund_status = 'rejected'
    }
    await Promise.all([load(), loadPendingRefunds()])
  } catch (e: any) {
    refundErr.value = e.response?.data?.message ?? 'Error procesando devolución'
  } finally { refundLoading.value = false }
}

const fmt     = (n: number) => (n ?? 0).toLocaleString('es-MX', { minimumFractionDigits: 2 })
const fmtDate = (d: string) => d ? new Date(d).toLocaleString('es-MX', { dateStyle: 'short', timeStyle: 'short' }) : '—'

onMounted(() => Promise.all([load(), loadPendingRefunds()]))
</script>

<style scoped>
.page        { display:flex; flex-direction:column; gap:24px; }
.page-header { display:flex; align-items:flex-start; justify-content:space-between; }
.page-title  { font-size:22px; font-weight:700; color:#eaf0f7; margin:0; }
.page-sub    { font-size:13px; color:#5a6a87; margin:4px 0 0; }

.filters      { display:flex; align-items:center; gap:12px; flex-wrap:wrap; }
.status-tabs  { display:flex; gap:4px; background:#1c2333; border-radius:8px; padding:4px; flex-wrap:wrap; }
.tab          { background:none; border:none; color:#5a6a87; padding:6px 12px; border-radius:6px; font-size:12px; cursor:pointer; white-space:nowrap; display:flex; align-items:center; gap:5px; }
.tab.active   { background:#253047; color:#38bdf8; }
.tab-refund.active { color:#fb923c; }
.refund-dot   { background:#fb923c; color:#080c14; border-radius:10px; padding:0 6px; font-size:10px; font-weight:800; }
.search-input { background:#1c2333; border:1px solid #2d3a52; color:#d6dfe8; padding:8px 12px; border-radius:7px; font-size:13px; width:220px; }
.search-input:focus { outline:none; border-color:#38bdf8; }

.table-wrap { background:#1c2333; border:1px solid #2d3a52; border-radius:10px; overflow:hidden; }
.loading    { padding:40px; text-align:center; color:#5a6a87; }
.tbl        { width:100%; border-collapse:collapse; font-size:13px; }
.tbl thead th { background:#253047; color:#8494ac; padding:10px 14px; text-align:left; font-weight:600; font-size:11px; text-transform:uppercase; letter-spacing:.5px; }
.tbl-row      { border-top:1px solid #253047; cursor:pointer; transition:background .15s; }
.tbl-row:hover { background:rgba(56,189,248,.04); }
.tbl-row.row-refund { background:rgba(251,146,60,.04); }
.tbl-row.row-refund:hover { background:rgba(251,146,60,.08); }
.tbl-row td   { padding:11px 14px; color:#d6dfe8; }
.empty        { padding:40px; text-align:center; color:#5a6a87; }
.mono         { font-family:monospace; font-size:12px; }
.td-muted     { color:#5a6a87; }
.td-amount    { font-weight:600; color:#4ade80; }
.green        { color:#4ade80; }
.btn-sm       { background:#253047; border:none; color:#38bdf8; padding:5px 10px; border-radius:5px; font-size:12px; cursor:pointer; }

/* Status badges */
.badge           { display:inline-block; padding:3px 9px; border-radius:12px; font-size:11px; font-weight:600; text-transform:uppercase; letter-spacing:.5px; }
.badge.pending    { background:rgba(251,191,36,.12); color:#fbbf24; }
.badge.confirmed  { background:rgba(56,189,248,.12); color:#38bdf8; }
.badge.processing { background:rgba(167,139,250,.12); color:#a78bfa; }
.badge.shipped    { background:rgba(94,234,212,.12); color:#5eead4; }
.badge.delivered  { background:rgba(74,222,128,.12); color:#4ade80; }
.badge.cancelled  { background:rgba(248,113,113,.12); color:#f87171; }
.badge.refunded   { background:rgba(148,163,184,.12); color:#94a3b8; }

/* Refund status badges */
.badge.refund-requested { background:rgba(251,146,60,.15); color:#fb923c; }
.badge.refund-approved  { background:rgba(74,222,128,.12); color:#4ade80; }
.badge.refund-rejected  { background:rgba(248,113,113,.12); color:#f87171; }

.pagination { display:flex; align-items:center; justify-content:center; gap:16px; padding:14px; border-top:1px solid #253047; }
.pagination button { background:#253047; border:none; color:#8494ac; padding:6px 12px; border-radius:6px; cursor:pointer; }
.pagination button:disabled { opacity:.4; cursor:not-allowed; }
.pagination span { font-size:13px; color:#5a6a87; }

.modal-overlay { position:fixed; inset:0; background:rgba(0,0,0,.7); display:flex; align-items:center; justify-content:center; z-index:500; }
.modal         { background:#1c2333; border:1px solid #2d3a52; border-radius:12px; width:520px; max-height:85vh; overflow-y:auto; }
.modal-header  { display:flex; align-items:center; justify-content:space-between; padding:20px 24px 0; }
.modal-header h3 { font-size:16px; color:#eaf0f7; margin:0; }
.modal-close   { background:none; border:none; color:#5a6a87; font-size:18px; cursor:pointer; }
.modal-body    { padding:20px 24px 24px; display:flex; flex-direction:column; gap:12px; }
.detail-row    { display:flex; align-items:center; justify-content:space-between; font-size:13px; color:#8494ac; }
.detail-row strong { color:#eaf0f7; }
.items-title   { font-size:11px; color:#8494ac; text-transform:uppercase; letter-spacing:.5px; margin:4px 0 0; }
.order-item    { display:flex; align-items:center; justify-content:space-between; padding:8px 12px; background:#253047; border-radius:7px; font-size:13px; color:#d6dfe8; }
.status-change { display:flex; gap:8px; margin-top:4px; }
.select-input  { flex:1; background:#0f1623; border:1px solid #2d3a52; color:#d6dfe8; padding:8px 10px; border-radius:7px; font-size:13px; }
.btn-primary   { background:#38bdf8; color:#080c14; border:none; padding:8px 16px; border-radius:7px; font-size:13px; font-weight:700; cursor:pointer; }
.btn-primary:disabled { opacity:.5; cursor:not-allowed; }
.err-msg       { color:#f87171; font-size:12px; }

/* Refund section */
.refund-section { background:#131d2e; border:1px solid #1e2d45; border-radius:10px; padding:16px; display:flex; flex-direction:column; gap:10px; margin-top:4px; }
.refund-header  { display:flex; align-items:center; justify-content:space-between; }
.refund-title   { font-size:12px; font-weight:700; color:#8494ac; text-transform:uppercase; letter-spacing:.5px; }
.refund-reason  { font-size:13px; color:#d6dfe8; font-style:italic; background:#0f1623; padding:10px 12px; border-radius:7px; border-left:3px solid #2d3a52; }
.refund-date    { font-size:11px; color:#5a7298; }
.refund-info    { font-size:12px; color:#5a7298; margin:0; }
.refund-actions { display:flex; flex-direction:column; gap:8px; }
.refund-buttons { display:flex; gap:8px; }
.btn-danger  { flex:1; background:rgba(248,113,113,.1); border:1px solid rgba(248,113,113,.3); color:#f87171; padding:9px 16px; border-radius:7px; font-size:13px; font-weight:600; cursor:pointer; transition:background .15s; }
.btn-danger:hover  { background:rgba(248,113,113,.2); }
.btn-danger:disabled  { opacity:.5; cursor:not-allowed; }
.btn-success { flex:1; background:#38bdf8; color:#080c14; border:none; padding:9px 16px; border-radius:7px; font-size:13px; font-weight:700; cursor:pointer; }
.btn-success:disabled { opacity:.5; cursor:not-allowed; }
</style>
