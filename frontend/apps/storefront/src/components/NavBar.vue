<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useCartStore } from '../stores/cart'
import { api } from '../api/client'

const router = useRouter()
const auth   = useAuthStore()
const cart   = useCartStore()

const storeName  = ref('Mi Tienda')
const menuOpen   = ref(false)
const userOpen   = ref(false)
const search     = ref('')

onMounted(async () => {
  try {
    const { data } = await api.get('/store')
    if (data.store_name) storeName.value = data.store_name
  } catch {}
})

function goSearch() {
  if (search.value.trim())
    router.push({ name: 'Catalog', query: { q: search.value.trim() } })
}

function logout() {
  auth.logout()
  userOpen.value = false
  router.push({ name: 'Home' })
}
</script>

<template>
  <header class="nav">
    <div class="nav-inner">
      <!-- Logo -->
      <router-link to="/" class="logo">🛒 {{ storeName }}</router-link>

      <!-- Search -->
      <form class="search-bar" @submit.prevent="goSearch">
        <input v-model="search" placeholder="Buscar productos…" />
        <button type="submit">🔍</button>
      </form>

      <!-- Actions -->
      <nav class="actions">
        <router-link to="/catalog" class="nav-link">Catálogo</router-link>

        <!-- Cart -->
        <router-link to="/cart" class="cart-btn">
          🛒
          <span v-if="cart.count > 0" class="badge">{{ cart.count }}</span>
        </router-link>

        <!-- Auth -->
        <div v-if="auth.isAuth" class="user-menu">
          <button class="user-btn" @click="userOpen = !userOpen">
            👤 {{ auth.user?.first_name || 'Mi cuenta' }} ▾
          </button>
          <div v-if="userOpen" class="dropdown">
            <router-link to="/orders" class="dd-item" @click="userOpen = false">📦 Mis pedidos</router-link>
            <button class="dd-item danger" @click="logout">Cerrar sesión</button>
          </div>
        </div>
        <div v-else class="auth-links">
          <router-link to="/login"    class="btn-login">Iniciar sesión</router-link>
          <router-link to="/register" class="btn-register">Registrarse</router-link>
        </div>
      </nav>

      <!-- Hamburger -->
      <button class="hamburger" @click="menuOpen = !menuOpen">☰</button>
    </div>

    <!-- Mobile menu -->
    <div v-if="menuOpen" class="mobile-menu">
      <router-link to="/catalog" @click="menuOpen = false">Catálogo</router-link>
      <router-link to="/cart"    @click="menuOpen = false">Carrito ({{ cart.count }})</router-link>
      <template v-if="auth.isAuth">
        <router-link to="/orders" @click="menuOpen = false">Mis pedidos</router-link>
        <button @click="logout">Cerrar sesión</button>
      </template>
      <template v-else>
        <router-link to="/login"    @click="menuOpen = false">Iniciar sesión</router-link>
        <router-link to="/register" @click="menuOpen = false">Registrarse</router-link>
      </template>
    </div>
  </header>
</template>

<style scoped>
.nav {
  position: sticky; top: 0; z-index: 100;
  background: #1e293b; color: #f1f5f9;
  box-shadow: 0 2px 8px rgba(0,0,0,.3);
}
.nav-inner {
  max-width: 1200px; margin: 0 auto;
  display: flex; align-items: center; gap: 1rem; padding: .75rem 1.5rem;
}
.logo {
  font-size: 1.2rem; font-weight: 700; color: #fff;
  text-decoration: none; white-space: nowrap;
}
.search-bar {
  flex: 1; display: flex; max-width: 480px;
}
.search-bar input {
  flex: 1; padding: .45rem .75rem; border: none; border-radius: 6px 0 0 6px;
  font-size: .9rem; outline: none; background: #334155; color: #f1f5f9;
}
.search-bar input::placeholder { color: #94a3b8; }
.search-bar button {
  padding: .45rem .8rem; background: #3b82f6; border: none;
  border-radius: 0 6px 6px 0; cursor: pointer; font-size: .9rem;
}
.search-bar button:hover { background: #2563eb; }
.actions {
  display: flex; align-items: center; gap: 1rem;
}
.nav-link {
  color: #cbd5e1; text-decoration: none; font-size: .9rem;
}
.nav-link:hover { color: #fff; }
.cart-btn {
  position: relative; font-size: 1.3rem; text-decoration: none;
  display: flex; align-items: center;
}
.badge {
  position: absolute; top: -8px; right: -10px;
  background: #ef4444; color: #fff; border-radius: 999px;
  font-size: .65rem; font-weight: 700; padding: 1px 5px; min-width: 18px;
  text-align: center;
}
.user-menu { position: relative; }
.user-btn {
  background: none; border: 1px solid #475569; color: #f1f5f9;
  padding: .35rem .75rem; border-radius: 6px; cursor: pointer; font-size: .85rem;
}
.user-btn:hover { background: #334155; }
.dropdown {
  position: absolute; right: 0; top: calc(100% + 6px);
  background: #1e293b; border: 1px solid #334155; border-radius: 8px;
  min-width: 160px; overflow: hidden; box-shadow: 0 4px 16px rgba(0,0,0,.3);
}
.dd-item {
  display: block; padding: .65rem 1rem; color: #cbd5e1;
  text-decoration: none; font-size: .875rem;
  background: none; border: none; width: 100%; text-align: left; cursor: pointer;
}
.dd-item:hover { background: #334155; color: #fff; }
.dd-item.danger:hover { color: #f87171; }
.auth-links { display: flex; align-items: center; gap: .5rem; }
.btn-login {
  background: transparent; color: #cbd5e1; border: 1px solid #475569;
  padding: .35rem .9rem; border-radius: 6px; text-decoration: none; font-size: .875rem;
}
.btn-login:hover { background: #334155; color: #fff; }
.btn-register {
  background: #3b82f6; color: #fff; padding: .4rem .9rem;
  border-radius: 6px; text-decoration: none; font-size: .875rem;
}
.btn-register:hover { background: #2563eb; }
.hamburger {
  display: none; background: none; border: none; color: #f1f5f9;
  font-size: 1.4rem; cursor: pointer;
}
.mobile-menu {
  display: none; flex-direction: column; padding: .5rem 1.5rem 1rem;
  border-top: 1px solid #334155;
}
.mobile-menu a, .mobile-menu button {
  padding: .6rem 0; color: #cbd5e1; text-decoration: none;
  background: none; border: none; text-align: left; font-size: .95rem; cursor: pointer;
}
@media (max-width: 768px) {
  .search-bar { display: none; }
  .actions    { display: none; }
  .hamburger  { display: block; margin-left: auto; }
  .mobile-menu { display: flex; }
}
</style>
