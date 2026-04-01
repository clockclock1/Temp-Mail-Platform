<template>
  <div class="grid">
    <section class="card">
      <h1>域名管理</h1>
      <p class="meta">只需为子域名设置 MX 记录指向本服务，即可接收该域名下的临时邮箱邮件。</p>
      <div class="grid grid-2">
        <label>
          域名
          <input v-model="form.name" placeholder="mail.example.com" />
        </label>
        <label>
          启用状态
          <select v-model="form.enabled">
            <option :value="true">启用</option>
            <option :value="false">禁用</option>
          </select>
        </label>
      </div>
      <button class="primary" style="margin-top: 12px" @click="create">新增域名</button>
      <p v-if="error" class="error">{{ error }}</p>
    </section>

    <section class="card">
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>域名</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in items" :key="item.id">
              <td>{{ item.id }}</td>
              <td>{{ item.name }}</td>
              <td><span class="badge">{{ item.enabled ? '启用' : '禁用' }}</span></td>
              <td>
                <button class="secondary" @click="toggle(item)">{{ item.enabled ? '禁用' : '启用' }}</button>
                <button class="danger" @click="remove(item.id)">删除</button>
              </td>
            </tr>
            <tr v-if="items.length === 0"><td colspan="4" class="meta">暂无域名</td></tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { DomainAPI } from '../api'

const form = reactive({ name: '', enabled: true })
const items = ref([])
const error = ref('')

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
    error.value = e?.response?.data?.error || '创建失败'
  }
}

async function toggle(item) {
  await DomainAPI.update(item.id, { name: item.name, enabled: !item.enabled })
  await load()
}

async function remove(id) {
  if (!confirm('确定删除该域名吗？')) return
  await DomainAPI.remove(id)
  await load()
}
</script>