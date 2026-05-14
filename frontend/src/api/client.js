import axios from 'axios'

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE || '/api/v1',
  timeout: 15000,
})

const LOCAL_STORAGE_KEY = 'tempmail.console.auth'
const SESSION_STORAGE_KEY = 'tempmail.console.auth.session'

function readToken() {
  const keys = [SESSION_STORAGE_KEY, LOCAL_STORAGE_KEY]
  for (const key of keys) {
    try {
      const payload = JSON.parse(window.sessionStorage.getItem(key) || window.localStorage.getItem(key) || 'null')
      if (payload?.token) return payload.token
    } catch {
      // Ignore malformed client-side cache and continue.
    }
  }
  return ''
}

function clearToken() {
  window.localStorage.removeItem(LOCAL_STORAGE_KEY)
  window.sessionStorage.removeItem(SESSION_STORAGE_KEY)
}

api.interceptors.request.use((config) => {
  const token = readToken()
  if (token) {
    config.headers = config.headers || {}
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error?.response?.status
    const requestURL = String(error?.config?.url || '')
    const isLoginRequest = requestURL.includes('/auth/login')

    if (status === 401 && !isLoginRequest) {
      clearToken()
      if (window.location.pathname !== '/login') {
        const next = `${window.location.pathname}${window.location.search}`
        window.location.replace(`/login?next=${encodeURIComponent(next)}`)
      }
    }

    return Promise.reject(error)
  },
)
