<template>
  <div class="page grid" style="gap: 12px">
    <section class="card soft">
      <h1 class="section-title">域名管理</h1>
      <p class="section-sub">支持 1-10 级邮箱后缀；推送域名池后会把层级配置一次性写入，新增主域再推送即可。</p>
      <div class="grid" style="margin-top: 10px; gap: 12px">
        <label>
          发信域名池
          <textarea v-model="poolText" rows="4" placeholder="mail.example.com&#10;notice.example.com"></textarea>
        </label>
        <div class="grid grid-3">
          <label>
            层级模式
            <select v-model="poolForm.randomLevel">
              <option :value="false">固定层级</option>
              <option :value="true">1-7级随机</option>
            </select>
          </label>
          <label>
            域名层级
            <select v-model.number="poolForm.level" :disabled="poolForm.randomLevel">
              <option v-for="n in 10" :key="n" :value="n">{{ n }}级</option>
            </select>
          </label>
          <label v-if="poolForm.randomLevel">
            随机范围
            <input :value="`${poolForm.levelMin}-${poolForm.levelMax}级`" disabled />
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
          <button class="primary" :disabled="pushing" @click="pushPool">{{ pushing ? '推送中...' : '推送域名池' }}</button>
          <button class="ghost" :disabled="pushing" @click="load">刷新</button>
        </div>
        <p v-if="pushMessage" class="success">{{ pushMessage }}</p>
      </div>
      <p class="meta" style="margin-top: 8px">域名只要解析到对应泛解析子域名，就可以使用这些层级的邮箱。</p>
    </section>

    <section class="card">
      <div class="grid grid-3" style="margin-top: 10px">
        <label>
          域名
          <input v-model="form.name" placeholder="mail.example.com" />
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
      <div class="row" style="margin-top: 10px">
        <button class="primary" @click="create">新增域名</button>
        <button class="ghost" @click="load">刷新</button>
      </div>
      <p v-if="error" class="error" style="margin-top: 8px">{{ error }}</p>
    </section>

    <section class="card">
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>域名</th>
              <th>层级</th>
              <th>状态</th>
              <th>创建者</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in filteredItems" :key="item.id">
              <td>{{ item.id }}</td>
              <td>{{ item.name }}</td>
              <td>{{ levelLabel(item) }}</td>
              <td>
                <span class="badge" :class="item.enabled ? 'ok' : 'off'">
                  {{ item.enabled ? '启用' : '禁用' }}
                </span>
              </td>
              <td>{{ item.createdBy || '-' }}</td>
              <td>
                <div class="row">
                  <button class="secondary" @click="toggle(item)">{{ item.enabled ? '禁用' : '启用' }}</button>
                  <button class="danger" @click="remove(item.id)">删除</button>
                </div>
              </td>
            </tr>
            <tr v-if="filteredItems.length === 0">
              <td colspan="6" class="meta">没有匹配的域名。</td>
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
    error.value = '请先填写发信域名池'
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
    pushMessage.value = `已推送 ${data.created || 0} 个新域名，更新 ${data.updated || 0} 个域名。`
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
    error.value = '请输入域名'
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
</script>
