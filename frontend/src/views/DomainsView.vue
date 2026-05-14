<template>
  <div class="page page-domains">
    <section class="hero-panel domains-hero">
      <div class="hero-copy">
        <span class="eyebrow">Domain Fabric</span>
        <h1 class="hero-title">域名管理</h1>
        <p class="hero-copy-text">
          根域名添加前会先校验 MX 记录，校验通过后才会写入。开启多级域名策略后，
          系统会按设定层级自动生成子域，接口调用行为与兼容模式保持一致。
        </p>
        <div class="hero-pills">
          <span class="pill">根域白名单</span>
          <span class="pill">1-10 级层级</span>
          <span class="pill">MX 校验</span>
          <span class="pill">兼容旧接口</span>
        </div>
      </div>

      <div class="hero-note">
        <strong>使用建议</strong>
        <p>
          第一次使用请先填写发信域名池并点击推送。后续新增根域名时，只需要重新推送一次，
          已有规则会持续生效。
        </p>
      </div>
    </section>

    <section class="stat-grid stat-grid--hero">
      <div class="stat-item">
        <div class="meta">域名总数</div>
        <div class="stat-value">{{ domainStats.total }}</div>
      </div>
      <div class="stat-item">
        <div class="meta">启用中</div>
        <div class="stat-value">{{ domainStats.enabled }}</div>
      </div>
      <div class="stat-item">
        <div class="meta">MX 已验证</div>
        <div class="stat-value">{{ domainStats.mxVerified }}</div>
      </div>
      <div class="stat-item">
        <div class="meta">随机层级</div>
        <div class="stat-value">{{ domainStats.randomLevel }}</div>
      </div>
      <div class="stat-item">
        <div class="meta">固定层级</div>
        <div class="stat-value">{{ domainStats.fixedLevel }}</div>
      </div>
    </section>

    <section class="surface-grid">
      <div class="card card-accent">
        <div class="panel-head">
          <div>
            <strong>根域名池推送</strong>
            <p class="meta">支持固定层级和随机层级。推送时会逐个验证 MX，验证失败会直接阻止写入。</p>
          </div>
          <span class="badge ok">Pool Push</span>
        </div>

        <div class="grid" style="margin-top: 14px">
          <label>
            根域名池
            <textarea
              v-model="poolText"
              rows="6"
              placeholder="example.com&#10;mail.example.net&#10;demo.example.org"
            ></textarea>
          </label>

          <div class="grid grid-3">
            <label>
              层级策略
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
            <label>
              当前模式
              <input :value="strategyText(poolForm)" disabled />
            </label>
          </div>

          <div class="row wrap" v-if="poolForm.randomLevel">
            <label class="compact-field">
              最小层级
              <input v-model.number="poolForm.levelMin" type="number" min="1" max="7" />
            </label>
            <label class="compact-field">
              最大层级
              <input v-model.number="poolForm.levelMax" type="number" min="1" max="7" />
            </label>
          </div>

          <div class="row wrap">
            <button class="primary" :disabled="pushing" @click="pushPool">
              {{ pushing ? '推送中...' : '推送根域名池' }}
            </button>
            <button class="ghost" :disabled="pushing" @click="load">刷新列表</button>
          </div>

          <p v-if="pushMessage" class="success">{{ pushMessage }}</p>
        </div>
      </div>

      <div class="card">
        <div class="panel-head">
          <div>
            <strong>手动添加根域</strong>
            <p class="meta">新增前会验证 MX 记录。失败时不会写入列表，避免无效根域进入生产配置。</p>
          </div>
          <button class="ghost" @click="load">刷新</button>
        </div>

        <div class="grid grid-2" style="margin-top: 14px">
          <label>
            根域名
            <input v-model.trim="form.name" placeholder="example.com" />
          </label>
          <label>
            状态
            <select v-model="form.enabled">
              <option :value="true">启用</option>
              <option :value="false">禁用</option>
            </select>
          </label>
        </div>

        <div class="grid grid-3" style="margin-top: 12px">
          <label>
            层级策略
            <select v-model="form.randomLevel">
              <option :value="false">固定层级</option>
              <option :value="true">1-7 级随机</option>
            </select>
          </label>
          <label>
            默认层级
            <select v-model.number="form.level" :disabled="form.randomLevel">
              <option v-for="n in 10" :key="n" :value="n">{{ n }} 级</option>
            </select>
          </label>
          <label>
            快速过滤
            <input v-model.trim="keyword" placeholder="输入域名关键字" />
          </label>
        </div>

        <div class="row wrap" style="margin-top: 12px" v-if="form.randomLevel">
          <label class="compact-field">
            最小层级
            <input v-model.number="form.levelMin" type="number" min="1" max="7" />
          </label>
          <label class="compact-field">
            最大层级
            <input v-model.number="form.levelMax" type="number" min="1" max="7" />
          </label>
        </div>

        <div class="row wrap" style="margin-top: 12px">
          <button class="primary" @click="create">新增根域</button>
          <button class="ghost" @click="resetForm">重置表单</button>
        </div>
        <p v-if="error" class="error" style="margin-top: 8px">{{ error }}</p>
      </div>
    </section>

    <section class="guide-grid">
      <article class="card guide-card">
        <div class="panel-head">
          <strong>MX 绑定指引</strong>
          <span class="badge ok">必做</span>
        </div>
        <ol class="dns-steps">
          <li>先准备一个可被公网访问的收信主机，例如 <code>mail.example.com</code>，并让它指向当前服务器公网 IP。</li>
          <li>在根域名处添加 <code>MX</code> 记录，让 <code>@</code> 指向你的收信主机，优先级建议为 <code>10</code>。</li>
          <li>如果要启用 1-10 级多级子域邮箱，再补充一个通配子域解析 <code>*</code> 指向同一入口。</li>
          <li>确认宿主机已放行并映射 SMTP 端口，外部邮件服务器需要能连到 <code>25</code> 端口。</li>
          <li>等待 DNS 生效后，再回到本页面新增根域；系统会立即执行 MX 校验。</li>
        </ol>
      </article>

      <article class="card guide-card guide-card--soft">
        <div class="panel-head">
          <strong>推荐记录示例</strong>
          <span class="badge">DNS</span>
        </div>
        <div class="guide-kv">
          <div><span>主机记录</span><code>@</code></div>
          <div><span>记录类型</span><code>MX</code></div>
          <div><span>记录值</span><code>mail.example.com</code></div>
          <div><span>优先级</span><code>10</code></div>
          <div><span>泛解析</span><code>* -> 同一收信入口</code></div>
          <div><span>检测方式</span><code>新增 / 推送时自动验证</code></div>
        </div>
        <p class="meta" style="margin-top: 12px">
          如果新增时提示 <code>lookup ... no such host</code>，通常表示域名本身未解析、MX 记录未生效，
          或容器无法通过当前 DNS 解析到该根域。
        </p>
      </article>
    </section>

    <section class="card">
      <div class="panel-head">
        <div>
          <strong>根域名列表</strong>
          <p class="meta">当前共 {{ filteredItems.length }} 条匹配记录，可在此查看验证状态、层级策略和 MX 明细。</p>
        </div>
        <button class="ghost" @click="load">刷新列表</button>
      </div>

      <div class="table-wrap" style="margin-top: 12px">
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>根域名</th>
              <th>层级策略</th>
              <th>MX 验证</th>
              <th>MX 记录</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in filteredItems" :key="item.id">
              <td>{{ item.id }}</td>
              <td>
                <div class="table-domain-name">{{ item.name }}</div>
                <div class="meta" v-if="item.mxCheckedAt">最近检测：{{ shortTime(item.mxCheckedAt) }}</div>
              </td>
              <td>{{ levelLabel(item) }}</td>
              <td>
                <span class="badge" :class="item.mxVerified ? 'ok' : 'off'">
                  {{ item.mxVerified ? '已验证' : '未验证' }}
                </span>
              </td>
              <td>
                <div class="mx-records" v-if="item.mxRecords">
                  <code v-for="record in mxList(item.mxRecords)" :key="record">{{ record }}</code>
                </div>
                <span class="meta" v-else>暂无记录</span>
              </td>
              <td>
                <span class="badge" :class="item.enabled ? 'ok' : 'off'">
                  {{ item.enabled ? '启用' : '禁用' }}
                </span>
              </td>
              <td>
                <div class="row wrap">
                  <button class="secondary" @click="recheckMx(item)">重验 MX</button>
                  <button class="ghost" @click="toggle(item)">
                    {{ item.enabled ? '禁用' : '启用' }}
                  </button>
                  <button class="danger" @click="remove(item.id)">删除</button>
                </div>
              </td>
            </tr>
            <tr v-if="filteredItems.length === 0">
              <td colspan="7" class="meta empty-state">没有匹配的根域名记录。</td>
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

const items = ref([])
const keyword = ref('')
const error = ref('')
const pushMessage = ref('')
const pushing = ref(false)

const poolText = ref('')
const poolForm = reactive({
  randomLevel: false,
  level: 2,
  levelMin: 1,
  levelMax: 7,
})

const form = reactive({
  name: '',
  enabled: true,
  randomLevel: false,
  level: 2,
  levelMin: 1,
  levelMax: 7,
})

const filteredItems = computed(() => {
  const q = keyword.value.trim().toLowerCase()
  if (!q) return items.value
  return items.value.filter((item) => item.name.toLowerCase().includes(q))
})

const domainStats = computed(() => ({
  total: items.value.length,
  enabled: items.value.filter((item) => item.enabled).length,
  mxVerified: items.value.filter((item) => item.mxVerified).length,
  randomLevel: items.value.filter((item) => item.randomLevel).length,
  fixedLevel: items.value.filter((item) => !item.randomLevel).length,
}))

onMounted(load)

async function load() {
  error.value = ''
  pushMessage.value = ''
  try {
    const { data } = await DomainAPI.list()
    items.value = data.items || []
  } catch (e) {
    error.value = e?.response?.data?.error || '加载域名列表失败。'
  }
}

function normalizeRange(model) {
  const next = { ...model }
  if (next.randomLevel) {
    next.levelMin = clamp(next.levelMin, 1, 7)
    next.levelMax = clamp(next.levelMax, 1, 7)
    if (next.levelMin > next.levelMax) {
      throw new Error('随机层级的最小值不能大于最大值。')
    }
  } else {
    next.level = clamp(next.level, 1, 10)
  }
  return next
}

async function pushPool() {
  error.value = ''
  pushMessage.value = ''

  const names = poolText.value
    .split(/[\n,]/g)
    .map((value) => value.trim())
    .filter(Boolean)

  if (!names.length) {
    error.value = '请先填写根域名池。'
    return
  }

  let strategy
  try {
    strategy = normalizeRange(poolForm)
  } catch (e) {
    error.value = e.message
    return
  }

  pushing.value = true
  try {
    const { data } = await DomainAPI.push({
      names,
      enabled: true,
      randomLevel: strategy.randomLevel,
      level: strategy.level,
      levelMin: strategy.levelMin,
      levelMax: strategy.levelMax,
    })
    pushMessage.value = `推送完成：新增 ${data.created || 0} 个根域，更新 ${data.updated || 0} 个根域。`
    await load()
  } catch (e) {
    error.value = e?.response?.data?.error || '推送失败，请检查 MX 和 DNS 解析状态。'
  } finally {
    pushing.value = false
  }
}

async function create() {
  error.value = ''
  if (!form.name) {
    error.value = '请输入根域名。'
    return
  }

  let strategy
  try {
    strategy = normalizeRange(form)
  } catch (e) {
    error.value = e.message
    return
  }

  try {
    await DomainAPI.create({
      name: form.name,
      enabled: form.enabled,
      randomLevel: strategy.randomLevel,
      level: strategy.level,
      levelMin: strategy.levelMin,
      levelMax: strategy.levelMax,
    })
    resetForm()
    await load()
  } catch (e) {
    error.value = e?.response?.data?.error || '新增根域失败。'
  }
}

async function toggle(item) {
  error.value = ''
  try {
    await DomainAPI.update(item.id, {
      name: item.name,
      enabled: !item.enabled,
      randomLevel: item.randomLevel,
      level: item.level,
      levelMin: item.levelMin,
      levelMax: item.levelMax,
    })
    await load()
  } catch (e) {
    error.value = e?.response?.data?.error || '更新域名状态失败。'
  }
}

async function recheckMx(item) {
  error.value = ''
  pushMessage.value = ''
  try {
    await DomainAPI.push({
      names: [item.name],
      enabled: item.enabled,
      randomLevel: item.randomLevel,
      level: item.level,
      levelMin: item.levelMin,
      levelMax: item.levelMax,
    })
    pushMessage.value = `已重新验证 ${item.name} 的 MX 记录。`
    await load()
  } catch (e) {
    error.value = e?.response?.data?.error || 'MX 重新验证失败。'
  }
}

async function remove(id) {
  if (!confirm('确定删除这个根域名吗？')) return
  error.value = ''
  try {
    await DomainAPI.remove(id)
    await load()
  } catch (e) {
    error.value = e?.response?.data?.error || '删除根域名失败。'
  }
}

function resetForm() {
  form.name = ''
  form.enabled = true
  form.randomLevel = false
  form.level = 2
  form.levelMin = 1
  form.levelMax = 7
}

function levelLabel(item) {
  if (item.randomLevel) {
    return `${item.levelMin || 1}-${item.levelMax || 7} 级随机`
  }
  return `${item.level || 2} 级固定`
}

function strategyText(model) {
  return model.randomLevel ? `${model.levelMin}-${model.levelMax} 级随机` : `${model.level} 级固定`
}

function mxList(records) {
  return String(records)
    .split('\n')
    .map((value) => value.trim())
    .filter(Boolean)
}

function shortTime(value) {
  return new Date(value).toLocaleString('zh-CN')
}

function clamp(value, min, max) {
  const num = Number(value)
  if (Number.isNaN(num)) return min
  return Math.min(max, Math.max(min, num))
}
</script>
