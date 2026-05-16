<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '../api/client'
import { useCartStore } from '../stores/cart'
import { useQuoteStore } from '../stores/quote'
import type { Product } from '../types'

const route  = useRoute()
const router = useRouter()
const cart   = useCartStore()
const quote  = useQuoteStore()

const product  = ref<Product | null>(null)
const loading  = ref(true)
const qty      = ref(1)
const imgIndex = ref(0)
const added    = ref(false)
const quoted   = ref(false)

onMounted(async () => {
  try {
    const { data } = await api.get(`/products/${route.params.id}`)
    product.value = data
  } catch { router.push({ name: 'Catalog' }) }
  loading.value = false
})

const price = computed(() => {
  if (!product.value) return ''
  return (product.value.branch_price ?? product.value.base_price).toLocaleString('es-MX', {
    style: 'currency', currency: 'MXN',
  })
})

const images = computed(() => product.value?.images ?? [])
const stock  = computed(() => product.value?.stock ?? 0)

function addToCart() {
  if (!product.value) return
  for (let i = 0; i < qty.value; i++) {
    cart.add({
      product_id: product.value.id,
      name:       product.value.name,
      sku:        product.value.sku,
      unit_price: product.value.branch_price ?? product.value.base_price,
      image_url:  images.value.find(x => x.is_main)?.url ?? images.value[0]?.url ?? '',
      stock:      stock.value,
    })
  }
  added.value = true
  setTimeout(() => added.value = false, 2000)
}

function addToQuote() {
  if (!product.value) return
  quote.add({
    product_id: product.value.id,
    sku:        product.value.sku,
    name:       product.value.name,
    unit_price: product.value.branch_price ?? product.value.base_price,
    image_url:  images.value.find(x => x.is_main)?.url ?? images.value[0]?.url ?? '',
  })
  quoted.value = true
  setTimeout(() => quoted.value = false, 2000)
}
</script>

<template>
  <div class="container">

    <!-- Breadcrumb -->
    <nav class="breadcrumb">
      <router-link to="/">Inicio</router-link>
      <span>/</span>
      <router-link to="/catalog">Catálogo</router-link>
      <span v-if="product">/ {{ product.name }}</span>
    </nav>

    <!-- Skeleton -->
    <div v-if="loading" class="skeleton-wrap">
      <div class="sk-img" />
      <div class="sk-body">
        <div class="sk-line long" />
        <div class="sk-line med"  />
        <div class="sk-line short" />
        <div class="sk-line med"  />
      </div>
    </div>

    <!-- Product detail -->
    <div v-else-if="product" class="product-layout">

      <!-- Gallery -->
      <div class="gallery">
        <div class="main-img">
          <img v-if="images[imgIndex]" :src="images[imgIndex].url" :alt="images[imgIndex].alt_text" />
          <div v-else class="no-img">📦</div>
        </div>
        <div v-if="images.length > 1" class="thumbs">
          <img
            v-for="(img, i) in images" :key="i"
            :src="img.url" :alt="img.alt_text"
            :class="{ active: i === imgIndex }"
            @click="imgIndex = i"
          />
        </div>
      </div>

      <!-- Info -->
      <div class="info">
        <p class="sku">SKU: {{ product.sku }}</p>
        <h1 class="name">{{ product.name }}</h1>

        <p class="price">{{ price }}</p>

        <div class="stock-row">
          <span v-if="stock > 10" class="stock-ok">✅ En stock</span>
          <span v-else-if="stock > 0" class="stock-low">⚠️ Últimas {{ stock }} unidades</span>
          <span v-else class="stock-none">❌ Sin stock</span>
        </div>

        <p v-if="product.description" class="description">{{ product.description }}</p>

        <div v-if="product.tags?.length" class="tags">
          <span v-for="t in product.tags" :key="t" class="tag">{{ t }}</span>
        </div>

        <!-- Qty + cart -->
        <div class="add-row">
          <div class="qty-ctrl">
            <button @click="qty = Math.max(1, qty - 1)">−</button>
            <span>{{ qty }}</span>
            <button @click="qty = Math.min(stock, qty + 1)" :disabled="qty >= stock">+</button>
          </div>
          <button
            class="btn-add"
            :disabled="stock === 0"
            :class="{ success: added }"
            @click="addToCart"
          >
            {{ added ? '✓ Agregado' : '🛒 Agregar al carrito' }}
          </button>
        </div>

        <!-- Quote action -->
        <button class="btn-quote-add" @click="addToQuote">
          {{ quoted ? '✓ Agregado a cotización' : '📄 Agregar a cotización' }}
        </button>

        <router-link v-if="added" to="/cart" class="btn-go-cart">Ver carrito →</router-link>
      </div>
    </div>

  </div>
</template>

<style scoped>
.container { max-width: 1100px; margin: 0 auto; padding: 2rem 1.5rem; }
.breadcrumb { display: flex; gap: .5rem; font-size: .85rem; color: #94a3b8; margin-bottom: 1.5rem; }
.breadcrumb a { color: #3b82f6; text-decoration: none; }
.breadcrumb a:hover { text-decoration: underline; }

/* skeleton */
.skeleton-wrap { display: grid; grid-template-columns: 1fr 1fr; gap: 3rem; }
.sk-img { aspect-ratio: 1; background: #e2e8f0; border-radius: 12px; animation: shimmer 1.4s infinite; }
.sk-body { display: flex; flex-direction: column; gap: 1rem; padding-top: 1rem; }
.sk-line { height: 1rem; border-radius: 4px; background: linear-gradient(90deg, #e2e8f0 25%, #f1f5f9 50%, #e2e8f0 75%); background-size: 200% 100%; animation: shimmer 1.4s infinite; }
.sk-line.long  { width: 80%; }
.sk-line.med   { width: 50%; }
.sk-line.short { width: 30%; }
@keyframes shimmer { to { background-position: -200% 0; } }

.product-layout { display: grid; grid-template-columns: 1fr 1fr; gap: 3rem; align-items: start; }

/* Gallery */
.gallery {}
.main-img { background: #f8fafc; border-radius: 16px; overflow: hidden; aspect-ratio: 1; display: flex; align-items: center; justify-content: center; margin-bottom: .75rem; }
.main-img img { width: 100%; height: 100%; object-fit: cover; }
.no-img { font-size: 5rem; color: #cbd5e1; }
.thumbs { display: flex; gap: .5rem; flex-wrap: wrap; }
.thumbs img { width: 72px; height: 72px; object-fit: cover; border-radius: 8px; cursor: pointer; border: 2px solid transparent; transition: border-color .15s; }
.thumbs img.active { border-color: #3b82f6; }

/* Info */
.info { display: flex; flex-direction: column; gap: 1rem; }
.sku  { font-size: .8rem; color: #94a3b8; text-transform: uppercase; letter-spacing: .05em; }
.name { font-size: 1.75rem; font-weight: 800; line-height: 1.2; color: #1e293b; }
.price { font-size: 2rem; font-weight: 700; color: #2563eb; }
.stock-ok   { color: #10b981; font-size: .9rem; font-weight: 600; }
.stock-low  { color: #f59e0b; font-size: .9rem; font-weight: 600; }
.stock-none { color: #ef4444; font-size: .9rem; font-weight: 600; }
.description { color: #475569; line-height: 1.7; font-size: .95rem; }
.tags { display: flex; gap: .5rem; flex-wrap: wrap; }
.tag { background: #eff6ff; color: #3b82f6; padding: .25rem .75rem; border-radius: 999px; font-size: .78rem; font-weight: 500; }

.add-row { display: flex; gap: 1rem; align-items: center; margin-top: .5rem; }
.qty-ctrl { display: flex; align-items: center; border: 1px solid #e2e8f0; border-radius: 8px; overflow: hidden; }
.qty-ctrl button { width: 36px; height: 40px; border: none; background: #f8fafc; font-size: 1.2rem; cursor: pointer; }
.qty-ctrl button:hover:not(:disabled) { background: #e2e8f0; }
.qty-ctrl button:disabled { color: #cbd5e1; cursor: not-allowed; }
.qty-ctrl span { width: 40px; text-align: center; font-weight: 600; }

.btn-add {
  flex: 1; padding: .75rem 1.5rem; border: none; border-radius: 10px;
  background: #1d4ed8; color: #fff; font-size: 1rem; font-weight: 700; cursor: pointer;
  transition: background .15s;
}
.btn-add:hover:not(:disabled) { background: #1e40af; }
.btn-add:disabled { background: #e2e8f0; color: #94a3b8; cursor: not-allowed; }
.btn-add.success { background: #10b981; }

.btn-quote-add {
  width: 100%; padding: .65rem 1rem; border: 1.5px solid #f59e0b; border-radius: 10px;
  background: #fffbeb; color: #92400e; font-size: .95rem; font-weight: 600; cursor: pointer;
  transition: all .15s;
}
.btn-quote-add:hover { background: #fef3c7; border-color: #d97706; }

.btn-go-cart {
  display: inline-block; margin-top: .25rem; color: #3b82f6;
  font-weight: 600; text-decoration: none;
}
.btn-go-cart:hover { text-decoration: underline; }

@media (max-width: 768px) {
  .product-layout { grid-template-columns: 1fr; }
  .skeleton-wrap  { grid-template-columns: 1fr; }
}
</style>
