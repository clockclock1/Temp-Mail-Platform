import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import DashboardView from '../views/DashboardView.vue'
import DomainsView from '../views/DomainsView.vue'
import UsersView from '../views/UsersView.vue'
import RolesView from '../views/RolesView.vue'
import ConfigView from '../views/ConfigView.vue'
import { useAuthStore } from '../stores/auth'

const routes = [
  { path: '/login', name: 'login', component: LoginView, meta: { guestOnly: true, title: '欢迎回来', subtitle: '登录后进入临时邮箱指挥台。' } },
  { path: '/', name: 'dashboard', component: DashboardView, meta: { requiresAuth: true, title: '收件台', subtitle: '创建地址、浏览邮件、快速处理收件流。' } },
  { path: '/domains', name: 'domains', component: DomainsView, meta: { requiresAuth: true, adminOnly: true, title: '域名管理', subtitle: '维护域名池、多级泛解析和推送策略。' } },
  { path: '/users', name: 'users', component: UsersView, meta: { requiresAuth: true, adminOnly: true, title: '用户管理', subtitle: '控制成员、角色和账号状态。' } },
  { path: '/roles', name: 'roles', component: RolesView, meta: { requiresAuth: true, adminOnly: true, title: '角色权限', subtitle: '按职责组织权限矩阵与成员分工。' } },
  { path: '/config', name: 'config', component: ConfigView, meta: { requiresAuth: true, permission: 'config:manage', title: '系统配置', subtitle: '调整运行参数并实时应用配置。' } },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (auth.isLoggedIn.value && !auth.state.initialized) {
    try {
      await auth.refreshMe()
    } catch {
      // The response interceptor will clear the token and redirect if needed.
    }
  }
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
