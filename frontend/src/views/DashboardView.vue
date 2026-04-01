<template>
  <div class="grid" style="gap: 16px">
    <section class="grid grid-3" v-if="canSeeStats">
      <div class="card" v-for="(value, key) in statsCards" :key="key">
        <p class="meta">{{ key }}</p>
        <h2>{{ value }}</h2>
      </div>
    </section>

    <section class="grid grid-2">
      <div class="card">
        <h2>创建临时邮箱</h2>
        <div class="grid">
          <label>
            邮箱前缀
            <input v-model="createForm.localPart" placeholder="demo" />
          </label>
          <label>
            域名
            <select v-model.number="createForm.domainId">
              <option :value="0">请选择域名</option>
              <option v-for="d in domains" :key="d.id" :value="d.id">{{ d.name }}</option>
            </select>
          </label>
          <label>
            存活时长(小时)
            <input type="number" min="1" max="720" v-model.number="createForm.ttlHours" />
          </label>
          <label>
            描述
            <input v-model="createForm.description" placeholder="用于注册测试" />
          </label>
        </div>
        <button class="primary" style="margin-top: 12px" @click="createMailbox" :disabled="loadingCreate">{{ loadingCreate ? '创建中...' : '创建邮箱' }}</button>
        <p v-if="createErr" class="error">{{ createErr }}</p>
      </div>

      <div class="card">
        <h2>邮箱列表</h2>
        <div class="table-wrap">
          <table>
            <thead>
              <tr>
                <th>地址</th>
                <th>剩余</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in mailboxes" :key="item.id" @click="selectMailbox(item)" style="cursor:pointer">
                <td>
                  <div>{{ item.address }}</div>
                  <div class="meta">{{ item.description || '-' }}</div>
                </td>
                <td>
                  <span class="badge">{{ formatRemaining(item.remainingSeconds) }}</span>
                </td>
                <td>
                  <button class="ghost" @click.stop="copy(item.address)">复制</button>
                  <button class="danger" @click.stop="removeMailbox(item.id)">删除</button>
                </td>
              </tr>
              <tr v-if="mailboxes.length === 0">
                <td colspan="3" class="meta">暂无邮箱</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </section>

    <section class="card" v-if="activeMailbox">
      <h2>邮件列表 - {{ activeMailbox.address }}</h2>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>时间</th>
              <th>发件人</th>
              <th>主题</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="msg in messages" :key="msg.id" @click="selectMessage(msg)" style="cursor:pointer">
              <td>{{ fmtDate(msg.receivedAt) }}</td>
              <td>{{ msg.from }}</td>
              <td>{{ msg.subject || '(无主题)' }}</td>
              <td><button class="danger" @click.stop="removeMessage(msg.id)">删除</button></td>
            </tr>
            <tr v-if="messages.length === 0"><td colspan="4" class="meta">暂无邮件</td></tr>
          </tbody>
        </table>
      </div>
    </section>

    <section class="card" v-if="activeMessage">
      <h2>邮件详情</h2>
      <p><strong>From:</strong> {{ activeMessage.from }}</p>
      <p><strong>To:</strong> {{ activeMessage.to }}</p>
      <p><strong>Subject:</strong> {{ activeMessage.subject || '(无主题)' }}</p>
      <p class="meta">{{ fmtDate(activeMessage.receivedAt) }}</p>
      <p>
        <button class="ghost" @click="downloadRaw(activeMessage.id)">下载原始 EML</button>
      </p>
      <pre style="white-space: pre-wrap; background: rgba(255,255,255,.65); padding: 10px; border-radius: 12px; max-height: 360px; overflow: auto">{{ activeMessage.textBody || activeMessage.htmlBody }}</pre>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { DomainAPI, MailboxAPI, MessageAPI, StatsAPI } from '../api'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const canSeeStats = computed(() => auth.can('stats:read'))

const domains = ref([])
const mailboxes = ref([])
const messages = ref([])
const activeMailbox = ref(null)
const activeMessage = ref(null)
const stats = ref({ users: '-', domains: '-', mailboxes: '-', messages: '-', messagesLast24Hours: '-' })

const createForm = reactive({ localPart: '', domainId: 0, ttlHours: 24, description: '' })
const createErr = ref('')
const loadingCreate = ref(false)

const statsCards = computed(() => ({
  用户总数: stats.value.users,
  域名总数: stats.value.domains,
  邮箱总数: stats.value.mailboxes,
  邮件总数: stats.value.messages,
  '24h 邮件': stats.value.messagesLast24Hours,
}))

onMounted(async () => {
  await Promise.all([loadDomains(), loadMailboxes(), loadStats()])
})

async function loadDomains() {
  const { data } = await DomainAPI.available()
  domains.value = data.items || []
}

async function loadMailboxes() {
  const { data } = await MailboxAPI.list()
  mailboxes.value = data.items || []
}

async function loadStats() {
  if (!canSeeStats.value) return
  const { data } = await StatsAPI.get()
  stats.value = data
}

async function createMailbox() {
  createErr.value = ''
  if (!createForm.localPart || !createForm.domainId) {
    createErr.value = '请先填写邮箱前缀并选择域名'
    return
  }
  loadingCreate.value = true
  try {
    await MailboxAPI.create(createForm)
    createForm.localPart = ''
    createForm.description = ''
    await loadMailboxes()
    await loadStats()
  } catch (e) {
    createErr.value = e?.response?.data?.error || '创建失败'
  } finally {
    loadingCreate.value = false
  }
}

function selectMailbox(item) {
  activeMailbox.value = item
  activeMessage.value = null
  loadMessages(item.id)
}

async function loadMessages(mailboxId) {
  const { data } = await MailboxAPI.messages(mailboxId)
  messages.value = data.items || []
}

async function selectMessage(msg) {
  const { data } = await MessageAPI.get(msg.id)
  activeMessage.value = data.item
}

async function removeMailbox(id) {
  if (!confirm('确定删除该邮箱以及其全部邮件吗？')) return
  await MailboxAPI.remove(id)
  if (activeMailbox.value?.id === id) {
    activeMailbox.value = null
    activeMessage.value = null
    messages.value = []
  }
  await loadMailboxes()
  await loadStats()
}

async function removeMessage(id) {
  await MessageAPI.remove(id)
  messages.value = messages.value.filter((m) => m.id !== id)
  if (activeMessage.value?.id === id) activeMessage.value = null
  await loadStats()
}

async function downloadRaw(id) {
  const { data } = await MessageAPI.raw(id)
  const blobUrl = URL.createObjectURL(data)
  const a = document.createElement('a')
  a.href = blobUrl
  a.download = `message-${id}.eml`
  a.click()
  URL.revokeObjectURL(blobUrl)
}

function formatRemaining(seconds) {
  if (seconds == null || seconds < 0) return '永久'
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  return `${h}h ${m}m`
}

function fmtDate(v) {
  return new Date(v).toLocaleString()
}

async function copy(text) {
  await navigator.clipboard.writeText(text)
}
</script>
