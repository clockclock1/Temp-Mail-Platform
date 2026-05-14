<template>
  <div class="login-wrap">
    <div class="login-stage">
      <section class="login-showcase login-showcase--auth">
        <span class="eyebrow">Access Portal</span>
        <h1 class="login-title">登录到临时邮箱控制台</h1>
        <p class="login-copy">
          通过标准账号密码登录进入系统，统一管理邮箱地址、域名池、权限和运行配置。
          登录后会自动校验身份状态，避免旧会话继续访问后台。
        </p>

        <div class="login-points">
          <article class="login-point">
            <strong>标准会话管理</strong>
            <span>支持记住登录、自动携带 Bearer Token，并在失效后回到登录页。</span>
          </article>
          <article class="login-point">
            <strong>权限分层可控</strong>
            <span>管理员与普通成员使用同一入口，根据角色自动显示可访问模块。</span>
          </article>
          <article class="login-point">
            <strong>适合生产环境</strong>
            <span>默认管理员来自配置文件，首次部署后建议立即在系统内修改账号密码。</span>
          </article>
        </div>
      </section>

      <section class="login-card login-card--auth">
        <div class="login-card-head">
          <span class="eyebrow">Secure Sign In</span>
          <h2 class="section-title">账号登录</h2>
          <p class="section-sub">请输入系统账号和密码，登录后将自动恢复上次可用的工作区权限。</p>
        </div>

        <form class="login-form" @submit.prevent="submit">
          <label class="login-field">
            <span>用户名</span>
            <input
              v-model.trim="form.username"
              autocomplete="username"
              placeholder="请输入用户名，例如 admin"
            />
          </label>

          <label class="login-field">
            <span>密码</span>
            <div class="password-field">
              <input
                v-model="form.password"
                :type="showPassword ? 'text' : 'password'"
                autocomplete="current-password"
                placeholder="请输入登录密码"
              />
              <button class="password-toggle" type="button" @click="showPassword = !showPassword">
                {{ showPassword ? '隐藏' : '显示' }}
              </button>
            </div>
          </label>

          <div class="login-options">
            <label class="check-row">
              <input v-model="form.remember" type="checkbox" />
              <span>记住登录</span>
            </label>
            <span class="meta">关闭时仅保存当前会话</span>
          </div>

          <button class="primary login-submit" :disabled="loading" type="submit">
            {{ loading ? '正在登录...' : '登录系统' }}
          </button>
        </form>

        <p v-if="error" class="login-alert error">{{ error }}</p>

        <div class="login-help">
          <div>
            <strong>首次登录提示</strong>
            <p class="meta">默认管理员账号读取 <code>default_admin_user</code> / <code>default_admin_pass</code>，建议登录后立刻修改。</p>
          </div>
          <div>
            <strong>安全建议</strong>
            <p class="meta">生产环境请搭配 HTTPS、强密码策略和独立管理员账号使用。</p>
          </div>
        </div>
      </section>
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

const form = reactive({
  username: '',
  password: '',
  remember: true,
})

const loading = ref(false)
const error = ref('')
const showPassword = ref(false)

async function submit() {
  error.value = ''
  if (!form.username || !form.password) {
    error.value = '请输入用户名和密码。'
    return
  }

  loading.value = true
  try {
    const { data } = await AuthAPI.login({
      username: form.username,
      password: form.password,
    })
    auth.setAuth(data, form.remember)
    router.push(typeof route.query.next === 'string' ? route.query.next : '/')
  } catch (e) {
    error.value = e?.response?.data?.error || '登录失败，请检查账号密码后重试。'
  } finally {
    loading.value = false
  }
}
</script>
