<template>
  <Teleport to="body">
    <div v-if="open" class="cp-backdrop" @click.self="hide()">
      <div class="cp-modal" ref="modalRef">
        <!-- Input -->
        <div class="cp-search">
          <span class="cp-icon">🔍</span>
          <input
            ref="inputRef"
            v-model="query"
            class="cp-input"
            placeholder="Buscar productos, usuarios, cupones…"
            @keydown.escape="hide()"
            @keydown.arrow-down.prevent="moveDown()"
            @keydown.arrow-up.prevent="moveUp()"
            @keydown.enter.prevent="selectCurrent()"
          />
          <kbd class="cp-esc">Esc</kbd>
        </div>

        <!-- Results -->
        <div class="cp-body" ref="listRef">
          <!-- Quick actions (sin query) -->
          <template v-if="!query">
            <div class="cp-group-label">Navegación rápida</div>
            <div v-for="(action, idx) in quickActions" :key="'qa-'+idx"
              class="cp-item" :class="{ active: flatIdx(idx, 'quick') === cursor }"
              @click="go(action)"
              @mousemove="cursor = flatIdx(idx, 'quick')">
              <span class="cp-item-icon">{{ action.icon }}</span>
              <span class="cp-item-label">{{ action.label }}</span>
              <span class="cp-item-sub">{{ action.sub }}</span>
            </div>
          </template>

          <template v-if="query">
            <!-- Products -->
            <template v-if="products.length > 0">
              <div class="cp-group-label">Productos</div>
              <div v-for="(p, idx) in products" :key="'p-'+p.id"
                class="cp-item" :class="{ active: flatIdx(idx, 'products') === cursor }"
                @click="go({ route: '/products', label: p.name })"
                @mousemove="cursor = flatIdx(idx, 'products')">
                <span class="cp-item-icon">📦</span>
                <div class="cp-item-main">
                  <span class="cp-item-label">{{ p.name }}</span>
                  <span class="cp-item-meta">{{ p.sku }} · ${{ fmt(p.base_price) }}</span>
                </div>
                <span class="cp-item-badge" :class="p.is_active ? 'badge-ok' : 'badge-off'">
                  {{ p.is_active ? 'Activo' : 'Inactivo' }}
                </span>
              </div>
            </template>

            <!-- Coupons -->
            <template v-if="filteredCoupons.length > 0">
              <div class="cp-group-label">Cupones</div>
              <div v-for="(c, idx) in filteredCoupons" :key="'c-'+c.id"
                class="cp-item" :class="{ active: flatIdx(idx, 'coupons') === cursor }"
                @click="go({ route: '/coupons', label: c.code })"
                @mousemove="cursor = flatIdx(idx, 'coupons')">
                <span class="cp-item-icon">🏷️</span>
                <div class="cp-item-main">
                  <span class="cp-item-label cp-mono">{{ c.code }}</span>
                  <span class="cp-item-meta">{{ c.type === 'percent' ? c.value+'%' : '$'+c.value }} descuento</span>
                </div>
                <span class="cp-item-badge" :class="c.is_active ? 'badge-ok' : 'badge-off'">
                  {{ c.is_active ? 'Activo' : 'Inactivo' }}
                </span>
              </div>
            </template>

            <!-- Users -->
            <template v-if="filteredUsers.length > 0">
              <div class="cp-group-label">Usuarios</div>
              <div v-for="(u, idx) in filteredUsers" :key="'u-'+u.id"
                class="cp-item" :class="{ active: flatIdx(idx, 'users') === cursor }"
                @click="go({ route: '/users', label: u.email })"
                @mousemove="cursor = flatIdx(idx, 'users')">
                <span class="cp-item-icon">👤</span>
                <div class="cp-item-main">
                  <span class="cp-item-label">{{ u.first_name }} {{ u.last_name }}</span>
                  <span class="cp-item-meta">{{ u.email }} · {{ u.role }}</span>
                </div>
              </div>
            </template>

            <div v-if="!loading && products.length === 0 && filteredCoupons.length === 0 && filteredUsers.length === 0"
              class="cp-empty">Sin resultados para "{{ query }}"</div>
            <div v-if="loading" class="cp-empty">Buscando…</div>
          </template>
        </div>

        <!-- Footer -->
        <div class="cp-footer">
          <span><kbd>↑↓</kbd> navegar</span>
          <span><kbd>↵</kbd> abrir</span>
          <span><kbd>Esc</kbd> cerrar</span>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, computed, nextTick, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api/client'
import { useCommandPalette } from '../composables/useCommandPalette'

const { open, hide } = useCommandPalette()
const router  = useRouter()
const inputRef = ref<HTMLInputElement | null>(null)
const listRef  = ref<HTMLElement | null>(null)
const query    = ref('')
const cursor   = ref(0)
const loading  = ref(false)

const products = ref<any[]>([])
const coupons  = ref<any[]>([])
const users    = ref<any[]>([])

const quickActions = [
  { icon: '📊', label: 'Dashboard',  sub: 'Inicio',        route: '/' },
  { icon: '🛒', label: 'Órdenes',    sub: 'Gestión',       route: '/orders' },
  { icon: '💳', label: 'POS',        sub: 'Punto de venta', route: '/pos' },
  { icon: '📦', label: 'Productos',  sub: 'Catálogo',      route: '/products' },
  { icon: '📋', label: 'Inventario', sub: 'Stock',         route: '/inventory' },
  { icon: '🏪', label: 'Sucursales', sub: 'Configuración', route: '/branches' },
  { icon: '👥', label: 'Usuarios',   sub: 'Equipo',        route: '/users' },
  { icon: '📈', label: 'Reportes',   sub: 'Análisis',      route: '/reports' },
  { icon: '🏷️', label: 'Cupones',    sub: 'Descuentos',    route: '/coupons' },
  { icon: '⚙️', label: 'Config',     sub: 'Tienda',        route: '/settings' },
]

const filteredCoupons = computed(() => {
  if (!query.value) return []
  const q = query.value.toLowerCase()
  return coupons.value.filter(c => c.code.toLowerCase().includes(q)).slice(0, 4)
})

const filteredUsers = computed(() => {
  if (!query.value) return []
  const q = query.value.toLowerCase()
  return users.value.filter(u =>
    u.email.toLowerCase().includes(q) ||
    (u.first_name + ' ' + u.last_name).toLowerCase().includes(q)
  ).slice(0, 4)
})

// Map section + local index → global cursor index
function flatIdx(idx: number, section: string) {
  if (!query.value) return idx
  let base = 0
  if (section === 'products') return idx
  base += products.value.length
  if (section === 'coupons') return base + idx
  base += filteredCoupons.value.length
  if (section === 'users') return base + idx
  return base + idx
}

const totalItems = computed(() => {
  if (!query.value) return quickActions.length
  return products.value.length + filteredCoupons.value.length + filteredUsers.value.length
})

function moveDown() { cursor.value = Math.min(cursor.value + 1, totalItems.value - 1) }
function moveUp()   { cursor.value = Math.max(cursor.value - 1, 0) }

function go(item: { route: string; label: string }) {
  router.push(item.route)
  hide()
  query.value = ''
}

function selectCurrent() {
  if (!query.value) {
    go(quickActions[cursor.value])
    return
  }
  const all = [
    ...products.value.map(p => ({ route: '/products', label: p.name })),
    ...filteredCoupons.value.map(c => ({ route: '/coupons', label: c.code })),
    ...filteredUsers.value.map(u => ({ route: '/users', label: u.email })),
  ]
  if (all[cursor.value]) go(all[cursor.value])
}

let searchTimer: any
watch(query, (q) => {
  cursor.value = 0
  clearTimeout(searchTimer)
  if (!q) { products.value = []; return }
  loading.value = true
  searchTimer = setTimeout(async () => {
    try {
      const { data } = await api.get(`/products?q=${encodeURIComponent(q)}&page_size=5`)
      products.value = data.data ?? []
    } catch {}
    loading.value = false
  }, 250)
})

watch(open, async (v) => {
  if (!v) { query.value = ''; return }
  await nextTick()
  inputRef.value?.focus()
  // Load full coupons + users for client-side filtering
  try {
    const [cr, ur] = await Promise.all([
      api.get('/admin/coupons'),
      api.get('/admin/users'),
    ])
    coupons.value = cr.data.data ?? []
    users.value   = ur.data.data ?? []
  } catch {}
})

// Global keyboard shortcut
function onKeydown(e: KeyboardEvent) {
  if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
    e.preventDefault()
    if (open.value) hide(); else { open.value = true }
  }
}
onMounted(() => window.addEventListener('keydown', onKeydown))
onUnmounted(() => window.removeEventListener('keydown', onKeydown))

const fmt = (n: number) => (n ?? 0).toLocaleString('es-MX', { minimumFractionDigits: 2 })
</script>

<style scoped>
.cp-backdrop { position:fixed; inset:0; background:rgba(0,0,0,.65); z-index:9000; display:flex; align-items:flex-start; justify-content:center; padding-top:12vh; }
.cp-modal    { background:#1c2333; border:1px solid #2d3a52; border-radius:14px; width:600px; max-width:95vw; box-shadow:0 24px 64px rgba(0,0,0,.6); overflow:hidden; display:flex; flex-direction:column; }

.cp-search   { display:flex; align-items:center; gap:10px; padding:14px 18px; border-bottom:1px solid #2d3a52; }
.cp-icon     { font-size:18px; flex-shrink:0; }
.cp-input    { flex:1; background:none; border:none; outline:none; color:#eaf0f7; font-size:16px; }
.cp-input::placeholder { color:#3d5070; }
.cp-esc      { background:#253047; border:1px solid #344762; color:#5a6a87; border-radius:5px; padding:2px 7px; font-size:11px; font-family:inherit; flex-shrink:0; }

.cp-body     { max-height:440px; overflow-y:auto; padding:8px 0; }
.cp-group-label { font-size:10px; font-weight:700; color:#5a6a87; text-transform:uppercase; letter-spacing:1px; padding:10px 18px 4px; }
.cp-item     { display:flex; align-items:center; gap:12px; padding:10px 18px; cursor:pointer; transition:background .1s; }
.cp-item.active { background:rgba(56,189,248,.1); }
.cp-item:hover  { background:rgba(56,189,248,.06); }
.cp-item-icon   { font-size:18px; flex-shrink:0; width:28px; text-align:center; }
.cp-item-main   { flex:1; display:flex; flex-direction:column; gap:2px; min-width:0; }
.cp-item-label  { font-size:14px; color:#eaf0f7; font-weight:500; white-space:nowrap; overflow:hidden; text-overflow:ellipsis; }
.cp-item-sub    { font-size:12px; color:#5a6a87; margin-left:auto; flex-shrink:0; }
.cp-item-meta   { font-size:12px; color:#5a6a87; }
.cp-mono        { font-family:monospace; letter-spacing:.05em; }
.cp-item-badge  { font-size:11px; font-weight:600; padding:2px 8px; border-radius:10px; flex-shrink:0; }
.badge-ok  { background:rgba(74,222,128,.1); color:#4ade80; }
.badge-off { background:rgba(248,113,113,.1); color:#f87171; }
.cp-empty    { text-align:center; color:#5a6a87; font-size:13px; padding:28px; }

.cp-footer   { display:flex; gap:20px; padding:10px 18px; border-top:1px solid #1a2235; background:#141c2c; }
.cp-footer span { font-size:11px; color:#5a6a87; display:flex; align-items:center; gap:5px; }
kbd { background:#253047; border:1px solid #344762; border-radius:4px; padding:1px 5px; font-size:10px; color:#8494ac; font-family:inherit; }
</style>
