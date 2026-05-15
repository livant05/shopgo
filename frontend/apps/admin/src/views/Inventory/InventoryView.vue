<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">Inventario</h1>
        <p class="page-sub">{{ items.length }} registros</p>
      </div>
      <div class="header-actions">
        <button class="btn-outline" @click="openTransfer()">↔ Transferir</button>
        <button class="btn-primary"  @click="openAdjust()">± Ajustar</button>
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
            <td class="td-muted">{{ inv.branch_id?.slice(0,8) }}…</td>
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

    <!-- Modal Ajustar -->
    <div v-if="adjustModal" class="modal-overlay" @click.self="adjustModal=false">
      <div class="modal">
        <div class="modal-header">
          <h3>Ajustar Stock</h3>
          <button class="modal-close" @click="adjustModal=false">✕</button>
        </div>
        <form @submit.prevent="doAdjust()" class="modal-body">
          <div class="field"><label>Product ID *</label>
            <input v-model="adj.product_id" required placeholder="UUID del producto" />
          </div>
          <div class="field"><label>Branch ID *</label>
            <input v-model="adj.branch_id" required placeholder="UUID de la sucursal" />
          </div>
          <div class="field-row">
            <div class="field"><label>Delta (+ entrada / - salida) *</label>
              <input v-model.number="adj.delta" type="number" required />
            </div>
            <div class="field"><label>Razón *</label>
              <input v-model="adj.reason" required placeholder="compra, merma…" />
            </div>
          </div>
          <div class="field"><label>Nota</label>
            <input v-model="adj.note" placeholder="Opcional" />
          </div>
          <div v-if="adjErr" class="err-msg">{{ adjErr }}</div>
          <div class="modal-footer">
            <button type="button" class="btn-ghost" @click="adjustModal=false">Cancelar</button>
            <button type="submit" class="btn-primary" :disabled="adjLoading">
              {{ adjLoading ? 'Ajustando…' : 'Confirmar ajuste' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Modal Transferir -->
    <div v-if="transferModal" class="modal-overlay" @click.self="transferModal=false">
      <div class="modal">
        <div class="modal-header">
          <h3>Transferir Stock</h3>
          <button class="modal-close" @click="transferModal=false">✕</button>
        </div>
        <form @submit.prevent="doTransfer()" class="modal-body">
          <div class="field"><label>Product ID *</label>
            <input v-model="tr.product_id" required placeholder="UUID del producto" />
          </div>
          <div class="field-row">
            <div class="field"><label>Desde Sucursal *</label>
              <select v-model="tr.from_branch_id" required class="sel">
                <option value="">— Seleccionar —</option>
                <option v-for="b in branches" :key="b.id" :value="b.id">{{ b.name }}</option>
              </select>
            </div>
            <div class="field"><label>Hacia Sucursal *</label>
              <select v-model="tr.to_branch_id" required class="sel">
                <option value="">— Seleccionar —</option>
                <option v-for="b in branches" :key="b.id" :value="b.id">{{ b.name }}</option>
              </select>
            </div>
          </div>
          <div class="field-row">
            <div class="field"><label>Cantidad *</label>
              <input v-model.number="tr.quantity" type="number" min="1" required />
            </div>
            <div class="field"><label>Nota</label>
              <input v-model="tr.note" placeholder="Opcional" />
            </div>
          </div>
          <div v-if="trErr" class="err-msg">{{ trErr }}</div>
          <div class="modal-footer">
            <button type="button" class="btn-ghost" @click="transferModal=false">Cancelar</button>
            <button type="submit" class="btn-primary" :disabled="trLoading">
              {{ trLoading ? 'Transfiriendo…' : 'Confirmar transferencia' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '../../api/client'
import type { Branch } from '../../types'

interface InvItem { product_id: string; branch_id: string; quantity: number; reserved_qty: number; reorder_point: number }

const items        = ref<InvItem[]>([])
const branches     = ref<Branch[]>([])
const loading      = ref(true)
const branchFilter = ref('')
const search       = ref('')
const adjustModal  = ref(false)
const transferModal = ref(false)
const adj = ref({ product_id:'', branch_id:'', delta:0, reason:'', note:'' })
const tr  = ref({ product_id:'', from_branch_id:'', to_branch_id:'', quantity:1, note:'' })
const adjLoading = ref(false); const adjErr = ref('')
const trLoading  = ref(false); const trErr  = ref('')

const available = (inv: InvItem) => inv.quantity - inv.reserved_qty

const filtered = computed(() =>
  items.value.filter(i => !search.value ||
    i.product_id.toLowerCase().includes(search.value.toLowerCase()))
)

async function load() {
  loading.value = true
  try {
    const url = branchFilter.value ? `/admin/inventory?branch_id=${branchFilter.value}` : '/admin/inventory'
    const { data } = await api.get(url)
    items.value = data.data ?? []
  } catch(e) { console.error(e) }
  finally { loading.value = false }
}

async function loadBranches() {
  const { data } = await api.get('/branches')
  branches.value = data.data ?? []
}

function openAdjust() {
  adj.value = { product_id:'', branch_id:'', delta:0, reason:'', note:'' }
  adjErr.value = ''; adjustModal.value = true
}

function openTransfer() {
  tr.value = { product_id:'', from_branch_id:'', to_branch_id:'', quantity:1, note:'' }
  trErr.value = ''; transferModal.value = true
}

async function doAdjust() {
  adjLoading.value = true; adjErr.value = ''
  try {
    await api.patch('/admin/inventory', adj.value)
    adjustModal.value = false; await load()
  } catch(e: any) { adjErr.value = e.response?.data?.message ?? 'Error' }
  finally { adjLoading.value = false }
}

async function doTransfer() {
  trLoading.value = true; trErr.value = ''
  try {
    await api.post('/admin/inventory/transfer', tr.value)
    transferModal.value = false; await load()
  } catch(e: any) { trErr.value = e.response?.data?.message ?? 'Error' }
  finally { trLoading.value = false }
}

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

.modal-overlay { position:fixed; inset:0; background:rgba(0,0,0,.7); display:flex; align-items:center; justify-content:center; z-index:100; }
.modal         { background:#1c2333; border:1px solid #2d3a52; border-radius:12px; width:520px; max-width:95vw; }
.modal-header  { display:flex; align-items:center; justify-content:space-between; padding:20px 24px 0; }
.modal-header h3 { font-size:16px; color:#eaf0f7; margin:0; }
.modal-close   { background:none; border:none; color:#5a6a87; font-size:18px; cursor:pointer; }
.modal-body    { padding:20px 24px 0; display:flex; flex-direction:column; gap:14px; }
.modal-footer  { display:flex; gap:10px; justify-content:flex-end; padding:20px 24px 24px; }
.field         { display:flex; flex-direction:column; gap:6px; flex:1; }
.field label   { font-size:11px; color:#8494ac; text-transform:uppercase; letter-spacing:.5px; }
.field input, .field .sel { background:#0f1623; border:1px solid #2d3a52; color:#d6dfe8; padding:9px 12px; border-radius:7px; font-size:13px; width:100%; }
.field input:focus { outline:none; border-color:#38bdf8; }
.field-row     { display:flex; gap:12px; }
.err-msg       { color:#f87171; font-size:12px; padding-bottom:4px; }
</style>
