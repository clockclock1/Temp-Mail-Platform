<template>
  <div class="page page-domains">
    <section class="hero-panel domains-hero">
      <div class="hero-copy">
        <span class="eyebrow">Domain Fabric</span>
        <h1 class="hero-title">域名管理</h1>
        <p class="hero-copy-text">
          泛域名规则已对齐旧接口风格：根域加入白名单后，创建地址接口可以直接使用它下面的多级子域名。
        </p>
        <div class="hero-pills">
          <span class="pill">根域白名单</span>
          <span class="pill">子域直接创建</span>
          <span class="pill">MX 验证</span>
        </div>
      </div>

      <div class="hero-note">
        <strong>添加规则</strong>
        <p>添加根域时会先验证该域名是否已经解析了 MX 记录；验证通过后，旧接口可直接使用这个根域下的任意层级子域。</p>
      </div>
    </section>

    <section class="stat-grid stat-grid--hero">
      <div class="stat-item">
        <div class="meta">域名总数</div>
        <div class="stat-value">{{ domainStats.total }}</div>
      </div>
      <div class="stat-item">
        <div class="meta">已启用</div>
        <div class="stat-value">{{ domainStats.enabled }}</div>
      </div>
      <div class="stat-item">
        <div class="meta">MX 已验证</div>
        <div class="stat-value">{{ domainStats.mxVerified }}</div>
      </div>
    </section>

    <section class="surface-grid">
      <div class="card card-accent">
        <div class="panel-head">
          <div>
            <strong>根域名池推送</strong>
            <p class="meta">第一次使用先填写好根域名池，再点推送；后续新增主域时重新推送即可。</p>
          </div>
          <span class="badge ok">Root Pool</span>
        </div>

        <div class="grid" style="margin-top: 14px; gap: 12px">
          <label>
            根域名池
            <textarea v-model="poolText" rows="5" placeholder="example.com&#10;mail.example.net"></textarea>
          </label>

          <div class="grid grid-3">
            <label>
              层级模式
              <select v-model="poolForm.randomLevel">
                <option :value="false">固定层级</option>
                <option :value="true">1-7 级随机</option>
              </select>
            </label>
            <label>
              默认层级
              <select v-model.number="poolForm.level" :disabled="poolForm.randomLevel">
                <option v-for="n in 10" :key="n" :value="n">{{ n }} 级</option>
              </select>
            </label>
            <label v-if="poolForm.randomLevel">
              随机范围
              <input :value="`${poolForm.levelMin}-${poolForm.levelMax} 级`" disabled />
            </label>
          </div>

          <div class="row" v-if="poolForm.randomLevel">
            <label style="max-width: 160px">
              最小级别
              <input type="number" min="1" max="7" v-model.number="poolForm.levelMin" />
            </label>
            <label style="max-width: 160px">
              最大级别
              <input type="number" min="1" max="7" v-model.number="poolForm.levelMax" />
            </label>
          </div>

          <div class="row">
            <button class="primary" :disabled="pushing" @click="pushPool">{{ pushing ? '推送中...' : '推送根域名池' }}</button>
            <button class="ghost" :disabled="pushing" @click="load">刷新</button>
          </div>
          <p v-if="pushMessage" class="success">{{ pushMessage }}</p>
        </div>
      </div>

      <div class="card">
        <div class="panel-head">
          <div>
            <strong>手动添加根域</strong>
            <p class="meta">新增前会验证 MX 记录，失败则不会写入。</p>
          </div>
          <button class="ghost" @click="load">刷新列表</button>
        </div>

        <div class="grid grid-3" style="margin-top: 14px">
          <label>
            根域名
            <input v-model="form.name" placeholder="example.com" />
          </label>
          <label>
            状态
            <select v-model="form.enabled">
              <option :value="true">启用</option>
              <option :value="false">禁用</option>
            </select>
          </label>
          <label>
            快速过滤
            <input v-model="keyword" placeholder="输入域名关键字" />
          </label>
        </div>

        <div class="row" style="margin-top: 12px">
          <button class="primary" @click="create">新增根域</button>
          <button class="ghost" @click="load">刷新</button>
        </div>
        <p v-if="error" class="error" style="margin-top: 8px">{{ error }}</p>
      </div>
    </section>

    <section class="card">
      <div class="panel-head">
        <strong>域名列表</strong>
        <span class="meta">{{ filteredItems.length }} 条记录</span>
      </div>
      <div class="table-wrap" style="margin-top: 12px">
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>域名</th>
              <th>层级</th>
              <th>MX 状态</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in filteredItems" :key="item.id">
              <td>{{ item.id }}</td>
              <td>
                <div>{{ item.name }}</div>
                <div class="meta" v-if="item.mxRecords">{{ formatMX(item.mxRecords) }}</div>
              </td>
              <td>{{ levelLabel(item) }}</td>
              <td>
                <span class="badge" :class="item.mxVerified ? 'ok' : 'off'">
                  {{ item.mxVerified ? '已验证' : '未验证' }}
                </span>
                <div class="meta" v-if="item.mxCheckedAt" style="margin-top: 4px">{{ shortTime(item.mxCheckedAt) }}</div>
              </td>
              <td>
                <span class="badge" :class="item.enabled ? 'ok' : 'off'">
                  {{ item.enabled ? '启用' : '禁用' }}
                </span>
              </td>
              <td>
                <div class="row">
                  <button class="secondary" @click="toggle(item)">{{ item.enabled ? '禁用' : '启用' }}</button>
                  <button class="danger" @click="remove(item.id)">删除</button>
                </div>
              </td>
            </tr>
            <tr v-if="filteredItems.length === 0">
              <td colspan="6" class="meta empty-state">没有匹配的域名。</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { DomainAPI } from '../api'

const form = reactive({ name: '', enabled: true })
const poolText = ref('')
const poolForm = reactive({
  randomLevel: false,
  level: 2,
  levelMin: 1,
  levelMax: 7,
})
const items = ref([])
const keyword = ref('')
const error = ref('')
const pushMessage = ref('')
const pushing = ref(false)

const filteredItems = computed(() => {
  const kw = keyword.value.trim().toLowerCase()
  if (!kw) return items.value
  return items.value.filter((i) => i.name.toLowerCase().includes(kw))
})

const domainStats = computed(() => ({
  total: items.value.length,
  enabled: items.value.filter((i) => i.enabled).length,
  mxVerified: items.value.filter((i) => i.mxVerified).length,
}))

onMounted(load)

async function load() {
  const { data } = await DomainAPI.list()
  items.value = data.items || []
}

async function pushPool() {
  error.value = ''
  pushMessage.value = ''
  const names = poolText.value
    .split(/[\n,]/g)
    .map((v) => v.trim())
    .filter((v) => v)
  if (names.length === 0) {
    error.value = '请先填写根域名池'
    return
  }

  if (poolForm.randomLevel) {
    if (poolForm.levelMin < 1 || poolForm.levelMin > 7 || poolForm.levelMax < 1 || poolForm.levelMax > 7) {
      error.value = '随机层级范围只能是 1-7'
      return
    }
    if (poolForm.levelMin > poolForm.levelMax) {
      error.value = '随机层级最小值不能大于最大值'
      return
    }
  } else if (poolForm.level < 1 || poolForm.level > 10) {
    error.value = '域名层级只能是 1-10'
    return
  }

  pushing.value = true
  try {
    const payload = {
      names,
      enabled: true,
      randomLevel: poolForm.randomLevel,
      level: poolForm.level,
      levelMin: poolForm.levelMin,
      levelMax: poolForm.levelMax,
    }
    const { data } = await DomainAPI.push(payload)
    pushMessage.value = `已推送 ${data.created || 0} 个新根域，更新 ${data.updated || 0} 个根域。`
    await load()
  } catch (e) {
    error.value = e?.response?.data?.error || '推送失败'
  } finally {
    pushing.value = false
  }
}

async function create() {
  error.value = ''
  if (!form.name) {
    error.value = '请输入根域名'
    return
  }
  try {
    await DomainAPI.create(form)
    form.name = ''
    form.enabled = true
    await load()
  } catch (e) {
    error.value = e?.response?.data?.error || '新增失败'
  }
}

async function toggle(item) {
  await DomainAPI.update(item.id, { name: item.name, enabled: !item.enabled })
  await load()
}

async function remove(id) {
  if (!confirm('确定删除这个域名吗？')) return
  await DomainAPI.remove(id)
  await load()
}

function levelLabel(item) {
  if (item.randomLevel) {
    const min = item.levelMin || 1
    const max = item.levelMax || 7
    return `${min}-${max}级随机`
  }
  return `${item.level || countLabels(item.name)}级`
}

function countLabels(name) {
  const count = String(name || '')
    .split('.')
    .map((v) => v.trim())
    .filter(Boolean).length
  return count > 0 ? count : 2
}

function formatMX(text) {
  return String(text).split('\n').join(' / ')
}

function shortTime(value) {
  return new Date(value).toLocaleString()
}
</script>
