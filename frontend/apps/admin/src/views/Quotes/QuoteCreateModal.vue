<script setup lang="ts">
import { ref, computed } from 'vue'
import { api } from '../../api/client'

const emit = defineEmits<{ (e: 'close'): void; (e: 'created', q: any): void }>()

// ── Customer ──────────────────────────────────────────────────
const customerName  = ref('')
const customerEmail = ref('')
const customerPhone = ref('')
const note          = ref('')

// ── Product search ────────────────────────────────────────────
const productQuery   = ref('')
const searchResults  = ref<any[]>([])
const showDropdown   = ref(false)
let   searchTimer: ReturnType<typeof setTimeout> | null = null

function scheduleSearch() {
  if (searchTimer) clearTimeout(searchTimer)
  if (!productQuery.value.trim()) { searchResults.value = []; return }
  searchTimer = setTimeout(async () => {
    try {
      const { data } = await api.get('/products', { params: { q: productQuery.value, page_size: 8 } })
      searchResults.value = data.data ?? []
    } catch { searchResults.value = [] }
  }, 280)
}

function hideDropdown() {
  setTimeout(() => { showDropdown.value = false }, 180)
}

// ── Items ─────────────────────────────────────────────────────
interface Item {
  product_id: string
  sku: string
  name: string
  qty: number
  unit_price: number
  subtotal: number
}
const items = ref<Item[]>([])

function addProduct(p: any) {
  const price = p.branch_price ?? p.base_price ?? 0
  const existing = items.value.find(it => it.product_id === p.id)
  if (existing) { existing.qty++; existing.subtotal = existing.qty * existing.unit_price; return }
  items.value.push({
    product_id: p.id,
    sku:        p.sku,
    name:       p.name,
    qty:        1,
    unit_price: price,
    subtotal:   price,
  })
  productQuery.value  = ''
  searchResults.value = []
  showDropdown.value  = false
}

function recalc(i: number) {
  const it = items.value[i]
  it.subtotal = parseFloat((it.qty * it.unit_price).toFixed(2))
}

// ── Totals ────────────────────────────────────────────────────
const TAX_RATE = 0.07
const subtotal  = computed(() => items.value.reduce((s, it) => s + it.subtotal, 0))
const taxAmount = computed(() => parseFloat((subtotal.value * TAX_RATE).toFixed(2)))
const total     = computed(() => parseFloat((subtotal.value + taxAmount.value).toFixed(2)))

function fmt(v: number) {
  return `B/. ${v.toLocaleString('es-PA', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

// ── Submit ────────────────────────────────────────────────────
const saving   = ref(false)
const error    = ref('')
const canSubmit = computed(() => customerName.value.trim() && items.value.length > 0)

async function submit() {
  if (!canSubmit.value) return
  saving.value = true
  error.value  = ''
  try {
    const { data } = await api.post('/admin/quotes', {
      customer_name:  customerName.value.trim(),
      customer_email: customerEmail.value.trim(),
      customer_phone: customerPhone.value.trim(),
      note:           note.value.trim(),
      items: items.value.map(it => ({
        product_id: it.product_id,
        sku:        it.sku,
        name:       it.name,
        qty:        it.qty,
        unit_price: it.unit_price,
      })),
    })
    emit('created', data)
  } catch (e: any) {
    error.value = e?.response?.data?.message ?? 'Error al crear la cotización.'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="overlay" @click.self="$emit('close')">
    <div class="modal">

      <div class="modal-header">
        <h2 class="modal-title">Nueva cotización</h2>
        <button class="btn-close" @click="$emit('close')">✕</button>
      </div>

      <div class="modal-body">

        <!-- Customer -->
        <div class="section">
          <p class="section-label">Cliente</p>
          <div class="field-row">
            <input v-model="customerName"  class="field" placeholder="Nombre *" />
            <input v-model="customerEmail" class="field" placeholder="Correo electrónico" type="email" />
          </div>
          <input v-model="customerPhone" class="field" placeholder="Teléfono" style="margin-top:.5rem" />
        </div>

        <!-- Product search -->
        <div class="section">
          <p class="section-label">Productos</p>
          <div class="search-wrap">
            <input
              v-model="productQuery"
              class="field"
              placeholder="Buscar producto por nombre o SKU…"
              autocomplete="off"
              @input="scheduleSearch"
              @focus="showDropdown = true"
              @blur="hideDropdown"
            />
            <div v-if="showDropdown && searchResults.length" class="search-dropdown">
              <div
                v-for="p in searchResults"
                :key="p.id"
                class="search-result"
                @mousedown.prevent="addProduct(p)"
              >
                <div class="r-info">
                  <span class="r-name">{{ p.name }}</span>
                  <span class="r-sku">{{ p.sku }}</span>
                </div>
                <span class="r-price">{{ fmt(p.branch_price ?? p.base_price ?? 0) }}</span>
              </div>
            </div>
          </div>

          <div v-if="items.length" class="items-wrap">
            <table class="items-tbl">
              <thead>
                <tr>
                  <th>Producto</th>
                  <th class="th-num">Cant.</th>
                  <th class="th-price">Precio unit.</th>
                  <th class="th-price">Subtotal</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(it, i) in items" :key="it.product_id">
                  <td class="td-name">
                    {{ it.name }}
                    <span class="item-sku">{{ it.sku }}</span>
                  </td>
                  <td>
                    <input
                      v-model.number="it.qty"
                      type="number" min="1" step="1"
                      class="num-input"
                      @input="recalc(i)"
                    />
                  </td>
                  <td>
                    <input
                      v-model.number="it.unit_price"
                      type="number" min="0" step="0.01"
                      class="num-input price-input"
                      @input="recalc(i)"
                    />
                  </td>
                  <td class="td-sub">{{ fmt(it.subtotal) }}</td>
                  <td>
                    <button class="btn-rm" @click="items.splice(i, 1)">✕</button>
                  </td>
                </tr>
              </tbody>
            </table>

            <div class="totals-row">
              <span>Subtotal <b>{{ fmt(subtotal) }}</b></span>
              <span>ITBMS 7% <b>{{ fmt(taxAmount) }}</b></span>
              <span class="grand-total">Total <b>{{ fmt(total) }}</b></span>
            </div>
          </div>

          <p v-else class="no-items">Agrega al menos un producto para continuar.</p>
        </div>

        <!-- Note -->
        <div class="section">
          <p class="section-label">Nota para el cliente (opcional)</p>
          <textarea v-model="note" class="field note-field" rows="2" placeholder="Condiciones especiales, instrucciones de entrega…" />
        </div>

        <p v-if="error" class="err-msg">{{ error }}</p>
      </div>

      <div class="modal-footer">
        <button class="btn-cancel" @click="$emit('close')">Cancelar</button>
        <button class="btn-create" :disabled="saving || !canSubmit" @click="submit">
          {{ saving ? 'Creando…' : 'Crear cotización' }}
        </button>
      </div>

    </div>
  </div>
</template>

<style scoped>
.overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,.65); z-index: 300;
  display: flex; align-items: center; justify-content: center; padding: 1rem;
}
.modal {
  width: 680px; max-width: 100%; max-height: 90vh;
  background: #0f1623; border: 1px solid #253047; border-radius: 14px;
  display: flex; flex-direction: column; overflow: hidden;
}
.modal-header {
  display: flex; justify-content: space-between; align-items: center;
  padding: 1.25rem 1.5rem; border-bottom: 1px solid #1a2540;
}
.modal-title { font-size: 1.1rem; font-weight: 700; color: #e2e8f0; margin: 0; }
.btn-close { background: none; border: none; color: #5a7298; font-size: 1.1rem; cursor: pointer; }
.btn-close:hover { color: #e2e8f0; }

.modal-body { flex: 1; overflow-y: auto; padding: 1.25rem 1.5rem; display: flex; flex-direction: column; gap: 1.25rem; }
.section { display: flex; flex-direction: column; gap: .5rem; }
.section-label { font-size: .7rem; font-weight: 700; text-transform: uppercase; letter-spacing: .08em; color: #5a7298; margin: 0; }

.field-row { display: grid; grid-template-columns: 1fr 1fr; gap: .5rem; }
.field {
  padding: .5rem .75rem; background: #0a0f1a; border: 1px solid #253047;
  border-radius: 8px; color: #d6dfe8; font-size: .875rem; outline: none;
  font-family: inherit; width: 100%; box-sizing: border-box;
}
.field:focus { border-color: #38bdf8; }
.note-field { resize: vertical; }

/* Product search */
.search-wrap { position: relative; }
.search-dropdown {
  position: absolute; top: calc(100% + 4px); left: 0; right: 0;
  background: #0a0f1a; border: 1px solid #253047; border-radius: 8px;
  max-height: 240px; overflow-y: auto; z-index: 10;
  box-shadow: 0 8px 24px rgba(0,0,0,.4);
}
.search-result {
  display: flex; justify-content: space-between; align-items: center;
  padding: .65rem 1rem; cursor: pointer; transition: background .1s;
}
.search-result:hover { background: rgba(56,189,248,.07); }
.r-info { display: flex; flex-direction: column; gap: .1rem; }
.r-name { font-size: .875rem; color: #d6dfe8; font-weight: 500; }
.r-sku  { font-size: .75rem; color: #5a7298; font-family: monospace; }
.r-price { font-size: .875rem; font-weight: 700; color: #38bdf8; white-space: nowrap; }

/* Items table */
.items-wrap { margin-top: .5rem; }
.items-tbl { width: 100%; border-collapse: collapse; font-size: .82rem; }
.items-tbl thead tr { background: #070d17; }
.items-tbl th { padding: .45rem .5rem; text-align: left; color: #5a7298; font-size: .72rem; text-transform: uppercase; border-bottom: 1px solid #1a2540; }
.items-tbl td { padding: .5rem .5rem; color: #a8b8cc; border-bottom: 1px solid #1a2540; vertical-align: middle; }
.items-tbl tr:last-child td { border-bottom: none; }
.td-name { color: #d6dfe8; font-weight: 500; }
.item-sku { display: block; font-size: .71rem; color: #5a7298; font-family: monospace; }
.th-num   { width: 64px; }
.th-price { width: 100px; text-align: right; }
.td-sub   { text-align: right; font-weight: 700; color: #38bdf8; }
.num-input {
  width: 58px; padding: .3rem .4rem; background: #0a0f1a; border: 1px solid #253047;
  border-radius: 6px; color: #d6dfe8; font-size: .82rem; text-align: right; outline: none;
}
.num-input:focus { border-color: #38bdf8; }
.price-input { width: 88px; }
.btn-rm { background: none; border: none; color: #5a7298; cursor: pointer; padding: .2rem .3rem; border-radius: 4px; font-size: .85rem; }
.btn-rm:hover { color: #f87171; background: rgba(239,68,68,.1); }

/* Totals preview */
.totals-row {
  display: flex; gap: 1.25rem; justify-content: flex-end; flex-wrap: wrap;
  padding: .65rem .5rem; font-size: .82rem; color: #5a7298; border-top: 1px solid #1a2540;
}
.grand-total { font-weight: 700; color: #38bdf8; }

.no-items { font-size: .82rem; color: #3d5070; text-align: center; padding: 1rem 0; }
.err-msg  { font-size: .82rem; color: #f87171; margin: 0; }

/* Footer */
.modal-footer {
  display: flex; justify-content: flex-end; gap: .75rem;
  padding: 1rem 1.5rem; border-top: 1px solid #1a2540;
}
.btn-cancel {
  padding: .55rem 1.1rem; background: #0a0f1a; border: 1px solid #253047;
  border-radius: 8px; color: #5a7298; font-size: .875rem; cursor: pointer;
}
.btn-cancel:hover { border-color: #38bdf8; color: #d6dfe8; }
.btn-create {
  padding: .55rem 1.4rem; background: rgba(56,189,248,.12); border: 1px solid rgba(56,189,248,.3);
  border-radius: 8px; color: #38bdf8; font-size: .875rem; font-weight: 700; cursor: pointer;
}
.btn-create:hover:not(:disabled) { background: rgba(56,189,248,.22); }
.btn-create:disabled { opacity: .45; cursor: not-allowed; }
</style>
