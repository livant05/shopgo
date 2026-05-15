import axios from 'axios'

let refreshing = false
let queue: Array<{ resolve: (t: string) => void; reject: (e: unknown) => void }> = []

function flush(err: unknown, token: string | null = null) {
  queue.forEach(p => err ? p.reject(err) : p.resolve(token!))
  queue = []
}

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL ?? '/api/v1',
  timeout: 20_000,
})

api.interceptors.request.use(cfg => {
  const t = localStorage.getItem('token')
  if (t) cfg.headers.Authorization = `Bearer ${t}`
  return cfg
})

api.interceptors.response.use(
  r => r,
  async err => {
    const orig = err.config
    if (err.response?.status !== 401 || orig._retry) return Promise.reject(err)

    if (refreshing) {
      return new Promise((resolve, reject) => queue.push({ resolve, reject }))
        .then(t => { orig.headers.Authorization = `Bearer ${t}`; return api(orig) })
    }

    orig._retry = true
    refreshing  = true

    try {
      const { data } = await axios.post('/api/v1/auth/refresh', {
        refresh_token: localStorage.getItem('refresh_token'),
      })
      localStorage.setItem('token',         data.access_token)
      localStorage.setItem('refresh_token', data.refresh_token)
      api.defaults.headers.common.Authorization = `Bearer ${data.access_token}`
      flush(null, data.access_token)
      orig.headers.Authorization = `Bearer ${data.access_token}`
      return api(orig)
    } catch (e) {
      flush(e)
      localStorage.removeItem('token')
      localStorage.removeItem('refresh_token')
      window.location.href = '/login'
    } finally {
      refreshing = false
    }
  }
)
