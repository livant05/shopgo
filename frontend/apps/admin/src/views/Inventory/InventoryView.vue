<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">Inventario</h1>
        <p class="page-sub">{{ items.length }} registros</p>
      </div>
      <div class="header-actions">
        <button class="btn-ghost"   @click="openHistory()">📋 Historial</button>
        <button class="btn-outline" @click="openTransfer()">↔ Transferir</button>
        <button class="btn-primary" @click="openAdjust()">± Ajustar</button>
      </div>
    </div>

    <!-- Filtro por sucursal -->
    <div class="filters">
      <select v-model="branchFilter" @change="load()" class="select-input">
        <option value="">Todas las sucursales</option>
        <option v-for="b in branches" :key="b.id" :value="b.id">{{ b.name }}</option>
      </select>
      <input v-model="search" class="search-input" placeholder="Buscar producto…" />
    </div>

    <!-- Tabla -->
    <div class="table-wrap">
      <div v-if="loading" class="loading">Cargando inventario…</div>
      <table v-else class="tbl">
        <thead>
          <tr>
            <th>SKU</th><th>Producto</th><th>Sucursal</th>
            <th>Cantidad</th><th>Reservado</th><th>Disponible</th>
            <th>Punto reorden</th><th>Estado</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="filtered.length === 0"><td colspan="8" class="empty">Sin registros</td></tr>
          <tr v-for="inv in filtered" :key="inv.product_id + inv.branch_id" class="tbl-row"
            :class="{ 'row-low': available(inv) <= inv.reorder_point }">
            <td class="mono">{{ inv.product_id?.slice(0,8) }}…</td>
            <td class="td-name">{{ inv.product_id }}</td>
            <td class="td-muted">{{ branchName(inv.branch_id) }}</td>
            <td>{{ inv.quantity }}</td>
            <td class="td-muted">{{ inv.reserved_qty }}</td>
            <td :class="available(inv) === 0 ? 'stock-zero' : available(inv) <= inv.reorder_point ? 'stock-low' : 'stock-ok'">
              {{ available(inv) }}
            </td>
            <td class="td-muted">{{ inv.reorder_point }}</td>
            <td>
              <span class="badge" :class="available(inv) <= inv.reorder_point ? 'badge-warn' : 'badge-ok'">
                {{ available(inv) <= inv.reorder_point ? 'Stock bajo' : 'OK' }}
              </span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Panel Historial -->
    <Teleport to="body">
      <div v-if="historyOpen" class="drawer-backdrop" @click.self="historyOpen=false">
        <div class="drawer">
          <div class="drawer-header">
            <h3>📋 Historial de Movimientos</h3>
            <button class="modal-close" @click="historyOpen=false">✕</button>
          </div>

          <div class="hist-filters">
            <select v-model="hf.type" @change="loadHistory()" class="select-input">
              <option value="">Todos los tipos</option>
              <option value="adjustment">Ajuste entrada</option>
              <option value="reduction">Ajuste salida</option>
              <option value="transfer">Transferencia</option>
              <option value="sale">Venta</option>
              <option value="return">Devolución</option>
            </select>
            <select v-model="hf.branch_id" @change="loadHistory()" class="select-input">
              <option value="">Todas las sucursales</option>
              <option v-for="b in branches" :key="b.id" :value="b.id">{{ b.name }}</option>
            </select>
            <input v-model="hf.from" type="date" class="date-input" @change="loadHistory()" />
            <input v-model="hf.to"   type="date" class="date-input" @change="loadHistory()" />
          </div>

          <div v-if="histLoading" class="hist-loading">Cargando…</div>
          <div v-else-if="history.length === 0" class="hist-empty">Sin movimientos para los filtros seleccionados</div>
          <div v-else class="hist-table-wrap">
            <table class="hist-tbl">
              <thead>
                <tr>
                  <th>Fecha</th><th>Tipo</th><th>Producto</th><th>Ruta</th><th>Cant.</th><th>Razón</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="m in history" :key="m.id" class="hist-row">
                  <td class="hist-date">{{ fmtDate(m.created_at) }}</td>
                  <td><span class="mov-badge" :class="'mov-'+m.type">{{ movLabel(m.type) }}</span></td>
                  <td>
                    <div class="mov-product">{{ m.product_name || m.product_id.slice(0,8)+'…' }}</div>
                    <div class="mov-sku">{{ m.product_sku }}</div>
                  </td>
                  <td class="mov-route">
                    <template v-if="m.type === 'transfer'">
                      <span>{{ m.from_branch_name || '—' }}</span>
                      <span class="route-arrow">→</span>
                      <span>{{ m.to_branch_name || '—' }}</span>
                    </template>
                    <template v-else>{{ m.to_branch_name || m.from_branch_name || '—' }}</template>
                  </td>
                  <td :class="m.quantity >= 0 ? 'qty-pos' : 'qty-neg'">
                    {{ m.quantity >= 0 ? '+' : '' }}{{ m.quantity }}
                  </td>
                  <td class="mov-reason">
                    <div>{{ m.reason }}</div>
                    <div v-if="m.note" class="mov-note">{{ m.note }}</div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="hist-pagination" v-if="histTotal > hf.page_size">
            <button class="pag-btn" :disabled="hf.page <= 1" @click="hf.page--; loadHistory()">← Anterior</button>
            <span class="pag-info">{{ hf.page }} / {{ Math.ceil(histTotal / hf.page_size) }}</span>
            <button class="pag-btn" :disabled="hf.page * hf.page_size >= histTotal" @click="hf.page++; loadHistory()">Siguiente →</button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Modal Ajustar -->
    <Teleport to="body">
      <div v-if="adjustModal" class="modal-overlay" @click.self="adjustModal=false">
        <div class="modal">
          <div class="modal-header">
            <h3>± Ajustar Stock</h3>
            <button class="modal-close" @click="adjustModal=false">✕</button>
          </div>
          <form @submit.prevent="doAdjust()" class="modal-body">

            <!-- Product search -->
            <div class="field">
              <label>Producto *</label>
              <div class="ac-wrap" v-click-outside="() => adjProductResults = []">
                <input
                  v-model="adjProductQuery"
                  @input="searchAdjProducts"
                  class="ac-input"
                  :class="{ 'ac-input--selected': adjProduct }"
                  placeholder="Buscar por nombre o SKU…"
                  autocomplete="off"
                />
                <div v-if="adjProductResults.length" class="ac-list">
                  <div v-for="p in adjProductResults" :key="p.id" class="ac-item"
                    @mousedown.prevent="selectAdjProduct(p)">
                    <div class="ac-item-main">
                      <span class="ac-name">{{ p.name }}</span>
                      <span class="ac-sku">{{ p.sku }}</span>
                    </div>
                    <span class="ac-stock">{{ p.stock ?? '—' }} uds.</span>
                  </div>
                </div>
              </div>
              <div v-if="adjProduct" class="selected-tag">
                ✓ {{ adjProduct.name }} <span class="selected-sku">{{ adjProduct.sku }}</span>
              </div>
            </div>

            <!-- Branch select -->
            <div class="field">
              <label>Sucursal *</label>
              <select v-model="adj.branch_id" required class="sel" @change="refreshAdjStock">
                <option value="">— Seleccionar sucursal —</option>
                <option v-for="b in branches" :key="b.id" :value="b.id">{{ b.name }}</option>
              </select>
            </div>

            <!-- Stock info -->
            <div v-if="adjCurrentStock !== null" class="stock-info">
              <span class="stock-info-label">Stock actual en sucursal:</span>
              <span class="stock-info-value" :class="adjCurrentStock === 0 ? 'stock-zero' : 'stock-ok'">
                {{ adjCurrentStock }} unidades
              </span>
            </div>

            <div class="field-row">
              <div class="field">
                <label>Delta (+ entrada / − salida) *</label>
                <input v-model.number="adj.delta" type="number" required
                  :class="{ 'input-warn': adjCurrentStock !== null && adj.delta < 0 && Math.abs(adj.delta) > adjCurrentStock }" />
                <span v-if="adjCurrentStock !== null && adj.delta < 0 && Math.abs(adj.delta) > adjCurrentStock"
                  class="field-hint warn">
                  Excede stock disponible ({{ adjCurrentStock }})
                </span>
              </div>
              <div class="field">
                <label>Razón *</label>
                <input v-model="adj.reason" required placeholder="compra, merma, pérdida…" />
              </div>
            </div>

            <div class="field">
              <label>Nota</label>
              <input v-model="adj.note" placeholder="Opcional" />
            </div>

            <div v-if="adjErr" class="err-msg">{{ adjErr }}</div>
            <div class="modal-footer">
              <button type="button" class="btn-ghost" @click="adjustModal=false">Cancelar</button>
              <button type="submit" class="btn-primary" :disabled="adjLoading || !adjProduct || !adj.branch_id">
                {{ adjLoading ? 'Ajustando…' : 'Confirmar ajuste' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

    <!-- Modal Transferir -->
    <Teleport to="body">
      <div v-if="transferModal" class="modal-overlay" @click.self="transferModal=false">
        <div class="modal modal--wide">
          <div class="modal-header">
            <h3>↔ Transferir Stock</h3>
            <button class="modal-close" @click="transferModal=false">✕</button>
          </div>
          <form @submit.prevent="doTransfer()" class="modal-body">

            <!-- Product search -->
            <div class="field">
              <label>Producto *</label>
              <div class="ac-wrap" v-click-outside="() => trProductResults = []">
                <input
                  v-model="trProductQuery"
                  @input="searchTrProducts"
                  class="ac-input"
                  :class="{ 'ac-input--selected': trProduct }"
                  placeholder="Buscar por nombre o SKU…"
                  autocomplete="off"
                />
                <div v-if="trProductResults.length" class="ac-list">
                  <div v-for="p in trProductResults" :key="p.id" class="ac-item"
                    @mousedown.prevent="selectTrProduct(p)">
                    <div class="ac-item-main">
                      <span class="ac-name">{{ p.name }}</span>
                      <span class="ac-sku">{{ p.sku }}</span>
                    </div>
                    <span class="ac-stock">{{ p.stock ?? '—' }} uds.</span>
                  </div>
                </div>
              </div>
              <div v-if="trProduct" class="selected-tag">
                ✓ {{ trProduct.name }} <span class="selected-sku">{{ trProduct.sku }}</span>
              </div>
            </div>

            <!-- Branch selects -->
            <div class="field-row">
              <div class="field">
                <label>Desde sucursal *</label>
                <select v-model="tr.from_branch_id" required class="sel" @change="refreshTrStock">
                  <option value="">— Seleccionar —</option>
                  <option v-for="b in branches" :key="b.id" :value="b.id"
                    :disabled="b.id === tr.to_branch_id">{{ b.name }}</option>
                </select>
              </div>
              <div class="field">
                <label>Hacia sucursal *</label>
                <select v-model="tr.to_branch_id" required class="sel">
                  <option value="">— Seleccionar —</option>
                  <option v-for="b in branches" :key="b.id" :value="b.id"
                    :disabled="b.id === tr.from_branch_id">{{ b.name }}</option>
                </select>
              </div>
            </div>

            <!-- Stock disponible en origen -->
            <div v-if="trAvailableStock !== null" class="stock-info">
              <span class="stock-info-label">Stock en sucursal origen:</span>
              <span class="stock-info-value" :class="trAvailableStock === 0 ? 'stock-zero' : trAvailableStock < 5 ? 'stock-low' : 'stock-ok'">
                {{ trAvailableStock }} unidades disponibles
              </span>
            </div>

            <div class="field-row">
              <div class="field">
                <label>Cantidad *</label>
                <input
                  v-model.number="tr.quantity" type="number" min="1"
                  :max="trAvailableStock ?? undefined"
                  required
                  :class="{ 'input-warn': trAvailableStock !== null && tr.quantity > trAvailableStock }"
                />
                <span v-if="trAvailableStock !== null && tr.quantity > trAvailableStock"
                  class="field-hint warn">
                  Excede stock disponible ({{ trAvailableStock }})
                </span>
                <span v-else-if="trAvailableStock !== null" class="field-hint">
                  Máx. {{ trAvailableStock }}
                </span>
              </div>
              <div class="field">
                <label>Nota</label>
                <input v-model="tr.note" placeholder="Opcional" />
              </div>
            </div>

            <div v-if="trErr" class="err-msg">{{ trErr }}</div>

            <!-- Success state -->
            <div v-if="trSuccess" class="success-banner">
              ✓ Transferencia registrada.
              <button type="button" class="link-btn" @click="transferModal=false; openHistory()">Ver historial →</button>
            </div>

            <div class="modal-footer">
              <button type="button" class="btn-ghost" @click="transferModal=false">Cancelar</button>
              <button type="submit" class="btn-primary"
                :disabled="trLoading || !trProduct || !tr.from_branch_id || !tr.to_branch_id ||
                           (trAvailableStock !== null && tr.quantity > trAvailableStock)">
                {{ trLoading ? 'Transfiriendo…' : 'Confirmar transferencia' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { api } from '../../api/client'
import type { Branch } from '../../types'

// Directive to close dropdowns on outside click
const vClickOutside = {
  mounted(el: HTMLElement, binding: any) {
    el._clickOutside = (e: Event) => { if (!el.contains(e.target as Node)) binding.value(e) }
    document.addEventListener('mousedown', el._clickOutside)
  },
  unmounted(el: HTMLElement) {
    document.removeEventListener('mousedown', el._clickOutside)
  },
}

interface InvItem { product_id: string; branch_id: string; quantity: number; reserved_qty: number; reorder_point: number }

const items         = ref<InvItem[]>([])
const branches      = ref<Branch[]>([])
const loading       = ref(true)
const branchFilter  = ref('')
const search        = ref('')
const adjustModal   = ref(false)
const transferModal = ref(false)
const historyOpen   = ref(false)

// ── Adjust modal state ────────────────────────────────────────────────────────
const adj             = ref({ product_id: '', branch_id: '', delta: 0, reason: '', note: '' })
const adjLoading      = ref(false)
const adjErr          = ref('')
const adjProductQuery   = ref('')
const adjProductResults = ref<any[]>([])
const adjProduct        = ref<any>(null)
const adjCurrentStock   = ref<number | null>(null)
let adjSearchTimer: any

// ── Transfer modal state ──────────────────────────────────────────────────────
const tr              = ref({ product_id: '', from_branch_id: '', to_branch_id: '', quantity: 1, note: '' })
const trLoading       = ref(false)
const trErr           = ref('')
const trSuccess       = ref(false)
const trProductQuery    = ref('')
const trProductResults  = ref<any[]>([])
const trProduct         = ref<any>(null)
const trAvailableStock  = ref<number | null>(null)
let trSearchTimer: any

// ── History state ─────────────────────────────────────────────────────────────
interface Movement {
  id: string; product_id: string; product_name: string; product_sku: string
  from_branch_id: string; from_branch_name: string
  to_branch_id: string; to_branch_name: string
  quantity: number; type: string; reason: string; note: string
  user_id: string; created_at: string
}
const history     = ref<Movement[]>([])
const histTotal   = ref(0)
const histLoading = ref(false)
const hf = ref({ type: '', branch_id: '', from: '', to: '', page: 1, page_size: 30 })

// ── Helpers ───────────────────────────────────────────────────────────────────
const available  = (inv: InvItem) => inv.quantity - inv.reserved_qty
const branchName = (id: string) => branches.value.find(b => b.id === id)?.name ?? id?.slice(0, 8) + '…'

const filtered = computed(() =>
  items.value.filter(i => !search.value ||
    i.product_id.toLowerCase().includes(search.value.toLowerCase()))
)

// ── Data loading ──────────────────────────────────────────────────────────────
async function load() {
  loading.value = true
  try {
    const url = branchFilter.value ? `/admin/inventory?branch_id=${branchFilter.value}` : '/admin/inventory'
    const { data } = await api.get(url)
    items.value = data.data ?? []
  } catch (e) { console.error(e) }
  finally { loading.value = false }
}

async function loadBranches() {
  const { data } = await api.get('/branches')
  branches.value = data.data ?? []
}

// ── Adjust modal ──────────────────────────────────────────────────────────────
function openAdjust() {
  adj.value = { product_id: '', branch_id: '', delta: 0, reason: '', note: '' }
  adjErr.value = ''
  adjProduct.value = null
  adjProductQuery.value = ''
  adjProductResults.value = []
  adjCurrentStock.value = null
  adjustModal.value = true
}

function searchAdjProducts() {
  clearTimeout(adjSearchTimer)
  adjProduct.value = null
  adj.value.product_id = ''
  adjCurrentStock.value = null
  if (!adjProductQuery.value) { adjProductResults.value = []; return }
  adjSearchTimer = setTimeout(async () => {
    try {
      const params: any = { q: adjProductQuery.value, page_size: 6 }
      if (adj.value.branch_id) params.branch_id = adj.value.branch_id
      const { data } = await api.get('/products', { params })
      adjProductResults.value = data.data ?? []
    } catch {}
  }, 250)
}

function selectAdjProduct(p: any) {
  adjProduct.value = p
  adj.value.product_id = p.id
  adjProductQuery.value = p.name
  adjProductResults.value = []
  if (adj.value.branch_id) adjCurrentStock.value = p.stock ?? null
}

async function refreshAdjStock() {
  if (!adjProduct.value || !adj.value.branch_id) return
  try {
    const { data } = await api.get('/products', {
      params: { q: adjProduct.value.sku, branch_id: adj.value.branch_id, page_size: 5 }
    })
    const found = (data.data ?? []).find((p: any) => p.id === adjProduct.value.id)
    adjCurrentStock.value = found?.stock ?? null
  } catch {}
}

async function doAdjust() {
  adjLoading.value = true; adjErr.value = ''
  try {
    await api.patch('/admin/inventory', adj.value)
    adjustModal.value = false; await load()
  } catch (e: any) { adjErr.value = e.response?.data?.message ?? 'Error al ajustar' }
  finally { adjLoading.value = false }
}

// ── Transfer modal ────────────────────────────────────────────────────────────
function openTransfer() {
  tr.value = { product_id: '', from_branch_id: '', to_branch_id: '', quantity: 1, note: '' }
  trErr.value = ''
  trSuccess.value = false
  trProduct.value = null
  trProductQuery.value = ''
  trProductResults.value = []
  trAvailableStock.value = null
  transferModal.value = true
}

function searchTrProducts() {
  clearTimeout(trSearchTimer)
  trProduct.value = null
  tr.value.product_id = ''
  trAvailableStock.value = null
  if (!trProductQuery.value) { trProductResults.value = []; return }
  trSearchTimer = setTimeout(async () => {
    try {
      const params: any = { q: trProductQuery.value, page_size: 6 }
      if (tr.value.from_branch_id) params.branch_id = tr.value.from_branch_id
      const { data } = await api.get('/products', { params })
      trProductResults.value = data.data ?? []
    } catch {}
  }, 250)
}

function selectTrProduct(p: any) {
  trProduct.value = p
  tr.value.product_id = p.id
  trProductQuery.value = p.name
  trProductResults.value = []
  if (tr.value.from_branch_id) trAvailableStock.value = p.stock ?? null
}

async function refreshTrStock() {
  if (!trProduct.value || !tr.value.from_branch_id) return
  try {
    const { data } = await api.get('/products', {
      params: { q: trProduct.value.sku, branch_id: tr.value.from_branch_id, page_size: 5 }
    })
    const found = (data.data ?? []).find((p: any) => p.id === trProduct.value.id)
    trAvailableStock.value = found?.stock ?? null
  } catch {}
}

async function doTransfer() {
  trLoading.value = true; trErr.value = ''
  try {
    await api.post('/admin/inventory/transfer', tr.value)
    trSuccess.value = true
    await load()
  } catch (e: any) { trErr.value = e.response?.data?.message ?? 'Error al transferir' }
  finally { trLoading.value = false }
}

// ── History ───────────────────────────────────────────────────────────────────
async function loadHistory() {
  histLoading.value = true
  try {
    const p = new URLSearchParams()
    if (hf.value.type)      p.set('type',      hf.value.type)
    if (hf.value.branch_id) p.set('branch_id', hf.value.branch_id)
    if (hf.value.from)      p.set('from',      hf.value.from)
    if (hf.value.to)        p.set('to',        hf.value.to)
    p.set('page',      String(hf.value.page))
    p.set('page_size', String(hf.value.page_size))
    const { data } = await api.get(`/admin/inventory/history?${p}`)
    history.value   = data.data  ?? []
    histTotal.value = data.total ?? 0
  } catch (e) { console.error(e) }
  finally { histLoading.value = false }
}

function openHistory() {
  hf.value = { type: '', branch_id: '', from: '', to: '', page: 1, page_size: 30 }
  historyOpen.value = true
  loadHistory()
}

const movLabel = (t: string) => ({
  adjustment: 'Entrada', reduction: 'Salida', transfer: 'Transferencia',
  sale: 'Venta', return: 'Devolución',
}[t] ?? t)

const fmtDate = (s: string) => new Date(s).toLocaleString('es-MX', { dateStyle: 'short', timeStyle: 'short' })

onMounted(() => Promise.all([load(), loadBranches()]))
</script>

<style scoped>
.page        { display:flex; flex-direction:column; gap:24px; }
.page-header { display:flex; align-items:flex-start; justify-content:space-between; }
.page-title  { font-size:22px; font-weight:700; color:#eaf0f7; margin:0; }
.page-sub    { font-size:13px; color:#5a6a87; margin:4px 0 0; }
.header-actions { display:flex; gap:8px; }
.filters     { display:flex; gap:10px; flex-wrap:wrap; }
.select-input, .search-input { background:#1c2333; border:1px solid #2d3a52; color:#d6dfe8; padding:9px 12px; border-radius:7px; font-size:13px; }
.search-input { width:240px; }
.search-input:focus { outline:none; border-color:#38bdf8; }

.table-wrap { background:#1c2333; border:1px solid #2d3a52; border-radius:10px; overflow:hidden; }
.loading    { padding:40px; text-align:center; color:#5a6a87; }
.tbl        { width:100%; border-collapse:collapse; font-size:13px; }
.tbl thead th { background:#253047; color:#8494ac; padding:10px 14px; text-align:left; font-weight:600; font-size:11px; text-transform:uppercase; letter-spacing:.5px; }
.tbl-row    { border-top:1px solid #253047; transition:background .15s; }
.tbl-row:hover { background:rgba(56,189,248,.04); }
.tbl-row td { padding:10px 14px; color:#d6dfe8; }
.row-low    { background:rgba(251,146,60,.04) !important; }
.empty      { padding:40px; text-align:center; color:#5a6a87; }
.mono       { font-family:monospace; font-size:12px; color:#8494ac; }
.td-name    { font-weight:500; }
.td-muted   { color:#5a6a87; }
.stock-ok   { color:#4ade80; font-weight:600; }
.stock-low  { color:#fb923c; font-weight:600; }
.stock-zero { color:#f87171; font-weight:600; }
.badge      { display:inline-block; padding:3px 8px; border-radius:10px; font-size:11px; font-weight:600; }
.badge-ok   { background:rgba(74,222,128,.1); color:#4ade80; }
.badge-warn { background:rgba(251,146,60,.12); color:#fb923c; }

.btn-primary { background:#38bdf8; color:#080c14; border:none; padding:9px 18px; border-radius:7px; font-size:13px; font-weight:700; cursor:pointer; }
.btn-primary:disabled { opacity:.5; cursor:not-allowed; }
.btn-outline { background:none; border:1px solid #38bdf8; color:#38bdf8; padding:9px 18px; border-radius:7px; font-size:13px; font-weight:600; cursor:pointer; }
.btn-ghost   { background:none; border:1px solid #2d3a52; color:#8494ac; padding:9px 16px; border-radius:7px; font-size:13px; cursor:pointer; }

.modal-overlay { position:fixed; inset:0; background:rgba(0,0,0,.7); display:flex; align-items:center; justify-content:center; z-index:500; }
.modal         { background:#1c2333; border:1px solid #2d3a52; border-radius:12px; width:520px; max-width:95vw; }
.modal--wide   { width:580px; }
.modal-header  { display:flex; align-items:center; justify-content:space-between; padding:20px 24px 0; }
.modal-header h3 { font-size:16px; color:#eaf0f7; margin:0; }
.modal-close   { background:none; border:none; color:#5a6a87; font-size:18px; cursor:pointer; }
.modal-body    { padding:20px 24px 0; display:flex; flex-direction:column; gap:14px; }
.modal-footer  { display:flex; gap:10px; justify-content:flex-end; padding:20px 24px 24px; }
.field         { display:flex; flex-direction:column; gap:6px; flex:1; }
.field label   { font-size:11px; color:#8494ac; text-transform:uppercase; letter-spacing:.5px; }
.field input, .field .sel, .sel { background:#0f1623; border:1px solid #2d3a52; color:#d6dfe8; padding:9px 12px; border-radius:7px; font-size:13px; width:100%; box-sizing:border-box; }
.field input:focus { outline:none; border-color:#38bdf8; }
.field-row     { display:flex; gap:12px; }
.err-msg       { color:#f87171; font-size:12px; padding-bottom:4px; }

/* Autocomplete */
.ac-wrap  { position:relative; }
.ac-input { width:100%; box-sizing:border-box; background:#0f1623; border:1px solid #2d3a52; color:#d6dfe8; padding:9px 12px; border-radius:7px; font-size:13px; }
.ac-input:focus { outline:none; border-color:#38bdf8; }
.ac-input--selected { border-color:#4ade80 !important; }
.ac-list  { position:absolute; top:calc(100% + 4px); left:0; right:0; background:#1c2333; border:1px solid #2d3a52; border-radius:8px; z-index:600; box-shadow:0 8px 24px rgba(0,0,0,.5); overflow:hidden; }
.ac-item  { display:flex; align-items:center; gap:10px; padding:10px 14px; cursor:pointer; transition:background .1s; }
.ac-item:hover { background:rgba(56,189,248,.08); }
.ac-item-main { flex:1; display:flex; flex-direction:column; gap:2px; min-width:0; }
.ac-name  { font-size:13px; color:#eaf0f7; font-weight:500; white-space:nowrap; overflow:hidden; text-overflow:ellipsis; }
.ac-sku   { font-size:11px; color:#5a6a87; font-family:monospace; }
.ac-stock { font-size:12px; font-weight:700; color:#38bdf8; white-space:nowrap; flex-shrink:0; }

/* Selected product tag */
.selected-tag { display:inline-flex; align-items:center; gap:6px; background:rgba(74,222,128,.08); border:1px solid rgba(74,222,128,.2); color:#4ade80; padding:5px 10px; border-radius:6px; font-size:12px; font-weight:600; }
.selected-sku { color:#5a7298; font-family:monospace; font-size:11px; }

/* Stock info panel */
.stock-info { display:flex; align-items:center; gap:10px; background:#131d2e; border:1px solid #1e2d45; border-radius:8px; padding:10px 14px; }
.stock-info-label { font-size:12px; color:#5a7298; }
.stock-info-value { font-size:14px; font-weight:700; margin-left:auto; }

/* Quantity hints */
.field-hint { font-size:11px; color:#5a7298; }
.field-hint.warn { color:#fb923c; }
.input-warn { border-color:#fb923c !important; }

/* Success banner */
.success-banner { background:rgba(74,222,128,.08); border:1px solid rgba(74,222,128,.2); border-radius:8px; padding:12px 16px; font-size:13px; color:#4ade80; display:flex; align-items:center; gap:10px; }
.link-btn { background:none; border:none; color:#38bdf8; font-size:13px; font-weight:600; cursor:pointer; text-decoration:underline; padding:0; }

/* Drawer historial */
.drawer-backdrop { position:fixed; inset:0; background:rgba(0,0,0,.6); z-index:400; display:flex; justify-content:flex-end; }
.drawer { background:#0f1623; border-left:1px solid #2d3a52; width:820px; max-width:95vw; height:100vh; display:flex; flex-direction:column; }
.drawer-header { display:flex; align-items:center; justify-content:space-between; padding:20px 24px; border-bottom:1px solid #2d3a52; flex-shrink:0; }
.drawer-header h3 { font-size:16px; color:#eaf0f7; margin:0; }
.hist-filters { display:flex; gap:8px; flex-wrap:wrap; padding:16px 24px; border-bottom:1px solid #1a2235; flex-shrink:0; }
.date-input   { background:#1c2333; border:1px solid #2d3a52; color:#d6dfe8; padding:8px 10px; border-radius:7px; font-size:12px; }
.hist-loading, .hist-empty { padding:40px; text-align:center; color:#5a6a87; font-size:13px; }
.hist-table-wrap { flex:1; overflow-y:auto; }
.hist-tbl { width:100%; border-collapse:collapse; font-size:12px; }
.hist-tbl thead th { background:#1a2235; color:#8494ac; padding:9px 12px; text-align:left; font-size:11px; text-transform:uppercase; letter-spacing:.5px; position:sticky; top:0; }
.hist-row { border-top:1px solid #1a2235; }
.hist-row:hover { background:rgba(56,189,248,.03); }
.hist-row td { padding:9px 12px; color:#d6dfe8; vertical-align:top; }
.hist-date { color:#5a6a87; white-space:nowrap; font-size:11px; }
.mov-badge { display:inline-block; padding:2px 8px; border-radius:10px; font-size:11px; font-weight:600; white-space:nowrap; }
.mov-adjustment { background:rgba(74,222,128,.1); color:#4ade80; }
.mov-reduction  { background:rgba(248,113,113,.1); color:#f87171; }
.mov-transfer   { background:rgba(56,189,248,.1); color:#38bdf8; }
.mov-sale       { background:rgba(99,102,241,.1); color:#818cf8; }
.mov-return     { background:rgba(251,146,60,.1); color:#fb923c; }
.mov-product { font-weight:500; color:#eaf0f7; }
.mov-sku     { font-size:11px; color:#5a6a87; font-family:monospace; }
.mov-route   { font-size:12px; color:#8494ac; white-space:nowrap; }
.route-arrow { margin:0 4px; color:#5a6a87; }
.qty-pos { color:#4ade80; font-weight:700; }
.qty-neg { color:#f87171; font-weight:700; }
.mov-reason { color:#8494ac; }
.mov-note   { font-size:11px; color:#5a6a87; margin-top:2px; }
.hist-pagination { display:flex; align-items:center; justify-content:center; gap:16px; padding:16px; border-top:1px solid #1a2235; flex-shrink:0; }
.pag-btn  { background:none; border:1px solid #2d3a52; color:#8494ac; padding:7px 14px; border-radius:6px; font-size:12px; cursor:pointer; }
.pag-btn:disabled { opacity:.4; cursor:not-allowed; }
.pag-info { font-size:12px; color:#5a6a87; }
</style>
