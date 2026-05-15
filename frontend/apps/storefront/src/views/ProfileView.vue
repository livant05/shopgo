<template>
  <div class="page">
    <div class="container">
      <div class="profile-header">
        <div class="avatar">{{ initials }}</div>
        <div>
          <h1 class="profile-name">{{ auth.fullName }}</h1>
          <p class="profile-email">{{ auth.user?.email }}</p>
        </div>
      </div>

      <!-- Tabs -->
      <div class="tabs">
        <button class="tab" :class="{ active: tab === 'info' }"    @click="tab = 'info'">👤 Información</button>
        <button class="tab" :class="{ active: tab === 'address' }" @click="tab = 'address'">📍 Dirección</button>
        <button class="tab" :class="{ active: tab === 'security' }" @click="tab = 'security'">🔒 Seguridad</button>
      </div>

      <!-- Info tab -->
      <div v-if="tab === 'info'" class="card">
        <h2 class="card-title">Información personal</h2>
        <form @submit.prevent="saveProfile" class="form">
          <div class="form-row">
            <div class="field">
              <label>Nombre *</label>
              <input v-model="profile.first_name" required placeholder="Juan" />
            </div>
            <div class="field">
              <label>Apellido</label>
              <input v-model="profile.last_name" placeholder="García" />
            </div>
          </div>
          <div class="field">
            <label>Teléfono</label>
            <input v-model="profile.phone" type="tel" placeholder="+52 55 1234 5678" />
          </div>
          <div class="field field--readonly">
            <label>Correo electrónico</label>
            <div class="readonly-value">{{ auth.user?.email }}</div>
            <p class="field-hint">El email no puede modificarse desde aquí.</p>
          </div>
          <div v-if="profileMsg" class="msg" :class="profileErr ? 'msg-err' : 'msg-ok'">{{ profileMsg }}</div>
          <div class="form-footer">
            <button type="submit" class="btn-primary" :disabled="profileLoading">
              {{ profileLoading ? 'Guardando…' : 'Guardar cambios' }}
            </button>
          </div>
        </form>
      </div>

      <!-- Address tab -->
      <div v-if="tab === 'address'" class="card">
        <h2 class="card-title">Dirección de envío predeterminada</h2>
        <p class="card-sub">Se usará automáticamente al hacer checkout.</p>
        <form @submit.prevent="saveProfile" class="form">
          <div class="field">
            <label>Calle y número</label>
            <input v-model="profile.default_address.street" placeholder="Av. Insurgentes 123" />
          </div>
          <div class="form-row">
            <div class="field">
              <label>Ciudad</label>
              <input v-model="profile.default_address.city" placeholder="Ciudad de México" />
            </div>
            <div class="field">
              <label>Estado</label>
              <input v-model="profile.default_address.state" placeholder="CDMX" />
            </div>
          </div>
          <div class="form-row">
            <div class="field">
              <label>Código postal</label>
              <input v-model="profile.default_address.zip" placeholder="06600" />
            </div>
            <div class="field">
              <label>País</label>
              <input v-model="profile.default_address.country" placeholder="México" />
            </div>
          </div>
          <div v-if="profileMsg" class="msg" :class="profileErr ? 'msg-err' : 'msg-ok'">{{ profileMsg }}</div>
          <div class="form-footer">
            <button type="submit" class="btn-primary" :disabled="profileLoading">
              {{ profileLoading ? 'Guardando…' : 'Guardar dirección' }}
            </button>
          </div>
        </form>
      </div>

      <!-- Security tab -->
      <div v-if="tab === 'security'" class="card">
        <h2 class="card-title">Cambiar contraseña</h2>
        <form @submit.prevent="savePassword" class="form">
          <div class="field">
            <label>Contraseña actual *</label>
            <div class="pass-wrap">
              <input v-model="passwords.current" :type="showCurrent ? 'text' : 'password'" required placeholder="••••••••" />
              <button type="button" class="eye-btn" @click="showCurrent = !showCurrent">{{ showCurrent ? '🙈' : '👁' }}</button>
            </div>
          </div>
          <div class="field">
            <label>Nueva contraseña *</label>
            <div class="pass-wrap">
              <input v-model="passwords.new" :type="showNew ? 'text' : 'password'" required minlength="8" placeholder="Mín. 8 caracteres" />
              <button type="button" class="eye-btn" @click="showNew = !showNew">{{ showNew ? '🙈' : '👁' }}</button>
            </div>
            <div class="strength-bar">
              <div class="strength-fill" :class="strengthClass" :style="{ width: strengthPct + '%' }"></div>
            </div>
            <p class="strength-label" :class="strengthClass">{{ strengthLabel }}</p>
          </div>
          <div class="field">
            <label>Confirmar contraseña *</label>
            <input v-model="passwords.confirm" type="password" required placeholder="Repetir nueva contraseña" />
            <p v-if="passwords.confirm && passwords.new !== passwords.confirm" class="field-hint err">Las contraseñas no coinciden</p>
          </div>
          <div v-if="passMsg" class="msg" :class="passErr ? 'msg-err' : 'msg-ok'">{{ passMsg }}</div>
          <div class="form-footer">
            <button type="submit" class="btn-primary" :disabled="passLoading || passwords.new !== passwords.confirm || passwords.new.length < 8">
              {{ passLoading ? 'Actualizando…' : 'Cambiar contraseña' }}
            </button>
          </div>
        </form>
      </div>

      <!-- Quick links -->
      <div class="quick-links">
        <router-link to="/orders" class="quick-link">📦 Mis pedidos</router-link>
        <router-link to="/catalog" class="quick-link">🛍 Ver catálogo</router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '../api/client'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const tab  = ref<'info' | 'address' | 'security'>('info')

const profile = ref({
  first_name: '',
  last_name:  '',
  phone:      '',
  default_address: { street: '', city: '', state: '', zip: '', country: '' },
})

const profileLoading = ref(false)
const profileMsg     = ref('')
const profileErr     = ref(false)

const passwords = ref({ current: '', new: '', confirm: '' })
const passLoading = ref(false)
const passMsg     = ref('')
const passErr     = ref(false)
const showCurrent = ref(false)
const showNew     = ref(false)

const initials = computed(() => {
  const u = auth.user
  if (!u) return '?'
  return ((u.first_name?.[0] ?? '') + (u.last_name?.[0] ?? '')).toUpperCase() || u.email?.[0]?.toUpperCase() || '?'
})

// Password strength
const strengthScore = computed(() => {
  const p = passwords.value.new
  if (p.length < 4) return 0
  let s = 0
  if (p.length >= 8) s++
  if (p.length >= 12) s++
  if (/[A-Z]/.test(p)) s++
  if (/[0-9]/.test(p)) s++
  if (/[^A-Za-z0-9]/.test(p)) s++
  return s
})
const strengthPct   = computed(() => Math.min(100, strengthScore.value * 20))
const strengthClass = computed(() => ['', 'weak', 'weak', 'fair', 'good', 'strong'][strengthScore.value] ?? 'strong')
const strengthLabel = computed(() => ['', 'Muy débil', 'Débil', 'Regular', 'Buena', 'Fuerte'][strengthScore.value] ?? 'Fuerte')

onMounted(() => {
  const u = auth.user
  if (u) {
    profile.value.first_name = u.first_name ?? ''
    profile.value.last_name  = u.last_name  ?? ''
    profile.value.phone      = u.phone      ?? ''
    if (u.default_address) {
      profile.value.default_address = { ...u.default_address }
    }
  }
})

async function saveProfile() {
  profileLoading.value = true; profileMsg.value = ''; profileErr.value = false
  try {
    const { data } = await api.put('/auth/profile', profile.value)
    await auth.me()
    profileMsg.value = 'Cambios guardados correctamente.'
  } catch (e: any) {
    profileErr.value = true
    profileMsg.value = e.response?.data?.message ?? 'Error al guardar'
  } finally { profileLoading.value = false }
}

async function savePassword() {
  if (passwords.value.new !== passwords.value.confirm) return
  passLoading.value = true; passMsg.value = ''; passErr.value = false
  try {
    await api.put('/auth/password', {
      current_password: passwords.value.current,
      new_password:     passwords.value.new,
    })
    passMsg.value = 'Contraseña actualizada correctamente.'
    passwords.value = { current: '', new: '', confirm: '' }
  } catch (e: any) {
    passErr.value = true
    passMsg.value = e.response?.data?.message ?? 'Error al cambiar contraseña'
  } finally { passLoading.value = false }
}
</script>

<style scoped>
.page      { min-height: 80vh; background: #f1f5f9; padding: 2rem 1rem; }
.container { max-width: 680px; margin: 0 auto; display: flex; flex-direction: column; gap: 1.5rem; }

/* Header */
.profile-header { display: flex; align-items: center; gap: 1.25rem; }
.avatar  { width: 64px; height: 64px; border-radius: 50%; background: linear-gradient(135deg, #3b82f6, #6366f1); color: #fff; font-size: 1.5rem; font-weight: 800; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.profile-name  { font-size: 1.5rem; font-weight: 800; color: #1e293b; margin: 0 0 .2rem; }
.profile-email { font-size: .875rem; color: #64748b; margin: 0; }

/* Tabs */
.tabs { display: flex; gap: 4px; background: #e2e8f0; border-radius: 10px; padding: 4px; }
.tab  { flex: 1; background: none; border: none; padding: .55rem .75rem; border-radius: 7px; font-size: .85rem; font-weight: 600; color: #64748b; cursor: pointer; transition: all .15s; }
.tab.active { background: #fff; color: #1d4ed8; box-shadow: 0 1px 4px rgba(0,0,0,.1); }

/* Card */
.card      { background: #fff; border-radius: 16px; padding: 2rem; box-shadow: 0 1px 4px rgba(0,0,0,.07); }
.card-title { font-size: 1.1rem; font-weight: 700; color: #1e293b; margin: 0 0 .4rem; }
.card-sub  { font-size: .85rem; color: #64748b; margin: 0 0 1.5rem; }

/* Form */
.form      { display: flex; flex-direction: column; gap: 1.1rem; margin-top: 1.5rem; }
.form-row  { display: flex; gap: 1rem; }
.field     { display: flex; flex-direction: column; gap: .35rem; flex: 1; }
.field label { font-size: .78rem; font-weight: 700; color: #475569; text-transform: uppercase; letter-spacing: .05em; }
.field input { border: 1.5px solid #e2e8f0; border-radius: 8px; padding: .65rem .9rem; font-size: .9rem; color: #1e293b; transition: border-color .15s; }
.field input:focus { outline: none; border-color: #3b82f6; }
.field--readonly .readonly-value { padding: .65rem .9rem; background: #f8fafc; border: 1.5px solid #e2e8f0; border-radius: 8px; font-size: .9rem; color: #64748b; }
.field-hint { font-size: .75rem; color: #94a3b8; margin: 0; }
.field-hint.err { color: #ef4444; }
.form-footer { display: flex; justify-content: flex-end; padding-top: .5rem; }

/* Messages */
.msg    { font-size: .875rem; padding: .75rem 1rem; border-radius: 8px; }
.msg-ok  { background: #dcfce7; color: #15803d; }
.msg-err { background: #fee2e2; color: #b91c1c; }

/* Password */
.pass-wrap { position: relative; display: flex; }
.pass-wrap input { flex: 1; padding-right: 2.5rem; }
.eye-btn { position: absolute; right: .6rem; top: 50%; transform: translateY(-50%); background: none; border: none; cursor: pointer; font-size: 1rem; line-height: 1; }
.strength-bar  { height: 4px; background: #e2e8f0; border-radius: 2px; overflow: hidden; margin-top: .4rem; }
.strength-fill { height: 100%; border-radius: 2px; transition: width .3s, background .3s; }
.strength-fill.weak   { background: #ef4444; }
.strength-fill.fair   { background: #f59e0b; }
.strength-fill.good   { background: #3b82f6; }
.strength-fill.strong { background: #10b981; }
.strength-label { font-size: .75rem; margin: .2rem 0 0; }
.strength-label.weak   { color: #ef4444; }
.strength-label.fair   { color: #f59e0b; }
.strength-label.good   { color: #3b82f6; }
.strength-label.strong { color: #10b981; }

/* Buttons */
.btn-primary { background: #1d4ed8; color: #fff; border: none; padding: .75rem 2rem; border-radius: 9px; font-size: .9rem; font-weight: 700; cursor: pointer; }
.btn-primary:hover { background: #1e40af; }
.btn-primary:disabled { opacity: .5; cursor: not-allowed; }

/* Quick links */
.quick-links { display: flex; gap: 1rem; flex-wrap: wrap; }
.quick-link  { display: flex; align-items: center; gap: .4rem; background: #fff; border: 1.5px solid #e2e8f0; color: #475569; padding: .6rem 1.2rem; border-radius: 9px; text-decoration: none; font-size: .875rem; font-weight: 600; }
.quick-link:hover { border-color: #3b82f6; color: #1d4ed8; }

@media (max-width: 520px) {
  .form-row { flex-direction: column; }
  .tabs .tab { font-size: .75rem; padding: .5rem .4rem; }
}
</style>
