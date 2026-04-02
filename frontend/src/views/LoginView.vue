<template>
  <div class="login-wrap">
    <div class="card login-card">
      <h1 class="section-title">登录控制台</h1>
      <p class="section-sub">管理员与授权用户可创建临时邮箱、查看邮件、管理系统。</p>

      <div class="grid" style="margin-top: 12px">
        <label>
          用户名
          <input v-model="form.username" placeholder="admin" @keyup.enter="submit" />
        </label>
        <label>
          密码
          <input v-model="form.password" type="password" placeholder="请输入密码" @keyup.enter="submit" />
        </label>
      </div>

      <button class="primary" style="width: 100%; margin-top: 12px" :disabled="loading" @click="submit">
        {{ loading ? '登录中...' : '登录' }}
      </button>

      <p v-if="error" class="error" style="margin-top: 10px">{{ error }}</p>
      <p class="meta" style="margin-top: 10px">默认账号来自配置文件：`default_admin_user` / `default_admin_pass`</p>
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
    router.push(typeof route.query.next === 'string' ? route.query.next : '/')
  } catch (e) {
    error.value = e?.response?.data?.error || '登录失败'
  } finally {
    loading.value = false
  }
}
</script>