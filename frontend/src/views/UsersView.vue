<template>
  <div class="grid">
    <section class="card">
      <h1>用户管理</h1>
      <div class="grid grid-3">
        <label>
          用户名
          <input v-model="form.username" placeholder="dev01" />
        </label>
        <label>
          昵称
          <input v-model="form.displayName" placeholder="开发同学" />
        </label>
        <label>
          角色
          <select v-model.number="form.roleId">
            <option :value="0">请选择角色</option>
            <option v-for="role in roles" :key="role.id" :value="role.id">{{ role.name }}</option>
          </select>
        </label>
      </div>
      <label>
        初始密码
        <input v-model="form.password" type="password" placeholder="至少8位" />
      </label>
      <button class="primary" style="margin-top: 12px" @click="create">新增用户</button>
      <p v-if="error" class="error">{{ error }}</p>
    </section>

    <section class="card">
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>用户名</th>
              <th>角色</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="u in users" :key="u.id">
              <td>{{ u.id }}</td>
              <td>{{ u.username }}</td>
              <td>{{ u.role?.name }}</td>
              <td><span class="badge">{{ u.active ? '启用' : '禁用' }}</span></td>
              <td>
                <button class="secondary" @click="toggleActive(u)">{{ u.active ? '禁用' : '启用' }}</button>
                <button class="danger" @click="remove(u.id)">删除</button>
              </td>
            </tr>
            <tr v-if="users.length === 0"><td colspan="5" class="meta">暂无用户</td></tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { RoleAPI, UserAPI } from '../api'

const users = ref([])
const roles = ref([])
const error = ref('')
const form = reactive({ username: '', displayName: '', password: '', roleId: 0 })

onMounted(async () => {
  await Promise.all([loadUsers(), loadRoles()])
})

async function loadUsers() {
  const { data } = await UserAPI.list()
  users.value = data.items || []
}

async function loadRoles() {
  const { data } = await RoleAPI.list()
  roles.value = data.items || []
}

async function create() {
  error.value = ''
  if (!form.username || !form.password || !form.roleId) {
    error.value = '请填写完整信息'
    return
  }
  try {
    await UserAPI.create(form)
    form.username = ''
    form.displayName = ''
    form.password = ''
    form.roleId = 0
    await loadUsers()
  } catch (e) {
    error.value = e?.response?.data?.error || '创建失败'
  }
}

async function toggleActive(user) {
  await UserAPI.update(user.id, { active: !user.active })
  await loadUsers()
}

async function remove(id) {
  if (!confirm('确定删除该用户吗？')) return
  await UserAPI.remove(id)
  await loadUsers()
}
</script>