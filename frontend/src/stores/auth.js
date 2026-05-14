import { computed, reactive } from 'vue'
import { AuthAPI } from '../api'

const LOCAL_STORAGE_KEY = 'tempmail.console.auth'
const SESSION_STORAGE_KEY = 'tempmail.console.auth.session'

function safeParse(raw) {
  if (!raw) return null
  try {
    return JSON.parse(raw)
  } catch {
    return null
  }
}

function readSnapshot() {
  const sessionPayload = safeParse(window.sessionStorage.getItem(SESSION_STORAGE_KEY))
  if (sessionPayload?.token) {
    return { payload: sessionPayload, remember: false }
  }

  const localPayload = safeParse(window.localStorage.getItem(LOCAL_STORAGE_KEY))
  if (localPayload?.token) {
    return { payload: localPayload, remember: true }
  }

  return { payload: null, remember: true }
}

function clearPersisted() {
  window.localStorage.removeItem(LOCAL_STORAGE_KEY)
  window.sessionStorage.removeItem(SESSION_STORAGE_KEY)
}

function persistSnapshot(snapshot, remember) {
  clearPersisted()
  const storage = remember ? window.localStorage : window.sessionStorage
  const key = remember ? LOCAL_STORAGE_KEY : SESSION_STORAGE_KEY
  storage.setItem(key, JSON.stringify(snapshot))
}

const restored = readSnapshot()

const state = reactive({
  token: restored.payload?.token || '',
  user: restored.payload?.user || null,
  perms: Array.isArray(restored.payload?.perms) ? restored.payload.perms : [],
  remember: restored.remember,
  loading: false,
  initialized: !restored.payload?.token,
})

const isLoggedIn = computed(() => Boolean(state.token))
const isAdmin = computed(() => String(state.user?.role?.name || '').toLowerCase() === 'admin')

let refreshPromise = null

function setAuth(payload, remember = true) {
  state.token = payload.token || state.token || ''
  state.user = payload.user || null
  state.perms = Array.isArray(payload.perms) ? payload.perms : []
  state.remember = remember
  state.initialized = true

  if (state.token) {
    persistSnapshot(
      {
        token: state.token,
        user: state.user,
        perms: state.perms,
      },
      remember,
    )
  } else {
    clearPersisted()
  }
}

function can(permission) {
  if (!permission) return true
  return isAdmin.value || state.perms.includes(permission)
}

function logout() {
  state.token = ''
  state.user = null
  state.perms = []
  state.remember = true
  state.loading = false
  state.initialized = true
  clearPersisted()
}

async function refreshMe(force = false) {
  if (!state.token) {
    state.initialized = true
    return null
  }

  if (refreshPromise && !force) {
    return refreshPromise
  }

  refreshPromise = (async () => {
    state.loading = true
    try {
      const { data } = await AuthAPI.me()
      setAuth(
        {
          token: state.token,
          user: data.user,
          perms: data.perms,
        },
        state.remember,
      )
      return data.user
    } catch (error) {
      logout()
      throw error
    } finally {
      state.loading = false
      state.initialized = true
      refreshPromise = null
    }
  })()

  return refreshPromise
}

export function useAuthStore() {
  return {
    state,
    isLoggedIn,
    isAdmin,
    setAuth,
    refreshMe,
    can,
    logout,
  }
}
