<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">Productos</h1>
        <p class="page-sub">{{ page.total }} productos en total</p>
      </div>
      <button class="btn-primary" @click="openCreate()">+ Nuevo producto</button>
    </div>

    <!-- Filtros -->
    <div class="filters">
      <input v-model="search" class="search-input" placeholder="Buscar por nombre o SKU…" @input="debouncedLoad()" />
      <select v-model="onlyActive" @change="load()" class="select-input">
        <option value="">Todos</option>
        <option value="true">Solo activos</option>
        <option value="false">Solo inactivos</option>
      </select>
    </div>

    <!-- Tabla -->
    <div class="table-wrap">
      <div v-if="loading" class="loading">Cargando productos…</div>
      <table v-else class="tbl">
        <thead>
          <tr><th>SKU</th><th>Nombre</th><th>Precio base</th><th>Stock</th><th>Activo</th><th></th></tr>
        </thead>
        <tbody>
          <tr v-if="products.length === 0"><td colspan="6" class="empty">Sin productos</td></tr>
          <tr v-for="p in products" :key="p.id" class="tbl-row">
            <td class="mono">{{ p.sku }}</td>
            <td class="td-name">
              <span>{{ p.name }}</span>
              <span class="td-desc">{{ p.description?.slice(0,60) }}</span>
            </td>
            <td class="td-amount">${{ fmt(p.base_price) }}</td>
            <td>
              <span v-if="p.stock != null" :class="p.stock <= 10 ? 'stock-low' : 'stock-ok'">{{ p.stock }}</span>
              <span v-else class="td-muted">—</span>
            </td>
            <td>
              <button class="toggle" :class="p.is_active ? 'on' : 'off'"
                @click="toggleActive(p)">{{ p.is_active ? 'Activo' : 'Inactivo' }}</button>
            </td>
            <td class="td-actions">
              <button class="btn-sm" @click="openEdit(p)">Editar</button>
            </td>
          </tr>
        </tbody>
      </table>
      <div class="pagination" v-if="page.total_pages > 1">
        <button @click="page.page--; load()" :disabled="!page.has_prev">‹ Ant</button>
        <span>{{ page.page }} / {{ page.total_pages }}</span>
        <button @click="page.page++; load()" :disabled="!page.has_next">Sig ›</button>
      </div>
    </div>

    <!-- Drawer crear/editar -->
    <div v-if="drawer" class="modal-overlay" @click.self="drawer=false">
      <div class="drawer">
        <div class="drawer-header">
          <h3>{{ editing ? 'Editar producto' : 'Nuevo producto' }}</h3>
          <button class="modal-close" @click="drawer=false">✕</button>
        </div>
        <form @submit.prevent="save()" class="drawer-body">
          <div class="field-row">
            <div class="field"><label>SKU *</label>
              <input v-model="form.sku" required :disabled="!!editing" />
            </div>
            <div class="field"><label>Precio base *</label>
              <input v-model.number="form.base_price" type="number" step="0.01" min="0" required />
            </div>
          </div>
          <div class="field"><label>Nombre *</label>
            <input v-model="form.name" required />
          </div>
          <div class="field"><label>Descripción</label>
            <textarea v-model="form.description" rows="3" />
          </div>
          <div class="field-row">
            <div class="field"><label>Tags (comas)</label>
              <input v-model="formTags" placeholder="ropa,oferta" />
            </div>
            <div class="field check-field">
              <label>
                <input type="checkbox" v-model="form.is_active" />
                Producto activo
              </label>
            </div>
          </div>
          <div v-if="saveErr" class="err-msg">{{ saveErr }}</div>
          <div class="drawer-footer">
            <button type="button" class="btn-ghost" @click="drawer=false">Cancelar</button>
            <button type="submit" class="btn-primary" :disabled="saving">
              {{ saving ? 'Guardando…' : (editing ? 'Guardar cambios' : 'Crear producto') }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../../api/client'
import type { Product } from '../../types'

const products  = ref<Product[]>([])
const loading   = ref(true)
const search    = ref('')
const onlyActive = ref('')
const page      = ref({ page: 1, total: 0, total_pages: 1, has_next: false, has_prev: false })
const drawer    = ref(false)
const editing   = ref<Product | null>(null)
const saving    = ref(false)
const saveErr   = ref('')
const formTags  = ref('')

const blankForm = () => ({ sku:'', name:'', description:'', base_price:0, is_active:true, tags:[] as string[], images:[], attributes:{} })
const form = ref(blankForm())

async function load() {
  loading.value = true
  try {
    const p = new URLSearchParams({ page: String(page.value.page), page_size:'20' })
    if (search.value)    p.set('q', search.value)
    const { data } = await api.get(`/products?${p}`)
    products.value = data.data ?? []
    page.value     = data
  } catch(e) { console.error(e) }
  finally { loading.value = false }
}

let debTimer: any
function debouncedLoad() { clearTimeout(debTimer); debTimer = setTimeout(load, 350) }

function openCreate() {
  editing.value = null
  form.value    = blankForm()
  formTags.value = ''
  saveErr.value  = ''
  drawer.value   = true
}

function openEdit(p: Product) {
  editing.value  = p
  form.value     = { ...p }
  formTags.value = (p.tags ?? []).join(', ')
  saveErr.value  = ''
  drawer.value   = true
}

async function save() {
  saving.value  = true
  saveErr.value = ''
  form.value.tags = formTags.value.split(',').map(t => t.trim()).filter(Boolean)
  try {
    if (editing.value) {
      await api.put(`/admin/products/${editing.value.id}`, form.value)
    } else {
      await api.post('/admin/products', form.value)
    }
    drawer.value = false
    await load()
  } catch(e: any) {
    saveErr.value = e.response?.data?.message ?? 'Error guardando'
  } finally { saving.value = false }
}

async function toggleActive(p: Product) {
  try {
    await api.patch(`/admin/products/${p.id}/active`, { is_active: !p.is_active })
    p.is_active = !p.is_active
  } catch(e: any) { alert(e.response?.data?.message ?? 'Error') }
}

const fmt = (n: number) => (n ?? 0).toLocaleString('es-MX', { minimumFractionDigits: 2 })
onMounted(load)
</script>

<style scoped>
.page        { display:flex; flex-direction:column; gap:24px; }
.page-header { display:flex; align-items:flex-start; justify-content:space-between; }
.page-title  { font-size:22px; font-weight:700; color:#eaf0f7; margin:0; }
.page-sub    { font-size:13px; color:#5a6a87; margin:4px 0 0; }

.filters      { display:flex; gap:10px; flex-wrap:wrap; }
.search-input { background:#1c2333; border:1px solid #2d3a52; color:#d6dfe8; padding:9px 14px; border-radius:7px; font-size:13px; width:280px; }
.search-input:focus { outline:none; border-color:#38bdf8; }
.select-input { background:#1c2333; border:1px solid #2d3a52; color:#d6dfe8; padding:9px 12px; border-radius:7px; font-size:13px; }

.table-wrap   { background:#1c2333; border:1px solid #2d3a52; border-radius:10px; overflow:hidden; }
.loading      { padding:40px; text-align:center; color:#5a6a87; }
.tbl          { width:100%; border-collapse:collapse; font-size:13px; }
.tbl thead th { background:#253047; color:#8494ac; padding:10px 14px; text-align:left; font-weight:600; font-size:11px; text-transform:uppercase; letter-spacing:.5px; }
.tbl-row      { border-top:1px solid #253047; transition:background .15s; }
.tbl-row:hover { background:rgba(56,189,248,.04); }
.tbl-row td   { padding:10px 14px; color:#d6dfe8; }
.empty        { padding:40px; text-align:center; color:#5a6a87; }
.mono         { font-family:monospace; font-size:12px; color:#8494ac; }
.td-name      { display:flex; flex-direction:column; gap:2px; }
.td-desc      { font-size:11px; color:#5a6a87; }
.td-muted     { color:#5a6a87; }
.td-amount    { font-weight:600; color:#4ade80; }
.td-actions   { display:flex; gap:6px; }
.stock-ok     { color:#4ade80; font-weight:600; }
.stock-low    { color:#f87171; font-weight:600; }
.btn-sm       { background:#253047; border:none; color:#38bdf8; padding:5px 10px; border-radius:5px; font-size:12px; cursor:pointer; }
.btn-primary  { background:#38bdf8; color:#080c14; border:none; padding:9px 20px; border-radius:7px; font-size:13px; font-weight:700; cursor:pointer; }
.btn-primary:disabled { opacity:.5; cursor:not-allowed; }
.btn-ghost    { background:none; border:1px solid #2d3a52; color:#8494ac; padding:9px 16px; border-radius:7px; font-size:13px; cursor:pointer; }
.toggle       { border:none; padding:4px 10px; border-radius:12px; font-size:11px; font-weight:600; cursor:pointer; }
.toggle.on    { background:rgba(74,222,128,.12); color:#4ade80; }
.toggle.off   { background:rgba(248,113,113,.12); color:#f87171; }

.pagination { display:flex; align-items:center; justify-content:center; gap:16px; padding:14px; border-top:1px solid #253047; }
.pagination button { background:#253047; border:none; color:#8494ac; padding:6px 12px; border-radius:6px; cursor:pointer; }
.pagination button:disabled { opacity:.4; cursor:not-allowed; }
.pagination span { font-size:13px; color:#5a6a87; }

.modal-overlay { position:fixed; inset:0; background:rgba(0,0,0,.7); display:flex; align-items:center; justify-content:flex-end; z-index:100; }
.drawer        { background:#1c2333; border-left:1px solid #2d3a52; width:480px; max-width:100vw; height:100vh; overflow-y:auto; display:flex; flex-direction:column; }
.drawer-header { display:flex; align-items:center; justify-content:space-between; padding:24px 24px 0; }
.drawer-header h3 { font-size:16px; color:#eaf0f7; margin:0; }
.modal-close   { background:none; border:none; color:#5a6a87; font-size:18px; cursor:pointer; }
.drawer-body   { padding:20px 24px; display:flex; flex-direction:column; gap:16px; flex:1; }
.drawer-footer { display:flex; gap:10px; justify-content:flex-end; margin-top:auto; padding-top:16px; border-top:1px solid #253047; }
.field         { display:flex; flex-direction:column; gap:6px; flex:1; }
.field label   { font-size:11px; color:#8494ac; text-transform:uppercase; letter-spacing:.5px; }
.field input, .field textarea, .field select { background:#0f1623; border:1px solid #2d3a52; color:#d6dfe8; padding:9px 12px; border-radius:7px; font-size:13px; font-family:inherit; }
.field input:focus, .field textarea:focus { outline:none; border-color:#38bdf8; }
.field textarea { resize:vertical; }
.field-row     { display:flex; gap:12px; }
.check-field   { justify-content:flex-end; }
.check-field label { display:flex; align-items:center; gap:8px; font-size:13px; color:#d6dfe8; cursor:pointer; text-transform:none; letter-spacing:0; }
.err-msg       { color:#f87171; font-size:12px; }
</style>
