<template>
  <div class="login-shell">
    <div class="login-card">
      <div class="login-logo">🛒</div>
      <h1 class="login-title">{{ storeName }}</h1>
      <p class="login-sub">Panel de administración</p>

      <form @submit.prevent="handleLogin" class="login-form">
        <!-- Paso 1: Credenciales -->
        <template v-if="step === 1">
          <div class="field">
            <label>Email</label>
            <input v-model="email" type="email" placeholder="admin@tienda.com" required autocomplete="email" />
          </div>
          <div class="field">
            <label>Contraseña</label>
            <input v-model="password" type="password" placeholder="••••••••" required autocomplete="current-password" />
          </div>
        </template>

        <!-- Paso 2: 2FA -->
        <template v-else>
          <p class="mfa-prompt">Ingresa el código de tu app autenticadora</p>
          <div class="field">
            <label>Código 2FA</label>
            <input v-model="totpCode" placeholder="123456" maxlength="6" pattern="[0-9]{6}"
              autocomplete="one-time-code" class="mfa-input" ref="totpRef" />
          </div>
          <button type="button" @click="step = 1; totpCode=''" class="btn-ghost">← Atrás</button>
        </template>

        <div v-if="error" class="login-error">{{ error }}</div>

        <button type="submit" class="btn-primary" :disabled="loading">
          <span v-if="loading" class="spinner" />
          <span v-else>{{ step === 1 ? 'Iniciar sesión' : 'Verificar' }}</span>
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const route  = useRoute()
const auth   = useAuthStore()

const storeName = import.meta.env.VITE_STORE_NAME ?? 'Mi Tienda'
const email    = ref('')
const password = ref('')
const totpCode = ref('')
const step     = ref(1)
const loading  = ref(false)
const error    = ref('')
const totpRef  = ref<HTMLInputElement>()

async function handleLogin() {
  loading.value = true
  error.value   = ''
  try {
    await auth.login(email.value, password.value, step.value === 2 ? totpCode.value : undefined)
    const redirect = (route.query.redirect as string) ?? '/'
    router.push(redirect)
  } catch (e: any) {
    const code = e.response?.data?.code
    if (code === 'MFA_REQUIRED') {
      step.value = 2
      await nextTick()
      totpRef.value?.focus()
    } else {
      error.value = e.response?.data?.message ?? 'Error al iniciar sesión'
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-shell { min-height: 100vh; display: flex; align-items: center; justify-content: center; background: #080c14; padding: 24px; }
.login-card  { background: #0f1623; border: 1px solid #253047; border-radius: 14px; padding: 48px 40px; width: 100%; max-width: 380px; text-align: center; }
.login-logo  { font-size: 48px; margin-bottom: 12px; }
.login-title { font-size: 20px; font-weight: 700; color: #f1f5f9; margin-bottom: 4px; }
.login-sub   { font-size: 13px; color: #5a7298; margin-bottom: 32px; }
.login-form  { display: flex; flex-direction: column; gap: 16px; text-align: left; }
.field       { display: flex; flex-direction: column; gap: 6px; }
.field label { font-size: 12px; color: #8494ac; text-transform: uppercase; letter-spacing: .5px; }
.field input { background: #080c14; border: 1px solid #253047; border-radius: 8px; color: #e2e8f0; padding: 10px 14px; font-size: 14px; width: 100%; }
.field input:focus { outline: none; border-color: #38bdf8; }
.mfa-prompt  { font-size: 13px; color: #7a95b8; text-align: center; margin-bottom: 4px; }
.mfa-input   { font-size: 22px; letter-spacing: 8px; text-align: center; font-family: monospace; }
.login-error { background: rgba(248,113,113,.1); border: 1px solid rgba(248,113,113,.3); color: #f87171; padding: 10px; border-radius: 7px; font-size: 13px; }
.btn-primary { background: #38bdf8; color: #080c14; border: none; padding: 12px; border-radius: 8px; font-size: 14px; font-weight: 700; cursor: pointer; display: flex; align-items: center; justify-content: center; gap: 8px; }
.btn-primary:disabled { opacity: .5; cursor: not-allowed; }
.btn-ghost   { background: none; border: 1px solid #253047; color: #5a7298; padding: 10px; border-radius: 8px; font-size: 13px; cursor: pointer; }
.spinner     { width: 16px; height: 16px; border: 2px solid rgba(8,12,20,.3); border-top-color: #080c14; border-radius: 50%; animation: spin .7s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
