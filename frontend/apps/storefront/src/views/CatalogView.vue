<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '../api/client'
import ProductCard from '../components/ProductCard.vue'
import type { Product, Category } from '../types'

const route  = useRoute()
const router = useRouter()

const products   = ref<Product[]>([])
const categories = ref<Category[]>([])
const total      = ref(0)
const loading    = ref(true)
const page       = ref(1)
const pageSize   = 16

const search     = ref((route.query.q as string) || '')
const categoryId = ref((route.query.category as string) || '')
const sortBy     = ref('name')
const inStock    = ref(false)
const priceMin   = ref('')
const priceMax   = ref('')

const totalPages = computed(() => Math.ceil(total.value / pageSize))

let debounceTimer: ReturnType<typeof setTimeout> | null = null

async function load() {
  loading.value = true
  try {
    const { data } = await api.get('/products', {
      params: {
        page:        page.value,
        page_size:   pageSize,
        q:           search.value || undefined,
        category_id: categoryId.value || undefined,
        sort:        sortBy.value,
        in_stock:    inStock.value ? 'true' : undefined,
        price_min:   priceMin.value ? Number(priceMin.value) : undefined,
        price_max:   priceMax.value ? Number(priceMax.value) : undefined,
      },
    })
    products.value = data.data  ?? []
    total.value    = data.total ?? 0
  } catch {}
  loading.value = false
}

function scheduleLoad() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => { page.value = 1; load() }, 350)
}

onMounted(async () => {
  const { data } = await api.get('/categories').catch(() => ({ data: { data: [] } }))
  categories.value = data.data ?? []
  await load()
})

// Debounce text search; immediate reload for dropdowns and toggles
watch(search, scheduleLoad)
watch([categoryId, sortBy, inStock], () => { page.value = 1; load() })
watch(page, load)

function applyPriceFilter() { page.value = 1; load() }

function clearFilters() {
  search.value = ''; categoryId.value = ''; sortBy.value = 'name'
  inStock.value = false; priceMin.value = ''; priceMax.value = ''
  page.value = 1; load()
}

function applySearch(e: Event) {
  e.preventDefault()
  if (debounceTimer) clearTimeout(debounceTimer)
  router.replace({ query: { ...route.query, q: search.value || undefined } })
  page.value = 1; load()
}

const activeFilters = computed(() =>
  [categoryId.value, inStock.value, priceMin.value, priceMax.value, search.value].some(Boolean)
)
</script>

<template>
  <div class="page">

    <!-- Filters bar -->
    <div class="filters-bar">
      <div class="inner">

        <!-- Search -->
        <form class="filter-search" @submit="applySearch">
          <input v-model="search" placeholder="Buscar productos…" />
          <button type="submit">🔍</button>
        </form>

        <!-- Category -->
        <select v-model="categoryId" class="select">
          <option value="">Todas las categorías</option>
          <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
        </select>

        <!-- Sort -->
        <select v-model="sortBy" class="select">
          <option value="name">Nombre A-Z</option>
          <option value="-name">Nombre Z-A</option>
          <option value="price">Precio ↑</option>
          <option value="-price">Precio ↓</option>
        </select>

        <!-- Price range -->
        <div class="price-range">
          <input v-model="priceMin" type="number" min="0" placeholder="$ Min"
                 class="price-input" @change="applyPriceFilter" />
          <span class="price-sep">–</span>
          <input v-model="priceMax" type="number" min="0" placeholder="$ Max"
                 class="price-input" @change="applyPriceFilter" />
        </div>

        <!-- In stock toggle -->
        <label class="stock-toggle">
          <input type="checkbox" v-model="inStock" />
          <span>En stock</span>
        </label>

        <!-- Result count + clear -->
        <span class="result-count">{{ total }} resultado{{ total !== 1 ? 's' : '' }}</span>
        <button v-if="activeFilters" class="btn-clear-inline" @click="clearFilters">✕ Limpiar</button>

      </div>
    </div>

    <div class="container">

      <!-- Grid -->
      <div v-if="loading" class="skeleton-grid">
        <div v-for="n in 16" :key="n" class="skeleton-card" />
      </div>

      <div v-else-if="products.length" class="product-grid">
        <ProductCard v-for="p in products" :key="p.id" :product="p" />
      </div>

      <div v-else class="empty-state">
        <p class="empty-icon">🔍</p>
        <h3>Sin resultados</h3>
        <p>Intenta con otra búsqueda o ajusta los filtros</p>
        <button class="btn-clear" @click="clearFilters">Limpiar filtros</button>
      </div>

      <!-- Pagination -->
      <div v-if="totalPages > 1" class="pagination">
        <button :disabled="page === 1" @click="page--">← Anterior</button>
        <span>Página {{ page }} de {{ totalPages }}</span>
        <button :disabled="page >= totalPages" @click="page++">Siguiente →</button>
      </div>

    </div>
  </div>
</template>

<style scoped>
.page { min-height: 80vh; }
.filters-bar {
  background: #fff; border-bottom: 1px solid #e2e8f0;
  padding: .75rem 1.5rem; position: sticky; top: 56px; z-index: 10;
}
.inner {
  max-width: 1200px; margin: 0 auto;
  display: flex; align-items: center; gap: .75rem; flex-wrap: wrap;
}
.filter-search {
  display: flex; border: 1px solid #e2e8f0; border-radius: 8px; overflow: hidden;
}
.filter-search input {
  padding: .4rem .7rem; border: none; outline: none;
  font-size: .875rem; color: #1e293b; width: 180px;
}
.filter-search button {
  padding: .4rem .7rem; background: #f8fafc; border: none;
  cursor: pointer; border-left: 1px solid #e2e8f0;
}
.select {
  padding: .4rem .7rem; border: 1px solid #e2e8f0; border-radius: 8px;
  font-size: .875rem; outline: none; background: #fff; color: #1e293b; cursor: pointer;
}
.price-range { display: flex; align-items: center; gap: .3rem; }
.price-input {
  width: 80px; padding: .4rem .5rem; border: 1px solid #e2e8f0;
  border-radius: 8px; font-size: .8rem; outline: none; color: #1e293b;
}
.price-sep { color: #94a3b8; font-size: .8rem; }
.stock-toggle {
  display: flex; align-items: center; gap: .35rem;
  font-size: .875rem; color: #475569; cursor: pointer; white-space: nowrap;
}
.stock-toggle input { cursor: pointer; accent-color: #3b82f6; }
.result-count { font-size: .8rem; color: #94a3b8; margin-left: auto; white-space: nowrap; }
.btn-clear-inline {
  padding: .3rem .65rem; background: #fee2e2; color: #ef4444;
  border: none; border-radius: 6px; cursor: pointer; font-size: .8rem; font-weight: 600;
  white-space: nowrap;
}
.btn-clear-inline:hover { background: #fecaca; }

.container { max-width: 1200px; margin: 0 auto; padding: 2rem 1.5rem; }
.product-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: 1.25rem; }
.skeleton-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: 1.25rem; }
.skeleton-card {
  aspect-ratio: .8; border-radius: 12px;
  background: linear-gradient(90deg, #e2e8f0 25%, #f1f5f9 50%, #e2e8f0 75%);
  background-size: 200% 100%; animation: shimmer 1.4s infinite;
}
@keyframes shimmer { to { background-position: -200% 0; } }
.empty-state { text-align: center; padding: 5rem 2rem; color: #94a3b8; }
.empty-icon { font-size: 3rem; margin-bottom: .75rem; }
.empty-state h3 { font-size: 1.2rem; color: #64748b; margin-bottom: .4rem; }
.btn-clear {
  margin-top: 1rem; padding: .5rem 1.25rem; background: #3b82f6;
  color: #fff; border: none; border-radius: 8px; cursor: pointer; font-weight: 600;
}
.btn-clear:hover { background: #2563eb; }
.pagination { display: flex; align-items: center; justify-content: center; gap: 1.5rem; margin-top: 2.5rem; }
.pagination button {
  padding: .5rem 1.25rem; border: 1px solid #e2e8f0; border-radius: 8px;
  background: #fff; cursor: pointer; font-weight: 600; color: #3b82f6;
}
.pagination button:disabled { color: #cbd5e1; cursor: not-allowed; }
.pagination span { font-size: .875rem; color: #64748b; }
</style>
