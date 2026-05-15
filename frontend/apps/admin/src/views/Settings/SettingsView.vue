<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">Configuración</h1>
        <p class="page-sub">Ajustes generales de la tienda</p>
      </div>
    </div>

    <div v-if="loading" class="loading">Cargando configuración…</div>

    <template v-else>
      <form @submit.prevent="save()" class="settings-form">
        <!-- Información de la tienda -->
        <div class="section">
          <h3 class="section-title">🏪 Información de la tienda</h3>
          <div class="field-row">
            <div class="field">
              <label>Nombre de la tienda *</label>
              <input v-model="form.store_name" required />
            </div>
            <div class="field">
              <label>Moneda</label>
              <select v-model="form.currency" class="sel">
                <option value="MXN">MXN — Peso mexicano</option>
                <option value="USD">USD — Dólar</option>
                <option value="EUR">EUR — Euro</option>
              </select>
            </div>
          </div>
          <div class="field">
            <label>URL del logo</label>
            <input v-model="form.logo_url" type="url" placeholder="https://…/logo.png" />
          </div>
          <div v-if="form.logo_url" class="logo-preview">
            <img :src="form.logo_url" alt="Logo" />
          </div>
        </div>

        <!-- Impuestos -->
        <div class="section">
          <h3 class="section-title">📊 Impuestos y precios</h3>
          <div class="field-row">
            <div class="field">
              <label>Tasa IVA (0 a 1)</label>
              <input v-model.number="form.tax_rate" type="number" step="0.01" min="0" max="1" />
              <span class="field-hint">Ejemplo: 0.16 = 16%</span>
            </div>
            <div class="field">
              <label>IVA incluido en precios</label>
              <div class="toggle-row">
                <button type="button" class="toggle-btn" :class="form.tax_inclusive ? 'on' : 'off'"
                  @click="form.tax_inclusive = !form.tax_inclusive">
                  {{ form.tax_inclusive ? 'Sí — precios con IVA' : 'No — IVA se suma al checkout' }}
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Contacto -->
        <div class="section">
          <h3 class="section-title">📧 Contacto</h3>
          <div class="field-row">
            <div class="field">
              <label>Email de contacto *</label>
              <input v-model="form.contact_email" type="email" required />
            </div>
            <div class="field">
              <label>Teléfono de soporte</label>
              <input v-model="form.support_phone" type="tel" placeholder="+52 55 1234 5678" />
            </div>
          </div>
        </div>

        <div class="form-footer">
          <div v-if="saved" class="success-msg">✓ Configuración guardada correctamente</div>
          <div v-if="saveErr" class="err-msg">{{ saveErr }}</div>
          <button type="submit" class="btn-primary" :disabled="saving">
            {{ saving ? 'Guardando…' : 'Guardar configuración' }}
          </button>
        </div>
      </form>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../../api/client'

const loading = ref(true)
const saving  = ref(false)
const saved   = ref(false)
const saveErr = ref('')

const form = ref({
  store_name: '',
  logo_url: '',
  currency: 'MXN',
  tax_rate: 0.16,
  tax_inclusive: false,
  contact_email: '',
  support_phone: '',
})

async function load() {
  loading.value = true
  try {
    const { data } = await api.get('/store')
    form.value = { ...form.value, ...data }
  } catch(e) { console.error(e) }
  finally { loading.value = false }
}

async function save() {
  saving.value = true; saveErr.value = ''; saved.value = false
  try {
    await api.put('/admin/store', form.value)
    saved.value = true
    setTimeout(() => { saved.value = false }, 3000)
  } catch(e: any) { saveErr.value = e.response?.data?.message ?? 'Error guardando' }
  finally { saving.value = false }
}

onMounted(load)
</script>

<style scoped>
.page        { display:flex; flex-direction:column; gap:24px; max-width:720px; }
.page-header { display:flex; align-items:flex-start; justify-content:space-between; }
.page-title  { font-size:22px; font-weight:700; color:#eaf0f7; margin:0; }
.page-sub    { font-size:13px; color:#5a6a87; margin:4px 0 0; }
.loading     { padding:40px; text-align:center; color:#5a6a87; }

.settings-form { display:flex; flex-direction:column; gap:20px; }
.section       { background:#1c2333; border:1px solid #2d3a52; border-radius:12px; padding:24px; display:flex; flex-direction:column; gap:16px; }
.section-title { font-size:14px; font-weight:700; color:#d6dfe8; margin:0; }

.field         { display:flex; flex-direction:column; gap:6px; flex:1; }
.field label   { font-size:11px; color:#8494ac; text-transform:uppercase; letter-spacing:.5px; }
.field input, .field .sel { background:#0f1623; border:1px solid #2d3a52; color:#d6dfe8; padding:10px 12px; border-radius:8px; font-size:14px; width:100%; }
.field input:focus { outline:none; border-color:#38bdf8; }
.field-hint    { font-size:11px; color:#5a6a87; }
.field-row     { display:flex; gap:16px; }

.logo-preview img { max-height:60px; border-radius:6px; margin-top:4px; border:1px solid #2d3a52; }

.toggle-row    { margin-top:4px; }
.toggle-btn    { border:none; padding:10px 16px; border-radius:8px; font-size:13px; font-weight:600; cursor:pointer; }
.toggle-btn.on  { background:rgba(74,222,128,.1); color:#4ade80; border:1px solid rgba(74,222,128,.3); }
.toggle-btn.off { background:#0f1623; color:#5a6a87; border:1px solid #2d3a52; }

.form-footer   { display:flex; align-items:center; justify-content:flex-end; gap:16px; }
.btn-primary   { background:#38bdf8; color:#080c14; border:none; padding:12px 28px; border-radius:8px; font-size:14px; font-weight:700; cursor:pointer; }
.btn-primary:disabled { opacity:.5; cursor:not-allowed; }
.success-msg   { color:#4ade80; font-size:13px; }
.err-msg       { color:#f87171; font-size:13px; }
</style>
