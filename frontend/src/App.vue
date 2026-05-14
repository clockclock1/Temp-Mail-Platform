<template>
  <div class="app-shell" :class="{ 'app-shell--guest': !auth.isLoggedIn.value }">
    <div class="ambient ambient-a"></div>
    <div class="ambient ambient-b"></div>
    <div class="ambient ambient-c"></div>

    <aside v-if="auth.isLoggedIn.value" class="sidebar">
      <div class="logo-wrap">
        <div class="logo-mark">
          <span></span>
        </div>
        <div>
          <h1 class="logo-title">MoMail Console</h1>
          <p class="logo-sub">Temp Mail Platform</p>
        </div>
      </div>

      <div class="side-kicker">
        <span class="kicker-dot"></span>
        <span>Live inbox orbit</span>
      </div>

      <nav class="side-nav">
        <RouterLink v-for="item in navItems" :key="item.to" :to="item.to" class="nav-item">
          <span class="nav-icon">{{ item.icon }}</span>
          <span class="nav-body">
            <strong>{{ item.label }}</strong>
            <small>{{ item.desc }}</small>
          </span>
        </RouterLink>
      </nav>

      <div class="sidebar-footer">
        <div class="user-card">
          <div class="user-avatar">{{ userInitial }}</div>
          <div>
            <div class="user-name">{{ auth.state.user?.username || 'guest' }}</div>
            <div class="user-meta">{{ auth.state.user?.role?.name || 'member' }}</div>
          </div>
        </div>
        <button class="ghost" @click="handleLogout">退出登录</button>
      </div>
    </aside>

    <main class="main-view">
      <header v-if="auth.isLoggedIn.value" class="view-topbar">
        <div>
          <div class="page-badge">{{ currentMeta.title || '控制台' }}</div>
          <h2 class="topbar-title">{{ currentMeta.title || '控制台' }}</h2>
          <p class="topbar-sub">{{ currentMeta.subtitle || '在这里管理你的临时邮箱系统。' }}</p>
        </div>
        <div class="topbar-chip">
          <span class="chip-glow"></span>
          <span>{{ currentTime }}</span>
        </div>
      </header>

      <section class="view-stage">
        <RouterView v-slot="{ Component }">
          <Transition name="view" mode="out-in">
            <component :is="Component" />
          </Transition>
        </RouterView>
      </section>
    </main>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { useAuthStore } from './stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const currentTime = ref('')
let timer = null

const navItems = computed(() => [
  { to: '/', label: '收件台', desc: '创建地址与查看邮件', icon: '收' },
  ...(auth.isAdmin.value ? [{ to: '/domains', label: '域名管理', desc: '域名池与层级推送', icon: '域' }] : []),
  ...(auth.can('user:manage') ? [{ to: '/users', label: '用户管理', desc: '账号与状态控制', icon: '用' }] : []),
  ...(auth.can('role:manage') ? [{ to: '/roles', label: '角色权限', desc: '权限矩阵编排', icon: '权' }] : []),
  ...(auth.can('config:manage') ? [{ to: '/config', label: '系统配置', desc: '运行参数与热更新', icon: '配' }] : []),
])

const currentMeta = computed(() => route.meta || {})
const userInitial = computed(() => String(auth.state.user?.username || 'M').slice(0, 1).toUpperCase())

onMounted(() => {
  const tick = () => {
    currentTime.value = new Intl.DateTimeFormat('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      weekday: 'short',
      hour: '2-digit',
      minute: '2-digit',
    }).format(new Date())
  }
  tick()
  timer = setInterval(tick, 30000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

function handleLogout() {
  auth.logout()
  router.push('/login')
}
</script>
