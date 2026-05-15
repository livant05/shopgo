<template>
  <div class="shell">
    <!-- Sidebar -->
    <aside class="sidebar" :class="{ collapsed }">
      <div class="sb-header">
        <span class="sb-logo" v-if="!collapsed">🛒 {{ storeName }}</span>
        <button @click="collapsed = !collapsed" class="sb-toggle">{{ collapsed ? '→' : '←' }}</button>
      </div>

      <button class="sb-search" @click="cmdPalette.show()" :title="collapsed ? 'Buscar (⌘K)' : ''">
        <span class="sb-icon">🔍</span>
        <span v-if="!collapsed" class="sb-search-label">Buscar…</span>
        <kbd v-if="!collapsed" class="sb-search-kbd">⌘K</kbd>
      </button>

      <nav class="sb-nav">
        <router-link v-for="item in navItems" :key="item.to" :to="item.to"
          class="sb-link" :class="{ 'sb-link-hidden': !auth.hasRole(item.minRole) }">
          <span class="sb-icon">{{ item.icon }}</span>
          <span v-if="!collapsed" class="sb-label">{{ item.label }}</span>
        </router-link>
      </nav>

      <div class="sb-footer">
        <div class="sb-user" v-if="!collapsed">
          <div class="sb-user-name">{{ auth.fullName }}</div>
          <div class="sb-user-role">{{ auth.user?.role }}</div>
        </div>
        <button @click="auth.logout(); router.push('/login')" class="sb-logout" title="Cerrar sesión">🚪</button>
      </div>
    </aside>

    <!-- Main content -->
    <main class="content">
      <router-view v-slot="{ Component }">
        <Transition name="fade" mode="out-in">
          <component :is="Component" />
        </Transition>
      </router-view>
    </main>

    <!-- Command Palette -->
    <CommandPalette />

    <!-- Notification toasts -->
    <div class="toast-stack">
      <TransitionGroup name="toast">
        <div v-for="n in notifications" :key="n.id" class="toast" @click="dismiss(n.id)">
          <span class="toast-icon">🛒</span>
          <div class="toast-body">
            <p class="toast-title">Nuevo pedido</p>
            <p class="toast-msg">{{ n.message }}</p>
          </div>
          <button class="toast-close">✕</button>
        </div>
      </TransitionGroup>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useNotifications } from '../composables/useNotifications'
import { useCommandPalette } from '../composables/useCommandPalette'
import CommandPalette from '../components/CommandPalette.vue'

const auth      = useAuthStore()
const router    = useRouter()
const collapsed = ref(false)
const storeName = import.meta.env.VITE_STORE_NAME ?? 'Mi Tienda'

const { notifications, connect, dismiss } = useNotifications()
const cmdPalette = useCommandPalette()

onMounted(() => {
  if (auth.hasRole('admin')) connect()
})

const navItems = [
  { to: '/',          icon: '📊', label: 'Dashboard',  minRole: 'staff'   },
  { to: '/orders',    icon: '🛒', label: 'Órdenes',    minRole: 'staff'   },
  { to: '/pos',       icon: '💳', label: 'POS',        minRole: 'staff'   },
  { to: '/products',  icon: '📦', label: 'Productos',  minRole: 'manager' },
  { to: '/inventory',  icon: '📋', label: 'Inventario',  minRole: 'manager' },
  { to: '/categories', icon: '🗂️', label: 'Categorías', minRole: 'admin'   },
  { to: '/branches',   icon: '🏪', label: 'Sucursales',  minRole: 'admin'   },
  { to: '/users',     icon: '👥', label: 'Usuarios',   minRole: 'admin'   },
  { to: '/reports',   icon: '📈', label: 'Reportes',   minRole: 'admin'   },
  { to: '/coupons',   icon: '🏷️', label: 'Cupones',    minRole: 'admin'   },
  { to: '/settings',  icon: '⚙️', label: 'Config',     minRole: 'admin'   },
]
</script>

<style scoped>
.shell    { display: flex; min-height: 100vh; background: #080c14; }
.sidebar  { width: 220px; background: #0f1623; border-right: 1px solid #253047; display: flex; flex-direction: column; transition: width .2s; position: sticky; top: 0; height: 100vh; overflow: hidden; flex-shrink: 0; }
.sidebar.collapsed { width: 56px; }
.sb-header { display: flex; align-items: center; justify-content: space-between; padding: 16px 12px; border-bottom: 1px solid #253047; min-height: 56px; }
.sb-logo { font-size: 14px; font-weight: 700; color: #38bdf8; white-space: nowrap; overflow: hidden; }
.sb-toggle { background: none; border: 1px solid #253047; color: #5a7298; border-radius: 5px; width: 26px; height: 26px; cursor: pointer; font-size: 11px; flex-shrink: 0; }
.sb-search { display:flex; align-items:center; gap:8px; margin:8px; padding:8px 10px; background:#0f1623; border:1px solid #253047; border-radius:7px; cursor:pointer; color:#5a6a87; font-size:13px; transition:all .15s; }
.sb-search:hover { border-color:#38bdf8; color:#38bdf8; }
.sb-search-label { flex:1; text-align:left; }
.sb-search-kbd { background:#253047; border:1px solid #344762; border-radius:4px; padding:1px 5px; font-size:10px; color:#8494ac; }
.sb-nav { flex: 1; padding: 8px 8px; display: flex; flex-direction: column; gap: 3px; overflow-y: auto; }
.sb-link { display: flex; align-items: center; gap: 10px; padding: 9px 10px; border-radius: 7px; text-decoration: none; color: #5a7298; font-size: 13px; transition: all .15s; white-space: nowrap; }
.sb-link:hover, .sb-link.router-link-active { background: rgba(56,189,248,.08); color: #38bdf8; }
.sb-link.router-link-exact-active { background: rgba(56,189,248,.12); color: #38bdf8; }
.sb-link-hidden { display: none; }
.sb-icon { font-size: 16px; flex-shrink: 0; }
.sb-footer { padding: 12px; border-top: 1px solid #253047; display: flex; align-items: center; gap: 8px; }
.sb-user { flex: 1; min-width: 0; }
.sb-user-name { font-size: 12px; font-weight: 600; color: #d6dfe8; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.sb-user-role { font-size: 10px; color: #5a7298; text-transform: uppercase; letter-spacing: 1px; }
.sb-logout { background: none; border: none; cursor: pointer; font-size: 16px; padding: 4px; flex-shrink: 0; }
.content { flex: 1; padding: 32px 36px; overflow-x: hidden; }
.fade-enter-active, .fade-leave-active { transition: opacity .2s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }

/* Toast notifications */
.toast-stack {
  position: fixed; bottom: 24px; right: 24px;
  display: flex; flex-direction: column; gap: 10px; z-index: 9999;
}
.toast {
  display: flex; align-items: flex-start; gap: 12px;
  background: #1c2333; border: 1px solid #2d3a52;
  border-left: 3px solid #38bdf8;
  border-radius: 10px; padding: 14px 16px;
  width: 320px; box-shadow: 0 8px 24px rgba(0,0,0,.4);
  cursor: pointer;
}
.toast-icon { font-size: 20px; flex-shrink: 0; line-height: 1; }
.toast-body { flex: 1; min-width: 0; }
.toast-title { font-size: 12px; font-weight: 700; color: #38bdf8; margin: 0 0 2px; text-transform: uppercase; letter-spacing: .5px; }
.toast-msg   { font-size: 13px; color: #d6dfe8; margin: 0; }
.toast-close { background: none; border: none; color: #5a6a87; cursor: pointer; font-size: 14px; padding: 0; flex-shrink: 0; }
.toast-close:hover { color: #d6dfe8; }

.toast-enter-active { transition: all .3s ease; }
.toast-leave-active { transition: all .25s ease; }
.toast-enter-from   { opacity: 0; transform: translateX(40px); }
.toast-leave-to     { opacity: 0; transform: translateX(40px); }
</style>
