<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">Cupones</h1>
        <p class="page-sub">Gestión de códigos de descuento</p>
      </div>
      <button class="btn-primary" @click="showForm = true">+ Nuevo cupón</button>
    </div>

    <!-- Table -->
    <div class="table-wrap">
      <div v-if="loading" class="loading">Cargando cupones…</div>
      <table v-else class="tbl">
        <thead>
          <tr>
            <th>Código</th><th>Tipo</th><th>Valor</th>
            <th>Usos</th><th>Vence</th><th>Estado</th><th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="coupons.length === 0">
            <td colspan="7" class="empty">Sin cupones. Crea el primero.</td>
          </tr>
          <tr v-for="c in coupons" :key="c.id" class="tbl-row">
            <td class="mono code-cell">{{ c.code }}</td>
            <td>
              <span class="badge-type" :class="c.type">
                {{ c.type === 'percent' ? '%' : '$' }}
              </span>
            </td>
            <td class="td-val">
              {{ c.type === 'percent' ? c.value + '%' : '$' + fmt(c.value) }}
            </td>
            <td class="td-muted">
              {{ c.uses_count }}{{ c.max_uses ? ' / ' + c.max_uses : '' }}
            </td>
            <td class="td-muted">
              {{ c.valid_until ? fmtDate(c.valid_until) : '—' }}
            </td>
            <td>
              <span class="badge-status" :class="{ active: c.is_active, inactive: !c.is_active }">
                {{ c.is_active ? 'Activo' : 'Inactivo' }}
              </span>
            </td>
            <td>
              <button class="btn-toggle"
                :class="{ deactivate: c.is_active }"
                :disabled="toggling === c.id"
                @click="toggle(c)">
                {{ c.is_active ? 'Desactivar' : 'Activar' }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Create modal -->
    <div v-if="showForm" class="modal-overlay" @click.self="closeForm">
      <div class="modal">
        <div class="modal-header">
          <h3>Nuevo cupón</h3>
          <button class="modal-close" @click="closeForm">✕</button>
        </div>
        <div class="modal-body">
          <div class="field">
            <label>Código *</label>
            <input v-model="form.code" class="inp" placeholder="VERANO20" style="text-transform:uppercase"
                   @input="form.code = form.code.toUpperCase()" />
          </div>
          <div class="row2">
            <div class="field">
              <label>Tipo *</label>
              <select v-model="form.type" class="inp">
                <option value="percent">Porcentaje (%)</option>
                <option value="fixed">Fijo ($)</option>
              </select>
            </div>
            <div class="field">
              <label>Valor *</label>
              <input v-model.number="form.value" type="number" min="0.01" step="0.01"
                     class="inp" :placeholder="form.type === 'percent' ? 'ej. 10' : 'ej. 50'" />
            </div>
          </div>
          <div class="row2">
            <div class="field">
              <label>Válido hasta</label>
              <input v-model="form.valid_until" type="datetime-local" class="inp" />
            </div>
            <div class="field">
              <label>Máx. usos</label>
              <input v-model.number="form.max_uses" type="number" min="1" class="inp" placeholder="Sin límite" />
            </div>
          </div>
          <p v-if="formErr" class="err-msg">{{ formErr }}</p>
          <div class="modal-actions">
            <button class="btn-cancel" @click="closeForm">Cancelar</button>
            <button class="btn-primary" :disabled="creating" @click="create">
              {{ creating ? 'Creando…' : 'Crear cupón' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../../api/client'

interface Coupon {
  id: string; code: string; type: string; value: number
  valid_until?: string; max_uses?: number; uses_count: number; is_active: boolean
}

const coupons  = ref<Coupon[]>([])
const loading  = ref(true)
const toggling = ref<string | null>(null)
const showForm = ref(false)
const creating = ref(false)
const formErr  = ref('')

const form = ref({ code: '', type: 'percent', value: 0, valid_until: '', max_uses: undefined as number | undefined })

async function load() {
  loading.value = true
  try {
    const { data } = await api.get('/admin/coupons')
    coupons.value = data.data ?? []
  } catch { coupons.value = [] }
  loading.value = false
}

async function toggle(c: Coupon) {
  toggling.value = c.id
  try {
    await api.patch(`/admin/coupons/${c.id}/active`, { active: !c.is_active })
    c.is_active = !c.is_active
  } catch {}
  toggling.value = null
}

async function create() {
  formErr.value = ''
  if (!form.value.code || !form.value.value) {
    formErr.value = 'Código y valor son requeridos'; return
  }
  if (form.value.type === 'percent' && form.value.value > 100) {
    formErr.value = 'El porcentaje no puede ser mayor a 100'; return
  }
  creating.value = true
  try {
    const payload: any = {
      code: form.value.code,
      type: form.value.type,
      value: form.value.value,
    }
    if (form.value.valid_until) payload.valid_until = new Date(form.value.valid_until).toISOString()
    if (form.value.max_uses)    payload.max_uses    = form.value.max_uses
    const { data } = await api.post('/admin/coupons', payload)
    coupons.value.unshift(data)
    closeForm()
  } catch (e: any) {
    formErr.value = e.response?.data?.message ?? 'Error al crear el cupón'
  }
  creating.value = false
}

function closeForm() {
  showForm.value = false; formErr.value = ''
  form.value = { code: '', type: 'percent', value: 0, valid_until: '', max_uses: undefined }
}

const fmt     = (n: number) => (n ?? 0).toLocaleString('es-MX', { minimumFractionDigits: 2 })
const fmtDate = (s: string) => new Date(s).toLocaleDateString('es-MX', { year: 'numeric', month: 'short', day: 'numeric' })

onMounted(load)
</script>

<style scoped>
.page        { display: flex; flex-direction: column; gap: 24px; }
.page-header { display: flex; align-items: flex-start; justify-content: space-between; }
.page-title  { font-size: 22px; font-weight: 700; color: #eaf0f7; margin: 0; }
.page-sub    { font-size: 13px; color: #5a6a87; margin: 4px 0 0; }

.btn-primary { background: #38bdf8; color: #080c14; border: none; padding: 8px 16px; border-radius: 7px; font-size: 13px; font-weight: 700; cursor: pointer; }
.btn-primary:disabled { opacity: .5; cursor: not-allowed; }
.btn-primary:hover:not(:disabled) { background: #7dd3fc; }

.table-wrap { background: #1c2333; border: 1px solid #2d3a52; border-radius: 10px; overflow: hidden; }
.loading    { padding: 40px; text-align: center; color: #5a6a87; }
.tbl        { width: 100%; border-collapse: collapse; font-size: 13px; }
.tbl thead th { background: #253047; color: #8494ac; padding: 10px 14px; text-align: left; font-weight: 600; font-size: 11px; text-transform: uppercase; letter-spacing: .5px; }
.tbl-row    { border-top: 1px solid #253047; }
.tbl-row td { padding: 11px 14px; color: #d6dfe8; }
.empty      { padding: 40px; text-align: center; color: #5a6a87; }
.mono       { font-family: monospace; }
.code-cell  { font-weight: 700; color: #38bdf8; letter-spacing: .06em; }
.td-muted   { color: #5a6a87; }
.td-val     { font-weight: 600; color: #4ade80; }

.badge-type         { display: inline-block; padding: 2px 8px; border-radius: 10px; font-size: 11px; font-weight: 700; }
.badge-type.percent { background: rgba(167,139,250,.15); color: #a78bfa; }
.badge-type.fixed   { background: rgba(74,222,128,.15);  color: #4ade80; }

.badge-status         { display: inline-block; padding: 3px 9px; border-radius: 12px; font-size: 11px; font-weight: 600; }
.badge-status.active   { background: rgba(74,222,128,.12); color: #4ade80; }
.badge-status.inactive { background: rgba(248,113,113,.12); color: #f87171; }

.btn-toggle          { background: #253047; border: none; color: #38bdf8; padding: 5px 10px; border-radius: 5px; font-size: 11px; cursor: pointer; }
.btn-toggle.deactivate { color: #f87171; }
.btn-toggle:disabled   { opacity: .4; cursor: not-allowed; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,.7); display: flex; align-items: center; justify-content: center; z-index: 100; }
.modal         { background: #1c2333; border: 1px solid #2d3a52; border-radius: 12px; width: 480px; }
.modal-header  { display: flex; align-items: center; justify-content: space-between; padding: 20px 24px 0; }
.modal-header h3 { font-size: 16px; color: #eaf0f7; margin: 0; }
.modal-close   { background: none; border: none; color: #5a6a87; font-size: 18px; cursor: pointer; }
.modal-body    { padding: 20px 24px 24px; display: flex; flex-direction: column; gap: 14px; }
.field         { display: flex; flex-direction: column; gap: 5px; }
.field label   { font-size: 11px; font-weight: 600; color: #8494ac; text-transform: uppercase; letter-spacing: .5px; }
.row2          { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.inp           { background: #0f1623; border: 1px solid #2d3a52; color: #d6dfe8; padding: 8px 10px; border-radius: 7px; font-size: 13px; outline: none; width: 100%; box-sizing: border-box; }
.inp:focus     { border-color: #38bdf8; }
.err-msg       { color: #f87171; font-size: 12px; }
.modal-actions { display: flex; justify-content: flex-end; gap: 8px; margin-top: 4px; }
.btn-cancel    { background: #253047; border: none; color: #8494ac; padding: 8px 14px; border-radius: 7px; font-size: 13px; cursor: pointer; }
.btn-cancel:hover { background: #2d3a52; }
</style>
