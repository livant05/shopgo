<template>
  <div class="page">
    <div class="container">

      <div class="page-header">
        <h1 class="page-title">📄 Mis cotizaciones</h1>
        <router-link to="/catalog" class="btn-back">← Volver al catálogo</router-link>
      </div>

      <!-- Empty state -->
      <div v-if="store.history.length === 0" class="empty">
        <p class="empty-icon">📄</p>
        <p class="empty-title">No tienes cotizaciones guardadas</p>
        <p class="empty-sub">Agrega productos a tu cotización desde el catálogo.</p>
        <router-link to="/catalog" class="btn-explore">Explorar catálogo</router-link>
      </div>

      <!-- Quote list -->
      <div v-else class="quote-list">
        <div v-for="item in enriched" :key="item.id" class="quote-card">
          <div class="card-left">
            <p class="card-num">N.° {{ String(item.quoteNumber).padStart(5, '0') }}</p>
            <p class="card-name">{{ item.customerName }}</p>
            <p class="card-date">{{ fmtDate(item.createdAt) }}</p>
          </div>

          <div class="card-mid">
            <p class="card-total">{{ fmt(item.total) }}</p>
            <span v-if="item.liveStatus" class="status-badge" :class="'badge-' + item.liveStatus">
              {{ statusLabel(item.liveStatus) }}
            </span>
            <span v-else class="status-badge badge-loading">…</span>
          </div>

          <router-link :to="`/quote/${item.id}`" class="btn-view">Ver →</router-link>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useQuoteStore } from '../stores/quote'
import { api } from '../api/client'

interface EnrichedItem {
  id: string
  quoteNumber: number
  total: number
  createdAt: string
  customerName: string
  liveStatus: string | null
}

const store    = useQuoteStore()
const enriched = ref<EnrichedItem[]>(
  store.history.map(h => ({ ...h, liveStatus: null }))
)

function fmt(v: number) {
  return `B/. ${v.toLocaleString('es-PA', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

function fmtDate(d: string) {
  return new Date(d).toLocaleDateString('es-PA', { day: '2-digit', month: 'short', year: 'numeric' })
}

function statusLabel(s: string) {
  if (s === 'accepted') return 'Aprobada'
  if (s === 'rejected') return 'No aprobada'
  return 'En revisión'
}

onMounted(async () => {
  await Promise.allSettled(
    enriched.value.map(async (item, i) => {
      try {
        const { data } = await api.get(`/quotes/${item.id}`)
        enriched.value[i].liveStatus = data.status ?? 'pending'
      } catch {
        enriched.value[i].liveStatus = 'pending'
      }
    })
  )
})
</script>

<style scoped>
.page { min-height: 80vh; background: #f1f5f9; padding: 2rem 1.5rem; }
.container { max-width: 720px; margin: 0 auto; }

.page-header {
  display: flex; justify-content: space-between; align-items: center;
  margin-bottom: 1.75rem; flex-wrap: wrap; gap: .75rem;
}
.page-title { font-size: 1.5rem; font-weight: 800; color: #0f172a; margin: 0; }
.btn-back { color: #3b82f6; text-decoration: none; font-size: .875rem; font-weight: 600; }
.btn-back:hover { text-decoration: underline; }

/* Empty */
.empty {
  text-align: center; padding: 4rem 2rem;
  background: #fff; border-radius: 1rem; box-shadow: 0 2px 8px rgba(0,0,0,.06);
}
.empty-icon  { font-size: 3rem; margin: 0 0 .75rem; }
.empty-title { font-size: 1.1rem; font-weight: 700; color: #0f172a; margin: 0 0 .35rem; }
.empty-sub   { font-size: .875rem; color: #64748b; margin: 0 0 1.5rem; }
.btn-explore {
  display: inline-block; background: #1d4ed8; color: #fff;
  text-decoration: none; border-radius: .5rem; padding: .6rem 1.5rem; font-weight: 700;
}

/* List */
.quote-list { display: flex; flex-direction: column; gap: .75rem; }
.quote-card {
  background: #fff; border-radius: .875rem; padding: 1.1rem 1.25rem;
  display: flex; align-items: center; gap: 1rem;
  box-shadow: 0 1px 4px rgba(0,0,0,.06);
  transition: box-shadow .15s;
}
.quote-card:hover { box-shadow: 0 4px 16px rgba(0,0,0,.1); }

.card-left { flex: 1; min-width: 0; }
.card-num  { font-size: .75rem; font-weight: 700; color: #94a3b8; font-family: monospace; margin: 0 0 .2rem; letter-spacing: .05em; }
.card-name { font-size: .95rem; font-weight: 700; color: #1e293b; margin: 0 0 .15rem; }
.card-date { font-size: .78rem; color: #94a3b8; margin: 0; }

.card-mid { display: flex; flex-direction: column; align-items: flex-end; gap: .4rem; flex-shrink: 0; }
.card-total { font-size: 1rem; font-weight: 800; color: #0f172a; margin: 0; font-variant-numeric: tabular-nums; }

.status-badge {
  font-size: .72rem; font-weight: 700; border-radius: 999px;
  padding: .2rem .7rem; white-space: nowrap;
}
.badge-pending  { background: #fef3c7; color: #92400e; }
.badge-accepted { background: #dcfce7; color: #166534; }
.badge-rejected { background: #fee2e2; color: #991b1b; }
.badge-loading  { background: #f1f5f9; color: #94a3b8; }

.btn-view {
  background: #1d4ed8; color: #fff; text-decoration: none;
  border-radius: .5rem; padding: .5rem 1rem; font-size: .85rem;
  font-weight: 700; flex-shrink: 0; white-space: nowrap;
}
.btn-view:hover { background: #1e40af; }

@media (max-width: 600px) {
  .quote-card { flex-wrap: wrap; }
  .card-mid { flex-direction: row; align-items: center; }
}
</style>
