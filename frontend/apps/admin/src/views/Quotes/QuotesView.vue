<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api } from '../../api/client'

interface QuoteItem {
  product_id: string
  sku: string
  name: string
  qty: number
  unit_price: number
  subtotal: number
}

interface Quote {
  id: string
  quote_number: number
  items: QuoteItem[]
  subtotal: number
  tax_rate: number
  tax_amount: number
  total: number
  currency: string
  store_name: string
  customer_name: string
  customer_email: string
  customer_phone: string
  note: string
  status: string
  status_note: string
  status_at?: string
  created_at: string
  expires_at?: string
}

interface Page {
  data: Quote[]
  total: number
  page: number
  total_pages: number
  has_next: boolean
  has_prev: boolean
}

const STATUS_LABEL: Record<string, string> = {
  pending:  'Pendiente',
  accepted: 'Aceptada',
  rejected: 'Rechazada',
}

const quotes     = ref<Quote[]>([])
const total      = ref(0)
const page       = ref(1)
const totalPages = ref(1)
const loading    = ref(false)
const saving     = ref(false)
const search     = ref('')
const statusFilter = ref('')
const from       = ref('')
const to         = ref('')
const selected   = ref<Quote | null>(null)
const noteInput  = ref('')
const errMsg     = ref('')

let searchTimer: ReturnType<typeof setTimeout> | null = null

async function load(resetPage = false) {
  if (resetPage) page.value = 1
  loading.value = true
  try {
    const { data } = await api.get<Page>('/admin/quotes', {
      params: {
        q:         search.value        || undefined,
        status:    statusFilter.value  || undefined,
        from:      from.value          || undefined,
        to:        to.value            || undefined,
        page:      page.value,
        page_size: 20,
      },
    })
    quotes.value     = data.data
    total.value      = Number(data.total)
    totalPages.value = data.total_pages
  } catch { /* keep current */ }
  loading.value = false
}

async function setStatus(status: 'accepted' | 'rejected') {
  if (!selected.value) return
  saving.value = true
  errMsg.value = ''
  try {
    const { data } = await api.patch<Quote>(`/admin/quotes/${selected.value.id}/status`, {
      status,
      note: noteInput.value,
    })
    selected.value = data
    // update row in list
    const idx = quotes.value.findIndex(q => q.id === data.id)
    if (idx !== -1) quotes.value[idx] = data
    noteInput.value = ''
  } catch (e: any) {
    errMsg.value = e?.response?.data?.message ?? 'Error al actualizar estado.'
  } finally {
    saving.value = false
  }
}

function scheduleLoad() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => load(true), 350)
}

function fmt(v: number) {
  return v.toLocaleString('es-MX', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function fmtDate(d: string) {
  return new Date(d).toLocaleDateString('es-MX', { day: '2-digit', month: 'short', year: 'numeric' })
}

function storeFrontLink(id: string) {
  const base = import.meta.env.VITE_STOREFRONT_URL ?? 'http://localhost:5177'
  return `${base}/quote/${id}`
}

watch(search, scheduleLoad)
watch([statusFilter, from, to], () => load(true))
watch(page, () => load(false))

onMounted(() => load())
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">Cotizaciones</h1>
        <p class="page-sub">{{ total }} cotización{{ total !== 1 ? 'es' : '' }}</p>
      </div>
    </div>

    <!-- Filtros -->
    <div class="filters">
      <input
        v-model="search"
        class="search-input"
        placeholder="Buscar por cliente o correo…"
      />
      <div class="status-tabs">
        <button
          v-for="s in [['', 'Todas'], ['pending', 'Pendientes'], ['accepted', 'Aceptadas'], ['rejected', 'Rechazadas']]"
          :key="s[0]"
          class="tab"
          :class="{ active: statusFilter === s[0] }"
          @click="statusFilter = s[0]"
        >{{ s[1] }}</button>
      </div>
      <div class="date-range">
        <input v-model="from" type="date" class="date-input" title="Desde" />
        <span class="sep">—</span>
        <input v-model="to" type="date" class="date-input" title="Hasta" />
      </div>
    </div>

    <!-- Tabla -->
    <div class="table-wrap">
      <div v-if="loading" class="loading">Cargando cotizaciones…</div>
      <table v-else class="tbl">
        <thead>
          <tr>
            <th>N.°</th>
            <th>Cliente</th>
            <th>Correo</th>
            <th>Items</th>
            <th>Total</th>
            <th>Estado</th>
            <th>Fecha</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="quotes.length === 0">
            <td colspan="8" class="empty">Sin cotizaciones</td>
          </tr>
          <tr
            v-for="q in quotes"
            :key="q.id"
            class="tbl-row"
            @click="selected = q; noteInput = ''; errMsg = ''"
          >
            <td class="mono">{{ String(q.quote_number).padStart(5, '0') }}</td>
            <td class="td-main">{{ q.customer_name || '—' }}</td>
            <td class="td-muted">{{ q.customer_email || '—' }}</td>
            <td>{{ q.items?.length ?? 0 }} ítem{{ (q.items?.length ?? 0) !== 1 ? 's' : '' }}</td>
            <td class="td-amount">{{ q.currency }} {{ fmt(q.total) }}</td>
            <td><span class="badge" :class="q.status">{{ STATUS_LABEL[q.status] ?? q.status }}</span></td>
            <td class="td-muted">{{ fmtDate(q.created_at) }}</td>
            <td><button class="btn-sm" @click.stop="selected = q; noteInput = ''; errMsg = ''">Ver</button></td>
          </tr>
        </tbody>
      </table>

      <div class="pagination" v-if="totalPages > 1">
        <button @click="page--" :disabled="page <= 1">‹ Ant</button>
        <span>{{ page }} / {{ totalPages }}</span>
        <button @click="page++" :disabled="page >= totalPages">Sig ›</button>
      </div>
    </div>

    <!-- Panel detalle -->
    <Transition name="slide">
      <div v-if="selected" class="detail-overlay" @click.self="selected = null">
        <div class="detail-panel">
          <div class="detail-header">
            <div>
              <h2 class="detail-title">Cotización N.° {{ String(selected.quote_number).padStart(5, '0') }}</h2>
              <p class="detail-date">{{ fmtDate(selected.created_at) }}</p>
            </div>
            <div class="header-right">
              <span class="badge" :class="selected.status">{{ STATUS_LABEL[selected.status] ?? selected.status }}</span>
              <button class="btn-close" @click="selected = null">✕</button>
            </div>
          </div>

          <!-- Vigencia -->
          <div v-if="selected.expires_at" class="expiry-bar" :class="{ expired: new Date(selected.expires_at) < new Date() }">
            <span>{{ new Date(selected.expires_at) < new Date() ? '⚠ Vencida' : '✓ Vigente hasta' }}</span>
            <span>{{ fmtDate(selected.expires_at) }}</span>
          </div>

          <!-- Cliente -->
          <div class="detail-section">
            <p class="section-label">Cliente</p>
            <p class="detail-value">{{ selected.customer_name || 'Sin nombre' }}</p>
            <p v-if="selected.customer_email" class="detail-sub">{{ selected.customer_email }}</p>
            <p v-if="selected.customer_phone" class="detail-sub">{{ selected.customer_phone }}</p>
          </div>

          <!-- Items -->
          <div class="detail-section">
            <p class="section-label">Productos</p>
            <table class="items-tbl">
              <thead>
                <tr><th>Producto</th><th>SKU</th><th>Cant.</th><th>Precio</th><th>Subtotal</th></tr>
              </thead>
              <tbody>
                <tr v-for="it in selected.items" :key="it.product_id">
                  <td>{{ it.name }}</td>
                  <td class="mono">{{ it.sku }}</td>
                  <td>{{ it.qty }}</td>
                  <td>{{ fmt(it.unit_price) }}</td>
                  <td class="td-amount">{{ fmt(it.subtotal) }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Totales -->
          <div class="totals">
            <div class="total-row"><span>Subtotal</span><span>{{ selected.currency }} {{ fmt(selected.subtotal) }}</span></div>
            <div class="total-row"><span>Impuesto ({{ (selected.tax_rate * 100).toFixed(0) }}%)</span><span>{{ selected.currency }} {{ fmt(selected.tax_amount) }}</span></div>
            <div class="total-row grand"><span>Total</span><span>{{ selected.currency }} {{ fmt(selected.total) }}</span></div>
          </div>

          <!-- Nota del cliente -->
          <div v-if="selected.note" class="detail-section">
            <p class="section-label">Nota del cliente</p>
            <p class="detail-note">{{ selected.note }}</p>
          </div>

          <!-- Nota de estado previa -->
          <div v-if="selected.status_note" class="detail-section">
            <p class="section-label">Nota interna</p>
            <p class="detail-note note-internal">{{ selected.status_note }}</p>
          </div>

          <!-- Acciones de estado -->
          <div v-if="selected.status === 'pending'" class="detail-section">
            <p class="section-label">Cambiar estado</p>
            <textarea
              v-model="noteInput"
              class="note-input"
              rows="2"
              placeholder="Nota interna opcional…"
            />
            <p v-if="errMsg" class="err-msg">{{ errMsg }}</p>
            <div class="action-btns">
              <button
                class="btn-accept"
                :disabled="saving"
                @click="setStatus('accepted')"
              >✓ Aceptar</button>
              <button
                class="btn-reject"
                :disabled="saving"
                @click="setStatus('rejected')"
              >✕ Rechazar</button>
            </div>
          </div>

          <!-- Acciones -->
          <div class="detail-actions">
            <a :href="storeFrontLink(selected.id)" target="_blank" class="btn-view">
              🔗 Ver cotización
            </a>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.page { max-width: 1100px; }
.page-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 1.5rem; }
.page-title { font-size: 1.5rem; font-weight: 700; color: #e2e8f0; margin: 0; }
.page-sub   { font-size: .875rem; color: #5a7298; margin: .2rem 0 0; }

/* Filtros */
.filters { display: flex; gap: 1rem; margin-bottom: 1.25rem; flex-wrap: wrap; align-items: center; }
.search-input {
  flex: 1; min-width: 200px; padding: .5rem .85rem;
  background: #0f1623; border: 1px solid #253047; border-radius: 8px;
  color: #d6dfe8; font-size: .875rem; outline: none;
}
.search-input:focus { border-color: #38bdf8; }
.status-tabs { display: flex; gap: .35rem; }
.tab {
  padding: .35rem .8rem; background: #0a0f1a; border: 1px solid #253047;
  border-radius: 7px; color: #5a7298; font-size: .8rem; cursor: pointer;
}
.tab:hover { border-color: #38bdf8; color: #38bdf8; }
.tab.active { background: rgba(56,189,248,.1); border-color: #38bdf8; color: #38bdf8; font-weight: 600; }
.date-range { display: flex; align-items: center; gap: .5rem; }
.date-input {
  padding: .5rem .65rem; background: #0f1623; border: 1px solid #253047;
  border-radius: 8px; color: #d6dfe8; font-size: .875rem; outline: none;
}
.date-input:focus { border-color: #38bdf8; }
.sep { color: #5a7298; }

/* Tabla */
.table-wrap { background: #0f1623; border: 1px solid #253047; border-radius: 12px; overflow: hidden; }
.loading { padding: 2.5rem; text-align: center; color: #5a7298; font-size: .9rem; }
.tbl { width: 100%; border-collapse: collapse; }
.tbl thead tr { background: #0a0f1a; }
.tbl th { padding: .75rem 1rem; text-align: left; font-size: .75rem; font-weight: 600; color: #5a7298; text-transform: uppercase; letter-spacing: .07em; border-bottom: 1px solid #1a2540; }
.tbl td { padding: .8rem 1rem; font-size: .875rem; color: #a8b8cc; border-bottom: 1px solid #1a2540; }
.tbl-row { cursor: pointer; transition: background .15s; }
.tbl-row:hover { background: rgba(56,189,248,.05); }
.tbl-row:last-child td { border-bottom: none; }
.empty { text-align: center; color: #3d5070; padding: 2.5rem !important; }
.mono { font-family: monospace; font-size: .82rem; }
.td-main { color: #d6dfe8; font-weight: 500; }
.td-muted { color: #5a7298; }
.td-amount { font-weight: 700; color: #38bdf8; }
.btn-sm { padding: .3rem .75rem; background: rgba(56,189,248,.1); border: 1px solid rgba(56,189,248,.2); border-radius: 6px; color: #38bdf8; font-size: .78rem; cursor: pointer; }
.btn-sm:hover { background: rgba(56,189,248,.2); }

/* Badges */
.badge { display: inline-block; padding: .2rem .6rem; border-radius: 999px; font-size: .75rem; font-weight: 600; }
.badge.pending  { background: rgba(251,191,36,.12); color: #fbbf24; }
.badge.accepted { background: rgba(16,185,129,.12); color: #10b981; }
.badge.rejected { background: rgba(239,68,68,.12);  color: #ef4444; }

/* Paginación */
.pagination { display: flex; align-items: center; justify-content: center; gap: 1rem; padding: 1rem; border-top: 1px solid #1a2540; }
.pagination button { padding: .35rem .85rem; background: #0a0f1a; border: 1px solid #253047; border-radius: 7px; color: #38bdf8; cursor: pointer; font-size: .85rem; }
.pagination button:disabled { color: #3d5070; cursor: not-allowed; }
.pagination span { font-size: .85rem; color: #5a7298; }

/* Detail panel */
.detail-overlay { position: fixed; inset: 0; background: rgba(0,0,0,.6); z-index: 200; display: flex; justify-content: flex-end; }
.detail-panel {
  width: 520px; max-width: 100vw; height: 100%; background: #0f1623;
  border-left: 1px solid #253047; overflow-y: auto;
  display: flex; flex-direction: column; gap: 1.25rem; padding: 1.5rem;
}
.detail-header { display: flex; justify-content: space-between; align-items: flex-start; }
.header-right  { display: flex; align-items: center; gap: .65rem; }
.detail-title  { font-size: 1.1rem; font-weight: 700; color: #e2e8f0; margin: 0; }
.detail-date   { font-size: .8rem; color: #5a7298; margin: .2rem 0 0; }
.btn-close { background: none; border: none; color: #5a7298; font-size: 1.1rem; cursor: pointer; padding: .25rem; }
.btn-close:hover { color: #e2e8f0; }

.expiry-bar {
  display: flex; justify-content: space-between; align-items: center;
  background: rgba(16,185,129,.08); border: 1px solid rgba(16,185,129,.2);
  border-radius: 8px; padding: .5rem .85rem; font-size: .82rem; color: #10b981;
}
.expiry-bar.expired {
  background: rgba(239,68,68,.08); border-color: rgba(239,68,68,.2); color: #ef4444;
}

.detail-section { border-top: 1px solid #1a2540; padding-top: 1rem; }
.section-label { font-size: .7rem; font-weight: 700; text-transform: uppercase; letter-spacing: .08em; color: #5a7298; margin: 0 0 .5rem; }
.detail-value  { font-size: .95rem; font-weight: 600; color: #d6dfe8; margin: 0; }
.detail-sub    { font-size: .82rem; color: #5a7298; margin: .15rem 0 0; }
.detail-note   { font-size: .875rem; color: #a8b8cc; background: #0a0f1a; padding: .65rem .85rem; border-radius: 8px; margin: 0; }
.note-internal { border-left: 2px solid #253047; }

.note-input {
  width: 100%; padding: .55rem .75rem; background: #0a0f1a; border: 1px solid #253047;
  border-radius: 8px; color: #d6dfe8; font-size: .875rem; outline: none;
  resize: vertical; box-sizing: border-box; margin-bottom: .65rem; font-family: inherit;
}
.note-input:focus { border-color: #38bdf8; }
.err-msg { font-size: .82rem; color: #ef4444; margin-bottom: .5rem; }
.action-btns { display: flex; gap: .65rem; }
.btn-accept {
  flex: 1; padding: .55rem; background: rgba(16,185,129,.12); border: 1px solid rgba(16,185,129,.3);
  border-radius: 8px; color: #10b981; font-weight: 700; font-size: .875rem; cursor: pointer;
}
.btn-accept:hover:not(:disabled) { background: rgba(16,185,129,.22); }
.btn-reject {
  flex: 1; padding: .55rem; background: rgba(239,68,68,.1); border: 1px solid rgba(239,68,68,.25);
  border-radius: 8px; color: #ef4444; font-weight: 700; font-size: .875rem; cursor: pointer;
}
.btn-reject:hover:not(:disabled) { background: rgba(239,68,68,.2); }
.btn-accept:disabled, .btn-reject:disabled { opacity: .5; cursor: not-allowed; }

.items-tbl { width: 100%; border-collapse: collapse; font-size: .82rem; }
.items-tbl th { color: #5a7298; text-align: left; padding: .4rem .5rem; border-bottom: 1px solid #1a2540; font-size: .72rem; text-transform: uppercase; }
.items-tbl td { padding: .5rem .5rem; color: #a8b8cc; border-bottom: 1px solid #1a2540; }
.items-tbl tr:last-child td { border-bottom: none; }

.totals { background: #0a0f1a; border-radius: 8px; padding: .85rem 1rem; display: flex; flex-direction: column; gap: .35rem; }
.total-row { display: flex; justify-content: space-between; font-size: .875rem; color: #5a7298; }
.total-row.grand { font-weight: 700; font-size: 1rem; color: #38bdf8; border-top: 1px solid #253047; padding-top: .5rem; margin-top: .1rem; }

.detail-actions { border-top: 1px solid #1a2540; padding-top: 1rem; display: flex; gap: .75rem; }
.btn-view { display: inline-flex; align-items: center; gap: .4rem; padding: .55rem 1rem; background: rgba(56,189,248,.1); border: 1px solid rgba(56,189,248,.25); border-radius: 8px; color: #38bdf8; font-size: .875rem; font-weight: 600; text-decoration: none; }
.btn-view:hover { background: rgba(56,189,248,.18); }

/* Transition */
.slide-enter-active, .slide-leave-active { transition: transform .25s ease; }
.slide-enter-from, .slide-leave-to { transform: translateX(100%); }
</style>
