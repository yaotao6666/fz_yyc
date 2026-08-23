<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getMerchantStaffList,
  createMerchantStaff,
  updateMerchantStaff,
  deleteMerchantStaff,
  resetMerchantStaffPassword,
  getDepartmentTree,
  getAllRoles
} from '@/api/sp'
import type { MerchantStaffListItem, SysDepartment, SysRole } from '@/types/sp'

const loading = ref(false)
const list = ref<MerchantStaffListItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const keyword = ref('')

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, unknown> = { page: page.value, page_size: pageSize.value }
    if (keyword.value) params.keyword = keyword.value
    const res = await getMerchantStaffList(params)
    list.value = res.list || []
    total.value = res.pagination?.total || 0
  } catch (_e) {
    // 错误已由拦截器提示
  } finally {
    loading.value = false
  }
}

const statusText = (s: number) => (s === 1 ? '启用' : '禁用')
const statusType = (s: number): 'success' | 'danger' => (s === 1 ? 'success' : 'danger')
const roleText = (r: string) => (r === 'owner' ? '店铺负责人' : '普通员工')

/* ----- 部门/角色下拉数据 ----- */
const departmentTree = ref<SysDepartment[]>([])
const allRoles = ref<SysRole[]>([])

async function loadSelectOptions() {
  try {
    const [departments, roles] = await Promise.all([getDepartmentTree(), getAllRoles()])
    departmentTree.value = departments
    allRoles.value = roles
  } catch (_e) {
    // 错误已由拦截器提示
  }
}

/* ----- 员工弹窗 ----- */
const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'update'>('create')
const saving = ref(false)
const form = reactive({
  id: 0,
  name: '',
  phone: '',
  username: '',
  password: '',
  role: 'staff',
  department_id: null as number | null,
  role_ids: [] as number[]
})

function openCreate() {
  Object.assign(form, { id: 0, name: '', phone: '', username: '', password: '', role: 'staff', department_id: null, role_ids: [] })
  dialogMode.value = 'create'
  dialogVisible.value = true
}

function openUpdate(row: MerchantStaffListItem) {
  Object.assign(form, {
    id: row.id,
    name: row.name || '',
    phone: row.phone || '',
    username: row.username,
    password: '',
    role: row.role || 'staff',
    department_id: row.department_id ?? null,
    role_ids: (row.roles || []).map((r) => r.id)
  })
  dialogMode.value = 'update'
  dialogVisible.value = true
}

async function handleSave() {
  if (!form.name.trim() || !form.phone.trim()) {
    ElMessage.warning('请填写姓名和手机号')
    return
  }
  if (dialogMode.value === 'create' && (!form.username.trim() || !form.password.trim())) {
    ElMessage.warning('请填写用户名和密码')
    return
  }
  saving.value = true
  try {
    if (dialogMode.value === 'create') {
      await createMerchantStaff({
        name: form.name.trim(),
        phone: form.phone.trim(),
        username: form.username.trim(),
        password: form.password,
        role: form.role,
        department_id: form.department_id,
        role_ids: form.role_ids
      })
      ElMessage.success('创建成功')
    } else {
      await updateMerchantStaff(form.id, {
        name: form.name.trim(),
        phone: form.phone.trim(),
        role: form.role,
        department_id: form.department_id,
        role_ids: form.role_ids
      })
      ElMessage.success('更新成功')
    }
    dialogVisible.value = false
    loadData()
  } catch (_e) {
    // 错误已由拦截器提示
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: MerchantStaffListItem) {
  try {
    await ElMessageBox.confirm(`确认删除员工「${row.name || row.username}」？此操作不可恢复。`, '删除员工', { type: 'warning' })
    await deleteMerchantStaff(row.id)
    ElMessage.success('删除成功')
    loadData()
  } catch (_e) {
    // 取消
  }
}

async function handleResetPassword(row: MerchantStaffListItem) {
  try {
    const { value } = await ElMessageBox.prompt('请输入新密码（至少6位）', '重置密码', {
      inputPattern: /^.{6,}$/,
      inputErrorMessage: '密码至少6位'
    })
    await resetMerchantStaffPassword(row.id, value)
    ElMessage.success('密码重置成功')
  } catch (_e) {
    // 取消
  }
}

async function handleToggleStatus(row: MerchantStaffListItem) {
  const newStatus = row.status === 1 ? 0 : 1
  const action = newStatus === 1 ? '启用' : '禁用'
  try {
    await ElMessageBox.confirm(`确认${action}员工「${row.name || row.username}」？`, action, { type: 'warning' })
    await updateMerchantStaff(row.id, { status: newStatus })
    ElMessage.success(`${action}成功`)
    loadData()
  } catch (_e) {
    // 取消
  }
}

onMounted(() => {
  loadData()
  loadSelectOptions()
})
</script>

<template>
  <div class="page-card">
    <div class="toolbar">
      <div class="filter-bar">
        <el-input v-model="keyword" placeholder="姓名/手机号/用户名" style="width: 220px" clearable @clear="() => { page = 1; loadData() }" @keyup.enter="() => { page = 1; loadData() }" />
        <el-button type="primary" @click="() => { page = 1; loadData() }">搜索</el-button>
      </div>
      <div>
        <el-button type="primary" v-permission="'system:staff:create'" @click="openCreate">新增员工</el-button>
      </div>
    </div>

    <el-table v-loading="loading" :data="list" stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="name" label="姓名" width="100" />
      <el-table-column prop="phone" label="手机号" width="130" />
      <el-table-column prop="username" label="用户名" width="120" />
      <el-table-column label="角色" width="120">
        <template #default="{ row }">
          <el-tag size="small" :type="row.role === 'owner' ? 'danger' : 'default'">{{ roleText(row.role || 'staff') }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="部门" width="140">
        <template #default="{ row }">
          {{ row.department?.name || '-' }}
        </template>
      </el-table-column>
      <el-table-column label="关联角色" min-width="160">
        <template #default="{ row }">
          <el-tag v-for="r in (row.roles || [])" :key="r.id" size="small" style="margin-right: 4px; margin-bottom: 2px">{{ r.name }}</el-tag>
          <span v-if="!row.roles || !row.roles.length">-</span>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag size="small" :type="statusType(row.status || 1)">{{ statusText(row.status || 1) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="last_login_at" label="最后登录" width="170" />
      <el-table-column label="操作" width="280" fixed="right">
        <template #default="{ row }">
          <el-button size="small" v-permission="'system:staff:update'" @click="openUpdate(row)">编辑</el-button>
          <el-button size="small" v-permission="'system:staff:reset-password'" @click="handleResetPassword(row)">重置密码</el-button>
          <el-button
            size="small"
            :type="row.status === 1 ? 'warning' : 'success'"
            v-permission="'system:staff:update'"
            @click="handleToggleStatus(row)"
          >{{ row.status === 1 ? '禁用' : '启用' }}</el-button>
          <el-button size="small" type="danger" v-permission="'system:staff:delete'" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-if="total > pageSize"
      class="pagination"
      :current-page="page"
      :page-size="pageSize"
      :total="total"
      layout="prev, pager, next, total"
      @current-change="(p: number) => { page = p; loadData() }"
    />

    <!-- 员工编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="dialogMode === 'create' ? '新增员工' : '编辑员工'" width="520px" destroy-on-close>
      <el-form label-width="90px">
        <el-form-item label="姓名" required>
          <el-input v-model="form.name" placeholder="员工显示名称" />
        </el-form-item>
        <el-form-item label="手机号" required>
          <el-input v-model="form.phone" placeholder="手机号" />
        </el-form-item>
        <template v-if="dialogMode === 'create'">
          <el-form-item label="用户名" required>
            <el-input v-model="form.username" placeholder="登录用户名" />
          </el-form-item>
          <el-form-item label="密码" required>
            <el-input v-model="form.password" type="password" placeholder="密码（至少6位）" />
          </el-form-item>
        </template>
        <el-form-item label="角色">
          <el-radio-group v-model="form.role">
            <el-radio value="staff">普通员工</el-radio>
            <el-radio value="owner">店铺负责人</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="部门">
          <el-tree-select
            v-model="form.department_id"
            :data="departmentTree"
            :props="{ label: 'name', children: 'children' }"
            node-key="id"
            value-key="id"
            check-strictly
            clearable
            placeholder="选择所属部门"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="关联角色">
          <el-select v-model="form.role_ids" multiple placeholder="选择角色" style="width: 100%">
            <el-option v-for="r in allRoles" :key="r.id" :label="r.name" :value="r.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-card { background: #fff; border-radius: 16px; padding: 24px; }
.toolbar { display: flex; align-items: center; justify-content: space-between; margin-bottom: 20px; }
.filter-bar { display: flex; gap: 12px; }
.pagination { margin-top: 20px; justify-content: flex-end; }
</style>