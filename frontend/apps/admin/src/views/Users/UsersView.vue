<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">Usuarios</h1>
        <p class="page-sub">{{ page.total ?? 0 }} usuarios registrados</p>
      </div>
      <button class="btn-primary" @click="openCreate()">+ Nuevo usuario</button>
    </div>

    <div class="table-wrap">
      <div v-if="loading" class="loading">Cargando usuarios…</div>
      <table v-else class="tbl">
        <thead>
          <tr><th>Nombre</th><th>Email</th><th>Rol</th><th>Sucursal</th><th>Activo</th><th></th></tr>
        </thead>
        <tbody>
          <tr v-if="users.length === 0"><td colspan="6" class="empty">Sin usuarios</td></tr>
          <tr v-for="u in users" :key="u.id" class="tbl-row">
            <td>
              <div class="user-name">{{ u.first_name }} {{ u.last_name }}</div>
            </td>
            <td class="td-muted">{{ u.email }}</td>
            <td><span class="role-badge" :class="u.role">{{ u.role }}</span></td>
            <td class="td-muted mono">{{ u.branch_id ? u.branch_id.slice(0,8) + '…' : '—' }}</td>
            <td>
              <button class="toggle" :class="u.is_active ? 'on' : 'off'"
                @click="toggleActive(u)">{{ u.is_active ? 'Activo' : 'Inactivo' }}</button>
            </td>
            <td><button class="btn-sm" @click="openEdit(u)">Editar</button></td>
          </tr>
        </tbody>
      </table>
      <div class="pagination" v-if="page.total_pages > 1">
        <button @click="curPage--; load()" :disabled="curPage <= 1">‹ Ant</button>
        <span>{{ curPage }} / {{ page.total_pages }}</span>
        <button @click="curPage++; load()" :disabled="curPage >= page.total_pages">Sig ›</button>
      </div>
    </div>

    <!-- Drawer -->
    <div v-if="drawer" class="modal-overlay" @click.self="drawer=false">
      <div class="drawer">
        <div class="drawer-header">
          <h3>{{ editing ? 'Editar usuario' : 'Nuevo usuario' }}</h3>
          <button class="modal-close" @click="drawer=false">✕</button>
        </div>
        <form @submit.prevent="save()" class="drawer-body">
          <div class="field-row">
            <div class="field"><label>Nombre</label>
              <input v-model="form.first_name" />
            </div>
            <div class="field"><label>Apellido</label>
              <input v-model="form.last_name" />
            </div>
          </div>
          <div class="field"><label>Email *</label>
            <input v-model="form.email" type="email" required :disabled="!!editing" />
          </div>
          <div class="field" v-if="!editing"><label>Contraseña *</label>
            <input v-model="form.password" type="password" required minlength="8" />
          </div>
          <div class="field-row">
            <div class="field"><label>Rol *</label>
              <select v-model="form.role" required class="sel">
                <option value="admin">admin</option>
                <option value="manager">manager</option>
                <option value="staff">staff</option>
                <option value="customer">customer</option>
              </select>
            </div>
            <div class="field"><label>Sucursal</label>
              <select v-model="form.branch_id" class="sel">
                <option value="">— Sin sucursal —</option>
                <option v-for="b in branches" :key="b.id" :value="b.id">{{ b.name }}</option>
              </select>
            </div>
          </div>
          <div class="field"><label>Teléfono</label>
            <input v-model="form.phone" type="tel" />
          </div>
          <div v-if="saveErr" class="err-msg">{{ saveErr }}</div>
          <div class="drawer-footer">
            <button type="button" class="btn-ghost" @click="drawer=false">Cancelar</button>
            <button type="submit" class="btn-primary" :disabled="saving">
              {{ saving ? 'Guardando…' : (editing ? 'Guardar' : 'Crear usuario') }}
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
import type { User, Branch } from '../../types'

const users    = ref<User[]>([])
const branches = ref<Branch[]>([])
const loading  = ref(true)
const curPage  = ref(1)
const page     = ref({ total:0, total_pages:1 })
const drawer   = ref(false)
const editing  = ref<User | null>(null)
const saving   = ref(false)
const saveErr  = ref('')

const blankForm = () => ({ email:'', password:'', first_name:'', last_name:'', role:'staff' as any, branch_id:'', phone:'' })
const form = ref(blankForm())

async function load() {
  loading.value = true
  try {
    const { data } = await api.get(`/admin/users?page=${curPage.value}&page_size=20`)
    users.value = data.data ?? []
    page.value  = data
  } catch(e) { console.error(e) }
  finally { loading.value = false }
}

async function loadBranches() {
  const { data } = await api.get('/branches')
  branches.value = data.data ?? []
}

function openCreate() { editing.value = null; form.value = blankForm(); saveErr.value=''; drawer.value=true }
function openEdit(u: User) { editing.value = u; form.value = { ...u, password:'' } as any; saveErr.value=''; drawer.value=true }

async function save() {
  saving.value = true; saveErr.value = ''
  try {
    if (editing.value) await api.put(`/admin/users/${editing.value.id}`, form.value)
    else await api.post('/admin/users', form.value)
    drawer.value = false; await load()
  } catch(e: any) { saveErr.value = e.response?.data?.message ?? 'Error guardando' }
  finally { saving.value = false }
}

async function toggleActive(u: User) {
  try {
    await api.patch(`/admin/users/${u.id}/active`, { is_active: !u.is_active })
    u.is_active = !u.is_active
  } catch(e: any) { alert(e.response?.data?.message ?? 'Error') }
}

onMounted(() => Promise.all([load(), loadBranches()]))
</script>

<style scoped>
.page        { display:flex; flex-direction:column; gap:24px; }
.page-header { display:flex; align-items:flex-start; justify-content:space-between; }
.page-title  { font-size:22px; font-weight:700; color:#eaf0f7; margin:0; }
.page-sub    { font-size:13px; color:#5a6a87; margin:4px 0 0; }

.table-wrap { background:#1c2333; border:1px solid #2d3a52; border-radius:10px; overflow:hidden; }
.loading    { padding:40px; text-align:center; color:#5a6a87; }
.tbl        { width:100%; border-collapse:collapse; font-size:13px; }
.tbl thead th { background:#253047; color:#8494ac; padding:10px 14px; text-align:left; font-weight:600; font-size:11px; text-transform:uppercase; letter-spacing:.5px; }
.tbl-row    { border-top:1px solid #253047; transition:background .15s; }
.tbl-row:hover { background:rgba(56,189,248,.04); }
.tbl-row td { padding:10px 14px; color:#d6dfe8; }
.empty      { padding:40px; text-align:center; color:#5a6a87; }
.td-muted   { color:#5a6a87; }
.mono       { font-family:monospace; font-size:12px; }
.user-name  { font-weight:600; }
.btn-sm     { background:#253047; border:none; color:#38bdf8; padding:5px 10px; border-radius:5px; font-size:12px; cursor:pointer; }
.toggle     { border:none; padding:4px 10px; border-radius:12px; font-size:11px; font-weight:600; cursor:pointer; }
.toggle.on  { background:rgba(74,222,128,.12); color:#4ade80; }
.toggle.off { background:rgba(248,113,113,.12); color:#f87171; }

.role-badge         { display:inline-block; padding:3px 9px; border-radius:10px; font-size:11px; font-weight:600; }
.role-badge.admin   { background:rgba(251,191,36,.12); color:#fbbf24; }
.role-badge.manager { background:rgba(167,139,250,.12); color:#a78bfa; }
.role-badge.staff   { background:rgba(56,189,248,.12); color:#38bdf8; }
.role-badge.customer { background:rgba(148,163,184,.12); color:#94a3b8; }

.pagination { display:flex; align-items:center; justify-content:center; gap:16px; padding:14px; border-top:1px solid #253047; }
.pagination button { background:#253047; border:none; color:#8494ac; padding:6px 12px; border-radius:6px; cursor:pointer; }
.pagination button:disabled { opacity:.4; cursor:not-allowed; }
.pagination span { font-size:13px; color:#5a6a87; }

.btn-primary { background:#38bdf8; color:#080c14; border:none; padding:9px 20px; border-radius:7px; font-size:13px; font-weight:700; cursor:pointer; }
.btn-primary:disabled { opacity:.5; cursor:not-allowed; }
.btn-ghost   { background:none; border:1px solid #2d3a52; color:#8494ac; padding:9px 16px; border-radius:7px; font-size:13px; cursor:pointer; }

.modal-overlay { position:fixed; inset:0; background:rgba(0,0,0,.7); display:flex; align-items:center; justify-content:flex-end; z-index:100; }
.drawer        { background:#1c2333; border-left:1px solid #2d3a52; width:480px; max-width:100vw; height:100vh; overflow-y:auto; display:flex; flex-direction:column; }
.drawer-header { display:flex; align-items:center; justify-content:space-between; padding:24px 24px 0; }
.drawer-header h3 { font-size:16px; color:#eaf0f7; margin:0; }
.modal-close   { background:none; border:none; color:#5a6a87; font-size:18px; cursor:pointer; }
.drawer-body   { padding:20px 24px; display:flex; flex-direction:column; gap:16px; flex:1; }
.drawer-footer { display:flex; gap:10px; justify-content:flex-end; margin-top:auto; padding-top:16px; border-top:1px solid #253047; }
.field         { display:flex; flex-direction:column; gap:6px; flex:1; }
.field label   { font-size:11px; color:#8494ac; text-transform:uppercase; letter-spacing:.5px; }
.field input, .field .sel { background:#0f1623; border:1px solid #2d3a52; color:#d6dfe8; padding:9px 12px; border-radius:7px; font-size:13px; width:100%; }
.field input:focus { outline:none; border-color:#38bdf8; }
.field input:disabled { opacity:.5; }
.field-row     { display:flex; gap:12px; }
.err-msg       { color:#f87171; font-size:12px; }
</style>
