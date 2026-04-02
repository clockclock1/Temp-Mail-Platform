<template>
  <div class="page grid" style="gap: 12px">
    <section class="card soft">
      <div class="row wrap" style="justify-content: space-between">
        <div>
          <h1 class="section-title">用户管理</h1>
          <p class="section-sub">支持筛选、分页、编辑、重置密码、状态管控。</p>
        </div>
        <button class="primary" @click="openCreate">新增用户</button>
      </div>

      <div class="grid grid-4" style="margin-top: 10px">
        <label>
          关键词
          <input v-model="query.q" placeholder="用户名/昵称" @keyup.enter="loadUsers" />
        </label>
        <label>
          角色
          <select v-model="query.roleId">
            <option value="">全部角色</option>
            <option v-for="r in roles" :key="r.id" :value="String(r.id)">{{ r.name }}</option>
          </select>
        </label>
        <label>
          状态
          <select v-model="query.active">
            <option value="">全部</option>
            <option value="true">启用</option>
            <option value="false">禁用</option>
          </select>
        </label>
        <label>
          每页
          <select v-model.number="query.pageSize">
            <option :value="10">10</option>
            <option :value="20">20</option>
            <option :value="50">50</option>
          </select>
        </label>
      </div>

      <div class="row" style="margin-top: 10px">
        <button class="secondary" @click="loadUsers">查询</button>
        <button class="ghost" @click="resetQuery">重置</button>
      </div>
      <p v-if="error" class="error" style="margin-top: 8px">{{ error }}</p>
    </section>

    <section class="card">
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>用户名</th>
              <th>昵称</th>
              <th>角色</th>
              <th>权限数</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="u in users" :key="u.id">
              <td>{{ u.id }}</td>
              <td>{{ u.username }}</td>
              <td>{{ u.displayName || '-' }}</td>
              <td>{{ u.role?.name || '-' }}</td>
              <td>{{ u.role?.permissions?.length || 0 }}</td>
              <td>
                <span class="badge" :class="u.active ? 'ok' : 'off'">{{ u.active ? '启用' : '禁用' }}</span>
              </td>
              <td>
                <div class="row wrap">
                  <button class="ghost" @click="openEdit(u)">编辑</button>
                  <button class="secondary" @click="toggleActive(u)">{{ u.active ? '禁用' : '启用' }}</button>
                  <button class="secondary" @click="resetPassword(u)">重置密码</button>
                  <button class="danger" @click="remove(u)">删除</button>
                </div>
              </td>
            </tr>
            <tr v-if="users.length === 0">
              <td colspan="7" class="meta">暂无数据。</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="row" style="justify-content: space-between; margin-top: 10px">
        <span class="meta">共 {{ total }} 条，第 {{ query.page }} / {{ totalPages }} 页</span>
        <div class="row">
          <button class="ghost" :disabled="query.page <= 1" @click="prevPage">上一页</button>
          <button class="ghost" :disabled="query.page >= totalPages" @click="nextPage">下一页</button>
        </div>
      </div>
    </section>

    <div class="modal" v-if="showCreate" @click.self="showCreate = false">
      <div class="modal-card">
        <h2 class="section-title">新增用户</h2>
        <div class="grid grid-2" style="margin-top: 10px">
          <label>
            用户名
            <input v-model="createForm.username" />
          </label>
          <label>
            昵称
            <input v-model="createForm.displayName" />
          </label>
          <label>
            角色
            <select v-model.number="createForm.roleId">
              <option :value="0">请选择角色</option>
              <option v-for="r in roles" :key="r.id" :value="r.id">{{ r.name }}</option>
            </select>
          </label>
          <label>
            初始密码
            <input type="password" v-model="createForm.password" placeholder="至少 8 位" />
          </label>
        </div>
        <div class="row" style="margin-top: 12px">
          <button class="primary" @click="submitCreate">创建</button>
          <button class="ghost" @click="showCreate = false">取消</button>
        </div>
      </div>
    </div>

    <div class="modal" v-if="showEdit" @click.self="showEdit = false">
      <div class="modal-card">
        <h2 class="section-title">编辑用户</h2>
        <p class="section-sub">用户名：{{ editForm.username }}</p>

        <div class="grid grid-2" style="margin-top: 10px">
          <label>
            昵称
            <input v-model="editForm.displayName" />
          </label>
          <label>
            角色
            <select v-model.number="editForm.roleId">
              <option v-for="r in roles" :key="r.id" :value="r.id">{{ r.name }}</option>
            </select>
          </label>
          <label>
            状态
            <select v-model="editForm.active">
              <option :value="true">启用</option>
              <option :value="false">禁用</option>
            </select>
          </label>
          <label>
            新密码（可留空）
            <input type="password" v-model="editForm.password" />
          </label>
        </div>

        <div class="card soft" style="margin-top: 10px">
          <div class="meta">该角色权限</div>
          <div class="row wrap" style="margin-top: 6px">
            <span class="badge" v-for="p in selectedRolePerms" :key="p">{{ p }}</span>
            <span v-if="selectedRolePerms.length === 0" class="meta">无权限</span>
          </div>
        </div>

        <div class="row" style="margin-top: 12px">
          <button class="primary" @click="submitEdit">保存</button>
          <button class="ghost" @click="showEdit = false">关闭</button>
        </div>
      </div>
    </div>

    <div class="modal" v-if="tempPassword" @click.self="tempPassword = ''">
      <div class="modal-card">
        <h2 class="section-title">已重置密码</h2>
        <p class="section-sub">系统自动生成了新密码，请立即保存并通知用户。</p>
        <pre class="code" style="margin-top: 10px">{{ tempPassword }}</pre>
        <div class="row" style="margin-top: 10px">
          <button class="primary" @click="copy(tempPassword)">复制密码</button>
          <button class="ghost" @click="tempPassword = ''">关闭</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { RoleAPI, UserAPI } from '../api'

const users = ref([])
const roles = ref([])
const total = ref(0)
const error = ref('')
const showCreate = ref(false)
const showEdit = ref(false)
const tempPassword = ref('')

const query = reactive({ q: '', roleId: '', active: '', page: 1, pageSize: 20 })

const createForm = reactive({ username: '', displayName: '', password: '', roleId: 0 })
const editForm = reactive({ id: 0, username: '', displayName: '', password: '', roleId: 0, active: true })

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / query.pageSize)))
const selectedRolePerms = computed(() => {
  const role = roles.value.find((r) => r.id === editForm.roleId)
  return (role?.permissions || []).map((p) => p.key)
})

onMounted(async () => {
  await Promise.all([loadRoles(), loadUsers()])
})

async function loadRoles() {
  const { data } = await RoleAPI.list()
  roles.value = data.items || []
}

async function loadUsers() {
  error.value = ''
  try {
    const params = {
      q: query.q || undefined,
      roleId: query.roleId || undefined,
      active: query.active || undefined,
      page: query.page,
      pageSize: query.pageSize,
    }
    const { data } = await UserAPI.list(params)
    users.value = data.items || []
    total.value = data.total || users.value.length
  } catch (e) {
    error.value = e?.response?.data?.error || '加载用户失败'
  }
}

function resetQuery() {
  query.q = ''
  query.roleId = ''
  query.active = ''
  query.page = 1
  query.pageSize = 20
  loadUsers()
}

function prevPage() {
  if (query.page <= 1) return
  query.page -= 1
  loadUsers()
}

function nextPage() {
  if (query.page >= totalPages.value) return
  query.page += 1
  loadUsers()
}

function openCreate() {
  createForm.username = ''
  createForm.displayName = ''
  createForm.password = ''
  createForm.roleId = 0
  showCreate.value = true
}

async function submitCreate() {
  if (!createForm.username || !createForm.password || !createForm.roleId) {
    alert('请填写用户名、密码和角色')
    return
  }
  await UserAPI.create(createForm)
  showCreate.value = false
  query.page = 1
  await loadUsers()
}

async function openEdit(user) {
  const { data } = await UserAPI.get(user.id)
  const item = data.item
  editForm.id = item.id
  editForm.username = item.username
  editForm.displayName = item.displayName || ''
  editForm.password = ''
  editForm.roleId = item.roleId
  editForm.active = item.active
  showEdit.value = true
}

async function submitEdit() {
  const payload = {
    displayName: editForm.displayName,
    roleId: editForm.roleId,
    active: editForm.active,
  }
  if (editForm.password) payload.password = editForm.password
  await UserAPI.update(editForm.id, payload)
  showEdit.value = false
  await loadUsers()
}

async function toggleActive(user) {
  await UserAPI.update(user.id, { active: !user.active })
  await loadUsers()
}

async function resetPassword(user) {
  const custom = prompt('输入新密码（留空则自动生成强密码）', '')
  const payload = custom && custom.trim() ? { password: custom.trim() } : {}
  const { data } = await UserAPI.resetPassword(user.id, payload)
  if (data.password) {
    tempPassword.value = data.password
  } else {
    alert('密码已重置')
  }
}

async function remove(user) {
  if (!confirm(`确定删除用户 ${user.username} 吗？`)) return
  await UserAPI.remove(user.id)
  await loadUsers()
}

async function copy(text) {
  await navigator.clipboard.writeText(text)
  alert('已复制')
}
</script>