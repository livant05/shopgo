<template>
  <div class="page">

    <!-- Actions (hidden on print) -->
    <div class="no-print action-bar">
      <router-link to="/catalog" class="btn-back">← Volver al catálogo</router-link>
      <div class="action-btns">
        <button class="btn-wa" v-if="waLink" @click="shareWhatsApp()">
          📱 Compartir WhatsApp
        </button>
        <button class="btn-copy" @click="copyLink()">
          {{ copied ? '✓ Copiado' : '🔗 Copiar enlace' }}
        </button>
        <button class="btn-print" @click="window.print()">
          🖨 Descargar / Imprimir
        </button>
      </div>
    </div>

    <div v-if="loading" class="loading">Cargando cotización…</div>
    <div v-else-if="!quote" class="not-found">Cotización no encontrada.</div>

    <template v-else>
      <!-- Status banner (hidden on print) -->
      <div class="no-print status-banner"
           :class="bannerClass">
        <div class="banner-left">
          <span class="banner-icon">{{ bannerIcon }}</span>
          <div>
            <p class="banner-title">{{ bannerTitle }}</p>
            <p v-if="quote.status_note" class="banner-note">{{ quote.status_note }}</p>
          </div>
        </div>
        <div class="banner-right">
          <span v-if="isExpired" class="tag-expired">Vencida</span>
          <button v-else-if="quote.status === 'accepted'"
                  class="btn-order" @click="loadIntoCart">
            Proceder al pago →
          </button>
        </div>
      </div>

      <div id="quote-doc" class="quote-doc">

        <!-- Header -->
        <div class="doc-header">
          <div class="doc-company">
            <h1 class="company-name">{{ quote.store_name }}</h1>
            <p v-if="quote.contact_email" class="company-contact">📧 {{ quote.contact_email }}</p>
            <p v-if="quote.support_phone" class="company-contact">📞 {{ quote.support_phone }}</p>
          </div>
          <div class="doc-meta">
            <div class="quote-badge">COTIZACIÓN</div>
            <p class="quote-num">N.° {{ String(quote.quote_number).padStart(5, '0') }}</p>
            <p class="quote-date">Fecha: {{ fmtDate(quote.created_at) }}</p>
            <p class="quote-valid">
              Válida hasta: {{ quote.expires_at ? fmtDate(quote.expires_at) : '30 días' }}
            </p>
          </div>
        </div>

        <div class="doc-divider" />

        <!-- Customer info -->
        <div class="doc-customer">
          <div>
            <p class="section-label">COTIZADO A</p>
            <p class="customer-name">{{ quote.customer_name || 'Cliente' }}</p>
            <p v-if="quote.customer_email" class="customer-detail">{{ quote.customer_email }}</p>
            <p v-if="quote.customer_phone" class="customer-detail">{{ quote.customer_phone }}</p>
          </div>
          <div class="currency-badge">
            <p class="currency-label">Moneda</p>
            <p class="currency-val">Balboas / USD (B/.)</p>
          </div>
        </div>

        <!-- Items table -->
        <table class="items-table">
          <thead>
            <tr>
              <th class="col-num">#</th>
              <th>Descripción</th>
              <th class="col-sku">SKU</th>
              <th class="col-num">Cant.</th>
              <th class="col-price">Precio unit.</th>
              <th class="col-price">Subtotal</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(item, i) in quote.items" :key="item.product_id">
              <td class="col-num">{{ i + 1 }}</td>
              <td class="item-name-cell">{{ item.name }}</td>
              <td class="col-sku">{{ item.sku }}</td>
              <td class="col-num">{{ item.qty }}</td>
              <td class="col-price">{{ fmt(item.unit_price) }}</td>
              <td class="col-price">{{ fmt(item.subtotal) }}</td>
            </tr>
          </tbody>
        </table>

        <!-- Totals -->
        <div class="totals-wrap">
          <div class="totals-box">
            <div class="tot-row">
              <span>Subtotal</span>
              <span>{{ fmt(quote.subtotal) }}</span>
            </div>
            <div class="tot-row itbms">
              <span>ITBMS ({{ (quote.tax_rate * 100).toFixed(0) }}%)</span>
              <span>{{ fmt(quote.tax_amount) }}</span>
            </div>
            <div class="tot-row grand">
              <span>TOTAL</span>
              <span>{{ fmt(quote.total) }}</span>
            </div>
          </div>
        </div>

        <!-- Note -->
        <div v-if="quote.note" class="note-box">
          <p class="section-label">NOTAS</p>
          <p class="note-text">{{ quote.note }}</p>
        </div>

        <!-- Footer -->
        <div class="doc-footer">
          <p>Esta cotización es válida hasta el {{ quote.expires_at ? fmtDate(quote.expires_at) : '30 días desde la emisión' }}.</p>
          <p>Los precios indicados están en Balboas (B/.) equivalentes a Dólares Americanos (USD) e incluyen el ITBMS del {{ (quote.tax_rate * 100).toFixed(0) }}% según la legislación panameña.</p>
          <p v-if="quote.contact_email">Para consultas: {{ quote.contact_email }}</p>
        </div>

      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '../api/client'
import { useCartStore } from '../stores/cart'

interface QuoteItem {
  product_id: string
  sku: string
  name: string
  qty: number
  unit_price: number
  subtotal: number
}

interface Quote {
  id: string
  quote_number: number
  items: QuoteItem[]
  subtotal: number
  tax_rate: number
  tax_amount: number
  total: number
  currency: string
  store_name: string
  contact_email: string
  support_phone: string
  customer_name: string
  customer_email: string
  customer_phone: string
  note: string
  status: string
  status_note?: string
  status_at?: string
  created_at: string
  expires_at?: string
}

const route   = useRoute()
const router  = useRouter()
const cart    = useCartStore()
const quote   = ref<Quote | null>(null)
const loading = ref(true)
const copied  = ref(false)

const isExpired = computed(() => {
  if (!quote.value?.expires_at) return false
  return new Date(quote.value.expires_at) < new Date()
})

const bannerClass = computed(() => ({
  'banner-pending':  quote.value?.status === 'pending',
  'banner-accepted': quote.value?.status === 'accepted',
  'banner-rejected': quote.value?.status === 'rejected',
}))

const bannerIcon = computed(() => {
  if (quote.value?.status === 'accepted') return '✅'
  if (quote.value?.status === 'rejected') return '✗'
  return '⏳'
})

const bannerTitle = computed(() => {
  if (quote.value?.status === 'accepted') return 'Cotización aceptada'
  if (quote.value?.status === 'rejected') return 'Cotización no aprobada'
  return 'Cotización en revisión — te notificaremos cuando esté lista'
})

const waLink = computed(() => {
  if (!quote.value) return ''
  const text = encodeURIComponent(
    `Hola, te comparto una cotización de ${quote.value.store_name}:\n${window.location.href}`
  )
  return `https://wa.me/?text=${text}`
})

function fmt(v: number) {
  return `B/. ${v.toLocaleString('es-PA', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

function fmtDate(d: string) {
  return new Date(d).toLocaleDateString('es-PA', { day: '2-digit', month: 'long', year: 'numeric' })
}

function shareWhatsApp() {
  window.open(waLink.value, '_blank')
}

async function copyLink() {
  try {
    await navigator.clipboard.writeText(window.location.href)
    copied.value = true
    setTimeout(() => copied.value = false, 2000)
  } catch {}
}

function loadIntoCart() {
  if (!quote.value) return
  cart.clear()
  quote.value.items.forEach(item => {
    cart.add({ product_id: item.product_id, name: item.name, sku: item.sku, unit_price: item.unit_price, stock: 99 })
    if (item.qty > 1) cart.qty(item.product_id, item.qty)
  })
  cart.isOpen = false
  router.push('/checkout')
}

onMounted(async () => {
  try {
    const { data } = await api.get(`/quotes/${route.params.id}`)
    quote.value = data
  } catch {}
  loading.value = false
})
</script>

<style scoped>
.page { min-height: 80vh; background: #f1f5f9; padding: 2rem 1.5rem; }
.action-bar {
  max-width: 820px; margin: 0 auto 1.5rem;
  display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: .75rem;
}
.btn-back { color: #3b82f6; text-decoration: none; font-size: .875rem; font-weight: 600; }
.btn-back:hover { text-decoration: underline; }
.action-btns { display: flex; gap: .6rem; flex-wrap: wrap; }
.btn-wa    { background: #25d366; color: #fff; border: none; border-radius: .5rem; padding: .5rem 1rem; font-size: .85rem; font-weight: 600; cursor: pointer; }
.btn-copy  { background: #f1f5f9; color: #475569; border: 1px solid #d1d5db; border-radius: .5rem; padding: .5rem 1rem; font-size: .85rem; cursor: pointer; }
.btn-print { background: #1d4ed8; color: #fff; border: none; border-radius: .5rem; padding: .5rem 1rem; font-size: .85rem; font-weight: 600; cursor: pointer; }
.btn-print:hover { background: #1e40af; }

.loading, .not-found { text-align: center; padding: 4rem; color: #94a3b8; }

/* Status banner */
.status-banner {
  max-width: 820px; margin: 0 auto 1.25rem;
  display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: .75rem;
  border-radius: .75rem; padding: 1rem 1.25rem;
}
.banner-pending  { background: #fffbeb; border: 1px solid #fcd34d; }
.banner-accepted { background: #f0fdf4; border: 1px solid #86efac; }
.banner-rejected { background: #fef2f2; border: 1px solid #fca5a5; }

.banner-left { display: flex; align-items: flex-start; gap: .75rem; }
.banner-icon { font-size: 1.3rem; line-height: 1; flex-shrink: 0; }
.banner-title {
  font-size: .9rem; font-weight: 700; margin: 0 0 .15rem;
  color: #0f172a;
}
.banner-note { font-size: .82rem; color: #64748b; margin: 0; }
.banner-right { display: flex; align-items: center; gap: .5rem; flex-shrink: 0; }

.tag-expired {
  font-size: .78rem; font-weight: 700; color: #b91c1c;
  background: #fee2e2; border-radius: .375rem; padding: .25rem .75rem;
}
.btn-order {
  background: #16a34a; color: #fff; border: none; border-radius: .5rem;
  padding: .6rem 1.25rem; font-size: .9rem; font-weight: 700; cursor: pointer;
  transition: background .15s;
}
.btn-order:hover { background: #15803d; }

/* Quote document */
.quote-doc {
  max-width: 820px; margin: 0 auto; background: #fff;
  border-radius: 1rem; padding: 3rem;
  box-shadow: 0 4px 24px rgba(0,0,0,.08);
}

.doc-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 2rem; }
.company-name { font-size: 1.6rem; font-weight: 800; color: #0f172a; margin: 0 0 .35rem; }
.company-contact { font-size: .85rem; color: #64748b; margin: .15rem 0; }

.doc-meta { text-align: right; }
.quote-badge {
  display: inline-block; background: #1d4ed8; color: #fff;
  font-size: .75rem; font-weight: 800; letter-spacing: .1em;
  padding: .3rem .9rem; border-radius: .4rem; margin-bottom: .5rem;
}
.quote-num  { font-size: 1.1rem; font-weight: 700; color: #0f172a; margin: 0 0 .25rem; }
.quote-date, .quote-valid { font-size: .82rem; color: #64748b; margin: .1rem 0; }

.doc-divider { height: 2px; background: linear-gradient(to right, #1d4ed8, #e2e8f0); margin: 0 0 2rem; border-radius: 1px; }

.doc-customer { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 2rem; gap: 1rem; }
.section-label { font-size: .7rem; font-weight: 700; text-transform: uppercase; letter-spacing: .1em; color: #94a3b8; margin: 0 0 .4rem; }
.customer-name { font-size: 1.1rem; font-weight: 700; color: #0f172a; margin: 0 0 .2rem; }
.customer-detail { font-size: .85rem; color: #64748b; margin: .1rem 0; }
.currency-badge { text-align: right; }
.currency-label { font-size: .7rem; font-weight: 700; text-transform: uppercase; letter-spacing: .1em; color: #94a3b8; margin: 0 0 .2rem; }
.currency-val { font-size: .875rem; color: #1d4ed8; font-weight: 700; margin: 0; }

/* Table */
.items-table { width: 100%; border-collapse: collapse; margin-bottom: 1.5rem; }
.items-table thead tr { background: #f8fafc; border-bottom: 2px solid #e2e8f0; }
.items-table th { padding: .65rem .75rem; font-size: .75rem; font-weight: 700; text-transform: uppercase; letter-spacing: .05em; color: #64748b; text-align: left; }
.items-table tbody tr { border-bottom: 1px solid #f1f5f9; }
.items-table tbody tr:hover { background: #fafbff; }
.items-table td { padding: .75rem .75rem; font-size: .875rem; color: #374151; }
.item-name-cell { font-weight: 600; color: #1e293b; }
.col-num { width: 48px; text-align: center; }
.col-sku { width: 110px; color: #94a3b8 !important; font-family: monospace; font-size: .78rem !important; }
.col-price { width: 120px; text-align: right; font-variant-numeric: tabular-nums; }

/* Totals */
.totals-wrap { display: flex; justify-content: flex-end; margin-bottom: 2rem; }
.totals-box { width: 300px; display: flex; flex-direction: column; gap: .4rem; }
.tot-row { display: flex; justify-content: space-between; font-size: .875rem; color: #475569; padding: .3rem 0; }
.tot-row.itbms { color: #64748b; border-top: 1px dashed #e2e8f0; padding-top: .5rem; }
.tot-row.grand { font-size: 1.15rem; font-weight: 800; color: #0f172a; border-top: 2px solid #1d4ed8; padding-top: .65rem; margin-top: .1rem; }

/* Note */
.note-box { background: #f8fafc; border-left: 4px solid #1d4ed8; border-radius: 0 .5rem .5rem 0; padding: 1rem 1.25rem; margin-bottom: 2rem; }
.note-text { font-size: .875rem; color: #475569; margin: 0; white-space: pre-wrap; line-height: 1.6; }

/* Footer */
.doc-footer { border-top: 1px solid #e2e8f0; padding-top: 1.25rem; }
.doc-footer p { font-size: .78rem; color: #94a3b8; margin: .25rem 0; line-height: 1.5; }

/* ── Print styles ──────────────────── */
@media print {
  .no-print { display: none !important; }
  .page { padding: 0; background: #fff; }
  .quote-doc {
    box-shadow: none; border-radius: 0; padding: 2cm 2.5cm;
    max-width: 100%; margin: 0;
  }
  .items-table tbody tr:hover { background: none; }
}

@media (max-width: 600px) {
  .quote-doc { padding: 1.5rem 1rem; }
  .doc-header { flex-direction: column; gap: 1rem; }
  .doc-meta { text-align: left; }
  .doc-customer { flex-direction: column; }
  .totals-wrap { justify-content: stretch; }
  .totals-box { width: 100%; }
  .col-sku { display: none; }
  .status-banner { flex-direction: column; align-items: flex-start; }
}
</style>
