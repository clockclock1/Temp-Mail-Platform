<template>
  <div class="page grid" style="gap: 12px">
    <section class="card soft">
      <div class="row wrap" style="justify-content: space-between">
        <div>
          <h1 class="section-title">角色与权限管理</h1>
          <p class="section-sub">按权限集合控制不同用户能力，支持快速创建、编辑、克隆角色。</p>
        </div>
        <div class="row">
          <button class="primary" @click="createRole">新建角色</button>
          <button class="ghost" :disabled="!selectedRole" @click="cloneRole">克隆选中角色</button>
        </div>
      </div>
      <p v-if="error" class="error" style="margin-top: 8px">{{ error }}</p>
    </section>

    <section class="grid grid-2">
      <div class="card">
        <h2 class="section-title" style="font-size: 16px">角色列表</h2>
        <div class="table-wrap" style="margin-top: 8px">
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>角色名</th>
                <th>权限数</th>
                <th>用户数</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="role in roles"
                :key="role.id"
                :style="selectedRole?.id === role.id ? 'background:#f1f3ff' : ''"
              >
                <td>{{ role.id }}</td>
                <td>{{ role.name }}</td>
                <td>{{ role.permissions?.length || 0 }}</td>
                <td>{{ role.userCount || 0 }}</td>
                <td>
                  <div class="row wrap">
                    <button class="ghost" @click="selectRole(role)">编辑</button>
                    <button class="danger" @click="deleteRole(role)">删除</button>
                  </div>
                </td>
              </tr>
              <tr v-if="roles.length === 0"><td colspan="5" class="meta">暂无角色</td></tr>
            </tbody>
          </table>
        </div>
      </div>

      <div class="card">
        <h2 class="section-title" style="font-size: 16px">角色编辑</h2>
        <p class="section-sub" v-if="!selectedRole">先选择左侧一个角色，或点击“新建角色”。</p>

        <div v-if="selectedRole" class="grid" style="margin-top: 8px">
          <label>
            角色名
            <input v-model="form.name" :disabled="selectedRole.name === 'admin'" />
          </label>
          <label>
            描述
            <input v-model="form.description" :disabled="selectedRole.name === 'admin'" />
          </label>

          <div class="card soft">
            <div class="row" style="justify-content: space-between">
              <strong>权限矩阵</strong>
              <span class="meta">已选 {{ form.permissionKeys.length }} 项</span>
            </div>

            <div class="grid" style="margin-top: 8px">
              <div v-for="group in groupedPermissions" :key="group.group" class="card">
                <div class="row" style="justify-content: space-between; margin-bottom: 6px">
                  <strong>{{ group.group }}</strong>
                  <button class="ghost" :disabled="selectedRole.name === 'admin'" @click="toggleGroup(group)">
                    {{ isGroupAllChecked(group) ? '取消全选' : '全选' }}
                  </button>
                </div>

                <div class="row wrap">
                  <label v-for="perm in group.items" :key="perm.key" class="badge" style="padding: 6px 8px">
                    <input
                      type="checkbox"
                      :value="perm.key"
                      v-model="form.permissionKeys"
                      :disabled="selectedRole.name === 'admin'"
                    />
                    {{ perm.key }}
                  </label>
                </div>
              </div>
            </div>
          </div>

          <div class="card soft">
            <strong>绑定用户</strong>
            <div class="row wrap" style="margin-top: 8px">
              <span class="badge" v-for="u in roleUsers" :key="u.id">{{ u.username }}{{ u.active ? '' : '(禁用)' }}</span>
              <span v-if="roleUsers.length === 0" class="meta">暂无绑定用户</span>
            </div>
          </div>

          <div class="row">
            <button class="primary" :disabled="selectedRole.name === 'admin'" @click="saveRole">保存修改</button>
            <button class="ghost" @click="reload">刷新</button>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { PermissionAPI, RoleAPI } from '../api'

const roles = ref([])
const permissions = ref([])
const selectedRole = ref(null)
const roleUsers = ref([])
const error = ref('')

const form = reactive({
  name: '',
  description: '',
  permissionKeys: [],
})

const groupedPermissions = computed(() => {
  const groups = {}
  for (const p of permissions.value) {
    const group = p.key.includes(':') ? p.key.split(':')[0] : 'other'
    if (!groups[group]) groups[group] = []
    groups[group].push(p)
  }
  return Object.keys(groups)
    .sort()
    .map((k) => ({ group: k, items: groups[k].sort((a, b) => a.key.localeCompare(b.key)) }))
})

onMounted(reload)

async function reload() {
  error.value = ''
  await Promise.all([loadRoles(), loadPermissions()])
  if (selectedRole.value) {
    const latest = roles.value.find((r) => r.id === selectedRole.value.id)
    if (latest) {
      await selectRole(latest)
    } else {
      selectedRole.value = null
      roleUsers.value = []
    }
  }
}

async function loadRoles() {
  const { data } = await RoleAPI.list()
  roles.value = data.items || []
}

async function loadPermissions() {
  const { data } = await PermissionAPI.list()
  permissions.value = data.items || []
}

async function selectRole(role) {
  selectedRole.value = role
  form.name = role.name
  form.description = role.description || ''
  form.permissionKeys = (role.permissions || []).map((p) => p.key)
  const { data } = await RoleAPI.users(role.id)
  roleUsers.value = data.items || []
}

function toggleGroup(group) {
  const keys = group.items.map((i) => i.key)
  const allChecked = keys.every((k) => form.permissionKeys.includes(k))
  if (allChecked) {
    form.permissionKeys = form.permissionKeys.filter((k) => !keys.includes(k))
  } else {
    const set = new Set([...form.permissionKeys, ...keys])
    form.permissionKeys = Array.from(set)
  }
}

function isGroupAllChecked(group) {
  return group.items.every((i) => form.permissionKeys.includes(i.key))
}

async function createRole() {
  const name = prompt('请输入角色名（如 reviewer）', '')
  if (!name || !name.trim()) return
  try {
    await RoleAPI.create({ name: name.trim(), description: '', permissionKeys: [] })
    await reload()
  } catch (e) {
    error.value = e?.response?.data?.error || '创建角色失败'
  }
}

async function cloneRole() {
  if (!selectedRole.value) return
  const base = selectedRole.value
  const name = prompt('请输入新角色名', `${base.name}-copy`)
  if (!name || !name.trim()) return
  try {
    await RoleAPI.create({
      name: name.trim(),
      description: (base.description || '') + ' (copy)',
      permissionKeys: (base.permissions || []).map((p) => p.key),
    })
    await reload()
  } catch (e) {
    error.value = e?.response?.data?.error || '克隆失败'
  }
}

async function saveRole() {
  if (!selectedRole.value) return
  try {
    await RoleAPI.update(selectedRole.value.id, {
      name: form.name,
      description: form.description,
      permissionKeys: form.permissionKeys,
    })
    await reload()
  } catch (e) {
    error.value = e?.response?.data?.error || '保存失败'
  }
}

async function deleteRole(role) {
  if (!confirm(`确定删除角色 ${role.name} 吗？`)) return
  try {
    await RoleAPI.remove(role.id)
    if (selectedRole.value?.id === role.id) {
      selectedRole.value = null
      roleUsers.value = []
    }
    await reload()
  } catch (e) {
    error.value = e?.response?.data?.error || '删除失败'
  }
}
</script>