<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth   = useAuthStore()
const router = useRouter()
const route  = useRoute()

const email    = ref('')
const password = ref('')
const error    = ref('')
const loading  = ref(false)
const showPass = ref(false)

async function login() {
  if (!email.value || !password.value) { error.value = 'Completa todos los campos'; return }
  error.value = ''; loading.value = true
  try {
    await auth.login(email.value, password.value)
    const redirect = (route.query.redirect as string) || '/'
    router.push(redirect)
  } catch (e: any) {
    error.value = e.response?.data?.message ?? 'Credenciales incorrectas'
  }
  loading.value = false
}
</script>

<template>
  <div class="page">
    <div class="card">

      <div class="logo-wrap">
        <router-link to="/" class="logo-link">🛒</router-link>
      </div>

      <h1>Iniciar sesión</h1>
      <p class="sub">Accede a tu cuenta para continuar</p>

      <form @submit.prevent="login" class="form">
        <div class="field">
          <label>Email</label>
          <input
            v-model="email" type="email" class="input"
            placeholder="tu@email.com" autocomplete="email" required
          />
        </div>

        <div class="field">
          <label>Contraseña</label>
          <div class="pass-wrap">
            <input
              v-model="password" :type="showPass ? 'text' : 'password'"
              class="input" placeholder="••••••••" required
            />
            <button type="button" class="toggle-pass" @click="showPass = !showPass">
              {{ showPass ? '🙈' : '👁' }}
            </button>
          </div>
        </div>

        <p v-if="error" class="error-msg">⚠️ {{ error }}</p>

        <button type="submit" class="btn-login" :disabled="loading">
          <span v-if="loading" class="spinner" />
          {{ loading ? 'Ingresando…' : 'Iniciar sesión' }}
        </button>
      </form>

      <p class="back-link">
        <router-link to="/">← Volver a la tienda</router-link>
      </p>

    </div>
  </div>
</template>

<style scoped>
.page {
  min-height: 100vh; display: flex; align-items: center; justify-content: center;
  background: linear-gradient(135deg, #1e293b 0%, #1d4ed8 100%);
  padding: 2rem 1.5rem;
}
.card {
  background: #fff; border-radius: 20px; padding: 2.5rem 2rem;
  width: 100%; max-width: 400px; box-shadow: 0 8px 40px rgba(0,0,0,.25);
}
.logo-wrap { text-align: center; margin-bottom: 1.25rem; }
.logo-link { font-size: 2.5rem; text-decoration: none; }
h1 { font-size: 1.5rem; font-weight: 800; text-align: center; margin-bottom: .35rem; color: #1e293b; }
.sub { text-align: center; color: #94a3b8; margin-bottom: 2rem; font-size: .9rem; }
.form { display: flex; flex-direction: column; gap: 1.25rem; }
.field { display: flex; flex-direction: column; gap: .4rem; }
.field label { font-size: .8rem; font-weight: 700; color: #475569; text-transform: uppercase; letter-spacing: .05em; }
.input {
  width: 100%; padding: .7rem .9rem; border: 1px solid #e2e8f0; border-radius: 10px;
  font-size: .95rem; outline: none; color: #1e293b; transition: border-color .15s; font-family: inherit;
}
.input:focus { border-color: #3b82f6; box-shadow: 0 0 0 3px rgba(59,130,246,.15); }
.pass-wrap { position: relative; }
.pass-wrap .input { padding-right: 2.5rem; }
.toggle-pass {
  position: absolute; right: .75rem; top: 50%; transform: translateY(-50%);
  background: none; border: none; cursor: pointer; font-size: 1rem; line-height: 1;
}
.error-msg { color: #ef4444; font-size: .875rem; background: #fef2f2; padding: .6rem .9rem; border-radius: 8px; border: 1px solid #fecaca; }
.btn-login {
  width: 100%; padding: .8rem; background: #1d4ed8; color: #fff;
  border: none; border-radius: 12px; font-size: 1rem; font-weight: 700;
  cursor: pointer; transition: background .15s; display: flex; align-items: center; justify-content: center; gap: .5rem;
}
.btn-login:hover:not(:disabled) { background: #1e40af; }
.btn-login:disabled { background: #93c5fd; cursor: not-allowed; }
.spinner {
  width: 16px; height: 16px; border: 2px solid rgba(255,255,255,.4);
  border-top-color: #fff; border-radius: 50%; animation: spin .7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
.back-link { text-align: center; margin-top: 1.5rem; font-size: .875rem; }
.back-link a { color: #3b82f6; text-decoration: none; }
.back-link a:hover { text-decoration: underline; }
</style>
