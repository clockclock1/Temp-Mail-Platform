<template>
  <div class="login-wrap">
    <div class="login-stage">
      <section class="login-showcase">
        <span class="eyebrow">Mailbox Control Center</span>
        <h1 class="login-title">让临时邮箱管理，像一套真正的运营控制台。</h1>
        <p class="login-copy">
          在这里统一完成地址生成、邮件预览、域名池推送与权限控制。界面更轻，操作更顺，状态反馈也更清楚。
        </p>

        <div class="login-points">
          <article class="login-point">
            <strong>实时收件</strong>
            <span>地址、邮件和过期状态集中显示。</span>
          </article>
          <article class="login-point">
            <strong>多级域名</strong>
            <span>支持域名池推送与 1-10 级策略配置。</span>
          </article>
          <article class="login-point">
            <strong>权限分层</strong>
            <span>角色、用户、配置入口统一收束。</span>
          </article>
        </div>
      </section>

      <section class="card login-card">
        <div class="login-card-head">
          <span class="eyebrow">Secure Sign-in</span>
          <h2 class="section-title">登录控制台</h2>
          <p class="section-sub">管理员与授权用户可创建临时邮箱、查看邮件、管理系统。</p>
        </div>

        <div class="grid" style="margin-top: 18px">
          <label>
            用户名
            <input v-model="form.username" placeholder="admin" @keyup.enter="submit" />
          </label>
          <label>
            密码
            <input v-model="form.password" type="password" placeholder="请输入密码" @keyup.enter="submit" />
          </label>
        </div>

        <button class="primary login-submit" :disabled="loading" @click="submit">
          {{ loading ? '登录中...' : '进入控制台' }}
        </button>

        <p v-if="error" class="error" style="margin-top: 12px">{{ error }}</p>
        <p class="meta login-meta">默认账号来自配置文件：`default_admin_user` / `default_admin_pass`</p>
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
