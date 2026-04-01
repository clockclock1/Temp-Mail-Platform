<template>
  <div class="grid" style="gap: 16px">
    <section class="card">
      <h1>系统配置</h1>
      <p class="meta">配置修改后会写入配置文件，并立即应用可热更新项。</p>
      <p class="meta">需要重启才生效的项：`http_addr`、`smtp_addr`、`db_path`。</p>

      <div class="grid grid-2">
        <label>
          app_name
          <input v-model="form.appName" />
        </label>
        <label>
          web_dir
          <input v-model="form.webDir" />
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

      <label>
        cors_origins（每行一个或逗号分隔）
        <textarea v-model="corsText" rows="4"></textarea>
      </label>

      <div style="display:flex; gap: 8px; margin-top: 12px">
        <button class="primary" @click="save" :disabled="saving">{{ saving ? '保存中...' : '保存并应用' }}</button>
        <button class="secondary" @click="reloadFromFile" :disabled="saving">从配置文件重新加载</button>
      </div>

      <p v-if="error" class="error">{{ error }}</p>
      <p v-if="success" class="success">{{ success }}</p>
      <p v-if="restartRequired" class="error">检测到需要重启服务的配置变更，请重启后端使其完全生效。</p>
    </section>

    <section class="card" v-if="warnings.length">
      <h2>应用提示</h2>
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
    const payload = {
      ...form,
      corsOrigins: parseCorsOrigins(corsText.value),
    }
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
    success.value = '已从配置文件重新加载并应用。'
  } catch (e) {
    error.value = e?.response?.data?.error || '重载失败'
  } finally {
    saving.value = false
  }
}
</script>