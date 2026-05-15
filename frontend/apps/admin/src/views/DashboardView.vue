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

    <!-- Gráfica de ventas -->
    <div class="chart-card">
      <div class="chart-header">
        <h3 class="section-title">Ventas — últimos 30 días</h3>
        <div class="chart-legend">
          <span class="legend-dot revenue"></span><span class="legend-text">Ingresos (MXN)</span>
          <span class="legend-dot orders"></span><span class="legend-text">Órdenes</span>
        </div>
      </div>
      <div v-if="chartLoading" class="chart-placeholder">Cargando datos…</div>
      <v-chart v-else class="chart" :option="chartOption" autoresize />
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
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, BarChart } from 'echarts/charts'
import {
  GridComponent, TooltipComponent, LegendComponent,
  DataZoomComponent,
} from 'echarts/components'
import VChart from 'vue-echarts'
import { api } from '../api/client'
import { useAuthStore } from '../stores/auth'

use([CanvasRenderer, LineChart, BarChart, GridComponent, TooltipComponent, LegendComponent, DataZoomComponent])

const auth = useAuthStore()
const storeName = import.meta.env.VITE_STORE_NAME ?? 'Mi Tienda'

const kpis = ref([
  { icon:'💰', label:'Ventas hoy',        value:'$0',   delta:'0%',  up:false },
  { icon:'🛒', label:'Órdenes hoy',       value:'0',    delta:'0',   up:false },
  { icon:'📦', label:'Productos activos', value:'0',    delta:'0',   up:true  },
  { icon:'⚠️', label:'Alertas stock',     value:'0',    delta:'0',   up:false },
])
const lowStock   = ref<any[]>([])
const chartLoading = ref(true)
const chartOption  = ref<any>({})

function buildChart(days: { day: string; orders: number; revenue: number }[]) {
  const labels   = days.map(d => d.day)
  const revenues = days.map(d => d.revenue)
  const orders   = days.map(d => d.orders)

  chartOption.value = {
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'axis',
      backgroundColor: '#1c2333',
      borderColor: '#2d3a52',
      textStyle: { color: '#d6dfe8', fontSize: 12 },
      formatter(params: any[]) {
        const rev = params.find(p => p.seriesName === 'Ingresos')
        const ord = params.find(p => p.seriesName === 'Órdenes')
        return `<b>${params[0].axisValue}</b><br/>
          💰 Ingresos: <b>$${Number(rev?.value ?? 0).toLocaleString('es-MX', { minimumFractionDigits: 2 })}</b><br/>
          🛒 Órdenes: <b>${ord?.value ?? 0}</b>`
      },
    },
    legend: { show: false },
    grid: { left: 60, right: 24, top: 16, bottom: 40 },
    xAxis: {
      type: 'category',
      data: labels,
      axisLine:  { lineStyle: { color: '#2d3a52' } },
      axisLabel: { color: '#5a6a87', fontSize: 11 },
      splitLine: { show: false },
    },
    yAxis: [
      {
        type: 'value',
        name: 'MXN',
        nameTextStyle: { color: '#5a6a87', fontSize: 10 },
        axisLine:  { show: false },
        axisLabel: { color: '#5a6a87', fontSize: 11,
          formatter: (v: number) => v >= 1000 ? `$${(v/1000).toFixed(0)}k` : `$${v}` },
        splitLine: { lineStyle: { color: '#1a2235' } },
      },
      {
        type: 'value',
        name: 'Órdenes',
        nameTextStyle: { color: '#5a6a87', fontSize: 10 },
        axisLine:  { show: false },
        axisLabel: { color: '#5a6a87', fontSize: 11 },
        splitLine: { show: false },
      },
    ],
    series: [
      {
        name: 'Ingresos',
        type: 'line',
        data: revenues,
        smooth: true,
        symbol: 'none',
        lineStyle: { color: '#38bdf8', width: 2 },
        areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
          colorStops: [{ offset: 0, color: 'rgba(56,189,248,.25)' }, { offset: 1, color: 'rgba(56,189,248,.02)' }] } },
        yAxisIndex: 0,
      },
      {
        name: 'Órdenes',
        type: 'bar',
        data: orders,
        barMaxWidth: 12,
        itemStyle: { color: 'rgba(99,102,241,.6)', borderRadius: [3, 3, 0, 0] },
        yAxisIndex: 1,
      },
    ],
  }
}

async function loadChart() {
  chartLoading.value = true
  try {
    const to   = new Date()
    const from = new Date(to)
    from.setDate(from.getDate() - 29)
    const fmt = (d: Date) => d.toISOString().slice(0, 10)
    const res = await api.get('/admin/reports/sales', { params: { from: fmt(from), to: fmt(to) } })
    buildChart(res.data.data ?? [])
  } catch (e) {
    console.error('[chart]', e)
    buildChart([])
  } finally {
    chartLoading.value = false
  }
}

async function loadDashboard() {
  try {
    const [dashRes, alertsRes] = await Promise.all([
      api.get('/admin/reports/dashboard'),
      api.get('/ops/inventory/alerts'),
    ])
    const d = dashRes.data
    kpis.value[0].value = `$${(d.gmv ?? 0).toLocaleString('es-MX', { minimumFractionDigits: 0 })}`
    kpis.value[1].value = String(d.orders ?? 0)
    lowStock.value = alertsRes.data.data ?? []
    kpis.value[3].value = String(lowStock.value.length)
  } catch (e) { console.error(e) }
}

onMounted(() => {
  loadDashboard()
  loadChart()
})
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

.chart-card {
  background: #1c2333; border: 1px solid #2d3a52; border-radius: 10px; padding: 20px;
}
.chart-header {
  display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px;
}
.chart-legend {
  display: flex; align-items: center; gap: 14px;
}
.legend-dot {
  width: 10px; height: 10px; border-radius: 50%; display: inline-block;
}
.legend-dot.revenue { background: #38bdf8; }
.legend-dot.orders  { background: rgba(99,102,241,.8); }
.legend-text { font-size: 12px; color: #5a6a87; }
.chart { height: 260px; width: 100%; }
.chart-placeholder { height: 260px; display: flex; align-items: center; justify-content: center; color: #5a6a87; font-size: 13px; }

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
