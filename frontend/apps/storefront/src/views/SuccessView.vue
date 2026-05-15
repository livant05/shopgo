<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '../api/client'
import type { Order } from '../types'

const route = useRoute()
const order = ref<Order | null>(null)

const fmt = (n: number) => n.toLocaleString('es-MX', { style: 'currency', currency: 'MXN' })

const statusLabel: Record<string, string> = {
  pending:    'Pendiente',
  confirmed:  'Confirmado',
  processing: 'En proceso',
  shipped:    'Enviado',
  delivered:  'Entregado',
  cancelled:  'Cancelado',
}

onMounted(async () => {
  const id = route.params.id as string
  if (id && id !== 'ok') {
    try {
      const { data } = await api.get(`/orders/${id}`)
      order.value = data
    } catch {}
  }
})
</script>

<template>
  <div class="page">
    <div class="card">
      <div class="icon-wrap">✅</div>
      <h1>¡Pedido confirmado!</h1>
      <p class="sub">Gracias por tu compra. Recibirás una confirmación pronto.</p>

      <div v-if="order" class="order-detail">
        <div class="detail-grid">
          <div class="detail-item">
            <span class="label">Número de pedido</span>
            <strong class="mono">{{ order.id.slice(0, 8).toUpperCase() }}</strong>
          </div>
          <div class="detail-item">
            <span class="label">Estado</span>
            <span class="status-badge">{{ statusLabel[order.status] ?? order.status }}</span>
          </div>
          <div class="detail-item">
            <span class="label">Total</span>
            <strong>{{ fmt(order.total) }}</strong>
          </div>
          <div class="detail-item">
            <span class="label">Fecha</span>
            <span>{{ new Date(order.created_at).toLocaleDateString('es-MX') }}</span>
          </div>
        </div>

        <div v-if="order.items?.length" class="items-section">
          <h3>Artículos</h3>
          <div v-for="item in order.items" :key="item.product_id" class="order-item">
            <div>
              <p class="item-name">{{ item.name }}</p>
              <p class="item-meta">{{ item.sku }} · × {{ item.quantity }}</p>
            </div>
            <span class="item-total">{{ fmt(item.line_total) }}</span>
          </div>
        </div>
      </div>

      <div class="actions">
        <router-link to="/" class="btn-primary">Volver al inicio</router-link>
        <router-link to="/catalog" class="btn-secondary">Seguir comprando</router-link>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page {
  min-height: 80vh; display: flex; align-items: center; justify-content: center;
  padding: 2rem 1.5rem; background: #f1f5f9;
}
.card {
  background: #fff; border-radius: 20px; padding: 3rem 2.5rem;
  max-width: 600px; width: 100%; text-align: center;
  box-shadow: 0 4px 24px rgba(0,0,0,.1);
}
.icon-wrap { font-size: 4rem; margin-bottom: 1rem; }
h1 { font-size: 1.75rem; font-weight: 800; margin-bottom: .5rem; color: #1e293b; }
.sub { color: #64748b; margin-bottom: 2rem; font-size: 1rem; }

.order-detail { text-align: left; background: #f8fafc; border-radius: 12px; padding: 1.5rem; margin-bottom: 2rem; }
.detail-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; margin-bottom: 1.5rem; }
.detail-item { display: flex; flex-direction: column; gap: .25rem; }
.label { font-size: .75rem; color: #94a3b8; text-transform: uppercase; letter-spacing: .05em; font-weight: 600; }
.mono { font-family: monospace; font-size: 1rem; letter-spacing: .08em; }
.status-badge { display: inline-block; background: #dbeafe; color: #1d4ed8; border-radius: 999px; padding: .2rem .75rem; font-size: .8rem; font-weight: 600; }

.items-section h3 { font-size: .9rem; font-weight: 700; color: #475569; margin-bottom: .75rem; text-transform: uppercase; letter-spacing: .04em; }
.order-item { display: flex; justify-content: space-between; align-items: center; padding: .6rem 0; border-bottom: 1px solid #e2e8f0; }
.order-item:last-child { border-bottom: none; }
.item-name { font-size: .9rem; font-weight: 600; }
.item-meta { font-size: .78rem; color: #94a3b8; }
.item-total { font-weight: 700; font-size: .9rem; }

.actions { display: flex; gap: 1rem; justify-content: center; flex-wrap: wrap; }
.btn-primary {
  background: #1d4ed8; color: #fff; padding: .75rem 2rem;
  border-radius: 10px; text-decoration: none; font-weight: 700;
}
.btn-primary:hover { background: #1e40af; }
.btn-secondary {
  background: #f1f5f9; color: #475569; padding: .75rem 2rem;
  border-radius: 10px; text-decoration: none; font-weight: 600;
}
.btn-secondary:hover { background: #e2e8f0; }
</style>
