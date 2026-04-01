<template>
  <div class="grid grid-2">
    <section class="card">
      <h1>角色与权限</h1>
      <label>
        角色名
        <input v-model="form.name" placeholder="support" />
      </label>
      <label>
        描述
        <input v-model="form.description" placeholder="客服查看邮箱" />
      </label>
      <p class="meta">权限</p>
      <div class="grid" style="max-height: 220px; overflow: auto">
        <label v-for="p in permissions" :key="p.id">
          <input type="checkbox" :value="p.key" v-model="form.permissionKeys" />
          {{ p.key }}
        </label>
      </div>
      <div style="display: flex; gap: 8px; margin-top: 12px">
        <button class="primary" @click="create">新增角色</button>
        <button class="secondary" :disabled="!selectedRoleId" @click="update">更新选中角色</button>
      </div>
      <p v-if="error" class="error">{{ error }}</p>
    </section>

    <section class="card">
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>角色</th>
              <th>权限数量</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="role in roles" :key="role.id" :style="role.id === selectedRoleId ? 'background: rgba(17,100,102,.08)' : ''">
              <td>{{ role.id }}</td>
              <td>{{ role.name }}</td>
              <td>{{ role.permissions?.length || 0 }}</td>
              <td>
                <button class="ghost" @click="pick(role)">编辑</button>
                <button class="danger" @click="remove(role)">删除</button>
              </td>
            </tr>
            <tr v-if="roles.length === 0"><td colspan="4" class="meta">暂无角色</td></tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { PermissionAPI, RoleAPI } from '../api'

const roles = ref([])
const permissions = ref([])
const selectedRoleId = ref(0)
const error = ref('')

const form = reactive({
  name: '',
  description: '',
  permissionKeys: [],
})

onMounted(async () => {
  await Promise.all([loadRoles(), loadPermissions()])
})

async function loadRoles() {
  const { data } = await RoleAPI.list()
  roles.value = data.items || []
}

async function loadPermissions() {
  const { data } = await PermissionAPI.list()
  permissions.value = data.items || []
}

function pick(role) {
  selectedRoleId.value = role.id
  form.name = role.name
  form.description = role.description || ''
  form.permissionKeys = (role.permissions || []).map((p) => p.key)
}

async function create() {
  error.value = ''
  if (!form.name) {
    error.value = '角色名不能为空'
    return
  }
  try {
    await RoleAPI.create(form)
    resetForm()
    await loadRoles()
  } catch (e) {
    error.value = e?.response?.data?.error || '创建失败'
  }
}

async function update() {
  if (!selectedRoleId.value) return
  error.value = ''
  try {
    await RoleAPI.update(selectedRoleId.value, form)
    await loadRoles()
  } catch (e) {
    error.value = e?.response?.data?.error || '更新失败'
  }
}

async function remove(role) {
  if (!confirm(`确定删除角色 ${role.name} 吗？`)) return
  try {
    await RoleAPI.remove(role.id)
    if (selectedRoleId.value === role.id) resetForm()
    await loadRoles()
  } catch (e) {
    error.value = e?.response?.data?.error || '删除失败'
  }
}

function resetForm() {
  selectedRoleId.value = 0
  form.name = ''
  form.description = ''
  form.permissionKeys = []
}
</script>