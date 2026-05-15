<template>
  <div class="page">
    <div class="card">

      <div class="logo-wrap">
        <router-link to="/" class="logo-link">🛒</router-link>
      </div>

      <h1>Crear cuenta</h1>
      <p class="sub">Únete para hacer seguimiento de tus pedidos</p>

      <form @submit.prevent="register" class="form">
        <div class="row2">
          <div class="field">
            <label>Nombre</label>
            <input v-model="firstName" class="input" placeholder="Ana" required />
          </div>
          <div class="field">
            <label>Apellido</label>
            <input v-model="lastName" class="input" placeholder="García" required />
          </div>
        </div>

        <div class="field">
          <label>Email</label>
          <input v-model="email" type="email" class="input" placeholder="tu@email.com" required autocomplete="email" />
        </div>

        <div class="field">
          <label>Contraseña</label>
          <div class="pass-wrap">
            <input
              v-model="password" :type="showPass ? 'text' : 'password'"
              class="input" placeholder="Mínimo 8 caracteres" required minlength="8"
            />
            <button type="button" class="toggle-pass" @click="showPass = !showPass">
              {{ showPass ? '🙈' : '👁' }}
            </button>
          </div>
          <div class="strength" v-if="password">
            <div class="bar" :class="strengthClass" :style="{ width: strengthPct + '%' }" />
            <span>{{ strengthLabel }}</span>
          </div>
        </div>

        <p v-if="error" class="error-msg">⚠️ {{ error }}</p>

        <button type="submit" class="btn-register" :disabled="loading">
          <span v-if="loading" class="spinner" />
          {{ loading ? 'Creando cuenta…' : 'Crear cuenta' }}
        </button>
      </form>

      <p class="login-link">
        ¿Ya tienes cuenta? <router-link to="/login">Inicia sesión</router-link>
      </p>
      <p class="back-link">
        <router-link to="/">← Volver a la tienda</router-link>
      </p>

    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth   = useAuthStore()
const router = useRouter()
const route  = useRoute()

const firstName = ref('')
const lastName  = ref('')
const email     = ref('')
const password  = ref('')
const showPass  = ref(false)
const error     = ref('')
const loading   = ref(false)

const strengthPct = computed(() => {
  const p = password.value
  let s = 0
  if (p.length >= 8)  s += 25
  if (p.length >= 12) s += 15
  if (/[A-Z]/.test(p)) s += 20
  if (/[0-9]/.test(p)) s += 20
  if (/[^A-Za-z0-9]/.test(p)) s += 20
  return Math.min(s, 100)
})

const strengthClass = computed(() => {
  const v = strengthPct.value
  if (v < 40) return 'weak'
  if (v < 70) return 'ok'
  return 'strong'
})

const strengthLabel = computed(() => {
  const v = strengthPct.value
  if (v < 40) return 'Débil'
  if (v < 70) return 'Regular'
  return 'Fuerte'
})

async function register() {
  error.value = ''; loading.value = true
  try {
    await auth.register(email.value, password.value, firstName.value, lastName.value)
    const redirect = (route.query.redirect as string) || '/'
    router.push(redirect)
  } catch (e: any) {
    error.value = e.response?.data?.message ?? 'Error al crear la cuenta'
  }
  loading.value = false
}
</script>

<style scoped>
.page {
  min-height: 100vh; display: flex; align-items: center; justify-content: center;
  background: linear-gradient(135deg, #1e293b 0%, #1d4ed8 100%);
  padding: 2rem 1.5rem;
}
.card {
  background: #fff; border-radius: 20px; padding: 2.5rem 2rem;
  width: 100%; max-width: 440px; box-shadow: 0 8px 40px rgba(0,0,0,.25);
}
.logo-wrap { text-align: center; margin-bottom: 1.25rem; }
.logo-link { font-size: 2.5rem; text-decoration: none; }
h1  { font-size: 1.5rem; font-weight: 800; text-align: center; margin-bottom: .35rem; color: #1e293b; }
.sub { text-align: center; color: #94a3b8; margin-bottom: 2rem; font-size: .9rem; }
.form  { display: flex; flex-direction: column; gap: 1.15rem; }
.row2  { display: grid; grid-template-columns: 1fr 1fr; gap: .75rem; }
.field { display: flex; flex-direction: column; gap: .4rem; }
.field label { font-size: .78rem; font-weight: 700; color: #475569; text-transform: uppercase; letter-spacing: .05em; }
.input {
  width: 100%; padding: .7rem .9rem; border: 1px solid #e2e8f0; border-radius: 10px;
  font-size: .95rem; outline: none; color: #1e293b; transition: border-color .15s; font-family: inherit;
}
.input:focus { border-color: #3b82f6; box-shadow: 0 0 0 3px rgba(59,130,246,.15); }
.pass-wrap { position: relative; }
.pass-wrap .input { padding-right: 2.5rem; }
.toggle-pass {
  position: absolute; right: .75rem; top: 50%; transform: translateY(-50%);
  background: none; border: none; cursor: pointer; font-size: 1rem;
}
.strength { margin-top: .4rem; }
.bar { height: 4px; border-radius: 2px; transition: width .3s, background .3s; }
.bar.weak   { background: #ef4444; }
.bar.ok     { background: #f59e0b; }
.bar.strong { background: #10b981; }
.strength span { font-size: .72rem; color: #94a3b8; margin-top: 2px; display: block; }
.error-msg { color: #ef4444; font-size: .875rem; background: #fef2f2; padding: .6rem .9rem; border-radius: 8px; border: 1px solid #fecaca; }
.btn-register {
  width: 100%; padding: .8rem; background: #1d4ed8; color: #fff;
  border: none; border-radius: 12px; font-size: 1rem; font-weight: 700;
  cursor: pointer; transition: background .15s; display: flex; align-items: center; justify-content: center; gap: .5rem;
}
.btn-register:hover:not(:disabled) { background: #1e40af; }
.btn-register:disabled { background: #93c5fd; cursor: not-allowed; }
.spinner {
  width: 16px; height: 16px; border: 2px solid rgba(255,255,255,.4);
  border-top-color: #fff; border-radius: 50%; animation: spin .7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
.login-link { text-align: center; margin-top: 1.25rem; font-size: .875rem; color: #64748b; }
.login-link a { color: #3b82f6; font-weight: 600; text-decoration: none; }
.login-link a:hover { text-decoration: underline; }
.back-link  { text-align: center; margin-top: .75rem; font-size: .8rem; }
.back-link a { color: #94a3b8; text-decoration: none; }
.back-link a:hover { color: #3b82f6; }
</style>
