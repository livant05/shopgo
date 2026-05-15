<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api/client'
import ProductCard from '../components/ProductCard.vue'
import type { Product, Category } from '../types'

const router     = useRouter()
const products   = ref<Product[]>([])
const categories = ref<Category[]>([])
const loading    = ref(true)

onMounted(async () => {
  try {
    const [pRes, cRes] = await Promise.all([
      api.get('/products', { params: { page: 1, page_size: 8 } }),
      api.get('/categories'),
    ])
    products.value   = pRes.data.data ?? []
    categories.value = cRes.data.data ?? []
  } catch {}
  loading.value = false
})

const search = ref('')
function goSearch() {
  router.push({ name: 'Catalog', query: { q: search.value } })
}
</script>

<template>
  <!-- Hero -->
  <section class="hero">
    <div class="hero-content">
      <h1>Encuentra lo que necesitas</h1>
      <p>Calidad, variedad y los mejores precios en un solo lugar</p>
      <form class="hero-search" @submit.prevent="goSearch">
        <input v-model="search" placeholder="¿Qué estás buscando?" />
        <button type="submit">Buscar</button>
      </form>
    </div>
  </section>

  <div class="container">

    <!-- Categorías -->
    <section v-if="categories.length" class="section">
      <h2 class="section-title">Categorías</h2>
      <div class="cat-grid">
        <router-link
          v-for="cat in categories" :key="cat.id"
          :to="{ name: 'Catalog', query: { category: cat.id } }"
          class="cat-card"
        >
          <span class="cat-icon">🏷️</span>
          <span>{{ cat.name }}</span>
        </router-link>
      </div>
    </section>

    <!-- Productos destacados -->
    <section class="section">
      <div class="section-header">
        <h2 class="section-title">Productos destacados</h2>
        <router-link to="/catalog" class="see-all">Ver todos →</router-link>
      </div>

      <div v-if="loading" class="skeleton-grid">
        <div v-for="n in 8" :key="n" class="skeleton-card" />
      </div>
      <div v-else-if="products.length" class="product-grid">
        <ProductCard v-for="p in products" :key="p.id" :product="p" />
      </div>
      <div v-else class="empty-state">
        <p class="empty-icon">📦</p>
        <h3>No hay productos aún</h3>
        <p>Pronto encontrarás artículos aquí.</p>
      </div>
    </section>

    <!-- Banner CTA -->
    <section class="banner-cta">
      <div>
        <h2>¿Tienes dudas?</h2>
        <p>Nuestro equipo está listo para asistirte</p>
      </div>
      <router-link to="/catalog" class="btn-cta">Ver catálogo completo</router-link>
    </section>

  </div>

  <footer class="footer">
    <p>© {{ new Date().getFullYear() }} Mi Tienda · Todos los derechos reservados</p>
  </footer>
</template>

<style scoped>
.hero {
  background: linear-gradient(135deg, #1e293b 0%, #1d4ed8 100%);
  color: #fff; padding: 5rem 1.5rem; text-align: center;
}
.hero-content { max-width: 600px; margin: 0 auto; }
.hero h1 { font-size: 2.4rem; font-weight: 800; margin-bottom: .75rem; line-height: 1.2; }
.hero p  { font-size: 1.1rem; color: #bfdbfe; margin-bottom: 2rem; }
.hero-search {
  display: flex; max-width: 480px; margin: 0 auto;
  background: #fff; border-radius: 50px; overflow: hidden;
  box-shadow: 0 4px 20px rgba(0,0,0,.25);
}
.hero-search input {
  flex: 1; padding: .75rem 1.25rem; border: none; font-size: 1rem; outline: none; color: #1e293b;
}
.hero-search button {
  padding: .75rem 1.5rem; background: #3b82f6; color: #fff;
  border: none; font-weight: 700; cursor: pointer; border-radius: 0 50px 50px 0;
}
.hero-search button:hover { background: #2563eb; }
.container { max-width: 1200px; margin: 0 auto; padding: 2rem 1.5rem; }
.section   { margin-bottom: 3rem; }
.section-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 1.25rem; }
.section-title  { font-size: 1.4rem; font-weight: 700; }
.see-all { color: #3b82f6; text-decoration: none; font-size: .9rem; font-weight: 600; }
.see-all:hover { text-decoration: underline; }
.cat-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(130px, 1fr)); gap: 1rem; }
.cat-card {
  display: flex; flex-direction: column; align-items: center; gap: .4rem;
  background: #fff; border-radius: 12px; padding: 1.25rem 1rem;
  text-decoration: none; color: #1e293b; font-size: .875rem; font-weight: 500;
  box-shadow: 0 1px 4px rgba(0,0,0,.07); transition: transform .15s;
}
.cat-card:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0,0,0,.12); }
.cat-icon { font-size: 1.8rem; }
.product-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: 1.25rem; }
.skeleton-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: 1.25rem; }
.skeleton-card {
  aspect-ratio: .8; border-radius: 12px;
  background: linear-gradient(90deg, #e2e8f0 25%, #f1f5f9 50%, #e2e8f0 75%);
  background-size: 200% 100%; animation: shimmer 1.4s infinite;
}
@keyframes shimmer { to { background-position: -200% 0; } }
.empty-state { text-align: center; padding: 4rem 2rem; color: #94a3b8; }
.empty-icon { font-size: 3rem; margin-bottom: .75rem; }
.empty-state h3 { font-size: 1.2rem; color: #64748b; margin-bottom: .4rem; }
.banner-cta {
  background: #eff6ff; border: 1px solid #bfdbfe; border-radius: 16px;
  padding: 2.5rem; display: flex; align-items: center;
  justify-content: space-between; gap: 2rem; flex-wrap: wrap; margin-bottom: 2rem;
}
.banner-cta h2 { font-size: 1.4rem; font-weight: 700; color: #1e3a8a; margin-bottom: .3rem; }
.banner-cta p  { color: #3b82f6; }
.btn-cta {
  background: #1d4ed8; color: #fff; padding: .8rem 1.75rem;
  border-radius: 8px; text-decoration: none; font-weight: 600; white-space: nowrap;
}
.btn-cta:hover { background: #1e40af; }
.footer { background: #1e293b; color: #94a3b8; text-align: center; padding: 1.5rem; font-size: .875rem; }
</style>
