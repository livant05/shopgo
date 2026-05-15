import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const LEVELS: Record<string,number> = { admin:100, manager:60, staff:40, customer:10 }

export const router = createRouter({
  history: createWebHistory(),
  scrollBehavior: () => ({ top: 0 }),
  routes: [
    { path: '/login', name: 'Login', component: () => import('../views/LoginView.vue'), meta: { public: true } },
    {
      path: '/',
      component: () => import('../layouts/DashboardLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        { path: '',          name: 'Dashboard', component: () => import('../views/DashboardView.vue') },
        { path: 'orders',    name: 'Orders',    component: () => import('../views/Orders/OrderListView.vue'), meta: { minRole: 'staff' } },
        { path: 'pos',       name: 'POS',       component: () => import('../views/POS/POSView.vue'),          meta: { minRole: 'staff' } },
        { path: 'products',  name: 'Products',  component: () => import('../views/Products/ProductListView.vue'), meta: { minRole: 'manager' } },
        { path: 'inventory', name: 'Inventory', component: () => import('../views/Inventory/InventoryView.vue'), meta: { minRole: 'manager' } },
        { path: 'branches',  name: 'Branches',  component: () => import('../views/Branches/BranchListView.vue'), meta: { minRole: 'admin' } },
        { path: 'users',     name: 'Users',     component: () => import('../views/Users/UsersView.vue'),         meta: { minRole: 'admin' } },
        { path: 'reports',   name: 'Reports',   component: () => import('../views/Reports/ReportsView.vue'),     meta: { minRole: 'admin' } },
        { path: 'settings',  name: 'Settings',  component: () => import('../views/Settings/SettingsView.vue'),   meta: { minRole: 'admin' } },
      ],
    },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

router.beforeEach(async (to, _, next) => {
  const auth = useAuthStore()
  if (to.meta.public) return next()
  if (!auth.token) return next({ name: 'Login', query: { redirect: to.fullPath } })
  if (!auth.user) {
    try { await auth.me() } catch { auth.logout(); return next({ name: 'Login' }) }
  }
  const min = to.meta.minRole as string | undefined
  if (min && (LEVELS[auth.user?.role ?? ''] ?? 0) < (LEVELS[min] ?? 0)) return next({ path: '/' })
  next()
})
