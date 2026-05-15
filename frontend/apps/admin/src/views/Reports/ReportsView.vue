<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">Reportes</h1>
        <p class="page-sub">Análisis de ventas e inventario</p>
      </div>
      <div class="date-filters">
        <input v-model="from" type="date" class="date-input" @change="loadAll()" />
        <span class="date-sep">→</span>
        <input v-model="to"   type="date" class="date-input" @change="loadAll()" />
        <button class="btn-outline" @click="setToday()">Hoy</button>
        <button class="btn-outline" @click="setMonth()">Este mes</button>
      </div>
    </div>

    <!-- KPI cards -->
    <div class="kpi-grid">
      <div class="kpi-card">
        <div class="kpi-icon">💰</div>
        <div class="kpi-val">${{ fmt(revenue.gmv) }}</div>
        <div class="kpi-lbl">Ventas (GMV)</div>
      </div>
      <div class="kpi-card">
        <div class="kpi-icon">🛒</div>
        <div class="kpi-val">{{ revenue.orders ?? 0 }}</div>
        <div class="kpi-lbl">Órdenes</div>
      </div>
      <div class="kpi-card">
        <div class="kpi-icon">📊</div>
        <div class="kpi-val">${{ fmt(revenue.aov) }}</div>
        <div class="kpi-lbl">Ticket promedio</div>
      </div>
      <div class="kpi-card">
        <div class="kpi-icon">👥</div>
        <div class="kpi-val">{{ revenue.customers ?? 0 }}</div>
        <div class="kpi-lbl">Clientes únicos</div>
      </div>
    </div>

    <!-- Gráfica de serie diaria -->
    <div class="section" v-if="daily.length > 0">
      <h3 class="section-title">Ventas diarias</h3>
      <div class="chart">
        <div v-for="d in daily" :key="d.day" class="bar-col">
          <div class="bar-tip">${{ fmtK(d.revenue) }}</div>
          <div class="bar" :style="{ height: barH(d.revenue) + 'px' }" :title="`${d.day}: $${fmt(d.revenue)}`"></div>
          <div class="bar-lbl">{{ d.day.slice(5) }}</div>
        </div>
      </div>
    </div>

    <!-- Ventas por sucursal -->
    <div class="section" v-if="byBranch.length > 0">
      <h3 class="section-title">Ventas por sucursal</h3>
      <table class="tbl">
        <thead><tr><th>Sucursal</th><th>Órdenes</th><th>Ingresos</th><th>Clientes</th></tr></thead>
        <tbody>
          <tr v-for="b in byBranch" :key="b.branch_id" class="tbl-row">
            <td class="td-bold">{{ b.branch_name }}</td>
            <td>{{ b.orders }}</td>
            <td class="td-amount">${{ fmt(b.revenue) }}</td>
            <td>{{ b.customers }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Top productos -->
    <div class="section" v-if="topProds.length > 0">
      <h3 class="section-title">Top productos</h3>
      <table class="tbl">
        <thead><tr><th>#</th><th>SKU</th><th>Nombre</th><th>Unidades</th><th>Ingresos</th></tr></thead>
        <tbody>
          <tr v-for="(p, i) in topProds" :key="p.id" class="tbl-row">
            <td class="td-muted">{{ i + 1 }}</td>
            <td class="mono">{{ p.sku }}</td>
            <td>{{ p.name }}</td>
            <td class="td-bold">{{ p.units_sold }}</td>
            <td class="td-amount">${{ fmt(p.revenue) }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="loading" class="loading">Cargando reportes…</div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '../../api/client'

const from = ref('')
const to   = ref('')
const loading = ref(false)
const revenue  = ref<any>({ gmv:0, orders:0, aov:0, customers:0 })
const byBranch = ref<any[]>([])
const topProds = ref<any[]>([])
const daily    = ref<any[]>([])

const maxRev = computed(() => Math.max(...daily.value.map((d: any) => d.revenue), 1))
const barH   = (rev: number) => Math.max(4, Math.round((rev / maxRev.value) * 140))

function setToday() {
  const d = new Date().toISOString().slice(0,10)
  from.value = d; to.value = d; loadAll()
}
function setMonth() {
  const now = new Date()
  from.value = new Date(now.getFullYear(), now.getMonth(), 1).toISOString().slice(0,10)
  to.value   = now.toISOString().slice(0,10)
  loadAll()
}

async function loadAll() {
  loading.value = true
  const p = new URLSearchParams()
  if (from.value) p.set('from', from.value)
  if (to.value)   p.set('to', to.value)
  const qs = p.toString() ? '?' + p : ''
  try {
    const [r, b, t, d] = await Promise.all([
      api.get('/admin/reports/dashboard' + qs),
      api.get('/admin/reports/branches'  + qs),
      api.get('/admin/reports/products'  + qs),
      api.get('/admin/reports/sales'     + qs),
    ])
    revenue.value  = r.data
    byBranch.value = b.data.data ?? []
    topProds.value = t.data.data ?? []
    daily.value    = d.data.data ?? []
  } catch(e) { console.error(e) }
  finally { loading.value = false }
}

const fmt  = (n: number) => (n ?? 0).toLocaleString('es-MX', { minimumFractionDigits: 2 })
const fmtK = (n: number) => n >= 1000 ? `${(n/1000).toFixed(1)}k` : String(Math.round(n))

onMounted(() => { setMonth() })
</script>

<style scoped>
.page        { display:flex; flex-direction:column; gap:28px; }
.page-header { display:flex; align-items:flex-start; justify-content:space-between; flex-wrap:wrap; gap:12px; }
.page-title  { font-size:22px; font-weight:700; color:#eaf0f7; margin:0; }
.page-sub    { font-size:13px; color:#5a6a87; margin:4px 0 0; }
.date-filters { display:flex; align-items:center; gap:8px; flex-wrap:wrap; }
.date-input   { background:#1c2333; border:1px solid #2d3a52; color:#d6dfe8; padding:8px 10px; border-radius:7px; font-size:13px; }
.date-sep     { color:#5a6a87; }
.btn-outline  { background:none; border:1px solid #38bdf8; color:#38bdf8; padding:8px 14px; border-radius:7px; font-size:12px; cursor:pointer; }

.kpi-grid    { display:grid; grid-template-columns:repeat(auto-fill, minmax(180px,1fr)); gap:14px; }
.kpi-card    { background:#1c2333; border:1px solid #2d3a52; border-radius:10px; padding:20px; }
.kpi-icon    { font-size:22px; margin-bottom:8px; }
.kpi-val     { font-size:28px; font-weight:700; color:#eaf0f7; margin-bottom:4px; }
.kpi-lbl     { font-size:12px; color:#5a6a87; text-transform:uppercase; letter-spacing:.5px; }

.section       { background:#1c2333; border:1px solid #2d3a52; border-radius:10px; padding:20px; }
.section-title { font-size:14px; font-weight:700; color:#d6dfe8; margin:0 0 16px; text-transform:uppercase; letter-spacing:.5px; }

.chart   { display:flex; align-items:flex-end; gap:4px; height:160px; overflow-x:auto; padding-bottom:4px; }
.bar-col { display:flex; flex-direction:column; align-items:center; gap:4px; min-width:40px; }
.bar     { width:28px; background:linear-gradient(to top,#0ea5e9,#38bdf8); border-radius:4px 4px 0 0; transition:height .3s; }
.bar-tip { font-size:10px; color:#5a6a87; white-space:nowrap; }
.bar-lbl { font-size:10px; color:#5a6a87; white-space:nowrap; }

.tbl        { width:100%; border-collapse:collapse; font-size:13px; }
.tbl thead th { color:#8494ac; padding:8px 12px; text-align:left; font-weight:600; font-size:11px; text-transform:uppercase; letter-spacing:.5px; border-bottom:1px solid #2d3a52; }
.tbl-row    { border-top:1px solid rgba(45,58,82,.5); }
.tbl-row td { padding:10px 12px; color:#d6dfe8; }
.td-bold    { font-weight:600; color:#eaf0f7; }
.td-muted   { color:#5a6a87; }
.td-amount  { color:#4ade80; font-weight:600; }
.mono       { font-family:monospace; font-size:12px; color:#8494ac; }
.loading    { text-align:center; color:#5a6a87; padding:20px; }
</style>
