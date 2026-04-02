<template>
  <div class="page grid" style="gap: 12px">
    <section class="card soft">
      <h1 class="section-title">系统配置中心</h1>
      <p class="section-sub">保存即写入配置文件并应用运行时配置。涉及监听地址和数据库路径的修改会提示重启。</p>
    </section>

    <section class="card">
      <div class="grid grid-3">
        <label>
          app_name
          <input v-model="form.appName" />
        </label>
        <label>
          http_addr
          <input v-model="form.httpAddr" />
        </label>
        <label>
          smtp_addr
          <input v-model="form.smtpAddr" />
        </label>

        <label>
          web_dir
          <input v-model="form.webDir" />
        </label>
        <label>
          db_path
          <input v-model="form.dbPath" />
        </label>
        <label>
          data_dir
          <input v-model="form.dataDir" />
        </label>

        <label>
          jwt_secret
          <input v-model="form.jwtSecret" />
        </label>
        <label>
          jwt_expire_hours
          <input type="number" min="1" v-model.number="form.jwtExpireHours" />
        </label>
        <label>
          legacy_admin_auth
          <input v-model="form.legacyAdminAuth" />
        </label>

        <label>
          legacy_custom_auth
          <input v-model="form.legacyCustomAuth" />
        </label>
        <label>
          legacy_address_jwt_expire_hours
          <input type="number" min="1" v-model.number="form.legacyAddrExpire" />
        </label>
        <label>
          cleanup_interval_minutes
          <input type="number" min="1" v-model.number="form.cleanupIntervalMinutes" />
        </label>

        <label>
          default_admin_user
          <input v-model="form.defaultAdminUser" />
        </label>
        <label>
          default_admin_pass
          <input v-model="form.defaultAdminPass" />
        </label>
      </div>

      <label style="margin-top: 10px; display: block">
        cors_origins（每行一个）
        <textarea v-model="corsText" rows="4"></textarea>
      </label>

      <div class="row" style="margin-top: 10px">
        <button class="primary" :disabled="saving" @click="save">{{ saving ? '保存中...' : '保存并应用' }}</button>
        <button class="secondary" :disabled="saving" @click="reloadFromFile">从文件重载</button>
      </div>

      <p v-if="success" class="success" style="margin-top: 8px">{{ success }}</p>
      <p v-if="error" class="error" style="margin-top: 8px">{{ error }}</p>
      <p v-if="restartRequired" class="error" style="margin-top: 8px">部分配置需要重启服务才能完全生效。</p>
    </section>

    <section class="card" v-if="warnings.length">
      <h2 class="section-title" style="font-size: 16px">应用提示</h2>
      <ul>
        <li v-for="(w, i) in warnings" :key="i">{{ w }}</li>
      </ul>
    </section>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { SystemAPI } from '../api'

const form = reactive({
  appName: '',
  httpAddr: '',
  smtpAddr: '',
  webDir: '',
  jwtSecret: '',
  jwtExpireHours: 24,
  legacyAdminAuth: '',
  legacyCustomAuth: '',
  legacyAddrExpire: 720,
  dbPath: '',
  dataDir: '',
  corsOrigins: [],
  defaultAdminUser: '',
  defaultAdminPass: '',
  cleanupIntervalMinutes: 10,
})

const corsText = ref('')
const warnings = ref([])
const restartRequired = ref(false)
const error = ref('')
const success = ref('')
const saving = ref(false)

onMounted(load)

async function load() {
  error.value = ''
  const { data } = await SystemAPI.getConfig()
  assignForm(data.item)
}

function assignForm(item) {
  Object.assign(form, item)
  corsText.value = (item.corsOrigins || []).join('\n')
}

function parseCorsOrigins(input) {
  return input
    .split(/[\n,]/g)
    .map((v) => v.trim())
    .filter((v) => v)
}

async function save() {
  saving.value = true
  error.value = ''
  success.value = ''
  warnings.value = []
  restartRequired.value = false

  try {
    const payload = { ...form, corsOrigins: parseCorsOrigins(corsText.value) }
    const { data } = await SystemAPI.updateConfig(payload)
    assignForm(data.item)
    warnings.value = data.warnings || []
    restartRequired.value = !!data.restartRequired
    success.value = '配置已保存并应用。'
  } catch (e) {
    error.value = e?.response?.data?.error || '保存失败'
  } finally {
    saving.value = false
  }
}

async function reloadFromFile() {
  saving.value = true
  error.value = ''
  success.value = ''
  warnings.value = []
  restartRequired.value = false

  try {
    const { data } = await SystemAPI.reloadConfig()
    assignForm(data.item)
    warnings.value = data.warnings || []
    restartRequired.value = !!data.restartRequired
    success.value = '已从配置文件重载并应用。'
  } catch (e) {
    error.value = e?.response?.data?.error || '重载失败'
  } finally {
    saving.value = false
  }
}
</script>