import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import DashboardView from '../views/DashboardView.vue'
import DomainsView from '../views/DomainsView.vue'
import UsersView from '../views/UsersView.vue'
import RolesView from '../views/RolesView.vue'
import ConfigView from '../views/ConfigView.vue'
import { useAuthStore } from '../stores/auth'

const routes = [
  { path: '/login', name: 'login', component: LoginView, meta: { guestOnly: true } },
  { path: '/', name: 'dashboard', component: DashboardView, meta: { requiresAuth: true } },
  { path: '/domains', name: 'domains', component: DomainsView, meta: { requiresAuth: true, adminOnly: true } },
  { path: '/users', name: 'users', component: UsersView, meta: { requiresAuth: true, adminOnly: true } },
  { path: '/roles', name: 'roles', component: RolesView, meta: { requiresAuth: true, adminOnly: true } },
  { path: '/config', name: 'config', component: ConfigView, meta: { requiresAuth: true, permission: 'config:manage' } },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isLoggedIn.value) {
    return { name: 'login', query: { next: to.fullPath } }
  }
  if (to.meta.guestOnly && auth.isLoggedIn.value) {
    return { name: 'dashboard' }
  }
  if (to.meta.adminOnly && !auth.isAdmin.value) {
    return { name: 'dashboard' }
  }
  if (to.meta.permission && !auth.can(to.meta.permission)) {
    return { name: 'dashboard' }
  }
  return true
})

export default router
