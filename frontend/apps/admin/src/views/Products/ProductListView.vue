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
      <select v-model="filterCategory" @change="load()" class="select-input">
        <option value="">Todas las categorías</option>
        <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
      </select>
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
          <tr><th>SKU</th><th>Nombre</th><th>Categoría</th><th>Precio base</th><th>Stock</th><th>Activo</th><th></th></tr>
        </thead>
        <tbody>
          <tr v-if="products.length === 0"><td colspan="7" class="empty">Sin productos</td></tr>
          <tr v-for="p in products" :key="p.id" class="tbl-row">
            <td class="mono">{{ p.sku }}</td>
            <td class="td-name">
              <span>{{ p.name }}</span>
              <span class="td-desc">{{ p.description?.slice(0,60) }}</span>
            </td>
            <td class="td-muted">{{ categoryName(p.category_id) || '—' }}</td>
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
    <Teleport to="body">
      <div v-if="drawer" class="drawer-backdrop" @click.self="drawer=false">
        <div class="drawer">
          <div class="drawer-header">
            <h3>{{ editing ? 'Editar producto' : 'Nuevo producto' }}</h3>
            <button class="modal-close" @click="drawer=false">✕</button>
          </div>

          <form @submit.prevent="save()" class="drawer-body">

            <!-- Sección: Información general -->
            <div class="form-section">
              <p class="sec-title">Información general</p>
              <div class="field-row">
                <div class="field">
                  <label>SKU *</label>
                  <input v-model="form.sku" required :disabled="!!editing" />
                </div>
                <div class="field">
                  <label>Precio base *</label>
                  <div class="input-prefix">
                    <span class="prefix">$</span>
                    <input v-model.number="form.base_price" type="number" step="0.01" min="0" required />
                  </div>
                </div>
              </div>
              <div class="field">
                <label>Nombre *</label>
                <input v-model="form.name" required />
              </div>
              <div class="field">
                <label>Descripción</label>
                <textarea v-model="form.description" rows="3" />
              </div>
              <div class="field-row">
                <div class="field">
                  <label>Categoría</label>
                  <select v-model="form.category_id" class="sel">
                    <option value="">— Sin categoría —</option>
                    <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
                  </select>
                </div>
                <div class="field">
                  <label>Tags (comas)</label>
                  <input v-model="formTags" placeholder="ropa, oferta, nuevo" />
                </div>
              </div>
              <div class="check-row">
                <label class="check-label">
                  <input type="checkbox" v-model="form.is_active" />
                  Producto activo (visible en tienda)
                </label>
              </div>
            </div>

            <!-- Sección: Imágenes -->
            <div class="form-section">
              <p class="sec-title">Imágenes</p>
              <div v-for="(img, idx) in form.images" :key="idx" class="image-row">
                <div class="image-thumb">
                  <img v-if="img.url" :src="img.url" @error="img.url=''" />
                  <span v-else>📷</span>
                </div>
                <div class="image-fields">
                  <input v-model="img.url" placeholder="https://… URL de la imagen" class="img-url-input" />
                  <input v-model="img.alt_text" placeholder="Texto alternativo" class="img-alt-input" />
                </div>
                <div class="image-actions">
                  <button type="button" class="btn-main-img" :class="{ active: img.is_main }"
                    @click="setMainImage(idx)" :title="img.is_main ? 'Imagen principal' : 'Marcar como principal'">
                    {{ img.is_main ? '★' : '☆' }}
                  </button>
                  <button type="button" class="btn-rm-img" @click="removeImage(idx)">✕</button>
                </div>
              </div>
              <button type="button" class="btn-add-img" @click="addImage()">+ Agregar imagen</button>
            </div>

            <!-- Sección: Precio por sucursal (solo al editar) -->
            <div class="form-section" v-if="editing">
              <p class="sec-title">Precio por sucursal</p>
              <div class="field-row">
                <div class="field">
                  <label>Sucursal</label>
                  <select v-model="priceOverride.branch_id" class="sel">
                    <option value="">— Seleccionar —</option>
                    <option v-for="b in branches" :key="b.id" :value="b.id">{{ b.name }}</option>
                  </select>
                </div>
                <div class="field">
                  <label>Precio especial</label>
                  <div class="input-prefix">
                    <span class="prefix">$</span>
                    <input v-model.number="priceOverride.price" type="number" step="0.01" min="0" />
                  </div>
                </div>
                <div class="field" style="justify-content:flex-end">
                  <label>&nbsp;</label>
                  <button type="button" class="btn-outline" :disabled="!priceOverride.branch_id || !priceOverride.price"
                    @click="savePriceOverride()">Aplicar</button>
                </div>
              </div>
              <p v-if="priceMsg" class="price-msg" :class="priceErr ? 'err' : 'ok'">{{ priceMsg }}</p>
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
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../../api/client'
import type { Product, Branch } from '../../types'

interface Category { id: string; name: string; slug: string }
interface ImgForm  { url: string; alt_text: string; is_main: boolean }

const products      = ref<Product[]>([])
const categories    = ref<Category[]>([])
const branches      = ref<Branch[]>([])
const loading       = ref(true)
const search        = ref('')
const onlyActive    = ref('')
const filterCategory = ref('')
const page          = ref({ page: 1, total: 0, total_pages: 1, has_next: false, has_prev: false })
const drawer        = ref(false)
const editing       = ref<Product | null>(null)
const saving        = ref(false)
const saveErr       = ref('')
const formTags      = ref('')
const priceOverride = ref({ branch_id: '', price: 0 })
const priceMsg      = ref('')
const priceErr      = ref(false)

const blankForm = (): { sku:string; name:string; description:string; base_price:number; is_active:boolean; tags:string[]; images:ImgForm[]; attributes:Record<string,any>; category_id:string } =>
  ({ sku:'', name:'', description:'', base_price:0, is_active:true, tags:[], images:[], attributes:{}, category_id:'' })

const form = ref(blankForm())

function categoryName(id?: string) {
  return categories.value.find(c => c.id === id)?.name ?? ''
}

async function loadCategories() {
  try {
    const { data } = await api.get('/categories')
    categories.value = (data.data ?? []).flatMap((c: any) => [c, ...(c.children ?? [])])
  } catch {}
}

async function loadBranches() {
  try {
    const { data } = await api.get('/branches')
    branches.value = data.data ?? []
  } catch {}
}

async function load() {
  loading.value = true
  try {
    const p = new URLSearchParams({ page: String(page.value.page), page_size: '20' })
    if (search.value)          p.set('q', search.value)
    if (filterCategory.value)  p.set('category_id', filterCategory.value)
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
  priceOverride.value = { branch_id: '', price: 0 }
  priceMsg.value = ''
  drawer.value   = true
}

function openEdit(p: Product) {
  editing.value  = p
  form.value     = {
    ...blankForm(),
    ...p,
    images: (p.images ?? []).map(i => ({ ...i })),
    category_id: p.category_id ?? '',
  }
  formTags.value = (p.tags ?? []).join(', ')
  saveErr.value  = ''
  priceOverride.value = { branch_id: '', price: 0 }
  priceMsg.value = ''
  drawer.value   = true
}

// Image helpers
function addImage() {
  form.value.images.push({ url: '', alt_text: '', is_main: form.value.images.length === 0 })
}
function removeImage(idx: number) {
  form.value.images.splice(idx, 1)
  if (form.value.images.length > 0 && !form.value.images.some(i => i.is_main)) {
    form.value.images[0].is_main = true
  }
}
function setMainImage(idx: number) {
  form.value.images.forEach((i, j) => { i.is_main = j === idx })
}

async function save() {
  saving.value  = true
  saveErr.value = ''
  const payload = {
    ...form.value,
    tags: formTags.value.split(',').map(t => t.trim()).filter(Boolean),
    images: form.value.images.filter(i => i.url),
  }
  try {
    if (editing.value) {
      await api.put(`/admin/products/${editing.value.id}`, payload)
    } else {
      await api.post('/admin/products', payload)
    }
    drawer.value = false
    await load()
  } catch(e: any) {
    saveErr.value = e.response?.data?.message ?? 'Error guardando'
  } finally { saving.value = false }
}

async function savePriceOverride() {
  priceMsg.value = ''; priceErr.value = false
  try {
    await api.put(`/admin/products/${editing.value!.id}/price`, {
      branch_id: priceOverride.value.branch_id,
      price: priceOverride.value.price,
    })
    priceMsg.value = '✓ Precio especial aplicado'
  } catch(e: any) {
    priceErr.value = true
    priceMsg.value = e.response?.data?.message ?? 'Error aplicando precio'
  }
}

async function toggleActive(p: Product) {
  try {
    await api.patch(`/admin/products/${p.id}/active`, { is_active: !p.is_active })
    p.is_active = !p.is_active
  } catch(e: any) { alert(e.response?.data?.message ?? 'Error') }
}

const fmt = (n: number) => (n ?? 0).toLocaleString('es-MX', { minimumFractionDigits: 2 })

onMounted(() => Promise.all([load(), loadCategories(), loadBranches()]))
</script>

<style scoped>
.page        { display:flex; flex-direction:column; gap:24px; }
.page-header { display:flex; align-items:flex-start; justify-content:space-between; }
.page-title  { font-size:22px; font-weight:700; color:#eaf0f7; margin:0; }
.page-sub    { font-size:13px; color:#5a6a87; margin:4px 0 0; }

.filters      { display:flex; gap:10px; flex-wrap:wrap; }
.search-input { background:#1c2333; border:1px solid #2d3a52; color:#d6dfe8; padding:9px 14px; border-radius:7px; font-size:13px; width:260px; }
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
.btn-outline  { background:none; border:1px solid #38bdf8; color:#38bdf8; padding:9px 16px; border-radius:7px; font-size:13px; cursor:pointer; }
.btn-outline:disabled { opacity:.4; cursor:not-allowed; }
.toggle       { border:none; padding:4px 10px; border-radius:12px; font-size:11px; font-weight:600; cursor:pointer; }
.toggle.on    { background:rgba(74,222,128,.12); color:#4ade80; }
.toggle.off   { background:rgba(248,113,113,.12); color:#f87171; }

.pagination { display:flex; align-items:center; justify-content:center; gap:16px; padding:14px; border-top:1px solid #253047; }
.pagination button { background:#253047; border:none; color:#8494ac; padding:6px 12px; border-radius:6px; cursor:pointer; }
.pagination button:disabled { opacity:.4; cursor:not-allowed; }
.pagination span { font-size:13px; color:#5a6a87; }

/* Drawer */
.drawer-backdrop { position:fixed; inset:0; background:rgba(0,0,0,.7); display:flex; align-items:stretch; justify-content:flex-end; z-index:200; }
.drawer          { background:#1c2333; border-left:1px solid #2d3a52; width:560px; max-width:100vw; height:100vh; overflow-y:auto; display:flex; flex-direction:column; }
.drawer-header   { display:flex; align-items:center; justify-content:space-between; padding:22px 24px; border-bottom:1px solid #2d3a52; position:sticky; top:0; background:#1c2333; z-index:1; }
.drawer-header h3 { font-size:16px; color:#eaf0f7; margin:0; }
.modal-close      { background:none; border:none; color:#5a6a87; font-size:18px; cursor:pointer; }
.drawer-body      { padding:20px 24px; display:flex; flex-direction:column; gap:20px; flex:1; }
.drawer-footer    { display:flex; gap:10px; justify-content:flex-end; padding:16px 24px; border-top:1px solid #253047; position:sticky; bottom:0; background:#1c2333; }

/* Form sections */
.form-section  { display:flex; flex-direction:column; gap:14px; background:#0f1623; border:1px solid #2d3a52; border-radius:10px; padding:18px; }
.sec-title     { font-size:11px; font-weight:700; color:#5a6a87; text-transform:uppercase; letter-spacing:.8px; margin:0 0 -4px; }

.field         { display:flex; flex-direction:column; gap:6px; flex:1; }
.field label   { font-size:11px; color:#8494ac; text-transform:uppercase; letter-spacing:.5px; }
.field input, .field textarea, .sel {
  background:#1c2333; border:1px solid #2d3a52; color:#d6dfe8;
  padding:9px 12px; border-radius:7px; font-size:13px; font-family:inherit; width:100%; box-sizing:border-box;
}
.field input:focus, .field textarea:focus { outline:none; border-color:#38bdf8; }
.field textarea { resize:vertical; }
.field-row     { display:flex; gap:12px; flex-wrap:wrap; }

.input-prefix  { display:flex; align-items:center; background:#1c2333; border:1px solid #2d3a52; border-radius:7px; overflow:hidden; }
.prefix        { padding:9px 10px; font-size:13px; color:#5a6a87; background:#0f1623; border-right:1px solid #2d3a52; flex-shrink:0; }
.input-prefix input { border:none; border-radius:0; background:transparent; }
.input-prefix input:focus { outline:none; }

.check-row     { display:flex; }
.check-label   { display:flex; align-items:center; gap:8px; font-size:13px; color:#d6dfe8; cursor:pointer; }
.check-label input { width:auto; }

/* Images */
.image-row     { display:flex; align-items:center; gap:10px; background:#1a2235; border-radius:8px; padding:10px; }
.image-thumb   { width:48px; height:48px; border-radius:6px; background:#253047; display:flex; align-items:center; justify-content:center; font-size:20px; flex-shrink:0; overflow:hidden; }
.image-thumb img { width:100%; height:100%; object-fit:cover; }
.image-fields  { flex:1; display:flex; flex-direction:column; gap:6px; min-width:0; }
.img-url-input, .img-alt-input { background:#1c2333; border:1px solid #2d3a52; color:#d6dfe8; padding:7px 10px; border-radius:6px; font-size:12px; width:100%; box-sizing:border-box; }
.img-url-input:focus, .img-alt-input:focus { outline:none; border-color:#38bdf8; }
.image-actions { display:flex; flex-direction:column; gap:6px; flex-shrink:0; }
.btn-main-img  { background:none; border:1px solid #2d3a52; color:#5a6a87; width:28px; height:28px; border-radius:6px; cursor:pointer; font-size:14px; }
.btn-main-img.active { background:rgba(251,191,36,.1); border-color:rgba(251,191,36,.4); color:#fbbf24; }
.btn-rm-img    { background:none; border:1px solid rgba(248,113,113,.3); color:#f87171; width:28px; height:28px; border-radius:6px; cursor:pointer; font-size:12px; }
.btn-add-img   { background:none; border:1px dashed #2d3a52; color:#5a6a87; padding:8px; border-radius:7px; font-size:12px; cursor:pointer; text-align:center; }
.btn-add-img:hover { border-color:#38bdf8; color:#38bdf8; }

.price-msg { font-size:12px; margin:0; }
.price-msg.ok  { color:#4ade80; }
.price-msg.err { color:#f87171; }
.err-msg       { color:#f87171; font-size:12px; }
</style>
