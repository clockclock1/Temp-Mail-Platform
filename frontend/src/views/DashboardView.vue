<template>
  <div class="page grid" style="gap: 12px">
    <section v-if="canSeeStats" class="stat-grid">
      <div class="stat-item" v-for="(item, idx) in statItems" :key="idx">
        <div class="meta">{{ item.label }}</div>
        <div class="stat-value">{{ item.value }}</div>
      </div>
    </section>

    <section class="card soft">
      <div class="row wrap" style="justify-content: space-between">
        <div>
          <h2 class="section-title">收件台</h2>
          <p class="section-sub">左侧邮箱，中间邮件列表，右侧正文预览</p>
        </div>
        <div class="row wrap">
          <input v-model="createForm.localPart" placeholder="邮箱前缀，例如 qa-user" style="width: 210px" />
          <select v-model.number="createForm.domainId" style="width: 190px">
            <option :value="0">选择域名</option>
            <option v-for="d in domains" :key="d.id" :value="d.id">{{ d.name }}</option>
          </select>
          <input v-model.number="createForm.ttlHours" type="number" min="1" max="720" style="width: 110px" />
          <button class="primary" :disabled="creating" @click="createMailbox">{{ creating ? '创建中...' : '创建邮箱' }}</button>
        </div>
      </div>
      <p v-if="error" class="error" style="margin-top: 8px">{{ error }}</p>
    </section>

    <section class="layout-mail">
      <div class="card list-panel">
        <div class="row" style="justify-content: space-between; margin-bottom: 8px">
          <strong>邮箱列表</strong>
          <button class="ghost" @click="loadMailboxes">刷新</button>
        </div>
        <div class="item-list">
          <article
            v-for="m in mailboxes"
            :key="m.id"
            class="item"
            :class="{ active: activeMailbox?.id === m.id }"
            @click="selectMailbox(m)"
          >
            <div class="row" style="justify-content: space-between">
              <strong>{{ m.localPart }}</strong>
              <span class="badge" :class="m.enabled ? 'ok' : 'off'">{{ m.enabled ? '启用' : '禁用' }}</span>
            </div>
            <div class="meta" style="margin-top: 4px">@{{ m.domain?.name }}</div>
            <div class="meta" style="margin-top: 2px">{{ formatRemaining(m.remainingSeconds) }}</div>
            <div class="row" style="margin-top: 8px">
              <button class="ghost" @click.stop="copy(m.address)">复制</button>
              <button class="danger" @click.stop="removeMailbox(m.id)">删除</button>
            </div>
          </article>

          <p v-if="mailboxes.length === 0" class="meta">暂无邮箱，先在上方创建一个。</p>
        </div>
      </div>

      <div class="card list-panel">
        <div class="row" style="justify-content: space-between; margin-bottom: 8px">
          <strong>邮件列表</strong>
          <input v-model="mailFilter" placeholder="按主题/发件人过滤" style="width: 190px" />
        </div>

        <div class="item-list">
          <article
            v-for="msg in filteredMessages"
            :key="msg.id"
            class="item"
            :class="{ active: activeMessage?.id === msg.id }"
            @click="selectMessage(msg)"
          >
            <div class="row" style="justify-content: space-between">
              <strong>{{ msg.subject || '(无主题)' }}</strong>
              <span class="meta">{{ shortTime(msg.receivedAt) }}</span>
            </div>
            <div class="meta" style="margin-top: 4px">{{ msg.from || '未知发件人' }}</div>
            <div class="row" style="margin-top: 8px">
              <button class="danger" @click.stop="removeMessage(msg.id)">删除</button>
            </div>
          </article>

          <p v-if="!activeMailbox" class="meta">先在左侧选择邮箱。</p>
          <p v-else-if="filteredMessages.length === 0" class="meta">该邮箱暂无邮件。</p>
        </div>
      </div>

      <div class="card preview">
        <div v-if="activeMessage">
          <h3 class="section-title" style="margin-bottom: 8px">{{ activeMessage.subject || '(无主题)' }}</h3>
          <p class="meta"><strong>From:</strong> {{ activeMessage.from }}</p>
          <p class="meta"><strong>To:</strong> {{ activeMessage.to }}</p>
          <p class="meta"><strong>时间:</strong> {{ fullTime(activeMessage.receivedAt) }}</p>
          <div class="row" style="margin: 10px 0">
            <button class="secondary" @click="downloadRaw(activeMessage.id)">下载原始 EML</button>
          </div>
          <pre class="code">{{ activeMessage.textBody || activeMessage.htmlBody || '(无正文)' }}</pre>
        </div>
        <p v-else class="meta">选择一封邮件查看详情。</p>
      </div>
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
const mailFilter = ref('')
const error = ref('')
const creating = ref(false)

const createForm = reactive({
  localPart: '',
  domainId: 0,
  ttlHours: 24,
  description: 'console-created',
})

const statItems = computed(() => [
  { label: '用户总数', value: stats.value.users },
  { label: '域名总数', value: stats.value.domains },
  { label: '邮箱总数', value: stats.value.mailboxes },
  { label: '邮件总数', value: stats.value.messages },
  { label: '24 小时新增邮件', value: stats.value.messagesLast24Hours },
])

const filteredMessages = computed(() => {
  const kw = mailFilter.value.trim().toLowerCase()
  if (!kw) return messages.value
  return messages.value.filter((m) =>
    [m.subject, m.from].some((v) => String(v || '').toLowerCase().includes(kw)),
  )
})

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
  if (activeMailbox.value) {
    const current = mailboxes.value.find((m) => m.id === activeMailbox.value.id)
    if (!current) {
      activeMailbox.value = null
      activeMessage.value = null
      messages.value = []
    }
  }
}

async function loadStats() {
  if (!canSeeStats.value) return
  const { data } = await StatsAPI.get()
  stats.value = data
}

async function createMailbox() {
  error.value = ''
  if (!createForm.localPart || !createForm.domainId) {
    error.value = '请输入邮箱前缀并选择域名'
    return
  }

  creating.value = true
  try {
    await MailboxAPI.create(createForm)
    createForm.localPart = ''
    await Promise.all([loadMailboxes(), loadStats()])
  } catch (e) {
    error.value = e?.response?.data?.error || '创建失败'
  } finally {
    creating.value = false
  }
}

async function selectMailbox(mailbox) {
  activeMailbox.value = mailbox
  activeMessage.value = null
  mailFilter.value = ''
  const { data } = await MailboxAPI.messages(mailbox.id)
  messages.value = data.items || []
}

async function selectMessage(message) {
  const { data } = await MessageAPI.get(message.id)
  activeMessage.value = data.item
}

async function removeMailbox(id) {
  if (!confirm('确定删除这个邮箱和其全部邮件吗？')) return
  await MailboxAPI.remove(id)
  if (activeMailbox.value?.id === id) {
    activeMailbox.value = null
    activeMessage.value = null
    messages.value = []
  }
  await Promise.all([loadMailboxes(), loadStats()])
}

async function removeMessage(id) {
  if (!confirm('确定删除该邮件吗？')) return
  await MessageAPI.remove(id)
  messages.value = messages.value.filter((m) => m.id !== id)
  if (activeMessage.value?.id === id) activeMessage.value = null
  await loadStats()
}

async function downloadRaw(id) {
  const { data } = await MessageAPI.raw(id)
  const url = URL.createObjectURL(data)
  const a = document.createElement('a')
  a.href = url
  a.download = `message-${id}.eml`
  a.click()
  URL.revokeObjectURL(url)
}

async function copy(text) {
  await navigator.clipboard.writeText(text)
}

function formatRemaining(seconds) {
  if (seconds == null || seconds < 0) return '不过期'
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  return `剩余 ${h}h ${m}m`
}

function shortTime(v) {
  return new Date(v).toLocaleTimeString()
}

function fullTime(v) {
  return new Date(v).toLocaleString()
}
</script>