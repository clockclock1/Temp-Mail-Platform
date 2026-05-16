<template>
  <div class="page page-domains">
    <section class="hero-panel domains-hero">
      <div class="hero-copy">
        <span class="eyebrow">Domain Fabric</span>
        <h1 class="hero-title">域名管理</h1>
        <p class="hero-copy-text">
          这里只维护固定根域名。新邮箱会直接使用这个根域，任意子域名地址是否能收信，交给 DNS 的通配记录和 MX
          入口来负责。
        </p>
        <div class="hero-pills">
          <span class="pill">固定接收域</span>
          <span class="pill">MX 校验</span>
          <span class="pill">泛子域路由</span>
        </div>
      </div>

      <div class="hero-note">
        <strong>推荐模式</strong>
        <p>
          先添加一个真实可收信的根域名，例如 <code>example.com</code>。系统里创建出的邮箱会固定为
          <code>user@example.com</code>，同时也会接受类似 <code>user@foo.example.com</code>、
          <code>user@bar.foo.example.com</code> 的泛子域来信。
        </p>
      </div>
    </section>

    <section class="stat-grid stat-grid--hero">
      <div class="stat-item">
        <div class="meta">根域总数</div>
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
            <strong>批量推送根域名</strong>
            <p class="meta">每个根域都会先验证 MX，验证通过才会写入。</p>
          </div>
          <span class="badge ok">Pool Push</span>
        </div>

        <div class="grid" style="margin-top: 14px">
          <label>
            根域名池
            <textarea
              v-model="poolText"
              rows="6"
              placeholder="example.com&#10;example.net&#10;example.org"
            ></textarea>
          </label>

          <div class="row wrap">
            <button class="primary" :disabled="pushing" @click="pushPool">
              {{ pushing ? '推送中...' : '推送根域名池' }}
            </button>
            <button class="ghost" :disabled="pushing" @click="load">刷新列表</button>
          </div>

          <p v-if="pushMessage" class="success">{{ pushMessage }}</p>
          <p class="meta">
            推送成功后，新建邮箱会固定绑定到对应根域；如果你同时配置了通配 DNS，任意子域地址也会被路由到同一个邮箱。
          </p>
        </div>
      </div>

      <div class="card">
        <div class="panel-head">
          <div>
            <strong>手动添加根域</strong>
            <p class="meta">新增前会即时验证 MX 记录，失败时不会写入。</p>
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

        <div class="grid grid-2" style="margin-top: 12px">
          <label>
            接收模式
            <input value="固定根域 + 泛子域路由" disabled />
          </label>
          <label>
            快速过滤
            <input v-model.trim="keyword" placeholder="输入域名关键字" />
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
          <strong>DNS / MX 绑定指引</strong>
          <span class="badge ok">必做</span>
        </div>
        <ol class="dns-steps">
          <li>先准备一个收信主机，例如 <code>mail.example.com</code>，并让它解析到你的服务器公网 IP。</li>
          <li>添加根域 MX：<code>@ MX 10 mail.example.com</code>。</li>
          <li>添加通配 MX：<code>* MX 10 mail.example.com</code>，让任意子域邮箱都走同一个 SMTP 入口。</li>
          <li>建议再添加通配解析：<code>* A 服务器公网 IP</code>，方便子域探测和部分隐式投递场景。</li>
          <li>确认宿主机公网 <code>25</code> 端口已放行，并映射到容器内的 <code>2525</code>。</li>
          <li>等待 DNS 生效后，再回到本页面新增根域；系统会立即校验根域 MX。</li>
        </ol>
      </article>

      <article class="card guide-card guide-card--soft">
        <div class="panel-head">
          <strong>推荐记录示例</strong>
          <span class="badge">DNS</span>
        </div>
        <div class="guide-kv">
          <div><span>mail 主机</span><code>mail A 服务器公网 IP</code></div>
          <div><span>根域 MX</span><code>@ MX 10 mail.example.com</code></div>
          <div><span>通配 MX</span><code>* MX 10 mail.example.com</code></div>
          <div><span>通配解析</span><code>* A 服务器公网 IP</code></div>
          <div><span>容器端口</span><code>25 -> 2525</code></div>
          <div><span>系统收件</span><code>user@example.com / user@any.example.com</code></div>
        </div>
        <p class="meta" style="margin-top: 12px">
          如果新增时提示 <code>lookup ... no such host</code>，通常表示根域本身还没有生效，或者容器当前 DNS 无法解析该域名。
          现在系统会自动尝试备用 DNS，但根域和 MX 记录本身仍然必须真实存在。
        </p>
      </article>
    </section>

    <section class="card">
      <div class="panel-head">
        <div>
          <strong>根域名列表</strong>
          <p class="meta">当前共 {{ filteredItems.length }} 条匹配记录，可查看验证状态与 MX 明细。</p>
        </div>
        <button class="ghost" @click="load">刷新列表</button>
      </div>

      <div class="table-wrap" style="margin-top: 12px">
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>根域名</th>
              <th>接收模式</th>
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
              <td>固定根域 + 泛子域路由</td>
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

const form = reactive({
  name: '',
  enabled: true,
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

  pushing.value = true
  try {
    const { data } = await DomainAPI.push({
      names,
      enabled: true,
      randomLevel: false,
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

  try {
    await DomainAPI.create({
      name: form.name,
      enabled: form.enabled,
      randomLevel: false,
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
      randomLevel: false,
      level: item.level,
    })
    await load()
  } catch (e) {
    error.value = e?.response?.data?.error || '更新根域状态失败。'
  }
}

async function recheckMx(item) {
  error.value = ''
  pushMessage.value = ''
  try {
    await DomainAPI.push({
      names: [item.name],
      enabled: item.enabled,
      randomLevel: false,
      level: item.level,
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
</script>
