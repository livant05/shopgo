<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '../api/client'
import ProductCard from '../components/ProductCard.vue'
import type { Product, Category } from '../types'

const route  = useRoute()
const router = useRouter()

// — state —
const products   = ref<Product[]>([])
const categories = ref<Category[]>([])
const tags       = ref<string[]>([])
const total      = ref(0)
const loading    = ref(true)
const page       = ref(1)
const pageSize   = 16
const sidebarOpen = ref(true)

// — filters (read initial values from URL) —
const search     = ref((route.query.q as string)        || '')
const categoryId = ref((route.query.category as string) || '')
const sortBy     = ref((route.query.sort as string)     || 'name')
const inStock    = ref(route.query.in_stock === 'true')
const priceMin   = ref((route.query.price_min as string) || '')
const priceMax   = ref((route.query.price_max as string) || '')
const activeTag  = ref((route.query.tag as string)       || '')

const totalPages = computed(() => Math.ceil(total.value / pageSize))

// — category tree —
interface CatNode extends Category { children: CatNode[] }
const catTree = computed<CatNode[]>(() => {
  const map = new Map<string, CatNode>()
  categories.value.forEach(c => map.set(c.id, { ...c, children: [] }))
  const roots: CatNode[] = []
  map.forEach(node => {
    if (node.parent_id && map.has(node.parent_id)) {
      map.get(node.parent_id)!.children.push(node)
    } else {
      roots.push(node)
    }
  })
  return roots.sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0) || a.name.localeCompare(b.name))
})

// — active filter chips —
interface Chip { label: string; clear: () => void }
const chips = computed<Chip[]>(() => {
  const out: Chip[] = []
  if (search.value)     out.push({ label: `"${search.value}"`,          clear: () => { search.value = '' } })
  if (categoryId.value) {
    const cat = categories.value.find(c => c.id === categoryId.value)
    out.push({ label: cat?.name ?? categoryId.value, clear: () => { categoryId.value = '' } })
  }
  if (activeTag.value)  out.push({ label: `#${activeTag.value}`,        clear: () => { activeTag.value = '' } })
  if (inStock.value)    out.push({ label: 'En stock',                    clear: () => { inStock.value = false } })
  if (priceMin.value)   out.push({ label: `Desde $${priceMin.value}`,   clear: () => { priceMin.value = '' } })
  if (priceMax.value)   out.push({ label: `Hasta $${priceMax.value}`,   clear: () => { priceMax.value = '' } })
  return out
})

const hasFilters = computed(() => chips.value.length > 0)

// — debounced load —
let debounceTimer: ReturnType<typeof setTimeout> | null = null

async function load(resetPage = false) {
  if (resetPage) page.value = 1
  loading.value = true
  try {
    const { data } = await api.get('/products', {
      params: {
        page:        page.value,
        page_size:   pageSize,
        q:           search.value    || undefined,
        category_id: categoryId.value || undefined,
        sort:        sortBy.value,
        in_stock:    inStock.value ? 'true' : undefined,
        price_min:   priceMin.value  ? Number(priceMin.value)  : undefined,
        price_max:   priceMax.value  ? Number(priceMax.value)  : undefined,
        tag:         activeTag.value || undefined,
      },
    })
    products.value = data.data  ?? []
    total.value    = data.total ?? 0
  } catch { /* keep showing previous results */ }
  loading.value = false
  syncURL()
}

function scheduleLoad() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => load(true), 350)
}

// — sync URL query params —
function syncURL() {
  router.replace({
    query: {
      q:         search.value     || undefined,
      category:  categoryId.value || undefined,
      sort:      sortBy.value !== 'name' ? sortBy.value : undefined,
      in_stock:  inStock.value ? 'true' : undefined,
      price_min: priceMin.value  || undefined,
      price_max: priceMax.value  || undefined,
      tag:       activeTag.value || undefined,
      page:      page.value > 1 ? String(page.value) : undefined,
    },
  })
}

function clearAll() {
  search.value = ''; categoryId.value = ''; sortBy.value = 'name'
  inStock.value = false; priceMin.value = ''; priceMax.value = ''
  activeTag.value = ''
  load(true)
}

function selectCategory(id: string) {
  categoryId.value = categoryId.value === id ? '' : id
  load(true)
}

function selectTag(tag: string) {
  activeTag.value = activeTag.value === tag ? '' : tag
  load(true)
}

function applyPrice() { load(true) }

watch(search, scheduleLoad)
watch([sortBy, inStock, activeTag, categoryId], () => load(true))
watch(page, () => load(false))

onMounted(async () => {
  page.value = Number(route.query.page) || 1
  const [catRes, tagRes] = await Promise.allSettled([
    api.get('/categories'),
    api.get('/tags'),
  ])
  if (catRes.status === 'fulfilled') categories.value = catRes.value.data?.data ?? []
  if (tagRes.status === 'fulfilled') tags.value = tagRes.value.data?.data ?? []
  await load(false)
})
</script>

<template>
  <div class="catalog-shell">

    <!-- Mobile filter toggle -->
    <div class="mobile-bar">
      <button class="mobile-filter-btn" @click="sidebarOpen = !sidebarOpen">
        🔧 Filtros <span v-if="hasFilters" class="chip-dot">{{ chips.length }}</span>
      </button>
      <span class="result-count-mobile">{{ total }} resultado{{ total !== 1 ? 's' : '' }}</span>
    </div>

    <div class="catalog-body">
      <!-- ── Sidebar ─────────────────────────────────── -->
      <aside class="sidebar" :class="{ open: sidebarOpen }">
        <div class="sidebar-inner">

          <!-- Search -->
          <div class="filter-section">
            <h3 class="filter-heading">Buscar</h3>
            <div class="search-wrap">
              <input
                v-model="search"
                class="search-input"
                placeholder="Nombre, SKU, descripción…"
                @keyup.enter="scheduleLoad"
              />
              <span class="search-icon">🔍</span>
            </div>
          </div>

          <!-- Categories -->
          <div class="filter-section">
            <h3 class="filter-heading">Categoría</h3>
            <button
              class="cat-btn"
              :class="{ active: !categoryId }"
              @click="selectCategory('')"
            >Todas</button>
            <template v-for="root in catTree" :key="root.id">
              <button
                class="cat-btn"
                :class="{ active: categoryId === root.id }"
                @click="selectCategory(root.id)"
              >{{ root.name }} <span class="cat-count">{{ root.product_count }}</span></button>
              <button
                v-for="child in root.children"
                :key="child.id"
                class="cat-btn cat-child"
                :class="{ active: categoryId === child.id }"
                @click="selectCategory(child.id)"
              >└ {{ child.name }} <span class="cat-count">{{ child.product_count }}</span></button>
            </template>
          </div>

          <!-- Price range -->
          <div class="filter-section">
            <h3 class="filter-heading">Precio</h3>
            <div class="price-row">
              <input
                v-model="priceMin"
                type="number"
                min="0"
                placeholder="$ Mín"
                class="price-input"
              />
              <span class="price-sep">–</span>
              <input
                v-model="priceMax"
                type="number"
                min="0"
                placeholder="$ Máx"
                class="price-input"
              />
            </div>
            <button class="btn-apply" @click="applyPrice()">Aplicar precio</button>
          </div>

          <!-- In stock -->
          <div class="filter-section">
            <label class="toggle-label">
              <div class="toggle-track" :class="{ on: inStock }" @click="inStock = !inStock">
                <div class="toggle-thumb" />
              </div>
              <span>Solo en stock</span>
            </label>
          </div>

          <!-- Tags -->
          <div v-if="tags.length" class="filter-section">
            <h3 class="filter-heading">Etiquetas</h3>
            <div class="tag-cloud">
              <button
                v-for="tag in tags"
                :key="tag"
                class="tag-pill"
                :class="{ active: activeTag === tag }"
                @click="selectTag(tag)"
              >#{{ tag }}</button>
            </div>
          </div>

          <!-- Sort -->
          <div class="filter-section">
            <h3 class="filter-heading">Ordenar por</h3>
            <select v-model="sortBy" class="sort-select">
              <option value="name">Nombre A-Z</option>
              <option value="-name">Nombre Z-A</option>
              <option value="price">Precio ↑</option>
              <option value="-price">Precio ↓</option>
            </select>
          </div>

          <button v-if="hasFilters" class="btn-clear-all" @click="clearAll()">
            ✕ Limpiar todos los filtros
          </button>
        </div>
      </aside>

      <!-- ── Main results ─────────────────────────────── -->
      <main class="results">

        <!-- Top bar: chips + count -->
        <div class="results-topbar">
          <div class="chips-wrap">
            <span class="result-count">{{ total }} resultado{{ total !== 1 ? 's' : '' }}</span>
            <div v-if="hasFilters" class="chips">
              <span
                v-for="chip in chips"
                :key="chip.label"
                class="chip"
                @click="chip.clear(); load(true)"
              >{{ chip.label }} ✕</span>
            </div>
          </div>
        </div>

        <!-- Skeleton loading -->
        <div v-if="loading" class="product-grid">
          <div v-for="n in pageSize" :key="n" class="skeleton-card" />
        </div>

        <!-- Product grid -->
        <div v-else-if="products.length" class="product-grid">
          <ProductCard v-for="p in products" :key="p.id" :product="p" />
        </div>

        <!-- Empty state -->
        <div v-else class="empty-state">
          <p class="empty-icon">🔍</p>
          <h3>Sin resultados</h3>
          <p>Intenta con otra búsqueda o ajusta los filtros</p>
          <button class="btn-reset" @click="clearAll()">Limpiar filtros</button>
        </div>

        <!-- Pagination -->
        <div v-if="totalPages > 1" class="pagination">
          <button :disabled="page === 1" @click="page--">← Anterior</button>
          <div class="page-numbers">
            <button
              v-for="p in totalPages"
              :key="p"
              :class="{ current: p === page }"
              @click="page = p"
            >{{ p }}</button>
          </div>
          <button :disabled="page >= totalPages" @click="page++">Siguiente →</button>
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
/* ── Shell ──────────────────────────────── */
.catalog-shell { min-height: 80vh; }
.mobile-bar {
  display: none;
  align-items: center; justify-content: space-between;
  padding: .6rem 1rem; background: #fff; border-bottom: 1px solid #e2e8f0;
  position: sticky; top: 56px; z-index: 20;
}
.mobile-filter-btn {
  display: flex; align-items: center; gap: .4rem;
  background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 8px;
  padding: .4rem .75rem; font-size: .85rem; cursor: pointer; font-weight: 600;
}
.chip-dot {
  background: #3b82f6; color: #fff; border-radius: 999px;
  font-size: .7rem; padding: .1rem .4rem; min-width: 18px; text-align: center;
}
.result-count-mobile { font-size: .8rem; color: #94a3b8; }

.catalog-body { display: flex; max-width: 1280px; margin: 0 auto; padding: 1.5rem 1.5rem 3rem; gap: 1.75rem; }

/* ── Sidebar ────────────────────────────── */
.sidebar { width: 240px; flex-shrink: 0; }
.sidebar-inner { position: sticky; top: 72px; display: flex; flex-direction: column; gap: .25rem; }

.filter-section { border-bottom: 1px solid #f1f5f9; padding-bottom: 1rem; margin-bottom: 1rem; }
.filter-section:last-of-type { border-bottom: none; }

.filter-heading {
  font-size: .75rem; font-weight: 700; text-transform: uppercase; letter-spacing: .08em;
  color: #94a3b8; margin: 0 0 .65rem;
}

.search-wrap { position: relative; }
.search-input {
  width: 100%; padding: .5rem .75rem .5rem 2rem;
  border: 1px solid #e2e8f0; border-radius: 8px; font-size: .875rem;
  outline: none; color: #1e293b; box-sizing: border-box;
}
.search-input:focus { border-color: #3b82f6; }
.search-icon { position: absolute; left: .6rem; top: 50%; transform: translateY(-50%); font-size: .85rem; }

.cat-btn {
  display: flex; justify-content: space-between; align-items: center;
  width: 100%; padding: .4rem .6rem; background: none; border: none;
  text-align: left; font-size: .875rem; color: #475569; cursor: pointer;
  border-radius: 6px; transition: all .15s;
}
.cat-btn:hover { background: #f1f5f9; }
.cat-btn.active { background: #eff6ff; color: #1d4ed8; font-weight: 600; }
.cat-child { padding-left: 1.25rem; font-size: .83rem; color: #64748b; }
.cat-count { font-size: .75rem; color: #94a3b8; background: #f1f5f9; border-radius: 999px; padding: .1rem .4rem; }
.cat-btn.active .cat-count { background: #dbeafe; color: #2563eb; }

.price-row { display: flex; align-items: center; gap: .4rem; margin-bottom: .6rem; }
.price-input {
  flex: 1; padding: .4rem .5rem; border: 1px solid #e2e8f0;
  border-radius: 6px; font-size: .8rem; outline: none; color: #1e293b; min-width: 0;
}
.price-input:focus { border-color: #3b82f6; }
.price-sep { color: #cbd5e1; flex-shrink: 0; }
.btn-apply {
  width: 100%; padding: .4rem; background: #f8fafc; border: 1px solid #e2e8f0;
  border-radius: 6px; font-size: .8rem; cursor: pointer; color: #475569;
}
.btn-apply:hover { background: #eff6ff; border-color: #bfdbfe; color: #2563eb; }

.toggle-label { display: flex; align-items: center; gap: .6rem; cursor: pointer; font-size: .875rem; color: #475569; }
.toggle-track {
  width: 36px; height: 20px; background: #e2e8f0; border-radius: 999px;
  position: relative; transition: background .2s; flex-shrink: 0;
}
.toggle-track.on { background: #3b82f6; }
.toggle-thumb {
  position: absolute; left: 2px; top: 2px;
  width: 16px; height: 16px; background: #fff; border-radius: 50%;
  transition: left .2s; box-shadow: 0 1px 3px rgba(0,0,0,.2);
}
.toggle-track.on .toggle-thumb { left: 18px; }

.tag-cloud { display: flex; flex-wrap: wrap; gap: .35rem; }
.tag-pill {
  padding: .25rem .6rem; border: 1px solid #e2e8f0; border-radius: 999px;
  font-size: .78rem; cursor: pointer; background: #f8fafc; color: #64748b;
  transition: all .15s;
}
.tag-pill:hover { border-color: #93c5fd; color: #1d4ed8; background: #eff6ff; }
.tag-pill.active { background: #3b82f6; color: #fff; border-color: #3b82f6; }

.sort-select {
  width: 100%; padding: .45rem .6rem; border: 1px solid #e2e8f0;
  border-radius: 6px; font-size: .875rem; outline: none; background: #fff; color: #1e293b; cursor: pointer;
}
.sort-select:focus { border-color: #3b82f6; }

.btn-clear-all {
  width: 100%; padding: .5rem; background: #fee2e2; color: #dc2626;
  border: none; border-radius: 8px; cursor: pointer; font-size: .8rem; font-weight: 600;
  margin-top: .5rem;
}
.btn-clear-all:hover { background: #fecaca; }

/* ── Results ───────────────────────────── */
.results { flex: 1; min-width: 0; }

.results-topbar { display: flex; align-items: center; margin-bottom: 1rem; gap: .75rem; min-height: 32px; }
.chips-wrap { display: flex; align-items: center; flex-wrap: wrap; gap: .5rem; }
.result-count { font-size: .875rem; color: #64748b; white-space: nowrap; }
.chips { display: flex; flex-wrap: wrap; gap: .4rem; }
.chip {
  display: inline-flex; align-items: center; gap: .3rem;
  background: #eff6ff; color: #1d4ed8; border: 1px solid #bfdbfe;
  border-radius: 999px; font-size: .78rem; padding: .2rem .65rem;
  cursor: pointer; font-weight: 500; transition: all .15s;
}
.chip:hover { background: #fee2e2; color: #dc2626; border-color: #fca5a5; }

.product-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(210px, 1fr));
  gap: 1.25rem;
}
.skeleton-card {
  aspect-ratio: .8; border-radius: 12px;
  background: linear-gradient(90deg, #e2e8f0 25%, #f1f5f9 50%, #e2e8f0 75%);
  background-size: 200% 100%; animation: shimmer 1.4s infinite;
}
@keyframes shimmer { to { background-position: -200% 0; } }

.empty-state { text-align: center; padding: 5rem 2rem; color: #94a3b8; }
.empty-icon { font-size: 3rem; margin-bottom: .75rem; }
.empty-state h3 { font-size: 1.2rem; color: #64748b; margin-bottom: .4rem; }
.btn-reset {
  margin-top: 1rem; padding: .5rem 1.25rem; background: #3b82f6;
  color: #fff; border: none; border-radius: 8px; cursor: pointer; font-weight: 600;
}
.btn-reset:hover { background: #2563eb; }

.pagination {
  display: flex; align-items: center; justify-content: center;
  gap: .75rem; margin-top: 2.5rem; flex-wrap: wrap;
}
.pagination > button {
  padding: .45rem 1rem; border: 1px solid #e2e8f0; border-radius: 8px;
  background: #fff; cursor: pointer; font-weight: 600; color: #3b82f6; font-size: .875rem;
}
.pagination > button:disabled { color: #cbd5e1; cursor: not-allowed; }
.page-numbers { display: flex; gap: .35rem; flex-wrap: wrap; }
.page-numbers button {
  width: 34px; height: 34px; border: 1px solid #e2e8f0; border-radius: 7px;
  background: #fff; cursor: pointer; font-size: .85rem; color: #475569;
}
.page-numbers button.current {
  background: #3b82f6; color: #fff; border-color: #3b82f6; font-weight: 700;
}
.page-numbers button:hover:not(.current) { background: #f1f5f9; }

/* ── Responsive ─────────────────────────── */
@media (max-width: 768px) {
  .mobile-bar { display: flex; }
  .catalog-body { flex-direction: column; padding: .75rem 1rem 2rem; gap: 0; }
  .sidebar {
    width: 100%; max-height: 0; overflow: hidden;
    transition: max-height .3s ease;
    background: #f8fafc; border-bottom: 1px solid #e2e8f0;
  }
  .sidebar.open { max-height: 2000px; }
  .sidebar-inner { position: static; padding: .75rem 0 .25rem; }
  .product-grid { grid-template-columns: repeat(auto-fill, minmax(155px, 1fr)); }
}
</style>
