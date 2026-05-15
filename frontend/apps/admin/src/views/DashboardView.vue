<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">Dashboard</h1>
        <p class="page-sub">{{ storeName }} — últimos 30 días</p>
      </div>
      <div class="period-badge">📅 {{ fmtShort(dateFrom) }} → {{ fmtShort(dateTo) }}</div>
    </div>

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

    <!-- Gráfica de ventas diarias -->
    <div class="chart-card">
      <div class="chart-header">
        <h3 class="section-title">Ventas diarias</h3>
        <div class="chart-legend">
          <span class="legend-dot revenue"></span><span class="legend-text">Ingresos (MXN)</span>
          <span class="legend-dot orders"></span><span class="legend-text">Órdenes</span>
        </div>
      </div>
      <div v-if="chartLoading" class="chart-placeholder">Cargando datos…</div>
      <v-chart v-else class="chart" :option="chartOption" autoresize />
    </div>

    <!-- Row: Top Productos + Top Clientes -->
    <div class="two-col">

      <!-- Top Productos -->
      <div class="panel-card">
        <div class="panel-header">
          <h3 class="section-title">🏆 Top productos</h3>
          <span class="panel-sub">Por unidades vendidas</span>
        </div>
        <div v-if="topProductsLoading" class="panel-loading">Cargando…</div>
        <div v-else-if="topProducts.length === 0" class="panel-empty">Sin datos</div>
        <div v-else class="rank-list">
          <div v-for="(p, idx) in topProducts" :key="p.id" class="rank-row">
            <span class="rank-num" :class="'rank-'+( idx+1 <= 3 ? idx+1 : 'rest')">{{ idx + 1 }}</span>
            <div class="rank-info">
              <div class="rank-name">{{ p.name }}</div>
              <div class="rank-meta">{{ p.sku }}</div>
            </div>
            <div class="rank-stats">
              <div class="rank-primary">{{ p.units_sold }} uds.</div>
              <div class="rank-secondary">{{ fmtMXN(p.revenue) }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- Top Clientes -->
      <div class="panel-card">
        <div class="panel-header">
          <h3 class="section-title">👥 Top clientes</h3>
          <span class="panel-sub">Por gasto total</span>
        </div>
        <div v-if="topCustomersLoading" class="panel-loading">Cargando…</div>
        <div v-else-if="topCustomers.length === 0" class="panel-empty">Sin datos</div>
        <div v-else class="rank-list">
          <div v-for="(c, idx) in topCustomers" :key="c.customer_id" class="rank-row">
            <span class="rank-num" :class="'rank-'+( idx+1 <= 3 ? idx+1 : 'rest')">{{ idx + 1 }}</span>
            <div class="rank-info">
              <div class="rank-name">{{ c.full_name || 'Cliente' }}</div>
              <div class="rank-meta">{{ c.email || c.customer_id.slice(0,12) + '…' }}</div>
            </div>
            <div class="rank-stats">
              <div class="rank-primary">{{ fmtMXN(c.revenue) }}</div>
              <div class="rank-secondary">{{ c.orders }} pedidos</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Horas pico -->
    <div class="chart-card">
      <div class="chart-header">
        <h3 class="section-title">⏰ Horas pico de ventas</h3>
        <span class="panel-sub">Distribución horaria de órdenes</span>
      </div>
      <div v-if="hourlyLoading" class="chart-placeholder">Cargando datos…</div>
      <v-chart v-else class="chart chart--hourly" :option="hourlyOption" autoresize />
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
import { ref, onMounted } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, BarChart } from 'echarts/charts'
import {
  GridComponent, TooltipComponent, LegendComponent, DataZoomComponent,
} from 'echarts/components'
import VChart from 'vue-echarts'
import { api } from '../api/client'

use([CanvasRenderer, LineChart, BarChart, GridComponent, TooltipComponent, LegendComponent, DataZoomComponent])

const storeName = import.meta.env.VITE_STORE_NAME ?? 'Mi Tienda'

// Date range: last 30 days
const dateTo   = new Date()
const dateFrom = new Date(dateTo); dateFrom.setDate(dateFrom.getDate() - 29)
const fmtISO   = (d: Date) => d.toISOString().slice(0, 10)
const fmtShort = (d: Date) => d.toLocaleDateString('es-MX', { day: 'numeric', month: 'short' })

const kpis = ref([
  { icon: '💰', label: 'Ventas hoy',        value: '$0', delta: '0%', up: false },
  { icon: '🛒', label: 'Órdenes hoy',       value: '0',  delta: '0',  up: false },
  { icon: '📦', label: 'Productos activos', value: '0',  delta: '0',  up: true  },
  { icon: '⚠️', label: 'Alertas stock',    value: '0',  delta: '0',  up: false },
])
const lowStock   = ref<any[]>([])

// Sales chart
const chartLoading = ref(true)
const chartOption  = ref<any>({})

// Top products
const topProducts        = ref<any[]>([])
const topProductsLoading = ref(true)

// Top customers
const topCustomers        = ref<any[]>([])
const topCustomersLoading = ref(true)

// Hourly
const hourlyLoading = ref(true)
const hourlyOption  = ref<any>({})

// ── Formatters ────────────────────────────────────────────────────────────────
const fmtMXN = (n: number) =>
  (n ?? 0).toLocaleString('es-MX', { style: 'currency', currency: 'MXN', maximumFractionDigits: 0 })

// ── Charts ────────────────────────────────────────────────────────────────────
function buildSalesChart(days: { day: string; orders: number; revenue: number }[]) {
  chartOption.value = {
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'axis',
      backgroundColor: '#1c2333', borderColor: '#2d3a52',
      textStyle: { color: '#d6dfe8', fontSize: 12 },
      formatter(params: any[]) {
        const rev = params.find(p => p.seriesName === 'Ingresos')
        const ord = params.find(p => p.seriesName === 'Órdenes')
        return `<b>${params[0].axisValue}</b><br/>
          💰 Ingresos: <b>${fmtMXN(rev?.value ?? 0)}</b><br/>
          🛒 Órdenes: <b>${ord?.value ?? 0}</b>`
      },
    },
    legend: { show: false },
    grid: { left: 60, right: 24, top: 16, bottom: 40 },
    xAxis: {
      type: 'category', data: days.map(d => d.day),
      axisLine: { lineStyle: { color: '#2d3a52' } },
      axisLabel: { color: '#5a6a87', fontSize: 11 },
      splitLine: { show: false },
    },
    yAxis: [
      {
        type: 'value', name: 'MXN',
        nameTextStyle: { color: '#5a6a87', fontSize: 10 },
        axisLine: { show: false },
        axisLabel: { color: '#5a6a87', fontSize: 11,
          formatter: (v: number) => v >= 1000 ? `$${(v/1000).toFixed(0)}k` : `$${v}` },
        splitLine: { lineStyle: { color: '#1a2235' } },
      },
      {
        type: 'value', name: 'Órdenes',
        nameTextStyle: { color: '#5a6a87', fontSize: 10 },
        axisLine: { show: false },
        axisLabel: { color: '#5a6a87', fontSize: 11 },
        splitLine: { show: false },
      },
    ],
    series: [
      {
        name: 'Ingresos', type: 'line', data: days.map(d => d.revenue),
        smooth: true, symbol: 'none',
        lineStyle: { color: '#38bdf8', width: 2 },
        areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
          colorStops: [{ offset: 0, color: 'rgba(56,189,248,.25)' }, { offset: 1, color: 'rgba(56,189,248,.02)' }] } },
        yAxisIndex: 0,
      },
      {
        name: 'Órdenes', type: 'bar', data: days.map(d => d.orders),
        barMaxWidth: 12,
        itemStyle: { color: 'rgba(99,102,241,.6)', borderRadius: [3, 3, 0, 0] },
        yAxisIndex: 1,
      },
    ],
  }
}

function buildHourlyChart(raw: { hour: number; orders: number; revenue: number }[]) {
  // Fill 0-23, even if no data
  const byHour = new Map(raw.map(r => [r.hour, r]))
  const hours  = Array.from({ length: 24 }, (_, h) => h)
  const orders  = hours.map(h => byHour.get(h)?.orders ?? 0)
  const maxOrd  = Math.max(...orders, 1)

  hourlyOption.value = {
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'axis',
      backgroundColor: '#1c2333', borderColor: '#2d3a52',
      textStyle: { color: '#d6dfe8', fontSize: 12 },
      formatter(params: any[]) {
        const h   = Number(params[0].axisValue)
        const ord = params[0].value
        const rev = byHour.get(h)?.revenue ?? 0
        return `<b>${h}:00 – ${h + 1}:00</b><br/>🛒 Órdenes: <b>${ord}</b><br/>💰 ${fmtMXN(rev)}`
      },
    },
    grid: { left: 36, right: 16, top: 16, bottom: 36 },
    xAxis: {
      type: 'category',
      data: hours.map(h => `${h}h`),
      axisLine: { lineStyle: { color: '#2d3a52' } },
      axisLabel: { color: '#5a6a87', fontSize: 10 },
      splitLine: { show: false },
    },
    yAxis: {
      type: 'value',
      axisLine: { show: false },
      axisLabel: { color: '#5a6a87', fontSize: 11 },
      splitLine: { lineStyle: { color: '#1a2235' } },
    },
    series: [{
      type: 'bar',
      data: orders.map((v, h) => ({
        value: v,
        itemStyle: {
          color: v === 0 ? '#1a2235' :
            v >= maxOrd * .8 ? '#f59e0b' :
            v >= maxOrd * .5 ? '#38bdf8' : '#3d5a7a',
          borderRadius: [3, 3, 0, 0],
        },
      })),
      barCategoryGap: '30%',
    }],
  }
}

// ── Data loading ──────────────────────────────────────────────────────────────
async function loadDashboard() {
  try {
    const [dashRes, alertsRes] = await Promise.all([
      api.get('/admin/reports/dashboard'),
      api.get('/ops/inventory/alerts'),
    ])
    const d = dashRes.data
    kpis.value[0].value = fmtMXN(d.gmv ?? 0)
    kpis.value[1].value = String(d.orders ?? 0)
    lowStock.value = alertsRes.data.data ?? []
    kpis.value[3].value = String(lowStock.value.length)
  } catch (e) { console.error(e) }
}

async function loadSalesChart() {
  chartLoading.value = true
  try {
    const res = await api.get('/admin/reports/sales', {
      params: { from: fmtISO(dateFrom), to: fmtISO(dateTo) },
    })
    buildSalesChart(res.data.data ?? [])
  } catch { buildSalesChart([]) }
  finally { chartLoading.value = false }
}

async function loadTopProducts() {
  topProductsLoading.value = true
  try {
    const res = await api.get('/admin/reports/products', {
      params: { from: fmtISO(dateFrom), to: fmtISO(dateTo), n: 5 },
    })
    topProducts.value = res.data.data ?? []
  } catch {}
  finally { topProductsLoading.value = false }
}

async function loadTopCustomers() {
  topCustomersLoading.value = true
  try {
    const res = await api.get('/admin/reports/customers', {
      params: { from: fmtISO(dateFrom), to: fmtISO(dateTo), n: 5 },
    })
    topCustomers.value = res.data.data ?? []
  } catch {}
  finally { topCustomersLoading.value = false }
}

async function loadHourly() {
  hourlyLoading.value = true
  try {
    const res = await api.get('/admin/reports/hourly', {
      params: { from: fmtISO(dateFrom), to: fmtISO(dateTo) },
    })
    buildHourlyChart(res.data.data ?? [])
  } catch { buildHourlyChart([]) }
  finally { hourlyLoading.value = false }
}

onMounted(() => {
  loadDashboard()
  loadSalesChart()
  loadTopProducts()
  loadTopCustomers()
  loadHourly()
})
</script>

<style scoped>
.page        { display: flex; flex-direction: column; gap: 28px; }
.page-header { display: flex; align-items: flex-start; justify-content: space-between; }
.page-title  { font-size: 22px; font-weight: 700; color: #eaf0f7; margin: 0; }
.page-sub    { font-size: 13px; color: #5a6a87; margin: 4px 0 0; }
.period-badge { background: #1c2333; border: 1px solid #2d3a52; color: #5a6a87; font-size: 12px; padding: 6px 12px; border-radius: 8px; white-space: nowrap; align-self: center; }

/* KPIs */
.kpi-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px,1fr)); gap: 16px; }
.kpi-card { background: #1c2333; border: 1px solid #2d3a52; border-radius: 10px; padding: 20px; transition: all .2s; }
.kpi-card:hover { border-color: rgba(79,195,247,.3); transform: translateY(-2px); }
.kpi-icon  { font-size: 24px; margin-bottom: 10px; }
.kpi-value { font-size: 26px; font-weight: 700; color: #eaf0f7; margin-bottom: 4px; }
.kpi-label { font-size: 12px; color: #5a6a87; text-transform: uppercase; letter-spacing: 1px; }
.kpi-delta { font-size: 12px; margin-top: 8px; }
.kpi-delta.up   { color: #4ade80; }
.kpi-delta.down { color: #f87171; }

/* Charts */
.chart-card { background: #1c2333; border: 1px solid #2d3a52; border-radius: 10px; padding: 20px; }
.chart-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; flex-wrap: wrap; gap: 8px; }
.chart-legend { display: flex; align-items: center; gap: 14px; }
.legend-dot { width: 10px; height: 10px; border-radius: 50%; display: inline-block; }
.legend-dot.revenue { background: #38bdf8; }
.legend-dot.orders  { background: rgba(99,102,241,.8); }
.legend-text { font-size: 12px; color: #5a6a87; }
.chart         { height: 260px; width: 100%; }
.chart--hourly { height: 200px; }
.chart-placeholder { height: 260px; display: flex; align-items: center; justify-content: center; color: #5a6a87; font-size: 13px; }

/* Two-column layout */
.two-col { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
@media (max-width: 900px) { .two-col { grid-template-columns: 1fr; } }

/* Panel cards */
.panel-card { background: #1c2333; border: 1px solid #2d3a52; border-radius: 10px; padding: 20px; display: flex; flex-direction: column; }
.panel-header { display: flex; align-items: baseline; justify-content: space-between; margin-bottom: 16px; }
.panel-sub  { font-size: 11px; color: #5a6a87; }
.panel-loading, .panel-empty { font-size: 13px; color: #5a6a87; padding: 20px 0; text-align: center; }

/* Rank list */
.rank-list { display: flex; flex-direction: column; gap: 2px; }
.rank-row  { display: flex; align-items: center; gap: 12px; padding: 9px 10px; border-radius: 7px; transition: background .1s; }
.rank-row:hover { background: rgba(56,189,248,.04); }
.rank-num  { width: 24px; height: 24px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 11px; font-weight: 800; flex-shrink: 0; }
.rank-1    { background: rgba(251,191,36,.15); color: #fbbf24; }
.rank-2    { background: rgba(148,163,184,.12); color: #94a3b8; }
.rank-3    { background: rgba(180,120,60,.12);  color: #cd7f3a; }
.rank-rest { background: #1a2235; color: #5a6a87; }
.rank-info { flex: 1; min-width: 0; }
.rank-name { font-size: 13px; color: #d6dfe8; font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.rank-meta { font-size: 11px; color: #5a6a87; font-family: monospace; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.rank-stats { text-align: right; flex-shrink: 0; }
.rank-primary   { font-size: 13px; font-weight: 700; color: #4ade80; }
.rank-secondary { font-size: 11px; color: #5a6a87; }

/* Section title & alerts */
.section-title { font-size: 15px; font-weight: 600; color: #d6dfe8; margin: 0; }
.alert-list { display: flex; flex-direction: column; gap: 8px; margin-top: 12px; }
.alert-row  { display: flex; align-items: center; gap: 16px; background: rgba(251,146,60,.06); border: 1px solid rgba(251,146,60,.2); border-radius: 8px; padding: 11px 14px; }
.alert-name   { flex: 1; font-size: 13px; color: #d6dfe8; font-weight: 500; }
.alert-branch { font-size: 12px; color: #5a6a87; }
.alert-stock  { font-size: 12px; font-weight: 600; }
.alert-stock.zero { color: #f87171; }
.alert-stock.low  { color: #fb923c; }
.alert-link { font-size: 12px; color: #4fc3f7; text-decoration: none; white-space: nowrap; }
</style>
