<template>
  <div class="page grid" style="gap: 12px">
    <section class="card soft">
      <h1 class="section-title">域名管理</h1>
      <p class="section-sub">MX 指向本服务后，用户即可创建该域名下的临时邮箱。</p>
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
              <th>状态</th>
              <th>创建者</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in filteredItems" :key="item.id">
              <td>{{ item.id }}</td>
              <td>{{ item.name }}</td>
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
              <td colspan="5" class="meta">没有匹配的域名。</td>
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
const items = ref([])
const keyword = ref('')
const error = ref('')

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
</script>