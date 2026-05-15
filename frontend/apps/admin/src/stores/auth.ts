import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '../api/client'

interface User {
  id: string; email: string; role: string
  first_name: string; last_name: string
  branch_id?: string; mfa_enabled: boolean; is_active: boolean
}

function loadUser(): User | null {
  try { return JSON.parse(localStorage.getItem('user') ?? 'null') } catch { return null }
}

export const useAuthStore = defineStore('auth', () => {
  const user  = ref<User | null>(loadUser())
  const token = ref(localStorage.getItem('token'))

  const isAuth    = computed(() => !!token.value)
  const isAdmin   = computed(() => user.value?.role === 'admin')
  const isManager = computed(() => ['admin','manager'].includes(user.value?.role ?? ''))
  const fullName  = computed(() => user.value ? `${user.value.first_name} ${user.value.last_name}` : '')

  if (token.value) api.defaults.headers.common.Authorization = `Bearer ${token.value}`

  async function login(email: string, password: string, totpCode?: string) {
    const { data } = await api.post('/auth/login', { email, password, totp_code: totpCode })
    token.value = data.access_token
    localStorage.setItem('token',         data.access_token)
    localStorage.setItem('refresh_token', data.refresh_token)
    api.defaults.headers.common.Authorization = `Bearer ${data.access_token}`
    if (data.user) {
      user.value = data.user
      localStorage.setItem('user', JSON.stringify(data.user))
    } else {
      await me()
    }
  }

  async function logout() {
    token.value = null; user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('refresh_token')
    localStorage.removeItem('user')
    delete api.defaults.headers.common.Authorization
  }

  async function me() {
    const { data } = await api.get('/auth/me')
    user.value = data
    localStorage.setItem('user', JSON.stringify(data))
  }

  function hasRole(min: string): boolean {
    const levels: Record<string,number> = { admin:100, manager:60, staff:40, customer:10 }
    return (levels[user.value?.role ?? ''] ?? 0) >= (levels[min] ?? 0)
  }

  return { user, token, isAuth, isAdmin, isManager, fullName, login, logout, me, hasRole }
})
