<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">Configuración</h1>
        <p class="page-sub">Ajustes generales de la tienda</p>
      </div>
      <div class="header-actions">
        <span v-if="dirty && !saved" class="unsaved-badge">● Cambios sin guardar</span>
        <button class="btn-primary" :disabled="saving || !dirty" @click="save()">
          {{ saving ? 'Guardando…' : 'Guardar cambios' }}
        </button>
      </div>
    </div>

    <!-- Tabs -->
    <div class="tabs">
      <button v-for="t in tabs" :key="t.id" class="tab-btn"
        :class="{ active: activeTab === t.id }" @click="activeTab = t.id">
        {{ t.icon }} {{ t.label }}
      </button>
    </div>

    <div v-if="loading" class="loading">Cargando configuración…</div>

    <template v-else>
      <!-- Tab: Tienda -->
      <div v-show="activeTab === 'store'" class="section">
        <h3 class="section-title">Información de la tienda</h3>
        <div class="field-row">
          <div class="field">
            <label>Nombre *</label>
            <input v-model="form.store_name" required @input="dirty=true" />
          </div>
          <div class="field">
            <label>Moneda</label>
            <select v-model="form.currency" class="sel" @change="dirty=true">
              <option value="MXN">MXN — Peso mexicano</option>
              <option value="USD">USD — Dólar</option>
              <option value="EUR">EUR — Euro</option>
            </select>
          </div>
        </div>
        <div class="field">
          <label>URL del logo</label>
          <input v-model="form.logo_url" type="url" placeholder="https://…/logo.png" @input="dirty=true" />
        </div>
        <div v-if="form.logo_url" class="logo-preview">
          <img :src="form.logo_url" alt="Logo" />
        </div>

        <h3 class="section-title" style="margin-top:24px">Impuestos y precios</h3>
        <div class="field-row">
          <div class="field">
            <label>Tasa IVA (0–1)</label>
            <input v-model.number="form.tax_rate" type="number" step="0.01" min="0" max="1" @input="dirty=true" />
            <span class="field-hint">Ej: 0.16 = 16%</span>
          </div>
          <div class="field">
            <label>IVA incluido en precios</label>
            <button type="button" class="toggle-btn" :class="form.tax_inclusive ? 'on' : 'off'"
              @click="form.tax_inclusive = !form.tax_inclusive; dirty=true">
              {{ form.tax_inclusive ? 'Sí — precios con IVA' : 'No — IVA se suma al checkout' }}
            </button>
          </div>
        </div>
      </div>

      <!-- Tab: Contacto -->
      <div v-show="activeTab === 'contact'" class="section">
        <h3 class="section-title">Datos de contacto</h3>
        <div class="field-row">
          <div class="field">
            <label>Email de contacto *</label>
            <input v-model="form.contact_email" type="email" required @input="dirty=true" />
          </div>
          <div class="field">
            <label>Teléfono de soporte</label>
            <input v-model="form.support_phone" type="tel" placeholder="+52 55 1234 5678" @input="dirty=true" />
          </div>
        </div>

        <h3 class="section-title" style="margin-top:24px">Redes sociales</h3>
        <div class="field-row">
          <div class="field">
            <label>Instagram</label>
            <div class="input-prefix">
              <span class="prefix">instagram.com/</span>
              <input v-model="form.social_instagram" placeholder="mi_tienda" @input="dirty=true" />
            </div>
          </div>
          <div class="field">
            <label>Facebook</label>
            <div class="input-prefix">
              <span class="prefix">facebook.com/</span>
              <input v-model="form.social_facebook" placeholder="mi.tienda" @input="dirty=true" />
            </div>
          </div>
        </div>
        <div class="field" style="max-width:340px">
          <label>WhatsApp</label>
          <input v-model="form.social_whatsapp" type="tel" placeholder="+52 55 1234 5678" @input="dirty=true" />
          <span class="field-hint">Número con código de país para el botón de WhatsApp</span>
        </div>
      </div>

      <!-- Tab: Pagos -->
      <div v-show="activeTab === 'payments'" class="section">
        <h3 class="section-title">Stripe</h3>

        <div class="stripe-status" :class="stripeConfigured ? 'ok' : 'warn'">
          <span class="status-dot"></span>
          <span>{{ stripeConfigured ? 'Stripe configurado (clave secreta activa via env)' : 'Stripe no configurado — falta STRIPE_SECRET_KEY en el entorno' }}</span>
        </div>

        <div class="field" style="max-width:480px; margin-top:16px">
          <label>Clave pública (pk_…)</label>
          <input v-model="form.stripe_public_key" placeholder="pk_live_…  o  pk_test_…" @input="dirty=true" />
          <span class="field-hint">Se muestra en el frontend para inicializar Stripe.js. Es segura guardarla en BD.</span>
        </div>

        <div class="info-box" style="margin-top:20px">
          <strong>🔒 Clave secreta</strong><br />
          La clave secreta (<code>sk_…</code>) se configura únicamente como variable de entorno
          <code>STRIPE_SECRET_KEY</code> y nunca se almacena en la base de datos.
        </div>
      </div>

      <!-- Feedback global -->
      <div class="form-footer">
        <div v-if="saved" class="success-msg">✓ Configuración guardada</div>
        <div v-if="saveErr" class="err-msg">{{ saveErr }}</div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../../api/client'

const loading  = ref(true)
const saving   = ref(false)
const saved    = ref(false)
const saveErr  = ref('')
const dirty    = ref(false)
const activeTab = ref('store')
const stripeConfigured = ref(false)

const tabs = [
  { id: 'store',    icon: '🏪', label: 'Tienda'   },
  { id: 'contact',  icon: '📧', label: 'Contacto' },
  { id: 'payments', icon: '💳', label: 'Pagos'    },
]

const form = ref({
  store_name:      '',
  logo_url:        '',
  currency:        'MXN',
  tax_rate:        0.16,
  tax_inclusive:   false,
  contact_email:   '',
  support_phone:   '',
  stripe_public_key: '',
  social_instagram:  '',
  social_facebook:   '',
  social_whatsapp:   '',
})

async function load() {
  loading.value = true
  try {
    const [storeRes, payRes] = await Promise.all([
      api.get('/store'),
      api.get('/admin/store/payment'),
    ])
    Object.assign(form.value, storeRes.data)
    stripeConfigured.value = payRes.data.stripe_configured ?? false
  } catch(e) { console.error(e) }
  finally { loading.value = false }
}

async function save() {
  saving.value = true; saveErr.value = ''; saved.value = false
  try {
    await api.put('/admin/store', form.value)
    saved.value = true
    dirty.value = false
    setTimeout(() => { saved.value = false }, 3000)
  } catch(e: any) { saveErr.value = e.response?.data?.message ?? 'Error guardando' }
  finally { saving.value = false }
}

onMounted(load)
</script>

<style scoped>
.page        { display:flex; flex-direction:column; gap:20px; max-width:760px; }
.page-header { display:flex; align-items:flex-start; justify-content:space-between; gap:12px; flex-wrap:wrap; }
.page-title  { font-size:22px; font-weight:700; color:#eaf0f7; margin:0; }
.page-sub    { font-size:13px; color:#5a6a87; margin:4px 0 0; }
.header-actions { display:flex; align-items:center; gap:12px; }
.unsaved-badge { font-size:12px; color:#fb923c; }
.loading     { padding:40px; text-align:center; color:#5a6a87; }

.tabs { display:flex; gap:4px; border-bottom:1px solid #2d3a52; padding-bottom:0; }
.tab-btn { background:none; border:none; color:#5a6a87; font-size:13px; padding:10px 18px; cursor:pointer; border-bottom:2px solid transparent; transition:all .15s; margin-bottom:-1px; }
.tab-btn:hover  { color:#d6dfe8; }
.tab-btn.active { color:#38bdf8; border-bottom-color:#38bdf8; }

.section       { background:#1c2333; border:1px solid #2d3a52; border-radius:12px; padding:28px; display:flex; flex-direction:column; gap:16px; }
.section-title { font-size:13px; font-weight:700; color:#8494ac; margin:0; text-transform:uppercase; letter-spacing:.5px; }

.field         { display:flex; flex-direction:column; gap:6px; flex:1; }
.field label   { font-size:11px; color:#8494ac; text-transform:uppercase; letter-spacing:.5px; }
.field input, .field .sel {
  background:#0f1623; border:1px solid #2d3a52; color:#d6dfe8;
  padding:10px 12px; border-radius:8px; font-size:14px; width:100%; box-sizing:border-box;
}
.field input:focus { outline:none; border-color:#38bdf8; }
.field-hint    { font-size:11px; color:#5a6a87; }
.field-row     { display:flex; gap:16px; flex-wrap:wrap; }

.logo-preview img { max-height:60px; border-radius:6px; border:1px solid #2d3a52; }

.toggle-btn    { border:none; padding:10px 16px; border-radius:8px; font-size:13px; font-weight:600; cursor:pointer; text-align:left; }
.toggle-btn.on  { background:rgba(74,222,128,.1); color:#4ade80; border:1px solid rgba(74,222,128,.3); }
.toggle-btn.off { background:#0f1623; color:#5a6a87; border:1px solid #2d3a52; }

.input-prefix  { display:flex; align-items:center; background:#0f1623; border:1px solid #2d3a52; border-radius:8px; overflow:hidden; }
.prefix        { padding:10px 10px; font-size:12px; color:#5a6a87; background:#0a1019; border-right:1px solid #2d3a52; white-space:nowrap; }
.input-prefix input { border:none; border-radius:0; padding:10px 10px; background:transparent; }
.input-prefix input:focus { outline:none; }

.stripe-status { display:flex; align-items:center; gap:10px; padding:12px 16px; border-radius:8px; font-size:13px; }
.stripe-status.ok   { background:rgba(74,222,128,.08); border:1px solid rgba(74,222,128,.25); color:#4ade80; }
.stripe-status.warn { background:rgba(251,146,60,.08); border:1px solid rgba(251,146,60,.25); color:#fb923c; }
.status-dot         { width:8px; height:8px; border-radius:50%; background:currentColor; flex-shrink:0; }

.info-box { background:#0f1623; border:1px solid #2d3a52; border-radius:8px; padding:14px 16px; font-size:13px; color:#8494ac; line-height:1.6; }
.info-box code { background:#1c2333; padding:2px 6px; border-radius:4px; color:#38bdf8; font-size:12px; }

.form-footer   { display:flex; align-items:center; justify-content:flex-end; gap:16px; min-height:24px; }
.btn-primary   { background:#38bdf8; color:#080c14; border:none; padding:11px 24px; border-radius:8px; font-size:14px; font-weight:700; cursor:pointer; }
.btn-primary:disabled { opacity:.4; cursor:not-allowed; }
.success-msg   { color:#4ade80; font-size:13px; }
.err-msg       { color:#f87171; font-size:13px; }
</style>
