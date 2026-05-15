<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useCartStore } from '../stores/cart'
import { useAuthStore } from '../stores/auth'

const cart   = useCartStore()
const auth   = useAuthStore()
const router = useRouter()

const couponCode  = ref('')
const couponError = ref('')
const couponOk    = ref(false)
const applying    = ref(false)

async function applyCoupon() {
  if (!couponCode.value.trim()) return
  applying.value = true
  const result = await cart.applyCoupon(couponCode.value.trim())
  applying.value = false
  if (result.ok) { couponOk.value = true; couponError.value = '' }
  else { couponError.value = result.message ?? 'Cupón inválido'; couponOk.value = false }
}

function fmt(n: number) {
  return n.toLocaleString('es-MX', { style: 'currency', currency: 'MXN' })
}

function goCheckout() {
  if (!auth.isAuth) {
    router.push({ name: 'Login', query: { redirect: '/checkout' } })
  } else {
    router.push({ name: 'Checkout' })
  }
}
</script>

<template>
  <div class="container">
    <h1 class="title">🛒 Tu carrito</h1>

    <!-- Empty -->
    <div v-if="cart.isEmpty" class="empty">
      <p class="empty-icon">🛒</p>
      <h2>Tu carrito está vacío</h2>
      <p>Agrega productos para continuar</p>
      <router-link to="/catalog" class="btn-shop">Ir al catálogo</router-link>
    </div>

    <!-- Cart items + summary -->
    <div v-else class="layout">

      <!-- Items -->
      <div class="items-col">
        <div v-for="item in cart.items" :key="item.product_id" class="cart-item">
          <div class="item-img">
            <img v-if="item.image_url" :src="item.image_url" :alt="item.name" />
            <span v-else>📦</span>
          </div>
          <div class="item-info">
            <p class="item-name">{{ item.name }}</p>
            <p class="item-sku">SKU: {{ item.sku }}</p>
            <p class="item-price">{{ fmt(item.unit_price) }} / u</p>
          </div>
          <div class="item-controls">
            <div class="qty-ctrl">
              <button @click="cart.qty(item.product_id, item.quantity - 1)">−</button>
              <span>{{ item.quantity }}</span>
              <button @click="cart.qty(item.product_id, item.quantity + 1)" :disabled="item.quantity >= item.stock">+</button>
            </div>
            <p class="item-total">{{ fmt(item.unit_price * item.quantity) }}</p>
            <button class="btn-remove" @click="cart.remove(item.product_id)">🗑</button>
          </div>
        </div>
      </div>

      <!-- Summary -->
      <div class="summary">
        <h2>Resumen</h2>

        <div class="summary-row">
          <span>Subtotal</span>
          <strong>{{ fmt(cart.subtotal) }}</strong>
        </div>
        <div v-if="cart.discount > 0" class="summary-row discount">
          <span>Descuento</span>
          <strong>−{{ fmt(cart.discount) }}</strong>
        </div>
        <hr />
        <div class="summary-row total">
          <span>Total</span>
          <strong>{{ fmt(cart.total) }}</strong>
        </div>

        <!-- Coupon -->
        <div class="coupon-section">
          <p class="coupon-label">¿Tienes un cupón?</p>
          <div class="coupon-row">
            <input v-model="couponCode" placeholder="Código" :disabled="couponOk" />
            <button :disabled="couponOk || applying" @click="applyCoupon">
              {{ applying ? '…' : 'Aplicar' }}
            </button>
          </div>
          <p v-if="couponOk"    class="msg-ok">✅ Cupón aplicado</p>
          <p v-if="couponError" class="msg-err">{{ couponError }}</p>
        </div>

        <button class="btn-checkout" @click="goCheckout">
          Proceder al pago →
        </button>

        <router-link to="/catalog" class="link-continue">← Seguir comprando</router-link>
      </div>

    </div>
  </div>
</template>

<style scoped>
.container { max-width: 1100px; margin: 0 auto; padding: 2rem 1.5rem; }
.title { font-size: 1.75rem; font-weight: 800; margin-bottom: 2rem; }

.empty { text-align: center; padding: 5rem 2rem; }
.empty-icon { font-size: 4rem; margin-bottom: 1rem; }
.empty h2  { font-size: 1.4rem; font-weight: 700; margin-bottom: .5rem; }
.empty p   { color: #94a3b8; margin-bottom: 1.5rem; }
.btn-shop {
  display: inline-block; background: #1d4ed8; color: #fff;
  padding: .75rem 2rem; border-radius: 10px; text-decoration: none; font-weight: 700;
}
.btn-shop:hover { background: #1e40af; }

.layout { display: grid; grid-template-columns: 1fr 360px; gap: 2rem; align-items: start; }

/* Items */
.items-col { display: flex; flex-direction: column; gap: 1rem; }
.cart-item {
  background: #fff; border-radius: 12px; padding: 1rem;
  display: flex; gap: 1rem; align-items: center;
  box-shadow: 0 1px 4px rgba(0,0,0,.06);
}
.item-img { width: 72px; height: 72px; border-radius: 8px; overflow: hidden; background: #f8fafc; display: flex; align-items: center; justify-content: center; flex-shrink: 0; font-size: 1.8rem; }
.item-img img { width: 100%; height: 100%; object-fit: cover; }
.item-info { flex: 1; min-width: 0; }
.item-name  { font-weight: 600; font-size: .95rem; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.item-sku   { font-size: .75rem; color: #94a3b8; margin-top: 2px; }
.item-price { font-size: .85rem; color: #64748b; margin-top: 4px; }
.item-controls { display: flex; flex-direction: column; align-items: flex-end; gap: .5rem; }
.qty-ctrl { display: flex; align-items: center; border: 1px solid #e2e8f0; border-radius: 8px; overflow: hidden; }
.qty-ctrl button { width: 30px; height: 32px; border: none; background: #f8fafc; cursor: pointer; font-size: 1rem; }
.qty-ctrl button:hover:not(:disabled) { background: #e2e8f0; }
.qty-ctrl button:disabled { color: #cbd5e1; cursor: not-allowed; }
.qty-ctrl span { width: 32px; text-align: center; font-weight: 600; font-size: .9rem; }
.item-total { font-weight: 700; font-size: 1rem; color: #1e293b; }
.btn-remove { background: none; border: none; font-size: 1.1rem; cursor: pointer; opacity: .5; }
.btn-remove:hover { opacity: 1; }

/* Summary */
.summary { background: #fff; border-radius: 16px; padding: 1.5rem; box-shadow: 0 2px 8px rgba(0,0,0,.08); position: sticky; top: 80px; }
.summary h2 { font-size: 1.15rem; font-weight: 700; margin-bottom: 1.25rem; }
.summary-row { display: flex; justify-content: space-between; font-size: .95rem; margin-bottom: .75rem; }
.summary-row.discount strong { color: #10b981; }
.summary-row.total { font-size: 1.15rem; margin-top: .5rem; }
hr { border: none; border-top: 1px solid #e2e8f0; margin: .75rem 0; }
.coupon-section { margin: 1.25rem 0; }
.coupon-label { font-size: .8rem; color: #94a3b8; margin-bottom: .4rem; }
.coupon-row { display: flex; gap: .5rem; }
.coupon-row input { flex: 1; padding: .45rem .75rem; border: 1px solid #e2e8f0; border-radius: 8px; font-size: .875rem; outline: none; }
.coupon-row button { padding: .45rem .9rem; background: #1d4ed8; color: #fff; border: none; border-radius: 8px; cursor: pointer; font-weight: 600; font-size: .875rem; }
.coupon-row button:disabled { background: #e2e8f0; color: #94a3b8; cursor: not-allowed; }
.msg-ok  { font-size: .8rem; color: #10b981; margin-top: .4rem; }
.msg-err { font-size: .8rem; color: #ef4444; margin-top: .4rem; }
.btn-checkout {
  width: 100%; padding: .9rem; background: #1d4ed8; color: #fff;
  border: none; border-radius: 12px; font-size: 1rem; font-weight: 700;
  cursor: pointer; transition: background .15s; margin-top: .5rem;
}
.btn-checkout:hover { background: #1e40af; }
.link-continue { display: block; text-align: center; margin-top: 1rem; color: #64748b; font-size: .875rem; text-decoration: none; }
.link-continue:hover { color: #3b82f6; }

@media (max-width: 768px) {
  .layout { grid-template-columns: 1fr; }
}
</style>
