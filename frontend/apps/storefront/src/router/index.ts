import { createRouter, createWebHistory } from 'vue-router'
export const router = createRouter({
  history: createWebHistory(),
  scrollBehavior: () => ({ top: 0 }),
  routes: [
    { path: '/',               name: 'Home',     component: () => import('../views/HomeView.vue') },
    { path: '/catalog',        name: 'Catalog',  component: () => import('../views/CatalogView.vue') },
    { path: '/product/:id',    name: 'Product',  component: () => import('../views/ProductView.vue') },
    { path: '/cart',           name: 'Cart',     component: () => import('../views/CartView.vue') },
    { path: '/checkout',       name: 'Checkout', component: () => import('../views/CheckoutView.vue'), meta: { requiresAuth: true } },
    { path: '/orders',         name: 'Orders',      component: () => import('../views/OrdersView.vue'),       meta: { requiresAuth: true } },
    { path: '/orders/:id',     name: 'OrderDetail', component: () => import('../views/OrderDetailView.vue'),  meta: { requiresAuth: true } },
    { path: '/orders/:id/ok',  name: 'Success',     component: () => import('../views/SuccessView.vue') },
    { path: '/profile',        name: 'Profile',  component: () => import('../views/ProfileView.vue'),  meta: { requiresAuth: true } },
    { path: '/quote/:id',      name: 'Quote',    component: () => import('../views/QuoteView.vue') },
    { path: '/login',          name: 'Login',    component: () => import('../views/LoginView.vue') },
    { path: '/register',       name: 'Register', component: () => import('../views/RegisterView.vue') },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})
router.beforeEach((to, _, next) => {
  if (to.meta.requiresAuth && !localStorage.getItem('token')) return next({ name: 'Login', query: { redirect: to.fullPath } })
  next()
})
