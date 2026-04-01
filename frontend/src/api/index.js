import { api } from './client'

export const AuthAPI = {
  login(payload) {
    return api.post('/auth/login', payload)
  },
  me() {
    return api.get('/auth/me')
  },
}

export const DomainAPI = {
  available() {
    return api.get('/domains/available')
  },
  list() {
    return api.get('/domains')
  },
  create(payload) {
    return api.post('/domains', payload)
  },
  update(id, payload) {
    return api.put(`/domains/${id}`, payload)
  },
  remove(id) {
    return api.delete(`/domains/${id}`)
  },
}

export const MailboxAPI = {
  list() {
    return api.get('/mailboxes')
  },
  create(payload) {
    return api.post('/mailboxes', payload)
  },
  remove(id) {
    return api.delete(`/mailboxes/${id}`)
  },
  messages(mailboxId) {
    return api.get(`/mailboxes/${mailboxId}/messages`)
  },
}

export const MessageAPI = {
  get(id) {
    return api.get(`/messages/${id}`)
  },
  remove(id) {
    return api.delete(`/messages/${id}`)
  },
  raw(id) {
    return api.get(`/messages/${id}/raw`, { responseType: 'blob' })
  },
}

export const UserAPI = {
  list() {
    return api.get('/users')
  },
  create(payload) {
    return api.post('/users', payload)
  },
  update(id, payload) {
    return api.patch(`/users/${id}`, payload)
  },
  remove(id) {
    return api.delete(`/users/${id}`)
  },
}

export const RoleAPI = {
  list() {
    return api.get('/roles')
  },
  create(payload) {
    return api.post('/roles', payload)
  },
  update(id, payload) {
    return api.put(`/roles/${id}`, payload)
  },
  remove(id) {
    return api.delete(`/roles/${id}`)
  },
}

export const PermissionAPI = {
  list() {
    return api.get('/permissions')
  },
}

export const StatsAPI = {
  get() {
    return api.get('/stats')
  },
}
