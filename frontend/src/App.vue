<template>
  <div class="app-shell">
    <aside class="sidebar" v-if="auth.isLoggedIn.value">
      <div class="logo-wrap">
        <div class="logo-dot"></div>
        <div>
          <h1 class="logo-title">MoMail Console</h1>
          <p class="logo-sub">Temp Mail Platform</p>
        </div>
      </div>

      <nav class="side-nav">
        <RouterLink to="/">收件台</RouterLink>
        <RouterLink v-if="auth.isAdmin.value" to="/domains">域名管理</RouterLink>
        <RouterLink v-if="auth.can('user:manage')" to="/users">用户管理</RouterLink>
        <RouterLink v-if="auth.can('role:manage')" to="/roles">角色权限</RouterLink>
        <RouterLink v-if="auth.can('config:manage')" to="/config">系统配置</RouterLink>
      </nav>

      <div class="sidebar-footer">
        <div class="user-chip">{{ auth.state.user?.username }}</div>
        <button class="ghost" @click="handleLogout">退出登录</button>
      </div>
    </aside>

    <main class="main-view">
      <RouterView />
    </main>
  </div>
</template>

<script setup>
import { RouterLink, RouterView, useRouter } from 'vue-router'
import { useAuthStore } from './stores/auth'

const auth = useAuthStore()
const router = useRouter()

function handleLogout() {
  auth.logout()
  router.push('/login')
}
</script>