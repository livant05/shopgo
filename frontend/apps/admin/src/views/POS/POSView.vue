<template>
  <div class="pos">
    <!-- Panel izquierdo: catálogo -->
    <div class="catalog-panel">
      <div class="catalog-header">
        <h2 class="panel-title">Punto de Venta</h2>
        <select v-model="branchID" @change="loadProducts()" class="branch-select">
          <option value="">— Sucursal —</option>
          <option v-for="b in branches" :key="b.id" :value="b.id">{{ b.name }}</option>
        </select>
      </div>

      <input v-model="productSearch" @input="debouncedLoad()" class="search-box" placeholder="Buscar producto…" />

      <!-- Filtros de categoría -->
      <div class="cat-tabs">
        <button class="cat-tab" :class="{ active: categoryID === '' }" @click="setCategory('')">Todos</button>
        <button v-for="c in categories" :key="c.id"
          class="cat-tab" :class="{ active: categoryID === c.id }"
          @click="setCategory(c.id)">{{ c.name }}</button>
      </div>

      <div class="products-grid">
        <div v-if="loadingProds" class="loading-msg">Cargando…</div>
        <template v-else>
          <div v-for="p in products" :key="p.id" class="product-tile"
            @click="addToCart(p)" :class="{ 'out-of-stock': (p.stock ?? 0) === 0 }">
            <div class="tile-emoji">📦</div>
            <div class="tile-name">{{ p.name }}</div>
            <div class="tile-sku">{{ p.sku }}</div>
            <div class="tile-price">${{ fmt(p.branch_price ?? p.base_price) }}</div>
            <div class="tile-stock" :class="(p.stock ?? 0) === 0 ? 'zero' : (p.stock ?? 0) <= 5 ? 'low' : ''">
              {{ p.stock ?? '—' }} en stock
            </div>
          </div>
          <div v-if="products.length === 0" class="empty-grid">Sin productos</div>
        </template>
      </div>
    </div>

    <!-- Panel derecho: carrito -->
    <div class="cart-panel">
      <div class="cart-header">
        <span class="cart-title">Carrito</span>
        <button class="btn-clear" @click="cart=[]" :disabled="cart.length===0">Limpiar</button>
      </div>

      <div class="cart-items">
        <div v-if="cart.length === 0" class="cart-empty">Agrega productos del catálogo</div>
        <div v-for="(ci, idx) in cart" :key="ci.product_id" class="cart-item">
          <div class="ci-name">{{ ci.name }}</div>
          <div class="ci-price">${{ fmt(ci.unit_price) }}</div>
          <div class="ci-qty">
            <button @click="ci.qty = Math.max(1, ci.qty - 1)">−</button>
            <span>{{ ci.qty }}</span>
            <button @click="ci.qty++">+</button>
          </div>
          <div class="ci-total">${{ fmt(ci.unit_price * ci.qty) }}</div>
          <button class="ci-remove" @click="cart.splice(idx,1)">✕</button>
        </div>
      </div>

      <!-- Cupón -->
      <div class="coupon-row">
        <input v-model="couponCode" @keyup.enter="applyCoupon()" class="coupon-input" placeholder="Código de descuento" :disabled="!!couponApplied" />
        <button class="btn-coupon" @click="applyCoupon()" :disabled="!couponCode || !!couponApplied">
          {{ couponApplied ? '✓' : 'Aplicar' }}
        </button>
        <button v-if="couponApplied" class="btn-coupon-rm" @click="removeCoupon()">✕</button>
      </div>
      <div v-if="couponErr" class="coupon-err">{{ couponErr }}</div>

      <div class="cart-totals">
        <div class="tot-row"><span>Subtotal</span><span>${{ fmt(subtotal) }}</span></div>
        <div v-if="discount > 0" class="tot-row discount">
          <span>Descuento ({{ couponApplied }})</span><span>−${{ fmt(discount) }}</span>
        </div>
        <div class="tot-row"><span>IVA (16%)</span><span>${{ fmt(tax) }}</span></div>
        <div class="tot-row total"><span>Total</span><span>${{ fmt(total) }}</span></div>
      </div>

      <div class="cart-footer">
        <div class="field"><label>Nota / cliente</label>
          <input v-model="customerNote" placeholder="Venta mostrador" />
        </div>
        <button class="btn-checkout" @click="checkout()" :disabled="cart.length===0 || !branchID || checking">
          {{ checking ? 'Procesando…' : 'Cobrar $' + fmt(total) }}
        </button>
      </div>

      <div v-if="checkoutErr" class="err-box">{{ checkoutErr }}</div>
    </div>

    <!-- Modal: ticket de venta -->
    <Teleport to="body">
      <div v-if="receipt" class="receipt-backdrop" @click.self="receipt = null">
        <div class="receipt-modal" id="receipt-print">
          <div class="receipt-header">
            <div class="receipt-store">{{ storeName }}</div>
            <div class="receipt-title">TICKET DE VENTA</div>
            <div class="receipt-date">{{ receiptDate }}</div>
            <div class="receipt-order">Orden: <strong>{{ receipt.id.slice(0,8).toUpperCase() }}</strong></div>
          </div>
          <div class="receipt-divider">- - - - - - - - - - - - - - - -</div>
          <div class="receipt-items">
            <div v-for="i in receipt.items" :key="i.product_id" class="receipt-item">
              <span class="ri-name">{{ i.name }}</span>
              <span class="ri-qty">×{{ i.quantity }}</span>
              <span class="ri-total">${{ fmt(i.line_total) }}</span>
            </div>
          </div>
          <div class="receipt-divider">- - - - - - - - - - - - - - - -</div>
          <div class="receipt-totals">
            <div class="rt-row"><span>Subtotal</span><span>${{ fmt(receipt.subtotal) }}</span></div>
            <div v-if="(receipt.discount ?? 0) > 0" class="rt-row discount">
              <span>Descuento</span><span>−${{ fmt(receipt.discount) }}</span>
            </div>
            <div class="rt-row"><span>IVA</span><span>${{ fmt(receipt.tax) }}</span></div>
            <div class="rt-row total"><span>TOTAL</span><span>${{ fmt(receipt.total) }}</span></div>
          </div>
          <div class="receipt-divider">- - - - - - - - - - - - - - - -</div>
          <div class="receipt-thanks">¡Gracias por su compra!</div>
          <div class="receipt-actions no-print">
            <button class="btn-print" @click="printReceipt()">🖨 Imprimir</button>
            <button class="btn-new-sale" @click="receipt = null">Nueva venta</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '../../api/client'
import { useAuthStore } from '../../stores/auth'
import type { Product, Branch } from '../../types'

const auth = useAuthStore()
const storeName = import.meta.env.VITE_STORE_NAME ?? 'Mi Tienda'

interface CartItem { product_id: string; name: string; sku: string; unit_price: number; qty: number }
interface Category  { id: string; name: string }

const branches      = ref<Branch[]>([])
const categories    = ref<Category[]>([])
const products      = ref<Product[]>([])
const cart          = ref<CartItem[]>([])
const branchID      = ref(auth.user?.branch_id ?? '')
const categoryID    = ref('')
const productSearch = ref('')
const customerNote  = ref('')
const couponCode    = ref('')
const couponApplied = ref('')
const couponErr     = ref('')
const discount      = ref(0)
const loadingProds  = ref(false)
const checking      = ref(false)
const checkoutErr   = ref('')
const receipt       = ref<any>(null)
const receiptDate   = ref('')

const subtotal = computed(() => cart.value.reduce((s, i) => s + i.unit_price * i.qty, 0))
const tax      = computed(() => (subtotal.value - discount.value) * 0.16)
const total    = computed(() => subtotal.value - discount.value + tax.value)

async function loadBranches() {
  const { data } = await api.get('/branches')
  branches.value = data.data ?? []
}

async function loadCategories() {
  try {
    const { data } = await api.get('/categories')
    categories.value = data.data ?? []
  } catch {}
}

async function loadProducts() {
  if (!branchID.value) return
  loadingProds.value = true
  try {
    const p = new URLSearchParams({ branch_id: branchID.value, page_size: '60' })
    if (productSearch.value) p.set('q', productSearch.value)
    if (categoryID.value)    p.set('category_id', categoryID.value)
    const { data } = await api.get(`/products?${p}`)
    products.value = (data.data ?? []).filter((p: Product) => p.is_active)
  } catch(e) { console.error(e) }
  finally { loadingProds.value = false }
}

function setCategory(id: string) {
  categoryID.value = id
  loadProducts()
}

let debTimer: any
function debouncedLoad() { clearTimeout(debTimer); debTimer = setTimeout(loadProducts, 300) }

function addToCart(p: Product) {
  if ((p.stock ?? 1) === 0) return
  const existing = cart.value.find(i => i.product_id === p.id)
  if (existing) { existing.qty++; return }
  cart.value.push({
    product_id: p.id, name: p.name, sku: p.sku,
    unit_price: p.branch_price ?? p.base_price, qty: 1,
  })
}

async function applyCoupon() {
  if (!couponCode.value) return
  couponErr.value = ''
  try {
    const { data } = await api.post('/coupons/validate', {
      code: couponCode.value,
      subtotal: subtotal.value,
    })
    couponApplied.value = couponCode.value.toUpperCase()
    discount.value = data.discount ?? 0
  } catch(e: any) {
    couponErr.value = e.response?.data?.message ?? 'Cupón inválido'
  }
}

function removeCoupon() {
  couponCode.value = ''
  couponApplied.value = ''
  discount.value = 0
  couponErr.value = ''
}

async function checkout() {
  if (cart.value.length === 0 || !branchID.value) return
  checking.value = true; checkoutErr.value = ''
  try {
    const { data } = await api.post('/orders', {
      branch_id: branchID.value,
      customer_id: auth.user?.id,
      items: cart.value.map(i => ({ product_id: i.product_id, quantity: i.qty })),
      coupon_code: couponApplied.value || undefined,
      notes: customerNote.value || 'Venta mostrador POS',
      currency: 'MXN',
    })
    receiptDate.value = new Date().toLocaleString('es-MX')
    receipt.value = data
    cart.value = []
    customerNote.value = ''
    removeCoupon()
    await loadProducts()
  } catch(e: any) {
    checkoutErr.value = e.response?.data?.message ?? 'Error al procesar'
  } finally { checking.value = false }
}

function printReceipt() {
  window.print()
}

const fmt = (n: number) => (n ?? 0).toLocaleString('es-MX', { minimumFractionDigits: 2 })

onMounted(async () => {
  await Promise.all([loadBranches(), loadCategories()])
  if (branchID.value) await loadProducts()
})
</script>

<style scoped>
.pos { display:flex; gap:0; height:calc(100vh - 64px); margin:-32px -36px; overflow:hidden; }

.catalog-panel { flex:1; display:flex; flex-direction:column; padding:24px; gap:12px; overflow:hidden; border-right:1px solid #253047; }
.catalog-header { display:flex; align-items:center; justify-content:space-between; gap:12px; }
.panel-title    { font-size:18px; font-weight:700; color:#eaf0f7; margin:0; }
.branch-select  { background:#1c2333; border:1px solid #2d3a52; color:#d6dfe8; padding:8px 12px; border-radius:7px; font-size:13px; }
.search-box     { background:#1c2333; border:1px solid #2d3a52; color:#d6dfe8; padding:10px 14px; border-radius:8px; font-size:13px; width:100%; box-sizing:border-box; }
.search-box:focus { outline:none; border-color:#38bdf8; }

/* Category tabs */
.cat-tabs { display:flex; gap:6px; flex-wrap:wrap; }
.cat-tab  { background:none; border:1px solid #2d3a52; color:#5a6a87; padding:5px 12px; border-radius:20px; font-size:12px; cursor:pointer; transition:all .15s; white-space:nowrap; }
.cat-tab:hover  { border-color:#38bdf8; color:#38bdf8; }
.cat-tab.active { background:rgba(56,189,248,.12); border-color:#38bdf8; color:#38bdf8; }

.products-grid { flex:1; overflow-y:auto; display:grid; grid-template-columns:repeat(auto-fill, minmax(150px,1fr)); gap:10px; padding-right:4px; }
.loading-msg   { grid-column:1/-1; text-align:center; color:#5a6a87; padding:40px; }
.empty-grid    { grid-column:1/-1; text-align:center; color:#5a6a87; padding:40px; }

.product-tile   { background:#1c2333; border:1px solid #2d3a52; border-radius:10px; padding:14px; cursor:pointer; transition:all .15s; display:flex; flex-direction:column; gap:4px; }
.product-tile:hover { border-color:#38bdf8; background:rgba(56,189,248,.06); transform:translateY(-1px); }
.out-of-stock   { opacity:.4; cursor:not-allowed; }
.out-of-stock:hover { transform:none; border-color:#2d3a52; background:#1c2333; }
.tile-emoji     { font-size:28px; margin-bottom:4px; }
.tile-name      { font-size:13px; font-weight:600; color:#eaf0f7; line-height:1.3; }
.tile-sku       { font-size:11px; color:#5a6a87; font-family:monospace; }
.tile-price     { font-size:15px; font-weight:700; color:#4ade80; margin-top:4px; }
.tile-stock     { font-size:11px; color:#5a6a87; }
.tile-stock.low  { color:#fb923c; }
.tile-stock.zero { color:#f87171; }

/* Cart panel */
.cart-panel  { width:340px; flex-shrink:0; display:flex; flex-direction:column; background:#0f1623; overflow:hidden; }
.cart-header { display:flex; align-items:center; justify-content:space-between; padding:20px 20px 16px; border-bottom:1px solid #253047; }
.cart-title  { font-size:15px; font-weight:700; color:#eaf0f7; }
.btn-clear   { background:none; border:none; color:#5a6a87; font-size:12px; cursor:pointer; }
.btn-clear:disabled { opacity:.3; cursor:not-allowed; }

.cart-items  { flex:1; overflow-y:auto; padding:12px 16px; display:flex; flex-direction:column; gap:8px; min-height:0; }
.cart-empty  { text-align:center; color:#5a6a87; font-size:13px; margin-top:40px; }
.cart-item   { display:flex; align-items:center; gap:8px; background:#1c2333; border-radius:8px; padding:10px 12px; flex-shrink:0; }
.ci-name     { flex:1; font-size:12px; color:#d6dfe8; line-height:1.3; }
.ci-price    { font-size:11px; color:#5a6a87; white-space:nowrap; }
.ci-qty      { display:flex; align-items:center; gap:6px; }
.ci-qty button { background:#253047; border:none; color:#d6dfe8; width:22px; height:22px; border-radius:4px; cursor:pointer; font-size:14px; line-height:1; }
.ci-qty span   { font-size:13px; color:#eaf0f7; min-width:18px; text-align:center; }
.ci-total    { font-size:12px; color:#4ade80; font-weight:600; white-space:nowrap; }
.ci-remove   { background:none; border:none; color:#5a6a87; cursor:pointer; font-size:12px; }

/* Coupon */
.coupon-row   { display:flex; align-items:center; gap:6px; padding:8px 16px; border-top:1px solid #1a2235; }
.coupon-input { flex:1; background:#1c2333; border:1px solid #2d3a52; color:#d6dfe8; padding:7px 10px; border-radius:6px; font-size:12px; }
.coupon-input:focus { outline:none; border-color:#38bdf8; }
.coupon-input:disabled { opacity:.6; }
.btn-coupon   { background:none; border:1px solid #38bdf8; color:#38bdf8; padding:6px 10px; border-radius:6px; font-size:12px; cursor:pointer; white-space:nowrap; }
.btn-coupon:disabled { opacity:.4; cursor:not-allowed; }
.btn-coupon-rm { background:none; border:none; color:#f87171; font-size:14px; cursor:pointer; padding:4px; }
.coupon-err   { font-size:11px; color:#f87171; padding:0 16px 6px; }

.cart-totals { padding:12px 20px; border-top:1px solid #253047; display:flex; flex-direction:column; gap:6px; }
.tot-row     { display:flex; justify-content:space-between; font-size:13px; color:#8494ac; }
.tot-row.discount { color:#fb923c; }
.tot-row.total { font-size:16px; font-weight:700; color:#eaf0f7; margin-top:4px; border-top:1px solid #2d3a52; padding-top:8px; }

.cart-footer { padding:14px 16px; border-top:1px solid #253047; display:flex; flex-direction:column; gap:10px; }
.field       { display:flex; flex-direction:column; gap:5px; }
.field label { font-size:11px; color:#8494ac; text-transform:uppercase; letter-spacing:.5px; }
.field input { background:#1c2333; border:1px solid #2d3a52; color:#d6dfe8; padding:8px 10px; border-radius:6px; font-size:13px; }
.field input:focus { outline:none; border-color:#38bdf8; }
.btn-checkout { background:#38bdf8; color:#080c14; border:none; padding:13px; border-radius:8px; font-size:14px; font-weight:700; cursor:pointer; width:100%; }
.btn-checkout:disabled { opacity:.5; cursor:not-allowed; }

.err-box { margin:0 16px 16px; background:rgba(248,113,113,.1); border:1px solid rgba(248,113,113,.3); color:#f87171; padding:10px 14px; border-radius:7px; font-size:12px; }

/* Receipt modal */
.receipt-backdrop { position:fixed; inset:0; background:rgba(0,0,0,.7); display:flex; align-items:center; justify-content:center; z-index:9999; }
.receipt-modal    { background:#fff; color:#111; width:320px; border-radius:10px; padding:28px 24px; font-family:monospace; display:flex; flex-direction:column; gap:10px; box-shadow:0 20px 60px rgba(0,0,0,.5); }
.receipt-header   { text-align:center; display:flex; flex-direction:column; gap:4px; }
.receipt-store    { font-size:16px; font-weight:700; }
.receipt-title    { font-size:13px; letter-spacing:2px; color:#555; }
.receipt-date     { font-size:11px; color:#888; }
.receipt-order    { font-size:12px; margin-top:4px; }
.receipt-divider  { color:#bbb; font-size:12px; letter-spacing:1px; text-align:center; }
.receipt-items    { display:flex; flex-direction:column; gap:6px; }
.receipt-item     { display:flex; align-items:baseline; gap:4px; font-size:12px; }
.ri-name { flex:1; }
.ri-qty  { color:#555; }
.ri-total { font-weight:600; }
.receipt-totals   { display:flex; flex-direction:column; gap:4px; }
.rt-row  { display:flex; justify-content:space-between; font-size:13px; }
.rt-row.discount  { color:#e07000; }
.rt-row.total     { font-size:15px; font-weight:700; margin-top:4px; }
.receipt-thanks   { text-align:center; font-size:12px; color:#555; }
.receipt-actions  { display:flex; gap:10px; margin-top:4px; }
.btn-print    { flex:1; background:#111; color:#fff; border:none; padding:10px; border-radius:6px; font-size:13px; cursor:pointer; }
.btn-new-sale { flex:1; background:none; border:1px solid #ccc; color:#333; padding:10px; border-radius:6px; font-size:13px; cursor:pointer; }

@media print {
  body > *:not(#receipt-print) { display: none !important; }
  .no-print { display: none !important; }
  .receipt-backdrop { position:static; background:none; }
  .receipt-modal { box-shadow:none; border-radius:0; }
}
</style>
