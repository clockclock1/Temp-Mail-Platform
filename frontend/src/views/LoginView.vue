<template>
  <div class="login-wrap">
    <div class="card login-card">
      <h1>登录控制台</h1>
      <p class="meta">登录后才能创建和管理临时邮箱。</p>

      <div class="grid">
        <label>
          用户名
          <input v-model="form.username" placeholder="admin" @keyup.enter="submit" />
        </label>
        <label>
          密码
          <input v-model="form.password" placeholder="请输入密码" type="password" @keyup.enter="submit" />
        </label>
      </div>

      <button class="primary" style="margin-top: 12px; width: 100%" @click="submit" :disabled="loading">
        {{ loading ? '登录中...' : '登录' }}
      </button>

      <p v-if="error" class="error">{{ error }}</p>
      <p class="meta" style="margin-top: 12px">默认管理员账号来自后端环境变量 `DEFAULT_ADMIN_USER/DEFAULT_ADMIN_PASS`。</p>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { AuthAPI } from '../api'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const form = reactive({ username: '', password: '' })
const loading = ref(false)
const error = ref('')

async function submit() {
  error.value = ''
  loading.value = true
  try {
    const { data } = await AuthAPI.login(form)
    auth.setAuth(data)
    const next = typeof route.query.next === 'string' ? route.query.next : '/'
    router.push(next)
  } catch (e) {
    error.value = e?.response?.data?.error || '登录失败'
  } finally {
    loading.value = false
  }
}
</script>