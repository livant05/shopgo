<template>
  <div class="page">
    <h1 class="page-title">Dashboard</h1>
    <p class="page-sub">{{ storeName }} — Resumen de operaciones</p>

    <!-- KPIs -->
    <div class="kpi-grid">
      <div v-for="k in kpis" :key="k.label" class="kpi-card">
        <div class="kpi-icon">{{ k.icon }}</div>
        <div class="kpi-value">{{ k.value }}</div>
        <div class="kpi-label">{{ k.label }}</div>
        <div class="kpi-delta" :class="k.up ? 'up' : 'down'">
          {{ k.up ? '↑' : '↓' }} {{ k.delta }} vs ayer
        </div>
      </div>
    </div>

    <!-- Alertas de stock bajo -->
    <div class="alerts-section" v-if="lowStock.length > 0">
      <h3 class="section-title">⚠️ Stock Bajo ({{ lowStock.length }})</h3>
      <div class="alert-list">
        <div v-for="a in lowStock" :key="a.product_id" class="alert-row">
          <span class="alert-name">{{ a.product_name }}</span>
          <span class="alert-branch">{{ a.branch_name }}</span>
          <span class="alert-stock" :class="a.available === 0 ? 'zero' : 'low'">
            {{ a.available }} disponibles / {{ a.reorder_point }} mínimo
          </span>
          <router-link to="/inventory" class="alert-link">Reponer →</router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '../api/client'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const storeName = import.meta.env.VITE_STORE_NAME ?? 'Mi Tienda'

const kpis = ref([
  { icon:'💰', label:'Ventas hoy',        value:'$0',   delta:'0%',  up:false },
  { icon:'🛒', label:'Órdenes hoy',       value:'0',    delta:'0',   up:false },
  { icon:'📦', label:'Productos activos', value:'0',    delta:'0',   up:true  },
  { icon:'⚠️', label:'Alertas stock',     value:'0',    delta:'0',   up:false },
])
const lowStock = ref<any[]>([])
const loading  = ref(true)

async function load() {
  loading.value = true
  try {
    const [dashRes, alertsRes] = await Promise.all([
      api.get('/admin/reports/dashboard'),
      api.get('/ops/inventory/alerts'),
    ])
    const d = dashRes.data
    kpis.value[0].value = `$${(d.gmv ?? 0).toLocaleString('es-MX', {minimumFractionDigits:0})}`
    kpis.value[1].value = String(d.orders ?? 0)
    lowStock.value = alertsRes.data.data ?? []
    kpis.value[3].value = String(lowStock.value.length)
  } catch (e) { console.error(e) }
  finally { loading.value = false }
}

onMounted(load)
</script>

<style scoped>
.page       { display: flex; flex-direction: column; gap: 28px; }
.page-title { font-size: 22px; font-weight: 700; color: #eaf0f7; margin: 0; }
.page-sub   { font-size: 13px; color: #5a6a87; margin: 0; }

.kpi-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px,1fr)); gap: 16px; }
.kpi-card { background: #1c2333; border: 1px solid #2d3a52; border-radius: 10px; padding: 20px; transition: all .2s; }
.kpi-card:hover { border-color: rgba(79,195,247,.3); transform: translateY(-2px); }
.kpi-icon  { font-size: 24px; margin-bottom: 10px; }
.kpi-value { font-size: 26px; font-weight: 700; color: #eaf0f7; margin-bottom: 4px; }
.kpi-label { font-size: 12px; color: #5a6a87; text-transform: uppercase; letter-spacing: 1px; }
.kpi-delta { font-size: 12px; margin-top: 8px; }
.kpi-delta.up   { color: #4ade80; }
.kpi-delta.down { color: #f87171; }

.section-title { font-size: 15px; font-weight: 600; color: #d6dfe8; margin: 0 0 12px; }
.alert-list { display: flex; flex-direction: column; gap: 8px; }
.alert-row  { display: flex; align-items: center; gap: 16px; background: rgba(251,146,60,.06); border: 1px solid rgba(251,146,60,.2); border-radius: 8px; padding: 11px 14px; }
.alert-name   { flex: 1; font-size: 13px; color: #d6dfe8; font-weight: 500; }
.alert-branch { font-size: 12px; color: #5a6a87; }
.alert-stock  { font-size: 12px; font-weight: 600; }
.alert-stock.zero { color: #f87171; }
.alert-stock.low  { color: #fb923c; }
.alert-link { font-size: 12px; color: #4fc3f7; text-decoration: none; white-space: nowrap; }
</style>
