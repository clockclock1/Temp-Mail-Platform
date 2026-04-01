<template>
  <div class="app-shell">
    <header class="topbar" v-if="auth.isLoggedIn.value">
      <div class="brand">
        <span class="dot"></span>
        <strong>TempMail Console</strong>
      </div>
      <nav class="menu">
        <RouterLink to="/">邮箱</RouterLink>
        <RouterLink v-if="auth.isAdmin.value" to="/domains">域名</RouterLink>
        <RouterLink v-if="auth.isAdmin.value" to="/users">用户</RouterLink>
        <RouterLink v-if="auth.isAdmin.value" to="/roles">角色</RouterLink>
        <RouterLink v-if="auth.can('config:manage')" to="/config">配置</RouterLink>
      </nav>
      <div class="actions">
        <span class="user">{{ auth.state.user?.username }}</span>
        <button class="ghost" @click="handleLogout">退出</button>
      </div>
    </header>

    <main class="content">
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
