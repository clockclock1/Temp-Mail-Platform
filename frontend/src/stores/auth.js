import { reactive, computed } from 'vue'
import { api } from '../api/client'

const TOKEN_KEY = 'tm_token'
const USER_KEY = 'tm_user'
const PERMS_KEY = 'tm_perms'

const state = reactive({
  token: localStorage.getItem(TOKEN_KEY) || '',
  user: JSON.parse(localStorage.getItem(USER_KEY) || 'null'),
  perms: JSON.parse(localStorage.getItem(PERMS_KEY) || '[]'),
})

if (state.token) {
  api.defaults.headers.common.Authorization = `Bearer ${state.token}`
}

const isLoggedIn = computed(() => !!state.token)
const isAdmin = computed(() => state.user?.role?.name === 'admin')

function persist() {
  if (state.token) {
    localStorage.setItem(TOKEN_KEY, state.token)
    localStorage.setItem(USER_KEY, JSON.stringify(state.user))
    localStorage.setItem(PERMS_KEY, JSON.stringify(state.perms))
    api.defaults.headers.common.Authorization = `Bearer ${state.token}`
  } else {
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(USER_KEY)
    localStorage.removeItem(PERMS_KEY)
    delete api.defaults.headers.common.Authorization
  }
}

function setAuth(payload) {
  state.token = payload.token
  state.user = payload.user
  state.perms = payload.perms || []
  persist()
}

function logout() {
  state.token = ''
  state.user = null
  state.perms = []
  persist()
}

async function refreshMe() {
  if (!state.token) return
  const { data } = await api.get('/auth/me')
  state.user = data.user
  state.perms = data.perms || []
  persist()
}

function can(permission) {
  if (isAdmin.value) return true
  return state.perms.includes(permission)
}

export function useAuthStore() {
  return {
    state,
    isLoggedIn,
    isAdmin,
    setAuth,
    logout,
    refreshMe,
    can,
  }
}