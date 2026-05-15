<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">Sucursales</h1>
        <p class="page-sub">{{ branches.length }} sucursal(es)</p>
      </div>
      <button class="btn-primary" @click="openCreate()">+ Nueva sucursal</button>
    </div>

    <div class="cards">
      <div v-if="loading" class="loading">Cargando…</div>
      <div v-else-if="branches.length === 0" class="empty">Sin sucursales registradas</div>
      <div v-for="b in branches" :key="b.id" class="branch-card">
        <div class="bc-header">
          <div class="bc-name">{{ b.name }}</div>
          <span class="badge" :class="b.is_active ? 'badge-ok' : 'badge-off'">
            {{ b.is_active ? 'Activa' : 'Inactiva' }}
          </span>
        </div>
        <div class="bc-address">
          {{ b.address?.street }}, {{ b.address?.city }}, {{ b.address?.state }} {{ b.address?.zip }}
        </div>
        <div class="bc-meta">
          <span>💰 {{ b.settings?.currency ?? 'MXN' }}</span>
          <span>📊 IVA {{ ((b.settings?.tax_rate ?? 0.16) * 100).toFixed(0) }}%</span>
          <span>🏭 {{ b.warehouse_mode ? 'Almacén' : 'Tienda' }}</span>
        </div>
        <div class="bc-actions">
          <button class="btn-sm" @click="openEdit(b)">Editar</button>
          <button class="btn-sm btn-toggle" @click="toggleActive(b)">
            {{ b.is_active ? 'Desactivar' : 'Activar' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Modal crear/editar -->
    <div v-if="drawer" class="modal-overlay" @click.self="drawer=false">
      <div class="drawer">
        <div class="drawer-header">
          <h3>{{ editing ? 'Editar sucursal' : 'Nueva sucursal' }}</h3>
          <button class="modal-close" @click="drawer=false">✕</button>
        </div>
        <form @submit.prevent="save()" class="drawer-body">
          <div class="field"><label>Nombre *</label>
            <input v-model="form.name" required />
          </div>

          <fieldset class="fieldset">
            <legend>Dirección</legend>
            <div class="field-row">
              <div class="field"><label>Calle *</label>
                <input v-model="form.address.street" required />
              </div>
              <div class="field"><label>Ciudad *</label>
                <input v-model="form.address.city" required />
              </div>
            </div>
            <div class="field-row">
              <div class="field"><label>Estado</label>
                <input v-model="form.address.state" />
              </div>
              <div class="field"><label>CP</label>
                <input v-model="form.address.zip" />
              </div>
              <div class="field"><label>País</label>
                <input v-model="form.address.country" placeholder="MX" />
              </div>
            </div>
          </fieldset>

          <fieldset class="fieldset">
            <legend>Configuración</legend>
            <div class="field-row">
              <div class="field"><label>Moneda</label>
                <input v-model="form.settings.currency" placeholder="MXN" />
              </div>
              <div class="field"><label>IVA (0–1)</label>
                <input v-model.number="form.settings.tax_rate" type="number" step="0.01" min="0" max="1" />
              </div>
            </div>
            <div class="check-field">
              <label><input type="checkbox" v-model="form.warehouse_mode" /> Modo almacén (sin ventas directas)</label>
            </div>
          </fieldset>

          <div v-if="saveErr" class="err-msg">{{ saveErr }}</div>
          <div class="drawer-footer">
            <button type="button" class="btn-ghost" @click="drawer=false">Cancelar</button>
            <button type="submit" class="btn-primary" :disabled="saving">
              {{ saving ? 'Guardando…' : (editing ? 'Guardar cambios' : 'Crear sucursal') }}
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
import type { Branch } from '../../types'

const branches = ref<Branch[]>([])
const loading  = ref(true)
const drawer   = ref(false)
const editing  = ref<Branch | null>(null)
const saving   = ref(false)
const saveErr  = ref('')

const blankForm = () => ({
  name: '', warehouse_mode: false,
  address: { street:'', city:'', state:'', zip:'', country:'MX' },
  settings: { tax_rate: 0.16, currency: 'MXN' },
  is_active: true,
})
const form = ref(blankForm())

async function load() {
  loading.value = true
  try { const { data } = await api.get('/branches'); branches.value = data.data ?? [] }
  catch(e) { console.error(e) }
  finally { loading.value = false }
}

function openCreate() { editing.value = null; form.value = blankForm(); saveErr.value=''; drawer.value=true }
function openEdit(b: Branch) { editing.value = b; form.value = JSON.parse(JSON.stringify(b)); saveErr.value=''; drawer.value=true }

async function save() {
  saving.value = true; saveErr.value = ''
  try {
    if (editing.value) await api.put(`/admin/branches/${editing.value.id}`, form.value)
    else await api.post('/admin/branches', form.value)
    drawer.value = false; await load()
  } catch(e: any) { saveErr.value = e.response?.data?.message ?? 'Error guardando' }
  finally { saving.value = false }
}

async function toggleActive(b: Branch) {
  try { await api.patch(`/admin/branches/${b.id}/active`, { is_active: !b.is_active }); b.is_active = !b.is_active }
  catch(e: any) { alert(e.response?.data?.message ?? 'Error') }
}

onMounted(load)
</script>

<style scoped>
.page        { display:flex; flex-direction:column; gap:24px; }
.page-header { display:flex; align-items:flex-start; justify-content:space-between; }
.page-title  { font-size:22px; font-weight:700; color:#eaf0f7; margin:0; }
.page-sub    { font-size:13px; color:#5a6a87; margin:4px 0 0; }
.loading, .empty { padding:40px; text-align:center; color:#5a6a87; }

.cards       { display:grid; grid-template-columns:repeat(auto-fill, minmax(320px,1fr)); gap:16px; }
.branch-card { background:#1c2333; border:1px solid #2d3a52; border-radius:12px; padding:20px; display:flex; flex-direction:column; gap:12px; }
.bc-header   { display:flex; align-items:center; justify-content:space-between; }
.bc-name     { font-size:16px; font-weight:700; color:#eaf0f7; }
.badge       { display:inline-block; padding:3px 9px; border-radius:10px; font-size:11px; font-weight:600; }
.badge-ok    { background:rgba(74,222,128,.1); color:#4ade80; }
.badge-off   { background:rgba(248,113,113,.1); color:#f87171; }
.bc-address  { font-size:13px; color:#8494ac; line-height:1.4; }
.bc-meta     { display:flex; gap:12px; }
.bc-meta span { font-size:12px; color:#5a6a87; }
.bc-actions  { display:flex; gap:8px; }
.btn-sm      { background:#253047; border:none; color:#38bdf8; padding:6px 12px; border-radius:6px; font-size:12px; cursor:pointer; }
.btn-toggle  { color:#fb923c; }

.btn-primary { background:#38bdf8; color:#080c14; border:none; padding:9px 20px; border-radius:7px; font-size:13px; font-weight:700; cursor:pointer; }
.btn-primary:disabled { opacity:.5; cursor:not-allowed; }
.btn-ghost   { background:none; border:1px solid #2d3a52; color:#8494ac; padding:9px 16px; border-radius:7px; font-size:13px; cursor:pointer; }

.modal-overlay { position:fixed; inset:0; background:rgba(0,0,0,.7); display:flex; align-items:center; justify-content:flex-end; z-index:100; }
.drawer        { background:#1c2333; border-left:1px solid #2d3a52; width:520px; max-width:100vw; height:100vh; overflow-y:auto; display:flex; flex-direction:column; }
.drawer-header { display:flex; align-items:center; justify-content:space-between; padding:24px 24px 0; }
.drawer-header h3 { font-size:16px; color:#eaf0f7; margin:0; }
.modal-close   { background:none; border:none; color:#5a6a87; font-size:18px; cursor:pointer; }
.drawer-body   { padding:20px 24px; display:flex; flex-direction:column; gap:16px; flex:1; }
.drawer-footer { display:flex; gap:10px; justify-content:flex-end; margin-top:auto; padding-top:16px; border-top:1px solid #253047; }
.field         { display:flex; flex-direction:column; gap:6px; flex:1; }
.field label   { font-size:11px; color:#8494ac; text-transform:uppercase; letter-spacing:.5px; }
.field input   { background:#0f1623; border:1px solid #2d3a52; color:#d6dfe8; padding:9px 12px; border-radius:7px; font-size:13px; }
.field input:focus { outline:none; border-color:#38bdf8; }
.field-row     { display:flex; gap:12px; }
.fieldset      { border:1px solid #2d3a52; border-radius:8px; padding:16px; display:flex; flex-direction:column; gap:12px; }
.fieldset legend { color:#8494ac; font-size:11px; text-transform:uppercase; letter-spacing:.5px; padding:0 6px; }
.check-field   { display:flex; align-items:center; }
.check-field label { display:flex; align-items:center; gap:8px; font-size:13px; color:#d6dfe8; cursor:pointer; }
.err-msg       { color:#f87171; font-size:12px; }
</style>
