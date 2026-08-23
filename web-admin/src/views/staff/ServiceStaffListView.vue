<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createServiceStaff, deleteServiceStaff, getServiceStaffList, resetServiceStaffPassword, updateServiceStaffStatus } from '@/api/sp'
import type { ServiceStaffItem } from '@/types/sp'

const loading = ref(false)
const list = ref<ServiceStaffItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const statusFilter = ref<number | string>('')
const keyword = ref('')

const statusOptions = [
  { label: '全部', value: '' },
  { label: '待审核', value: 0 },
  { label: '启用', value: 1 },
  { label: '禁用', value: 2 }
]

function statusText(s: number) {
  return { 0: '待审核', 1: '启用', 2: '禁用' }[s] || '未知'
}
function statusType(s: number) {
  return ({ 0: 'warning', 1: 'success', 2: 'danger' } as const)[s] || 'info'
}

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, unknown> = { page: page.value, page_size: pageSize.value }
    if (statusFilter.value !== '') params.status = statusFilter.value
    if (keyword.value) params.keyword = keyword.value
    const res = await getServiceStaffList(params as any)
    list.value = res.list || []
    total.value = res.total || 0
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    loading.value = false
  }
}

async function handleStatusChange(row: ServiceStaffItem, status: number) {
  const action = status === 1 ? (row.status === 0 ? '审核通过' : '启用') : '禁用'
  try {
    await ElMessageBox.confirm(`确认${action}服务人员「${row.name || row.username}」？`, action, { type: 'warning' })
    await updateServiceStaffStatus(row.id, status)
    ElMessage.success(`${action}成功`)
    loadData()
  } catch (e) {
    // 取消或失败
  }
}

async function handleResetPassword(row: ServiceStaffItem) {
  try {
    const { value } = await ElMessageBox.prompt('请输入新密码（至少6位）', '重置密码', {
      inputPattern: /^.{6,}$/,
      inputErrorMessage: '密码至少6位'
    })
    await resetServiceStaffPassword(row.id, value)
    ElMessage.success('密码重置成功')
  } catch (e) {
    // 取消
  }
}

async function handleDelete(row: ServiceStaffItem) {
  try {
    await ElMessageBox.confirm(`确认删除服务人员「${row.name || row.username}」？此操作不可恢复。`, '删除', { type: 'warning' })
    await deleteServiceStaff(row.id)
    ElMessage.success('删除成功')
    loadData()
  } catch (e) {
    // 取消
  }
}

function handlePageChange(p: number) {
  page.value = p
  loadData()
}

/* ----- 添加服务人员 ----- */
const dialogVisible = ref(false)
const saving = ref(false)
const form = ref({ username: '', password: '', name: '', phone: '' })

function openCreate() {
  form.value = { username: '', password: '', name: '', phone: '' }
  dialogVisible.value = true
}

async function handleCreate() {
  if (!form.value.username || !form.value.password || !form.value.name || !form.value.phone) {
    ElMessage.warning('请完整填写服务人员信息')
    return
  }
  saving.value = true
  try {
    await createServiceStaff(form.value)
    ElMessage.success('添加成功')
    dialogVisible.value = false
    loadData()
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    saving.value = false
  }
}

onMounted(loadData)
</script>

<template>
  <div class="page-card">
    <div class="filter-bar">
      <el-select v-model="statusFilter" placeholder="状态筛选" style="width: 140px" @change="() => { page = 1; loadData() }">
        <el-option v-for="o in statusOptions" :key="o.value" :label="o.label" :value="o.value" />
      </el-select>
      <el-input v-model="keyword" placeholder="姓名/手机号/用户名" style="width: 220px" clearable @clear="() => { page = 1; loadData() }" @keyup.enter="() => { page = 1; loadData() }" />
      <el-button type="primary" @click="() => { page = 1; loadData() }">搜索</el-button>
      <el-button type="primary" v-permission="'staff:create'" @click="openCreate">添加服务人员</el-button>
    </div>

    <el-dialog v-model="dialogVisible" title="添加服务人员" width="500px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="姓名" required>
          <el-input v-model="form.name" placeholder="服务人员姓名" />
        </el-form-item>
        <el-form-item label="手机号" required>
          <el-input v-model="form.phone" placeholder="手机号" />
        </el-form-item>
        <el-form-item label="登录账号" required>
          <el-input v-model="form.username" placeholder="登录用户名（2-64位）" />
        </el-form-item>
        <el-form-item label="登录密码" required>
          <el-input v-model="form.password" type="password" placeholder="登录密码（至少6位）" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleCreate">确认添加</el-button>
      </template>
    </el-dialog>

    <el-table v-loading="loading" :data="list" stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="name" label="姓名" width="120" />
      <el-table-column prop="phone" label="手机号" width="140" />
      <el-table-column prop="username" label="用户名" width="140" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="last_login_at" label="最后登录" width="180" />
      <el-table-column prop="created_at" label="注册时间" width="180" />
      <el-table-column label="操作" width="280" fixed="right">
        <template #default="{ row }">
          <el-button v-if="row.status !== 1" size="small" type="primary" @click="handleStatusChange(row, 1)">{{ row.status === 0 ? '审核通过' : '启用' }}</el-button>
          <el-button v-if="row.status === 1" size="small" type="warning" @click="handleStatusChange(row, 2)">禁用</el-button>
          <el-button size="small" @click="handleResetPassword(row)">重置密码</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
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
      @current-change="handlePageChange"
    />
  </div>
</template>

<style scoped>
.page-card { background: #fff; border-radius: 16px; padding: 24px; }
.filter-bar { display: flex; gap: 12px; margin-bottom: 20px; }
.pagination { margin-top: 20px; justify-content: flex-end; }
</style>
