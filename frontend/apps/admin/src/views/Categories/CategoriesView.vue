<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">Categorías</h1>
        <p class="page-sub">{{ categories.length }} categorías en total</p>
      </div>
      <button class="btn-primary" @click="openCreate()">+ Nueva categoría</button>
    </div>

    <div v-if="loading" class="loading">Cargando categorías…</div>

    <div v-else class="tree-wrap">
      <div v-if="roots.length === 0" class="empty-state">No hay categorías aún. ¡Crea la primera!</div>

      <div v-for="root in roots" :key="root.id" class="tree-root">
        <div class="cat-row" :class="{ inactive: !root.is_active }">
          <div class="cat-info">
            <span class="cat-name">{{ root.name }}</span>
            <span class="cat-slug">{{ root.slug }}</span>
            <span class="badge-count">{{ root.product_count }} productos</span>
            <span v-if="!root.is_active" class="badge-inactive">Archivada</span>
          </div>
          <div class="cat-actions">
            <button class="btn-icon" title="Editar" @click="openEdit(root)">✏️</button>
            <button
              class="btn-icon"
              :title="root.is_active ? 'Archivar' : 'Activar'"
              @click="toggleActive(root)"
            >{{ root.is_active ? '📦' : '✅' }}</button>
          </div>
        </div>

        <div v-if="childrenOf(root.id).length" class="tree-children">
          <div v-for="child in childrenOf(root.id)" :key="child.id" class="cat-row child-row" :class="{ inactive: !child.is_active }">
            <div class="cat-info">
              <span class="tree-indent">└─</span>
              <span class="cat-name">{{ child.name }}</span>
              <span class="cat-slug">{{ child.slug }}</span>
              <span class="badge-count">{{ child.product_count }} productos</span>
              <span v-if="!child.is_active" class="badge-inactive">Archivada</span>
            </div>
            <div class="cat-actions">
              <button class="btn-icon" title="Editar" @click="openEdit(child)">✏️</button>
              <button
                class="btn-icon"
                :title="child.is_active ? 'Archivar' : 'Activar'"
                @click="toggleActive(child)"
              >{{ child.is_active ? '📦' : '✅' }}</button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal crear/editar -->
    <Teleport to="body">
      <div v-if="modal" class="overlay" @click.self="closeModal()">
        <div class="modal">
          <div class="modal-header">
            <h2>{{ editing ? 'Editar categoría' : 'Nueva categoría' }}</h2>
            <button class="btn-close" @click="closeModal()">✕</button>
          </div>

          <div class="form-grid">
            <label class="form-label">
              Nombre *
              <input v-model="form.name" class="form-input" placeholder="Ej. Electrónica" @input="autoSlug()" />
            </label>

            <label class="form-label">
              Slug (URL)
              <input v-model="form.slug" class="form-input" placeholder="electronica" />
            </label>

            <label class="form-label" style="grid-column: 1/-1">
              Descripción
              <textarea v-model="form.description" class="form-input" rows="3" placeholder="Descripción breve de la categoría…" />
            </label>

            <label class="form-label">
              Categoría padre
              <select v-model="form.parent_id" class="form-input">
                <option value="">— Sin padre (raíz) —</option>
                <option
                  v-for="c in rootOptions"
                  :key="c.id"
                  :value="c.id"
                  :disabled="editing && c.id === form.id"
                >{{ c.name }}</option>
              </select>
            </label>

            <label class="form-label">
              Orden de visualización
              <input v-model.number="form.sort_order" type="number" min="0" class="form-input" />
            </label>
          </div>

          <p v-if="modalErr" class="err-msg">{{ modalErr }}</p>

          <div class="modal-footer">
            <button class="btn-ghost" @click="closeModal()">Cancelar</button>
            <button class="btn-primary" :disabled="saving" @click="save()">
              {{ saving ? 'Guardando…' : (editing ? 'Actualizar' : 'Crear') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

interface Category {
  id: string
  name: string
  slug: string
  description: string
  parent_id: string | null
  sort_order: number
  is_active: boolean
  product_count: number
}

const API = import.meta.env.VITE_API_URL ?? '/api/v1'
const token = () => localStorage.getItem('token') ?? ''

const categories = ref<Category[]>([])
const loading = ref(false)
const modal = ref(false)
const editing = ref(false)
const saving = ref(false)
const modalErr = ref('')

const emptyForm = (): any => ({
  id: '',
  name: '',
  slug: '',
  description: '',
  parent_id: '',
  sort_order: 0,
})
const form = ref(emptyForm())

const roots = computed(() => categories.value.filter(c => !c.parent_id).sort((a, b) => a.sort_order - b.sort_order || a.name.localeCompare(b.name)))
const rootOptions = computed(() => categories.value.filter(c => !c.parent_id))

function childrenOf(parentId: string): Category[] {
  return categories.value.filter(c => c.parent_id === parentId).sort((a, b) => a.sort_order - b.sort_order || a.name.localeCompare(b.name))
}

async function load() {
  loading.value = true
  try {
    const r = await fetch(`${API}/categories`)
    const j = await r.json()
    categories.value = j.data ?? []
  } finally {
    loading.value = false
  }
}

function autoSlug() {
  if (!editing.value) {
    form.value.slug = form.value.name.toLowerCase().normalize('NFD').replace(/[̀-ͯ]/g, '').replace(/\s+/g, '-').replace(/[^a-z0-9-]/g, '')
  }
}

function openCreate() {
  editing.value = false
  form.value = emptyForm()
  modalErr.value = ''
  modal.value = true
}

function openEdit(cat: Category) {
  editing.value = true
  form.value = {
    id: cat.id,
    name: cat.name,
    slug: cat.slug,
    description: cat.description ?? '',
    parent_id: cat.parent_id ?? '',
    sort_order: cat.sort_order,
  }
  modalErr.value = ''
  modal.value = true
}

function closeModal() {
  modal.value = false
}

async function save() {
  if (!form.value.name.trim()) {
    modalErr.value = 'El nombre es obligatorio.'
    return
  }
  saving.value = true
  modalErr.value = ''
  try {
    const payload = {
      name: form.value.name,
      slug: form.value.slug || undefined,
      description: form.value.description,
      parent_id: form.value.parent_id || null,
      sort_order: form.value.sort_order,
    }
    const url = editing.value ? `${API}/admin/categories/${form.value.id}` : `${API}/admin/categories`
    const method = editing.value ? 'PUT' : 'POST'
    const r = await fetch(url, {
      method,
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token()}` },
      body: JSON.stringify(payload),
    })
    if (!r.ok) {
      const e = await r.json()
      modalErr.value = e.message ?? 'Error al guardar'
      return
    }
    closeModal()
    await load()
  } finally {
    saving.value = false
  }
}

async function toggleActive(cat: Category) {
  const r = await fetch(`${API}/admin/categories/${cat.id}/active`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token()}` },
    body: JSON.stringify({ active: !cat.is_active }),
  })
  if (r.ok) await load()
}

onMounted(load)
</script>

<style scoped>
.page { padding: 1.5rem 2rem; max-width: 860px; }
.page-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 1.5rem; }
.page-title { font-size: 1.5rem; font-weight: 700; color: #0f172a; margin: 0; }
.page-sub { color: #64748b; font-size: .875rem; margin: .25rem 0 0; }
.loading { color: #64748b; padding: 2rem 0; }
.empty-state { text-align: center; color: #94a3b8; padding: 3rem; border: 2px dashed #e2e8f0; border-radius: .75rem; }

.tree-wrap { display: flex; flex-direction: column; gap: .5rem; }
.tree-root { border: 1px solid #e2e8f0; border-radius: .75rem; overflow: hidden; background: #fff; }

.cat-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: .875rem 1rem;
  transition: background .15s;
}
.cat-row:hover { background: #f8fafc; }
.cat-row.inactive { opacity: .55; }
.child-row { padding-left: 1.5rem; border-top: 1px solid #f1f5f9; background: #fafafa; }

.cat-info { display: flex; align-items: center; gap: .75rem; flex-wrap: wrap; }
.cat-name { font-weight: 600; color: #1e293b; }
.cat-slug { font-size: .8rem; color: #94a3b8; font-family: monospace; }
.tree-indent { color: #94a3b8; font-family: monospace; }
.badge-count { font-size: .75rem; background: #e0f2fe; color: #0369a1; border-radius: 999px; padding: .1rem .55rem; }
.badge-inactive { font-size: .75rem; background: #fef3c7; color: #92400e; border-radius: 999px; padding: .1rem .55rem; }

.tree-children { }

.cat-actions { display: flex; gap: .25rem; }
.btn-icon { background: none; border: none; cursor: pointer; font-size: 1rem; padding: .25rem .4rem; border-radius: .35rem; }
.btn-icon:hover { background: #f1f5f9; }

/* modal */
.overlay { position: fixed; inset: 0; background: rgba(0,0,0,.45); display: flex; align-items: center; justify-content: center; z-index: 9999; padding: 1rem; }
.modal { background: #fff; border-radius: 1rem; width: 100%; max-width: 540px; box-shadow: 0 20px 60px rgba(0,0,0,.25); display: flex; flex-direction: column; gap: 1.25rem; padding: 1.5rem; }
.modal-header { display: flex; justify-content: space-between; align-items: center; }
.modal-header h2 { margin: 0; font-size: 1.15rem; font-weight: 700; color: #0f172a; }
.btn-close { background: none; border: none; font-size: 1.2rem; cursor: pointer; color: #64748b; }
.modal-footer { display: flex; justify-content: flex-end; gap: .75rem; }

.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
.form-label { display: flex; flex-direction: column; gap: .35rem; font-size: .875rem; font-weight: 500; color: #374151; }
.form-input { padding: .55rem .75rem; border: 1px solid #d1d5db; border-radius: .5rem; font-size: .9rem; outline: none; transition: border-color .15s; }
.form-input:focus { border-color: #6366f1; }

.err-msg { color: #dc2626; font-size: .85rem; background: #fef2f2; padding: .5rem .75rem; border-radius: .5rem; margin: 0; }

.btn-primary { background: #6366f1; color: #fff; border: none; border-radius: .5rem; padding: .6rem 1.25rem; font-size: .9rem; font-weight: 600; cursor: pointer; }
.btn-primary:hover:not(:disabled) { background: #4f46e5; }
.btn-primary:disabled { opacity: .55; cursor: not-allowed; }
.btn-ghost { background: none; border: 1px solid #d1d5db; border-radius: .5rem; padding: .6rem 1.25rem; font-size: .9rem; cursor: pointer; }
.btn-ghost:hover { background: #f1f5f9; }
</style>
