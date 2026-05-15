import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '../api/client'

function loadUser() {
  try { return JSON.parse(localStorage.getItem('sf_user') ?? 'null') } catch { return null }
}

export const useAuthStore = defineStore('auth', () => {
  const user  = ref<any>(loadUser())
  const token = ref(localStorage.getItem('token'))

  const isAuth   = computed(() => !!token.value)
  const fullName = computed(() => user.value ? `${user.value.first_name} ${user.value.last_name}` : '')

  async function login(email: string, password: string) {
    const { data } = await api.post('/auth/login', { email, password })
    token.value = data.access_token
    localStorage.setItem('token',         data.access_token)
    localStorage.setItem('refresh_token', data.refresh_token ?? '')
    api.defaults.headers.common.Authorization = `Bearer ${data.access_token}`
    if (data.user) {
      user.value = data.user
      localStorage.setItem('sf_user', JSON.stringify(data.user))
    } else {
      await me()
    }
  }

  async function register(email: string, password: string, firstName: string, lastName: string) {
    await api.post('/auth/register', { email, password, first_name: firstName, last_name: lastName })
    await login(email, password)
  }

  async function logout() {
    token.value = null; user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('refresh_token')
    localStorage.removeItem('sf_user')
    delete api.defaults.headers.common.Authorization
  }

  async function me() {
    try {
      const { data } = await api.get('/auth/me')
      user.value = data
      localStorage.setItem('sf_user', JSON.stringify(data))
    } catch {}
  }

  if (token.value) {
    api.defaults.headers.common.Authorization = `Bearer ${token.value}`
    if (!user.value) me()
  }

  return { user, token, isAuth, fullName, login, register, logout, me }
})
