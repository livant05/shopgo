<template>
  <div class="shell">
    <!-- Sidebar -->
    <aside class="sidebar" :class="{ collapsed }">
      <div class="sb-header">
        <span class="sb-logo" v-if="!collapsed">🛒 {{ storeName }}</span>
        <button @click="collapsed = !collapsed" class="sb-toggle">{{ collapsed ? '→' : '←' }}</button>
      </div>

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
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth     = useAuthStore()
const router   = useRouter()
const collapsed = ref(false)
const storeName = import.meta.env.VITE_STORE_NAME ?? 'Mi Tienda'

const navItems = [
  { to: '/',          icon: '📊', label: 'Dashboard',  minRole: 'staff'   },
  { to: '/orders',    icon: '🛒', label: 'Órdenes',    minRole: 'staff'   },
  { to: '/pos',       icon: '💳', label: 'POS',        minRole: 'staff'   },
  { to: '/products',  icon: '📦', label: 'Productos',  minRole: 'manager' },
  { to: '/inventory', icon: '📋', label: 'Inventario', minRole: 'manager' },
  { to: '/branches',  icon: '🏪', label: 'Sucursales', minRole: 'admin'   },
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
.sb-nav { flex: 1; padding: 12px 8px; display: flex; flex-direction: column; gap: 3px; overflow-y: auto; }
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
</style>
