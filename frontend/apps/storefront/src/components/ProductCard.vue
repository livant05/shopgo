<script setup lang="ts">
import type { Product } from '../types'
import { useCartStore } from '../stores/cart'
import { useRouter } from 'vue-router'

const props = defineProps<{ product: Product }>()
const cart   = useCartStore()
const router = useRouter()

function mainImage() {
  const main = props.product.images?.find(i => i.is_main)
  return main?.url || props.product.images?.[0]?.url || ''
}

function price() {
  return (props.product.branch_price ?? props.product.base_price).toLocaleString('es-MX', {
    style: 'currency', currency: 'MXN',
  })
}

function addToCart() {
  cart.add({
    product_id: props.product.id,
    name:       props.product.name,
    sku:        props.product.sku,
    unit_price: props.product.branch_price ?? props.product.base_price,
    image_url:  mainImage(),
    stock:      props.product.stock ?? 99,
  })
}
</script>

<template>
  <div class="card" @click="router.push(`/product/${product.id}`)">
    <div class="img-wrap">
      <img v-if="mainImage()" :src="mainImage()" :alt="product.name" />
      <div v-else class="img-placeholder">📦</div>
      <span v-if="product.stock === 0" class="out-badge">Sin stock</span>
    </div>
    <div class="body">
      <p class="sku">{{ product.sku }}</p>
      <h3 class="name">{{ product.name }}</h3>
      <p class="price">{{ price() }}</p>
      <button
        class="btn-cart"
        :disabled="product.stock === 0"
        @click.stop="addToCart"
      >
        {{ product.stock === 0 ? 'Sin stock' : '+ Agregar' }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.card {
  background: #fff; border-radius: 12px; overflow: hidden;
  box-shadow: 0 1px 4px rgba(0,0,0,.08); cursor: pointer;
  transition: transform .15s, box-shadow .15s;
  display: flex; flex-direction: column;
}
.card:hover { transform: translateY(-3px); box-shadow: 0 6px 20px rgba(0,0,0,.12); }
.img-wrap {
  position: relative; aspect-ratio: 1; background: #f8fafc;
  display: flex; align-items: center; justify-content: center; overflow: hidden;
}
.img-wrap img { width: 100%; height: 100%; object-fit: cover; }
.img-placeholder { font-size: 3.5rem; color: #cbd5e1; }
.out-badge {
  position: absolute; top: 8px; left: 8px; background: #ef4444;
  color: #fff; font-size: .7rem; font-weight: 600; padding: 2px 8px; border-radius: 999px;
}
.body { padding: 1rem; display: flex; flex-direction: column; gap: .35rem; flex: 1; }
.sku  { font-size: .7rem; color: #94a3b8; margin: 0; text-transform: uppercase; letter-spacing: .05em; }
.name { font-size: .95rem; font-weight: 600; color: #1e293b; margin: 0; line-height: 1.3; }
.price { font-size: 1.05rem; font-weight: 700; color: #2563eb; margin: 0; }
.btn-cart {
  margin-top: auto; padding: .5rem; border: none; border-radius: 8px;
  background: #3b82f6; color: #fff; font-weight: 600; font-size: .85rem;
  cursor: pointer; transition: background .15s;
}
.btn-cart:hover:not(:disabled) { background: #2563eb; }
.btn-cart:disabled { background: #e2e8f0; color: #94a3b8; cursor: not-allowed; }
</style>
